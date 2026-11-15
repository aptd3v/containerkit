// Package pidns is the PID namespace mode vocabulary.
package pidns

// Mode is an Engine PidMode / Compose pid value.
type Mode string

const Host Mode = "host"

// Container joins another container's PID namespace.
func Container(id string) Mode { return Mode("container:" + id) }

// All returns the fixed PID modes (not container:<id> forms).
func All() []Mode { return []Mode{Host} }
