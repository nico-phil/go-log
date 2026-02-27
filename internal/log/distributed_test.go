package log_test

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/log"
	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
)

// TestDistributedLog test the distributedLog
func TestDistributedLog(t *testing.T) {
	var logs []*log.DistributedLog
	nodeCount := 3
	ports := dynaport.Get(nodeCount)
	for i := 0; i < nodeCount; i++ {
		dataDir, err := os.MkdirTemp("", "distributed-log-test")
		require.NoError(t, err)
		defer func(dir string) {
			_ = os.Remove(dir)
		}(dataDir)

		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", ports[i]))
		require.NoError(t, err)

		config := log.Config{}

		config.Raft.StreamLayer = log.NewStreamLayer(ln, nil, nil)
		config.Raft.LocalID = raft.ServerID(fmt.Sprintf("%d", i))
		config.Raft.HeartbeatTimeout = 50 * time.Millisecond
		config.Raft.ElectionTimeout = 50 * time.Millisecond
		config.Raft.LeaderLeaseTimeout = 50 * time.Millisecond
		config.Raft.CommitTimeout = 5 * time.Millisecond

		if i == 0 {
			config.Raft.Bootstrap = true
		}

		l, err := log.NewDistrubutedLog(dataDir, config)
		require.NoError(t, err)

		if i != 0 {
			err = logs[0].Join(fmt.Sprintf("%d", i), ln.Addr().String())
			require.NoError(t, err)
		} else {
			err = l.WaitForLeader(3 * time.Second)
		}

		logs = append(logs, l)
	}

	records := []*api.Record{
		{Value: []byte("first")},
		{Value: []byte("second")},
	}

	for _, record := range records {
		off, err := logs[0].Append(record)
		require.NoError(t, err)

		fmt.Printf("AppendOffset:%d\n", off)
		time.Sleep(time.Second)

		for i := 0; i < nodeCount; i++ {
			got, err := logs[i].Read(off)
			require.NoError(t, err)
			record.Offset = off

			fmt.Printf("record: %v --- got: %v --- i=%d --- offset=%d\n", string(record.Value), string(got.Value), i, off)

			require.Equal(t, record.Value, got.Value)
		}
	}

}
