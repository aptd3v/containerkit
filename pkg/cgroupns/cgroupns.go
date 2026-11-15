// Package cgroupns is the cgroup namespace mode vocabulary.
package cgroupns

// Mode is an Engine CgroupnsMode / Compose cgroupns value.
type Mode string

const (
	Host    Mode = "host"
	Private Mode = "private"
)

// All returns every known cgroup namespace mode.
func All() []Mode { return []Mode{Host, Private} }
