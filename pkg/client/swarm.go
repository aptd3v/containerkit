package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/response"
)

// SwarmInit initializes the swarm.
func (c *Client) SwarmInit(ctx context.Context, opt *SwarmInit) (token string, err error) {
	return c.wrapped.SwarmInit(ctx, opt.apply())
}

// SwarmJoin joins a node to the swarm.
func (c *Client) SwarmJoin(ctx context.Context, opt *SwarmJoin) error {
	return c.wrapped.SwarmJoin(ctx, opt.apply())
}

// SwarmLeave leaves the swarm
func (c *Client) SwarmLeave(ctx context.Context, force bool) error {
	return c.wrapped.SwarmLeave(ctx, force)
}

// SwarmInspect inspects the swarm
func (c *Client) SwarmInspect(ctx context.Context) (*response.Swarm, error) {
	resp, err := c.wrapped.SwarmInspect(ctx)
	if err != nil {
		return nil, err
	}
	return &response.Swarm{
		Swarm: resp,
	}, nil
}
