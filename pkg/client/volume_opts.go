package client

import (
	"github.com/moby/moby/api/types/volume"
	sdk "github.com/moby/moby/client"
)

// VolumeCreate is options for VolumeCreate.
type VolumeCreate struct {
	Name              string
	Driver            string
	DriverOpts        map[string]string
	Labels            map[string]string
	ClusterVolumeSpec *volume.ClusterVolumeSpec
}

func (o *VolumeCreate) apply() sdk.VolumeCreateOptions {
	if o == nil {
		return sdk.VolumeCreateOptions{}
	}
	return sdk.VolumeCreateOptions{
		Name:              o.Name,
		Driver:            o.Driver,
		DriverOpts:        o.DriverOpts,
		Labels:            o.Labels,
		ClusterVolumeSpec: o.ClusterVolumeSpec,
	}
}

// VolumeList is options for VolumeList.
type VolumeList struct {
	Filters []Filter
}

func (o *VolumeList) apply() sdk.VolumeListOptions {
	op := sdk.VolumeListOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op
	}
	op.Filters = argsFromFilters(o.Filters)
	return op
}

// VolumeUpdate is options for VolumeUpdate.
type VolumeUpdate struct {
	ClusterVolumeSpec *volume.ClusterVolumeSpec
}

func (o *VolumeUpdate) apply() sdk.VolumeUpdateOptions {
	if o == nil {
		return sdk.VolumeUpdateOptions{}
	}
	return sdk.VolumeUpdateOptions{Spec: o.ClusterVolumeSpec}
}

// VolumePrune is options for VolumesPrune.
type VolumePrune struct {
	All     bool
	Filters []Filter
}

func (o *VolumePrune) apply() sdk.VolumePruneOptions {
	if o == nil {
		return sdk.VolumePruneOptions{Filters: argsFromFilters(nil)}
	}
	return sdk.VolumePruneOptions{All: o.All, Filters: argsFromFilters(o.Filters)}
}
