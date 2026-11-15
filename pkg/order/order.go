// Package order is the Swarm update/rollback order vocabulary.
package order

// Order is update/rollback order (start-first or stop-first).
type Order string

const (
	StartFirst Order = "start-first"
	StopFirst  Order = "stop-first"
)

// All returns every known update order.
func All() []Order { return []Order{StartFirst, StopFirst} }
