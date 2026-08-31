package containerkit

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Up is the options for the compose up command
type Up struct {
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
	Scale                   []UpScale
	Timeout                 *int
	Timestamps              bool
	Wait                    bool
	WaitTimeout             *int
	Watch                   bool
	Yes                     bool

	Profiles     []string
	Flags        []string
	Writer       io.Writer
	ServiceNames []string
	Errs         []error
}

// UpScale is the scale for the compose up command
type UpScale struct {
	Service string
	Num     int
}

// addErrorWhen adds an error to the error slice if the condition is true
func (opt *Up) addErrorWhen(cond bool, flag string, msg string) {
	if cond {
		opt.Errs = append(opt.Errs, NewComposeFlagError(flag, msg))
	}
}

// GenerateFlags generates the flags for the compose up command
//
// It will return a slice of flags to append to the command Eg.
//
//	[]string{"up", "--detach", ...}
func (opt *Up) GenerateFlags() (flags []string, err error) {
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

// Down is the options for the compose down command
type Down struct {
	RemoveOrphans bool
	Timeout       *int
	RemoveImage   *string
	RemoveVolumes bool

	Flags        []string
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

func (opt *Down) GenerateFlags() ([]string, error) {
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

// Logs is the options for the compose logs command
type Logs struct {
	Tail        *int
	Follow      bool
	NoLogPrefix bool

	Writer       io.Writer
	Flags        []string
	Profiles     []string
	ServiceNames []string
}

func (opt *Logs) GenerateFlags() ([]string, error) {
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

// Kill is the options for the compose kill command
type Kill struct {
	Signal        *string
	RemoveOrphans bool

	Flags        []string
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// GenerateFlags generates the flags for the compose kill command
//
// It will return a slice of flags to append to the command Eg.
//
//	[]string{"kill", "--signal", "SIGKILL", "--remove-orphans"}
func (opt *Kill) GenerateFlags() ([]string, error) {
	flags := []string{"kill"}
	if opt.Signal != nil {
		flags = append(flags, "--signal", *opt.Signal)
	}
	if opt.RemoveOrphans {
		flags = append(flags, "--remove-orphans")
	}
	return flags, nil
}

// Ps is the options for the compose ps command
type Ps struct {
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

func (opt *Ps) GenerateFlags() ([]string, error) {
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

// Start is the options for the compose start command
type Start struct {
	Wait         bool
	WaitTimeout  *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *Start) GenerateFlags() ([]string, error) {
	flags := []string{"start"}
	if opt.Wait {
		flags = append(flags, "--wait")
	}
	if opt.WaitTimeout != nil {
		flags = append(flags, "--wait-timeout", strconv.Itoa(*opt.WaitTimeout))
	}
	return flags, nil
}

// Stop is the options for the compose stop command
type Stop struct {
	Timeout      *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *Stop) GenerateFlags() ([]string, error) {
	flags := []string{"stop"}
	if opt.Timeout != nil {
		flags = append(flags, "--timeout", strconv.Itoa(*opt.Timeout))
	}
	return flags, nil
}

// Restart is the options for the compose restart command
type Restart struct {
	NoDeps       bool
	Timeout      *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
	Flags        []string
}

func (opt *Restart) GenerateFlags() ([]string, error) {
	flags := []string{"restart"}
	if opt.NoDeps {
		flags = append(flags, "--no-deps")
	}
	if opt.Timeout != nil {
		flags = append(flags, "--timeout", strconv.Itoa(*opt.Timeout))
	}
	return flags, nil
}

// ComposeBuild is the options for the compose build command
type ComposeBuild struct {
	BuildArg         []string // key=value, appended as --build-arg each
	Builder          string
	Check            bool
	Memory           string // e.g. "2G"
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

func (opt *ComposeBuild) GenerateFlags() ([]string, error) {
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

// Pull is the options for the compose pull command
type Pull struct {
	IgnoreBuildable    bool
	IgnorePullFailures bool
	IncludeDeps        bool
	Policy             string // e.g. "missing", "always", "never"
	Quiet              bool
	Writer             io.Writer
	Profiles           []string
	ServiceNames       []string
	Flags              []string
}

func (opt *Pull) GenerateFlags() ([]string, error) {
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

// Exec is the options for the compose exec command.
// Usage: docker compose exec [OPTIONS] SERVICE COMMAND [ARGS...]
// Service and Command are required; set via WithService and WithCommand.
type Exec struct {
	Detach     bool
	Env        []string // key=value, passed as -e each
	Index      *int     // --index for multi-replica services
	NoTTY      bool     // -T: disable TTY
	Privileged bool
	User       string // -u
	Workdir    string // -w
	Writer     io.Writer
	Stdin      io.Reader // optional; forwarded to the container after compose file is read
	Profiles   []string
	Service    string   // required: service name
	Command    []string // required: command and args (e.g. ["sh", "-c", "echo hi"])
}

func (opt *Exec) GenerateFlags() ([]string, error) {
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

// Attach is the options for the compose attach command.
// Service is required.
type Attach struct {
	DetachKeys string
	Index      *int
	NoStdin    bool
	NoSigProxy bool // --sig-proxy=false; Compose defaults to true

	Writer   io.Writer
	Stdin    io.Reader
	Profiles []string
	Flags    []string
	Service  string // required
}

func (opt *Attach) GenerateFlags() ([]string, error) {
	flags := []string{"attach"}
	if opt.DetachKeys != "" {
		flags = append(flags, "--detach-keys", opt.DetachKeys)
	}
	if opt.Index != nil {
		flags = append(flags, "--index", strconv.Itoa(*opt.Index))
	}
	if opt.NoStdin {
		flags = append(flags, "--no-stdin")
	}
	if opt.NoSigProxy {
		flags = append(flags, "--sig-proxy=false")
	}
	return flags, nil
}

// Commit is the options for the compose commit command.
// Service is required.
type Commit struct {
	Author  string
	Change  []string // Dockerfile instructions applied to the image
	Index   *int
	Message string
	NoPause bool // --pause=false; Compose defaults to true

	Writer     io.Writer
	Profiles   []string
	Flags      []string
	Service    string // required
	Repository string // optional REPOSITORY[:TAG]
}

func (opt *Commit) GenerateFlags() ([]string, error) {
	flags := []string{"commit"}
	if opt.Author != "" {
		flags = append(flags, "--author", opt.Author)
	}
	for _, c := range opt.Change {
		flags = append(flags, "--change", c)
	}
	if opt.Index != nil {
		flags = append(flags, "--index", strconv.Itoa(*opt.Index))
	}
	if opt.Message != "" {
		flags = append(flags, "--message", opt.Message)
	}
	if opt.NoPause {
		flags = append(flags, "--pause=false")
	}
	return flags, nil
}

// ComposeConfig is the options for the compose config command.
// Named ComposeConfig because Config is too generic (MergedConfig, container.Config).
type ComposeConfig struct {
	PrintEnvironment    bool
	Format              string // yaml | json
	Hash                string // service name or "*"
	PrintImages         bool
	LockImageDigests    bool
	PrintModels         bool
	PrintNetworks       bool
	NoConsistency       bool
	NoEnvResolution     bool
	NoInterpolate       bool
	NoNormalize         bool
	NoPathResolution    bool
	Output              string
	PrintProfiles       bool
	Quiet               bool
	ResolveImageDigests bool
	PrintServices       bool
	PrintVariables      bool
	PrintVolumes        bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *ComposeConfig) GenerateFlags() ([]string, error) {
	flags := []string{"config"}
	if opt.PrintEnvironment {
		flags = append(flags, "--environment")
	}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.Hash != "" {
		flags = append(flags, "--hash", opt.Hash)
	}
	if opt.PrintImages {
		flags = append(flags, "--images")
	}
	if opt.LockImageDigests {
		flags = append(flags, "--lock-image-digests")
	}
	if opt.PrintModels {
		flags = append(flags, "--models")
	}
	if opt.PrintNetworks {
		flags = append(flags, "--networks")
	}
	if opt.NoConsistency {
		flags = append(flags, "--no-consistency")
	}
	if opt.NoEnvResolution {
		flags = append(flags, "--no-env-resolution")
	}
	if opt.NoInterpolate {
		flags = append(flags, "--no-interpolate")
	}
	if opt.NoNormalize {
		flags = append(flags, "--no-normalize")
	}
	if opt.NoPathResolution {
		flags = append(flags, "--no-path-resolution")
	}
	if opt.Output != "" {
		flags = append(flags, "--output", opt.Output)
	}
	if opt.PrintProfiles {
		flags = append(flags, "--profiles")
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	if opt.ResolveImageDigests {
		flags = append(flags, "--resolve-image-digests")
	}
	if opt.PrintServices {
		flags = append(flags, "--services")
	}
	if opt.PrintVariables {
		flags = append(flags, "--variables")
	}
	if opt.PrintVolumes {
		flags = append(flags, "--volumes")
	}
	return flags, nil
}

// Cp is the options for the compose cp command.
// Source and Dest are required (SERVICE:PATH or PATH or "-").
type Cp struct {
	All        bool
	Archive    bool
	FollowLink bool
	Index      *int

	Writer   io.Writer
	Stdin    io.Reader // used when Source or Dest is "-"
	Profiles []string
	Flags    []string
	Source   string // required
	Dest     string // required
}

func (opt *Cp) GenerateFlags() ([]string, error) {
	flags := []string{"cp"}
	if opt.All {
		flags = append(flags, "--all")
	}
	if opt.Archive {
		flags = append(flags, "--archive")
	}
	if opt.FollowLink {
		flags = append(flags, "--follow-link")
	}
	if opt.Index != nil {
		flags = append(flags, "--index", strconv.Itoa(*opt.Index))
	}
	return flags, nil
}

// Create is the options for the compose create command.
type Create struct {
	Build         bool
	ForceRecreate bool
	NoBuild       bool
	NoRecreate    bool
	Pull          string
	QuietPull     bool
	RemoveOrphans bool
	Scale         []UpScale
	Yes           bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Create) GenerateFlags() ([]string, error) {
	flags := []string{"create"}
	if opt.Build {
		flags = append(flags, "--build")
	}
	if opt.ForceRecreate {
		flags = append(flags, "--force-recreate")
	}
	if opt.NoBuild {
		flags = append(flags, "--no-build")
	}
	if opt.NoRecreate {
		flags = append(flags, "--no-recreate")
	}
	if opt.Pull != "" {
		flags = append(flags, "--pull", opt.Pull)
	}
	if opt.QuietPull {
		flags = append(flags, "--quiet-pull")
	}
	if opt.RemoveOrphans {
		flags = append(flags, "--remove-orphans")
	}
	for _, s := range opt.Scale {
		flags = append(flags, "--scale", fmt.Sprintf("%s=%d", s.Service, s.Num))
	}
	if opt.Yes {
		flags = append(flags, "--yes")
	}
	return flags, nil
}

// ComposeExport is the options for the compose export command.
// Named ComposeExport to avoid confusion with Project.Export.
// Service is required.
type ComposeExport struct {
	Index  *int
	Output string

	Writer   io.Writer
	Profiles []string
	Flags    []string
	Service  string // required
}

func (opt *ComposeExport) GenerateFlags() ([]string, error) {
	flags := []string{"export"}
	if opt.Index != nil {
		flags = append(flags, "--index", strconv.Itoa(*opt.Index))
	}
	if opt.Output != "" {
		flags = append(flags, "--output", opt.Output)
	}
	return flags, nil
}

// Images is the options for the compose images command.
type Images struct {
	Format string // table | json
	Quiet  bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Images) GenerateFlags() ([]string, error) {
	flags := []string{"images"}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	return flags, nil
}

// Ls is the options for the compose ls command (lists Compose projects on the engine).
type Ls struct {
	All    bool
	Filter []string
	Format string
	Quiet  bool

	Writer   io.Writer
	Profiles []string
	Flags    []string
}

func (opt *Ls) GenerateFlags() ([]string, error) {
	flags := []string{"ls"}
	if opt.All {
		flags = append(flags, "--all")
	}
	for _, f := range opt.Filter {
		flags = append(flags, "--filter", f)
	}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	return flags, nil
}

// Pause is the options for the compose pause command.
type Pause struct {
	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Pause) GenerateFlags() ([]string, error) {
	return []string{"pause"}, nil
}

// Port is the options for the compose port command.
// Service and PrivatePort are required.
type Port struct {
	Index    *int
	Protocol string // tcp | udp

	Writer      io.Writer
	Profiles    []string
	Flags       []string
	Service     string // required
	PrivatePort int    // required
}

func (opt *Port) GenerateFlags() ([]string, error) {
	flags := []string{"port"}
	if opt.Index != nil {
		flags = append(flags, "--index", strconv.Itoa(*opt.Index))
	}
	if opt.Protocol != "" {
		flags = append(flags, "--protocol", opt.Protocol)
	}
	return flags, nil
}

// Publish is the options for the compose publish command.
// Repository is required.
type Publish struct {
	App                 bool
	OCIVersion          string
	ResolveImageDigests bool
	WithEnv             bool
	Yes                 bool

	Writer     io.Writer
	Profiles   []string
	Flags      []string
	Repository string // required REPOSITORY[:TAG]
}

func (opt *Publish) GenerateFlags() ([]string, error) {
	flags := []string{"publish"}
	if opt.App {
		flags = append(flags, "--app")
	}
	if opt.OCIVersion != "" {
		flags = append(flags, "--oci-version", opt.OCIVersion)
	}
	if opt.ResolveImageDigests {
		flags = append(flags, "--resolve-image-digests")
	}
	if opt.WithEnv {
		flags = append(flags, "--with-env")
	}
	if opt.Yes {
		flags = append(flags, "--yes")
	}
	return flags, nil
}

// Push is the options for the compose push command.
type Push struct {
	IgnorePushFailures bool
	IncludeDeps        bool
	Quiet              bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Push) GenerateFlags() ([]string, error) {
	flags := []string{"push"}
	if opt.IgnorePushFailures {
		flags = append(flags, "--ignore-push-failures")
	}
	if opt.IncludeDeps {
		flags = append(flags, "--include-deps")
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	return flags, nil
}

// Rm is the options for the compose rm command.
type Rm struct {
	Force   bool
	Stop    bool
	Volumes bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Rm) GenerateFlags() ([]string, error) {
	flags := []string{"rm"}
	if opt.Force {
		flags = append(flags, "--force")
	}
	if opt.Stop {
		flags = append(flags, "--stop")
	}
	if opt.Volumes {
		flags = append(flags, "--volumes")
	}
	return flags, nil
}

// Run is the options for the compose run command.
// Service is required. Command is optional (defaults to the service command).
type Run struct {
	Build         bool
	CapAdd        []string
	CapDrop       []string
	Detach        bool
	Entrypoint    string
	Env           []string
	EnvFromFile   []string
	Interactive   *bool // nil = Compose default (true)
	Label         []string
	Name          string
	NoDeps        bool
	NoTTY         bool
	Publish       []string
	Pull          string
	Quiet         bool
	QuietBuild    bool
	QuietPull     bool
	RemoveOrphans bool
	Rm            bool
	ServicePorts  bool
	UseAliases    bool
	User          string
	Volume        []string
	Workdir       string

	Writer   io.Writer
	Stdin    io.Reader
	Profiles []string
	Flags    []string
	Service  string   // required
	Command  []string // optional
}

func (opt *Run) GenerateFlags() ([]string, error) {
	flags := []string{"run"}
	if opt.Build {
		flags = append(flags, "--build")
	}
	for _, c := range opt.CapAdd {
		flags = append(flags, "--cap-add", c)
	}
	for _, c := range opt.CapDrop {
		flags = append(flags, "--cap-drop", c)
	}
	if opt.Detach {
		flags = append(flags, "--detach")
	}
	if opt.Entrypoint != "" {
		flags = append(flags, "--entrypoint", opt.Entrypoint)
	}
	for _, e := range opt.Env {
		flags = append(flags, "--env", e)
	}
	for _, f := range opt.EnvFromFile {
		flags = append(flags, "--env-from-file", f)
	}
	if opt.Interactive != nil {
		if *opt.Interactive {
			flags = append(flags, "--interactive")
		} else {
			flags = append(flags, "--interactive=false")
		}
	}
	for _, l := range opt.Label {
		flags = append(flags, "--label", l)
	}
	if opt.Name != "" {
		flags = append(flags, "--name", opt.Name)
	}
	if opt.NoDeps {
		flags = append(flags, "--no-deps")
	}
	if opt.NoTTY {
		flags = append(flags, "--no-TTY")
	}
	for _, p := range opt.Publish {
		flags = append(flags, "--publish", p)
	}
	if opt.Pull != "" {
		flags = append(flags, "--pull", opt.Pull)
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	if opt.QuietBuild {
		flags = append(flags, "--quiet-build")
	}
	if opt.QuietPull {
		flags = append(flags, "--quiet-pull")
	}
	if opt.RemoveOrphans {
		flags = append(flags, "--remove-orphans")
	}
	if opt.Rm {
		flags = append(flags, "--rm")
	}
	if opt.ServicePorts {
		flags = append(flags, "--service-ports")
	}
	if opt.UseAliases {
		flags = append(flags, "--use-aliases")
	}
	if opt.User != "" {
		flags = append(flags, "--user", opt.User)
	}
	for _, v := range opt.Volume {
		flags = append(flags, "--volume", v)
	}
	if opt.Workdir != "" {
		flags = append(flags, "--workdir", opt.Workdir)
	}
	return flags, nil
}

// Scale is the options for the compose scale command.
// Replicas is required (at least one SERVICE=NUM).
type Scale struct {
	NoDeps   bool
	Replicas []UpScale

	Writer   io.Writer
	Profiles []string
	Flags    []string
}

func (opt *Scale) GenerateFlags() ([]string, error) {
	flags := []string{"scale"}
	if opt.NoDeps {
		flags = append(flags, "--no-deps")
	}
	return flags, nil
}

// Stats is the options for the compose stats command.
type Stats struct {
	All      bool
	Format   string
	NoStream bool
	NoTrunc  bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Stats) GenerateFlags() ([]string, error) {
	flags := []string{"stats"}
	if opt.All {
		flags = append(flags, "--all")
	}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.NoStream {
		flags = append(flags, "--no-stream")
	}
	if opt.NoTrunc {
		flags = append(flags, "--no-trunc")
	}
	return flags, nil
}

// Top is the options for the compose top command.
type Top struct {
	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Top) GenerateFlags() ([]string, error) {
	return []string{"top"}, nil
}

// Unpause is the options for the compose unpause command.
type Unpause struct {
	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Unpause) GenerateFlags() ([]string, error) {
	return []string{"unpause"}, nil
}

// Version is the options for the compose version command.
type Version struct {
	Format string // pretty | json
	Short  bool

	Writer   io.Writer
	Profiles []string
	Flags    []string
}

func (opt *Version) GenerateFlags() ([]string, error) {
	flags := []string{"version"}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.Short {
		flags = append(flags, "--short")
	}
	return flags, nil
}

// Volumes is the options for the compose volumes command.
type Volumes struct {
	Format string
	Quiet  bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Volumes) GenerateFlags() ([]string, error) {
	flags := []string{"volumes"}
	if opt.Format != "" {
		flags = append(flags, "--format", opt.Format)
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	return flags, nil
}

// Wait is the options for the compose wait command.
// ServiceNames is required.
type Wait struct {
	DownProject bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string // required
}

func (opt *Wait) GenerateFlags() ([]string, error) {
	flags := []string{"wait"}
	if opt.DownProject {
		flags = append(flags, "--down-project")
	}
	return flags, nil
}

// Watch is the options for the compose watch command.
type Watch struct {
	NoUp    bool
	NoPrune bool // --prune=false; Compose defaults to true
	Quiet   bool

	Writer       io.Writer
	Profiles     []string
	Flags        []string
	ServiceNames []string
}

func (opt *Watch) GenerateFlags() ([]string, error) {
	flags := []string{"watch"}
	if opt.NoUp {
		flags = append(flags, "--no-up")
	}
	if opt.NoPrune {
		flags = append(flags, "--prune=false")
	}
	if opt.Quiet {
		flags = append(flags, "--quiet")
	}
	return flags, nil
}
