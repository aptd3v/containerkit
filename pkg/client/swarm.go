package client

import (
	"context"

	"github.com/aptd3v/go-contain/pkg/client/options/swarm/swarminit"
	"github.com/aptd3v/go-contain/pkg/client/options/swarm/swarmjoin"
	"github.com/aptd3v/go-contain/pkg/client/response"
	"github.com/docker/docker/api/types/swarm"
)

// SwarmInit initializes the swarm.
func (c *Client) SwarmInit(ctx context.Context, setters ...swarminit.SetSwarmInitOption) (token string, err error) {
	o := swarm.InitRequest{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return "", err
		}
	}
	return c.wrapped.SwarmInit(ctx, o)
}

// SwarmJoin joins a node to the swarm.
func (c *Client) SwarmJoin(ctx context.Context, setters ...swarmjoin.SetSwarmJoinOption) error {
	o := swarm.JoinRequest{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return err
		}
	}
	return c.wrapped.SwarmJoin(ctx, o)
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
