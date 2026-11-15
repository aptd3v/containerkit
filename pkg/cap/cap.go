// Package cap is the Linux capability vocabulary for CapAdd and CapDrop.
// Names match the kernel (CAP_* without the prefix), which is the form
// Docker Engine and Compose accept: cap.NET_ADMIN.
package cap

// Capability is a Linux capability name.
type Capability string

// Docker accepts ALL as every capability.
const ALL Capability = "ALL"

// Kernel capabilities from uapi/linux/capability.h (CAP_LAST_CAP = 40).
const (
	CHOWN              Capability = "CHOWN"
	DAC_OVERRIDE       Capability = "DAC_OVERRIDE"
	DAC_READ_SEARCH    Capability = "DAC_READ_SEARCH"
	FOWNER             Capability = "FOWNER"
	FSETID             Capability = "FSETID"
	KILL               Capability = "KILL"
	SETGID             Capability = "SETGID"
	SETUID             Capability = "SETUID"
	SETPCAP            Capability = "SETPCAP"
	LINUX_IMMUTABLE    Capability = "LINUX_IMMUTABLE"
	NET_BIND_SERVICE   Capability = "NET_BIND_SERVICE"
	NET_BROADCAST      Capability = "NET_BROADCAST"
	NET_ADMIN          Capability = "NET_ADMIN"
	NET_RAW            Capability = "NET_RAW"
	IPC_LOCK           Capability = "IPC_LOCK"
	IPC_OWNER          Capability = "IPC_OWNER"
	SYS_MODULE         Capability = "SYS_MODULE"
	SYS_RAWIO          Capability = "SYS_RAWIO"
	SYS_CHROOT         Capability = "SYS_CHROOT"
	SYS_PTRACE         Capability = "SYS_PTRACE"
	SYS_PACCT          Capability = "SYS_PACCT"
	SYS_ADMIN          Capability = "SYS_ADMIN"
	SYS_BOOT           Capability = "SYS_BOOT"
	SYS_NICE           Capability = "SYS_NICE"
	SYS_RESOURCE       Capability = "SYS_RESOURCE"
	SYS_TIME           Capability = "SYS_TIME"
	SYS_TTY_CONFIG     Capability = "SYS_TTY_CONFIG"
	MKNOD              Capability = "MKNOD"
	LEASE              Capability = "LEASE"
	AUDIT_WRITE        Capability = "AUDIT_WRITE"
	AUDIT_CONTROL      Capability = "AUDIT_CONTROL"
	SETFCAP            Capability = "SETFCAP"
	MAC_OVERRIDE       Capability = "MAC_OVERRIDE"
	MAC_ADMIN          Capability = "MAC_ADMIN"
	SYSLOG             Capability = "SYSLOG"
	WAKE_ALARM         Capability = "WAKE_ALARM"
	BLOCK_SUSPEND      Capability = "BLOCK_SUSPEND"
	AUDIT_READ         Capability = "AUDIT_READ"
	PERFMON            Capability = "PERFMON"
	BPF                Capability = "BPF"
	CHECKPOINT_RESTORE Capability = "CHECKPOINT_RESTORE"
)

// All returns every known capability, including ALL.
func All() []Capability {
	return []Capability{
		ALL,
		CHOWN, DAC_OVERRIDE, DAC_READ_SEARCH, FOWNER, FSETID, KILL,
		SETGID, SETUID, SETPCAP, LINUX_IMMUTABLE,
		NET_BIND_SERVICE, NET_BROADCAST, NET_ADMIN, NET_RAW,
		IPC_LOCK, IPC_OWNER,
		SYS_MODULE, SYS_RAWIO, SYS_CHROOT, SYS_PTRACE, SYS_PACCT,
		SYS_ADMIN, SYS_BOOT, SYS_NICE, SYS_RESOURCE, SYS_TIME, SYS_TTY_CONFIG,
		MKNOD, LEASE, AUDIT_WRITE, AUDIT_CONTROL, SETFCAP,
		MAC_OVERRIDE, MAC_ADMIN, SYSLOG, WAKE_ALARM, BLOCK_SUSPEND,
		AUDIT_READ, PERFMON, BPF, CHECKPOINT_RESTORE,
	}
}
