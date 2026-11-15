// Package failact is the Swarm update/rollback failure_action vocabulary.
package failact

// Action is update/rollback failure_action.
type Action string

const (
	Continue Action = "continue"
	Pause    Action = "pause"
	Rollback Action = "rollback"
)

// All returns every known failure action.
func All() []Action { return []Action{Continue, Pause, Rollback} }
