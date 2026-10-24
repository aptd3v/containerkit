package kit

import (
	"github.com/aptd3v/containerkit/pkg/kit/types"
	"github.com/docker/docker/api/types/container"
)

// RestartPolicy sets the container restart policy.
// maxRetryCount applies to RestartPolicyOnFailure.
func (c *Container) RestartPolicy(mode types.RestartPolicy, maxRetryCount int) *Container {
	name := container.RestartPolicyMode(mode)
	if name == "" {
		name = container.RestartPolicyDisabled
	}
	c.Config.Host.RestartPolicy = container.RestartPolicy{
		Name:              name,
		MaximumRetryCount: maxRetryCount,
	}
	return c
}

// RestartPolicyAlways restarts the container whenever it exits.
func (c *Container) RestartPolicyAlways() *Container {
	return c.RestartPolicy(types.RestartPolicyAlways, 0)
}

// RestartPolicyOnFailure restarts the container when it exits with an error.
func (c *Container) RestartPolicyOnFailure(maxRetryCount int) *Container {
	return c.RestartPolicy(types.RestartPolicyOnFailure, maxRetryCount)
}

// RestartPolicyUnlessStopped restarts the container unless it was stopped.
func (c *Container) RestartPolicyUnlessStopped() *Container {
	return c.RestartPolicy(types.RestartPolicyUnlessStopped, 0)
}

// RestartPolicyNever does not restart the container.
func (c *Container) RestartPolicyNever() *Container {
	return c.RestartPolicy(types.RestartPolicyNo, 0)
}
