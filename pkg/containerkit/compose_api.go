package containerkit

import (
	"errors"
	"fmt"
	"io"
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
	IgnoreOrphans           bool
	RenewAnonVolumes        bool
	Scale                   []UpScale
	Timeout                 *int
	Timestamps              bool
	Wait                    bool
	WaitTimeout             *int
	Watch                   bool
	Yes                     bool

	Profiles     []string
	Writer       io.Writer
	ServiceNames []string
}

// UpScale is the scale for the compose up command
type UpScale struct {
	Service string
	Num     int
}

// validate rejects option combinations the Compose SDK cannot run together.
func (opt *Up) validate() error {
	var errs []error
	detachConflict := opt.Attach != nil ||
		opt.AttachDependencies ||
		opt.Watch ||
		opt.AbortOnContainerExit ||
		opt.AbortOnContainerFailure
	if opt.Detach && detachConflict {
		errs = append(errs, fmt.Errorf(
			"attach, attach dependencies, watch, abort on container exit, "+
				"and abort on container failure cannot be combined with detach",
		))
	}
	if opt.Attach != nil && opt.AttachDependencies {
		errs = append(errs, fmt.Errorf("attach and attach dependencies cannot be used together"))
	}
	for _, scale := range opt.Scale {
		if scale.Service == "" || scale.Num < 0 {
			errs = append(errs, fmt.Errorf("service name required and scale must be non-negative"))
		}
	}
	waitConflict := opt.Watch ||
		opt.Attach != nil ||
		opt.AttachDependencies ||
		opt.AbortOnContainerFailure ||
		opt.AbortOnContainerExit
	if opt.Wait && waitConflict {
		errs = append(errs, fmt.Errorf(
			"watch, attach, attach dependencies, "+
				"and abort on container exit cannot be combined with wait",
		))
	}
	return errors.Join(errs...)
}

