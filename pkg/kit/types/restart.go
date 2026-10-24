package types

import "github.com/docker/docker/api/types/container"

// RestartPolicy is [container.RestartPolicyMode].
type RestartPolicy = container.RestartPolicyMode

const (
	RestartPolicyNo            = container.RestartPolicyDisabled
	RestartPolicyOnFailure     = container.RestartPolicyOnFailure
	RestartPolicyAlways        = container.RestartPolicyAlways
	RestartPolicyUnlessStopped = container.RestartPolicyUnlessStopped
)
