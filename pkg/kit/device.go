package kit

import "fmt"

func (d *device) setHost(p string) error {
	if p == "" {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: device host path", errEmpty))
	}
	d.host.assign(p)
	return nil
}

func (d *device) setContainer(p string) error {
	if p == "" {
		return cfgErr("", "target", "Target", fmt.Errorf("%w: device container path", errEmpty))
	}
	d.container.assign(p)
	return nil
}

func (d *device) setPerms(p string) error {
	d.perms.assign(p)
	return nil
}

// PermsOpt is the value returned by [Perms].
// It implements [DeviceOption].
type PermsOpt struct{ p string }

// Perms sets device cgroup permissions inside [Device] (e.g. "rwm").
func Perms(p string) PermsOpt { return PermsOpt{p: p} }

func (o PermsOpt) applyDevice(d *device) error { return d.setPerms(o.p) }

// DeviceOpt is the value returned by [Device].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type DeviceOpt struct{ opts []DeviceOption }

// Device maps a host device into the container (Engine Devices / Compose devices).
func Device(opts ...DeviceOption) DeviceOpt { return DeviceOpt{opts: opts} }

func (o DeviceOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o DeviceOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o DeviceOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o DeviceOpt) apply(s *spec) error {
	var d device
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("device.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyDevice(&d); err != nil {
			errs = append(errs, wrapPath("device", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !d.host.ok() {
		return cfgErr("", "device.source", "Source", fmt.Errorf("%w: device host path", errRequired))
	}
	s.devices = append(s.devices, d)
	return nil
}

// DeviceCgroupRule adds a device cgroup rule.
func DeviceCgroupRule(rule string) SpecOpt {
	return specFn(func(s *spec) error {
		if rule == "" {
			return cfgErr("", "device_cgroup_rule", "DeviceCgroupRule", fmt.Errorf("%w: rule", errEmpty))
		}
		s.devCgroup = append(s.devCgroup, rule)
		return nil
	})
}

func (g *gpu) setDriver(d string) error {
	g.driver.assign(d)
	return nil
}

// CountOpt is the value returned by [Count].
// It implements [GPUOption].
type CountOpt struct{ n int }

// Count sets GPU count inside [GPU]. -1 means all.
func Count(n int) CountOpt { return CountOpt{n: n} }

func (o CountOpt) applyGPU(g *gpu) error {
	g.count.assign(o.n)
	return nil
}

// CapabilitiesOpt is the value returned by [Capabilities].
// It implements [GPUOption].
type CapabilitiesOpt struct{ c string }

// Capabilities adds a GPU capability (typically "gpu") inside [GPU].
func Capabilities(c string) CapabilitiesOpt { return CapabilitiesOpt{c: c} }

func (o CapabilitiesOpt) applyGPU(g *gpu) error {
	if o.c == "" {
		return cfgErr("", "capabilities", "Capabilities", fmt.Errorf("%w: gpu capability", errEmpty))
	}
	g.capabilities = append(g.capabilities, o.c)
	return nil
}

// GPUOpt is the value returned by [GPU].
// It implements [ContainerOption], [DeployOption], [JobOption], [ReserveOption], and
// [ServiceOption].
type GPUOpt struct{ opts []GPUOption }

// GPU requests GPUs (Engine DeviceRequests / Compose gpus).
func GPU(opts ...GPUOption) GPUOpt { return GPUOpt{opts: opts} }

func (o GPUOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o GPUOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o GPUOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o GPUOpt) collect() (gpu, error) {
	var g gpu
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("gpu.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyGPU(&g); err != nil {
			errs = append(errs, wrapPath("gpu", err))
		}
	}
	if err := join(errs...); err != nil {
		return gpu{}, err
	}
	return g, nil
}

func (o GPUOpt) apply(s *spec) error {
	g, err := o.collect()
	if err != nil {
		return err
	}
	s.gpus = append(s.gpus, g)
	return nil
}

func (o GPUOpt) applyDeploy(d *deploy) error {
	g, err := o.collect()
	if err != nil {
		return err
	}
	d.devices = append(d.devices, g)
	return nil
}

func (o GPUOpt) applyReserve(r *resource) error {
	g, err := o.collect()
	if err != nil {
		return err
	}
	r.devices = append(r.devices, g)
	return nil
}
