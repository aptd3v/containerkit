// Package signal is the POSIX signal vocabulary for StopSignal and Kill.
package signal

// Signal is a signal name as accepted by Docker (SIGTERM, SIGKILL, ...).
type Signal string

const (
	SIGHUP    Signal = "SIGHUP"
	SIGINT    Signal = "SIGINT"
	SIGQUIT   Signal = "SIGQUIT"
	SIGILL    Signal = "SIGILL"
	SIGTRAP   Signal = "SIGTRAP"
	SIGABRT   Signal = "SIGABRT"
	SIGBUS    Signal = "SIGBUS"
	SIGFPE    Signal = "SIGFPE"
	SIGKILL   Signal = "SIGKILL"
	SIGUSR1   Signal = "SIGUSR1"
	SIGSEGV   Signal = "SIGSEGV"
	SIGUSR2   Signal = "SIGUSR2"
	SIGPIPE   Signal = "SIGPIPE"
	SIGALRM   Signal = "SIGALRM"
	SIGTERM   Signal = "SIGTERM"
	SIGSTKFLT Signal = "SIGSTKFLT"
	SIGCHLD   Signal = "SIGCHLD"
	SIGCONT   Signal = "SIGCONT"
	SIGSTOP   Signal = "SIGSTOP"
	SIGTSTP   Signal = "SIGTSTP"
	SIGTTIN   Signal = "SIGTTIN"
	SIGTTOU   Signal = "SIGTTOU"
	SIGURG    Signal = "SIGURG"
	SIGXCPU   Signal = "SIGXCPU"
	SIGXFSZ   Signal = "SIGXFSZ"
	SIGVTALRM Signal = "SIGVTALRM"
	SIGPROF   Signal = "SIGPROF"
	SIGWINCH  Signal = "SIGWINCH"
	SIGIO     Signal = "SIGIO"
	SIGPWR    Signal = "SIGPWR"
	SIGSYS    Signal = "SIGSYS"
)

// All returns every known POSIX/Linux signal name Docker accepts.
func All() []Signal {
	return []Signal{
		SIGHUP, SIGINT, SIGQUIT, SIGILL, SIGTRAP, SIGABRT, SIGBUS, SIGFPE,
		SIGKILL, SIGUSR1, SIGSEGV, SIGUSR2, SIGPIPE, SIGALRM, SIGTERM,
		SIGSTKFLT, SIGCHLD, SIGCONT, SIGSTOP, SIGTSTP, SIGTTIN, SIGTTOU,
		SIGURG, SIGXCPU, SIGXFSZ, SIGVTALRM, SIGPROF, SIGWINCH, SIGIO,
		SIGPWR, SIGSYS,
	}
}
