package log

import (
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// @Todo: later, i should unexport some filed for better api design. i should not expose to users some
// internal stuff

// Log represents a Log, this is api that control the WAL
type Log struct {
	mu            sync.Mutex
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
	return &Log{
		Dir:    dir,
		Config: c,
	}, nil
}

func (l *Log) setup() error {
	files, err := os.ReadDir(l.Dir)
	if err != nil {
		return err
	}
	var baseOffsets []uint64
	for _, file := range files {
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
