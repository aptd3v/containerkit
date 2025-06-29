// package list provides options for the network list.
package list

import (
	"github.com/docker/docker/api/types/network"
)

// SetNetworkListOption is a function that sets a parameter for the network list.
type SetNetworkListOption func(*network.ListOptions) error

// WithFilters appends the filters for the network list.
func WithFilters(key, value string) SetNetworkListOption {
	return func(o *network.ListOptions) error {
		o.Filters.Add(key, value)
		return nil
	}
}
