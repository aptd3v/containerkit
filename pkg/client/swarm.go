package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/response"
	sdk "github.com/moby/moby/client"
)

// SwarmInit initializes the swarm and returns the node ID.
func (c *Client) SwarmInit(ctx context.Context, opt *SwarmInit) (string, error) {
	op, err := opt.apply()
	if err != nil {
		return "", err
	}
	res, err := c.wrapped.SwarmInit(ctx, op)
	if err != nil {
		return "", err
	}
	return res.NodeID, nil
}

// SwarmJoin joins a node to the swarm.
func (c *Client) SwarmJoin(ctx context.Context, opt *SwarmJoin) error {
	_, err := c.wrapped.SwarmJoin(ctx, opt.apply())
	return err
}

// SwarmLeave leaves the swarm.
func (c *Client) SwarmLeave(ctx context.Context, force bool) error {
	_, err := c.wrapped.SwarmLeave(ctx, sdk.SwarmLeaveOptions{Force: force})
	return err
}

// SwarmUpdate updates the swarm spec and can rotate tokens or the unlock key.
func (c *Client) SwarmUpdate(ctx context.Context, opt *SwarmUpdate) error {
	_, err := c.wrapped.SwarmUpdate(ctx, opt.apply())
	return err
}

// SwarmUnlock unlocks a locked swarm manager.
func (c *Client) SwarmUnlock(ctx context.Context, opt *SwarmUnlock) error {
	_, err := c.wrapped.SwarmUnlock(ctx, opt.apply())
	return err
}

// SwarmGetUnlockKey returns the swarm manager unlock key.
func (c *Client) SwarmGetUnlockKey(ctx context.Context) (string, error) {
	res, err := c.wrapped.SwarmGetUnlockKey(ctx)
	if err != nil {
		return "", err
	}
	return res.Key, nil
}

// SwarmInspect inspects the swarm.
func (c *Client) SwarmInspect(ctx context.Context) (*response.Swarm, error) {
	resp, err := c.wrapped.SwarmInspect(ctx, sdk.SwarmInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Swarm{Swarm: resp.Swarm}, nil
}
