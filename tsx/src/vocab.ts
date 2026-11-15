// Vocabulary values for props and command options.
// Pass them where a component asks for a string (or, for cascade, a number).
// Names that would clash with a command live here with a longer name:
// pullPolicy, restartPolicy, waitFor, watchAction.

export const cap = {
	ALL: "ALL",
	CHOWN: "CHOWN",
	DAC_OVERRIDE: "DAC_OVERRIDE",
	DAC_READ_SEARCH: "DAC_READ_SEARCH",
	FOWNER: "FOWNER",
	FSETID: "FSETID",
	KILL: "KILL",
	SETGID: "SETGID",
	SETUID: "SETUID",
	SETPCAP: "SETPCAP",
	LINUX_IMMUTABLE: "LINUX_IMMUTABLE",
	NET_BIND_SERVICE: "NET_BIND_SERVICE",
	NET_BROADCAST: "NET_BROADCAST",
	NET_ADMIN: "NET_ADMIN",
	NET_RAW: "NET_RAW",
	IPC_LOCK: "IPC_LOCK",
	IPC_OWNER: "IPC_OWNER",
	SYS_MODULE: "SYS_MODULE",
	SYS_RAWIO: "SYS_RAWIO",
	SYS_CHROOT: "SYS_CHROOT",
	SYS_PTRACE: "SYS_PTRACE",
	SYS_PACCT: "SYS_PACCT",
	SYS_ADMIN: "SYS_ADMIN",
	SYS_BOOT: "SYS_BOOT",
	SYS_NICE: "SYS_NICE",
	SYS_RESOURCE: "SYS_RESOURCE",
	SYS_TIME: "SYS_TIME",
	SYS_TTY_CONFIG: "SYS_TTY_CONFIG",
	MKNOD: "MKNOD",
	LEASE: "LEASE",
	AUDIT_WRITE: "AUDIT_WRITE",
	AUDIT_CONTROL: "AUDIT_CONTROL",
	SETFCAP: "SETFCAP",
	MAC_OVERRIDE: "MAC_OVERRIDE",
	MAC_ADMIN: "MAC_ADMIN",
	SYSLOG: "SYSLOG",
	WAKE_ALARM: "WAKE_ALARM",
	BLOCK_SUSPEND: "BLOCK_SUSPEND",
	AUDIT_READ: "AUDIT_READ",
	PERFMON: "PERFMON",
	BPF: "BPF",
	CHECKPOINT_RESTORE: "CHECKPOINT_RESTORE",
} as const;

export const avail = {
	Active: "active",
	Pause: "pause",
	Drain: "drain",
} as const;

/** onExit values for composeUp() and composeStart(). */
export const cascade = {
	Ignore: 0,
	Stop: 1,
	Fail: 2,
} as const;

export const cgroupns = {
	Host: "host",
	Private: "private",
} as const;

export const depend = {
	Started: "service_started",
	Healthy: "service_healthy",
	CompletedSuccessfully: "service_completed_successfully",
} as const;

/** Disk-usage object filters for diskUsage(). */
export const df = {
	Container: "container",
	Image: "image",
	Volume: "volume",
	BuildCache: "build-cache",
} as const;

export const endpoint = {
	VIP: "vip",
	DNSRR: "dnsrr",
} as const;

export const failact = {
	Continue: "continue",
	Pause: "pause",
	Rollback: "rollback",
} as const;

export const ipc = {
	None: "none",
	Private: "private",
	Shareable: "shareable",
	Host: "host",
	container(id: string): `container:${string}` {
		return `container:${id}`;
	},
} as const;

export const isolation = {
	Default: "default",
	Process: "process",
	HyperV: "hyperv",
} as const;

export const logDriver = {
	JSONFile: "json-file",
	Local: "local",
	Syslog: "syslog",
	Journald: "journald",
	GELF: "gelf",
	Fluentd: "fluentd",
	AWSLogs: "awslogs",
	Splunk: "splunk",
	ETWLogs: "etwlogs",
	GCPLogs: "gcplogs",
	None: "none",
} as const;

export const netDriver = {
	Bridge: "bridge",
	Host: "host",
	None: "none",
	Overlay: "overlay",
	Macvlan: "macvlan",
	IPvlan: "ipvlan",
} as const;

export const netmode = {
	Bridge: "bridge",
	Host: "host",
	None: "none",
	Default: "default",
	container(id: string): `container:${string}` {
		return `container:${id}`;
	},
	service(name: string): `service:${string}` {
		return `service:${name}`;
	},
} as const;

export const oci = {
	V1_0: "1.0",
	V1_1: "1.1",
} as const;

export const order = {
	StartFirst: "start-first",
	StopFirst: "stop-first",
} as const;

