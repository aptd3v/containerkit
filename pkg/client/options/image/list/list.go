// package list provides options for the image list.
package list

import (
	"github.com/docker/docker/api/types/image"
)

// SetImageListOption is a function that sets a parameter for the image list.
type SetImageListOption func(*image.ListOptions) error

// WithAll sets the all flag for the image list.
// All controls whether all images in the graph are filtered, or just the heads.
func WithAll() SetImageListOption {
	return func(o *image.ListOptions) error {
		o.All = true
		return nil
	}
}

// WithFilters sets the filters for the image list.
func WithFilter(key, value string) SetImageListOption {
	return func(o *image.ListOptions) error {
		o.Filters.Add(key, value)
		return nil
	}
}

// WithSharedSize sets the shared size for the image list.
func WithSharedSize() SetImageListOption {
	return func(o *image.ListOptions) error {
		o.SharedSize = true
		return nil
	}
}

// WithContainerCount sets the container count for the image list.
func WithContainerCount() SetImageListOption {
	return func(o *image.ListOptions) error {
		o.ContainerCount = true
		return nil
	}
}

// WithManifests sets the manifests for the image list.
// Manifests indicates whether the image manifests should be returned
func WithManifests() SetImageListOption {
	return func(o *image.ListOptions) error {
		o.Manifests = true
		return nil
	}
}
