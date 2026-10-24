package kit

import (
	"fmt"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/aptd3v/containerkit/pkg/kit/types"
	"github.com/docker/go-connections/nat"
)

// Env appends an environment variable and its value to the container configuration
// Parameters:
//   - key: environment variable name
//   - value: environment variable value
func (c *Container) Env(key string, value string) *Container {
	if c.Config.Container.Env == nil {
		c.Config.Container.Env = make([]string, 0)
	}
	if key == "" || value == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("env", fmt.Sprintf("invalid environment variable: %s", key+"="+value)))
		return c
	}
	c.Config.Container.Env = append(c.Config.Container.Env, key+"="+value)
	return c
}

// EnvMap appends a map of environment variables to the container configuration
// Parameters:
//   - env: map of environment variables
func (c *Container) EnvMap(env map[string]string) *Container {
	if c.Config.Container.Env == nil {
		c.Config.Container.Env = make([]string, 0)
	}
	for key, value := range env {
		if key == "" || value == "" {
			c.Errors = append(c.Errors, errdefs.NewContainerConfigError("env", fmt.Sprintf("invalid environment variable: %s", key+"="+value)))
			return c
		}
		c.Config.Container.Env = append(c.Config.Container.Env, key+"="+value)
	}
	return c
}

// ExposedPort appends a port to be exposed from the container
// Parameter:
//   - port: port number or port range to be exposed in the container (e.g., "80-1000" or "80")
//   - protocol: protocol to be exposed from the container
func (c *Container) ExposedPort(protocol types.Protocol, port string) *Container {
	if c.Config.Container.ExposedPorts == nil {
		c.Config.Container.ExposedPorts = make(nat.PortSet)
	}
	p, err := nat.NewPort(string(protocol), port)
	if err != nil {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("exposed_port", fmt.Sprintf("invalid exposed port: %s, %s", port, err)))
		return c
	}
	c.Config.Container.ExposedPorts[p] = struct{}{}
	return c
}

// HostName sets the hostname of the container
func (c *Container) HostName(hostname string) *Container {
	if hostname == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("hostname", fmt.Sprintf("invalid hostname: %s", hostname)))
		return c
	}
	c.Config.Container.Hostname = hostname
	return c
}

// DomainName sets the domain name of the container
func (c *Container) DomainName(domainname string) *Container {
	if domainname == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("domainname", fmt.Sprintf("invalid domain name: %s", domainname)))
		return c
	}
	c.Config.Container.Domainname = domainname
	return c
}

// Image sets the image to use for the container
func (c *Container) Image(image string) *Container {
	if image == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("image", fmt.Sprintf("invalid image: %s", image)))
		return c
	}
	c.Config.Container.Image = image
	return c
}
func (c *Container) Imagef(stringFormat string, args ...any) *Container {

	image := fmt.Sprintf(stringFormat, args...)
	if image == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("image", fmt.Sprintf("invalid image: %s", image)))
		return c
	}
	c.Config.Container.Image = image
	return c
}

// Command appends the command to be run in the container
// Parameters:
//   - cmd: command and its arguments
func (c *Container) Command(cmd ...string) *Container {
	if len(cmd) == 0 {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("command", "command is empty"))
		return c
	}
	if c.Config.Container.Cmd == nil {
		c.Config.Container.Cmd = make([]string, 0, len(cmd))
	}
	c.Config.Container.Cmd = append(c.Config.Container.Cmd, cmd...)
	return c
}

// User sets the user that commands are run as inside the container
func (c *Container) User(user string) *Container {
	if user == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("user", fmt.Sprintf("invalid user: %s", user)))
		return c
	}
	c.Config.Container.User = user
	return c
}

func (c *Container) AttachedStdin() *Container {
	c.Config.Container.AttachStdin = true
	return c
}

func (c *Container) AttachedStdout() *Container {
	c.Config.Container.AttachStdout = true
	return c
}

