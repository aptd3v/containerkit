package containerkit

import (
	"errors"
	"testing"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestNewContainerAndValidate(t *testing.T) {
	c := NewContainer()
	require.Equal(t, "", c.Name)
	require.NoError(t, c.Validate())

	c.Errors = nil
	require.NoError(t, c.Validate())

	c.Errors = []error{errors.New("x")}
	require.Error(t, c.Validate())
}

func TestWithTypedSettersAndInvalidType(t *testing.T) {
	c := NewContainer("w")
	c.With(
		nil,
		SetContainerConfig(nil),
		SetHostConfig(nil),
		SetNetworkConfig(nil),
		SetPlatformConfig(nil),
		SetContainerConfig(func(cfg *container.Config) error {
			cfg.Image = "alpine"
			return nil
		}),
		SetHostConfig(func(cfg *container.HostConfig) error {
			cfg.Privileged = true
			return nil
		}),
		SetNetworkConfig(func(cfg *network.NetworkingConfig) error {
			cfg.EndpointsConfig = map[string]*network.EndpointSettings{"n": {}}
			return nil
		}),
		SetPlatformConfig(func(p *ocispec.Platform) error {
			p.OS = "linux"
			return nil
		}),
	)
	require.NoError(t, c.Validate())
	require.Equal(t, "alpine", c.Config.Container.Image)
	require.True(t, c.Config.Host.Privileged)
	require.Equal(t, "linux", c.Config.Platform.OS)

	c.With(
		SetContainerConfig(func(*container.Config) error { return errors.New("cc") }),
		SetHostConfig(func(*container.HostConfig) error { return errors.New("hc") }),
		SetNetworkConfig(func(*network.NetworkingConfig) error { return errors.New("nc") }),
		SetPlatformConfig(func(*ocispec.Platform) error { return errors.New("pc") }),
		"not-a-setter",
	)
	require.Error(t, c.Validate())
}

func TestWithXxxConfigErrorWrappers(t *testing.T) {
	c := NewContainer("e")
	c.WithContainerConfig(func(*container.Config) error { return errors.New("bad-cc") })
	c.WithHostConfig(func(*container.HostConfig) error { return errors.New("bad-hc") })
	c.WithNetworkConfig(func(*network.NetworkingConfig) error { return errors.New("bad-nc") })
	c.WithPlatformConfig(func(*ocispec.Platform) error { return errors.New("bad-pc") })
	err := c.Validate()
	require.Error(t, err)
	require.True(t, errdefs.IsContainerConfigError(c.Errors[0]))
	require.True(t, errdefs.IsHostConfigError(c.Errors[1]))
	require.True(t, errdefs.IsNetworkConfigError(c.Errors[2]))
	require.True(t, errdefs.IsPlatformConfigError(c.Errors[3]))
}
