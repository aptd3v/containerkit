package kit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aptd3v/containerkit/pkg/terminal"
)

// AttachTerminal hijacks a running container and returns a [terminal.Session]
// bound to in and out. Call [terminal.Session.Start] to copy I/O; pass a
// resize func that calls [ResizeContainer] with [ConsoleSize] when in is a
// real terminal. With no attach options, Stream, AttachStdin, AttachStdout,
// and AttachStderr are enabled.
func AttachTerminal(ctx context.Context, api Engine, id string, in *os.File, out io.Writer, opts ...EngineAttachOption) (*terminal.Session, error) {
	if len(opts) == 0 {
		opts = []EngineAttachOption{Stream(true), AttachStdin(true), AttachStdout(true), AttachStderr(true)}
	}
	res, err := AttachContainer(ctx, api, id, opts...)
	if err != nil {
		return nil, err
	}
	sess, err := terminal.NewSession(in, out, res.Conn, res.Reader)
	if err != nil {
		res.Close()
		return nil, cfgErr("engine", id, "AttachTerminal", err)
	}
	return sess, nil
}

// ExecTerminal creates an exec process, hijacks it, and returns a
// [terminal.Session] bound to in and out. Tty, AttachStdin, AttachStdout, and
// AttachStderr default on; pass [Command] for the remote argv. Call
// [terminal.Session.Start] with a resize func that calls [ResizeExec] when in
// is a real terminal.
func ExecTerminal(ctx context.Context, api Engine, containerID string, in *os.File, out io.Writer, opts ...EngineExecCreateOption) (*terminal.Session, error) {
	opts = append([]EngineExecCreateOption{
		Tty(true),
		AttachStdin(true),
		AttachStdout(true),
		AttachStderr(true),
		ConsoleSize(24, 80),
	}, opts...)
	created, err := CreateExec(ctx, api, containerID, opts...)
	if err != nil {
		return nil, err
	}
	res, err := AttachExec(ctx, api, created.ID, Tty(true), ConsoleSize(24, 80))
	if err != nil {
		return nil, err
	}
	sess, err := terminal.NewSession(in, out, res.Conn, res.Reader)
	if err != nil {
		res.Close()
		return nil, cfgErr("engine", created.ID, "ExecTerminal", err)
	}
	return sess, nil
}

// StartExecTerminal creates an exec, hijacks it, and copies in and out until the
// remote process exits. When in is a terminal, the session is raw and window
// changes call ResizeExecTerminal. A pipe skips raw mode and resize, so a test
// can drive the same path without a keyboard.
func StartExecTerminal(ctx context.Context, api Engine, containerID string, in *os.File, out io.Writer, opts ...EngineExecCreateOption) error {
	opts = append([]EngineExecCreateOption{
		Tty(true),
		AttachStdin(true),
		AttachStdout(true),
		AttachStderr(true),
		ConsoleSize(24, 80),
	}, opts...)
	created, err := CreateExec(ctx, api, containerID, opts...)
	if err != nil {
		return err
	}
	res, err := AttachExec(ctx, api, created.ID, Tty(true), ConsoleSize(24, 80))
	if err != nil {
		return err
	}
	sess, err := terminal.NewSession(in, out, res.Conn, res.Reader)
	if err != nil {
		res.Close()
		return cfgErr("engine", created.ID, "StartExecTerminal", err)
	}
	defer sess.Close()
	var resize func(terminal.Size) error
	if _, err := sess.Size(); err == nil {
		resize = func(sz terminal.Size) error {
			return ResizeExecTerminal(ctx, api, created.ID, sz)
		}
	}
	err = sess.Start(ctx, resize)
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, terminal.ErrClosed) {
		return err
	}
	return nil
}

// ResizeTerminal reports a window size to a running container TTY.
func ResizeTerminal(ctx context.Context, api Engine, id string, sz terminal.Size) error {
	_, err := ResizeContainer(ctx, api, id, ConsoleSize(sz.Height, sz.Width))
	if err != nil {
		return fmt.Errorf("resize terminal: %w", err)
	}
	return nil
}

// ResizeExecTerminal reports a window size to a running exec TTY.
func ResizeExecTerminal(ctx context.Context, api Engine, execID string, sz terminal.Size) error {
	_, err := ResizeExec(ctx, api, execID, ConsoleSize(sz.Height, sz.Width))
	if err != nil {
		return fmt.Errorf("resize exec terminal: %w", err)
	}
	return nil
}
