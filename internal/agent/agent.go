package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/nico-phil/go-log/internal/auth"
	"github.com/nico-phil/go-log/internal/discovery"
	"github.com/nico-phil/go-log/internal/log"
	"github.com/nico-phil/go-log/internal/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
}

// Config contains configurations for all components
type Config struct {
	ServerTlsConfig *tls.Config
	PeerTlsConfig   *tls.Config
	Datadir         string
	BindAddr        string
	RPCPort         int
	NodeName        string
	StartJoinAdd    []string
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
	}

	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}

	return a, nil
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

func (a *Agent) Shutdown() {}
