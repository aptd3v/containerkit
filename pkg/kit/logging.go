package kit

import "fmt"

func (l *logging) setDriver(d string) error {
	if d == "" {
		return cfgErr("", "driver", "Driver", fmt.Errorf("%w: log driver", errEmpty))
	}
	l.driver.assign(d)
	return nil
}

// DriverOptKV is the value returned by [DriverOpt].
// It implements [ComposeNetOption], [ComposeVolOption], [EngineNetworkOption],
// [EngineVolumeOption], [GPUOption], [LogOption], [NetworkOption], and [VolumeOption].
type DriverOptKV struct{ key, value string }

// DriverOpt adds a driver option inside [Logging], [Volume], [Network], or project resources.
func DriverOpt(key, value string) DriverOptKV { return DriverOptKV{key: key, value: value} }

func (o DriverOptKV) applyLog(l *logging) error {
	if o.key == "" {
		return cfgErr("", "option", "DriverOpt", fmt.Errorf("%w: log option key", errEmpty))
	}
	if l.options == nil {
		l.options = map[string]string{}
	}
	l.options[o.key] = o.value
	return nil
}

func (o DriverOptKV) applyVolume(v *volume) error {
	if o.key == "" {
		return cfgErr("", "option", "DriverOpt", fmt.Errorf("%w: volume option key", errEmpty))
	}
	if v.DriverValue == nil {
		v.DriverValue = map[string]string{}
	}
	v.DriverValue[o.key] = o.value
	return nil
}

func (o DriverOptKV) applyNetwork(n *netAttach) error {
	if o.key == "" {
		return cfgErr("", "option", "DriverOpt", fmt.Errorf("%w: network option key", errEmpty))
	}
	if n.driverOpts == nil {
		n.driverOpts = map[string]string{}
	}
	n.driverOpts[o.key] = o.value
	return nil
}

func (o DriverOptKV) applyComposeNet(n *projectNet) error {
	if n.driverOpts == nil {
		n.driverOpts = map[string]string{}
	}
	n.driverOpts[o.key] = o.value
	return nil
}

func (o DriverOptKV) applyComposeVol(v *projectVol) error {
	if v.driverOpts == nil {
		v.driverOpts = map[string]string{}
	}
	v.driverOpts[o.key] = o.value
	return nil
}

func (o DriverOptKV) applyGPU(g *gpu) error {
	// GPU options reuse DriverOpt as capability-adjacent extras; store as ids if key is id.
	if o.key == "id" && o.value != "" {
		g.ids = append(g.ids, o.value)
	}
	return nil
}

// LoggingOpt is the value returned by [Logging].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type LoggingOpt struct{ opts []LogOption }

// Logging sets Engine LogConfig / Compose logging.
func Logging(opts ...LogOption) LoggingOpt { return LoggingOpt{opts: opts} }

func (o LoggingOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o LoggingOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o LoggingOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o LoggingOpt) apply(s *spec) error {
	var l logging
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("logging.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyLog(&l); err != nil {
			errs = append(errs, wrapPath("logging", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !l.driver.ok() {
		return cfgErr("", "logging.driver", "Driver", fmt.Errorf("%w: log driver", errRequired))
	}
	s.logging.assign(l)
	return nil
}
