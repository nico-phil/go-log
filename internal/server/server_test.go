package server

import (
	"os"
	"testing"

	llog "github.com/nico-phil/go-log/internal/log"
	"github.com/stretchr/testify/require"
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
	NewGRPCServer(&c)
}
