package loadbalance

import (
	"context"
	"fmt"
	"log"
	"sync"

	api "github.com/nico-phil/go-log/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

var Name = "proglog"

// Resolver implement resolver.Builder and reolver.Resolver interface from grpc
type Resolver struct {
	mu            sync.Mutex
	clientConn    resolver.ClientConn
	resolverConn  *grpc.ClientConn
	serviceConfig *serviceconfig.ParseResult
	logger        *zap.Logger
}

// for compile-time check
var _ resolver.Builder = (*Resolver)(nil)

// Build sets up a client connecton to our server, so the resolver can call the Getservers API
func (r *Resolver) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {
	r.logger = zap.L().Named("resolver")
	r.clientConn = cc
	var dialOpts []grpc.DialOption

	if opts.DialCreds != nil {
		dialOpts = append(
			dialOpts,
			grpc.WithTransportCredentials(opts.DialCreds))
	}

	r.serviceConfig = r.clientConn.ParseServiceConfig(
		fmt.Sprintf(`{"loadBalancingConfig:[{"%s:{}}]}`, Name),
	)

	log.Printf("loadbalancer-config: %+v", r.serviceConfig)

	var err error
	r.resolverConn, err = grpc.NewClient(target.Endpoint(), dialOpts...)
	if err != nil {
		return nil, err
	}

	r.ResolveNow(resolver.ResolveNowOptions{})
	return nil, err
}

// Scheme return the resolver's scheme identifer like this: proglog://address:port
func (r *Resolver) Scheme() string {
	return Name
}

// register the resolver
func init() {
	resolver.Register(&Resolver{})
}

// Resolver should implment resolver.Resolver interface. I'ts a compile-time checking
var _ resolver.Resolver = (*Resolver)(nil)

// ResolveNow will be called by gRPC to try to resolve the target name,
// discover the servers and update the client connection with servers
func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client := api.NewLogClient(r.resolverConn)

	ctx := context.Background()

	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		r.logger.Error(
			"failed to resolve server",
			zap.Error(err),
		)
		return
	}

	var addrs []resolver.Address

	for _, server := range res.Servers {
		addrs = append(addrs, resolver.Address{
			Addr: server.RpcAddr,
			Attributes: attributes.New(
				"is_leader",
				server.IsLeader,
			),
		})
	}

	r.clientConn.UpdateState(resolver.State{
		Addresses:     addrs,
		ServiceConfig: r.serviceConfig,
	})
}

// Close closes the resolver.
func (r *Resolver) Close() {
	if err := r.resolverConn.Close(); err != nil {
		r.logger.Error(
			"failed to close connection",
			zap.Error(err),
		)
	}
}

type clientConn struct{}

func (s *clientConn) Report() {}

func (s *clientConn) NewAddress() {}

func (s *clientConn) NewServiceConfig() {}

func (s *clientConn) UpdateState() {}

func (s *clientConn) ParseServiceConfig(config string) {}
