package kit

import (
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/aptd3v/containerkit/pkg/kit/types"
	"github.com/stretchr/testify/require"
)

func TestNetworkContainerApplies(t *testing.T) {
	c := NewContainer("n").NetworkMode(types.NetworkContainer("123"))
	require.Empty(t, c.Errors)
	require.Equal(t, "container:123", string(c.Config.Host.NetworkMode))
}

func TestNetworkContainerEmptyIsHostConfigError(t *testing.T) {
	c := NewContainer("n").NetworkMode(types.NetworkContainer(""))
	require.Len(t, c.Errors, 1)
	require.True(t, errdefs.IsHostConfigError(c.Errors[0]))
}

func mustCpuset(list string) types.Cpuset {
	set, err := types.ParseCpuset(list)
	if err != nil {
		panic(err)
	}
	return set
}
