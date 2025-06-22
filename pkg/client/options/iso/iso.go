// Package iso provides options for image search.
package iso

import (
	"context"

	"github.com/aptd3v/go-contain/pkg/client/auth"
	"github.com/docker/docker/api/types/registry"
)

// SetImageSearchOption is a function that sets the image search options.
type SetImageSearchOption func(*registry.SearchOptions) error

// WithRegistryAuth sets the registry auth for the image search.
func WithRegistryAuth(creds auth.Auth) SetImageSearchOption {
	return func(o *registry.SearchOptions) error {
		o.RegistryAuth = auth.AuthToBase64(creds)
		return nil
	}
}

// WithPrivledgeFn sets the privilege function for the image search.
func WithPrivledgeFn(fn func(ctx context.Context) (string, error)) SetImageSearchOption {
	return func(so *registry.SearchOptions) error {
		so.PrivilegeFunc = fn
		return nil
	}
}

// WithFilter adds a filter for the image search.
func WithFilter(key, value string) SetImageSearchOption {
	return func(so *registry.SearchOptions) error {
		so.Filters.Add(key, value)
		return nil
	}
}

// WithLimit sets the limit for the image search.
func WithLimit(limit int) SetImageSearchOption {
	return func(so *registry.SearchOptions) error {
		so.Limit = limit
		return nil
	}
}
