// Package progress is the Compose / logger progress-mode vocabulary.
package progress

// Mode is a compose build progress or NewLogger progress setting.
type Mode string

const (
	Auto  Mode = "auto"
	Plain Mode = "plain"
	TTY   Mode = "tty"
	Quiet Mode = "quiet"
	JSON  Mode = "json"
)

// All returns every known progress mode.
func All() []Mode { return []Mode{Auto, Plain, TTY, Quiet, JSON} }
