// Package vuo provides options for volume update operations.
package update

import (
	"github.com/aptd3v/go-contain/pkg/client/options/volume/create/clusterspec"
	"github.com/docker/docker/api/types/volume"
)

// SetVolumeUpdateOption is a function that sets a volume update option.
type SetVolumeUpdateOption func(*volume.UpdateOptions) error

// WithClusterVolumeSpec sets the cluster volume spec for the volume update.
func WithClusterVolumeSpec(setters ...clusterspec.SetClusterVolumeSpecOption) SetVolumeUpdateOption {
	return func(o *volume.UpdateOptions) error {
		if o.Spec == nil {
			o.Spec = &volume.ClusterVolumeSpec{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(o.Spec); err != nil {
				return err
			}
		}
		return nil
	}
}
