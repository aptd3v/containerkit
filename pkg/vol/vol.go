// Package vol is the mount-type vocabulary for Volume Kind.
package vol

// Kind is an Engine mount type / Compose volume type.
type Kind string

const (
	Bind    Kind = "bind"
	Volume  Kind = "volume"
	Tmpfs   Kind = "tmpfs"
	Pipe    Kind = "npipe"
	Cluster Kind = "cluster"
	Image   Kind = "image"
)

// All returns every known mount type.
func All() []Kind { return []Kind{Bind, Volume, Tmpfs, Pipe, Cluster, Image} }
