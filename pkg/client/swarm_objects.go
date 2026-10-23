package client

import (
	"context"
	"io"

	"github.com/aptd3v/containerkit/pkg/client/response"
	sdk "github.com/moby/moby/client"
)

func cloneRaw(raw []byte) []byte {
	if raw == nil {
		return nil
	}
	return append([]byte(nil), raw...)
}

// NodeList returns the nodes in the swarm.
func (c *Client) NodeList(ctx context.Context, opt *NodeList) ([]response.Node, error) {
	listed, err := c.wrapped.NodeList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	nodes := make([]response.Node, len(listed.Items))
	for i, item := range listed.Items {
		nodes[i] = response.Node{Node: item}
	}
	return nodes, nil
}

// NodeInspect returns one swarm node. Raw is the daemon's JSON body.
func (c *Client) NodeInspect(ctx context.Context, id string) (*response.Node, error) {
	inspected, err := c.wrapped.NodeInspect(ctx, id, sdk.NodeInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Node{Node: inspected.Node, Raw: cloneRaw(inspected.Raw)}, nil
}

// NodeUpdate updates a swarm node. Version must match the node's current object version.
func (c *Client) NodeUpdate(ctx context.Context, id string, opt *NodeUpdate) error {
	_, err := c.wrapped.NodeUpdate(ctx, id, opt.apply())
	return err
}

// NodeRemove removes a swarm node.
func (c *Client) NodeRemove(ctx context.Context, id string, opt *NodeRemove) error {
	_, err := c.wrapped.NodeRemove(ctx, id, opt.apply())
	return err
}

// ServiceCreate creates a swarm service.
func (c *Client) ServiceCreate(ctx context.Context, opt *ServiceCreate) (*response.ServiceCreate, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	created, err := c.wrapped.ServiceCreate(ctx, op)
	if err != nil {
		return nil, err
	}
	return &response.ServiceCreate{ID: created.ID, Warnings: created.Warnings}, nil
}

// ServiceInspect returns one swarm service. Raw is the daemon's JSON body.
func (c *Client) ServiceInspect(ctx context.Context, id string, opt *ServiceInspect) (*response.Service, error) {
	inspected, err := c.wrapped.ServiceInspect(ctx, id, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.Service{Service: inspected.Service, Raw: cloneRaw(inspected.Raw)}, nil
}

// ServiceList returns swarm services.
func (c *Client) ServiceList(ctx context.Context, opt *ServiceList) ([]response.Service, error) {
	listed, err := c.wrapped.ServiceList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	services := make([]response.Service, len(listed.Items))
	for i, item := range listed.Items {
		services[i] = response.Service{Service: item}
	}
	return services, nil
}

// ServiceUpdate updates a swarm service.
func (c *Client) ServiceUpdate(ctx context.Context, id string, opt *ServiceUpdate) (*response.ServiceUpdate, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	updated, err := c.wrapped.ServiceUpdate(ctx, id, op)
	if err != nil {
		return nil, err
	}
	return &response.ServiceUpdate{Warnings: updated.Warnings}, nil
}

// ServiceRemove removes a swarm service.
func (c *Client) ServiceRemove(ctx context.Context, id string) error {
	_, err := c.wrapped.ServiceRemove(ctx, id, sdk.ServiceRemoveOptions{})
	return err
}

// ServiceLogs returns the logs of a swarm service. The caller closes the stream.
func (c *Client) ServiceLogs(ctx context.Context, id string, opt *ServiceLogs) (io.ReadCloser, error) {
	return c.wrapped.ServiceLogs(ctx, id, opt.apply())
}

// TaskList returns swarm tasks.
func (c *Client) TaskList(ctx context.Context, opt *TaskList) ([]response.Task, error) {
	listed, err := c.wrapped.TaskList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	tasks := make([]response.Task, len(listed.Items))
	for i, item := range listed.Items {
		tasks[i] = response.Task{Task: item}
	}
	return tasks, nil
}

// TaskInspect returns one swarm task. Raw is the daemon's JSON body.
func (c *Client) TaskInspect(ctx context.Context, id string) (*response.Task, error) {
	inspected, err := c.wrapped.TaskInspect(ctx, id, sdk.TaskInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Task{Task: inspected.Task, Raw: cloneRaw(inspected.Raw)}, nil
}

// TaskLogs returns the logs of one swarm task. The caller closes the stream.
func (c *Client) TaskLogs(ctx context.Context, id string, opt *TaskLogs) (io.ReadCloser, error) {
	return c.wrapped.TaskLogs(ctx, id, opt.apply())
}

// SecretCreate creates a swarm secret and returns its ID.
func (c *Client) SecretCreate(ctx context.Context, opt *SecretCreate) (string, error) {
	created, err := c.wrapped.SecretCreate(ctx, opt.apply())
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

// SecretInspect returns one swarm secret. Raw is the daemon's JSON body.
func (c *Client) SecretInspect(ctx context.Context, id string) (*response.Secret, error) {
	inspected, err := c.wrapped.SecretInspect(ctx, id, sdk.SecretInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Secret{Secret: inspected.Secret, Raw: cloneRaw(inspected.Raw)}, nil
}

// SecretList returns swarm secrets.
func (c *Client) SecretList(ctx context.Context, opt *SecretList) ([]response.Secret, error) {
	listed, err := c.wrapped.SecretList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	secrets := make([]response.Secret, len(listed.Items))
	for i, item := range listed.Items {
		secrets[i] = response.Secret{Secret: item}
	}
	return secrets, nil
}

// SecretUpdate updates a swarm secret.
func (c *Client) SecretUpdate(ctx context.Context, id string, opt *SecretUpdate) error {
	_, err := c.wrapped.SecretUpdate(ctx, id, opt.apply())
	return err
}

// SecretRemove removes a swarm secret.
func (c *Client) SecretRemove(ctx context.Context, id string) error {
	_, err := c.wrapped.SecretRemove(ctx, id, sdk.SecretRemoveOptions{})
	return err
}

// ConfigCreate creates a swarm config and returns its ID.
func (c *Client) ConfigCreate(ctx context.Context, opt *ConfigCreate) (string, error) {
	created, err := c.wrapped.ConfigCreate(ctx, opt.apply())
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

// ConfigInspect returns one swarm config. Raw is the daemon's JSON body.
func (c *Client) ConfigInspect(ctx context.Context, id string) (*response.SwarmConfig, error) {
	inspected, err := c.wrapped.ConfigInspect(ctx, id, sdk.ConfigInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.SwarmConfig{Config: inspected.Config, Raw: cloneRaw(inspected.Raw)}, nil
}

// ConfigList returns swarm configs.
func (c *Client) ConfigList(ctx context.Context, opt *ConfigList) ([]response.SwarmConfig, error) {
	listed, err := c.wrapped.ConfigList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	configs := make([]response.SwarmConfig, len(listed.Items))
	for i, item := range listed.Items {
		configs[i] = response.SwarmConfig{Config: item}
	}
	return configs, nil
}

// ConfigUpdate updates a swarm config.
func (c *Client) ConfigUpdate(ctx context.Context, id string, opt *ConfigUpdate) error {
	_, err := c.wrapped.ConfigUpdate(ctx, id, opt.apply())
	return err
}

// ConfigRemove removes a swarm config.
func (c *Client) ConfigRemove(ctx context.Context, id string) error {
	_, err := c.wrapped.ConfigRemove(ctx, id, sdk.ConfigRemoveOptions{})
	return err
}
