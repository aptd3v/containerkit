// Package netmode is the container network-mode vocabulary.
package netmode

// Mode is Engine NetworkMode / Compose network_mode.
type Mode string

const (
	Bridge  Mode = "bridge"
	Host    Mode = "host"
	None    Mode = "none"
	Default Mode = "default"
)

// Container uses another container's network namespace.
func Container(id string) Mode { return Mode("container:" + id) }

// Service uses a Compose service's network namespace.
func Service(name string) Mode { return Mode("service:" + name) }

// All returns the fixed network modes (not container:/service: forms).
func All() []Mode { return []Mode{Bridge, Host, None, Default} }
