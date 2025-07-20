// Package terminal
package terminal

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Session represents an interactive terminal session
type Session struct {
	stdin    *os.File
	oldState *term.State
	hijacked io.ReadWriteCloser
	reader   io.Reader
	resizeCh chan [2]uint
}

// NewSession creates a new terminal session
func NewSession(stdin *os.File, hijacked io.ReadWriteCloser, reader io.Reader) (*Session, error) {
	oldState, err := term.MakeRaw(int(stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}

	return &Session{
		stdin:    stdin,
		oldState: oldState,
		hijacked: hijacked,
		reader:   reader,
		resizeCh: make(chan [2]uint),
	}, nil
}

// Start begins the interactive session with bidirectional I/O
func (s *Session) Start() error {
	defer s.Close()

	// Set up error channel
	errCh := make(chan error, 1)

	// Copy container output to stdout
	go func() {
		_, err := io.Copy(os.Stdout, s.reader)
		errCh <- err
	}()

	// Copy stdin to container
	go func() {
		_, err := io.Copy(s.hijacked, s.stdin)
		errCh <- err
	}()

	// Wait for an error from either goroutine
	if err := <-errCh; err != nil {
		return fmt.Errorf("error during I/O: %w", err)
	}

	return nil
}

// Close restores the terminal state and cleans up resources
func (s *Session) Close() error {
	if s.oldState != nil {
		if err := term.Restore(int(s.stdin.Fd()), s.oldState); err != nil {
			return fmt.Errorf("failed to restore terminal state: %w", err)
		}
	}
	return nil
}

// GetSize returns the current terminal size
func (s *Session) GetSize() (width, height int, err error) {
	return term.GetSize(int(s.stdin.Fd()))
}

// MonitorSize starts monitoring terminal size changes
func (s *Session) MonitorSize() chan [2]uint {
	go func() {
		defer close(s.resizeCh)
		for {
			width, height, err := s.GetSize()
			if err != nil {
				return
			}
			s.resizeCh <- [2]uint{uint(height), uint(width)}
		}
	}()
	return s.resizeCh
}
