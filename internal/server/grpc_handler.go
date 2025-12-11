package server

import (
	"context"

	api "github.com/nico-phil/go-log/api/v1"
)

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	// validate input
	// pass it down
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}

	return &api.ProduceResponse{Offset: int64(offset)}, nil
}
