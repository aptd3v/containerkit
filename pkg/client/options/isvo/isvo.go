// Package isvo provides options for image save.
package isvo

import (
	"github.com/aptd3v/go-contain/pkg/create"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type SetImageSaveOption func(*ImageSaveOptions) error

type ImageSaveOptions struct {
	Platforms []v1.Platform
	ImageIDs  []string
}

// WithPlatform appends the platform for the image save.
func WithPlatform(setters ...create.SetPlatformConfig) SetImageSaveOption {
	return func(o *ImageSaveOptions) error {
		pc := create.PlatformConfig{
			Platform: &v1.Platform{},
		}
		for _, setter := range setters {
			if err := setter(&pc); err != nil {
				return err
			}
		}
		if o.Platforms == nil {
			o.Platforms = make([]v1.Platform, 0)
		}
		o.Platforms = append(o.Platforms, *pc.Platform)
		return nil
	}
}

// WithImageID appends the image id for the image save.
func WithImageID(imageID string) SetImageSaveOption {
	return func(o *ImageSaveOptions) error {
		if o.ImageIDs == nil {
			o.ImageIDs = make([]string, 0)
		}
		o.ImageIDs = append(o.ImageIDs, imageID)
		return nil
	}
}
