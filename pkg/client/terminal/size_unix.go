//go:build linux || darwin

package terminal

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// GetSize returns the current terminal size
func (s *Session) GetSize() (width, height int, err error) {
	return term.GetSize(int(s.stdin.Fd()))
}

// MonitorSize sends new terminal sizes over the channel when SIGWINCH is received.
func (s *Session) MonitorSize() chan MonitorSize {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)

	go func() {
		defer close(s.resizeCh)

		for range signals {
			width, height, err := s.GetSize()
			if err != nil {
				return
			}
			s.resizeCh <- MonitorSize{Width: uint(width), Height: uint(height)}
		}
	}()

	return s.resizeCh
}
