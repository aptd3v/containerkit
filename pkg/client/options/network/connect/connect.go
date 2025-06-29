// package connect provides options for the network connect.
package connect

import (
	"github.com/aptd3v/go-contain/pkg/create/config/nc/endpoint"
	"github.com/docker/docker/api/types/network"
)

// SetNetworkConnectOption is a function that sets a parameter for the network connect.
type SetNetworkConnectOption func(*network.ConnectOptions) error

// WithContainer sets the container ID for the network connect.
func WithContainer(id string) SetNetworkConnectOption {
	return func(o *network.ConnectOptions) error {
		o.Container = id
		return nil
	}
}

// WithEndpoint sets the endpoint configuration for the network connect.
func WithEndpoint(setters ...endpoint.SetEndpointConfig) SetNetworkConnectOption {
	return func(o *network.ConnectOptions) error {
		if o.EndpointConfig == nil {
			o.EndpointConfig = &network.EndpointSettings{}
		}
		o.EndpointConfig = &network.EndpointSettings{}
		for _, setter := range setters {
			if err := setter(o.EndpointConfig); err != nil {
				return err
			}
		}
		return nil
	}
}
