// package create provides options for the volume create.
package create

import (
	"github.com/aptd3v/go-contain/pkg/client/options/volume/create/clusterspec"
	"github.com/docker/docker/api/types/volume"
)

// SetVolumeCreateOption is a function that sets a parameter for the volume create.
type SetVolumeCreateOption func(*volume.CreateOptions) error

// WithClusterVolumeSpec sets the cluster volume spec.
//
// AccessMode defines how the volume is used by tasks.
func WithClusterVolumeSpec(setters ...clusterspec.SetClusterVolumeSpecOption) SetVolumeCreateOption {
	return func(o *volume.CreateOptions) error {
		if o.ClusterVolumeSpec == nil {
			o.ClusterVolumeSpec = &volume.ClusterVolumeSpec{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(o.ClusterVolumeSpec); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithDriver sets the driver of the volume.
func WithDriver(driver string) SetVolumeCreateOption {
	return func(o *volume.CreateOptions) error {
		o.Driver = driver
		return nil
	}
}

// WithDriverOpts appends the driver options of the volume.
//
// These options are passed directly to the driver and are driver specific.
func WithDriverOpts(key string, value string) SetVolumeCreateOption {
	return func(o *volume.CreateOptions) error {
		if o.DriverOpts == nil {
			o.DriverOpts = make(map[string]string)
		}
		o.DriverOpts[key] = value
		return nil
	}
}

// WithLabels appends the labels of the volume.
//
// User-defined key/value metadata.
func WithLabels(key string, value string) SetVolumeCreateOption {
	return func(o *volume.CreateOptions) error {
		if o.Labels == nil {
			o.Labels = make(map[string]string)
		}
		o.Labels[key] = value
		return nil
	}
}

// WithName sets the name of the volume.
//
// The new volume's name. If not specified, Docker generates a name.
func WithName(name string) SetVolumeCreateOption {
	return func(o *volume.CreateOptions) error {
		o.Name = name
		return nil
	}
}
