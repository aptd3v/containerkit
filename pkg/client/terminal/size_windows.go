//go:build windows

package terminal

import (
	"context"
	"time"
)

// watchResize polls for window changes until ctx is canceled.
// A failed size read is skipped so a transient error does not end the session.
func (s *Session) watchResize(ctx context.Context, resize func(Size) error) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	prev, err := s.Size()
	if err != nil {
		prev = Size{}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			sz, err := s.Size()
			if err != nil {
				continue
			}
			if sz == prev {
				continue
			}
			prev = sz
			if err := resize(sz); err != nil {
				return err
			}
		}
	}
}
