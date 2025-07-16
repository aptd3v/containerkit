// Package secretservice provides a set of functions to configure the secrets for the service
package secretservice

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

type SetSecretServiceConfig func(*types.ServiceSecretConfig) error

// WithSource sets the source for the secret
// parameters:
//   - source: the source for the secret
func WithSource(source string) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		opt.Source = source
		return nil
	}
}

// WithTarget sets the target for the secret
// parameters:
//   - target: the target for the secret
func WithTarget(target string) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		opt.Target = target
		return nil
	}
}

// WithUID sets the UID for the secret
// parameters:
//   - uid: the UID for the secret
func WithUID(uid string) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		opt.UID = uid
		return nil
	}
}

// WithGID sets the GID for the secret
// parameters:
//   - gid: the GID for the secret
func WithGID(gid string) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		opt.GID = gid
		return nil
	}
}

// WithMode sets the mode for the secret
// parameters:
//   - mode: the mode for the secret
func WithMode(mode int64) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		mode := types.FileMode(mode)
		opt.Mode = &mode
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the secret service config
// and append the error to the service config error collection
func Fail(err error) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		return errdefs.NewServiceConfigError("secrets", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the secret service config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		return errdefs.NewServiceConfigError("secrets", fmt.Sprintf(stringFormat, args...))
	}
}
