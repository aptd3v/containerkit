// Package ss provides a set of functions to configure the secrets for the service
package ss

import "github.com/compose-spec/compose-go/types"

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
func WithMode(mode uint32) SetSecretServiceConfig {
	return func(opt *types.ServiceSecretConfig) error {
		opt.Mode = &mode
		return nil
	}
}
