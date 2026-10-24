package kit

import (
	"github.com/aptd3v/containerkit/pkg/kit/types"
	"github.com/docker/docker/api/types/strslice"
)

// AddedCapabilities adds capabilities to the host configuration for the container.
// parameters:
//   - caps: the capabilities to add
func (c *Container) AddedCapabilities(caps ...types.Capability) *Container {
	if c.Config.Host.CapAdd == nil {
		c.Config.Host.CapAdd = make(strslice.StrSlice, 0, len(caps))
	}
	for _, cap := range caps {
		c.Config.Host.CapAdd = append(c.Config.Host.CapAdd, string(cap))
	}
	return c
}

// DroppedCapabilities drops capabilities from the host configuration for the container.
// parameters:
//   - caps: the capabilities to drop
func (c *Container) DroppedCapabilities(caps ...types.Capability) *Container {
	if c.Config.Host.CapDrop == nil {
		c.Config.Host.CapDrop = make(strslice.StrSlice, 0, len(caps))
	}
	for _, cap := range caps {
		c.Config.Host.CapDrop = append(c.Config.Host.CapDrop, string(cap))
	}
	return c
}

// DroppedAllCapabilities drops all capabilities
func (c *Container) DroppedAllCapabilities() *Container { return c.DroppedCapabilities(types.ALL) }

// DroppedSensitiveCapabilities drops a common set of sensitive capabilities
//
// SYS_ADMIN, SYS_MODULE, SYS_PTRACE, SYS_TIME, SYSLOG, MAC_ADMIN, MAC_OVERRIDE, CHECKPOINT_RESTORE
func (c *Container) DroppedSensitiveCapabilities() *Container {
	return c.DroppedCapabilities(
		types.SYS_ADMIN, types.SYS_MODULE, types.SYS_PTRACE, types.SYS_TIME, types.SYSLOG,
		types.MAC_ADMIN, types.MAC_OVERRIDE, types.CHECKPOINT_RESTORE,
	)
}
