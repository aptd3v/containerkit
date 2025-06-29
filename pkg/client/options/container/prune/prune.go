// package prune provides options for the container prune.
package prune

import "github.com/docker/docker/api/types/filters"

// SetContainerPruneOption is a function that sets a parameter for the container prune.
type SetContainerPruneOption func(filters.Args) error

// WithFilters sets the filters for the container prune options.
func WithFilters(key, value string) SetContainerPruneOption {
	return func(args filters.Args) error {
		args.Add(key, value)
		return nil
	}
}
