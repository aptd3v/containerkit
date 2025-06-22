// package nio provides options for the network inspect.
package nio

import "github.com/docker/docker/api/types/network"

// SetNetworkInspectOption is a function that sets a parameter for the network inspect.
type SetNetworkInspectOption func(*network.InspectOptions) error

// WithScope sets the scope for the network inspect.
func WithScope(scope string) SetNetworkInspectOption {
	return func(o *network.InspectOptions) error {
		o.Scope = scope
		return nil
	}
}

// WithVerbose sets the verbose flag for the network inspect.
func WithVerbose() SetNetworkInspectOption {
	return func(o *network.InspectOptions) error {
		o.Verbose = true
		return nil
	}
}
