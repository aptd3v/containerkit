package kit

import (
	"io"

	"github.com/moby/moby/api/pkg/stdcopy"
)

// CopyMultiplexed splits a Docker attach or logs stream.
// Without a TTY, the Engine prefixes each frame with a stream id.
// Stdout frames are written to out and stderr frames to errw.
// out and errw may be the same writer. src is read until EOF.
// The count is the total number of payload bytes written.
func CopyMultiplexed(out, errw io.Writer, src io.Reader) (int64, error) {
	return stdcopy.StdCopy(out, errw, src)
}
