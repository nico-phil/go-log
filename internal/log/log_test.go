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
		"append and read success":    testAppendRead,
		"offset out of range":        testOutOfRangeErr,
		"init with existing segment": testInitExisting,
		"truncate":                   testTruncate,
	}

	for sc, fn := range senarios {
		t.Run(sc, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := Config{}
			c.Segment.MaxStoreBytes = 36

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

	read, err := l.Read(off)
	require.NoError(t, err)
	require.Equal(t, rec.Value, read.Value)
}

func testOutOfRangeErr(t *testing.T, l *Log) {
	read, err := l.Read(1)
	require.Error(t, err)
	require.Nil(t, read)
}

func testInitExisting(t *testing.T, l *Log) {
	rec := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := l.Append(rec)
		require.NoError(t, err)
	}
	require.NoError(t, l.Close())

	off, err := l.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = l.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

	newL, err := NewLog(l.Dir, l.Config)
	require.NoError(t, err)

	off, err = newL.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = l.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

}

func testTruncate(t *testing.T, l *Log) {
	append := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := l.Append(append)
		require.NoError(t, err)
	}

	err := l.Truncate(1)
	require.NoError(t, err)

	_, err = l.Read(0)
	require.Error(t, err)
}
