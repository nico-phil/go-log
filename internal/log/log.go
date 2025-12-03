package log

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	api "github.com/nico-phil/go-log/api/v1"
)

// @Todo: later, i should unexport some filed for better api design. i should not expose to users some
// internal stuff

// Log represents the abstraction to manage the list of segments
type Log struct {
	mu            sync.RWMutex
	Dir           string
	Config        Config
	ActiveSegment *segment
	Segments      []*segment
}

// NewLog create a new Log
func NewLog(dir string, c Config) (*Log, error) {
	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}

	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}

	l := Log{Dir: dir, Config: c}
	return &l, l.setup()
}

// setup, reads all the files in the log dir, setup segment for each of them
func (l *Log) setup() error {
	files, err := os.ReadDir(l.Dir)
	if err != nil {
		return err
	}
	var baseOffsets []uint64
	for _, file := range files { // file: 16.index 16.store
		offStr := strings.TrimSuffix(
			file.Name(),
			path.Ext(file.Name()),
		)

		off, _ := strconv.ParseUint(offStr, 10, 10)
		baseOffsets = append(baseOffsets, off)
	}

	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})

	for i := 0; i < len(baseOffsets); i++ {
		if err := l.newSegment(baseOffsets[i]); err != nil {
			return err
		}
		// baseOffset contains dup for index and store so we skip
		// the dup
		i++
	}

	if l.Segments == nil {
		err = l.newSegment(
			l.Config.Segment.InitialOffset,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// newSegment creates a new segment and adds it to the segment list
func (l *Log) newSegment(off uint64) error {
	s, err := NewSegment(l.Dir, off, l.Config)
	if err != nil {
		return err
	}

	l.Segments = append(l.Segments, s)
	l.ActiveSegment = s
	return nil
}

// Append append a log record to the system, it return the offset and and error
func (l *Log) Append(record *api.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	off, err := l.ActiveSegment.Append(record)
	if err != nil {
		return 0, err
	}

	if l.ActiveSegment.IsMaxed() {
		if err = l.newSegment(off + 1); err != nil {
			return 0, err
		}
	}

	return off, nil
}

// Read takes an offet and return a record and error
func (l *Log) Read(off uint64) (*api.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var s *segment
	for _, segment := range l.Segments {
		if segment.baseOffset <= off && off < segment.nextOffset {
			s = segment
			break
		}
	}
	if s == nil || s.nextOffset <= off {
		return nil, fmt.Errorf("offset out of range: %d", off)
	}
	return s.Read(off)
}

// Close closes all the segments
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, segment := range l.Segments {
		if err := segment.Close(); err != nil {
			return err
		}
	}
	return nil
}
