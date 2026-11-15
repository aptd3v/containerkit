// Package selinux is the bind-mount SELinux relabel vocabulary.
package selinux

// Label is Compose volumes.bind.selinux ("z" shared or "Z" private).
type Label string

const (
	Shared  Label = "z"
	Private Label = "Z"
)

// All returns every known SELinux bind label.
func All() []Label { return []Label{Shared, Private} }
