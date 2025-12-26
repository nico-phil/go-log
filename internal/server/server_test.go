package server

import (
	"context"
	"net"
	"os"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/config"
	llog "github.com/nico-phil/go-log/internal/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Test_Server tests our grpc server
func Test_Server(t *testing.T) {

	scenarios := map[string]func(t *testing.T, client api.LogClient){
		"testProduceConsume":       testProduceConsume,
		"testProduceConsumeStream": testProduceConsumeStream,
	}

	for sc, fn := range scenarios {
		t.Run(sc, func(t *testing.T) {
			client, teardown := SetupTest(t)
			defer teardown()
			fn(t, client)
		})
	}

}

// SetupTest run before and after each test function
func SetupTest(t *testing.T) (api.LogClient, func()) {
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

	serverTlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCetFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: lis.Addr().String(),
	})
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(serverTlsConfig)

	c := Config{
		CommitLog: commitLog,
	}
	grpcServer, err := NewGRPCServer(&c, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		grpcServer.Serve(lis)

	}()

	// client
	clientTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CAFile: config.CAFile,
	})
	require.NoError(t, err)

	clientCreds := credentials.NewTLS(clientTLSConfig)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(clientCreds),
	}
	cc, err := grpc.NewClient(lis.Addr().String(), opts...)
	require.NoError(t, err)

	client := api.NewLogClient(cc)

	return client, func() {
		grpcServer.Stop()
		cc.Close()
		lis.Close()
		commitLog.Remove()
	}
}

// testProduceConsume tests produce and consume handler
func testProduceConsume(t *testing.T, client api.LogClient) {
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
func testProduceConsumeStream(t *testing.T, client api.LogClient) {

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
