package server

import (
	"context"
	"log"

	api "github.com/nico-phil/go-log/api/v1"
)

// Produce represents the grcp handler to append into the log
func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	// validate input
	// pass it down
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}

	return &api.ProduceResponse{Offset: int64(offset)}, nil
}

// Consume represents the grpc handler to read from the log
func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	rec, err := s.CommitLog.Read(uint64(req.Offset))
	if err != nil {
		log.Println("consume:", err)
		return nil, err
	}

	return &api.ConsumeResponse{Record: rec}, nil
}

// ProduceStream represents a grpc handler that append stream of records to the log
func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

// ConsumeStream consume a stream of record from the log
func (s *grpcServer) ConsumeStream(req *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}

			if err := stream.Send(res); err != nil {
				return nil
			}

			req.Offset++
		}
	}
}
