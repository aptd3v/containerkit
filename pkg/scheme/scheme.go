// Package scheme is the Engine client URL-scheme vocabulary.
package scheme

// Mode is a Docker daemon URL scheme.
type Mode string

const (
	Unix  Mode = "unix"
	Npipe Mode = "npipe"
	TCP   Mode = "tcp"
	HTTP  Mode = "http"
	HTTPS Mode = "https"
	SSH   Mode = "ssh"
	FD    Mode = "fd"
)

// All returns every known client scheme.
func All() []Mode { return []Mode{Unix, Npipe, TCP, HTTP, HTTPS, SSH, FD} }
