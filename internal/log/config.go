package log

import "github.com/hashicorp/raft"

// Config represents configuration for the the system
type Config struct {
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}

	Raft struct {
		raft.Config
		StreamLayer raft.StreamLayer
		Bootstrap   bool
	}
}
