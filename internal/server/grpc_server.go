package server

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	api "github.com/nico-phil/go-log/api/v1"
	llog "github.com/nico-phil/go-log/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	objectwildCard = "*"
	produceAction  = "produce"
	consume        = "consume"
)

// Config contents the commit log package
type Config struct {
	CommitLog  *llog.Log
	Authorizer Authorizer
}

// CommitLog represents the interface of the log
type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

type Authorizer interface {
	Authorize(subject, object, action string) error
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
	opts = append(opts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(authenticate),
			)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_auth.UnaryServerInterceptor(authenticate),
		)),
	)

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

// authenticate adds the client subject in the context
func authenticate(ctx context.Context) (context.Context, error) {
	peerCtx, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, status.New(
			codes.Unknown,
			"couldn't find peer info",
		).Err()
	}

	if peerCtx.AuthInfo == nil {
		return context.WithValue(ctx, subjectContextKey{}, ""), nil
	}

	tlsInfo := peerCtx.AuthInfo.(credentials.TLSInfo)
	subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName

	ctx = context.WithValue(ctx, subjectContextKey{}, subject)

	return ctx, nil
}
