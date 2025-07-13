// Package pool provides functions to set the ipam pool configuration for a project
package pool

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

// SetIpamPoolProjectConfig is a function that sets the ipam pool configuration for a project
type SetIpamPoolProjectConfig func(*types.IPAMPool) error

// WithSubnet sets the subnet for the ipam pool
// parameters:
//   - subnet: the subnet for the ipam pool
func WithSubnet(subnet string) SetIpamPoolProjectConfig {
	return func(opt *types.IPAMPool) error {
		opt.Subnet = subnet
		return nil
	}
}

// WithGateway sets the gateway for the ipam pool
// parameters:
//   - gateway: the gateway for the ipam pool
func WithGateway(gateway string) SetIpamPoolProjectConfig {
	return func(opt *types.IPAMPool) error {
		opt.Gateway = gateway
		return nil
	}
}

// WithIpRange sets the ip range for the ipam pool
// parameters:
//   - ipRange: the ip range for the ipam pool
func WithIpRange(ipRange string) SetIpamPoolProjectConfig {
	return func(opt *types.IPAMPool) error {
		opt.IPRange = ipRange
		return nil
	}
}

// WithAuxiliaryAddresses sets the auxiliary addresses for the ipam pool
// parameters:
//   - key: the key of the auxiliary address
//   - value: the value of the auxiliary address
func WithAuxiliaryAddresses(key, value string) SetIpamPoolProjectConfig {
	return func(opt *types.IPAMPool) error {
		if opt.AuxiliaryAddresses == nil {
			opt.AuxiliaryAddresses = make(map[string]string)
		}
		opt.AuxiliaryAddresses[key] = value
		return nil
	}
}

// Fail is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the pool config
// and append the error to the service config error collection
func Fail(err error) SetIpamPoolProjectConfig {
	return func(opt *types.IPAMPool) error {
		return errdefs.NewServiceConfigError("pool", err.Error())
	}
}

// Failf is a function that returns a setter that always returns the given error
//
// note: this is useful for when you want to fail the pool config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetIpamPoolProjectConfig {
	return func(opt *types.IPAMPool) error {
		return errdefs.NewServiceConfigError("pool", fmt.Sprintf(stringFormat, args...))
	}
}
