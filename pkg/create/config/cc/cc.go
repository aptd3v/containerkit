// Package cc provides the options for the container config.
package cc

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/docker/go-connections/nat"
)

// WithEnv appends an environment variable and its value to the container configuration
// Parameters:
//   - key: environment variable name
//   - value: environment variable value
func WithEnv(key string, value string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if config.Env == nil {
			config.Env = make([]string, 0)
		}
		if key == "" || value == "" {
			return create.NewContainerConfigError("env", fmt.Sprintf("invalid environment variable: %s", key+"="+value))
		}
		config.Env = append(config.Env, key+"="+value)
		return nil
	}
}

// WithExposedPort appends a port to be exposed from the container
// Parameter:
//   - port: port number or port range to be exposed in the container (e.g., "80-1000" or "80")
//   - protocol: protocol to be exposed from the container
func WithExposedPort(protocol string, port string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if config.ExposedPorts == nil {
			config.ExposedPorts = make(nat.PortSet)
		}
		p, err := nat.NewPort(protocol, port)
		if err != nil {
			return create.NewContainerConfigError("exposed_port", fmt.Sprintf("invalid exposed port: %s, %s", port, err))
		}
		config.ExposedPorts[p] = struct{}{}
		return nil
	}
}

// WithHostName sets the hostname of the container
func WithHostName(hostname string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if hostname == "" {
			return create.NewContainerConfigError("hostname", fmt.Sprintf("invalid hostname: %s", hostname))
		}
		config.Hostname = hostname
		return nil
	}
}

// WithDomainName sets the domain name of the container
func WithDomainName(domainname string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if domainname == "" {
			return create.NewContainerConfigError("domainname", fmt.Sprintf("invalid domain name: %s", domainname))
		}
		config.Domainname = domainname
		return nil
	}
}

// WithImage sets the image to use for the container
func WithImage(image string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if image == "" {
			return create.NewContainerConfigError("image", fmt.Sprintf("invalid image: %s", image))
		}
		config.Image = image
		return nil
	}
}
func WithImagef(stringFormat string, args ...interface{}) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		image := fmt.Sprintf(stringFormat, args...)
		if image == "" {
			return create.NewContainerConfigError("image", fmt.Sprintf("invalid image: %s", image))
		}
		config.Image = image
		return nil
	}
}

// WithCommand sets the command to be run in the container
// Parameters:
//   - cmd: command and its arguments
func WithCommand(cmd ...string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if len(cmd) == 0 {
			return create.NewContainerConfigError("command", "command is empty")
		}
		if config.Cmd == nil {
			config.Cmd = make([]string, 0)
		}
		config.Cmd = append(config.Cmd, cmd...)
		return nil
	}
}

// WithUser sets the user that commands are run as inside the container
func WithUser(user string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if user == "" {
			return create.NewContainerConfigError("user", fmt.Sprintf("invalid user: %s", user))
		}
		config.User = user
		return nil
	}
}

// WithAttachedStdin enables attaching to container's standard input
func WithAttachedStdin() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.AttachStdin = true
		return nil
	}
}

// WithAttachedStdout enables attaching to container's standard output
func WithAttachedStdout() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.AttachStdout = true
		return nil
	}
}

// WithAttachedStderr enables attaching to container's standard error
func WithAttachedStderr() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.AttachStderr = true
		return nil
	}
}

// WithTty allocates a pseudo-TTY for the container
func WithTty() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.Tty = true
		return nil
	}
}

// WithStdinOpen keeps STDIN open even if not attached
func WithStdinOpen() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.OpenStdin = true
		return nil
	}
}

// WithStdinOnce closes STDIN after the first attach
func WithStdinOnce() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.StdinOnce = true
		return nil
	}
}

// WithEscapedArgs indicates that command arguments are already escaped
func WithEscapedArgs() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.ArgsEscaped = true
		return nil
	}
}

// WithVolume appends a  short hand volume mount point to the container
// Parameter:
//   - volume: path where the volume should be mounted
//
// note: will not work within service config for compose file
// use hc.WithRWHostBindMount or other mount setter functions instead
func WithVolume(volume string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if volume == "" {
			return create.NewContainerConfigError("volume", fmt.Sprintf("invalid volume: '%s'", volume))
		}
		if config.Volumes == nil {
			config.Volumes = make(map[string]struct{})
		}
		config.Volumes[volume] = struct{}{}
		return nil
	}
}

// WithWorkingDir sets the working directory for commands to run in
func WithWorkingDir(dir string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if dir == "" {
			return create.NewContainerConfigError("working_dir", fmt.Sprintf("invalid working directory: '%s'", dir))
		}
		config.WorkingDir = dir
		return nil
	}
}

// WithDisabledNetwork disables networking for the container
func WithDisabledNetwork() create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		config.NetworkDisabled = true
		return nil
	}
}

// WithOnBuild appends ONBUILD metadata that will trigger when the image is used as a base image
func WithOnBuild(args ...string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if len(args) == 0 {
			return create.NewContainerConfigError("onbuild", "onbuild args are empty")
		}
		if config.OnBuild == nil {
			config.OnBuild = make([]string, 0)
		}
		config.OnBuild = append(config.OnBuild, args...)
		return nil
	}
}

// WithLabel appends a label to the container
// Parameters:
//   - label: label key
//   - value: label value
func WithLabel(label, value string) create.SetContainerConfig {

	return func(config *create.ContainerConfig) error {
		if label == "" || value == "" {
			return create.NewContainerConfigError("label", fmt.Sprintf("empty label: %s", label+"="+value))
		}
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		}
		config.Labels[label] = value
		return nil
	}
}

// WithStopSignal sets the signal that will be used to stop the container
func WithStopSignal(signal string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if signal == "" {
			return create.NewContainerConfigError("stop_signal", "empty stop signal")
		}
		config.StopSignal = signal
		return nil
	}
}

// WithEntrypoint sets the entrypoint to be run within the container
func WithEntrypoint(entrypoint ...string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if len(entrypoint) == 0 {
			return create.NewContainerConfigError("entrypoint", "entrypoint is empty")
		}
		if config.Entrypoint == nil {
			config.Entrypoint = make([]string, 0)
		}
		config.Entrypoint = append(config.Entrypoint, entrypoint...)
		return nil
	}
}

// WithShell sets the shell for shell-form of RUN, CMD, ENTRYPOINT
func WithShell(shell ...string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if len(shell) == 0 {
			return create.NewContainerConfigError("shell", "shell is empty")
		}
		if config.Shell == nil {
			config.Shell = make([]string, 0)
		}
		config.Shell = append(config.Shell, shell...)
		return nil
	}
}

// WithStopTimeout sets the timeout (in seconds) to stop the container
//
// note: is also stop_grace_period in compose
func WithStopTimeout(timeout int) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if timeout <= 0 {
			return create.NewContainerConfigError("stop_timeout", "invalid stop timeout")
		}
		config.StopTimeout = &timeout
		return nil
	}
}

// WithMacAddress sets the MAC address for the container
// Parameter:
//   - macAddress: MAC address to be used for the container
//
// Deprecated: this function is deprecated since docker API v1.44. Use nc.WithMacAddress(string) instead.
func WithMacAddress(macAddress string) create.SetContainerConfig {
	return func(config *create.ContainerConfig) error {
		if macAddress == "" {
			return create.NewContainerConfigError("mac_address", "empty mac address")
		}
		config.MacAddress = macAddress
		return nil
	}
}
