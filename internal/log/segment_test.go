package log

import (
	"io"
	"os"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/stretchr/testify/require"
)

func TestSegment(t *testing.T) {
	dir, err := os.MkdirTemp("", "segment")
	require.NoError(t, err)

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entryWidth * 3 //36 bytes

	want := api.Record{Value: []byte("hello world")}

	s, err := NewSegment(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset)
	require.False(t, s.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(&want)
		require.NoError(t, err)
		require.Equal(t, 16+i, off)

		got, err := s.Read(off)
		require.NoError(t, err)
		require.Equal(t, want.Value, got.Value)
	}

	_, err = s.Append(&want)
	require.Equal(t, io.EOF, err)

	//maxed index
	require.True(t, s.IsMaxed())

	c.Segment.MaxStoreBytes = uint64(len(want.Value) * 3)
	c.Segment.MaxIndexBytes = 1024

	s, err = NewSegment(dir, 16, c)
	require.NoError(t, err)

	require.True(t, s.IsMaxed())

	s.Remove()

	s, err = NewSegment(dir, 16, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())

}
