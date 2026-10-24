package client

import sdk "github.com/moby/moby/client"

// VolumeCreate is [sdk.VolumeCreateOptions].
type VolumeCreate sdk.VolumeCreateOptions

func (o *VolumeCreate) apply() sdk.VolumeCreateOptions { return sdk.VolumeCreateOptions(apply(o)) }

// VolumeList is [sdk.VolumeListOptions].
type VolumeList sdk.VolumeListOptions

func (o *VolumeList) apply() sdk.VolumeListOptions { return sdk.VolumeListOptions(apply(o)) }

// VolumeUpdate is [sdk.VolumeUpdateOptions].
// Version is filled from the method argument.
type VolumeUpdate sdk.VolumeUpdateOptions

func (o *VolumeUpdate) apply() sdk.VolumeUpdateOptions { return sdk.VolumeUpdateOptions(apply(o)) }

// VolumePrune is [sdk.VolumePruneOptions].
type VolumePrune sdk.VolumePruneOptions

func (o *VolumePrune) apply() sdk.VolumePruneOptions { return sdk.VolumePruneOptions(apply(o)) }
