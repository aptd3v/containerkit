// Package iio provides options for image import.
package iio

import (
	"io"
)

// ImageImport is the image import options.
type ImageImportOptions struct {
	Source     io.Reader
	SourceName string
	Tag        string
	Message    string
	Changes    []string
	Platform   string
}

// SetImageImportOption is a function that sets the image import options.
type SetImageImportOption func(*ImageImportOptions) error

// WithSource sets the source for the image import.
//
//	Source is the data to send to the server to create this image from. You must set SourceName to "-" to leverage this.
func WithSource(source io.Reader) SetImageImportOption {
	return func(o *ImageImportOptions) error {
		o.Source = source
		return nil
	}
}

// WithSourceName sets the source name for the image import.
//
// SourceName is the name of the image to pull. Set to "-" to leverage the Source attribute.
func WithSourceName(sourceName string) SetImageImportOption {
	return func(o *ImageImportOptions) error {
		o.SourceName = sourceName
		return nil
	}
}

// WithTag sets the tag for the image import.
//
// Deprecated: this attribute is deprecated.
func WithTag(tag string) SetImageImportOption {
	return func(o *ImageImportOptions) error {
		o.Tag = tag
		return nil
	}
}

// WithMessage sets the message to tag the image with.
func WithMessage(message string) SetImageImportOption {
	return func(o *ImageImportOptions) error {
		o.Message = message
		return nil
	}
}

// WithChanges appends the raw changes to apply to the image.
func WithChanges(changes ...string) SetImageImportOption {
	return func(o *ImageImportOptions) error {
		if o.Changes == nil {
			o.Changes = make([]string, 0)
		}
		o.Changes = append(o.Changes, changes...)
		return nil
	}
}

// WithPlatform sets the platform for the image import.
func WithPlatform(platform string) SetImageImportOption {
	return func(o *ImageImportOptions) error {
		o.Platform = platform
		return nil
	}
}
