// Package ipno provides options for image prune.
package ipno

import "github.com/docker/docker/api/types/filters"

// SetImagePruneOption is a function that sets a parameter for the image prune.
type SetImagePruneOption func(filters.Args) error

// WithFilters sets the filters for the image prune options.
func WithFilters(key, value string) SetImagePruneOption {
	return func(args filters.Args) error {
		args.Add(key, value)
		return nil
	}
}
