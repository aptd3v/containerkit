// Package endpoint is the Swarm endpoint_mode vocabulary.
package endpoint

// Mode is deploy.endpoint_mode (vip or dnsrr).
type Mode string

const (
	VIP   Mode = "vip"
	DNSRR Mode = "dnsrr"
)

// All returns every known endpoint mode.
func All() []Mode { return []Mode{VIP, DNSRR} }
