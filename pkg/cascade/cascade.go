// Package cascade is the Compose start/up OnExit vocabulary.
package cascade

// Action is a Compose start/up cascade-on-exit behavior.
type Action int

const (
	Ignore Action = iota
	Stop
	Fail
)

// All returns every known cascade action.
func All() []Action { return []Action{Ignore, Stop, Fail} }
