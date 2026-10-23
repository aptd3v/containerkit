package containerkit

import (
	"testing"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/stretchr/testify/require"
)

func TestParseByteSize(t *testing.T) {
	n, err := ParseByteSize("100M")
	require.NoError(t, err)
	require.Equal(t, 100*MiB, n)

	n, err = ParseByteSize("100MiB")
	require.NoError(t, err)
	require.Equal(t, 100*MiB, n)

	_, err = ParseByteSize("100Error")
	require.Error(t, err)
	_, err = ParseByteSize("not-a-size")
	require.Error(t, err)
	_, err = ParseByteSize("44S")
	require.Error(t, err)
}

func TestParseCpuset(t *testing.T) {
	set, err := ParseCpuset("0,1,3-5")
	require.NoError(t, err)
	require.Equal(t, "0,1,3-5", set.String())

	for _, bad := range []string{"", "error", "error-error", "error-error-error", "1,1,,-,,"} {
		_, err = ParseCpuset(bad)
		require.Error(t, err, bad)
		require.True(t, errdefs.IsHostConfigError(err), bad)
	}
}

func TestModeConstructors(t *testing.T) {
	for _, err := range []error{
		second(NetworkContainer("")),
		second(IpcContainer("")),
		second(PidContainer("")),
		second(UserNSContainer("")),
	} {
		require.Error(t, err)
		require.True(t, errdefs.IsHostConfigError(err))
	}

	net, err := NetworkContainer("123")
	require.NoError(t, err)
	c := NewContainer("n").NetworkMode(net)
	require.Empty(t, c.Errors)
	require.Equal(t, "container:123", string(c.Config.Host.NetworkMode))
}

func second[T any](v T, err error) error { return err }

func mustCpuset(list string) Cpuset {
	set, err := ParseCpuset(list)
	if err != nil {
		panic(err)
	}
	return set
}
