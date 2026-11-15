// Package netdrv is the Engine/Compose network-driver vocabulary.
package netdrv

// Driver is a built-in network driver name.
type Driver string

const (
	Bridge  Driver = "bridge"
	Host    Driver = "host"
	None    Driver = "none"
	Overlay Driver = "overlay"
	Macvlan Driver = "macvlan"
	IPvlan  Driver = "ipvlan"
)

// All returns every known built-in network driver.
func All() []Driver { return []Driver{Bridge, Host, None, Overlay, Macvlan, IPvlan} }
