package log_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

// ErrOffsetOutOfRange represents and out of range error
type ErrOffsetOutOfRange struct {
	Offset uint64
}

// GRPCStatus return a grpc status error with details
func (e ErrOffsetOutOfRange) GRPCStatus() *status.Status {
	st := status.New(404, fmt.Sprintf("offset out of range: %d", e.Offset))
	msg := fmt.Sprintf("The requested offset is out of range: %d", e.Offset)
	d := &errdetails.LocalizedMessage{
		Locale:  "en-us",
		Message: msg,
	}

	std, err := st.WithDetails(d)
	if err != nil {
		return st
	}

	return std
}

// Error implements the Error interface
func (e ErrOffsetOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}
