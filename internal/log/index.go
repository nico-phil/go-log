package log

import (
	"fmt"
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	// offWidth represents the number of bytes use to store the offset
	offWidth uint64 = 4
	// posWidth represents the number of bytes use to store the position
	posWidth uint64 = 8

	entryWidth = offWidth + posWidth
)

// index: offset and postition in the store file

// Index represents the file we store index entries
type index struct {
	file *os.File
	MMap gommap.MMap
	// size tell us the index and where to write the next entry
	size uint64
}

// NewIndex creates an Index for the given file
func NewIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	idx.size = size

	if err := os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	idx.MMap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	)

	fmt.Println("file size", idx.size)
	fmt.Println("mmap len", len(idx.MMap))

	if err != nil {
		return nil, err
	}

	return idx, nil
}

// Write appends the given offset and position to the index file
func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.MMap)) < i.size+entryWidth {
		return io.EOF
	}

	// encode the offset and write it to the memory-mapped file
	enc.PutUint32(i.MMap[i.size:i.size+offWidth], off)
	// encode the position and write it to the memory-mapped file
	enc.PutUint64(i.MMap[i.size+offWidth:i.size+entryWidth], pos)
	i.size += uint64(entryWidth)

	if err := os.Truncate(i.file.Name(), int64(i.size)); err != nil {
		return err
	}
	return nil
}

// Read takes an index and return the associated record's position in the store
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	if in == -1 {
		out = uint32((i.size / entryWidth) - 1) // 24/12 = 1
	} else {
		out = uint32(in)
	}

	pos = uint64(out) * entryWidth

	if i.size < pos+entryWidth {
		return 0, 0, io.EOF
	}

	out = enc.Uint32(i.MMap[pos : pos+offWidth])

	pos = enc.Uint64(i.MMap[pos+offWidth : pos+entryWidth])

	fmt.Println("mmap:", string(i.MMap))

	return out, pos, nil

}

// Name returns the index's file path
func (i *index) Name() string {
	return i.file.Name()
}

// Close closes the index file. it syncs its data and persists the data to stable storage
func (i *index) Close() error {
	if err := i.MMap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

	if err := i.file.Sync(); err != nil {
		return err
	}

	if err := os.Truncate(i.file.Name(), int64(i.size)); err != nil {
		return err
	}

	return i.file.Close()
}
