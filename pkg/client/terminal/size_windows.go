//go:build windows

package terminal

import (
	"time"

	"golang.org/x/term"
)

// GetSize returns the current terminal size
func (s *Session) GetSize() (width, height int, err error) {
	return term.GetSize(int(s.stdin.Fd()))
}

// MonitorSize polls for terminal size changes on Windows
func (s *Session) MonitorSize() chan MonitorSize {
	go func() {
		defer close(s.resizeCh)

		oldW, oldH, err := s.GetSize()
		if err != nil {
			return
		}

		for {
			time.Sleep(200 * time.Millisecond)

			newW, newH, err := s.GetSize()
			if err != nil {
				return
			}

			if newW != oldW || newH != oldH {
				s.resizeCh <- MonitorSize{Width: uint(newW), Height: uint(newH)}
				oldW, oldH = newW, newH
			}
		}
	}()

	return s.resizeCh
}
