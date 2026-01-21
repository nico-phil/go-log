package discovery_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemberShip(t *testing.T) {
	m, handler := setupMember(t, nil)
	m, _ = setupMember(t, m)
	m, _ = setupMember(t, m)

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
			3 == len(m[0].Members()) &&
			0 == len(handler.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.NoError(t, m[2].Leave())

	require.Eventually(t, func() bool {
		return 2 == len(handler.joins) &&
		3 == len(m[0].Member()) &&
	}, 3 * time.Second, 250 * time.Millisecond)

}

func setupMember(t *testing.T, members []*Membership) ([]*Membership, *Handler) {
	return nil, nil
}
