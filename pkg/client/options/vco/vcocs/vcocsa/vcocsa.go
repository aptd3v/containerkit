// package vcocsa provides options for the volume cluster volume spec access mode.
package vcocsa

import "github.com/docker/docker/api/types/volume"

// SharingMode defines the Sharing of a Cluster Volume. This is how Tasks using a
// Volume at the same time can use it.
type SharingMode string

const (
	// SharingNone indicates that only one Task may use the Volume at a
	// time.
	SharingNone SharingMode = "none"

	// SharingReadOnly indicates that the Volume may be shared by any
	// number of Tasks, but they must be read-only.
	SharingReadOnly SharingMode = "readonly"

	// SharingOneWriter indicates that the Volume may be shared by any
	// number of Tasks, but all after the first must be read-only.
	SharingOneWriter SharingMode = "onewriter"

	// SharingAll means that the Volume may be shared by any number of
	// Tasks, as readers or writers.
	SharingAll SharingMode = "all"
)

type Scope string

const (
	// ScopeSingleNode indicates the volume can be used on one node at a
	// time.
	ScopeSingleNode Scope = "single"

	// ScopeMultiNode indicates the volume can be used on many nodes at
	// the same time.
	ScopeMultiNode Scope = "multi"
)

// SetAccessModeOption is a function that sets a parameter for the volume cluster volume spec access mode.
type SetAccessModeOption func(*volume.AccessMode) error

// WithScope sets the scope of the volume cluster volume spec access mode.
// Scope defines the set of nodes this volume can be used on at one time.
func WithScope(scope Scope) SetAccessModeOption {
	return func(o *volume.AccessMode) error {
		o.Scope = volume.Scope(scope)
		return nil
	}
}

// WithSharing sets the sharing of the volume cluster volume spec access mode.
// Sharing defines the number and way that different tasks can use this
// volume at one time.
func WithSharing(sharing SharingMode) SetAccessModeOption {
	return func(o *volume.AccessMode) error {
		o.Sharing = volume.SharingMode(sharing)
		return nil
	}
}

// WithMountVolume	 sets the mount volume of the volume cluster volume spec access mode.
// MountVolume defines options for using this volume as a Mount-type
// volume.
//
// Either BlockVolume or MountVolume, but not both, must be present.
func WithMountVolume(fsType string, mountFlags ...string) SetAccessModeOption {
	return func(o *volume.AccessMode) error {
		o.MountVolume = &volume.TypeMount{
			FsType:     fsType,
			MountFlags: mountFlags,
		}
		return nil
	}
}

// WithBlockVolume sets the block volume of the volume cluster volume spec access mode.
// MountVolume defines options for using this volume as a Mount-type
// volume.
//
// Either BlockVolume or MountVolume, but not both, must be present.
func WithBlockVolume() SetAccessModeOption {
	return func(o *volume.AccessMode) error {
		o.BlockVolume = &volume.TypeBlock{}
		return nil
	}
}
