package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDistributedLog test the distributedLog
func TestDistributedLog(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "data-log")
	require.NoError(t, err)
	defer os.Remove(tempDir)

	config := Config{}
	config.Segment.MaxIndexBytes = 1024
	_, err = NewDistrubutedLog(tempDir, config)
	require.NoError(t, err)

}
