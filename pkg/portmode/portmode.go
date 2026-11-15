// Package portmode is the Compose ports.mode vocabulary.
package portmode

// Mode is a published-port mode (ingress or host).
type Mode string

const (
	Ingress Mode = "ingress"
	Host    Mode = "host"
)

// All returns every known port publish mode.
func All() []Mode { return []Mode{Ingress, Host} }
