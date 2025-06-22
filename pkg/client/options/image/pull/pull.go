// package pull provides options for the image pull.
package pull

import (
	"context"
	"runtime"

	"github.com/aptd3v/go-contain/pkg/client/auth"
	"github.com/docker/docker/api/types/image"
)

// SetImagePullOption is a function that sets a parameter for the image pull.
type SetImagePullOption func(*image.PullOptions) error

// WithRegistryAuth sets the registry auth for the image pull.
func WithRegistryAuth(creds auth.Auth) SetImagePullOption {
	return func(o *image.PullOptions) error {
		auth, err := auth.AuthToBase64(creds)
		if err != nil {
			return err
		}
		o.RegistryAuth = auth
		return nil
	}
}

// WithPullAllPlatforms sets the pull all platforms associated with the image
// (e.g., linux/amd64, linux/arm64, etc.).
func WithPullAllPlatforms() SetImagePullOption {
	return func(o *image.PullOptions) error {
		o.All = true
		return nil
	}
}

// WithCurrentPlatform sets the platform for the image pull to the current platform. (GOARCH)
func WithCurrentPlatform() SetImagePullOption {
	return func(o *image.PullOptions) error {
		o.Platform = runtime.GOARCH
		return nil
	}
}

// WithPrivledgeFn sets the privilege function for the image pull.
func WithPrivledgeFunc(authFn func(ctx context.Context) (string, error)) SetImagePullOption {
	return func(o *image.PullOptions) error {
		o.PrivilegeFunc = authFn
		return nil
	}
}

// WithPlatform sets the platform for the image pull.
func WithPlatform(platform string) SetImagePullOption {
	return func(o *image.PullOptions) error {
		o.Platform = platform
		return nil
	}
}
