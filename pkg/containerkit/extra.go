package containerkit

import (
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/cc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/cc/health"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/hc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/hc/mount"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/nc"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/nc/endpoint"
	"github.com/aptd3v/containerkit/pkg/containerkit/internal/config/nc/endpoint/ipam"
)

// HealthCheck sets the container healthcheck from a Health value.
func (c *Container) HealthCheck(h Health) *Container {
	setters := []health.SetHealthcheckConfig{}
	if len(h.Test) > 0 {
		setters = append(setters, health.WithTest(h.Test...))
	}
	if h.TimeoutD != "" {
		setters = append(setters, health.WithTimeout(h.TimeoutD))
	} else {
		setters = append(setters, health.WithTimeout(h.Timeout))
	}
	if h.IntervalD != "" {
		setters = append(setters, health.WithInterval(h.IntervalD))
	} else {
		setters = append(setters, health.WithInterval(h.Interval))
	}
	if h.StartPeriodD != "" {
		setters = append(setters, health.WithStartPeriod(h.StartPeriodD))
	} else {
		setters = append(setters, health.WithStartPeriod(h.StartPeriod))
	}
	setters = append(setters, health.WithRetries(h.Retries))
	return c.WithContainerConfig(cc.WithHealthCheck(setters...))
}

// DisabledHealthCheck disables the healthcheck.
func (c *Container) DisabledHealthCheck() *Container {
	return c.WithContainerConfig(cc.WithDisabledHealthCheck())
}

// RestartUnlessStopped is an alias for RestartPolicyUnlessStopped.
func (c *Container) RestartUnlessStopped() *Container {
	return c.RestartPolicyUnlessStopped()
}

// Mount adds a host mount point.
func (c *Container) Mount(m Mount) *Container {
	t := m.Type
	if t == "" {
		t = MountBind
	}
	setters := []mount.SetMountConfig{
		mount.WithType(t),
		mount.WithSource(m.Source),
		mount.WithTarget(m.Target),
	}
	if m.ReadOnly {
		setters = append(setters, mount.WithReadOnly())
	} else {
		setters = append(setters, mount.WithReadWrite())
	}
	return c.WithHostConfig(hc.WithMountPoint(setters...))
}

// Endpoint joins the named network. Optional EndpointOpts customize the endpoint.
func (c *Container) Endpoint(name string, opts ...EndpointOpts) *Container {
	var setters []endpoint.SetEndpointConfig
	if len(opts) > 0 {
		o := opts[0]
		if len(o.Aliases) > 0 {
			setters = append(setters, endpoint.WithAliases(o.Aliases...))
		}
		if len(o.Links) > 0 {
			setters = append(setters, endpoint.WithLinks(o.Links...))
		}
		if o.MacAddress != "" {
			setters = append(setters, endpoint.WithMacAddress(o.MacAddress))
		}
		for k, v := range o.DriverOptions {
			setters = append(setters, endpoint.WithDriverOptions(k, v))
		}
		if o.IPv4 != "" || o.IPv6 != "" || len(o.LinkLocalIPs) > 0 {
			var ipamSet []ipam.SetIPAMConfig
			if o.IPv4 != "" {
				ipamSet = append(ipamSet, ipam.WithIPv4Address(o.IPv4))
			}
			if o.IPv6 != "" {
				ipamSet = append(ipamSet, ipam.WithIPv6Address(o.IPv6))
			}
			if len(o.LinkLocalIPs) > 0 {
				ipamSet = append(ipamSet, ipam.WithLinkLocalIPs(o.LinkLocalIPs...))
			}
			setters = append(setters, endpoint.WithIPAMConfig(ipamSet...))
		}
	}
	return c.WithNetworkConfig(nc.WithEndpoint(name, setters...))
}
