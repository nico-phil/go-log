package log

import (
	"os"
	"path"

	"github.com/hashicorp/raft"
)

// DistributedLog repesent a wrapper around raft to replicate the log
type DistributedLog struct {
	config Config
	log    *Log
	raft   *raft.Raft
}

// NewDistrubutedLog create a new distributed log
func NewDistrubutedLog(dataDir string, config Config) (*DistributedLog, error) {
	l := &DistributedLog{
		config: config,
	}
	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *DistributedLog) setupLog(dataDir string) error {
	logDir := path.Join(dataDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	var err error
	l.log, err = NewLog(logDir, l.config)

	return err
}
