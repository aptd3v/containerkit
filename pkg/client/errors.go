package client

import (
	cerrdefs "github.com/containerd/errdefs"
)

// IsErrNotFound reports whether err is a Docker Engine "not found" error
// (container, image, network, volume, and similar). It unwraps wrapped errors.
func IsErrNotFound(err error) bool {
	return cerrdefs.IsNotFound(err)
}
