// Package projectsecret provides a set of functions to configure the secret for the project
package projectsecret

import "github.com/compose-spec/compose-go/v2/types"

type SetProjectSecretConfig func(*types.SecretConfig) error

func WithFile(path string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.File = path
		return nil
	}
}
