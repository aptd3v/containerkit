package client

import (
	"context"

	"github.com/aptd3v/go-contain/pkg/client/options/vco"
	"github.com/aptd3v/go-contain/pkg/client/options/vlo"
	"github.com/aptd3v/go-contain/pkg/client/options/vpo"
	"github.com/aptd3v/go-contain/pkg/client/options/vuo"
	"github.com/aptd3v/go-contain/pkg/client/response"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
)

func (c *Client) VolumeCreate(ctx context.Context, setters ...vco.SetVolumeCreateOption) (*response.Volume, error) {
	o := volume.CreateOptions{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return nil, err
		}
	}
	v, err := c.wrapped.VolumeCreate(ctx, o)
	if err != nil {
		return nil, err
	}
	return &response.Volume{Volume: v}, nil
}

func (c *Client) VolumeInspect(ctx context.Context, name string) (*response.Volume, error) {
	v, err := c.wrapped.VolumeInspect(ctx, name)
	if err != nil {
		return nil, err
	}

	return &response.Volume{Volume: v}, nil
}
func (c *Client) VolumeInspectWithRaw(ctx context.Context, name string) (*response.Volume, []byte, error) {
	v, b, err := c.wrapped.VolumeInspectWithRaw(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	return &response.Volume{Volume: v}, b, nil
}

func (c *Client) VolumeList(ctx context.Context, setters ...vlo.SetVolumeListOption) (*response.VolumeList, error) {
	o := volume.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		setter(o)
	}
	v, err := c.wrapped.VolumeList(ctx, o)
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
func (c *Client) VolumeRemove(ctx context.Context, name string, force bool) error {
	return c.wrapped.VolumeRemove(ctx, name, force)
}
func (c *Client) VolumeUpdate(ctx context.Context, name string, swarmVersionIndex uint64, setters ...vuo.SetVolumeUpdateOption) error {
	o := volume.UpdateOptions{}
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(&o); err != nil {
			return err
		}
	}
	return c.wrapped.VolumeUpdate(ctx, name, swarm.Version{Index: swarmVersionIndex}, o)
}
func (c *Client) VolumePrune(ctx context.Context, setters ...vpo.SetVolumePruneOption) (*response.VolumePruneReport, error) {
	filters := filters.NewArgs()
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		if err := setter(filters); err != nil {
			return nil, err
		}
	}
	prune, err := c.wrapped.VolumesPrune(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &response.VolumePruneReport{PruneReport: prune}, nil
}
