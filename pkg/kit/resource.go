package kit

import "fmt"

// MemoryReservation sets the soft memory limit in bytes.
func MemoryReservation(bytes int64) SpecOpt {
	return specFn(func(s *spec) error {
		if bytes < 0 {
			return cfgErr("", "memory_reservation", "MemoryReservation", fmt.Errorf("%w: memory reservation", errInvalid))
		}
		s.memReserve.assign(bytes)
		return nil
	})
}

// MemorySwap sets memory+swap in bytes. -1 is unlimited.
func MemorySwap(bytes int64) SpecOpt {
	return specFn(func(s *spec) error {
		if bytes < -1 {
			return cfgErr("", "memory_swap", "MemorySwap", fmt.Errorf("%w: memory swap", errInvalid))
		}
		s.memSwap.assign(bytes)
		return nil
	})
}

// MemorySwappiness sets vm.swappiness (0–100).
func MemorySwappiness(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 || n > 100 {
			return cfgErr("", "memory_swappiness", "MemorySwappiness", fmt.Errorf("%w: swappiness %d", errInvalid, n))
		}
		s.memSwapiness.assign(n)
		return nil
	})
}

// ShmSize sets /dev/shm size in bytes.
func ShmSize(bytes int64) SpecOpt {
	return specFn(func(s *spec) error {
		if bytes < 0 {
			return cfgErr("", "shm_size", "ShmSize", fmt.Errorf("%w: shm size", errInvalid))
		}
		s.shmSize.assign(bytes)
		return nil
	})
}

// CPUsOpt is the value returned by [CPUs].
// It implements [ContainerOption], [DeployOption], [JobOption], [ReserveOption], and
// [ServiceOption].
type CPUsOpt struct{ n float64 }

// CPUs sets CPU quota as a fractional count (Compose cpus / Engine NanoCPUs).
// Inside [Deploy] it maps to deploy.resources.limits.cpus.
func CPUs(n float64) CPUsOpt { return CPUsOpt{n: n} }

func (o CPUsOpt) applyContainer(c *Container) error { return c.setCPUs(o.n) }
func (o CPUsOpt) applyService(s *Service) error     { return s.setCPUs(o.n) }
func (o CPUsOpt) applyJob(j *Job) error             { return j.setCPUs(o.n) }
func (o CPUsOpt) applyDeploy(d *deploy) error {
	if o.n < 0 {
		return cfgErr("", "cpus", "CPUs", fmt.Errorf("%w: deploy cpus", errInvalid))
	}
	d.cpus.assign(o.n)
	return nil
}

// CPUShares sets relative CPU weight.
func CPUShares(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_shares", "CPUShares", fmt.Errorf("%w: cpu shares", errInvalid))
		}
		s.cpuShares.assign(n)
		return nil
	})
}

// CPUPeriod sets CFS period in microseconds.
func CPUPeriod(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_period", "CPUPeriod", fmt.Errorf("%w: cpu period", errInvalid))
		}
		s.cpuPeriod.assign(n)
		return nil
	})
}

// CPUQuota sets CFS quota in microseconds.
func CPUQuota(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_quota", "CPUQuota", fmt.Errorf("%w: cpu quota", errInvalid))
		}
		s.cpuQuota.assign(n)
		return nil
	})
}

// CPUSet sets CpusetCpus (e.g. "0-3").
func CPUSet(set string) SpecOpt {
	return specFn(func(s *spec) error {
		if set == "" {
			return cfgErr("", "cpuset", "CPUSet", fmt.Errorf("%w: cpuset", errEmpty))
		}
		s.cpusetCpus.assign(set)
		return nil
	})
}

// CPUSetMems sets CpusetMems.
func CPUSetMems(set string) SpecOpt {
	return specFn(func(s *spec) error {
		if set == "" {
			return cfgErr("", "cpuset_mems", "CPUSetMems", fmt.Errorf("%w: cpuset mems", errEmpty))
		}
		s.cpusetMems.assign(set)
		return nil
	})
}

