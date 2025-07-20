package client

import (
	"io"

	"github.com/docker/docker/pkg/stdcopy"
)

// LogCopier provides methods to copy Docker container logs
type LogCopier struct {
	stdout io.Writer
	stderr io.Writer
}

// NewLogCopier creates a new LogCopier instance for multiplexed streams
// If stderr is nil, it will use stdout for both streams
func NewLogCopier(stdout io.Writer, stderr io.Writer) *LogCopier {
	if stderr == nil {
		stderr = stdout
	}
	return &LogCopier{
		stdout: stdout,
		stderr: stderr,
	}
}

// Copy copies the container log stream to the configured writers
// It handles Docker's multiplexed output format where stdout and stderr are combined with headers
func (lc *LogCopier) Copy(src io.Reader) (written int64, err error) {
	return stdcopy.StdCopy(lc.stdout, lc.stderr, src)
}

// CopyWithPrefix copies the container log stream and adds prefixes to stdout and stderr
// This is useful when you want to distinguish between the two streams in the output
func (lc *LogCopier) CopyWithPrefix(src io.Reader, stdoutPrefix, stderrPrefix string) (written int64, err error) {
	stdout := &prefixWriter{writer: lc.stdout, prefix: stdoutPrefix}
	stderr := &prefixWriter{writer: lc.stderr, prefix: stderrPrefix}
	return stdcopy.StdCopy(stdout, stderr, src)
}

type prefixWriter struct {
	writer io.Writer
	prefix string
}

func (w *prefixWriter) Write(p []byte) (n int, err error) {
	// Add prefix to each line
	prefixedData := append([]byte(w.prefix), p...)
	_, err = w.writer.Write(prefixedData)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
