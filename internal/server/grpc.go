package server

import (
	api "github.com/nico-phil/go-log/api/v1"
)

type Config struct {
	CommitLog CommitLog
}

type grpcServer struct {
	*Config
	api.UnimplementedLogServer
}
