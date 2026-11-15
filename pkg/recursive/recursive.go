// Package recursive is the Compose bind.recursive vocabulary.
package recursive

// Mode is volumes.bind.recursive.
type Mode string

const (
	Enabled  Mode = "enabled"
	Disabled Mode = "disabled"
	Writable Mode = "writable"
	ReadOnly Mode = "readonly"
)

// All returns every known bind recursive mode.
func All() []Mode { return []Mode{Enabled, Disabled, Writable, ReadOnly} }
