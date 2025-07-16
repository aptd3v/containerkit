// Package projectsecret provides a set of functions to configure the secret for the project
package projectsecret

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
)

type SetProjectSecretConfig func(*types.SecretConfig) error

// WithDriverOptions sets the driver options for the secret
// parameters:
//   - key: the key for the driver option
//   - value: the value for the driver option
func WithDriverOptions(key, value string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		if opt.DriverOpts == nil {
			opt.DriverOpts = make(map[string]string)
		}
		opt.DriverOpts[key] = value
		return nil
	}
}

// WithDriver sets the driver for the secret
// parameters:
//   - driver: the driver for the secret
func WithDriver(driver string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.Driver = driver
		return nil
	}
}

// WithTemplateDriver sets the template driver for the secret
// parameters:
//   - templateDriver: the template driver for the secret
func WithTemplateDriver(templateDriver string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.TemplateDriver = templateDriver
		return nil
	}
}

// WithEnvironment sets the environment for the secret
// parameters:
//   - environment: the environment for the secret
func WithEnvironment(environment string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.Environment = environment
		return nil
	}
}

// WithContent sets the content for the secret
// parameters:
//   - content: the content for the secret
func WithContent(content string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.Content = content
		return nil
	}
}

// WithName sets the name for the secret
// parameters:
//   - name: the name for the secret
func WithName(name string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.Name = name
		return nil
	}
}

// WithExternal sets the external for the secret
// parameters:
//   - external: the external for the secret
func WithExternal() SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.External = true
		return nil
	}
}

// WithFile sets the file for the secret
// parameters:
//   - path: the path for the secret
func WithFile(path string) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		opt.File = path
		return nil
	}
}

// Fail is a function that returns an error
//
// note: this is useful for when you want to fail the project secret config
// and append the error to the service config error collection
func Fail(err error) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		return errdefs.NewServiceConfigError("secrets", err.Error())
	}
}

// Failf is a function that returns an error
//
// note: this is useful for when you want to fail the project secret config
// and append the error to the service config error collection
func Failf(stringFormat string, args ...any) SetProjectSecretConfig {
	return func(opt *types.SecretConfig) error {
		return errdefs.NewServiceConfigError("secrets", fmt.Sprintf(stringFormat, args...))
	}
}
