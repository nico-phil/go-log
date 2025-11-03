package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	write = []byte("hello world")
	width = uint64(len(write)) + lenWidth
)

// TestStoreAppenRead tests Append and Read method
func TestStoreAppenRead(t *testing.T) {
	f, err := os.CreateTemp("", "store")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s, err := NewStore(f)
	require.NoError(t, err)

	testAppend(t, s)
	testRead(t, s)
	testReatAt(t, s)

	s, err = NewStore(f)
	require.NoError(t, err)

	testRead(t, s)
}

// TestStoreClose tests close store method
func TestStoreClose(t *testing.T) {
	f, err := os.CreateTemp("", "store")
	require.NoError(t, err)

	s, err := NewStore(f)
	require.NoError(t, err)

	_, _, err = s.Append([]byte(write))
	require.NoError(t, err)

	_, beforeSize, err := openFile(f.Name())
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, afterSize, err := openFile(f.Name())
	require.NoError(t, err)

	require.True(t, afterSize > beforeSize)
}

// openFile represents a helper function to open file
func openFile(name string) (file *os.File, size int64, err error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil
}

// testReatAt tests the ReatAt method
func testReatAt(t *testing.T, s *store) {
	t.Helper()

	for i, off := uint64(1), int64(0); i < 4; i++ {
		b := make([]byte, lenWidth)
		n, err := s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, lenWidth, n)
		off += int64(n)

		size := enc.Uint64(b)
		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, write, b)
		require.Equal(t, int(size), n)
		off += int64(n)
	}

}

// testAppend represents a helper function to test Append
func testAppend(t *testing.T, s *store) {
	t.Helper()
	for i := uint64(1); i < 4; i++ {
		n, pos, err := s.Append(write)
		require.NoError(t, err)
		require.Equal(t, pos+n, width*i)
	}
}

// testRead represents a helper function to test Read
func testRead(t *testing.T, s *store) {
	t.Helper()
	var pos uint64
	for i := uint64(1); i < 4; i++ {
		read, err := s.Read(pos)
		require.NoError(t, err)
		require.Equal(t, write, read)
		pos += width
	}
}
