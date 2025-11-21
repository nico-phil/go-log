package log

import (
	"fmt"
	"log"
	"os"
	"path"

	api "github.com/nico-phil/go-log/api/v1"
	"google.golang.org/protobuf/proto"
)

// Sement ties a store and an index together
type segment struct {
	store      *store
	index      *index
	baseOffset uint64
	nextOffset uint64
	config     Config
}

// newSegment create a new segment when the current active segment hits max size
func NewSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}

	var err error
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")), //data/0.store
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)

	if err != nil {
		return nil, err
	}

	if s.store, err = NewStore(storeFile); err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)

	if err != nil {
		return nil, err
	}

	if s.index, err = NewIndex(indexFile, c); err != nil {
		return nil, err
	}

	off, _, err := s.index.Read(-1)
	if err != nil {
		log.Printf("read(-1) %v", err)
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}

	return s, nil
}

// Append appends the record to the store and the index file
func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur

	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}

	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}

	if err := s.index.Write(
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		fmt.Println("error here", err)
		return 0, err
	}

	s.nextOffset++

	return cur, nil
}

// Read returns the record for a given offset
func (s *segment) Read(off uint64) (*api.Record, error) {
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}

	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}

	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// IsMaxed returns whether the segment has reached its max size, either by writing too much to the store or the index
func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes ||
		s.index.size >= s.config.Segment.MaxIndexBytes
}

// Remove deletes files(index and store) for ossociated with the segemnt
func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}

	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}

	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}

	return nil
}

// Close closes both the store and index files associated with the segment
func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}

	if err := s.store.Close(); err != nil {
		return err
	}

	return nil
}

func nearestMultiple(j, k int64) int64 {
	if j >= 0 {
		return (j / k) * k
	}
	return ((j - k + 1) / k) * k

}

func (s *segment) ReadIndex(in int64) (out uint32, pos uint64, err error) {
	return s.index.Read(in)
}
