//go:build windows

package terminal

import (
	"context"
	"io"
	"os"
)

// copyInput copies src to dst until EOF or ctx is canceled.
//
// A read from a Windows console cannot be interrupted. On cancel this returns
// and leaves the read running until the next byte arrives and the write to the
// closed connection fails.
func copyInput(ctx context.Context, dst io.Writer, src *os.File) error {
	errCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(dst, src)
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
