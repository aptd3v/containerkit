package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/response"
	"github.com/docker/docker/api/types/swarm"
)

// VolumeCreate creates a volume in the docker host.
func (c *Client) VolumeCreate(ctx context.Context, opt *VolumeCreate) (*response.Volume, error) {
	v, err := c.wrapped.VolumeCreate(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.Volume{Volume: v}, nil
}

// VolumeInspect returns the information about a specific volume in the docker host.
func (c *Client) VolumeInspect(ctx context.Context, name string) (*response.Volume, error) {
	v, err := c.wrapped.VolumeInspect(ctx, name)
	if err != nil {
		return nil, err
	}

	return &response.Volume{Volume: v}, nil
}

// VolumeInspectWithRaw returns the information about a specific volume in the docker host and its raw representation
func (c *Client) VolumeInspectWithRaw(ctx context.Context, name string) (*response.Volume, []byte, error) {
	v, b, err := c.wrapped.VolumeInspectWithRaw(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	return &response.Volume{Volume: v}, b, nil
}

// VolumeList returns the volumes configured in the docker host.
func (c *Client) VolumeList(ctx context.Context, opt *VolumeList) (*response.VolumeList, error) {
	v, err := c.wrapped.VolumeList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	volumes := make([]*response.Volume, len(v.Volumes))
	for i, v := range v.Volumes {
		if v == nil {
			continue
		}
		volumes[i] = &response.Volume{Volume: *v}
	}
	return &response.VolumeList{Volumes: volumes, Warnings: v.Warnings}, nil
}

// VolumeRemove removes a volume from the docker host.
func (c *Client) VolumeRemove(ctx context.Context, name string, force bool) error {
	return c.wrapped.VolumeRemove(ctx, name, force)
}

// VolumeUpdate updates a volume. This only works for Cluster Volumes, and only some fields can be updated.
func (c *Client) VolumeUpdate(ctx context.Context, name string, swarmVersionIndex uint64, opt *VolumeUpdate) error {
	return c.wrapped.VolumeUpdate(ctx, name, swarm.Version{Index: swarmVersionIndex}, opt.apply())
}

// VolumesPrune requests the daemon to delete unused data
func (c *Client) VolumesPrune(ctx context.Context, opt *VolumePrune) (*response.VolumePruneReport, error) {
	prune, err := c.wrapped.VolumesPrune(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.VolumePruneReport{PruneReport: prune}, nil
}
