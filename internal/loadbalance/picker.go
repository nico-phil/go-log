package loadbalance

import (
	"debug/buildinfo"
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

func (p *Picker) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	p.mu.Lock()
	defer p.mu.Unlock()

	var followers []balancer.SubConn

	for sc, scInfo := range buildInfo.ReadySCs {
		isLeader := scInfo.Address.Attributes.Value("is_leader").(bool) {

		}

		if isLeader = p.leader {
			p.leader = sc
		}

		followers = append(followers, sc)
	}

	 p.followers = followers
	return p
}


var _ balancer.Picker = Picker(nil)

func(p *Picker) Pick(info balancer.PickInfo) (Pick balancer.PickResult, error){
	return 	nil, nil
}