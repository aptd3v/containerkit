// Package swarminit provides options for the swarm init command.
package swarminit

import (
	"github.com/docker/docker/api/types/swarm"
)

// SetSwarmInitOption is a function that sets a swarm init request option.
type SetSwarmInitOption func(*swarm.InitRequest) error

// WithListenAddr sets the listen address of the swarm init request.
func WithListenAddr(addr string) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.ListenAddr = addr
		return nil
	}
}

// WithAdvertiseAddr sets the advertise address of the swarm init request.
func WithAdvertiseAddr(addr string) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.AdvertiseAddr = addr
		return nil
	}
}

// WithDataPathAddr sets the data path address of the swarm init request.
func WithDataPathAddr(addr string) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.DataPathAddr = addr
		return nil
	}
}

// WithDataPathPort sets the data path port of the swarm init request.
func WithDataPathPort(port int) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.DataPathPort = uint32(port)
		return nil
	}
}

// WithForceNewCluster sets the force new cluster flag of the swarm init request.
func WithForceNewCluster() SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.ForceNewCluster = true
		return nil
	}
}

// WithSpec sets the spec of the swarm init request.
func WithSpec(spec swarm.Spec) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.Spec = spec
		return nil
	}
}

// WithAutoLockManagers sets the auto lock managers flag of the swarm init request.
func WithAutoLockManagers() SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.AutoLockManagers = true
		return nil
	}
}

// WithAvailability sets the availability of the swarm init request.
func WithAvailability(availability swarm.NodeAvailability) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.Availability = swarm.NodeAvailability(availability)
		return nil
	}
}

// WithDefaultAddrPool sets the default address pool of the swarm init request.
func WithDefaultAddrPool(addrs ...string) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		if o.DefaultAddrPool == nil {
			o.DefaultAddrPool = make([]string, 0, len(addrs))
		}
		o.DefaultAddrPool = append(o.DefaultAddrPool, addrs...)
		return nil
	}
}

// WithSubnetSize sets the subnet size of the swarm init request.
func WithSubnetSize(size int) SetSwarmInitOption {
	return func(o *swarm.InitRequest) error {
		o.SubnetSize = uint32(size)
		return nil
	}
}
