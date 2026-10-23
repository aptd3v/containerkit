//go:build linux || darwin

package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// copyInput copies src to dst until EOF or ctx is canceled.
//
// The read is polled so cancellation can interrupt a blocking stdin without
// setting the descriptor non-blocking for the rest of the process.
func copyInput(ctx context.Context, dst io.Writer, src *os.File) error {
	wakeR, wakeW, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("terminal: stdin wake pipe: %w", err)
	}
	defer wakeR.Close()
	go func() {
		<-ctx.Done()
		wakeW.Close()
	}()

	buf := make([]byte, 32*1024)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ready, err := waitReadable(int(src.Fd()), int(wakeR.Fd()))
		if err != nil {
			return err
		}
		if !ready {
			if err := ctx.Err(); err != nil {
				return err
			}
			return errors.New("terminal: stdin wait interrupted")
		}
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err != nil {
			return err
		}
	}
}

// waitReadable reports whether fd has data. A readable wake descriptor with
// no data on fd reports ready=false so the caller can observe ctx.
func waitReadable(fd, wake int) (bool, error) {
	for {
		fds := []unix.PollFd{
			{Fd: int32(fd), Events: unix.POLLIN},
			{Fd: int32(wake), Events: unix.POLLIN},
		}
		_, err := unix.Poll(fds, -1)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("terminal: poll stdin: %w", err)
		}
		if fds[0].Revents&(unix.POLLNVAL) != 0 {
			return false, errors.New("terminal: stdin is not pollable")
		}
		if fds[0].Revents&(unix.POLLIN|unix.POLLHUP|unix.POLLERR) != 0 {
			return true, nil
		}
		if fds[1].Revents&(unix.POLLIN|unix.POLLHUP|unix.POLLERR) != 0 {
			return false, nil
		}
	}
}
