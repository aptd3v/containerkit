// Package userns is the user namespace mode vocabulary.
package userns

// Mode is an Engine UsernsMode / Compose userns_mode value.
type Mode string

const Host Mode = "host"

// All returns every known user namespace mode.
func All() []Mode { return []Mode{Host} }
