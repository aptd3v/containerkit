package compose

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

// ComposeUpOptions is the options for the compose up command
type ComposeUpOptions struct {
	Detach                  bool
	AbortOnContainerExit    bool
	AbortOnContainerFailure bool
	AlwaysRecreateDeps      bool
	Attach                  *string
	AttachDependencies      bool
	Build                   bool
	NoBuild                 bool
	ForceRecreate           bool
	Menu                    bool
	ExitCodeFrom            *string
	NoAttach                *string
	NoColor                 bool
	NoDeps                  bool
	NoLogPrefix             bool
	NoRecreate              bool
	Pull                    *string
	NoStart                 bool
	QuietPull               bool
	RemoveOrphans           bool
	RenewAnonVolumes        bool
	Scale                   []ComposeUpScale
	Timeout                 *int
	Timestamps              bool
	Wait                    bool
	WaitTimeout             *int
	Watch                   bool
	Yes                     bool

	Profiles []string
	Flags    []string
	Writer   io.Writer
	Errs     []error
}

// ComposeUpScale is the scale for the compose up command
type ComposeUpScale struct {
	Service string
	Num     int
}

// addErrorWhen adds an error to the error slice if the condition is true
func (opt *ComposeUpOptions) addErrorWhen(cond bool, flag string, msg string) {
	if cond {
		opt.Errs = append(opt.Errs, NewComposeFlagError(flag, msg))
	}
}

// GenerateFlags generates the flags for the compose up command
//
// It will return a slice of flags to append to the command Eg.
//
//	[]string{"up", "--detach", ...}
func (opt *ComposeUpOptions) GenerateFlags() (flags []string, err error) {
	opt.Errs = []error{}
	flags = []string{"up"}
	if opt.Detach {
		fail := opt.Attach != nil || opt.AttachDependencies || opt.Watch || opt.AbortOnContainerExit || opt.AbortOnContainerFailure
		msg := "WithAttach, WithAttachDependencies, WithWatch, WithAbortOnContainerExit and WithAbortOnContainerFailure cannot be used together"
		opt.addErrorWhen(fail, "--detach", msg)
		flags = append(flags, "--detach")

	}
	if opt.NoAttach != nil {
		flags = append(flags, "--no-attach", *opt.NoAttach)
	}
	if opt.AbortOnContainerExit {
		flags = append(flags, "--abort-on-container-exit")
	}
	if opt.AbortOnContainerFailure {
		flags = append(flags, "--abort-on-container-failure")
	}
	if opt.AlwaysRecreateDeps {
		flags = append(flags, "--always-recreate-deps")
	}
	if opt.Attach != nil {
		opt.addErrorWhen(opt.AttachDependencies, "--attach", "WithAttach and WithAttachDependencies cannot be used together")
		flags = append(flags, "--attach", *opt.Attach)
	}
	if opt.AttachDependencies {
		opt.addErrorWhen(opt.Attach != nil, "--attach-dependencies", "WithAttach and WithAttachDependencies cannot be used together")
		flags = append(flags, "--attach-dependencies")
	}
	if opt.Build {
		flags = append(flags, "--build")
	}
	if opt.NoBuild {
		flags = append(flags, "--no-build")
	}

	if opt.Menu {
		flags = append(flags, "--menu")
	}
	if opt.ExitCodeFrom != nil {
		flags = append(flags, "--exit-code-from", *opt.ExitCodeFrom)
	}
	if opt.ForceRecreate {
		flags = append(flags, "--force-recreate")
	}
	if opt.NoColor {
		flags = append(flags, "--no-color")
	}
	if opt.NoDeps {
		flags = append(flags, "--no-deps")
	}
	if opt.NoLogPrefix {
		flags = append(flags, "--no-log-prefix")
	}
	if opt.NoRecreate {
		flags = append(flags, "--no-recreate")
	}
	if opt.Pull != nil {
		flags = append(flags, "--pull", *opt.Pull)
	}
	if opt.NoStart {
		flags = append(flags, "--no-start")
	}
	if opt.QuietPull {
		flags = append(flags, "--quiet-pull")
	}
	if opt.RemoveOrphans {
		flags = append(flags, "--remove-orphans")
	}
	if opt.RenewAnonVolumes {
		flags = append(flags, "--renew-anon-volumes")
	}
	if len(opt.Scale) > 0 {
		for _, scale := range opt.Scale {
			opt.addErrorWhen(scale.Service == "" || scale.Num < 0, "--scale", "Service name required and scale must be non-negative")
			flags = append(flags, "--scale", fmt.Sprintf("%s=%d", scale.Service, scale.Num))
		}
	}
	if opt.Timeout != nil {
		flags = append(flags, "--timeout", strconv.Itoa(*opt.Timeout))
	}
	if opt.Timestamps {
		flags = append(flags, "--timestamps")
	}
	if opt.Wait {
		fail := opt.Watch || opt.Attach != nil || opt.AttachDependencies || opt.Detach || opt.AbortOnContainerFailure || opt.AbortOnContainerExit
		msg := "WithWatch, WithAttach, WithAttachDependencies, WithDetach, WithAbortOnContainerFailure, WithAbortOnContainerExit and WithWait cannot be used together"
		opt.addErrorWhen(fail, "--wait", msg)
		flags = append(flags, "--wait")
	}
	if opt.WaitTimeout != nil {
		flags = append(flags, "--wait-timeout", strconv.Itoa(*opt.WaitTimeout))
	}
	if opt.Watch {
		flags = append(flags, "--watch")
	}
	if opt.Yes {
		flags = append(flags, "--yes")
	}
	if len(opt.Errs) > 0 {
		return nil, errors.Join(opt.Errs...)
	}

	return flags, nil
}

// ComposeDownOptions is the options for the compose down command
type ComposeDownOptions struct {
	RemoveOrphans bool
	Timeout       *int
	RemoveImage   *string
	RemoveVolumes bool

	Flags  []string
	Writer io.Writer
}

func (opt *ComposeDownOptions) GenerateFlags() ([]string, error) {
	flags := []string{"down"}
	if opt.RemoveOrphans {
		flags = append(flags, "--remove-orphans")
	}
	if opt.Timeout != nil {
		flags = append(flags, "--timeout", strconv.Itoa(*opt.Timeout))
	}
	if opt.RemoveImage != nil {
		flags = append(flags, "--rmi", *opt.RemoveImage)
	}
	if opt.RemoveVolumes {
		flags = append(flags, "--volumes")
	}
	return flags, nil
}

// ComposeLogsOptions is the options for the compose logs command
type ComposeLogsOptions struct {
	Tail        *int
	Follow      bool
	NoLogPrefix bool

	Writer io.Writer
	Flags  []string
}

func (opt *ComposeLogsOptions) GenerateFlags() ([]string, error) {
	flags := []string{"logs"}
	if opt.Tail != nil {
		flags = append(flags, "--tail", strconv.Itoa(*opt.Tail))
	}
	if opt.Follow {
		flags = append(flags, "--follow")
	}
	if opt.NoLogPrefix {
		flags = append(flags, "--no-log-prefix")
	}
	return flags, nil
}
