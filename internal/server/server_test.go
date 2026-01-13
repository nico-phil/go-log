package server

import (
	"context"
	"net"
	"os"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/auth"
	"github.com/nico-phil/go-log/internal/config"
	llog "github.com/nico-phil/go-log/internal/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Test_Server tests our grpc server
func Test_Server(t *testing.T) {

	scenarios := map[string]func(t *testing.T, client, _ api.LogClient, config *Config){
		"testProduceConsume":       testProduceConsume,
		"testProduceConsumeStream": testProduceConsumeStream,
	}

	for sc, fn := range scenarios {
		t.Run(sc, func(t *testing.T) {
			rootclient, nobodyClient, config, teardown := SetupTest(t)
			defer teardown()
			fn(t, rootclient, nobodyClient, config)
		})
	}

}

// SetupTest run before and after each test function
func SetupTest(t *testing.T) (rootClient api.LogClient, nobodyClient api.LogClient, c *Config, teardown func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)

	logConfig := llog.Config{}
	logConfig.Segment.MaxIndexBytes = 1024
	logConfig.Segment.MaxStoreBytes = 1024
	commitLog, err := llog.NewLog(dir, logConfig)
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	newClient := func(crtFile, keyFile string) (*grpc.ClientConn, api.LogClient, []grpc.DialOption) {
		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile:      crtFile,
			KeyFile:       keyFile,
			CAFile:        config.CAFile,
			ServerAddress: lis.Addr().String(),
		})
		require.NoError(t, err)

		tlsCreds := credentials.NewTLS(tlsConfig)

		opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
		conn, err := grpc.NewClient(lis.Addr().String(), opts...)

		client := api.NewLogClient(conn)
		return conn, client, opts
	}

	var rootConn *grpc.ClientConn
	rootConn, rootClient, _ = newClient(
		config.RootClientCertFile,
		config.RootClientKeyFile,
	)

	var nobodyConn *grpc.ClientConn
	nobodyConn, nobodyClient, _ = newClient(
		config.RootClientCertFile,
		config.RootClientKeyFile,
	)

	serverTlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: lis.Addr().String(),
		Server:        true,
	})

	require.NoError(t, err)

	serverCreds := credentials.NewTLS(serverTlsConfig)
	authorizer := auth.New("", "")
	c = &Config{
		CommitLog:  commitLog,
		Authorizer: authorizer,
	}
	grpcServer, err := NewGRPCServer(c, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		grpcServer.Serve(lis)

	}()

	return rootClient, nobodyClient, c, func() {
		grpcServer.Stop()
		rootConn.Close()
		nobodyConn.Close()
		commitLog.Remove()
	}
}

// testProduceConsume tests produce and consume handler
func testProduceConsume(t *testing.T, client, _ api.LogClient, config *Config) {
	t.Helper()
	record := api.Record{
		Value:  []byte("hello world"),
		Offset: 0,
	}
	produceResp, err := client.Produce(context.Background(), &api.ProduceRequest{Record: &record})
	require.NoError(t, err)

	consumeResp, err := client.Consume(context.Background(), &api.ConsumeRequest{Offset: produceResp.Offset})
	require.NoError(t, err)
	require.Equal(t, record.Value, consumeResp.Record.Value)
	require.Equal(t, produceResp.Offset, int64(consumeResp.Record.Offset))

}

// testProduceConsumeStream tests produce and consume stream
func testProduceConsumeStream(t *testing.T, client, _ api.LogClient, config *Config) {

	records := []*api.Record{
		{Value: []byte("first message"), Offset: 0},
		{Value: []byte("second message"), Offset: 1},
	}
	stream, err := client.ProduceStream(context.Background())
	require.NoError(t, err)

	for offset, rec := range records {
		err = stream.Send(&api.ProduceRequest{Record: rec})
		require.NoError(t, err)

		resp, err := stream.Recv()
		require.NoError(t, err)

		require.Equal(t, uint64(offset), uint64(resp.Offset))
	}

	streamC, err := client.ConsumeStream(context.Background(), &api.ConsumeRequest{Offset: 0})
	require.NoError(t, err)

	for range records {
		_, err := streamC.Recv()
		require.NoError(t, err)

		// require.Equal(t, uint64(resp.Record.Offset), uint64(i))
	}

}
