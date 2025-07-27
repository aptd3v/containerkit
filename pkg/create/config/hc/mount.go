package hc

import (
	"os"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/hc/mount"
	"github.com/aptd3v/go-contain/pkg/tools"
)

// WithRWHostBindMount creates a read-write host bound mount
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
func WithRWHostBindMount(source string, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeBind),
		mount.WithSource(source),
		mount.WithTarget(target),
		mount.WithReadWrite(),
	)
}

// WithROHostBindMount creates a readonly host bound mount
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
func WithROHostBindMount(source string, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeBind),
		mount.WithSource(source),
		mount.WithTarget(target),
		mount.WithReadOnly(),
	)
}

// WithTmpfsMount creates a tmpfs mount
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
//   - mode: the mode of the tmpfs mount
func WithTmpfsMount(target string, sizeBytes int, mode os.FileMode) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeTmpfs),
		mount.WithTarget(target),
		mount.WithTmpfsSizeBytes(sizeBytes),
		mount.WithTmpfsMode(mode),
	)
}

// WithRONamedVolumeMount creates a readonly named volume mount
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
func WithRONamedVolumeMount(name, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeVolume),
		mount.WithSource(name),
		mount.WithTarget(target),
		mount.WithReadOnly(),
	)
}

// WithRWNamedVolumeMount creates a read-write named volume mount
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
func WithRWNamedVolumeMount(name, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeVolume),
		mount.WithSource(name),
		mount.WithTarget(target),
		mount.WithReadWrite(),
	)
}

// WithHostBindMountRecursiveRO creates a readonly host bound mount with recursive read only
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
func WithHostBindMountRecursiveRO(source, target string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeBind),
		mount.WithSource(source),
		mount.WithTarget(target),
		mount.WithReadOnly(),
		mount.WithBindReadOnlyForceRecursive(),
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
		mount.WithType(mount.MountTypeTmpfs),
		mount.WithTarget(target),
		mount.WithTmpfsSizeBytes(sizeBytes),
		mount.WithTmpfsKeyValue("uid", uid),
		mount.WithTmpfsKeyValue("gid", gid),
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
		mount.WithType(mount.MountTypeVolume),
		mount.WithSource(name),
		mount.WithTarget(target),
		mount.WithReadWrite(),
		mount.WithVolumeLabel(labelKey, labelValue),
	)
}

// WithRWNamedVolumeSubPath creates a read-write named volume with a subpath
// parameters:
//   - name: the name of the volume
//   - target: the target of the mount
//   - subPath: the subpath of the volume
func WithRWNamedVolumeSubPath(name, target, subPath string) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeVolume),
		mount.WithSource(name),
		mount.WithTarget(target),
		mount.WithReadWrite(),
		mount.WithVolumeSubPath(subPath),
	)
}

// WithBindMountWithPropagation creates a host bind mount with custom propagation
// parameters:
//   - source: the source of the mount
//   - target: the target of the mount
//   - propagation: the propagation of the mount
func WithBindMountWithPropagation(source, target string, propagation mount.Propagation) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeBind),
		mount.WithSource(source),
		mount.WithTarget(target),
		mount.WithReadWrite(),
		mount.WithBindPropagation(propagation),
	)
}

// WithTmpfsMountExec creates a tmpfs mount with exec flag
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
func WithTmpfsMountExec(target string, sizeBytes int) create.SetHostConfig {
	return WithMountPoint(
		mount.WithType(mount.MountTypeTmpfs),
		mount.WithTarget(target),
		mount.WithTmpfsSizeBytes(sizeBytes),
		mount.WithTmpfsFlag("exec"),
	)
}

// WithTmpfsMountCustomOptions creates a tmpfs mount with arbitrary options
// parameters:
//   - target: the target of the mount
//   - sizeBytes: the size of the tmpfs mount
//   - flags: the flags of the tmpfs mount
func WithTmpfsMountCustomOptions(target string, sizeBytes int, flags ...[]string) create.SetHostConfig {

	opts := []mount.SetMountConfig{
		mount.WithType(mount.MountTypeTmpfs),
		mount.WithTarget(target),
		mount.WithTmpfsSizeBytes(sizeBytes),
	}
	for _, flag := range flags {
		if len(flag) == 1 {
			opts = append(opts, mount.WithTmpfsFlag(flag[0]))
		} else if len(flag) == 2 {
			opts = append(opts, mount.WithTmpfsKeyValue(flag[0], flag[1]))
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
	return WithMountPoint(
		mount.WithType(mount.MountTypeBind),
		mount.WithSource(source),
		mount.WithTarget(target),
		mount.WithBindNonRecursive(),
		tools.WhenTrueElse(readonly, mount.WithReadOnly(), mount.WithReadWrite()),
	)
}
