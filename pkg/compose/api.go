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

	Flags    []string
	Writer   io.Writer
	Profiles []string
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

	Writer   io.Writer
	Flags    []string
	Profiles []string
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

// ComposeKillOptions is the options for the compose kill command
type ComposeKillOptions struct {
	Signal        *string
	RemoveOrphans bool

	Flags    []string
	Writer   io.Writer
	Profiles []string
}

// GenerateFlags generates the flags for the compose kill command
//
// It will return a slice of flags to append to the command Eg.
//
//	[]string{"kill", "--signal", "SIGKILL", "--remove-orphans"}
func (opt *ComposeKillOptions) GenerateFlags() ([]string, error) {
	flags := []string{"kill"}
	if opt.Signal != nil {
		flags = append(flags, "--signal", *opt.Signal)
	}
	if opt.RemoveOrphans {
		flags = append(flags, "--remove-orphans")
	}
	return flags, nil
}

// ComposePsOptions is the options for the compose ps command
type ComposePsOptions struct {
	All          bool
	Filter       []string // e.g. "status=running"
	Format       string
	NoTrunc      bool
	Orphans      *bool // nil = default (true); false = --no-orphans
	Quiet        bool
	Services     bool // --services: display services
	Status       string
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *ComposePsOptions) GenerateFlags() ([]string, error) {
	flags := []string{"ps"}
	if opt.All {
		flags = append(flags, "--all")
	}
	for _, f := range opt.Filter {
		flags = append(flags, "--filter", f)
	}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.NoTrunc {
		flags = append(flags, "--no-trunc")
	}
	if opt.Orphans != nil && !*opt.Orphans {
		flags = append(flags, "--no-orphans")
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	if opt.Services {
		flags = append(flags, "--services")
	}
	if opt.Status != "" {
		flags = append(flags, "--status", opt.Status)
	}
	return flags, nil
}

// ComposeStartOptions is the options for the compose start command
type ComposeStartOptions struct {
	Wait         bool
	WaitTimeout  *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *ComposeStartOptions) GenerateFlags() ([]string, error) {
	flags := []string{"start"}
	if opt.Wait {
		flags = append(flags, "--wait")
	}
	if opt.WaitTimeout != nil {
		flags = append(flags, "--wait-timeout", strconv.Itoa(*opt.WaitTimeout))
	}
	return flags, nil
}

// ComposeStopOptions is the options for the compose stop command
type ComposeStopOptions struct {
	Timeout      *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *ComposeStopOptions) GenerateFlags() ([]string, error) {
	flags := []string{"stop"}
	if opt.Timeout != nil {
		flags = append(flags, "--timeout", strconv.Itoa(*opt.Timeout))
	}
	return flags, nil
}

// ComposeRestartOptions is the options for the compose restart command
type ComposeRestartOptions struct {
	NoDeps       bool
	Timeout      *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *ComposeRestartOptions) GenerateFlags() ([]string, error) {
	flags := []string{"restart"}
	if opt.NoDeps {
		flags = append(flags, "--no-deps")
	}
	if opt.Timeout != nil {
		flags = append(flags, "--timeout", strconv.Itoa(*opt.Timeout))
	}
	return flags, nil
}

// ComposeBuildOptions is the options for the compose build command
type ComposeBuildOptions struct {
	BuildArg         []string // key=value, appended as --build-arg each
	Builder           string
	Check            bool
	Memory           string   // e.g. "2G"
	NoCache          bool
	Print            bool
	Provenance       bool
	Pull             bool
	Push             bool
	Quiet            bool
	SBOM             bool
	SSH              []string // e.g. "default" or "key=path"
	WithDependencies bool
	Writer           io.Writer
	Profiles         []string
	ServiceNames     []string
	Flags            []string
}

func (opt *ComposeBuildOptions) GenerateFlags() ([]string, error) {
	flags := []string{"build"}
	for _, a := range opt.BuildArg {
		flags = append(flags, "--build-arg", a)
	}
	if opt.Builder != "" {
		flags = append(flags, "--builder", opt.Builder)
	}
	if opt.Check {
		flags = append(flags, "--check")
	}
	if opt.Memory != "" {
		flags = append(flags, "--memory", opt.Memory)
	}
	if opt.NoCache {
		flags = append(flags, "--no-cache")
	}
	if opt.Print {
		flags = append(flags, "--print")
	}
	if opt.Provenance {
		flags = append(flags, "--provenance")
	}
	if opt.Pull {
		flags = append(flags, "--pull")
	}
	if opt.Push {
		flags = append(flags, "--push")
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	if opt.SBOM {
		flags = append(flags, "--sbom")
	}
	for _, s := range opt.SSH {
		flags = append(flags, "--ssh", s)
	}
	if opt.WithDependencies {
		flags = append(flags, "--with-dependencies")
	}
	return flags, nil
}

// ComposePullOptions is the options for the compose pull command
type ComposePullOptions struct {
	IgnoreBuildable     bool
	IgnorePullFailures  bool
	IncludeDeps         bool
	Policy              string   // e.g. "missing", "always", "never"
	Quiet               bool
	Writer              io.Writer
	Profiles            []string
	ServiceNames        []string
	Flags               []string
}

func (opt *ComposePullOptions) GenerateFlags() ([]string, error) {
	flags := []string{"pull"}
	if opt.IgnoreBuildable {
		flags = append(flags, "--ignore-buildable")
	}
	if opt.IgnorePullFailures {
		flags = append(flags, "--ignore-pull-failures")
	}
	if opt.IncludeDeps {
		flags = append(flags, "--include-deps")
	}
	if opt.Policy != "" {
		flags = append(flags, "--policy", opt.Policy)
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	return flags, nil
}

// ComposeExecOptions is the options for the compose exec command.
// Usage: docker compose exec [OPTIONS] SERVICE COMMAND [ARGS...]
// Service and Command are required; set via WithService and WithCommand.
type ComposeExecOptions struct {
	Detach    bool
	Env       []string // key=value, passed as -e each
	Index     *int     // --index for multi-replica services
	NoTTY     bool     // -T: disable TTY
	Privileged bool
	User      string   // -u
	Workdir   string   // -w
	Writer    io.Writer
	Stdin     io.Reader // optional; forwarded to the container after compose file is read
	Profiles  []string
	Service   string   // required: service name
	Command   []string  // required: command and args (e.g. ["sh", "-c", "echo hi"])
}

func (opt *ComposeExecOptions) GenerateFlags() ([]string, error) {
	flags := []string{"exec"}
	if opt.Detach {
		flags = append(flags, "--detach")
	}
	for _, e := range opt.Env {
		flags = append(flags, "--env", e)
	}
	if opt.Index != nil {
		flags = append(flags, "--index", strconv.Itoa(*opt.Index))
	}
	if opt.NoTTY {
		flags = append(flags, "--no-tty")
	}
	if opt.Privileged {
		flags = append(flags, "--privileged")
	}
	if opt.User != "" {
		flags = append(flags, "--user", opt.User)
	}
	if opt.Workdir != "" {
		flags = append(flags, "--workdir", opt.Workdir)
	}
	return flags, nil
}
