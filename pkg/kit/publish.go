package kit

import (
	"fmt"
	"strconv"
)

func (p *port) setProtocol(v Proto) error {
	switch v {
	case TCP, UDP, SCTP, "":
		p.protocol.assign(v)
		return nil
	default:
		return cfgErr("", "protocol", "Protocol", fmt.Errorf("%w: protocol %q", errInvalid, v))
	}
}

func (p *port) setHostIP(ip string) error {
	p.hostIP.assign(ip)
	return nil
}

func (p *port) setHostPort(n uint16) error {
	p.hostPort.assign(n)
	return nil
}

func (p *port) setContainerPort(n uint16) error {
	p.containerPort.assign(n)
	return nil
}

// ProtocolOpt is the value returned by [Protocol].
// It implements [ComposePortOption] and [PortOption].
type ProtocolOpt struct{ p Proto }

// HostIPOpt is the value returned by [HostIP].
// It implements [PortOption].
type HostIPOpt struct{ ip string }

// HostPortOpt is the value returned by [HostPort].
// It implements [PortOption].
type HostPortOpt struct{ n uint16 }

// ContainerPortOpt is the value returned by [ContainerPort].
// It implements [PortOption].
type ContainerPortOpt struct{ n uint16 }

// Protocol sets the published-port protocol in a [Publish] scope.
func Protocol(p Proto) ProtocolOpt { return ProtocolOpt{p: p} }

func (o ProtocolOpt) applyPort(p *port) error { return p.setProtocol(o.p) }

// HostIP sets the host bind address in a [Publish] scope. Empty is explicit.
func HostIP(ip string) HostIPOpt { return HostIPOpt{ip: ip} }

func (o HostIPOpt) applyPort(p *port) error { return p.setHostIP(o.ip) }

// HostPort sets the host port in a [Publish] scope. Zero is explicit.
func HostPort(n uint16) HostPortOpt { return HostPortOpt{n: n} }

func (o HostPortOpt) applyPort(p *port) error { return p.setHostPort(o.n) }

// ContainerPort sets the container port in a [Publish] scope.
func ContainerPort(n uint16) ContainerPortOpt { return ContainerPortOpt{n: n} }

func (o ContainerPortOpt) applyPort(p *port) error { return p.setContainerPort(o.n) }

// PublishOpt is the value returned by [Publish].
// It implements [ContainerOption] and [ServiceOption].
type PublishOpt struct{ opts []PortOption }

// Publish adds a published port. Nested setters determine host/container ports.
func Publish(opts ...PortOption) PublishOpt { return PublishOpt{opts: opts} }

func (o PublishOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o PublishOpt) applyService(s *Service) error     { return o.apply(&s.spec) }

func (o PublishOpt) apply(s *spec) error {
	var p port
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("publish.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyPort(&p); err != nil {
			errs = append(errs, wrapPath("publish", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if err := p.validate(); err != nil {
		return wrapPath("publish", err)
	}
	return s.addPublish(p)
}

func (p port) validate() error {
	if !p.containerPort.ok() {
		return cfgErr("", "container_port", "ContainerPort", fmt.Errorf("%w: container port", errRequired))
	}
	if p.protocol.ok() {
		switch p.protocol.val {
		case TCP, UDP, SCTP, "":
		default:
			return cfgErr("", "protocol", "Protocol", fmt.Errorf("%w: protocol %q", errInvalid, p.protocol.val))
		}
	}
	return nil
}

func (p port) proto() Proto {
	if p.protocol.ok() && p.protocol.val != "" {
		return p.protocol.val
	}
	return TCP
}

func (p port) hostIPOrDefault() string {
	if p.hostIP.ok() {
		return p.hostIP.val
	}
	return "0.0.0.0"
}

func (p port) hostPortString() string {
	if p.published.ok() {
		return p.published.val
	}
	if !p.hostPort.ok() {
		return ""
	}
	return strconv.FormatUint(uint64(p.hostPort.val), 10)
}