// IOMaximumIOps sets Windows system-drive IOps limit.
func IOMaximumIOps(n uint64) SpecOpt {
	return specFn(func(s *spec) error {
		s.ioMaxIOps.assign(n)
		return nil
	})
}

// IOMaximumBandwidth sets Windows system-drive byte/s limit.
func IOMaximumBandwidth(n uint64) SpecOpt {
	return specFn(func(s *spec) error {
		s.ioMaxBandwidth.assign(n)
		return nil
	})
}

// ConsoleSizeOpt is the value returned by [ConsoleSize].
// It implements [ContainerOption], [EngineExecAttachOption], [EngineExecCreateOption],
// [EngineExecResizeOption], [EngineResizeOption], [JobOption], and [ServiceOption].
type ConsoleSizeOpt struct{ h, w uint }

// ConsoleSize sets the initial console height and width.
func ConsoleSize(height, width uint) ConsoleSizeOpt { return ConsoleSizeOpt{h: height, w: width} }

func (o ConsoleSizeOpt) applyContainer(c *Container) error { return c.setConsoleSize(o.h, o.w) }
func (o ConsoleSizeOpt) applyService(s *Service) error     { return s.setConsoleSize(o.h, o.w) }
func (o ConsoleSizeOpt) applyJob(j *Job) error             { return j.setConsoleSize(o.h, o.w) }

// CPURealtimePeriod sets CPU realtime period in microseconds.
func CPURealtimePeriod(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_rt_period", "CPURealtimePeriod", fmt.Errorf("%w: cpu rt period", errInvalid))
		}
		s.cpuRTPeriod.assign(n)
		return nil
	})
}

// CPURealtimeRuntime sets CPU realtime runtime in microseconds.
func CPURealtimeRuntime(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_rt_runtime", "CPURealtimeRuntime", fmt.Errorf("%w: cpu rt runtime", errInvalid))
		}
		s.cpuRTRuntime.assign(n)
		return nil
	})
}

// CPUCount sets Windows CPU count.
func CPUCount(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_count", "CPUCount", fmt.Errorf("%w: cpu count", errInvalid))
		}
		s.cpuCount.assign(n)
		return nil
	})
}

// CPUPercent sets Windows CPU percent.
func CPUPercent(n int64) SpecOpt {
	return specFn(func(s *spec) error {
		if n < 0 {
			return cfgErr("", "cpu_percent", "CPUPercent", fmt.Errorf("%w: cpu percent", errInvalid))
		}
		s.cpuPercent.assign(n)
		return nil
	})
}

// PidsLimitOpt is the value returned by [PidsLimit].
// It implements [ContainerOption], [DeployOption], [JobOption], [ReserveOption], and
// [ServiceOption].
type PidsLimitOpt struct{ n int64 }

// PidsLimit sets the pids cgroup limit. 0 or -1 is unlimited.
// Inside [Deploy] / [Reserve] it maps to deploy.resources pids.
func PidsLimit(n int64) PidsLimitOpt { return PidsLimitOpt{n: n} }

func (o PidsLimitOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o PidsLimitOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o PidsLimitOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o PidsLimitOpt) apply(s *spec) error {
	s.pidsLimit.assign(o.n)
	return nil
}

func (o PidsLimitOpt) applyDeploy(d *deploy) error {
	d.pids.assign(o.n)
	return nil
}

func (o PidsLimitOpt) applyReserve(r *resource) error {
	r.pids.assign(o.n)
	return nil
}

// OomScoreAdj sets OOM score adjustment.
func OomScoreAdj(n int) SpecOpt {
	return specFn(func(s *spec) error {
		s.oomScoreAdj.assign(n)
		return nil
	})
}

// OomKillDisable disables the OOM killer.
func OomKillDisable(v bool) SpecOpt {
	return specFn(func(s *spec) error {
		s.oomKillDisable.assign(v)
		return nil
	})
}

