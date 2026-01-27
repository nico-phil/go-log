package agent

import (
	"crypto/tls"
	"sync"

	"github.com/nico-phil/go-log/internal/discovery"
	"github.com/nico-phil/go-log/internal/log"
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
}

// New create an agent
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
