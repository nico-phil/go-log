package server

import (
	api "github.com/nico-phil/go-log/api/v1"
	CommitLog "github.com/nico-phil/go-log/internal/log"
)

type Config struct {
	CommitLog CommitLog.Log
}

type grpcServer struct {
	*Config
	api.UnimplementedLogServer
}

func newgrpcServer(config *Config) (grpcServer, error) {
	svr := grpcServer{
		Config: config,
	}

	return svr, nil
}
