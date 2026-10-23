package containerkit

import (
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

// HealthCheck sets the container healthcheck from a Health value.
func (c *Container) HealthCheck(h Health) *Container {
	check, err := h.healthConfig()
	if err != nil {
		c.Errors = append(c.Errors, err)
		return c
	}
	c.Config.Container.Healthcheck = check
	return c
}

func (h Health) healthConfig() (*container.HealthConfig, error) {
	timeout, err := healthDuration(h.Timeout, h.TimeoutD, "timeout")
	if err != nil {
		return nil, err
	}
	interval, err := healthDuration(h.Interval, h.IntervalD, "interval")
	if err != nil {
		return nil, err
	}
	start, err := healthDuration(h.StartPeriod, h.StartPeriodD, "start period")
	if err != nil {
		return nil, err
	}
	if h.Retries < 0 {
		return nil, errdefs.NewContainerConfigError("healthcheck", "retries must be greater than or equal to 0")
	}
	test := h.Test
	if len(test) == 0 {
		test = []string{"NONE"}
	}
	return &container.HealthConfig{
		Test:        test,
		Timeout:     timeout,
		Interval:    interval,
		StartPeriod: start,
		Retries:     h.Retries,
	}, nil
}

func healthDuration(seconds int, raw, field string) (time.Duration, error) {
	var duration time.Duration
	if raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return 0, errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf("error parsing %s: %s", field, err))
		}
		duration = parsed
	} else {
		duration = time.Duration(seconds) * time.Second
	}
	if duration < 0 {
		return 0, errdefs.NewContainerConfigError("healthcheck", field+" must be non-negative")
	}
	return duration, nil
}

// DisabledHealthCheck disables the healthcheck.
func (c *Container) DisabledHealthCheck() *Container {
	c.Config.Container.Healthcheck = &container.HealthConfig{Test: []string{"NONE"}}
	return c
}

// RestartUnlessStopped is an alias for RestartPolicyUnlessStopped.
func (c *Container) RestartUnlessStopped() *Container {
	return c.RestartPolicyUnlessStopped()
}

// Mount adds a host mount point.
func (c *Container) Mount(m Mount) *Container {
	spec := mount.Mount{
		Type:        mount.Type(mountType(m)),
		Source:      m.Source,
		Target:      m.Target,
		ReadOnly:    m.ReadOnly,
		Consistency: mount.Consistency(m.Consistency),
	}
	if spec.Type == mount.TypeTmpfs && spec.Source != "" {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("mount", "tmpfs source must be empty"))
	}
	if m.Bind != nil {
		spec.BindOptions = &mount.BindOptions{
			Propagation:            mount.Propagation(m.Bind.Propagation),
			NonRecursive:           m.Bind.NonRecursive,
			ReadOnlyNonRecursive:   m.Bind.ReadOnlyNonRecursive,
			ReadOnlyForceRecursive: m.Bind.ReadOnlyForceRecursive,
		}
		if m.Bind.CreateHostPath != nil {
			spec.BindOptions.CreateMountpoint = *m.Bind.CreateHostPath
		}
		c.noteMount(m.Target, m.Bind)
	}
	if m.Volume != nil {
		spec.VolumeOptions = &mount.VolumeOptions{
			NoCopy:  m.Volume.NoCopy,
			Subpath: m.Volume.Subpath,
			Labels:  maps.Clone(m.Volume.Labels),
		}
		if m.Volume.Driver != "" || len(m.Volume.DriverOptions) > 0 {
			spec.VolumeOptions.DriverConfig = &mount.Driver{
				Name:    m.Volume.Driver,
				Options: maps.Clone(m.Volume.DriverOptions),
			}
		}
	}
	if m.Tmpfs != nil {
		spec.TmpfsOptions = &mount.TmpfsOptions{
			SizeBytes: m.Tmpfs.SizeBytes,
			Mode:      m.Tmpfs.Mode,
			Options:   m.Tmpfs.Options,
		}
	}
	if m.Image != nil {
		spec.ImageOptions = &mount.ImageOptions{Subpath: m.Image.Subpath}
	}
	c.Config.Host.Mounts = append(c.Config.Host.Mounts, spec)
	return c
}

func mountType(m Mount) MountType {
	if m.Type != "" {
		return m.Type
	}
	switch {
	case m.Image != nil:
		return MountImage
	case m.Tmpfs != nil:
		return MountTmpfs
	case m.Volume != nil:
		return MountVolume
	default:
		return MountBind
	}
}

func (c *Container) noteMount(target string, bind *BindOptions) {
	if bind.SELinux != "" && bind.SELinux != SELinuxShared && bind.SELinux != SELinuxPrivate {
		c.Errors = append(c.Errors, errdefs.NewHostConfigError("mount", "selinux must be z or Z"))
		bind = &BindOptions{CreateHostPath: bind.CreateHostPath}
	}
	if bind.SELinux == "" && bind.CreateHostPath == nil {
		return
	}
	if c.mountNotes == nil {
		c.mountNotes = map[string]mountNote{}
	}
	note := c.mountNotes[target]
	if bind.SELinux != "" {
		note.selinux = bind.SELinux
	}
	if bind.CreateHostPath != nil {
		note.createHostPath = bind.CreateHostPath
	}
	c.mountNotes[target] = note
}

