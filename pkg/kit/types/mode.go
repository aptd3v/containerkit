package types

import (
	"errors"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

var errEmptyContainerName = errors.New("container name is empty")

func containerMode(name string) string {
	return "container:" + strings.TrimSpace(name)
}

func errIfEmptyJoin(isContainer bool, id string) error {
	if isContainer && id == "" {
		return errEmptyContainerName
	}
	return nil
}

// NetworkMode is [container.NetworkMode].
type NetworkMode = container.NetworkMode

var (
	NetworkBridge NetworkMode = network.NetworkBridge
	NetworkHost   NetworkMode = network.NetworkHost
	NetworkNone   NetworkMode = network.NetworkNone
)

// ValidateNetworkMode reports an invalid container-join mode. The zero value is valid.
func ValidateNetworkMode(m NetworkMode) error {
	return errIfEmptyJoin(m.IsContainer(), m.ConnectedContainer())
}

// NetworkContainer joins the named container's network namespace.
func NetworkContainer(name string) NetworkMode {
	return NetworkMode(containerMode(name))
}

// IpcMode is [container.IpcMode].
type IpcMode = container.IpcMode

var (
	IpcHost      = container.IPCModeHost
	IpcPrivate   = container.IPCModePrivate
	IpcShareable = container.IPCModeShareable
	IpcNone      = container.IPCModeNone
)

// ValidateIpcMode reports an invalid container-join mode. The zero value is valid.
func ValidateIpcMode(m IpcMode) error {
	return errIfEmptyJoin(m.IsContainer(), m.Container())
}

// IpcContainer joins the named container's IPC namespace.
func IpcContainer(name string) IpcMode {
	return IpcMode(containerMode(name))
}

// PidMode is [container.PidMode].
type PidMode = container.PidMode

// PidHost uses the host PID namespace.
var PidHost PidMode = "host"

// ValidatePidMode reports an invalid container-join mode. The zero value is valid.
func ValidatePidMode(m PidMode) error {
	return errIfEmptyJoin(m.IsContainer(), m.Container())
}

// PidContainer joins the named container's PID namespace.
func PidContainer(name string) PidMode {
	return PidMode(containerMode(name))
}

// UTSMode is [container.UTSMode].
type UTSMode = container.UTSMode

// UTSHost uses the host UTS namespace.
var UTSHost UTSMode = "host"

// UserNSMode is [container.UsernsMode].
type UserNSMode = container.UsernsMode

var (
	UserNSHost    UserNSMode = "host"
	UserNSPrivate UserNSMode
)

// CgroupMode is [container.CgroupnsMode].
type CgroupMode = container.CgroupnsMode

var (
	CgroupHost    = container.CgroupnsModeHost
	CgroupPrivate = container.CgroupnsModePrivate
)

// Isolation is [container.Isolation].
type Isolation = container.Isolation

var (
	IsolationDefault = container.IsolationDefault
	IsolationProcess = container.IsolationProcess
	IsolationHyperV  = container.IsolationHyperV
)
