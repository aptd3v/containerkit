package main

func vocabType(goType string) (string, bool) {
	switch goType {
	case "Proto", "proto.Proto":
		return `"tcp" | "udp" | "sctp"`, true
	case "Capability", "cap.Capability":
		return `"ALL" | "CHOWN" | "DAC_OVERRIDE" | "DAC_READ_SEARCH" | "FOWNER" | "FSETID" | "KILL" | "SETGID" | "SETUID" | "SETPCAP" | "LINUX_IMMUTABLE" | "NET_BIND_SERVICE" | "NET_BROADCAST" | "NET_ADMIN" | "NET_RAW" | "IPC_LOCK" | "IPC_OWNER" | "SYS_MODULE" | "SYS_RAWIO" | "SYS_CHROOT" | "SYS_PTRACE" | "SYS_PACCT" | "SYS_ADMIN" | "SYS_BOOT" | "SYS_NICE" | "SYS_RESOURCE" | "SYS_TIME" | "SYS_TTY_CONFIG" | "MKNOD" | "LEASE" | "AUDIT_WRITE" | "AUDIT_CONTROL" | "SETFCAP" | "MAC_OVERRIDE" | "MAC_ADMIN" | "SYSLOG" | "WAKE_ALARM" | "BLOCK_SUSPEND" | "AUDIT_READ" | "PERFMON" | "BPF" | "CHECKPOINT_RESTORE"`, true
	case "RestartPolicy", "restart.Policy":
		return `"no" | "on-failure" | "always" | "unless-stopped"`, true
	case "signal.Signal":
		return `"SIGHUP" | "SIGINT" | "SIGQUIT" | "SIGILL" | "SIGTRAP" | "SIGABRT" | "SIGBUS" | "SIGFPE" | "SIGKILL" | "SIGUSR1" | "SIGSEGV" | "SIGUSR2" | "SIGPIPE" | "SIGALRM" | "SIGTERM" | "SIGSTKFLT" | "SIGCHLD" | "SIGCONT" | "SIGSTOP" | "SIGTSTP" | "SIGTTIN" | "SIGTTOU" | "SIGURG" | "SIGXCPU" | "SIGXFSZ" | "SIGVTALRM" | "SIGPROF" | "SIGWINCH" | "SIGIO" | "SIGPWR" | "SIGSYS"`, true
	case "oci.Version":
		return `"1.0" | "1.1"`, true
	case "order.Order":
		return `"start-first" | "stop-first"`, true
	case "selinux.Label":
		return `"z" | "Z"`, true
	case "DependCondition", "depend.Condition":
		return `"service_started" | "service_healthy" | "service_completed_successfully"`, true
	case "VolumeKind", "vol.Kind":
		return `"bind" | "volume" | "tmpfs" | "npipe" | "cluster" | "image"`, true
	case "failact.Action":
		return `"continue" | "pause" | "rollback"`, true
	case "isolation.Mode":
		return `"default" | "process" | "hyperv"`, true
	case "cgroupns.Mode":
		return `"host" | "private"`, true
	case "endpoint.Mode":
		return `"vip" | "dnsrr"`, true
	case "propagate.Mode":
		return `"private" | "rprivate" | "shared" | "rshared" | "slave" | "rslave"`, true
	case "recursive.Mode":
		return `"enabled" | "disabled" | "writable" | "readonly"`, true
	case "progress.Mode":
		return `"auto" | "plain" | "tty" | "quiet" | "json"`, true
	case "scheme.Mode":
		return `"unix" | "npipe" | "tcp" | "http" | "https" | "ssh" | "fd"`, true
	case "WatchAction", "watch.Action":
		return `"sync" | "rebuild" | "sync+restart" | "sync+exec" | "restart"`, true
	case "WaitCondition", "wait.Condition":
		return `"not-running" | "next-exit" | "removed"`, true
	case "userns.Mode", "uts.Mode":
		return `"host"`, true
	case "PullMode", "pull.Mode":
		return "`every_${string}` | \"always\" | \"never\" | \"missing\" | \"if_not_present\" | \"build\" | \"refresh\" | \"daily\" | \"weekly\"", true
	case "netmode.Mode":
		return "`container:${string}` | `service:${string}` | \"bridge\" | \"host\" | \"none\" | \"default\"", true
	case "ipc.Mode":
		return "`container:${string}` | \"none\" | \"private\" | \"shareable\" | \"host\"", true
	case "pidns.Mode":
		return "`container:${string}` | \"host\"", true
	case "ocispec.Platform":
		return "ImagePlatform", true
	default:
		return "", false
	}
}
