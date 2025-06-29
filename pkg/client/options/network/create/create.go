// package create provides options for the network create.
package create

import (
	"github.com/aptd3v/go-contain/pkg/client/options/network/create/ipam"
	"github.com/docker/docker/api/types/network"
)

// SetNetworkCreateOption is a function that sets a parameter for the network create.
type SetNetworkCreateOption func(*network.CreateOptions) error

// WithDriver sets the driver for the network create.
//
// Driver is the driver-name used to create the network (e.g. `bridge`, `overlay`)
func WithDriver(driver string) SetNetworkCreateOption {
	return func(o *network.CreateOptions) error {
		o.Driver = driver
		return nil
	}
}

// WithScope sets the scope for the network create.
//
// Scope describes the level at which the network exists (e.g. `swarm` for cluster-wide or `local` for machine level).
func WithScope(scope string) SetNetworkCreateOption {
	return func(o *network.CreateOptions) error {
		o.Scope = scope
		return nil
	}
}

// WithEnableIPv4 sets the enable IPv4 for the network create.
//
// EnableIPv4 represents whether to enable IPv4.
func WithEnableIPv4() SetNetworkCreateOption {
	enableIPv4 := true
	return func(o *network.CreateOptions) error {
		o.EnableIPv4 = &enableIPv4
		return nil
	}
}

// WithEnableIPv6 sets the enable IPv6 for the network create.
//
// EnableIPv6 represents whether to enable IPv6.
func WithEnableIPv6() SetNetworkCreateOption {
	enableIPv6 := true
	return func(o *network.CreateOptions) error {
		o.EnableIPv6 = &enableIPv6
		return nil
	}
}

// WithIPAM sets the IPAM for the network create.
//
// IPAM is the network's IP Address Management.
func WithIPAM(setters ...ipam.SetIPAMOption) SetNetworkCreateOption {
	return func(o *network.CreateOptions) error {
		if o.IPAM == nil {
			o.IPAM = &network.IPAM{}
		}
		for _, setter := range setters {
			if err := setter(o.IPAM); err != nil {
				return err
			}
		}
		return nil
	}
}
