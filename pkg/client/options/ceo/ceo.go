// package ceo provides options for the container exec.
package ceo

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
)

// SetContainerExecOption is a function that sets a parameter for the container exec.
type SetContainerExecOption func(*container.ExecOptions) error

// WithUser sets the user for the exec options.
func WithUser(user string) SetContainerExecOption {
	return func(o *container.ExecOptions) error {
		o.User = user
		return nil
	}
}

// WithPrivileged sets the privileged flag for the exec options.
func WithPrivileged() SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.Privileged = true
		return nil
	}
}

// WithTty sets the tty flag for the exec options.
func WithTty() SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.Tty = true
		return nil
	}
}

// WithConsoleSize sets the console size for the exec options.
func WithConsoleSize(width, height uint) SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.ConsoleSize = &[2]uint{width, height}
		return nil
	}
}

// WithAttachStdin sets the attach stdin flag for the exec options.
func WithAttachStdin() SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.AttachStdin = true
		return nil
	}
}

// WithAttachStderr sets the attach stderr flag for the exec options.
func WithAttachStderr() SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.AttachStderr = true
		return nil
	}
}

// WithAttachStdout sets the attach stdout flag for the exec options.
func WithAttachStdout() SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.AttachStdout = true
		return nil
	}
}

// WithDetach sets the detach flag for the exec options.
func WithDetach() SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.Detach = true
		return nil
	}
}

// WithDetachKeys sets the detach keys for the exec options.
func WithDetachKeys(detachKeys string) SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.DetachKeys = detachKeys
		return nil
	}
}

// WithEnv appends the environment variable to the exec options.
func WithEnv(key, value string) SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		if o.Env == nil {
			o.Env = []string{}
		}
		o.Env = append(o.Env, fmt.Sprintf("%s=%s", key, value))
		return nil
	}
}

// WithWorkingDir sets the working directory for the exec options.
func WithWorkingDir(workingDir string) SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		o.WorkingDir = workingDir
		return nil
	}
}

// WithCmd appends the command to the exec options.
func WithCmd(cmd ...string) SetContainerExecOption {

	return func(o *container.ExecOptions) error {
		if o.Cmd == nil {
			o.Cmd = []string{}
		}
		o.Cmd = append(o.Cmd, cmd...)
		return nil
	}
}
