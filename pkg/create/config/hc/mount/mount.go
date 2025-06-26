// Package mount provides the options for the mount config in the host config.
package mount

import (
	"fmt"
	"os"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/mount"
)

// Propagation represents the propagation of a mount.
type Propagation string

const (
	// PropagationRPrivate RPRIVATE
	PropagationRPrivate Propagation = "rprivate"
	// PropagationPrivate PRIVATE
	PropagationPrivate Propagation = "private"
	// PropagationRShared RSHARED
	PropagationRShared Propagation = "rshared"
	// PropagationShared SHARED
	PropagationShared Propagation = "shared"
	// PropagationRSlave RSLAVE
	PropagationRSlave Propagation = "rslave"
	// PropagationSlave SLAVE
	PropagationSlave Propagation = "slave"
)

/*
MountType is constant for the type of mount

	// TypeBind is the type for mounting a host directory
	"bind"
	// TypeVolume is the type for remote storage volumes
	"volume"
	// TypeTmpfs is the type for mounting tmpfs
	"tmpfs"
	// TypeNamedPipe is the type for mounting Windows named pipes
	"npipe"
*/
type MountType string

const (
	// TypeBind is the type for mounting a host directory
	MountTypeBind MountType = "bind"
	// TypeVolume is the type for remote storage volumes
	MountTypeVolume MountType = "volume"
	// TypeTmpfs is the type for mounting tmpfs
	MountTypeTmpfs MountType = "tmpfs"
	// TypeNamedPipe is the type for mounting Windows named pipes
	MountTypeNamedPipe MountType = "npipe"
)

/*
Consistency is constant for the consistency of the mount

	// ConsistencyFull guarantees bind mount-like consistency
	"consistent"
	// ConsistencyCached mounts can cache read data and FS structure
	"cached"
	// ConsistencyDelegated mounts can cache read and written data and structure
	"delegated"
	// ConsistencyDefault provides "consistent" behavior unless overridden
	"default"
*/
type Consistency string

const (
	// ConsistencyFull guarantees bind mount-like consistency
	ConsistencyFull Consistency = "consistent"
	// ConsistencyCached mounts can cache read data and FS structure
	ConsistencyCached Consistency = "cached"
	// ConsistencyDelegated mounts can cache read and written data and structure
	ConsistencyDelegated Consistency = "delegated"
	// ConsistencyDefault provides "consistent" behavior unless overridden
	ConsistencyDefault Consistency = "default"
)

// SetMountConfig is a function that sets the mount config
type SetMountConfig func(opt *mount.Mount) error

// WithType sets the mount type
// parameters:
//   - t: the type of the mount
func WithType(t MountType) SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.Type = mount.Type(t)
		return nil
	}
}

// WithSource sets the mount source
// parameters:
//   - source: the source of the mount
func WithSource(source string) SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.Source = source
		return nil
	}
}

// WithTarget sets the mount target
// parameters:
//   - target: the target of the mount
func WithTarget(target string) SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.Target = target
		return nil
	}
}

// WithReadOnly sets the mount read only, (attempts recursive read-only if possible)
func WithReadOnly() SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.ReadOnly = true
		return nil
	}
}

// WithReadWrite sets the mount read write
func WithReadWrite() SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.ReadOnly = false
		return nil
	}
}

// WithConsistency sets the consistency of the mount
// parameters:
//   - consistency: the consistency of the mount (Consistency represents the consistency requirements of a mount)
func WithConsistency(consistency Consistency) SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.Consistency = mount.Consistency(consistency)
		return nil
	}
}

// WithTmpfsSizeBytes sets the size of the tmpfs mount in bytes
// parameters:
//   - size: the size of the tmpfs mount in bytes
//
// This will be converted to an operating system specific value
// depending on the host. For example, on linux, it will be converted to
// use a 'k', 'm' or 'g' syntax. BSD, though not widely supported with
// docker, uses a straight byte value.
//
// Percentages are not supported.
func WithTmpfsSizeBytes(size int) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.TmpfsOptions == nil {
			opt.TmpfsOptions = &mount.TmpfsOptions{}
		}
		opt.TmpfsOptions.SizeBytes = int64(size)
		return nil
	}
}

// WithTmpfsMode sets the mode of the tmpfs mount
// parameters:
//   - mode: the mode of the tmpfs mount upon creation
func WithTmpfsMode(mode os.FileMode) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.TmpfsOptions == nil {
			opt.TmpfsOptions = &mount.TmpfsOptions{}
		}
		opt.TmpfsOptions.Mode = mode
		return nil
	}
}

// WithTmpfsFlag sets the flag of the tmpfs mount
// parameters:
//   - flag: the flag to set on the tmpfs mount
//
// Example:
//   - WithTmpfsFlag("exec")
func WithTmpfsFlag(flag string) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.TmpfsOptions == nil {
			opt.TmpfsOptions = &mount.TmpfsOptions{}
		}
		if opt.TmpfsOptions.Options == nil {
			opt.TmpfsOptions.Options = make([][]string, 0)
		}
		opt.TmpfsOptions.Options = append(opt.TmpfsOptions.Options, []string{flag})
		return nil
	}
}

