package loadbalance_test

import (
	"net"
	"net/url"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/config"
	"github.com/nico-phil/go-log/internal/loadbalance"
	"github.com/nico-phil/go-log/internal/server"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

func TestResolver(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})

	require.NoError(t, err)
	serverCred := credentials.NewTLS(tlsConfig)

	srv, err := server.NewGRPCServer(&server.Config{
		GetServerer: &getServers{},
	}, grpc.Creds(serverCred))

	go srv.Serve(l)

	conn := clientConn{}

	tlsConfig, err = config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ClientKeyFile,
		CAFile:        config.CAFile,
		Server:        false,
		ServerAddress: "127.0.0.1",
	})

	require.NoError(t, err)

	clientCreds := credentials.NewTLS(tlsConfig)

	opts := resolver.BuildOptions{
		DialCreds: clientCreds,
	}

	r := &loadbalance.Resolver{}
	_, err = r.Build(
		resolver.Target{URL: url.URL{}},
		&conn,
		opts,
	)
}

type getServers struct{}

func (s *getServers) GetServers() ([]*api.Server, error) {
	return []*api.Server{
		{
			Id:       "leader",
			RpcAddr:  "localhost:9001",
			IsLeader: true,
		},
		{
			Id:      "follower",
			RpcAddr: "localhost:9002",
		},
	}, nil
}

type clientConn struct {
	resolver.ClientConn
	state resolver.State
}

func (c *clientConn) UpdateState(state resolver.State) error {
	c.state = state
	return nil
}

func (c *clientConn) ReportError(err error) {}

func (c *clientConn) NewAddress(addresses []resolver.Address) {}

func (c *clientConn) ParseServiceConfig(config string) *serviceconfig.ParseResult {
	return nil

}
