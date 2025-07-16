// Package ipam provides the options for the IPAM config in the endpoint config.
package ipam

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/network"
)

// SetIPAMOptionFn is a function that sets an IPAM option for the network
type SetIPAMConfig func(opt *network.EndpointIPAMConfig) error

// WithIPv4Address sets the IPv4 address for the IPAM config.
// If the address is already set, it will be overwritten.
// Parameter:
//   - address: the IPv4 address to be used for the network
func WithIPv4Address(address string) SetIPAMConfig {
	return func(opt *network.EndpointIPAMConfig) error {
		opt.IPv4Address = address
		return nil
	}
}

// WithIPv6Address sets the IPv6 address for the IPAM config.
// If the address is already set, it will be overwritten.
// Parameter:
//   - address: the IPv6 address to be used for the network
func WithIPv6Address(address string) SetIPAMConfig {
	return func(opt *network.EndpointIPAMConfig) error {
		opt.IPv6Address = address
		return nil
	}
}

// WithLinkLocalIPs appends the link local IPs for the IPAM config
// Parameter:
//   - ips: the link local IPs to be used for the network
func WithLinkLocalIPs(ips ...string) SetIPAMConfig {
	return func(opt *network.EndpointIPAMConfig) error {
		if opt.LinkLocalIPs == nil {
			opt.LinkLocalIPs = make([]string, 0)
		}
		opt.LinkLocalIPs = append(opt.LinkLocalIPs, ips...)
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the IPAM config
// and append the error to the network config error collection
func Fail(err error) SetIPAMConfig {
	return func(opt *network.EndpointIPAMConfig) error {
		return errdefs.NewNetworkConfigError("ipam", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the IPAM config
// and append the error to the network config error collection
func Failf(stringFormat string, args ...any) SetIPAMConfig {
	return func(opt *network.EndpointIPAMConfig) error {
		return errdefs.NewNetworkConfigError("ipam", fmt.Sprintf(stringFormat, args...))
	}
}
