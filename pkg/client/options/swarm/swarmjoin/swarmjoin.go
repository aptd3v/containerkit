// Package swarmjoin provides options for the swarm join command.
package swarmjoin

import (
	"github.com/docker/docker/api/types/swarm"
)

// SetSwarmJoinOption is a function that sets a join request option.
type SetSwarmJoinOption func(*swarm.JoinRequest) error

// WithJoinToken sets the join token of the node.
func WithJoinToken(token string) SetSwarmJoinOption {
	return func(o *swarm.JoinRequest) error {
		o.JoinToken = token
		return nil
	}
}

// WithAdvertiseAddr sets the advertise address of the node.
func WithAdvertiseAddr(addr string) SetSwarmJoinOption {
	return func(o *swarm.JoinRequest) error {
		o.AdvertiseAddr = addr
		return nil
	}
}

// WithDataPathAddr sets the data path address of the node.
func WithDataPathAddr(addr string) SetSwarmJoinOption {
	return func(o *swarm.JoinRequest) error {
		o.DataPathAddr = addr
		return nil
	}
}

// WithAvailability sets the availability of the node.
func WithAvailability(availability swarm.NodeAvailability) SetSwarmJoinOption {
	return func(o *swarm.JoinRequest) error {
		o.Availability = availability
		return nil
	}
}

// WithRemoteAddrs appends remote addresses to the join request.
func WithRemoteAddrs(addrs ...string) SetSwarmJoinOption {
	return func(o *swarm.JoinRequest) error {
		if o.RemoteAddrs == nil {
			o.RemoteAddrs = make([]string, 0, len(addrs))
		}
		o.RemoteAddrs = append(o.RemoteAddrs, addrs...)
		return nil
	}
}
