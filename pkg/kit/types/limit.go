package types

import (
	composetypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
)

// Ulimit is [container.Ulimit].
type Ulimit = container.Ulimit

// ConsoleSize is the Docker HostConfig console size (height, width).
type ConsoleSize [2]uint

// Resources is Compose [composetypes.Resource] (deploy limits / reservations).
type Resources = composetypes.Resource

// NanoCPUs is Compose [composetypes.NanoCPUs].
type NanoCPUs = composetypes.NanoCPUs
