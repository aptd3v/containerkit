// Package up provides options for the compose up command
package up

import (
	"io"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
)

type PullPolicy string

const (
	PullPolicyAlways  PullPolicy = "always"
	PullPolicyMissing PullPolicy = "missing"
	PullPolicyNever   PullPolicy = "never"
)

// WithDetach sets the detach option to true
//
// --detach Run containers in the background.
func WithDetach() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Detach = true
		return nil
	}
}

// WithAbortOnContainerExit sets the abort on container exit option to true
//
// --abort-on-container-exit Stops all containers if any container was stopped. Incompatible with --detach
func WithAbortOnContainerExit() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.AbortOnContainerExit = true
		return nil
	}
}

// WithAbortOnContainerFailure sets the abort on container failure option to true
//
// --abort-on-container-failure Stops all containers if any container exited with failure. Incompatible with --detach
func WithAbortOnContainerFailure() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.AbortOnContainerFailure = true
		return nil
	}
}

// WithAlwaysRecreateDeps sets the always recreate deps option to true
//
// --always-recreate-deps Recreate dependent containers. Incompatible with --no-recreate.
func WithAlwaysRecreateDeps() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.AlwaysRecreateDeps = true
		return nil
	}
}

// WithAttach sets the services to the attach option
//
// --attach Restricts attaching to the specified service. Incompatible with --attach-dependencies.
func WithAttach(service string) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Attach = &service
		return nil
	}
}

// WithAttachDependencies sets the attach dependencies option to true
//
// --attach-dependencies Automatically attach to log output of dependent services
func WithAttachDependencies() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.AttachDependencies = true
		return nil
	}
}

// WithBuild sets the build option to true
//
// --build		Build images before starting containers
func WithBuild() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Build = true
		return nil
	}
}

// WithExitCodeFrom sets the exit code from option
//
// --exit-code-from		Return the exit code of the selected service container. Implies --abort-on-container-exit
func WithExitCodeFrom(service string) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.ExitCodeFrom = &service
		return nil
	}
}

// WithForceRecreate sets the force recreate option to true
//
// --force-recreate		Recreate containers even if their configuration and image haven't changed
func WithForceRecreate() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.ForceRecreate = true
		return nil
	}
}

// WithMenu sets the menu option to true
//
// --menu		Enable interactive shortcuts when running attached. Incompatible with --detach. Can also be enable/disable by setting COMPOSE_MENU environment var.
func WithMenu() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Menu = true
		return nil
	}
}

// WithNoAttach sets a single service to not attach to the logs
//
// --no-attach Do not attach (stream logs) to the specified service
//
// note: there is a current limitation in the compose cli where you cannot use --no-attach aka (WithNoAttach) with multiple services
// use sc.WithNoAttach() instead on one or multiple services as needed.
func WithNoAttach(service string) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoAttach = &service
		return nil
	}
}

// WithNoBuild sets the no build option to true
//
// --no-build		Don't build an image, even if it's policy
func WithNoBuild() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoBuild = true
		return nil
	}
}

// WithNoColor sets the no color option to true
//
// --no-color		Produce monochrome output
func WithNoColor() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoColor = true
		return nil
	}
}

// WithNoDeps sets the no deps option to true
//
// --no-deps		Don't start linked services
func WithNoDeps() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoDeps = true
		return nil
	}
}

// WithNoLogPrefix sets the no log prefix option to true
//
// --no-log-prefix		Don't print prefix in logs
func WithNoLogPrefix() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoLogPrefix = true
		return nil
	}
}

// WithNoRecreate sets the no recreate option to true
//
// --no-recreate		If containers already exist, don't recreate them. Incompatible with --force-recreate.
func WithNoRecreate() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoRecreate = true
		return nil
	}
}

// WithPull sets the pull option
//
// --pull 	policy	Pull image before running ("always"|"missing"|"never")
func WithPull(policy PullPolicy) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		pull := string(policy)
		opt.Pull = &pull
		return nil
	}
}

// WithNoStart sets the no start option to true
//
// --no-start		Don't start the services after creating them
func WithNoStart() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.NoStart = true
		return nil
	}
}

// WithQuietPull sets the quiet pull option to true
//
// --quiet-pull		Pull without printing progress information
func WithQuietPull() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.QuietPull = true
		return nil
	}
}

// WithRemoveOrphans sets the remove orphans option to true
//
// --remove-orphans		Remove containers for services not defined in the Compose file
func WithRemoveOrphans() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.RemoveOrphans = true
		return nil
	}
}

// WithRenewAnonVolumes sets the renew anon volumes option to true
//
// --renew-anon-volumes		Recreate anonymous volumes instead of retrieving data from the previous containers
func WithRenewAnonVolumes() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.RenewAnonVolumes = true
		return nil
	}
}

// WithScale sets the scale option
//
// --scale		Scale SERVICE to NUM instances. Overrides the scale setting in the Compose file if present.
func WithScale(service string, num int) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		if opt.Scale == nil {
			opt.Scale = []compose.ComposeUpScale{}
		}
		opt.Scale = append(opt.Scale, compose.ComposeUpScale{Service: service, Num: num})
		return nil
	}
}

// WithTimeout sets the timeout option

// --timeout		Use this timeout in seconds for container shutdown when attached or when containers are already running
func WithTimeout(timeout int) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Timeout = &timeout
		return nil
	}
}

// WithTimestamps sets the timestamps option to true
//
// --timestamps		Show timestamps
func WithTimestamps() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Timestamps = true
		return nil
	}
}

// WithWait sets the wait option to true
//
// --wait		Wait for services to be running|healthy. Implies detached mode.
func WithWait() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Wait = true
		return nil
	}
}

// WithWaitTimeout sets the wait timeout option
//
// --wait-timeout		Maximum duration in seconds to wait for the project to be running|healthy
func WithWaitTimeout(timeout int) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.WaitTimeout = &timeout
		return nil
	}
}
func WithWatch() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Watch = true
		return nil
	}
}

// WithYes sets the yes option to true
//
// -y, --yes		Assume "yes" as answer to all prompts and run non-interactively
func WithYes() compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		opt.Yes = true
		return nil
	}
}

// WithWriter sets the writer for the compose up command stdout and stderr
//
// if writer is nil, it will use os.Stdout as a fallback
func WithWriter(writer io.Writer) compose.SetComposeUpOption {
	return func(opt *compose.ComposeUpOptions) error {
		if writer != nil {
			opt.Writer = os.Stdout
		}
		opt.Writer = writer
		return nil
	}
}
