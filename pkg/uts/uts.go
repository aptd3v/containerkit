// Package uts is the UTS namespace mode vocabulary.
package uts

// Mode is an Engine UTSMode / Compose uts value.
type Mode string

const Host Mode = "host"

// All returns every known UTS mode.
func All() []Mode { return []Mode{Host} }
