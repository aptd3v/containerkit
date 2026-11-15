// Package isolation is the Windows isolation vocabulary.
package isolation

// Mode is Engine Isolation / Compose isolation.
type Mode string

const (
	Default Mode = "default"
	Process Mode = "process"
	HyperV  Mode = "hyperv"
)

// All returns every known isolation mode.
func All() []Mode { return []Mode{Default, Process, HyperV} }
