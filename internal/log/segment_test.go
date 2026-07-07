package log

import (
	"io"
	"log"
	"os"
	"path"
	"testing"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/stretchr/testify/require"
)

func TestSegment(t *testing.T) {
	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entryWidth * 3 //36 bytes

	var dir = path.Join("./", "test-proglog")
	var err error
	if permamentFile {
		_, err := os.Stat(dir)
		if err != nil {
			log.Println("stat is no nil, the file is already exist:****")
			err = os.Mkdir(dir, os.ModePerm)
			require.NoError(t, err)

			err = os.Chmod(dir, os.ModePerm)
			require.NoError(t, err)
		}

	} else {
		dir, err = os.MkdirTemp("", dir)
		require.NoError(t, err)
	}

	defer func() {
		if !permamentFile {
			os.RemoveAll(dir)
		}

	}()

	want := api.Record{Value: []byte("hello world")}
	baseOffset := uint64(0)
	s, err := NewSegment(dir, baseOffset, c)
	require.NoError(t, err)
	require.Equal(t, baseOffset, s.nextOffset)
	require.False(t, s.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(&want)
		require.NoError(t, err)
		require.Equal(t, baseOffset+i, off)

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

	s, err = NewSegment(dir, baseOffset, c)
	require.NoError(t, err)
	require.True(t, s.IsMaxed())

	err = s.Remove()
	require.NoError(t, err)
	s, err = NewSegment(dir, baseOffset, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())

}
