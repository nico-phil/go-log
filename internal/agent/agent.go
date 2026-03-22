package agent

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/raft"
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
	log        *log.DistributedLog
	server     *grpc.Server
	membership *discovery.Membership

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
	Bootstrap       bool
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
		a.setupLogger,
		a.setMux,
		a.setupLog,
		a.setupServer,
		a.SetupMemberShip,
	}

	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}

	go a.Serve()

	return a, nil
}

// setMux Creates a listener on the rpc address. it will accept both raft and grpc connection
func (a *Agent) setMux() error {
	rpcAddr := fmt.Sprintf(":%d", a.Config.RPCPort)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}

	a.mux = cmux.New(ln)
	return nil
}

// setupLog sets up the distributed log
func (a *Agent) setupLog() error {

	raftLn := a.mux.Match(func(r io.Reader) bool {
		b := make([]byte, 1)
		if _, err := r.Read(b); err != nil {
			return false
		}

		return bytes.Compare(b, []byte{byte(log.RaftRPC)}) == 0

	})

	logConfig := log.Config{}

	logConfig.Raft.StreamLayer = log.NewStreamLayer(
		raftLn,
		a.ServerTlsConfig,
		a.PeerTlsConfig,
	)

	logConfig.Raft.LocalID = raft.ServerID(a.Config.NodeName)
	logConfig.Raft.Bootstrap = a.Config.Bootstrap

	var err error
	a.log, err = log.NewDistrubutedLog(a.Datadir, logConfig)
	if err != nil {
		return err
	}

	if a.Config.Bootstrap {
		err = a.log.WaitForLeader(time.Second * 3)
	}

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
		CommitLog:   a.log,
		Authorizer:  autorizer,
		GetServerer: a.log,
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

	grpcLn := a.mux.Match(cmux.Any())

	go func() {
		if err := a.server.Serve(grpcLn); err != nil {
			_ = a.Shutdown()
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

	a.membership, err = discovery.New(a.log, discovery.Config{
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
		// a.replicator.Close,
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

func (a *Agent) Serve() error {
	if err := a.mux.Serve(); err != nil {
		_ = a.Shutdown()
		return err
	}

	return nil
}
