package types

import (
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
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
	}
}

func TestModeConstructors(t *testing.T) {
	for _, err := range []error{
		ValidateNetworkMode(NetworkContainer("")),
		ValidateIpcMode(IpcContainer("")),
		ValidatePidMode(PidContainer("")),
		ValidateNetworkMode(NetworkContainer("   ")),
	} {
		require.Error(t, err)
		require.ErrorIs(t, err, errEmptyContainerName)
		require.False(t, errdefs.IsHostConfigError(err))
	}

	require.NoError(t, ValidateNetworkMode(""))
	require.NoError(t, ValidateNetworkMode(NetworkBridge))
	net := NetworkContainer("123")
	require.NoError(t, ValidateNetworkMode(net))
	require.Equal(t, "container:123", string(net))
}