// WithTmpfsKeyValue sets the key value pair of the tmpfs mount
// parameters:
//   - key: the key of the key value pair
//   - value: the value of the key value pair
//
// Example:
//   - WithTmpfsKeyValue("uid", "1000")
//   - WithTmpfsKeyValue("gid", "1000")
func WithTmpfsKeyValue(key string, value string) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.TmpfsOptions == nil {
			opt.TmpfsOptions = &mount.TmpfsOptions{}
		}
		if opt.TmpfsOptions.Options == nil {
			opt.TmpfsOptions.Options = make([][]string, 0)
		}
		opt.TmpfsOptions.Options = append(opt.TmpfsOptions.Options, []string{key, value})
		return nil
	}
}

// WittVolumeNoCopy sets the no copy flag of the volume mount to true
func WithVolumeNoCopy() SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.VolumeOptions == nil {
			opt.VolumeOptions = &mount.VolumeOptions{}
		}
		opt.VolumeOptions.NoCopy = true
		return nil
	}
}

// WithVolumeLabel sets the label of the volume mount
// parameters:
//   - key: the key of the label
//   - value: the value of the label
func WithVolumeLabel(key string, value string) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.VolumeOptions == nil {
			opt.VolumeOptions = &mount.VolumeOptions{}
		}
		if opt.VolumeOptions.Labels == nil {
			opt.VolumeOptions.Labels = make(map[string]string)
		}
		opt.VolumeOptions.Labels[key] = value
		return nil
	}
}

// WithVolumeSubPath sets the sub path of the volume mount
// parameters:
//   - subPath: the sub path of the volume will mount the sub directory of the volume
func WithVolumeSubPath(subPath string) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.VolumeOptions == nil {
			opt.VolumeOptions = &mount.VolumeOptions{}
		}
		opt.VolumeOptions.Subpath = subPath
		return nil
	}
}

// WithVolumeDriver sets the driver of the volume mount
// parameters:
//   - driver: the driver of the volume mount
//   - options: the options of the volume mount passed directly to the linux kernel mount -o
//   - device: the device of the volume mount
func WithVolumeDriver(driver string, options string, device string) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.VolumeOptions == nil {
			opt.VolumeOptions = &mount.VolumeOptions{}
		}
		if opt.VolumeOptions.DriverConfig == nil {
			opt.VolumeOptions.DriverConfig = &mount.Driver{}
		}
		opt.VolumeOptions.DriverConfig.Name = driver
		opt.VolumeOptions.DriverConfig.Options = map[string]string{
			"o":      options,
			"device": device,
		}
		return nil
	}
}

// WithBindPropagation sets the propagation of the bind mount
// parameters:
//   - propagation: the propagation of the bind mount
func WithBindPropagation(propagation Propagation) SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.BindOptions == nil {
			opt.BindOptions = &mount.BindOptions{}
		}
		opt.BindOptions.Propagation = mount.Propagation(propagation)
		return nil
	}
}

// WithBindNonRecursive sets the non recursive flag of the bind mount to true
func WithBindNonRecursive() SetMountConfig {
	return func(opt *mount.Mount) error {
		opt.BindOptions.NonRecursive = true
		return nil
	}
}

// WithBindReadOnlyNonRecursive sets the read only non recursive flag of the bind mount to true
func WithBindReadOnlyNonRecursive() SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.BindOptions == nil {
			opt.BindOptions = &mount.BindOptions{}
		}
		opt.BindOptions.ReadOnlyNonRecursive = true
		return nil
	}
}

// WithBindReadOnlyForceRecursive sets the read only force recursive flag of the bind mount to true
// ReadOnlyForceRecursive raises an error if the mount cannot be made recursively read-only.
func WithBindReadOnlyForceRecursive() SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.BindOptions == nil {
			opt.BindOptions = &mount.BindOptions{}
		}
		opt.BindOptions.ReadOnlyForceRecursive = true
		return nil
	}
}

// WithBindCreateMountpoint creates a mountpoint for the bind mount
func WithBindCreateMountpoint() SetMountConfig {
	return func(opt *mount.Mount) error {
		if opt.BindOptions == nil {
			opt.BindOptions = &mount.BindOptions{}
		}
		opt.BindOptions.CreateMountpoint = true
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the mount
// and append the error to the host config error collection
func Fail(err error) SetMountConfig {
	return func(opt *mount.Mount) error {
		return create.NewContainerConfigError("mount", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the mount
// and append the error to the host config error collection
func Failf(stringFormat string, args ...interface{}) SetMountConfig {
	return func(opt *mount.Mount) error {
		return create.NewContainerConfigError("mount", fmt.Sprintf(stringFormat, args...))
	}
}
