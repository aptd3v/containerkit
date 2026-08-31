package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/response"
)

// NetworkCreate creates a new network in the docker host.
func (c *Client) NetworkCreate(ctx context.Context, name string, opt *NetworkCreate) (*response.NetworkCreate, error) {
	resp, err := c.wrapped.NetworkCreate(ctx, name, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.NetworkCreate{CreateResponse: resp}, nil
}

// NetworkConnect connects a container to an existent network in the docker host.
func (c *Client) NetworkConnect(ctx context.Context, networkID string, opt *NetworkConnect) error {
	o := opt.apply()
	return c.wrapped.NetworkConnect(ctx, networkID, o.Container, o.EndpointConfig)
}

// NetworkDisconnect disconnects a container from an existent network in the docker host.
func (c *Client) NetworkDisconnect(ctx context.Context, networkID string, containerID string, force bool) error {
	return c.wrapped.NetworkDisconnect(ctx, networkID, containerID, force)
}

// NetworkRemove removes an existent network from the docker host.
func (c *Client) NetworkRemove(ctx context.Context, networkID string) error {
	return c.wrapped.NetworkRemove(ctx, networkID)
}

// NetworkInspect returns the information for a specific network configured in the docker host.
func (c *Client) NetworkInspect(ctx context.Context, networkID string, opt *NetworkInspect) (*response.NetworkInspect, error) {
	resp, err := c.wrapped.NetworkInspect(ctx, networkID, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.NetworkInspect{Inspect: resp}, nil
}

// NetworkList returns the list of networks configured in the docker host.
func (c *Client) NetworkList(ctx context.Context, opt *NetworkList) ([]*response.NetworkSummary, error) {
	resp, err := c.wrapped.NetworkList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}

	summaries := make([]*response.NetworkSummary, 0, len(resp))
	for _, summary := range resp {
		summaries = append(summaries, &response.NetworkSummary{Summary: summary})
	}
	return summaries, nil
}

// NetworksPrune requests the daemon to delete unused networks
func (c *Client) NetworksPrune(ctx context.Context, opt *NetworkPrune) (*response.NetworkPruneReport, error) {
	resp, err := c.wrapped.NetworksPrune(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.NetworkPruneReport{PruneReport: resp}, nil

}
