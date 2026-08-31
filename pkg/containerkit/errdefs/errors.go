package errdefs

import (
	"errors"
	"fmt"
)

var (
	ErrContainerConfig = errors.New("container config has errors")
	ErrHostConfig      = errors.New("host config has errors")
	ErrNetworkConfig   = errors.New("network config has errors")
	ErrPlatformConfig  = errors.New("platform config has errors")
	ErrServiceConfig   = errors.New("service config has errors")
	ErrProjectConfig   = errors.New("project config has errors")
	ErrValidation      = errors.New("container has errors")
)

type ContainerConfigError struct {
	Field   string
	Message string
}

func (e *ContainerConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ContainerConfigError) Unwrap() error {
	return ErrContainerConfig
}

func IsContainerConfigError(err error) bool {
	return errors.Is(err, ErrContainerConfig)
}
func NewContainerConfigError(field, message string) *ContainerConfigError {
	return &ContainerConfigError{
		Field:   field,
		Message: message,
	}
}

type ServiceConfigError struct {
	Field   string
	Message string
}

func (e *ServiceConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ServiceConfigError) Unwrap() error {
	return ErrServiceConfig
}

func IsServiceConfigError(err error) bool {
	return errors.Is(err, ErrServiceConfig)
}

func NewServiceConfigError(field, message string) *ServiceConfigError {
	return &ServiceConfigError{
		Field:   field,
		Message: message,
	}
}

type HostConfigError struct {
	Field   string
	Message string
}

func (e *HostConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *HostConfigError) Unwrap() error {
	return ErrHostConfig
}

func IsHostConfigError(err error) bool {
	return errors.Is(err, ErrHostConfig)
}

func NewHostConfigError(field, message string) *HostConfigError {
	return &HostConfigError{
		Field:   field,
		Message: message,
	}
}

type NetworkConfigError struct {
	Field   string
	Message string
}

func (e *NetworkConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *NetworkConfigError) Unwrap() error {
	return ErrNetworkConfig
}

func IsNetworkConfigError(err error) bool {
	return errors.Is(err, ErrNetworkConfig)
}

func NewNetworkConfigError(field, message string) *NetworkConfigError {
	return &NetworkConfigError{
		Field:   field,
		Message: message,
	}
}

type PlatformConfigError struct {
	Field   string
	Message string
}

func (e *PlatformConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *PlatformConfigError) Unwrap() error {
	return ErrPlatformConfig
}

func IsPlatformConfigError(err error) bool {
	return errors.Is(err, ErrPlatformConfig)
}

func NewPlatformConfigError(field, message string) *PlatformConfigError {
	return &PlatformConfigError{
		Field:   field,
		Message: message,
	}
}

type ProjectConfigError struct {
	Field   string
	Message string
}

func (e *ProjectConfigError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *ProjectConfigError) Unwrap() error {
	return ErrProjectConfig
}

func IsProjectConfigError(err error) bool {
	return errors.Is(err, ErrProjectConfig)
}

func NewProjectConfigError(field, message string) *ProjectConfigError {
	return &ProjectConfigError{
		Field:   field,
		Message: message,
	}
}
