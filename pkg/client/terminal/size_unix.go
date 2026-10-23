//go:build linux || darwin

package terminal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// watchResize reports window changes until ctx is canceled.
// A failed size read is skipped so a transient error does not end the session.
func (s *Session) watchResize(ctx context.Context, resize func(Size) error) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)
	defer signal.Stop(sigs)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sigs:
			if ctx.Err() != nil {
				return nil
			}
			sz, err := s.Size()
			if err != nil {
				continue
			}
			if err := resize(sz); err != nil {
				return err
			}
		}
	}
}
