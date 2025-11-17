package log

import "sync"

type Log struct {
	mu            sync.Mutex
	Dir           string
	Config        Config
	ActiveSegment *segment
	Segments      []segment
}

func NewLog() *Log {
	return &Log{}
}
