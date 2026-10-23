package client

import (
	"context"
	"strings"

	"github.com/aptd3v/containerkit/pkg/client/response"
	"github.com/moby/moby/api/types/network"
	sdk "github.com/moby/moby/client"
)

// NetworkCreate creates a new network in the docker host.
func (c *Client) NetworkCreate(ctx context.Context, name string, opt *NetworkCreate) (*response.NetworkCreate, error) {
	op, err := opt.apply()
	if err != nil {
		return nil, err
	}
	resp, err := c.wrapped.NetworkCreate(ctx, name, op)
	if err != nil {
		return nil, err
	}
	return &response.NetworkCreate{CreateResponse: network.CreateResponse{
		ID:      resp.ID,
		Warning: strings.Join(resp.Warning, "\n"),
	}}, nil
}

// NetworkConnect connects a container to an existent network in the docker host.
func (c *Client) NetworkConnect(ctx context.Context, networkID string, opt *NetworkConnect) error {
	_, err := c.wrapped.NetworkConnect(ctx, networkID, opt.apply())
	return err
}

// NetworkDisconnect disconnects a container from an existent network in the docker host.
func (c *Client) NetworkDisconnect(ctx context.Context, networkID string, containerID string, force bool) error {
	_, err := c.wrapped.NetworkDisconnect(ctx, networkID, sdk.NetworkDisconnectOptions{
		Container: containerID,
		Force:     force,
	})
	return err
}

// NetworkRemove removes an existent network from the docker host.
func (c *Client) NetworkRemove(ctx context.Context, networkID string) error {
	_, err := c.wrapped.NetworkRemove(ctx, networkID, sdk.NetworkRemoveOptions{})
	return err
}

// NetworkInspect returns the information for a specific network configured in the docker host.
func (c *Client) NetworkInspect(ctx context.Context, networkID string, opt *NetworkInspect) (*response.NetworkInspect, error) {
	resp, err := c.wrapped.NetworkInspect(ctx, networkID, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.NetworkInspect{Inspect: resp.Network}, nil
}

// NetworkList returns the list of networks configured in the docker host.
func (c *Client) NetworkList(ctx context.Context, opt *NetworkList) ([]*response.NetworkSummary, error) {
	resp, err := c.wrapped.NetworkList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	summaries := make([]*response.NetworkSummary, 0, len(resp.Items))
	for _, summary := range resp.Items {
		summaries = append(summaries, &response.NetworkSummary{Summary: summary})
	}
	return summaries, nil
}

// NetworksPrune requests the daemon to delete unused networks.
func (c *Client) NetworksPrune(ctx context.Context, opt *NetworkPrune) (*response.NetworkPruneReport, error) {
	resp, err := c.wrapped.NetworkPrune(ctx, sdk.NetworkPruneOptions{Filters: opt.apply()})
	if err != nil {
		return nil, err
	}
	return &response.NetworkPruneReport{PruneReport: resp.Report}, nil
}
