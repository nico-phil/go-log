package log

import (
	"os"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/stretchr/testify/require"
)

// TestLog tests the Log
func TestLog(t *testing.T) {

	senarios := map[string]func(t *testing.T, log *Log){
		"append and read success": testAppendRead,
	}

	for sc, fn := range senarios {
		t.Run(sc, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := Config{}
			c.Segment.MaxStoreBytes = 32

			log, err := NewLog(dir, c)
			require.NoError(t, err)

			fn(t, log)
		})
	}

}

// testApendRead represents an helper function to test append and read
func testAppendRead(t *testing.T, l *Log) {
	// t.Helper()
	rec := &api.Record{
		Value: []byte("hello world"),
	}

	off, err := l.Append(rec)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)
}
