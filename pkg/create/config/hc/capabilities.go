package hc

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
)

// Capability is a string that represents a capability
type Capability string

// Default capabilities granted to containers
const (
	// Write records to audit log
	AUDIT_WRITE Capability = "AUDIT_WRITE"
	// Make arbitrary changes to file ownership
	CHOWN Capability = "CHOWN"
	// Bypass file read/write/execute permission checks
	DAC_OVERRIDE Capability = "DAC_OVERRIDE"
	// Bypass file ownership checks
	FOWNER Capability = "FOWNER"
	// Set process UID/GID
	FSETID Capability = "FSETID"
	// Terminate processes
	KILL Capability = "KILL"
	// Create special files
	MKNOD Capability = "MKNOD"
	// Bind to low-numbered ports
	NET_BIND_SERVICE Capability = "NET_BIND_SERVICE"
	// Use raw sockets
	NET_RAW Capability = "NET_RAW"
	// Set file capabilities
	SETFCAP Capability = "SETFCAP"
	// Set group ID
	SETGID Capability = "SETGID"
	// Set process capabilities
	SETPCAP Capability = "SETPCAP"
	// Set user ID
	SETUID Capability = "SETUID"
	// Use chroot()
	SYS_CHROOT Capability = "SYS_CHROOT"
)

// Audit related capabilities
const (
	// Configure auditing and audit rules
	AUDIT_CONTROL Capability = "AUDIT_CONTROL"
	// Read auditing and audit rules
	AUDIT_READ Capability = "AUDIT_READ"
)

// System administration capabilities
const (
	// Employ block devices
	BLOCK_SUSPEND Capability = "BLOCK_SUSPEND"
	// Use BPF (Berkeley Packet Filter)
	BPF Capability = "BPF"
	// Use process checkpoint/restore
	CHECKPOINT_RESTORE Capability = "CHECKPOINT_RESTORE"
	// Read files and directories
	DAC_READ_SEARCH Capability = "DAC_READ_SEARCH"
	// Perform admin tasks, like mount filesystems
	SYS_ADMIN Capability = "SYS_ADMIN"
	// Use reboot()
	SYS_BOOT Capability = "SYS_BOOT"
	// Load and unload kernel modules
	SYS_MODULE Capability = "SYS_MODULE"
	// Configure process accounting
	SYS_PACCT Capability = "SYS_PACCT"
	// Perform I/O port operations
	SYS_RAWIO Capability = "SYS_RAWIO"
	// Set system time
	SYS_TIME Capability = "SYS_TIME"
	// Configure tty devices
	SYS_TTY_CONFIG Capability = "SYS_TTY_CONFIG"
	// Configure syslog
	SYSLOG Capability = "SYSLOG"
)

// Process and resource management capabilities
const (
	// Lock memory
	IPC_LOCK Capability = "IPC_LOCK"
	// Become IPC namespace owner
	IPC_OWNER Capability = "IPC_OWNER"
	// Establish leases on filesystem objects
	LEASE Capability = "LEASE"
	// Set immutable attributes on files
	LINUX_IMMUTABLE Capability = "LINUX_IMMUTABLE"
	// Modify priority for arbitrary processes
	SYS_NICE Capability = "SYS_NICE"
	// Trace arbitrary processes using ptrace
	SYS_PTRACE Capability = "SYS_PTRACE"
	// Override resource limits
	SYS_RESOURCE Capability = "SYS_RESOURCE"
	// Set alarm to wake system
	WAKE_ALARM Capability = "WAKE_ALARM"
)

// Network administration capabilities
const (
	// Perform network administration tasks
	NET_ADMIN Capability = "NET_ADMIN"
	// Broadcast and listen to multicast
	NET_BROADCAST Capability = "NET_BROADCAST"
)

// Security policy capabilities
const (
	// Configure MAC (Mandatory Access Control) policy
	MAC_ADMIN Capability = "MAC_ADMIN"
	// Override MAC policy
	MAC_OVERRIDE Capability = "MAC_OVERRIDE"
	// Access perf_event Open() hypercall
	PERFMON Capability = "PERFMON"
)

// ALL all capabilities
const ALL Capability = "ALL"

// WithAddCapabilities adds capabilities to the host configuration for the container.
// parameters:
//   - caps: the capabilities to add
func WithAddedCapabilities(caps ...Capability) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.CapAdd == nil {
			opt.CapAdd = make(strslice.StrSlice, 0)
		}
		for _, cap := range caps {
			opt.CapAdd = append(opt.CapAdd, string(cap))
		}
		return nil
	}
}

// WithDropCapabilities drops capabilities from the host configuration for the container.
// parameters:
//   - caps: the capabilities to drop
func WithDroppedCapabilities(caps ...Capability) create.SetHostConfig {
	return func(opt *container.HostConfig) error {
		if opt.CapDrop == nil {
			opt.CapDrop = make(strslice.StrSlice, 0)
		}
		for _, cap := range caps {
			opt.CapDrop = append(opt.CapDrop, string(cap))
		}
		return nil
	}
}

// WithNetCapabilities adds only networking-related capabilities
//
// NET_ADMIN, NET_RAW, NET_BIND_SERVICE
func WithAddedNetCapabilities() create.SetHostConfig {
	return WithAddedCapabilities(NET_ADMIN, NET_RAW, NET_BIND_SERVICE)
}

// WithDropAllCapabilities drops all capabilities
func WithDroppedAllCapabilities() create.SetHostConfig {
	return WithDroppedCapabilities(ALL)
}

// WithDropSensitiveCapabilities drops a common set of sensitive capabilities
//
// SYS_ADMIN, SYS_MODULE, SYS_PTRACE, SYS_TIME, SYSLOG, MAC_ADMIN, MAC_OVERRIDE, CHECKPOINT_RESTORE
func WithDroppedSensitiveCapabilities() create.SetHostConfig {
	return WithDroppedCapabilities(
		SYS_ADMIN, SYS_MODULE, SYS_PTRACE, SYS_TIME, SYSLOG,
		MAC_ADMIN, MAC_OVERRIDE, CHECKPOINT_RESTORE,
	)
}
