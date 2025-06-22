// package vcocs provides options for the volume cluster volume spec.
package vcocs

import (
	"github.com/aptd3v/go-contain/pkg/client/options/vco/vcocs/vcocsa"
	"github.com/aptd3v/go-contain/pkg/client/options/vco/vcocs/vcocsar"
	"github.com/docker/docker/api/types/volume"
)

// Availability specifies the availability of the volume.
type Availability string

const (
	// AvailabilityActive indicates that the volume is active and fully
	// schedulable on the cluster.
	AvailabilityActive Availability = "active"

	// AvailabilityPause indicates that no new workloads should use the
	// volume, but existing workloads can continue to use it.
	AvailabilityPause Availability = "pause"

	// AvailabilityDrain indicates that all workloads using this volume
	// should be rescheduled, and the volume unpublished from all nodes.
	AvailabilityDrain Availability = "drain"
)

// SetClusterVolumeSpecOption is a function that sets a parameter for the volume cluster volume spec.
type SetClusterVolumeSpecOption func(*volume.ClusterVolumeSpec) error

// WithGroup sets the volume group of this volume. Volumes belonging to the
// same group can be referred to by group name when creating Services.
// Referring to a volume by group instructs swarm to treat volumes in that
// group interchangeably for the purpose of scheduling. Volumes with an
// empty string for a group technically all belong to the same, emptystring
// group.
func WithGroup(group string) SetClusterVolumeSpecOption {
	return func(o *volume.ClusterVolumeSpec) error {
		o.Group = group
		return nil
	}
}

// WithAccessMode sets the access mode of the volume.
// AccessMode defines how the volume is used by tasks.
func WithAccessMode(setters ...vcocsa.SetAccessModeOption) SetClusterVolumeSpecOption {
	return func(o *volume.ClusterVolumeSpec) error {
		if o.AccessMode == nil {
			o.AccessMode = &volume.AccessMode{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(o.AccessMode); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithAccessabilityRequirements sets the accessability requirements of the volume.
// AccessibilityRequirements specifies where in the cluster a volume must
// be accessible from.
//
// This field must be empty if the plugin does not support
// VOLUME_ACCESSIBILITY_CONSTRAINTS capabilities. If it is present but the
// plugin does not support it, volume will not be created.
//
// If AccessibilityRequirements is empty, but the plugin does support
// VOLUME_ACCESSIBILITY_CONSTRAINTS, then Swarmkit will assume the entire
// cluster is a valid target for the volume.
func WithAccessibilityRequirements(setters ...vcocsar.SetAccessibilityRequirementsOption) SetClusterVolumeSpecOption {
	return func(o *volume.ClusterVolumeSpec) error {
		if o.AccessibilityRequirements == nil {
			o.AccessibilityRequirements = &volume.TopologyRequirement{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(o.AccessibilityRequirements); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithCapacityRange sets the capacity range of the volume.
//
// CapacityRange defines the desired capacity that the volume should be
// created with. If WithCapacityRange is not called, the plugin will decide the capacity.
func WithCapacityRange(requiredBytes int64, limitBytes int64) SetClusterVolumeSpecOption {
	return func(o *volume.ClusterVolumeSpec) error {
		o.CapacityRange = &volume.CapacityRange{
			RequiredBytes: requiredBytes,
			LimitBytes:    limitBytes,
		}
		return nil
	}
}

// WithSecrets appends the secrets of the volume.
//
// Secrets defines Swarm Secrets that are passed to the CSI storage plugin
// when operating on this volume.
func WithSecrets(key string, secret string) SetClusterVolumeSpecOption {
	return func(o *volume.ClusterVolumeSpec) error {
		if o.Secrets == nil {
			o.Secrets = make([]volume.Secret, 0)
		}
		o.Secrets = append(o.Secrets, volume.Secret{
			Key:    key,
			Secret: secret,
		})
		return nil
	}
}

// WithAvailability sets the availability of the volume.
//
// Availability is the Volume's desired availability.
// Analogous to Node Availability, this allows the user
// to take volumes offline in order to update or delete them.
func WithAvailability(availability Availability) SetClusterVolumeSpecOption {
	return func(o *volume.ClusterVolumeSpec) error {
		o.Availability = volume.Availability(availability)
		return nil
	}
}
