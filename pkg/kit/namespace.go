package kit

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/cgroupns"
	"github.com/aptd3v/containerkit/pkg/ipc"
	"github.com/aptd3v/containerkit/pkg/isolation"
	"github.com/aptd3v/containerkit/pkg/pidns"
	"github.com/aptd3v/containerkit/pkg/userns"
	"github.com/aptd3v/containerkit/pkg/uts"
)

// IPC sets the IPC namespace mode (Engine IpcMode / Compose ipc).
func IPC(mode ipc.Mode) SpecOpt {
	return specFn(func(s *spec) error {
		if mode == "" {
			return cfgErr("", "ipc", "IPC", fmt.Errorf("%w: ipc", errEmpty))
		}
		s.ipc.assign(string(mode))
		return nil
	})
}

// PID sets the PID namespace mode.
func PID(mode pidns.Mode) SpecOpt {
	return specFn(func(s *spec) error {
		if mode == "" {
			return cfgErr("", "pid", "PID", fmt.Errorf("%w: pid", errEmpty))
		}
		s.pid.assign(string(mode))
		return nil
	})
}

// UTS sets the UTS namespace mode.
func UTS(mode uts.Mode) SpecOpt {
	return specFn(func(s *spec) error {
		if mode == "" {
			return cfgErr("", "uts", "UTS", fmt.Errorf("%w: uts", errEmpty))
		}
		s.uts.assign(string(mode))
		return nil
	})
}

// UserNS sets the user namespace mode.
func UserNS(mode userns.Mode) SpecOpt {
	return specFn(func(s *spec) error {
		if mode == "" {
			return cfgErr("", "userns", "UserNS", fmt.Errorf("%w: userns", errEmpty))
		}
		s.userns.assign(string(mode))
		return nil
	})
}

// Cgroup sets the cgroup to join.
func Cgroup(specName string) SpecOpt {
	return specFn(func(s *spec) error {
		if specName == "" {
			return cfgErr("", "cgroup", "Cgroup", fmt.Errorf("%w: cgroup", errEmpty))
		}
		s.cgroup.assign(specName)
		return nil
	})
}

// CgroupParent sets the parent cgroup.
func CgroupParent(parent string) SpecOpt {
	return specFn(func(s *spec) error {
		if parent == "" {
			return cfgErr("", "cgroup_parent", "CgroupParent", fmt.Errorf("%w: cgroup parent", errEmpty))
		}
		s.cgroupParent.assign(parent)
		return nil
	})
}

// CgroupNS sets the cgroup namespace mode (host/private).
func CgroupNS(mode cgroupns.Mode) SpecOpt {
	return specFn(func(s *spec) error {
		if mode == "" {
			return cfgErr("", "cgroupns", "CgroupNS", fmt.Errorf("%w: cgroupns", errEmpty))
		}
		s.cgroupns.assign(string(mode))
		return nil
	})
}

func (s *spec) setIsolation(mode string) error {
	if mode == "" {
		return cfgErr("", "isolation", "Isolation", fmt.Errorf("%w: isolation", errEmpty))
	}
	s.isolation.assign(mode)
	return nil
}

// Isolation sets Windows isolation (default, process, hyperv). Inside [Build]
// it maps to compose build.isolation.
func Isolation(mode isolation.Mode) IsolationOpt { return IsolationOpt{m: string(mode)} }

// Runtime sets the OCI runtime.
func Runtime(name string) SpecOpt {
	return specFn(func(s *spec) error {
		if name == "" {
			return cfgErr("", "runtime", "Runtime", fmt.Errorf("%w: runtime", errEmpty))
		}
		s.runtime.assign(name)
		return nil
	})
}

// Umask sets the process umask.
func Umask(mode uint32) SpecOpt {
	return specFn(func(s *spec) error {
		s.umask.assign(mode)
		return nil
	})
}
