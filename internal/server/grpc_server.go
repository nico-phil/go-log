package server

import (
	"context"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/auth"
	llog "github.com/nico-phil/go-log/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	objectwildCard = "*"
	produceAction  = "produce"
	consume        = "consume"
)

// Config contents the commit log package
type Config struct {
	CommitLog  *llog.Log
	Authorizer *auth.Authorizer
}

// CommitLog represents the interface of the log
type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

// grpcServer represent our grpc server
type grpcServer struct {
	*Config
	api.UnimplementedLogServer
}

// newgrpcServer create a new grpcServer
func newgrpcServer(config *Config) (*grpcServer, error) {
	svr := grpcServer{
		Config: config,
	}

	return &svr, nil
}

// NewGRPCServer creates a Newserver and and register it
func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterLogServer(gsrv, srv)
	reflection.Register(gsrv)
	return gsrv, nil
}

type subjectContextKey struct{}

func subjectGetContext(ctx context.Context) string {
	return ctx.Value(subjectContextKey{}).(string)
}
