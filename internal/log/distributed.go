package log

import "github.com/hashicorp/raft"

type DistributedLog struct {
	config Config
	log    *Log
	raft   *raft.Raft
}

func NewDistrubutedLog(dataDir string, config Config) (*DistributedLog, error) {
	dl := &DistributedLog{
		config: config,
	}
	return dl, nil
}
