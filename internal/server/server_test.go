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
		Value:  []byte("hello world"),
		Offset: 0,
	}

	produceResp, err := client.Produce(context.Background(), &api.ProduceRequest{Record: &record})
	require.NoError(t, err)

	consumeResp, err := client.Consume(context.Background(), &api.ConsumeRequest{Offset: produceResp.Offset})
	require.NoError(t, err)
	require.Equal(t, record.Value, consumeResp.Record.Value)
	require.Equal(t, produceResp.Offset, int64(consumeResp.Record.Offset))

	// ProduceStream
	stream, err := client.ProduceStream(context.Background())
	require.NoError(t, err)
	records := []*api.Record{
		{Value: []byte("first message"), Offset: 0},
		{Value: []byte("second message"), Offset: 1},
	}

	for offset, rec := range records {
		err = stream.Send(&api.ProduceRequest{Record: rec})
		require.NoError(t, err)

		resp, err := stream.Recv()
		require.NoError(t, err)

		require.Equal(t, int64(offset+1), resp.Offset)
	}

	// cosume stream
	streamC, err := client.ConsumeStream(context.Background(), &api.ConsumeRequest{Offset: 0})
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		resp, err := streamC.Recv()
		require.NoError(t, err)
		if i == 0 {
			rec := &record
			require.Equal(t, rec.Offset, resp.Record.Offset)
		} else {
			rec := records[i-1]
			require.Equal(t, rec.Offset, resp.Record.Offset)
		}

	}

}
