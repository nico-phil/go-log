package loadbalance_test

import (
	"testing"

	"github.com/nico-phil/go-log/internal/loadbalance"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/balancer"
)

func TestPickerNoSubConnAvailable(t *testing.T) {
	picker := &loadbalance.Picker{}

	for _, method := range []string{
		"/log.vX.Log/Produce",
		"/log.vX.Log/Consume",
	} {
		info := balancer.PickInfo{
			FullMethodName: method,
		}

		result, err := picker.Pick(info)
		require.Equal(t, balancer.ErrNoSubConnAvailable, err)

		require.Nil(t, result.SubConn)
	}
}
