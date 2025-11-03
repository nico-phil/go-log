package log

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

var (
	// enc represents BinEndian order
	enc = binary.BigEndian
)

const (
	// lenWidth Defines the number of bytes used to store the record's length
	lenWidth = 8
)

// store represents the file we store records in
type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

// NewStore creates a store instance
func NewStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append store record to the buffer to minimize system call, later we call flush to persist the record into the file
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size

	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)
	s.buf.Flush()
	return uint64(w), pos, nil
}

// Read read a record from the file stored a given position, it flush the buffer before in case we want to read a record that has not flused yet
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	b := make([]byte, enc.Uint64(size))
	fmt.Println("b", b)
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}

// ReadAt read record from the file begining a the offset
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}

	return s.File.ReadAt(p, off)
}

// Close closes the file, it fluses the buffer before closing the file, in case there are records in the buffer
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.buf.Flush()
	if err != nil {
		return err
	}

	return s.File.Close()
}
