package client

import (
	"context"

	"github.com/aptd3v/containerkit/pkg/client/response"
	"github.com/moby/moby/api/types/swarm"
	sdk "github.com/moby/moby/client"
)

// VolumeCreate creates a volume in the docker host.
func (c *Client) VolumeCreate(ctx context.Context, opt *VolumeCreate) (*response.Volume, error) {
	created, err := c.wrapped.VolumeCreate(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.Volume{Volume: created.Volume}, nil
}

// VolumeInspect returns the information about a specific volume in the docker host.
func (c *Client) VolumeInspect(ctx context.Context, name string) (*response.Volume, error) {
	inspected, err := c.wrapped.VolumeInspect(ctx, name, sdk.VolumeInspectOptions{})
	if err != nil {
		return nil, err
	}
	return &response.Volume{Volume: inspected.Volume}, nil
}

// VolumeInspectWithRaw returns the information about a specific volume in the docker host and its raw representation.
func (c *Client) VolumeInspectWithRaw(ctx context.Context, name string) (*response.Volume, []byte, error) {
	inspected, err := c.wrapped.VolumeInspect(ctx, name, sdk.VolumeInspectOptions{})
	if err != nil {
		return nil, nil, err
	}
	return &response.Volume{Volume: inspected.Volume}, append([]byte(nil), inspected.Raw...), nil
}

// VolumeList returns the volumes configured in the docker host.
func (c *Client) VolumeList(ctx context.Context, opt *VolumeList) (*response.VolumeList, error) {
	listed, err := c.wrapped.VolumeList(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	volumes := make([]*response.Volume, len(listed.Items))
	for i, item := range listed.Items {
		volumes[i] = &response.Volume{Volume: item}
	}
	return &response.VolumeList{Volumes: volumes, Warnings: listed.Warnings}, nil
}

// VolumeRemove removes a volume from the docker host.
func (c *Client) VolumeRemove(ctx context.Context, name string, force bool) error {
	_, err := c.wrapped.VolumeRemove(ctx, name, sdk.VolumeRemoveOptions{Force: force})
	return err
}

// VolumeUpdate updates a volume. This only works for Cluster Volumes, and only some fields can be updated.
func (c *Client) VolumeUpdate(ctx context.Context, name string, swarmVersionIndex uint64, opt *VolumeUpdate) error {
	op := opt.apply()
	op.Version = swarm.Version{Index: swarmVersionIndex}
	_, err := c.wrapped.VolumeUpdate(ctx, name, op)
	return err
}

// VolumesPrune requests the daemon to delete unused data.
func (c *Client) VolumesPrune(ctx context.Context, opt *VolumePrune) (*response.VolumePruneReport, error) {
	pruned, err := c.wrapped.VolumePrune(ctx, opt.apply())
	if err != nil {
		return nil, err
	}
	return &response.VolumePruneReport{PruneReport: pruned.Report}, nil
}
