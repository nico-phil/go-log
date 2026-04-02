package loadbalance

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

var _ base.PickerBuilder = (*Picker)(nil)

type Picker struct {
	mu        sync.RWMutex
	leader    balancer.SubConn
	followers []balancer.SubConn
}

func New() *Picker {
	return nil
}

func (p *Picker) Build(info base.PickerBuildInfo) balancer.Picker {
	return nil
}