// Endpoint joins the named network. Optional EndpointOpts customize the endpoint.
func (c *Container) Endpoint(name string, opts ...EndpointOpts) *Container {
	cfg := c.Config.Network
	if cfg.EndpointsConfig == nil {
		cfg.EndpointsConfig = map[string]*network.EndpointSettings{}
	}
	ep := cfg.EndpointsConfig[name]
	if ep == nil {
		ep = &network.EndpointSettings{}
		cfg.EndpointsConfig[name] = ep
	}
	if len(opts) == 0 {
		return c
	}
	o := opts[0]
	if len(o.Aliases) > 0 {
		ep.Aliases = append(ep.Aliases, o.Aliases...)
	}
	if len(o.Links) > 0 {
		ep.Links = append(ep.Links, o.Links...)
	}
	if o.MacAddress != "" {
		ep.MacAddress = o.MacAddress
	}
	if len(o.DriverOptions) > 0 {
		if ep.DriverOpts == nil {
			ep.DriverOpts = map[string]string{}
		}
		maps.Copy(ep.DriverOpts, o.DriverOptions)
	}
	if o.IPv4 != "" || o.IPv6 != "" || len(o.LinkLocalIPs) > 0 {
		if ep.IPAMConfig == nil {
			ep.IPAMConfig = &network.EndpointIPAMConfig{}
		}
		if o.IPv4 != "" {
			ep.IPAMConfig.IPv4Address = o.IPv4
		}
		if o.IPv6 != "" {
			ep.IPAMConfig.IPv6Address = o.IPv6
		}
		if len(o.LinkLocalIPs) > 0 {
			ep.IPAMConfig.LinkLocalIPs = append(ep.IPAMConfig.LinkLocalIPs, o.LinkLocalIPs...)
		}
	}
	return c
}

func (c *Container) RWHostBindMount(source string, target string) *Container {
	return c.Mount(Mount{Source: source, Target: target, Type: MountBind})
}

func (c *Container) ROHostBindMount(source string, target string) *Container {
	return c.Mount(Mount{Source: source, Target: target, Type: MountBind, ReadOnly: true})
}

func (c *Container) TmpfsMount(target string, sizeBytes int, mode os.FileMode) *Container {
	return c.Mount(Mount{
		Target: target,
		Type:   MountTmpfs,
		Tmpfs:  &TmpfsOptions{SizeBytes: int64(sizeBytes), Mode: mode},
	})
}

func (c *Container) RONamedVolumeMount(name, target string) *Container {
	return c.Mount(Mount{Source: name, Target: target, Type: MountVolume, ReadOnly: true})
}

func (c *Container) RWNamedVolumeMount(name, target string) *Container {
	return c.Mount(Mount{Source: name, Target: target, Type: MountVolume})
}

func (c *Container) HostBindMountRecursiveRO(source, target string) *Container {
	return c.Mount(Mount{
		Source:   source,
		Target:   target,
		Type:     MountBind,
		ReadOnly: true,
		Bind:     &BindOptions{ReadOnlyForceRecursive: true},
	})
}

func (c *Container) TmpfsMountUIDGID(target string, sizeBytes int, uid, gid string) *Container {
	return c.Mount(Mount{
		Target: target,
		Type:   MountTmpfs,
		Tmpfs: &TmpfsOptions{
			SizeBytes: int64(sizeBytes),
			Options:   [][]string{{"uid", uid}, {"gid", gid}},
		},
	})
}

func (c *Container) RWNamedVolumeMountWithLabel(name, target, labelKey, labelValue string) *Container {
	return c.Mount(Mount{
		Source: name,
		Target: target,
		Type:   MountVolume,
		Volume: &VolumeOptions{Labels: map[string]string{labelKey: labelValue}},
	})
}

func (c *Container) RWNamedVolumeSubPath(name, target, subPath string) *Container {
	return c.Mount(Mount{
		Source: name,
		Target: target,
		Type:   MountVolume,
		Volume: &VolumeOptions{Subpath: subPath},
	})
}

func (c *Container) BindMountWithPropagation(source, target string, propagation MountPropagation) *Container {
	return c.Mount(Mount{
		Source: source,
		Target: target,
		Type:   MountBind,
		Bind:   &BindOptions{Propagation: propagation},
	})
}

func (c *Container) TmpfsMountExec(target string, sizeBytes int) *Container {
	return c.Mount(Mount{
		Target: target,
		Type:   MountTmpfs,
		Tmpfs:  &TmpfsOptions{SizeBytes: int64(sizeBytes), Options: [][]string{{"exec"}}},
	})
}

func (c *Container) TmpfsMountCustomOptions(target string, sizeBytes int, flags ...[]string) *Container {
	opts := make([][]string, 0, len(flags))
	for _, flag := range flags {
		if len(flag) == 1 || len(flag) == 2 {
			opts = append(opts, flag)
		}
	}
	return c.Mount(Mount{
		Target: target,
		Type:   MountTmpfs,
		Tmpfs:  &TmpfsOptions{SizeBytes: int64(sizeBytes), Options: opts},
	})
}

func (c *Container) NonRecursiveBindMount(source, target string, readonly bool) *Container {
	return c.Mount(Mount{
		Source:   source,
		Target:   target,
		Type:     MountBind,
		ReadOnly: readonly,
		Bind:     &BindOptions{NonRecursive: true},
	})
}
