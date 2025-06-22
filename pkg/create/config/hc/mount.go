package hc

import (
	"os"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
)

// WithRWHostBindMount creates a read-write host bound mount
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
func WithRWHostBindMount(source string, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeBind),
		mount.WithMountSource(source),
		mount.WithMountTarget(target),
		mount.WithMountReadWrite(),
	)
}

// WithROHostBindMount creates a readonly host bound mount
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
func WithROHostBindMount(source string, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeBind),
		mount.WithMountSource(source),
		mount.WithMountTarget(target),
		mount.WithMountReadOnly(),
	)
}

// WithTmpfsMount creates a tmpfs mount
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
//   - mode: the mode of the tmpfs mount
func WithTmpfsMount(target string, sizeBytes int, mode os.FileMode) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeTmpfs),
		mount.WithMountTarget(target),
		mount.WithMountTmpfsSizeBytes(sizeBytes),
		mount.WithMountTmpfsMode(mode),
	)
}

// WithRONamedVolumeMount creates a readonly named volume mount
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
func WithRONamedVolumeMount(name, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeVolume),
		mount.WithMountSource(name),
		mount.WithMountTarget(target),
		mount.WithMountReadOnly(),
	)
}

// WithRWNamedVolumeMount creates a read-write named volume mount
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
func WithRWNamedVolumeMount(name, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeVolume),
		mount.WithMountSource(name),
		mount.WithMountTarget(target),
		mount.WithMountReadWrite(),
	)
}

// WithHostBindMountRecursiveReadOnly creates a readonly host bound mount with recursive read only
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
func WithHostBindMountRecursiveReadOnly(source, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeBind),
		mount.WithMountSource(source),
		mount.WithMountTarget(target),
		mount.WithMountReadOnly(),
		mount.WithMountBindReadOnlyForceRecursive(),
	)
}

// WithTmpfsMountUIDGID creates a tmpfs mount with uid and gid
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
//   - uid: the uid of the tmpfs mount
//   - gid: the gid of the tmpfs mount
func WithTmpfsMountUIDGID(target string, sizeBytes int, uid, gid string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeTmpfs),
		mount.WithMountTarget(target),
		mount.WithMountTmpfsSizeBytes(sizeBytes),
		mount.WithMountTmpfsKeyValue("uid", uid),
		mount.WithMountTmpfsKeyValue("gid", gid),
	)
}

// WithRWNamedVolumeMountWithLabel creates a read-write named volume with a label
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
//   - labelKey: the key of the label
//   - labelValue: the value of the label
func WithRWNamedVolumeMountWithLabel(name, target, labelKey, labelValue string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeVolume),
		mount.WithMountSource(name),
		mount.WithMountTarget(target),
		mount.WithMountReadWrite(),
		mount.WithMountVolumeLabel(labelKey, labelValue),
	)
}

// WithRWNamedVolumeSubPath creates a read-write named volume with a subpath
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
//   - subPath: the subpath of the volume
func WithRWNamedVolumeSubPath(name, target, subPath string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeVolume),
		mount.WithMountSource(name),
		mount.WithMountTarget(target),
		mount.WithMountReadWrite(),
		mount.WithMountVolumeSubPath(subPath),
	)
}

// WithBindMountWithPropagation creates a host bind mount with custom propagation
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
//   - propagation: the propagation of the mount
func WithBindMountWithPropagation(source, target string, propagation mount.Propagation) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeBind),
		mount.WithMountSource(source),
		mount.WithMountTarget(target),
		mount.WithMountReadWrite(),
		mount.WithMountBindPropagation(propagation),
	)
}

// WithNamedVolumeWithDriver creates a named volume with a specified driver and options
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
//   - driver: the driver of the volume
//   - options: the options of the volume
//   - device: the device of the volume
func WithNamedVolumeWithDriver(name, target, driver, options, device string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeVolume),
		mount.WithMountSource(name),
		mount.WithMountTarget(target),
		mount.WithMountReadWrite(),
		mount.WithMountVolumeDriver(driver, options, device),
	)
}

// WithTmpfsMountExec creates a tmpfs mount with exec flag
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
func WithTmpfsMountExec(target string, sizeBytes int) create.SetHostConfig {
	return WithMountPoint(
		mount.WithMountType(mount.MountTypeTmpfs),
		mount.WithMountTarget(target),
		mount.WithMountTmpfsSizeBytes(sizeBytes),
		mount.WithMountTmpfsFlag("exec"),
	)
}

// WithTmpfsMountCustomOptions creates a tmpfs mount with arbitrary options
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
//   - flags: the flags of the tmpfs mount
func WithTmpfsMountCustomOptions(target string, sizeBytes int, flags ...[]string) create.SetHostConfig {

	opts := []mount.SetMountConfig{
		mount.WithMountType(mount.MountTypeTmpfs),
		mount.WithMountTarget(target),
		mount.WithMountTmpfsSizeBytes(sizeBytes),
	}
	for _, flag := range flags {
		if len(flag) == 1 {
			opts = append(opts, mount.WithMountTmpfsFlag(flag[0]))
		} else if len(flag) == 2 {
			opts = append(opts, mount.WithMountTmpfsKeyValue(flag[0], flag[1]))
		}
	}
	return WithMountPoint(opts...)
}

// WithNonRecursiveBindMount creates a non-recursive host bind mount
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
//   - readonly: true if the mount should be read only, false otherwise
func WithNonRecursiveBindMount(source, target string, readonly bool) create.SetHostConfig {
	opts := []mount.SetMountConfig{
		mount.WithMountType(mount.MountTypeBind),
		mount.WithMountSource(source),
		mount.WithMountTarget(target),
		mount.WithMountBindNonRecursive(),
	}
	if readonly {
		opts = append(opts, mount.WithMountReadOnly())
	} else {
		opts = append(opts, mount.WithMountReadWrite())
	}
	return WithMountPoint(opts...)
}