export const pidns = {
	Host: "host",
	container(id: string): `container:${string}` {
		return `container:${id}`;
	},
} as const;

export const portmode = {
	Ingress: "ingress",
	Host: "host",
} as const;

export const progress = {
	Auto: "auto",
	Plain: "plain",
	TTY: "tty",
	Quiet: "quiet",
	JSON: "json",
} as const;

export const propagate = {
	Private: "private",
	RPrivate: "rprivate",
	Shared: "shared",
	RShared: "rshared",
	Slave: "slave",
	RSlave: "rslave",
} as const;

export const proto = {
	TCP: "tcp",
	UDP: "udp",
	SCTP: "sctp",
} as const;

export const pullPolicy = {
	Always: "always",
	Never: "never",
	Missing: "missing",
	IfNotPresent: "if_not_present",
	Build: "build",
	Refresh: "refresh",
	Daily: "daily",
	Weekly: "weekly",
	every(period: string): `every_${string}` {
		return `every_${period}`;
	},
} as const;

export const recursive = {
	Enabled: "enabled",
	Disabled: "disabled",
	Writable: "writable",
	ReadOnly: "readonly",
} as const;

export const restartPolicy = {
	No: "no",
	OnFailure: "on-failure",
	Always: "always",
	UnlessStopped: "unless-stopped",
} as const;

export const role = {
	Manager: "manager",
	Worker: "worker",
} as const;

export const scheme = {
	Unix: "unix",
	Npipe: "npipe",
	TCP: "tcp",
	HTTP: "http",
	HTTPS: "https",
	SSH: "ssh",
	FD: "fd",
} as const;

/** Bind-mount SELinux relabel. Shared is "z", Private is "Z". */
export const selinux = {
	Shared: "z",
	Private: "Z",
} as const;

export const signal = {
	SIGHUP: "SIGHUP",
	SIGINT: "SIGINT",
	SIGQUIT: "SIGQUIT",
	SIGILL: "SIGILL",
	SIGTRAP: "SIGTRAP",
	SIGABRT: "SIGABRT",
	SIGBUS: "SIGBUS",
	SIGFPE: "SIGFPE",
	SIGKILL: "SIGKILL",
	SIGUSR1: "SIGUSR1",
	SIGSEGV: "SIGSEGV",
	SIGUSR2: "SIGUSR2",
	SIGPIPE: "SIGPIPE",
	SIGALRM: "SIGALRM",
	SIGTERM: "SIGTERM",
	SIGSTKFLT: "SIGSTKFLT",
	SIGCHLD: "SIGCHLD",
	SIGCONT: "SIGCONT",
	SIGSTOP: "SIGSTOP",
	SIGTSTP: "SIGTSTP",
	SIGTTIN: "SIGTTIN",
	SIGTTOU: "SIGTTOU",
	SIGURG: "SIGURG",
	SIGXCPU: "SIGXCPU",
	SIGXFSZ: "SIGXFSZ",
	SIGVTALRM: "SIGVTALRM",
	SIGPROF: "SIGPROF",
	SIGWINCH: "SIGWINCH",
	SIGIO: "SIGIO",
	SIGPWR: "SIGPWR",
	SIGSYS: "SIGSYS",
} as const;

/** POSIX signal name. Dockerfile StopSignal and Compose stopSignal both take this. */
export type SignalName = (typeof signal)[keyof typeof signal];

export const svcmode = {
	Replicated: "replicated",
	Global: "global",
	ReplicatedJob: "replicated-job",
	GlobalJob: "global-job",
} as const;

export const userns = {
	Host: "host",
} as const;

export const uts = {
	Host: "host",
} as const;

export const vol = {
	Bind: "bind",
	Volume: "volume",
	Tmpfs: "tmpfs",
	Pipe: "npipe",
	Cluster: "cluster",
	Image: "image",
} as const;

export const waitFor = {
	NotRunning: "not-running",
	NextExit: "next-exit",
	Removed: "removed",
} as const;

export const watchAction = {
	Sync: "sync",
	Rebuild: "rebuild",
	SyncRestart: "sync+restart",
	SyncExec: "sync+exec",
	Restart: "restart",
} as const;

/** Every vocabulary group. Use this when a short name would clash with a command. */
export const vocab = {
	cap,
	avail,
	cascade,
	cgroupns,
	depend,
	df,
	endpoint,
	failact,
	ipc,
	isolation,
	logDriver,
	netDriver,
	netmode,
	oci,
	order,
	pidns,
	portmode,
	progress,
	propagate,
	proto,
	pullPolicy,
	recursive,
	restartPolicy,
	role,
	scheme,
	selinux,
	signal,
	svcmode,
	userns,
	uts,
	vol,
	waitFor,
	watchAction,
} as const;
