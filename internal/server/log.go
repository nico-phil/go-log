package server

import (
	"errors"
	"sync"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Log struct {
	mu      sync.Mutex
	Records []Record
}

func NewLog() *Log {
	return &Log{
		Records: make([]Record, 0),
	}
}

type Record struct {
	Offset uint64 `json:"offset"`
	Value  []byte `json:"value"`
}

func (l *Log) Append(record Record) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	record.Offset = uint64(len(l.Records))
	l.Records = append(l.Records, record)

	return len(l.Records), nil
}

func (l *Log) Read(offset uint64) (Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if int(offset) < 0 {
		return Record{}, ErrRecordNotFound
	}

	if int(offset) > len(l.Records) {
		return Record{}, ErrRecordNotFound
	}

	r := l.Records[offset]

	return r, nil
}
