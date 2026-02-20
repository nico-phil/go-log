package log_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDistributedLog test the distributedLog
func TestDistributedLog(t *testing.T) {
	nodeCount := 3
	for i := 0; i < nodeCount; i++ {
		dataDir, err := os.MkdirTemp("", "distributed-log-test")
		require.NoError(t, err)
		defer func(dir string) {
			_ = os.Remove(dir)
		}(dataDir)
	}

}
