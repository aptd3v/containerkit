package types

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// MergedConfig is the Docker SDK create config for a container.
type MergedConfig struct {
	Container *container.Config         // the container configuration
	Host      *container.HostConfig     // the host configuration
	Network   *network.NetworkingConfig // the network configuration
	Platform  *ocispec.Platform         // the platform configuration
}
