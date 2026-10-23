package containerkit

import (
	"strings"

	"github.com/aptd3v/containerkit/pkg/containerkit/errdefs"
	"github.com/docker/docker/api/types/container"
)

// NetworkMode is a container network namespace mode.
// The zero value is Docker's default.
type NetworkMode struct{ mode string }

var (
	NetworkBridge = NetworkMode{mode: "bridge"}
	NetworkHost   = NetworkMode{mode: "host"}
	NetworkNone   = NetworkMode{mode: "none"}
)

// NetworkContainer joins the named container's network namespace.
func NetworkContainer(name string) (NetworkMode, error) {
	if strings.TrimSpace(name) == "" {
		return NetworkMode{}, errdefs.NewHostConfigError("network_mode", "container name is empty")
	}
	return NetworkMode{mode: "container:" + name}, nil
}

// NetworkMode sets the container network namespace.
func (c *Container) NetworkMode(mode NetworkMode) *Container {
	c.Config.Host.NetworkMode = container.NetworkMode(mode.mode)
	return c
}

// IpcMode is a container IPC namespace mode.
// The zero value is Docker's default.
type IpcMode struct{ mode string }

var (
	IpcHost      = IpcMode{mode: "host"}
	IpcPrivate   = IpcMode{mode: "private"}
	IpcShareable = IpcMode{mode: "shareable"}
)

// IpcContainer joins the named container's IPC namespace.
func IpcContainer(name string) (IpcMode, error) {
	if strings.TrimSpace(name) == "" {
		return IpcMode{}, errdefs.NewHostConfigError("ipc_mode", "container name is empty")
	}
	return IpcMode{mode: "container:" + name}, nil
}

// IpcMode sets the container IPC namespace.
func (c *Container) IpcMode(mode IpcMode) *Container {
	c.Config.Host.IpcMode = container.IpcMode(mode.mode)
	return c
}

// PidMode is a container PID namespace mode.
// The zero value is Docker's default.
type PidMode struct{ mode string }

// PidHost uses the host PID namespace.
var PidHost = PidMode{mode: "host"}

// PidContainer joins the named container's PID namespace.
func PidContainer(name string) (PidMode, error) {
	if strings.TrimSpace(name) == "" {
		return PidMode{}, errdefs.NewHostConfigError("pid_mode", "container name is empty")
	}
	return PidMode{mode: "container:" + name}, nil
}

// PidMode sets the container PID namespace.
func (c *Container) PidMode(mode PidMode) *Container {
	c.Config.Host.PidMode = container.PidMode(mode.mode)
	return c
}

// UTSMode is a container UTS namespace mode.
// The zero value is Docker's default.
type UTSMode struct{ mode string }

// UTSHost uses the host UTS namespace.
var UTSHost = UTSMode{mode: "host"}

// UTSMode sets the container UTS namespace.
func (c *Container) UTSMode(mode UTSMode) *Container {
	c.Config.Host.UTSMode = container.UTSMode(mode.mode)
	return c
}

// UserNSMode is a container user namespace mode.
// The zero value is Docker's default.
type UserNSMode struct{ mode string }

var (
	UserNSHost    = UserNSMode{mode: "host"}
	UserNSPrivate = UserNSMode{mode: "private"}
)

// UserNSContainer joins the named container's user namespace.
func UserNSContainer(name string) (UserNSMode, error) {
	if strings.TrimSpace(name) == "" {
		return UserNSMode{}, errdefs.NewHostConfigError("userns_mode", "container name is empty")
	}
	return UserNSMode{mode: "container:" + name}, nil
}

// UserNSMode sets the container user namespace.
func (c *Container) UserNSMode(mode UserNSMode) *Container {
	c.Config.Host.UsernsMode = container.UsernsMode(mode.mode)
	return c
}

// CgroupMode is a container cgroup namespace mode.
// The zero value is Docker's default.
type CgroupMode struct{ mode string }

var (
	CgroupHost    = CgroupMode{mode: "host"}
	CgroupPrivate = CgroupMode{mode: "private"}
	CgroupNone    = CgroupMode{mode: "none"}
)

// Cgroup sets the container cgroup namespace.
func (c *Container) Cgroup(mode CgroupMode) *Container {
	c.Config.Host.Cgroup = container.CgroupSpec(mode.mode)
	return c
}

// Isolation is a Windows isolation mode.
// The zero value is Docker's default.
type Isolation struct{ mode string }

var (
	IsolationDefault = Isolation{mode: "default"}
	IsolationProcess = Isolation{mode: "process"}
	IsolationHyperV  = Isolation{mode: "hyperv"}
)

// Isolation sets the container isolation mode.
func (c *Container) Isolation(mode Isolation) *Container {
	c.Config.Host.Isolation = container.Isolation(mode.mode)
	return c
}
