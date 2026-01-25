package log

import (
	"sync"

	api "github.com/nico-phil/go-log/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Replicator
type Replicator struct {
	DialOptions []grpc.DialOption
	LocalServer api.LogClient
	logger      *zap.Logger
	mu          sync.Mutex
	servers     map[string]chan struct{} // addr->chan
	closed      bool
}

// Join
func (r *Replicator) Join(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.init()
	if r.closed {
		return nil
	}

	if _, ok := r.servers[name]; ok {
		return nil
	}

	r.servers[name] = make(chan struct{})
	go r.replicate(addr, r.servers[name])

	return nil
}

// replicate
func (r *Replicator) replicate(addr string, leave chan struct{}) {

}

// init
func (r *Replicator) init()
