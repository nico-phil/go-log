package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/auth"
	"github.com/nico-phil/go-log/internal/discovery"
	"github.com/nico-phil/go-log/internal/log"
	"github.com/nico-phil/go-log/internal/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/soheilhy/cmux"
)

// Agent contains All the differents components of system
type Agent struct {
	Config
	log        *log.Log
	server     *grpc.Server
	membership *discovery.Membership
	replicator *log.Replicator

	shutdown     bool
	shutdowns    chan struct{}
	shutdownLock sync.Mutex

	mux cmux.CMux
}

// Config contains configurations for all components
type Config struct {
	ServerTlsConfig *tls.Config
	PeerTlsConfig   *tls.Config
	Datadir         string
	BindAddr        string
	RPCPort         int
	NodeName        string
	StartJoinAddr   []string
	ACLModelFile    string
	ACLPolicyFile   string
}

// RPCAddr splits a network address of the form "host:port
func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

// New create an agent, it contains a set of method to set up and run the agent components.
// after we run New, we expect to have a running, functioning service
func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}

	setup := []func() error{
		a.setupLog,
		a.setupLogger,
		a.setupServer,
		a.SetupMemberShip,
	}

	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}

	return a, nil
}

// setMux Creates a listener on our rpc address. it will accept both raft and grpc connection
func (a *Agent) setMux() error {
	rpcAddr := fmt.Sprintf(":%d", a.Config.RPCPort)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}

	a.mux = cmux.New(ln)
	return nil
}

// setupLog sets up the WAL
func (a *Agent) setupLog() error {
	var err error
	a.log, err = log.NewLog(
		a.Config.Datadir,
		log.Config{},
	)

	return err
}

// setupLogger sets up the logger agent
func (a *Agent) setupLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	return nil
}

// setupServer sets up the server for the agent
func (a *Agent) setupServer() error {
	autorizer := auth.New(
		a.Config.ACLModelFile,
		a.Config.ACLPolicyFile,
	)

	serverConfig := server.Config{
		CommitLog:  a.log,
		Authorizer: autorizer,
	}

	var opts []grpc.ServerOption
	if a.Config.ServerTlsConfig != nil {
		creds := credentials.NewTLS(a.Config.ServerTlsConfig)
		opts = append(opts, grpc.Creds(creds))
	}

	var err error
	a.server, err = server.NewGRPCServer(&serverConfig, opts...)
	if err != nil {
		return err
	}

	rpcAddr, err := a.RPCAddr()
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}

	go func() {
		if err := a.server.Serve(ln); err != nil {
			a.Shutdown()
		}
	}()

	return err

}

// SetupMemberShip sets up a replicator with grcp client so that the replicator can
// connect to other servers, consume their data and produce a copy of the data to the local server
func (a *Agent) SetupMemberShip() error {
	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}
	var opts []grpc.DialOption
	if a.Config.PeerTlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(a.Config.PeerTlsConfig),
		))
	}
	conn, err := grpc.NewClient(rpcAddr, opts...)
	if err != nil {
		return err
	}

	client := api.NewLogClient(conn)
	a.replicator = &log.Replicator{
		DialOptions: opts,
		LocalServer: client,
	}

	a.membership, err = discovery.New(a.replicator, discovery.Config{
		NodeName: a.Config.NodeName,
		BindAddr: a.Config.BindAddr,
		Tags: map[string]string{
			"rpc_addr": rpcAddr,
		},
		StartJoinAddrs: a.Config.StartJoinAddr,
	})

	return err
}

// Shutdown
func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()
	if a.shutdown {
		return nil
	}
	a.shutdown = true
	close(a.shutdowns)

	shutdown := []func() error{
		a.membership.Leave,
		a.replicator.Close,
		func() error {
			a.server.GracefulStop()
			return nil
		},
		a.log.Close,
	}
	for _, fn := range shutdown {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
