package client

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

// NetworkCreate is options for NetworkCreate.
type NetworkCreate struct {
	Driver     string
	Scope      string
	EnableIPv4 *bool
	EnableIPv6 *bool
	Internal   bool
	Attachable bool
	Ingress    bool
	ConfigOnly bool
	Options    map[string]string
	Labels     map[string]string
	IPAM       *NetworkIPAM
}

// NetworkIPAM is IPAM configuration for NetworkCreate.
type NetworkIPAM struct {
	Driver  string
	Options map[string]string
	Config  []NetworkIPAMConfig
}

// NetworkIPAMConfig is a single IPAM pool.
type NetworkIPAMConfig struct {
	Subnet    string
	IPRange   string
	Gateway   string
	Auxiliary map[string]string
}

func (o *NetworkCreate) apply() network.CreateOptions {
	if o == nil {
		return network.CreateOptions{}
	}
	op := network.CreateOptions{
		Driver:     o.Driver,
		Scope:      o.Scope,
		EnableIPv4: o.EnableIPv4,
		EnableIPv6: o.EnableIPv6,
		Internal:   o.Internal,
		Attachable: o.Attachable,
		Ingress:    o.Ingress,
		ConfigOnly: o.ConfigOnly,
		Options:    o.Options,
		Labels:     o.Labels,
	}
	if o.IPAM != nil {
		cfg := make([]network.IPAMConfig, 0, len(o.IPAM.Config))
		for _, c := range o.IPAM.Config {
			cfg = append(cfg, network.IPAMConfig{
				Subnet:     c.Subnet,
				IPRange:    c.IPRange,
				Gateway:    c.Gateway,
				AuxAddress: c.Auxiliary,
			})
		}
		op.IPAM = &network.IPAM{
			Driver:  o.IPAM.Driver,
			Options: o.IPAM.Options,
			Config:  cfg,
		}
	}
	return op
}

// NetworkConnect is options for NetworkConnect.
type NetworkConnect struct {
	Container string
	Endpoint  *network.EndpointSettings
}

func (o *NetworkConnect) apply() network.ConnectOptions {
	if o == nil {
		return network.ConnectOptions{}
	}
	return network.ConnectOptions{
		Container:      o.Container,
		EndpointConfig: o.Endpoint,
	}
}

// NetworkInspect is options for NetworkInspect.
type NetworkInspect struct {
	Scope   string
	Verbose bool
}

func (o *NetworkInspect) apply() network.InspectOptions {
	if o == nil {
		return network.InspectOptions{}
	}
	return network.InspectOptions{
		Scope:   o.Scope,
		Verbose: o.Verbose,
	}
}

// NetworkList is options for NetworkList.
type NetworkList struct {
	Filters []Filter
}

func (o *NetworkList) apply() network.ListOptions {
	op := network.ListOptions{Filters: argsFromFilters(nil)}
	if o == nil {
		return op
	}
	op.Filters = argsFromFilters(o.Filters)
	return op
}

// NetworkPrune is options for NetworksPrune.
type NetworkPrune struct {
	Filters []Filter
}

func (o *NetworkPrune) apply() filters.Args {
	if o == nil {
		return argsFromFilters(nil)
	}
	return argsFromFilters(o.Filters)
}
