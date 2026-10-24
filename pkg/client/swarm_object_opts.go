package client

import sdk "github.com/moby/moby/client"

// NodeList is [sdk.NodeListOptions].
type NodeList sdk.NodeListOptions

func (o *NodeList) apply() sdk.NodeListOptions { return sdk.NodeListOptions(apply(o)) }

// NodeUpdate is [sdk.NodeUpdateOptions].
type NodeUpdate sdk.NodeUpdateOptions

func (o *NodeUpdate) apply() sdk.NodeUpdateOptions { return sdk.NodeUpdateOptions(apply(o)) }

// NodeRemove is [sdk.NodeRemoveOptions].
type NodeRemove sdk.NodeRemoveOptions

func (o *NodeRemove) apply() sdk.NodeRemoveOptions { return sdk.NodeRemoveOptions(apply(o)) }

// ServiceCreate is [sdk.ServiceCreateOptions].
type ServiceCreate sdk.ServiceCreateOptions

func (o *ServiceCreate) apply() sdk.ServiceCreateOptions { return sdk.ServiceCreateOptions(apply(o)) }

// ServiceInspect is [sdk.ServiceInspectOptions].
type ServiceInspect sdk.ServiceInspectOptions

func (o *ServiceInspect) apply() sdk.ServiceInspectOptions {
	return sdk.ServiceInspectOptions(apply(o))
}

// ServiceList is [sdk.ServiceListOptions].
type ServiceList sdk.ServiceListOptions

func (o *ServiceList) apply() sdk.ServiceListOptions { return sdk.ServiceListOptions(apply(o)) }

// ServiceUpdate is [sdk.ServiceUpdateOptions].
type ServiceUpdate sdk.ServiceUpdateOptions

func (o *ServiceUpdate) apply() sdk.ServiceUpdateOptions { return sdk.ServiceUpdateOptions(apply(o)) }

// ServiceLogs is [sdk.ServiceLogsOptions].
type ServiceLogs sdk.ServiceLogsOptions

func (o *ServiceLogs) apply() sdk.ServiceLogsOptions { return sdk.ServiceLogsOptions(apply(o)) }

// TaskList is [sdk.TaskListOptions].
type TaskList sdk.TaskListOptions

func (o *TaskList) apply() sdk.TaskListOptions { return sdk.TaskListOptions(apply(o)) }

// TaskLogs is [sdk.TaskLogsOptions].
type TaskLogs sdk.TaskLogsOptions

func (o *TaskLogs) apply() sdk.TaskLogsOptions { return sdk.TaskLogsOptions(apply(o)) }

// SecretCreate is [sdk.SecretCreateOptions].
type SecretCreate sdk.SecretCreateOptions

func (o *SecretCreate) apply() sdk.SecretCreateOptions { return sdk.SecretCreateOptions(apply(o)) }

// SecretList is [sdk.SecretListOptions].
type SecretList sdk.SecretListOptions

func (o *SecretList) apply() sdk.SecretListOptions { return sdk.SecretListOptions(apply(o)) }

// SecretUpdate is [sdk.SecretUpdateOptions].
type SecretUpdate sdk.SecretUpdateOptions

func (o *SecretUpdate) apply() sdk.SecretUpdateOptions { return sdk.SecretUpdateOptions(apply(o)) }

// ConfigCreate is [sdk.ConfigCreateOptions].
type ConfigCreate sdk.ConfigCreateOptions

func (o *ConfigCreate) apply() sdk.ConfigCreateOptions { return sdk.ConfigCreateOptions(apply(o)) }

// ConfigList is [sdk.ConfigListOptions].
type ConfigList sdk.ConfigListOptions

func (o *ConfigList) apply() sdk.ConfigListOptions { return sdk.ConfigListOptions(apply(o)) }

// ConfigUpdate is [sdk.ConfigUpdateOptions].
type ConfigUpdate sdk.ConfigUpdateOptions

func (o *ConfigUpdate) apply() sdk.ConfigUpdateOptions { return sdk.ConfigUpdateOptions(apply(o)) }
