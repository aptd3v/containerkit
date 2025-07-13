package hc

import (
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/docker/api/types/container"
)

type RestartPolicy string

const (
	RestartPolicyNo            RestartPolicy = "no"
	RestartPolicyOnFailure     RestartPolicy = "on-failure"
	RestartPolicyAlways        RestartPolicy = "always"
	RestartPolicyUnlessStopped RestartPolicy = "unless-stopped"
)

// WithRestartPolicy adds a restart policy to the host configuration.
// parameters:
//   - mode: the restart policy to use
//   - maxRetryCount: the maximum number of retries before giving up
func WithRestartPolicy(mode RestartPolicy, maxRetryCount int) create.SetHostConfig {
	var pm container.RestartPolicyMode
	switch mode {
	case RestartPolicyNo:
		pm = container.RestartPolicyDisabled
	case RestartPolicyOnFailure:
		pm = container.RestartPolicyOnFailure
	case RestartPolicyAlways:
		pm = container.RestartPolicyAlways
	case RestartPolicyUnlessStopped:
		pm = container.RestartPolicyUnlessStopped
	default:
		pm = container.RestartPolicyDisabled
	}
	return func(opt *container.HostConfig) error {
		opt.RestartPolicy = container.RestartPolicy{
			Name:              pm,
			MaximumRetryCount: maxRetryCount,
		}
		return nil
	}
}

// WithRestartPolicyAlways sets the restart policy to always
func WithRestartPolicyAlways() create.SetHostConfig {
	return WithRestartPolicy(RestartPolicyAlways, 0)
}

// WithRestartPolicyOnFailure sets the restart policy to on-failure
// parameters:
//   - maxRetryCount: the maximum number of retries before giving up
func WithRestartPolicyOnFailure(maxRetryCount int) create.SetHostConfig {
	return WithRestartPolicy(RestartPolicyOnFailure, maxRetryCount)
}

// WithRestartPolicyUnlessStopped sets the restart policy to unless-stopped
func WithRestartPolicyUnlessStopped() create.SetHostConfig {
	return WithRestartPolicy(RestartPolicyUnlessStopped, 0)
}

// WithRestartPolicyNever sets the restart policy to no
func WithRestartPolicyNever() create.SetHostConfig {
	return WithRestartPolicy(RestartPolicyNo, 0)
}
