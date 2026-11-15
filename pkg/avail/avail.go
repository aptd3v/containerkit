// Package avail is the Swarm node-availability vocabulary.
package avail

// Availability is a swarm node availability.
type Availability string

const (
	Active Availability = "active"
	Pause  Availability = "pause"
	Drain  Availability = "drain"
)

// All returns every known node availability.
func All() []Availability { return []Availability{Active, Pause, Drain} }
