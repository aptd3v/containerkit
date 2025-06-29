package compose

import (
	"errors"
	"fmt"
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

	Flags []string
}

type ComposeUpScale struct {
	Service string
	Num     int
}

func (opt *ComposeUpOptions) addErrorWhen(cond bool, errs *[]error, flag string, msg string) {
	if cond {
		*errs = append(*errs, NewComposeFlagError(flag, msg))
	}
}

// GenerateFlags generates the flags for the compose up command
//
// It will return a slice of flags to append to the command Eg.
//
//	[]string{"up", "--detach", ...}
func (opt *ComposeUpOptions) GenerateFlags() ([]string, error) {
	errs := []error{}
	flags := []string{"up"}
	if opt.Detach {
		fail := opt.Attach != nil || opt.AttachDependencies || opt.Watch || opt.AbortOnContainerExit || opt.AbortOnContainerFailure
		msg := "WithAttach, WithAttachDependencies, WithWatch, WithAbortOnContainerExit and WithAbortOnContainerFailure cannot be used together"
		opt.addErrorWhen(fail, &errs, "--detach", msg)
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
		opt.addErrorWhen(opt.AttachDependencies, &errs, "--attach", "WithAttach and WithAttachDependencies cannot be used together")
		flags = append(flags, "--attach", *opt.Attach)
	}
	if opt.AttachDependencies {
		opt.addErrorWhen(opt.Attach != nil, &errs, "--attach-dependencies", "WithAttach and WithAttachDependencies cannot be used together")
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
			opt.addErrorWhen(scale.Service == "" || scale.Num < 0, &errs, "--scale", "Service name required and scale must be non-negative")
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
		opt.addErrorWhen(fail, &errs, "--wait", msg)
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
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return flags, nil
}
