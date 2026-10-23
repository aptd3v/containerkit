package client

import (
	"github.com/aptd3v/containerkit/pkg/client/auth"
	"github.com/moby/moby/api/types/swarm"
	sdk "github.com/moby/moby/client"
)

// NodeList is options for NodeList.
type NodeList struct {
	Filters []Filter
}

func (o *NodeList) apply() sdk.NodeListOptions {
	if o == nil {
		return sdk.NodeListOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.NodeListOptions{Filters: argsFromFilters(o.Filters)}
}

// NodeUpdate is options for NodeUpdate.
type NodeUpdate struct {
	Version swarm.Version
	Spec    swarm.NodeSpec
}

func (o *NodeUpdate) apply() sdk.NodeUpdateOptions {
	if o == nil {
		return sdk.NodeUpdateOptions{}
	}
	return sdk.NodeUpdateOptions{Version: o.Version, Spec: o.Spec}
}

// NodeRemove is options for NodeRemove.
type NodeRemove struct {
	Force bool
}

func (o *NodeRemove) apply() sdk.NodeRemoveOptions {
	if o == nil {
		return sdk.NodeRemoveOptions{}
	}
	return sdk.NodeRemoveOptions{Force: o.Force}
}

// ServiceCreate is options for ServiceCreate.
type ServiceCreate struct {
	Spec          swarm.ServiceSpec
	Auth          auth.Auth
	QueryRegistry bool
}

func (o *ServiceCreate) apply() (sdk.ServiceCreateOptions, error) {
	if o == nil {
		return sdk.ServiceCreateOptions{}, nil
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.ServiceCreateOptions{}, err
	}
	return sdk.ServiceCreateOptions{
		Spec:                o.Spec,
		EncodedRegistryAuth: encoded,
		QueryRegistry:       o.QueryRegistry,
	}, nil
}

// ServiceInspect is options for ServiceInspect.
type ServiceInspect struct {
	InsertDefaults bool
}

func (o *ServiceInspect) apply() sdk.ServiceInspectOptions {
	if o == nil {
		return sdk.ServiceInspectOptions{}
	}
	return sdk.ServiceInspectOptions{InsertDefaults: o.InsertDefaults}
}

// ServiceList is options for ServiceList.
type ServiceList struct {
	Filters []Filter
	Status  bool
}

func (o *ServiceList) apply() sdk.ServiceListOptions {
	if o == nil {
		return sdk.ServiceListOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.ServiceListOptions{Filters: argsFromFilters(o.Filters), Status: o.Status}
}

// ServiceUpdate is options for ServiceUpdate.
type ServiceUpdate struct {
	Version          swarm.Version
	Spec             swarm.ServiceSpec
	Auth             auth.Auth
	RegistryAuthFrom swarm.RegistryAuthSource
	Rollback         string
	QueryRegistry    bool
}

func (o *ServiceUpdate) apply() (sdk.ServiceUpdateOptions, error) {
	if o == nil {
		return sdk.ServiceUpdateOptions{}, nil
	}
	encoded, err := encodeAuth(o.Auth)
	if err != nil {
		return sdk.ServiceUpdateOptions{}, err
	}
	return sdk.ServiceUpdateOptions{
		Version:             o.Version,
		Spec:                o.Spec,
		EncodedRegistryAuth: encoded,
		RegistryAuthFrom:    o.RegistryAuthFrom,
		Rollback:            o.Rollback,
		QueryRegistry:       o.QueryRegistry,
	}, nil
}

// ServiceLogs is options for ServiceLogs.
type ServiceLogs struct {
	ShowStdout bool
	ShowStderr bool
	Since      string
	Until      string
	Timestamps bool
	Follow     bool
	Tail       string
	Details    bool
}

func (o *ServiceLogs) apply() sdk.ServiceLogsOptions {
	if o == nil {
		return sdk.ServiceLogsOptions{}
	}
	return sdk.ServiceLogsOptions{
		ShowStdout: o.ShowStdout,
		ShowStderr: o.ShowStderr,
		Since:      o.Since,
		Until:      o.Until,
		Timestamps: o.Timestamps,
		Follow:     o.Follow,
		Tail:       o.Tail,
		Details:    o.Details,
	}
}

// TaskList is options for TaskList.
type TaskList struct {
	Filters []Filter
}

func (o *TaskList) apply() sdk.TaskListOptions {
	if o == nil {
		return sdk.TaskListOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.TaskListOptions{Filters: argsFromFilters(o.Filters)}
}

// TaskLogs is options for TaskLogs.
type TaskLogs struct {
	ShowStdout bool
	ShowStderr bool
	Since      string
	Until      string
	Timestamps bool
	Follow     bool
	Tail       string
	Details    bool
}

func (o *TaskLogs) apply() sdk.TaskLogsOptions {
	if o == nil {
		return sdk.TaskLogsOptions{}
	}
	return sdk.TaskLogsOptions{
		ShowStdout: o.ShowStdout,
		ShowStderr: o.ShowStderr,
		Since:      o.Since,
		Until:      o.Until,
		Timestamps: o.Timestamps,
		Follow:     o.Follow,
		Tail:       o.Tail,
		Details:    o.Details,
	}
}

// SecretCreate is options for SecretCreate.
type SecretCreate struct {
	Spec swarm.SecretSpec
}

func (o *SecretCreate) apply() sdk.SecretCreateOptions {
	if o == nil {
		return sdk.SecretCreateOptions{}
	}
	return sdk.SecretCreateOptions{Spec: o.Spec}
}

// SecretList is options for SecretList.
type SecretList struct {
	Filters []Filter
}

func (o *SecretList) apply() sdk.SecretListOptions {
	if o == nil {
		return sdk.SecretListOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.SecretListOptions{Filters: argsFromFilters(o.Filters)}
}

// SecretUpdate is options for SecretUpdate.
type SecretUpdate struct {
	Version swarm.Version
	Spec    swarm.SecretSpec
}

func (o *SecretUpdate) apply() sdk.SecretUpdateOptions {
	if o == nil {
		return sdk.SecretUpdateOptions{}
	}
	return sdk.SecretUpdateOptions{Version: o.Version, Spec: o.Spec}
}

// ConfigCreate is options for ConfigCreate.
type ConfigCreate struct {
	Spec swarm.ConfigSpec
}

func (o *ConfigCreate) apply() sdk.ConfigCreateOptions {
	if o == nil {
		return sdk.ConfigCreateOptions{}
	}
	return sdk.ConfigCreateOptions{Spec: o.Spec}
}

// ConfigList is options for ConfigList.
type ConfigList struct {
	Filters []Filter
}

func (o *ConfigList) apply() sdk.ConfigListOptions {
	if o == nil {
		return sdk.ConfigListOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.ConfigListOptions{Filters: argsFromFilters(o.Filters)}
}

// ConfigUpdate is options for ConfigUpdate.
type ConfigUpdate struct {
	Version swarm.Version
	Spec    swarm.ConfigSpec
}

func (o *ConfigUpdate) apply() sdk.ConfigUpdateOptions {
	if o == nil {
		return sdk.ConfigUpdateOptions{}
	}
	return sdk.ConfigUpdateOptions{Version: o.Version, Spec: o.Spec}
}
