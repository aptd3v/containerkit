package client

import (
	"context"

	"github.com/aptd3v/go-contain/pkg/client/options/ncno"
	"github.com/aptd3v/go-contain/pkg/client/options/nco"
	"github.com/aptd3v/go-contain/pkg/client/options/nio"
	"github.com/aptd3v/go-contain/pkg/client/options/nlo"
	"github.com/aptd3v/go-contain/pkg/client/options/npo"
	"github.com/aptd3v/go-contain/pkg/client/response"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

// NetworkCreate creates a new network in the docker host.
func (c *Client) NetworkCreate(ctx context.Context, name string, setters ...nco.SetNetworkCreateOption) (*response.NetworkCreate, error) {
	o := network.CreateOptions{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return nil, err
		}
	}
	resp, err := c.wrapped.NetworkCreate(ctx, name, o)
	if err != nil {
		return nil, err
	}
	return &response.NetworkCreate{CreateResponse: resp}, nil
}

// NetworkConnect connects a container to an existent network in the docker host.
func (c *Client) NetworkConnect(ctx context.Context, networkID string, setters ...ncno.SetNetworkConnectOption) error {
	o := network.ConnectOptions{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return err
		}

	}
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
func (c *Client) NetworkInspect(ctx context.Context, networkID string, setters ...nio.SetNetworkInspectOption) (*response.NetworkInspect, error) {
	o := network.InspectOptions{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return nil, err
		}
	}
	resp, err := c.wrapped.NetworkInspect(ctx, networkID, o)
	if err != nil {
		return nil, err
	}
	return &response.NetworkInspect{Inspect: resp}, nil
}

// NetworkList returns the list of networks configured in the docker host.
func (c *Client) NetworkList(ctx context.Context, setters ...nlo.SetNetworkListOption) ([]*response.NetworkSummary, error) {
	o := network.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return nil, err
		}
	}
	resp, err := c.wrapped.NetworkList(ctx, o)
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
func (c *Client) NetworksPrune(ctx context.Context, setters ...npo.SetNetworkPruneOption) (*response.NetworkPruneReport, error) {
	o := network.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o.Filters); err != nil {
			return nil, err
		}
	}
	resp, err := c.wrapped.NetworksPrune(ctx, o.Filters)
	if err != nil {
		return nil, err
	}
	return &response.NetworkPruneReport{PruneReport: resp}, nil

}
