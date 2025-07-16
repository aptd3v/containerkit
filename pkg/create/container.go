// Package create provides the container config for the wrapped client.
// it is built to be created in a declarative way.
package create

import (
	"errors"
	"fmt"
	"strings"

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

// NewContainer creates a new container with the given name.
// It will return a container with empty configurations ready to be set.
// a container is a wrapper around the docker sdk container configuration.
// but it can be used as a service config in a compose project
//
// parameters:
//   - name: the name of the container
//
// Note: 'name' is optional when using this container as a Compose service config.
// It is required when using the Docker SDK directly.
// If multiple strings are passed as the name, they will be joined with hyphens (e.g., "foo", "bar" -> "foo-bar").
func NewContainer(name ...string) *Container {
	cName := ""
	if len(name) > 0 {
		cName = strings.Join(name, "-")
	}
	return &Container{
		Name: cName,
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
		return fmt.Errorf("container config has the following errors:\n%s", errors.Join(c.Errors...))
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

// With sets the container config via the setter functions.
// It will return a container with the container config set.
// If any of the setters return an error, the error will
// be appended to the container's error slice.
// parameters:
//   - setters: the setters to set the container config
//
// note: setters can be any type that implements the
// SetContainerConfig, SetHostConfig, SetNetworkConfig,
// or SetPlatformConfig function type. If a setter is
// nil, it will be skipped.
func (c *Container) With(setters ...any) *Container {
	var errs []error
	for i, setter := range setters {
		switch setter := any(setter).(type) {
		case nil:
			continue
		case SetContainerConfig:
			if setter == nil {
				continue
			}
			if err := setter(c.Config.Container); err != nil {
				errs = append(errs, fmt.Errorf("setter %d: %w", i, err))
			}
		case SetHostConfig:
			if setter == nil {
				continue
			}
			if err := setter(c.Config.Host); err != nil {
				errs = append(errs, fmt.Errorf("setter %d: %w", i, err))
			}
		case SetNetworkConfig:
			if setter == nil {
				continue
			}
			if err := setter(c.Config.Network); err != nil {
				errs = append(errs, fmt.Errorf("setter %d: %w", i, err))
			}
		case SetPlatformConfig:
			if setter == nil {
				continue
			}
			if err := setter(c.Config.Platform); err != nil {
				errs = append(errs, fmt.Errorf("setter %d: %w", i, err))
			}
		default:
			errs = append(errs, fmt.Errorf("setter %d: invalid setter type: %T", i, setter))
		}

	}
	c.Errors = append(c.Errors, errs...)
	return c
}

// WithContainerConfig sets the container config via the setter functions.
// It will return a container with the container config set.
// If any of the setters return an error, the error will be
// appended to the container's error slice.
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
// If any of the setters return an error, the error will be
// appended to the container's error slice.
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
// If any of the setters return an error, the error will be
// appended to the container's error slice.
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
// If any of the setters return an error, the error will be
// appended to the container's error slice.
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
