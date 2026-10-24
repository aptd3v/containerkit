package client

import sdk "github.com/moby/moby/client"

// NetworkCreate is [sdk.NetworkCreateOptions].
type NetworkCreate sdk.NetworkCreateOptions

func (o *NetworkCreate) apply() sdk.NetworkCreateOptions { return sdk.NetworkCreateOptions(apply(o)) }

// NetworkConnect is [sdk.NetworkConnectOptions].
type NetworkConnect sdk.NetworkConnectOptions

func (o *NetworkConnect) apply() sdk.NetworkConnectOptions {
	return sdk.NetworkConnectOptions(apply(o))
}

// NetworkInspect is [sdk.NetworkInspectOptions].
type NetworkInspect sdk.NetworkInspectOptions

func (o *NetworkInspect) apply() sdk.NetworkInspectOptions {
	return sdk.NetworkInspectOptions(apply(o))
}

// NetworkList is [sdk.NetworkListOptions].
type NetworkList sdk.NetworkListOptions

func (o *NetworkList) apply() sdk.NetworkListOptions { return sdk.NetworkListOptions(apply(o)) }

// NetworkPrune is [sdk.NetworkPruneOptions].
type NetworkPrune sdk.NetworkPruneOptions

func (o *NetworkPrune) apply() sdk.NetworkPruneOptions { return sdk.NetworkPruneOptions(apply(o)) }
