package server

import (
	"context"
	"net"
	"os"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	llog "github.com/nico-phil/go-log/internal/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Test_Server tests our grpc server
func Test_Server(t *testing.T) {
	dir, err := os.MkdirTemp("", "data_test")
	require.NoError(t, err)

	logConfig := llog.Config{}
	logConfig.Segment.MaxIndexBytes = 1024
	logConfig.Segment.MaxStoreBytes = 1024
	commitLog, err := llog.NewLog(dir, logConfig)
	require.NoError(t, err)

	c := Config{
		CommitLog: commitLog,
	}
	grpcServer, err := NewGRPCServer(&c)
	require.NoError(t, err)

	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		grpcServer.Serve(lis)
	}()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	cc, err := grpc.NewClient(lis.Addr().String(), opts...)
	client := api.NewLogClient(cc)
	require.NoError(t, err)

	record := api.Record{
		Value: []byte("hello world"),
	}

	produceResp, err := client.Produce(context.Background(), &api.ProduceRequest{Record: &record})
	require.NoError(t, err)

	consumeResp, err := client.Consume(context.Background(), &api.ConsumeRequest{Offset: produceResp.Offset})
	require.NoError(t, err)
	require.Equal(t, record.Value, consumeResp.Record.Value)

}
