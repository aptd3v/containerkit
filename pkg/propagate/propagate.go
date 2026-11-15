// Package propagate is the bind-mount propagation vocabulary.
package propagate

// Mode is Engine mount.Propagation / Compose bind.propagation.
type Mode string

const (
	Private  Mode = "private"
	RPrivate Mode = "rprivate"
	Shared   Mode = "shared"
	RShared  Mode = "rshared"
	Slave    Mode = "slave"
	RSlave   Mode = "rslave"
)

// All returns every known bind propagation mode.
func All() []Mode { return []Mode{Private, RPrivate, Shared, RShared, Slave, RSlave} }
