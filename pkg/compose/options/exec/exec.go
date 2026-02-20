// Package exec provides options for the compose exec command
package exec

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

// WithDetach runs the command in the background
func WithDetach() compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Detach = true
		return nil
	}
}

// WithEnv sets environment variables (key=value), can be used multiple times
func WithEnv(keyValue ...string) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Env = append(opt.Env, keyValue...)
		return nil
	}
}

// WithIndex sets the index of the container when the service has multiple replicas
func WithIndex(index int) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Index = &index
		return nil
	}
}

// WithNoTTY disables pseudo-TTY allocation (use for scripts or piping)
func WithNoTTY() compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.NoTTY = true
		return nil
	}
}

// WithPrivileged gives extended privileges to the process
func WithPrivileged() compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Privileged = true
		return nil
	}
}

// WithUser runs the command as this user (e.g. "root", "1000:1000")
func WithUser(user string) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.User = user
		return nil
	}
}

// WithWorkdir sets the working directory for the command
func WithWorkdir(dir string) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Workdir = dir
		return nil
	}
}

// WithWriter sets the writer for stdout/stderr
func WithWriter(writer io.Writer) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		if writer == nil {
			opt.Writer = os.Stdout
			return nil
		}
		opt.Writer = writer
		return nil
	}
}

// WithStdin sets the reader to forward to the container (e.g. os.Stdin for interactive)
func WithStdin(r io.Reader) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Stdin = r
		return nil
	}
}

// WithProfiles sets the profiles to activate
func WithProfiles(profiles ...string) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Profiles = profiles
		return nil
	}
}

// WithService sets the service name (required)
func WithService(service string) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Service = service
		return nil
	}
}

// WithCommand sets the command and arguments (required), e.g. WithCommand("sh", "-c", "echo hi")
func WithCommand(command ...string) compose.SetComposeExecOption {
	return func(opt *compose.ComposeExecOptions) error {
		opt.Command = append(opt.Command, command...)
		return nil
	}
}
