package log

import (
	"os"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/stretchr/testify/require"
)

// TestLog tests the Log
func TestLog(t *testing.T) {
	dir, err := os.MkdirTemp("", "log-test")
	require.NoError(t, err)
	require.NotEmpty(t, dir)

	c := Config{}
	// c.Segment.MaxStoreBytes = 32

	l, err := NewLog(dir, c)
	require.NoError(t, err)

	testAppendRead(t, l)

}

// testApendRead represents an helper function to test append and read
func testAppendRead(t *testing.T, l *Log) {
	t.Helper()
	rec := &api.Record{
		Value: []byte("hello world"),
	}

	off, err := l.Append(rec)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)
}
