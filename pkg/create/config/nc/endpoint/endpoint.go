// Package endpoint provides the options for the endpoint config in the network config.
package endpoint

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint/ipam"
	"github.com/docker/docker/api/types/network"
)

// SetEndpointConfig is a function that sets the endpoint config
type SetEndpointConfig func(options *network.EndpointSettings) error

// WithIPAMConfig sets the IPAM config for the network
// Parameter:
//   - setters: the setters to be used for the IPAM config of the endpoint
func WithIPAMConfig(setters ...ipam.SetIPAMConfig) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		if options.IPAMConfig == nil {
			options.IPAMConfig = &network.EndpointIPAMConfig{}
		}
		for _, setter := range setters {
			if setter != nil {
				if err := setter(options.IPAMConfig); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// WithLinks sets the links for the endpoint.
// Links holds the list of network endpoints to link to.
//
// Parameter:
//   - links: the links to be used for the network
func WithLinks(links ...string) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		if options.Links == nil {
			options.Links = make([]string, 0)
		}
		options.Links = append(options.Links, links...)
		return nil
	}
}

// WithAliases sets the aliases for the endpoint.
// Aliases holds the list of extra, user-specified DNS names for this endpoint.
//
// Parameter:
//   - aliases: the aliases to be used for the network
func WithAliases(aliases ...string) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		if options.Aliases == nil {
			options.Aliases = make([]string, 0)
		}
		options.Aliases = append(options.Aliases, aliases...)
		return nil
	}
}

// WithMacAddress sets the MAC address for the endpoint
// Parameter:
//   - macAddress: MAC address to be used for the endpoint
func WithMacAddress(macAddress string) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		options.MacAddress = macAddress
		return nil
	}
}

// WithDriverOptions sets the driver options for the endpoint
// Parameter:
//   - key: the key of the driver option
//   - value: the value of the driver option
func WithDriverOptions(key, value string) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		if options.DriverOpts == nil {
			options.DriverOpts = make(map[string]string)
		}
		options.DriverOpts[key] = value
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the endpoint
// and append the error to the network config error collection
func Fail(err error) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		return create.NewNetworkConfigError("endpoint", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the endpoint
// and append the error to the network config error collection
func Failf(stringFormat string, args ...interface{}) SetEndpointConfig {
	return func(options *network.EndpointSettings) error {
		return create.NewNetworkConfigError("endpoint", fmt.Sprintf(stringFormat, args...))
	}
}
