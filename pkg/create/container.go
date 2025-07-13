// Package create provides the container config for the wrapped client.
// it is built to be created in a declarative way.
package create

import (
	"errors"
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type MergedConfig struct {
	Container *container.Config         // the container configuration
	Host      *container.HostConfig     // the host configuration
	Network   *network.NetworkingConfig // the network configuration
	Platform  *ocispec.Platform         // the platform configuration
}
type Container struct {
	Name   string        // the name of the container
	Config *MergedConfig // merged docker sdk configuration for container creation
	Errors []error
}

func NewContainer(name string) *Container {
	return &Container{
		Name: name,
		Config: &MergedConfig{
			Container: &container.Config{},
			Host:      &container.HostConfig{},
			Network:   &network.NetworkingConfig{},
			Platform:  &ocispec.Platform{},
		},
		Errors: []error{},
	}
}

// Validate validates the container config.
// It will return an error if the container has errors.
func (c *Container) Validate() error {
	if c.Errors == nil {
		c.Errors = []error{}
	}
	if c == nil {
		return fmt.Errorf("container is nil")
	}
	if len(c.Errors) > 0 {
		return fmt.Errorf("container has the following errors:\n%s", errors.Join(c.Errors...))
	}
	return nil
}

// SetContainerConfig is a function that sets the container config
type SetContainerConfig func(config *container.Config) error

// SetHostConfig is a function that sets the host config
type SetHostConfig func(config *container.HostConfig) error

// SetNetworkConfig is a function that sets the network config
type SetNetworkConfig func(config *network.NetworkingConfig) error

// SetPlatformConfig is a function that sets the platform config
type SetPlatformConfig func(config *ocispec.Platform) error

// WithContainerConfig sets the container config via the setter functions.
// It will return a container with the container config set.
// If any of the setters return an error, that setter will
// be skipped and an error will be appended to the container's error slice.
// parameters:
//   - setters: the setters to set the container config
func (c *Container) WithContainerConfig(setters ...SetContainerConfig) *Container {
	if len(setters) == 0 {
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(c.Config.Container); err != nil {
			c.Errors = append(c.Errors, errdefs.NewContainerConfigError("container", err.Error()))
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
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(c.Config.Host); err != nil {
			c.Errors = append(c.Errors, errdefs.NewHostConfigError("host", err.Error()))
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
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(c.Config.Network); err != nil {
			c.Errors = append(c.Errors, errdefs.NewNetworkConfigError("network", err.Error()))
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
		return c
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(c.Config.Platform); err != nil {
			c.Errors = append(c.Errors, errdefs.NewPlatformConfigError("platform", err.Error()))
			continue
		}
	}
	return c
}
