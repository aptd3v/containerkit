package types

import "github.com/docker/docker/api/types/mount"

// MountPropagation is a bind-mount propagation mode.
type MountPropagation = mount.Propagation

const (
	PropagationRPrivate = mount.PropagationRPrivate
	PropagationPrivate  = mount.PropagationPrivate
	PropagationRShared  = mount.PropagationRShared
	PropagationShared   = mount.PropagationShared
	PropagationRSlave   = mount.PropagationRSlave
	PropagationSlave    = mount.PropagationSlave
)

// MountType is a volume/bind/tmpfs mount type.
type MountType = mount.Type

const (
	MountBind      = mount.TypeBind
	MountVolume    = mount.TypeVolume
	MountTmpfs     = mount.TypeTmpfs
	MountNamedPipe = mount.TypeNamedPipe
	MountImage     = mount.TypeImage
	MountCluster   = mount.TypeCluster
)

// MountConsistency is a bind-mount consistency mode.
type MountConsistency = mount.Consistency

const (
	ConsistencyFull      = mount.ConsistencyFull
	ConsistencyCached    = mount.ConsistencyCached
	ConsistencyDelegated = mount.ConsistencyDelegated
	ConsistencyDefault   = mount.ConsistencyDefault
)
