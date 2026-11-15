// Package watch is the Compose develop.watch action vocabulary.
package watch

// Action is a Compose develop.watch action.
type Action string

const (
	Sync        Action = "sync"
	Rebuild     Action = "rebuild"
	SyncRestart Action = "sync+restart"
	SyncExec    Action = "sync+exec"
	Restart     Action = "restart"
)

// All returns every known watch action.
func All() []Action { return []Action{Sync, Rebuild, SyncRestart, SyncExec, Restart} }
