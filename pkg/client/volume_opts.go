package client

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

// VolumeCreate is options for VolumeCreate.
type VolumeCreate struct {
	Name              string
	Driver            string
	DriverOpts        map[string]string
	Labels            map[string]string
	ClusterVolumeSpec *volume.ClusterVolumeSpec
}

func (o *VolumeCreate) apply() volume.CreateOptions {
	if o == nil {
		return volume.CreateOptions{}
	}
	return volume.CreateOptions{
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

func (o *VolumeList) apply() volume.ListOptions {
	op := volume.ListOptions{Filters: argsFromFilters(nil)}
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

func (o *VolumeUpdate) apply() volume.UpdateOptions {
	if o == nil {
		return volume.UpdateOptions{}
	}
	return volume.UpdateOptions{Spec: o.ClusterVolumeSpec}
}

// VolumePrune is options for VolumesPrune.
type VolumePrune struct {
	Filters []Filter
}

func (o *VolumePrune) apply() filters.Args {
	if o == nil {
		return argsFromFilters(nil)
	}
	return argsFromFilters(o.Filters)
}
