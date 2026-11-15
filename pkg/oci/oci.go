// Package oci is the Compose publish OCI manifest-version vocabulary.
package oci

// Version is a Compose publish OCI manifest version.
type Version string

const (
	V1_0 Version = "1.0"
	V1_1 Version = "1.1"
)

// All returns every known OCI manifest version.
func All() []Version { return []Version{V1_0, V1_1} }
