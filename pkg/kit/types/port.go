package types

import (
	"github.com/docker/go-connections/nat"
)

// Protocol is a published port protocol, as used by [nat.Port].
type Protocol string

const (
	TCP  Protocol = "tcp"
	UDP  Protocol = "udp"
	SCTP Protocol = "sctp"
)

// PortBinding publishes one container port on the host.
// Host() is the [nat.PortBinding]; Protocol and ContainerPort become the [nat.Port] map key.
type PortBinding struct {
	Protocol      Protocol
	HostIP        string
	HostPort      string
	ContainerPort string
}

// Host returns the Docker [nat.PortBinding].
func (p PortBinding) Host() nat.PortBinding {
	return nat.PortBinding{HostIP: p.HostIP, HostPort: p.HostPort}
}
func NewPortBinding(protocol Protocol, hostIP string, hostPort string, containerPort string) PortBinding {
	return PortBinding{Protocol: protocol, HostIP: hostIP, HostPort: hostPort, ContainerPort: containerPort}
}
