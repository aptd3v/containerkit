// package ico provides options for the image create.
package ico

import (
	"github.com/aptd3v/go-contain/pkg/client/auth"
	"github.com/docker/docker/api/types/image"
)

// SetImageCreateOption is a function that sets a parameter for the image create.
type SetImageCreateOption func(*image.CreateOptions) error

// WithRegistryAuth sets the registry auth for the image create.
func WithRegistryAuth(creds auth.Auth) SetImageCreateOption {
	return func(o *image.CreateOptions) error {
		o.RegistryAuth = auth.AuthToBase64(creds)
		return nil
	}
}

// WithPlatform sets the platform of the image if it needs to be pulled from the registry
func WithPlatform(platform string) SetImageCreateOption {
	return func(o *image.CreateOptions) error {
		o.Platform = platform
		return nil
	}
}
