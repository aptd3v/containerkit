package kit

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/propagate"
)

func (v *volume) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("volume", err)
	}
	v.source.assign(name)
	if !v.kind.ok() {
		v.kind.assign(NamedVolume)
	}
	return nil
}

func (v *volume) setSource(s string) error {
	if s == "" {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: volume source", errEmpty))
	}
	v.source.assign(s)
	return nil
}

func (v *volume) setTarget(t string) error {
	if t == "" {
		return cfgErr("", "target", "Target", fmt.Errorf("%w: volume target", errEmpty))
	}
	v.target.assign(t)
	return nil
}

func (v *volume) setReadOnly(ro bool) error {
	v.readOnly.assign(ro)
	return nil
}

func (v *volume) setDriver(d string) error {
	if d == "" {
		return cfgErr("", "driver", "Driver", fmt.Errorf("%w: volume driver", errEmpty))
	}
	v.driver.assign(d)
	return nil
}

func (v *volume) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if v.labels == nil {
		v.labels = map[string]string{}
	}
	v.labels[k] = val
	return nil
}

func (v *volume) setKind(k VolumeKind) error {
	switch k {
	case Bind, NamedVolume, Tmpfs, NamedPipe, Cluster, ImageVolume, "":
		v.kind.assign(k)
		return nil
	default:
		return cfgErr("", "type", "Kind", fmt.Errorf("%w: volume type %q", errInvalid, k))
	}
}

func (v *volume) setSize(n int64) error {
	if n < 0 {
		return cfgErr("", "size", "Size", fmt.Errorf("%w: tmpfs size", errInvalid))
	}
	v.size.assign(n)
	return nil
}

func (v volume) validate() error {
	if !v.target.ok() {
		return cfgErr("", "target", "Target", fmt.Errorf("%w: volume target", errRequired))
	}
	kind := v.kindOrDefault()
	if kind == Bind && !v.source.ok() {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: bind source", errRequired))
	}
	if kind == Tmpfs && v.source.ok() {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: tmpfs cannot have a source", errInvalid))
	}
	return nil
}

func (v volume) kindOrDefault() VolumeKind {
	if v.kind.ok() && v.kind.val != "" {
		return v.kind.val
	}
	if v.source.ok() {
		s := v.source.val
		if len(s) > 0 && (s[0] == '/' || s[0] == '.' || s[0] == '\\') {
			return Bind
		}
		return NamedVolume
	}
	if v.size.ok() {
		return Tmpfs
	}
	return NamedVolume
}

// KindOpt is the value returned by [Kind].
// It implements [VolumeOption].
type KindOpt struct{ k VolumeKind }

// Kind sets the mount type inside [Volume] (bind, volume, tmpfs, ...).
func Kind(k VolumeKind) KindOpt { return KindOpt{k: k} }

func (o KindOpt) applyVolume(v *volume) error { return v.setKind(o.k) }

// SizeOpt is the value returned by [Size].
// It implements [VolumeOption].
type SizeOpt struct{ n int64 }

// Size sets tmpfs size in bytes inside [Volume].
func Size(n int64) SizeOpt { return SizeOpt{n: n} }

func (o SizeOpt) applyVolume(v *volume) error { return v.setSize(o.n) }

// VolumeOpt is the value returned by [Volume].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type VolumeOpt struct{ opts []VolumeOption }

// Volume adds a mount shared by Engine HostConfig.Mounts and Compose volumes.
func Volume(opts ...VolumeOption) VolumeOpt { return VolumeOpt{opts: opts} }

func (o VolumeOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o VolumeOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o VolumeOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o VolumeOpt) apply(s *spec) error {
	var v volume
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("volume.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyVolume(&v); err != nil {
			errs = append(errs, wrapPath("volume", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if err := v.validate(); err != nil {
		return wrapPath("volume", err)
	}
	s.volumes = append(s.volumes, v)
	return nil
}

// VolumesFrom mounts volumes from another container.
func VolumesFrom(name string) SpecOpt {
	return specFn(func(s *spec) error {
		if err := validateName(name); err != nil {
			return wrapPath("volumes_from", err)
		}
		s.volumesFrom = append(s.volumesFrom, name)
		return nil
	})
}

// VolumeDriver sets the volume driver name.
func VolumeDriver(d string) SpecOpt {
	return specFn(func(s *spec) error {
		if d == "" {
			return cfgErr("", "volume_driver", "VolumeDriver", fmt.Errorf("%w: volume driver", errEmpty))
		}
		s.volumeDrv.assign(d)
		return nil
	})
}

// SubPathOpt is the value returned by [SubPath].
// It implements [VolumeOption].
type SubPathOpt struct{ p string }

// SubPath sets a volume subpath inside [Volume].
func SubPath(p string) SubPathOpt { return SubPathOpt{p: p} }

func (o SubPathOpt) applyVolume(v *volume) error {
	v.subPath.assign(o.p)
	return nil
}

// NoCopyOpt is the value returned by [NoCopy].
// It implements [VolumeOption].
type NoCopyOpt struct{ v bool }

// NoCopy disables copying image content into a named volume.
func NoCopy(v bool) NoCopyOpt { return NoCopyOpt{v: v} }

func (o NoCopyOpt) applyVolume(v *volume) error {
	v.noCopy.assign(o.v)
	return nil
}

// PropagationOpt is the value returned by [Propagation].
// It implements [VolumeOption].
type PropagationOpt struct{ p string }

// Propagation sets bind-mount propagation inside [Volume].
func Propagation(p propagate.Mode) PropagationOpt { return PropagationOpt{p: string(p)} }

func (o PropagationOpt) applyVolume(v *volume) error {
	v.prop.assign(o.p)
	return nil
}
