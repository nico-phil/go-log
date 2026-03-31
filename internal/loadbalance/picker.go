package loadbalance

import (
	"sync"

	"google.golang.org/grpc/balancer"
)

type Picker struct {
	mu        sync.RWMutex
	leader    balancer.SubConn
	followers []balancer.SubConn
}

func New() *Picker {
	return nil
}
