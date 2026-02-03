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

type fsm struct {
	log *Log
}

type logStore struct {
	*Log
}

// NewDistrubutedLog create a new distributed log
func NewDistrubutedLog(dataDir string, config Config) (*DistributedLog, error) {
	l := &DistributedLog{
		config: config,
	}
	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}

	if err := l.setupRaft(dataDir); err != nil {
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

// setupRaft configures and creates the server's raft instance
func (l *DistributedLog) setupRaft(dataDir string) error {
	_ = &fsm{log: l.log}

	logDir := path.Join(dataDir, "raft", "log")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return err
	}

	logConfig := l.config
	logConfig.Segment.InitialOffset = 1
	_, err = newLogStore(logDir, logConfig)
	if err != nil {
		return err
	}
	return nil
}

func newLogStore(logDir string, config Config) (*logStore, error) {
	return nil, nil
}

func (l *DistributedLog) Create() error {
	return nil
}
