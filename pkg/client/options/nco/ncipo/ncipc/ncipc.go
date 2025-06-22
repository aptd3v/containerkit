// package ncipc provides options for the network IPAM config.
package ncipc

import "github.com/docker/docker/api/types/network"

// SetIPAMConfig is a function that sets a parameter for the network IPAM config.
type SetIPAMConfig func(*network.IPAMConfig) error

// WithSubnet sets the subnet for the network IPAM config.
func WithSubnet(subnet string) SetIPAMConfig {
	return func(o *network.IPAMConfig) error {
		o.Subnet = subnet
		return nil
	}
}

// WithIPRange sets the IP range for the network IPAM config.
func WithIPRange(ipRange string) SetIPAMConfig {
	return func(o *network.IPAMConfig) error {
		o.IPRange = ipRange
		return nil
	}
}

// WithGateway sets the gateway for the network IPAM config.
func WithGateway(gateway string) SetIPAMConfig {
	return func(o *network.IPAMConfig) error {
		o.Gateway = gateway
		return nil
	}
}

// WithAuxiliaryAddress appends the auxiliary address for the network IPAM config.
func WithAuxiliaryAddress(key, value string) SetIPAMConfig {
	return func(o *network.IPAMConfig) error {
		if o.AuxAddress == nil {
			o.AuxAddress = make(map[string]string)
		}
		o.AuxAddress[key] = value
		return nil
	}
}
