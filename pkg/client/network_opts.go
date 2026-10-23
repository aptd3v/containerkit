package client

import (
	"net/netip"

	"github.com/moby/moby/api/types/network"
	sdk "github.com/moby/moby/client"
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

func (o *NetworkCreate) apply() (sdk.NetworkCreateOptions, error) {
	if o == nil {
		return sdk.NetworkCreateOptions{}, nil
	}
	op := sdk.NetworkCreateOptions{
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
	if o.IPAM == nil {
		return op, nil
	}
	cfg := make([]network.IPAMConfig, 0, len(o.IPAM.Config))
	for _, c := range o.IPAM.Config {
		item := network.IPAMConfig{}
		if c.Subnet != "" {
			prefix, err := netip.ParsePrefix(c.Subnet)
			if err != nil {
				return sdk.NetworkCreateOptions{}, err
			}
			item.Subnet = prefix
		}
		if c.IPRange != "" {
			prefix, err := netip.ParsePrefix(c.IPRange)
			if err != nil {
				return sdk.NetworkCreateOptions{}, err
			}
			item.IPRange = prefix
		}
		if c.Gateway != "" {
			addr, err := netip.ParseAddr(c.Gateway)
			if err != nil {
				return sdk.NetworkCreateOptions{}, err
			}
			item.Gateway = addr
		}
		if len(c.Auxiliary) > 0 {
			item.AuxAddress = make(map[string]netip.Addr, len(c.Auxiliary))
			for name, addr := range c.Auxiliary {
				parsed, err := netip.ParseAddr(addr)
				if err != nil {
					return sdk.NetworkCreateOptions{}, err
				}
				item.AuxAddress[name] = parsed
			}
		}
		cfg = append(cfg, item)
	}
	op.IPAM = &network.IPAM{
		Driver:  o.IPAM.Driver,
		Options: o.IPAM.Options,
		Config:  cfg,
	}
	return op, nil
}

// NetworkConnect is options for NetworkConnect.
type NetworkConnect struct {
	Container string
	Endpoint  *network.EndpointSettings
}

func (o *NetworkConnect) apply() sdk.NetworkConnectOptions {
	if o == nil {
		return sdk.NetworkConnectOptions{}
	}
	return sdk.NetworkConnectOptions{
		Container:      o.Container,
		EndpointConfig: o.Endpoint,
	}
}

// NetworkInspect is options for NetworkInspect.
type NetworkInspect struct {
	Scope   string
	Verbose bool
}

func (o *NetworkInspect) apply() sdk.NetworkInspectOptions {
	if o == nil {
		return sdk.NetworkInspectOptions{}
	}
	return sdk.NetworkInspectOptions{
		Scope:   o.Scope,
		Verbose: o.Verbose,
	}
}

// NetworkList is options for NetworkList.
type NetworkList struct {
	Filters []Filter
}

func (o *NetworkList) apply() sdk.NetworkListOptions {
	op := sdk.NetworkListOptions{Filters: argsFromFilters(nil)}
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

func (o *NetworkPrune) apply() sdk.Filters {
	if o == nil {
		return argsFromFilters(nil)
	}
	return argsFromFilters(o.Filters)
}