func (c *Container) AttachedStderr() *Container {
	c.Config.Container.AttachStderr = true
	return c
}

func (c *Container) Tty() *Container {
	c.Config.Container.Tty = true
	return c
}

func (c *Container) StdinOpen() *Container {
	c.Config.Container.OpenStdin = true
	return c
}

func (c *Container) StdinOnce() *Container {
	c.Config.Container.StdinOnce = true
	return c
}

func (c *Container) EscapedArgs() *Container {
	c.Config.Container.ArgsEscaped = true
	return c
}

// Volume appends a  short hand volume mount point to the container
// Parameter:
//   - volume: path where the volume should be mounted
//
// note: Compose maps this to an anonymous volume at the container path.
// Named volumes use Mount or RWNamedVolumeMount.
func (c *Container) Volume(volume string) *Container {
	if volume == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("volume", fmt.Sprintf("invalid volume: '%s'", volume)))
		return c
	}
	if c.Config.Container.Volumes == nil {
		c.Config.Container.Volumes = make(map[string]struct{})
	}
	c.Config.Container.Volumes[volume] = struct{}{}
	return c
}

// WorkingDir sets the working directory for commands to run in
func (c *Container) WorkingDir(dir string) *Container {
	if dir == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("working_dir", fmt.Sprintf("invalid working directory: '%s'", dir)))
		return c
	}
	c.Config.Container.WorkingDir = dir
	return c
}

func (c *Container) DisabledNetwork() *Container {
	c.Config.Container.NetworkDisabled = true
	return c
}

// OnBuild appends ONBUILD metadata that will trigger when the image is used as a base image
func (c *Container) OnBuild(args ...string) *Container {
	if len(args) == 0 {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("onbuild", "onbuild args are empty"))
		return c
	}
	if c.Config.Container.OnBuild == nil {
		c.Config.Container.OnBuild = make([]string, 0, len(args))
	}
	c.Config.Container.OnBuild = append(c.Config.Container.OnBuild, args...)
	return c
}

// Label appends a label to the container
// Parameters:
//   - label: label key
//   - value: label value
func (c *Container) Label(label, value string) *Container {
	if label == "" || value == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("label", fmt.Sprintf("empty label: %s", label+"="+value)))
		return c
	}
	if c.Config.Container.Labels == nil {
		c.Config.Container.Labels = make(map[string]string)
	}
	c.Config.Container.Labels[label] = value
	return c
}

// StopSignal sets the signal that will be used to stop the container
func (c *Container) StopSignal(signal string) *Container {
	if signal == "" {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("stop_signal", "empty stop signal"))
		return c
	}
	c.Config.Container.StopSignal = signal
	return c
}

// Entrypoint sets the entrypoint to be run within the container
func (c *Container) Entrypoint(entrypoint ...string) *Container {
	if len(entrypoint) == 0 {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("entrypoint", "entrypoint is empty"))
		return c
	}
	if c.Config.Container.Entrypoint == nil {
		c.Config.Container.Entrypoint = make([]string, 0, len(entrypoint))
	}
	c.Config.Container.Entrypoint = append(c.Config.Container.Entrypoint, entrypoint...)
	return c
}

// Shell sets the shell for shell-form of RUN, CMD, ENTRYPOINT
func (c *Container) Shell(shell ...string) *Container {
	if len(shell) == 0 {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("shell", "shell is empty"))
		return c
	}
	if c.Config.Container.Shell == nil {
		c.Config.Container.Shell = make([]string, 0, len(shell))
	}
	c.Config.Container.Shell = append(c.Config.Container.Shell, shell...)
	return c
}

// StopTimeout sets the timeout (in seconds) to stop the container
//
// note: is also stop_grace_period in compose
func (c *Container) StopTimeout(timeout int) *Container {
	if timeout <= 0 {
		c.Errors = append(c.Errors, errdefs.NewContainerConfigError("stop_timeout", "invalid stop timeout"))
		return c
	}
	c.Config.Container.StopTimeout = &timeout
	return c
}
