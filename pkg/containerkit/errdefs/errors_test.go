package errdefs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrdefs(t *testing.T) {
	cc := NewContainerConfigError("image", "missing")
	require.True(t, IsContainerConfigError(cc))
	require.Equal(t, "image: missing", cc.Error())
	require.Equal(t, ErrContainerConfig, errors.Unwrap(cc))

	sc := NewServiceConfigError("web", "dup")
	require.True(t, IsServiceConfigError(sc))
	require.Equal(t, "web: dup", sc.Error())

	hc := NewHostConfigError("memory", "bad")
	require.True(t, IsHostConfigError(hc))

	nc := NewNetworkConfigError("net", "bad")
	require.True(t, IsNetworkConfigError(nc))

	pc := NewPlatformConfigError("os", "bad")
	require.True(t, IsPlatformConfigError(pc))

	pj := NewProjectConfigError("proj", "empty")
	require.True(t, IsProjectConfigError(pj))
	require.Equal(t, ErrProjectConfig, errors.Unwrap(pj))
}
