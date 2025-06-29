// package ipam provides options for the network IPAM.
package ipam

import (
	"github.com/aptd3v/go-contain/pkg/client/options/network/create/ipam/ipamconfig"
	"github.com/docker/docker/api/types/network"
)

// SetIPAMConfig is a function that sets a parameter for the network IPAM.
type SetIPAMOption func(*network.IPAM) error

func WithDriver(driver string) SetIPAMOption {
	return func(o *network.IPAM) error {
		o.Driver = driver
		return nil
	}
}

// WithOptions appends the options for the network IPAM.
func WithOptions(key, value string) SetIPAMOption {
	return func(o *network.IPAM) error {
		if o.Options == nil {
			o.Options = make(map[string]string)
		}
		o.Options[key] = value
		return nil
	}
}

func WithIpamConfig(setters ...ipamconfig.SetIPAMConfig) SetIPAMOption {
	return func(o *network.IPAM) error {
		if o.Config == nil {
			o.Config = make([]network.IPAMConfig, 0)
		}
		ipamConfig := &network.IPAMConfig{}
		for _, setter := range setters {
			if err := setter(ipamConfig); err != nil {
				return err
			}
		}
		o.Config = append(o.Config, *ipamConfig)
		return nil
	}
}
