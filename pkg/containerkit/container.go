// Package containerkit provides the container config for the wrapped client.
// it is built to be created in a declarative way.
package containerkit

import (
	"errors"
	"fmt"
	"strings"

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
	Name       string        // the name of the container
	Config     *MergedConfig // merged docker sdk configuration for container creation
	Errors     []error
	mountNotes map[string]mountNote
}

// mountNote carries Compose mount fields the engine mount type does not store.
type mountNote struct {
	selinux        string
	createHostPath *bool
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
		return fmt.Errorf("container config has the following errors:\n%w", errors.Join(c.Errors...))
	}
	return nil
}