// Down is the options for the compose down command
type Down struct {
	RemoveOrphans bool
	Timeout       *int
	RemoveImage   *string
	RemoveVolumes bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Logs is the options for the compose logs command
type Logs struct {
	Tail        *int
	Follow      bool
	Index       *int
	NoColor     bool
	NoLogPrefix bool
	Since       string
	Timestamps  bool
	Until       string

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// EventsOpt is the options for the compose events command.
// A nil EventsOpt watches every service and enables no profiles.
type EventsOpt struct {
	Service  string
	Profiles []string
	Since    string
	Until    string
}

// Kill is the options for the compose kill command
type Kill struct {
	Signal        *string
	RemoveOrphans bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
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
}

// Start is the options for the compose start command
type Start struct {
	Wait         bool
	WaitTimeout  *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Stop is the options for the compose stop command
type Stop struct {
	Timeout      *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Restart is the options for the compose restart command
type Restart struct {
	NoDeps       bool
	Timeout      *int
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
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
	Service  string // required
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
	Service    string // required
	Repository string // optional REPOSITORY[:TAG]
}

// ComposeConfig is the options for the compose config command.
// Named ComposeConfig because Config is too generic (MergedConfig, container.Config).
type ComposeConfig struct {
	PrintEnvironment bool
	Format           string // yaml | json
	PrintImages      bool
	PrintModels      bool
	PrintNetworks    bool
	NoConsistency    bool
	NoEnvResolution  bool
	NoInterpolate    bool
	NoNormalize      bool
	NoPathResolution bool
	Output           string
	PrintProfiles    bool
	Quiet            bool
	PrintServices    bool
	PrintVariables   bool
	PrintVolumes     bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
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
	Source   string // required
	Dest     string // required
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
	IgnoreOrphans bool
	Scale         []UpScale
	Yes           bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// ComposeExport is the options for the compose export command.
// Named ComposeExport to avoid confusion with Project.Export.
// Service is required.
type ComposeExport struct {
	Index  *int
	Output string

	Writer   io.Writer
	Profiles []string
	Service  string // required
}

// Images is the options for the compose images command.
type Images struct {
	Format string // table | json
	Quiet  bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Ls is the options for the compose ls command (lists Compose projects on the engine).
type Ls struct {
	All    bool
	Filter []string
	Format string
	Quiet  bool

	Writer   io.Writer
	Profiles []string
}

// Pause is the options for the compose pause command.
type Pause struct {
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Port is the options for the compose port command.
// Service and PrivatePort are required.
type Port struct {
	Index    *int
	Protocol string // tcp | udp

	Writer      io.Writer
	Profiles    []string
	Service     string // required
	PrivatePort int    // required
}

// Publish is the options for the compose publish command.
// Repository is required.
type Publish struct {
	App                 bool
	OCIVersion          string
	ResolveImageDigests bool
	InsecureRegistry    bool
	WithEnv             bool
	Yes                 bool

	Writer     io.Writer
	Profiles   []string
	Repository string // required REPOSITORY[:TAG]
}

// Push is the options for the compose push command.
type Push struct {
	IgnorePushFailures bool
	IncludeDeps        bool
	Quiet              bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Rm is the options for the compose rm command.
type Rm struct {
	Force   bool
	Stop    bool
	Volumes bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
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
	Pull          string
	Quiet         bool
	QuietBuild    bool
	QuietPull     bool
	RemoveOrphans bool
	IgnoreOrphans bool
	Rm            bool
	UseAliases    bool
	User          string
	Workdir       string

	Writer   io.Writer
	Stdin    io.Reader
	Profiles []string
	Service  string   // required
	Command  []string // optional
}

// Scale is the options for the compose scale command.
// Replicas is required (at least one SERVICE=NUM).
type Scale struct {
	NoDeps   bool
	Replicas []UpScale

	Writer   io.Writer
	Profiles []string
}

// Stats is the options for the compose stats command.
type Stats struct {
	All      bool
	Format   string
	NoStream bool
	NoTrunc  bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Top is the options for the compose top command.
type Top struct {
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Unpause is the options for the compose unpause command.
type Unpause struct {
	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Version is the options for the compose version command.
type Version struct {
	Format string // pretty | json
	Short  bool

	Writer   io.Writer
	Profiles []string
}

// Volumes is the options for the compose volumes command.
type Volumes struct {
	Format string
	Quiet  bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Wait is the options for the compose wait command.
// ServiceNames is required.
type Wait struct {
	DownProject bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string // required
}

// Watch is the options for the compose watch command.
type Watch struct {
	NoUp    bool
	NoPrune bool // --prune=false; Compose defaults to true
	Quiet   bool

	Writer       io.Writer
	Profiles     []string
	ServiceNames []string
}

// Viz is the options for compose alpha viz.
type Viz struct {
	Networks        bool
	Ports           bool
	Image           bool
	IndentationSize *int // default 1
	Spaces          bool

	Writer   io.Writer
	Profiles []string
}

// Generate is the options for compose alpha generate.
type Generate struct {
	Name       string
	ProjectDir string
	Format     string // yaml or json; default yaml
	Containers []string

	Writer io.Writer
}

// BridgeConvert is the options for compose bridge convert.
type BridgeConvert struct {
	Output          string // default out
	Templates       string
	Transformations []string
	Yes             bool

	Writer   io.Writer
	Profiles []string
}

// BridgeTransformationsList is the options for compose bridge transformations list.
type BridgeTransformationsList struct {
	Format string // table or json; default table
	Quiet  bool

	Writer io.Writer
}

// BridgeTransformationsCreate is the options for compose bridge transformations create.
type BridgeTransformationsCreate struct {
	From string
	Dest string

	Writer io.Writer
}

// UnsupportedAttribute is a compose-file attribute the schema accepts and this runtime does not apply outside Swarm.
type UnsupportedAttribute struct {
	Service string
	Path    string
	Reason  string
}

// ProjectLoad is the options for loading a Compose project from files.
type ProjectLoad struct {
	Name               string
	Files              []string
	WorkingDir         string
	EnvFiles           []string
	Profiles           []string
	Services           []string
	Offline            bool
	AllResources       bool
	Compatibility      bool
	InsecureRegistries []string
	OnUnsupported      func([]UnsupportedAttribute)

	Writer io.Writer
}
