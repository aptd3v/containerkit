// Package create provides the container config for the wrapped client.
// it is built to be created in a declarative way.
package create

import (
	"errors"
	"fmt"
	"strings"

	c "github.com/docker/docker/api/types/container"
	n "github.com/docker/docker/api/types/network"
	p "github.com/opencontainers/image-spec/specs-go/v1"
)

type ContainerConfig struct {
	*c.Config
}
type HostConfig struct {
	*c.HostConfig
}
type NetworkConfig struct {
	*n.NetworkingConfig
}
type PlatformConfig struct {
	*p.Platform
}

type MergedConfig struct {
	Name      string // the name of the container
	Container *ContainerConfig
	Host      *HostConfig
	Network   *NetworkConfig
	Platform  *PlatformConfig
}
type Container struct {
	Warnings []string
	Errors   []error
	Config   *MergedConfig
}

func NewContainer(name string) *Container {
	return &Container{
		Config: &MergedConfig{
			Name: name,
			Container: &ContainerConfig{
				Config: &c.Config{},
			},
			Host: &HostConfig{
				HostConfig: &c.HostConfig{},
			},
			Network: &NetworkConfig{
				NetworkingConfig: &n.NetworkingConfig{},
			},
			Platform: &PlatformConfig{
				Platform: &p.Platform{},
			},
		},
	}
}

// Validate validates the container config.
// It will return an error if the container has errors.
func (c *Container) Validate() error {
	if len(c.Errors) > 0 {
		return fmt.Errorf("container has the following errors:\n%s", errors.Join(c.Errors...))
	}
	return nil
}
func (c *Container) GetWarnings() string {
	return strings.Join(c.Warnings, "\n")
}

// SetContainerConfig is a function that sets the container config
type SetContainerConfig func(config *ContainerConfig) error

// SetHostConfig is a function that sets the host config
type SetHostConfig func(config *HostConfig) error

// SetNetworkConfig is a function that sets the network config
type SetNetworkConfig func(config *NetworkConfig) error

// SetPlatformConfig is a function that sets the platform config
type SetPlatformConfig func(config *PlatformConfig) error

// WithContainerConfig sets the container config via the setter functions.
// It will return a container with the container config set.
// If any of the setters return an error, that setter will
// be skipped and an error will be appended to the container's error slice.
// parameters:
//   - setters: the setters to set the container config
func (c *Container) WithContainerConfig(setters ...SetContainerConfig) *Container {
	if len(setters) == 0 {
		c.Warnings = append(c.Warnings, "WithContainerConfig: function called without any setters")
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			c.Warnings = append(c.Warnings, "WithContainerConfig: setter is nil")
			continue
		}
		if err := setter(c.Config.Container); err != nil {
			c.Errors = append(c.Errors, NewContainerConfigError("container", err.Error()))
			continue
		}
	}

	return c
}

// WithHostConfig sets the host config via the setter functions.
// It will return a container with the host config set.
// If any of the setters return an error, that setter will
// be skipped and an error will be appended to the container's error slice.
// parameters:
//   - setters: the setters to set the host config
func (c *Container) WithHostConfig(setters ...SetHostConfig) *Container {
	if len(setters) == 0 {
		c.Warnings = append(c.Warnings, "WithHostConfig: function called without any setters")
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			c.Warnings = append(c.Warnings, "WithHostConfig: setter is nil")
			continue
		}
		if err := setter(c.Config.Host); err != nil {
			c.Errors = append(c.Errors, NewHostConfigError("host", err.Error()))
			continue
		}
	}
	return c
}

// WithNetworkConfig sets the network config via the setter functions.
// It will return a container with the network config set.
// If any of the setters return an error, that setter will
// be skipped and an error will be appended to the container's error slice.
// parameters:
//   - setters: the setters to set the network config
func (c *Container) WithNetworkConfig(setters ...SetNetworkConfig) *Container {
	if len(setters) == 0 {
		c.Warnings = append(c.Warnings, "WithNetworkConfig: function called without any setters")
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			c.Warnings = append(c.Warnings, "WithNetworkConfig: setter is nil")
			continue
		}
		if err := setter(c.Config.Network); err != nil {
			c.Errors = append(c.Errors, NewNetworkConfigError("network", err.Error()))
			continue
		}
	}
	return c
}

// WithPlatformConfig sets the platform config via the setter functions.
// It will return a container with the platform config set.
// If any of the setters return an error, that setter will
// be skipped and an error will be appended to the container's error slice.
// parameters:
//   - setters: the setters to set the platform config
func (c *Container) WithPlatformConfig(setters ...SetPlatformConfig) *Container {
	if len(setters) == 0 {
		c.Warnings = append(c.Warnings, "WithPlatformConfig: function called without any setters")
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			c.Warnings = append(c.Warnings, "WithPlatformConfig: setter is nil")
			continue
		}
		if err := setter(c.Config.Platform); err != nil {
			c.Errors = append(c.Errors, NewPlatformConfigError("platform", err.Error()))
			continue
		}
	}
	return c
}
