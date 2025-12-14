package server

import (
	api "github.com/nico-phil/go-log/api/v1"
	CommitLog "github.com/nico-phil/go-log/internal/log"
)

// Config contents the commit log package
type Config struct {
	CommitLog CommitLog.Log
}

// grpcServer represent our grpc server
type grpcServer struct {
	*Config
	api.UnimplementedLogServer
}

// newgrpcServer create a new grpc server
func newgrpcServer(config *Config) (grpcServer, error) {
	svr := grpcServer{
		Config: config,
	}

	return svr, nil
}
