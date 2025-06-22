// package prune provides options for the network prune.
package prune

import "github.com/docker/docker/api/types/filters"

// SetNetworkPruneOption is a function that sets a parameter for the network prune.
type SetNetworkPruneOption func(*filters.Args) error

// WithFilters appends the filters for the network prune.
func WithFilters(key, value string) SetNetworkPruneOption {
	return func(o *filters.Args) error {
		o.Add(key, value)
		return nil
	}
}
