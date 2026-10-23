package containerkit

import (
	"github.com/docker/docker/api/types/container"
)

// RestartPolicy is a Docker restart policy.
// The zero value leaves the policy unset.
type RestartPolicy struct{ name string }

var (
	RestartPolicyNo            = RestartPolicy{name: "no"}
	RestartPolicyOnFailure     = RestartPolicy{name: "on-failure"}
	RestartPolicyAlways        = RestartPolicy{name: "always"}
	RestartPolicyUnlessStopped = RestartPolicy{name: "unless-stopped"}
)

// String returns the Docker restart policy name. The zero value is empty.
func (p RestartPolicy) String() string { return p.name }

// RestartPolicy sets the container restart policy.
// maxRetryCount applies to RestartPolicyOnFailure.
func (c *Container) RestartPolicy(mode RestartPolicy, maxRetryCount int) *Container {
	var pm container.RestartPolicyMode
	switch mode.name {
	case RestartPolicyNo.name:
		pm = container.RestartPolicyDisabled
	case RestartPolicyOnFailure.name:
		pm = container.RestartPolicyOnFailure
	case RestartPolicyAlways.name:
		pm = container.RestartPolicyAlways
	case RestartPolicyUnlessStopped.name:
		pm = container.RestartPolicyUnlessStopped
	default:
		pm = container.RestartPolicyDisabled
	}
	c.Config.Host.RestartPolicy = container.RestartPolicy{
		Name:              pm,
		MaximumRetryCount: maxRetryCount,
	}
	return c
}

// RestartPolicyAlways restarts the container whenever it exits.
func (c *Container) RestartPolicyAlways() *Container {
	return c.RestartPolicy(RestartPolicyAlways, 0)
}

// RestartPolicyOnFailure restarts the container when it exits with an error.
func (c *Container) RestartPolicyOnFailure(maxRetryCount int) *Container {
	return c.RestartPolicy(RestartPolicyOnFailure, maxRetryCount)
}

// RestartPolicyUnlessStopped restarts the container unless it was stopped.
func (c *Container) RestartPolicyUnlessStopped() *Container {
	return c.RestartPolicy(RestartPolicyUnlessStopped, 0)
}

// RestartPolicyNever does not restart the container.
func (c *Container) RestartPolicyNever() *Container {
	return c.RestartPolicy(RestartPolicyNo, 0)
}
