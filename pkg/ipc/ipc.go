// Package ipc is the IPC namespace mode vocabulary.
package ipc

import "fmt"

// Mode is an Engine IpcMode / Compose ipc value.
type Mode string

const (
	None      Mode = "none"
	Private   Mode = "private"
	Shareable Mode = "shareable"
	Host      Mode = "host"
)

// Container joins another container's IPC namespace.
func Container(id string) Mode { return Mode("container:" + id) }

// All returns the fixed IPC modes (not container:<id> forms).
func All() []Mode { return []Mode{None, Private, Shareable, Host} }

// String returns the Docker mode string.
func (m Mode) String() string { return string(m) }

var _ fmt.Stringer = Mode("")