// BlkioWeight sets block IO weight (10–1000).
func BlkioWeight(n uint16) SpecOpt {
	return specFn(func(s *spec) error {
		if n != 0 && (n < 10 || n > 1000) {
			return cfgErr("", "blkio_weight", "BlkioWeight", fmt.Errorf("%w: blkio weight %d", errInvalid, n))
		}
		s.blkioWeight.assign(n)
		return nil
	})
}

// Sysctl sets a namespaced sysctl.
func Sysctl(key, value string) SpecOpt {
	return specFn(func(s *spec) error {
		if key == "" {
			return cfgErr("", "sysctl", "Sysctl", fmt.Errorf("%w: sysctl key", errEmpty))
		}
		if s.sysctls == nil {
			s.sysctls = map[string]string{}
		}
		s.sysctls[key] = value
		return nil
	})
}

// SecurityOpt adds a security option (SELinux, apparmor, ...).
func SecurityOpt(opt string) SpecOpt {
	return specFn(func(s *spec) error {
		if opt == "" {
			return cfgErr("", "security_opt", "SecurityOpt", fmt.Errorf("%w: security opt", errEmpty))
		}
		s.securityOpt = append(s.securityOpt, opt)
		return nil
	})
}

// StorageOpt sets a storage-driver option.
func StorageOpt(key, value string) SpecOpt {
	return specFn(func(s *spec) error {
		if key == "" {
			return cfgErr("", "storage_opt", "StorageOpt", fmt.Errorf("%w: storage opt key", errEmpty))
		}
		if s.storageOpt == nil {
			s.storageOpt = map[string]string{}
		}
		s.storageOpt[key] = value
		return nil
	})
}

// GroupAdd adds a supplementary group.
func GroupAdd(g string) SpecOpt {
	return specFn(func(s *spec) error {
		if g == "" {
			return cfgErr("", "group_add", "GroupAdd", fmt.Errorf("%w: group", errEmpty))
		}
		s.groupAdd = append(s.groupAdd, g)
		return nil
	})
}

// Annotation adds a runtime annotation.
func Annotation(key, value string) SpecOpt {
	return specFn(func(s *spec) error {
		if key == "" {
			return cfgErr("", "annotation", "Annotation", fmt.Errorf("%w: annotation key", errEmpty))
		}
		if s.annotations == nil {
			s.annotations = map[string]string{}
		}
		s.annotations[key] = value
		return nil
	})
}

func (u *ulimit) setName(name string) error {
	if name == "" {
		return cfgErr("", "name", "Name", fmt.Errorf("%w: ulimit name", errEmpty))
	}
	u.name.assign(name)
	return nil
}

// SoftOpt is the value returned by [Soft].
// It implements [UlimitOption].
type SoftOpt struct{ n int64 }

// HardOpt is the value returned by [Hard].
// It implements [UlimitOption].
type HardOpt struct{ n int64 }

// Soft sets the soft ulimit inside [Ulimit].
func Soft(n int64) SoftOpt { return SoftOpt{n: n} }
func (o SoftOpt) applyUlimit(u *ulimit) error {
	u.soft.assign(o.n)
	return nil
}

// Hard sets the hard ulimit inside [Ulimit].
func Hard(n int64) HardOpt { return HardOpt{n: n} }
func (o HardOpt) applyUlimit(u *ulimit) error {
	u.hard.assign(o.n)
	return nil
}

// UlimitOpt is the value returned by [Ulimit].
// It implements [BuildOption], [ContainerOption], [JobOption], and [ServiceOption].
type UlimitOpt struct{ opts []UlimitOption }

// Ulimit adds a ulimit shared by Engine Resources.Ulimits and Compose ulimits.
func Ulimit(opts ...UlimitOption) UlimitOpt { return UlimitOpt{opts: opts} }

func (o UlimitOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o UlimitOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o UlimitOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o UlimitOpt) apply(s *spec) error {
	var u ulimit
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("ulimit.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyUlimit(&u); err != nil {
			errs = append(errs, wrapPath("ulimit", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !u.name.ok() {
		return cfgErr("", "ulimit.name", "Name", fmt.Errorf("%w: ulimit name", errRequired))
	}
	s.ulimits = append(s.ulimits, u)
	return nil
}
