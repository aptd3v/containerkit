// Package terminal attaches a local terminal to a hijacked container stream.
//
// A Session puts stdin into raw mode when it is a terminal, copies bytes in
// both directions, and restores the terminal when the remote side finishes,
// the context is canceled, or Close is called. Window changes are reported
// through a callback so the caller can invoke the engine resize API.
package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"golang.org/x/term"
)

// ErrClosed is returned when Close is called, including when Close stops a
// running Start.
var ErrClosed = errors.New("terminal: session closed")

// Size is the terminal window, in characters.
type Size struct {
	Width  uint
	Height uint
}

// Session is an interactive terminal bound to a hijacked connection.
//
// Close is idempotent. It is safe to call if Start never runs, and it is safe
// to call again after Start returns. Close during Start stops the copies and
// makes Start return ErrClosed.
type Session struct {
	in     *os.File
	out    io.Writer
	reader io.Reader

	mu       sync.Mutex
	conn     io.WriteCloser
	oldState *term.State
	stop     context.CancelCauseFunc
	started  bool
	closed   bool
}

// NewSession binds in and out to a hijacked stream.
//
// conn is the write side of the stream. reader is the read side. For a Docker
// hijack those are the connection and its buffered reader, which must be
// drained separately. NewSession does not change terminal settings; Start does
// that when in is a terminal.
func NewSession(in *os.File, out io.Writer, conn io.WriteCloser, reader io.Reader) (*Session, error) {
	if in == nil {
		return nil, errors.New("terminal: nil stdin")
	}
	if out == nil {
		return nil, errors.New("terminal: nil stdout")
	}
	if conn == nil {
		return nil, errors.New("terminal: nil connection")
	}
	if reader == nil {
		return nil, errors.New("terminal: nil reader")
	}
	return &Session{in: in, out: out, conn: conn, reader: reader}, nil
}

// Size returns the current window size.
func (s *Session) Size() (Size, error) {
	if s == nil || s.in == nil || !term.IsTerminal(int(s.in.Fd())) {
		return Size{}, errors.New("terminal: stdin is not a terminal")
	}
	width, height, err := term.GetSize(int(s.in.Fd()))
	if err != nil {
		return Size{}, fmt.Errorf("terminal: size: %w", err)
	}
	if width < 0 || height < 0 {
		return Size{}, errors.New("terminal: negative size")
	}
	return Size{Width: uint(width), Height: uint(height)}, nil
}

// Start copies in to conn and reader to out.
//
// When in reaches EOF, Start half-closes conn so the remote sees EOF and
// keeps copying output. When reader returns, Start closes conn and returns.
// Canceling ctx closes conn and stops both copies.
//
// resize is called with the current size before copying starts, then again
// when the window changes. A nil resize skips watching. resize runs
// synchronously for the initial size and again from a goroutine afterward;
// it should return when ctx is canceled.
//
// Start is called once. It enters raw mode when in is a terminal and restores
// the previous state before returning.
func (s *Session) Start(ctx context.Context, resize func(Size) error) (err error) {
	if s == nil {
		return errors.New("terminal: nil session")
	}
	if ctx == nil {
		return errors.New("terminal: nil context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrClosed
	}
	if s.started {
		s.mu.Unlock()
		return errors.New("terminal: session already started")
	}
	s.started = true
	s.mu.Unlock()

	defer func() {
		if cerr := s.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	// Resize before raw mode so a stuck engine call still leaves Ctrl+C working.
	if resize != nil {
		if err := s.resize(resize); err != nil {
			return fmt.Errorf("terminal: resize: %w", err)
		}
	}
	if err := s.makeRaw(); err != nil {
		return err
	}

	parent := ctx
	ctx, stop := context.WithCancelCause(parent)
	s.mu.Lock()
	s.stop = stop
	conn := s.conn
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.stop = nil
		s.mu.Unlock()
		stop(nil)
	}()

	var (
		wg     sync.WaitGroup
		errMu  sync.Mutex
		runErr error
	)
	setErr := func(err error) {
		if err == nil || errors.Is(err, context.Canceled) {
			return
		}
		errMu.Lock()
		if runErr == nil {
			runErr = err
		}
		errMu.Unlock()
		s.requestStop(nil)
	}

	wg.Go(func() {
		<-ctx.Done()
		s.closeConn()
	})
	wg.Go(func() {
		_, err := io.Copy(s.out, s.reader)
		if err != nil && !isClosed(err) {
			setErr(fmt.Errorf("terminal: copy output: %w", err))
		}
		s.requestStop(nil)
	})
	wg.Go(func() {
		err := copyInput(ctx, conn, s.in)
		if err == nil || errors.Is(err, io.EOF) {
			if ctx.Err() != nil {
				return
			}
			if herr := s.halfClose(); herr != nil && !isClosed(herr) {
				setErr(fmt.Errorf("terminal: close stdin: %w", herr))
			}
			return
		}
		if ctx.Err() != nil || isClosed(err) {
			return
		}
		setErr(fmt.Errorf("terminal: copy input: %w", err))
	})
	if resize != nil {
		wg.Go(func() {
			err := s.watchResize(ctx, resize)
			if err != nil && ctx.Err() == nil && !errors.Is(err, context.Canceled) {
				setErr(fmt.Errorf("terminal: resize: %w", err))
			}
		})
	}

	wg.Wait()

	errMu.Lock()
	err = runErr
	errMu.Unlock()
	if err != nil {
		return err
	}
	if err := parent.Err(); err != nil {
		return err
	}
	if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, context.Canceled) {
		return cause
	}
	return nil
}

// Close restores the terminal state and closes the hijacked connection.
// A second call is a no-op. Close during Start makes Start return ErrClosed.
func (s *Session) Close() error {
	if s == nil {
		return nil
	}
	s.requestStop(ErrClosed)

	s.mu.Lock()
	state := s.oldState
	in := s.in
	conn := s.conn
	s.oldState = nil
	s.conn = nil
	s.closed = true
	s.mu.Unlock()

	var errs []error
	if state != nil && in != nil {
		if err := term.Restore(int(in.Fd()), state); err != nil {
			errs = append(errs, fmt.Errorf("terminal: restore: %w", err))
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("terminal: close connection: %w", err))
		}
	}
	return errors.Join(errs...)
}

func (s *Session) makeRaw() error {
	if !term.IsTerminal(int(s.in.Fd())) {
		return nil
	}
	state, err := term.MakeRaw(int(s.in.Fd()))
	if err != nil {
		return fmt.Errorf("terminal: set raw mode: %w", err)
	}
	s.mu.Lock()
	s.oldState = state
	s.mu.Unlock()
	return nil
}

func (s *Session) resize(resize func(Size) error) error {
	sz, err := s.Size()
	if err != nil {
		return err
	}
	return resize(sz)
}

func (s *Session) requestStop(err error) {
	s.mu.Lock()
	stop := s.stop
	s.mu.Unlock()
	if stop != nil {
		stop(err)
	}
}

func (s *Session) closeConn() {
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (s *Session) halfClose() error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return nil
	}
	cw, ok := conn.(interface{ CloseWrite() error })
	if !ok {
		return nil
	}
	return cw.CloseWrite()
}

func isClosed(err error) bool {
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, os.ErrClosed)
}
