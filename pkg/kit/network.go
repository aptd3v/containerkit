package kit

import (
	"fmt"
	"net/netip"

	"github.com/aptd3v/containerkit/pkg/netmode"
)

func (n *netAttach) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("network", err)
	}
	n.name.assign(name)
	return nil
}

func (n netAttach) validate() error {
	if !n.name.ok() {
		return cfgErr("", "name", "Name", fmt.Errorf("%w: network name", errRequired))
	}
	return nil
}

// AliasOpt is the value returned by [Alias].
// It implements [NetworkOption].
type AliasOpt struct{ a string }

// Alias adds a network alias inside [Network].
func Alias(a string) AliasOpt { return AliasOpt{a: a} }

func (o AliasOpt) applyNetwork(n *netAttach) error {
	if o.a == "" {
		return cfgErr("", "alias", "Alias", fmt.Errorf("%w: alias", errEmpty))
	}
	n.aliases = append(n.aliases, o.a)
	return nil
}

// IPv4Opt is the value returned by [IPv4].
// It implements [NetworkOption].
type IPv4Opt struct{ ip string }

// IPv4 sets the static IPv4 address of a service, job, or container on one network.
//
// Put it on the Network element inside that service. The Network element under
// Compose defines the network and does not take this address. The address has
// to fall inside that network's subnet.
//
//	<Network name="backend" ipv4="10.0.0.10" />
func IPv4(ip string) IPv4Opt { return IPv4Opt{ip: ip} }

func (o IPv4Opt) applyNetwork(n *netAttach) error {
	if _, err := netip.ParseAddr(o.ip); err != nil {
		return cfgErr("", "ipv4", "IPv4", fmt.Errorf("%w: ipv4 %q", errInvalid, o.ip))
	}
	n.ipv4.assign(o.ip)
	return nil
}

// IPv6Opt is the value returned by [IPv6].
// It implements [NetworkOption].
type IPv6Opt struct{ ip string }

// IPv6 sets the static IPv6 address of a service, job, or container on one network.
//
// Put it on the Network element inside that service, next to ipv4. The address
// has to fall inside the project network's IPv6 subnet.
//
//	<Network name="backend" ipv6="fd00:1::10" />
func IPv6(ip string) IPv6Opt { return IPv6Opt{ip: ip} }

func (o IPv6Opt) applyNetwork(n *netAttach) error {
	if _, err := netip.ParseAddr(o.ip); err != nil {
		return cfgErr("", "ipv6", "IPv6", fmt.Errorf("%w: ipv6 %q", errInvalid, o.ip))
	}
	n.ipv6.assign(o.ip)
	return nil
}

// NetworkOpt is the value returned by [Network].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type NetworkOpt struct{ opts []NetworkOption }

// Network attaches a named network (Engine EndpointsConfig / Compose networks).
//
// Under Compose it defines the network: driver, subnet, and gateway. Under
// Service, Job, or Container it joins that network and can set ipv4, ipv6, and
// macAddress. A static address has to fall inside the subnet of the project
// network with the same name.
func Network(opts ...NetworkOption) NetworkOpt { return NetworkOpt{opts: opts} }

func (o NetworkOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o NetworkOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o NetworkOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o NetworkOpt) apply(s *spec) error {
	var n netAttach
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("network.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyNetwork(&n); err != nil {
			errs = append(errs, wrapPath("network", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if err := n.validate(); err != nil {
		return wrapPath("network", err)
	}
	s.networks = append(s.networks, n)
	return nil
}

func (s *spec) setNetworkMode(mode string) error {
	if mode == "" {
		return cfgErr("", "network_mode", "NetworkMode", fmt.Errorf("%w: network mode", errEmpty))
	}
	s.networkMode.assign(mode)
	return nil
}

// NetworkMode sets Engine NetworkMode / Compose network_mode (bridge, host, none, container:id).
// Inside [Build] it maps to compose build.network.
func NetworkMode(mode netmode.Mode) NetworkModeOpt {
	return NetworkModeOpt{m: string(mode)}
}

// NetworkDisabled disables networking (Engine NetworkDisabled, Compose network_mode=none).
func NetworkDisabled(v bool) SpecOpt {
	return specFn(func(s *spec) error {
		s.networkDisabled.assign(v)
		return nil
	})
}

// DNS adds one DNS server, such as 1.1.1.1.
// Put it on the service or container. Repeat it for another server.
func DNS(addr string) SpecOpt {
	return specFn(func(s *spec) error {
		if addr == "" {
			return cfgErr("", "dns", "DNS", fmt.Errorf("%w: dns", errEmpty))
		}
		s.dns = append(s.dns, addr)
		return nil
	})
}

// DNSOption adds a DNS resolver option.
func DNSOption(opt string) SpecOpt {
	return specFn(func(s *spec) error {
		if opt == "" {
			return cfgErr("", "dns_option", "DNSOption", fmt.Errorf("%w: dns option", errEmpty))
		}
		s.dnsOptions = append(s.dnsOptions, opt)
		return nil
	})
}

// DNSSearch adds a DNS search domain.
func DNSSearch(d string) SpecOpt {
	return specFn(func(s *spec) error {
		if d == "" {
			return cfgErr("", "dns_search", "DNSSearch", fmt.Errorf("%w: dns search", errEmpty))
		}
		s.dnsSearch = append(s.dnsSearch, d)
		return nil
	})
}

// ExtraHost maps a hostname to an IP inside the container (Engine extra_hosts / Compose extra_hosts).
// Put it on the service or container. Inside [Build] it maps to compose build.extra_hosts.
//
//	<ExtraHost hostname="db" ip="10.0.0.2" />
func ExtraHost(hostname, ip string) ExtraHostOpt {
	return ExtraHostOpt{hostname: hostname, ip: ip}
}

// Link adds a legacy container link.
func Link(name string) SpecOpt {
	return specFn(func(s *spec) error {
		if name == "" {
			return cfgErr("", "link", "Link", fmt.Errorf("%w: link", errEmpty))
		}
		s.links = append(s.links, name)
		return nil
	})
}
