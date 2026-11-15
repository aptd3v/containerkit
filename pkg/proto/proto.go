// Package proto is the IP protocol vocabulary for Publish and Expose.
package proto

// Proto is tcp, udp, or sctp.
type Proto string

const (
	TCP  Proto = "tcp"
	UDP  Proto = "udp"
	SCTP Proto = "sctp"
)

// All returns every known IP protocol.
func All() []Proto { return []Proto{TCP, UDP, SCTP} }
