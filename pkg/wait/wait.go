// Package wait is the Engine container-wait condition vocabulary.
package wait

// Condition is an Engine wait-for-container condition.
type Condition string

const (
	NotRunning Condition = "not-running"
	NextExit   Condition = "next-exit"
	Removed    Condition = "removed"
)

// All returns every known wait condition.
func All() []Condition { return []Condition{NotRunning, NextExit, Removed} }
