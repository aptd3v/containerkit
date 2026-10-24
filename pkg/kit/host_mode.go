package kit

import (
	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

func (c *Container) hostModeError(field string, err error) bool {
	if err == nil {
		return false
	}
	c.Errors = append(c.Errors, errdefs.NewHostConfigError(field, err.Error()))
	return true
}

// NetworkMode sets the container network namespace.
func (c *Container) NetworkMode(mode types.NetworkMode) *Container {
	if c.hostModeError("network_mode", types.ValidateNetworkMode(mode)) {
		return c
	}
	c.Config.Host.NetworkMode = mode
	return c
}

// IpcMode sets the container IPC namespace.
func (c *Container) IpcMode(mode types.IpcMode) *Container {
	if c.hostModeError("ipc_mode", types.ValidateIpcMode(mode)) {
		return c
	}
	c.Config.Host.IpcMode = mode
	return c
}

// PidMode sets the container PID namespace.
func (c *Container) PidMode(mode types.PidMode) *Container {
	if c.hostModeError("pid_mode", types.ValidatePidMode(mode)) {
		return c
	}
	c.Config.Host.PidMode = mode
	return c
}

// UTSMode sets the container UTS namespace.
func (c *Container) UTSMode(mode types.UTSMode) *Container {
	c.Config.Host.UTSMode = mode
	return c
}

// UserNSMode sets the container user namespace.
func (c *Container) UserNSMode(mode types.UserNSMode) *Container {
	c.Config.Host.UsernsMode = mode
	return c
}

// Cgroup sets the container cgroup namespace.
func (c *Container) Cgroup(mode types.CgroupMode) *Container {
	c.Config.Host.CgroupnsMode = mode
	return c
}

// Isolation sets the container isolation mode.
func (c *Container) Isolation(mode types.Isolation) *Container {
	c.Config.Host.Isolation = mode
	return c
}
