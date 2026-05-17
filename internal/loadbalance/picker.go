package loadbalance

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// compile-time checking
var _ base.PickerBuilder = (*Picker)(nil)

// Picker represents a grpc picker.
// It routes consume RPCs to follower servers and produce RPCs to the leader server
type Picker struct {
	mu sync.RWMutex

	// leader connection
	leader balancer.SubConn

	// list of the follower connection
	followers []balancer.SubConn

	current uint64
}

// Build uses ReadySCs(subconnections) map to differeciate leader and follower servers
func (p *Picker) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	p.mu.Lock()
	defer p.mu.Unlock()

	var followers []balancer.SubConn

	for sc, scInfo := range buildInfo.ReadySCs {
		isLeader := scInfo.Address.Attributes.Value("is_leader").(bool)

		if isLeader {
			p.leader = sc
			continue
		}

		followers = append(followers, sc)
	}

	p.followers = followers
	return p
}

var _ balancer.Picker = (*Picker)(nil)

// Pick routes consume rpc request to the follower and produce rpc request to the leader
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var result balancer.PickResult

	if strings.Contains(info.FullMethodName, "Produce") || len(p.followers) == 0 {
		result.SubConn = p.leader
	} else if strings.Contains(info.FullMethodName, "Consume") {
		result.SubConn = p.nextFollower()
	}

	log.Printf("result.SubConn: %+v", result.SubConn)
	if result.SubConn == nil {
		return result, balancer.ErrNoSubConnAvailable
	}

	return result, nil
}

// nextFollower returns the next follower uses round-robin algorithm
func (p *Picker) nextFollower() balancer.SubConn {
	cur := atomic.AddUint64(&p.current, uint64(1))
	len := uint64(len(p.followers))
	idx := cur % len
	return p.followers[idx]
}

func init() {
	balancer.Register(
		base.NewBalancerBuilder(Name, &Picker{}, base.Config{}),
	)
}
