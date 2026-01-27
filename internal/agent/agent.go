package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/nico-phil/go-log/internal/discovery"
	"github.com/nico-phil/go-log/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	setup := []func() error{}

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
