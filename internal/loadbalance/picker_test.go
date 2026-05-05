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

func TestPickerProduceToLeader(t *testing.T) {
	picker, subConns := setupTest()

	info := balancer.PickInfo{
		FullMethodName: "/log.vX.Log/Produce",
	}

	for i := 0; i < 5; i++ {
		gotPick, err := picker.Pick(info)
		require.NoError(t, err)
		require.Equal(t, subConns[0], gotPick.SubConn)
	}

}

func setupTest() (*loadbalance.Picker, []*balancer.SubConn) {
	return nil, nil
}
