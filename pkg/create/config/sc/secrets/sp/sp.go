// Package sp provides a set of functions to configure the secret for the project
package sp

import "github.com/compose-spec/compose-go/types"

type SetSecretProjectConfig func(*types.SecretConfig) error

func WithFile(path string) SetSecretProjectConfig {
	return func(opt *types.SecretConfig) error {
		opt.File = path
		return nil
	}
}
