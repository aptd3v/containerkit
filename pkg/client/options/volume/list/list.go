// Package list provides options for volume list operations.
package list

import "github.com/docker/docker/api/types/volume"

// SetVolumeListOption is a function that sets a volume list option.
type SetVolumeListOption func(volume.ListOptions) volume.ListOptions

// WithFilter is a function that sets the filter option.
func WithFilter(key, value string) SetVolumeListOption {
	return func(o volume.ListOptions) volume.ListOptions {
		o.Filters.Add(key, value)
		return o
	}
}
