// Package vpo provides options for volume prune operations.
package vpo

import (
	"github.com/docker/docker/api/types/filters"
)

// SetVolumePruneOption is a function that sets a volume prune option.
type SetVolumePruneOption func(filters.Args) error

// WithFilters sets the filters for the volume prune.
func WithFilters(key, value string) SetVolumePruneOption {
	return func(args filters.Args) error {
		args.Add(key, value)
		return nil
	}
}
