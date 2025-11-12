package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	// offWidth represents the number of bytes use to store the offset
	offWidth uint64 = 4
	// posWidth represents the number of bytes use to store the position
	posWidth uint64 = 8

	entWidth = offWidth + posWidth
)

// index: offset and postition in the store file

// Index represents the file we store index entries
type Index struct {
	file *os.File
	mmap gommap.MMap
	// size tell us the index and where to write the next entry
	size uint64
}

// Config represents configuration for the index file
type Config struct {
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
}

// NewIndex creates an Index for the given file
func NewIndex(f *os.File, c Config) (*Index, error) {
	idx := &Index{
		file: f,
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	idx.size = uint64(fi.Size())
	if err := os.Truncate(fi.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}

	return idx, nil
}

// Read takes an offset and return the associated record's position in the store
func (i *Index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}

	pos = uint64(out) * entWidth

	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}

	out = enc.Uint32(i.mmap[pos : pos+offWidth])

	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])

	return out, pos, nil

}

// Write appends the given offset and position to the index file
func (i *Index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}

	// encode the offset and write it to the memory-mapped file
	enc.PutUint32(i.mmap[i.size:offWidth], off)
	// encode the position and write it to the memory-mapped file
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

// Name return the index's file path
func (i *Index) Name() string {
	return i.file.Name()
}

// Close closes the index file. it syncs its data and persists the data to stable storage
func (i *Index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

	if err := i.file.Sync(); err != nil {
		return err
	}

	return i.file.Close()
}
