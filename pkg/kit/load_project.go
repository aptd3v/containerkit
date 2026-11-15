package kit

import (
	"context"
	"fmt"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	clitypes "github.com/aptd3v/containerkit/internal/compose/cli/types"
	composecli "github.com/compose-spec/compose-go/v2/cli"
)

type loadProjectAct struct {
	opts composeapi.ProjectLoadOptions
}

func (a loadProjectAct) sdk() composeapi.ProjectLoadOptions {
	out := a.opts
	out.ConfigPaths = append([]string(nil), a.opts.ConfigPaths...)
	out.EnvFiles = append([]string(nil), a.opts.EnvFiles...)
	out.Profiles = append([]string(nil), a.opts.Profiles...)
	out.Services = append([]string(nil), a.opts.Services...)
	out.ProjectOptionsFns = append([]clitypes.ProjectOptionsFn(nil), a.opts.ProjectOptionsFns...)
	out.LoadListeners = append([]composeapi.LoadListener(nil), a.opts.LoadListeners...)
	out.OCI.InsecureRegistries = append([]string(nil), a.opts.OCI.InsecureRegistries...)
	return out
}

// LoadProject loads a Compose project through a [Composer] using the Compose
// v5 SDK loader (remote includes, OCI, unsupported-attribute detection).
// Local YAML without a Composer is [ImportFile] / [ImportYAML].
func LoadProject(ctx context.Context, api Composer, opts ...LoadProjectOption) (*Project, error) {
	if api == nil {
		return nil, cfgErr("compose", "load", "LoadProject", fmt.Errorf("%w: compose client", errRequired))
	}
	var act loadProjectAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("load.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyLoadProject(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	out, err := api.LoadProject(ctx, act.sdk())
	if err != nil {
		resource := act.opts.ProjectName
		if resource == "" {
			resource = "load"
		}
		return nil, cfgErr("compose", resource, "LoadProject", err)
	}
	return out, nil
}

func (o NameOpt) applyLoadProject(a *loadProjectAct) error {
	if err := validateName(o.name); err != nil {
		return err
	}
	a.opts.ProjectName = o.name
	return nil
}

func (o FileOpt) applyLoadProject(a *loadProjectAct) error {
	if o.path == "" {
		return cfgErr("", "file", "File", fmt.Errorf("%w: compose file", errEmpty))
	}
	a.opts.ConfigPaths = append(a.opts.ConfigPaths, o.path)
	return nil
}

func (o ProjectDirectoryOpt) applyLoadProject(a *loadProjectAct) error {
	if o.s == "" {
		return cfgErr("", "project_directory", "ProjectDirectory", fmt.Errorf("%w: project directory", errEmpty))
	}
	a.opts.WorkingDir = o.s
	return nil
}

func (o EnvFileOpt) applyLoadProject(a *loadProjectAct) error {
	if o.path == "" {
		return cfgErr("", "env_file", "EnvFile", fmt.Errorf("%w: env file", errEmpty))
	}
	a.opts.EnvFiles = append(a.opts.EnvFiles, o.path)
	return nil
}

func (o ProfileOpt) applyLoadProject(a *loadProjectAct) error {
	if o.p == "" {
		return cfgErr("", "profile", "Profile", fmt.Errorf("%w: profile", errEmpty))
	}
	a.opts.Profiles = append(a.opts.Profiles, o.p)
	return nil
}

func (o ServicesOpt) applyLoadProject(a *loadProjectAct) error {
	o.set(&a.opts.Services)
	return nil
}

// AllResourcesOpt is the value returned by [AllResources].
// It implements [LoadProjectOption].
type AllResourcesOpt struct{ v bool }

// AllResources includes unused Compose resources on [LoadProject].
func AllResources(v bool) AllResourcesOpt { return AllResourcesOpt{v: v} }

func (o AllResourcesOpt) applyLoadProject(a *loadProjectAct) error {
	a.opts.All = o.v
	return nil
}

// OfflineOpt is the value returned by [Offline].
// It implements [LoadProjectOption].
type OfflineOpt struct{ v bool }

// Offline disables remote resource loading on [LoadProject].
func Offline(v bool) OfflineOpt { return OfflineOpt{v: v} }

func (o OfflineOpt) applyLoadProject(a *loadProjectAct) error {
	a.opts.Offline = o.v
	return nil
}

// CompatibilityOpt is the value returned by [Compatibility].
// It implements [LoadProjectOption].
type CompatibilityOpt struct{ v bool }

// Compatibility enables Compose v1 compatibility mode on [LoadProject].
func Compatibility(v bool) CompatibilityOpt { return CompatibilityOpt{v: v} }

func (o CompatibilityOpt) applyLoadProject(a *loadProjectAct) error {
	a.opts.Compatibility = o.v
	return nil
}

// InsecureRegistryOpt is the value returned by [InsecureRegistry].
// It implements [LoadProjectOption].
type InsecureRegistryOpt struct{ host string }

// InsecureRegistry adds an OCI insecure registry on [LoadProject].
func InsecureRegistry(host string) InsecureRegistryOpt {
	return InsecureRegistryOpt{host: host}
}

func (o InsecureRegistryOpt) applyLoadProject(a *loadProjectAct) error {
	if o.host == "" {
		return cfgErr("", "oci", "InsecureRegistry", fmt.Errorf("%w: registry", errEmpty))
	}
	a.opts.OCI.InsecureRegistries = append(a.opts.OCI.InsecureRegistries, o.host)
	return nil
}

// OnLoadOpt is the value returned by [OnLoad].
// It implements [LoadProjectOption].
type OnLoadOpt struct{ fn LoadListener }

// OnLoad registers a [LoadListener] on [LoadProject].
func OnLoad(fn LoadListener) OnLoadOpt { return OnLoadOpt{fn: fn} }

func (o OnLoadOpt) applyLoadProject(a *loadProjectAct) error {
	if o.fn == nil {
		return cfgErr("", "load", "OnLoad", fmt.Errorf("%w: listener", errRequired))
	}
	a.opts.LoadListeners = append(a.opts.LoadListeners, o.fn)
	return nil
}

// OnUnsupportedOpt is the value returned by [OnUnsupported].
// It implements [LoadProjectOption].
type OnUnsupportedOpt struct {
	fn func([]UnsupportedAttribute)
}

// OnUnsupported is invoked with compose-file attributes the runtime does not
// honor outside Swarm mode on [LoadProject].
func OnUnsupported(fn func([]UnsupportedAttribute)) OnUnsupportedOpt {
	return OnUnsupportedOpt{fn: fn}
}

func (o OnUnsupportedOpt) applyLoadProject(a *loadProjectAct) error {
	if o.fn == nil {
		return cfgErr("", "unsupported", "OnUnsupported", fmt.Errorf("%w: callback", errRequired))
	}
	a.opts.OnUnsupportedAttribute = o.fn
	return nil
}

// ProjectFnOpt is the value returned by [ProjectOption], [Interpolation], [Normalization],
// [Consistency], [ResolvedPaths], [DiscardEnvFile], [OsEnv], [DotEnv], [ConfigFileEnv],
// [DefaultConfigPath], [EnvironmentResolution], and [LabelsResolution].
// It implements [LoadProjectOption].
type ProjectFnOpt struct{ fn ProjectOptionsFn }

// ProjectOption appends a compose-go [ProjectOptionsFn] on [LoadProject].
func ProjectOption(fn ProjectOptionsFn) ProjectFnOpt { return ProjectFnOpt{fn: fn} }

func (o ProjectFnOpt) applyLoadProject(a *loadProjectAct) error {
	if o.fn == nil {
		return cfgErr("", "project_option", "ProjectOption", fmt.Errorf("%w: project option", errRequired))
	}
	a.opts.ProjectOptionsFns = append(a.opts.ProjectOptionsFns, o.fn)
	return nil
}

// Interpolation enables or disables compose-go interpolation on [LoadProject].
func Interpolation(v bool) ProjectFnOpt {
	return ProjectFnOpt{fn: composecli.WithInterpolation(v)}
}

// Normalization enables or disables compose-go normalization on [LoadProject].
func Normalization(v bool) ProjectFnOpt {
	return ProjectFnOpt{fn: composecli.WithNormalization(v)}
}

// Consistency enables or disables compose-go consistency checks on [LoadProject].
func Consistency(v bool) ProjectFnOpt {
	return ProjectFnOpt{fn: composecli.WithConsistency(v)}
}

// ResolvedPaths enables or disables compose-go path resolution on [LoadProject].
func ResolvedPaths(v bool) ProjectFnOpt {
	return ProjectFnOpt{fn: composecli.WithResolvedPaths(v)}
}

// DiscardEnvFile skips applying env_file values onto the project environment.
func DiscardEnvFile() ProjectFnOpt {
	return ProjectFnOpt{fn: composecli.WithDiscardEnvFile}
}

// OsEnv loads process environment into the project on [LoadProject].
func OsEnv() ProjectFnOpt { return ProjectFnOpt{fn: composecli.WithOsEnv} }

// DotEnv reads .env into the project environment on [LoadProject].
func DotEnv() ProjectFnOpt { return ProjectFnOpt{fn: composecli.WithDotEnv} }

// ConfigFileEnv reads COMPOSE_FILE on [LoadProject].
func ConfigFileEnv() ProjectFnOpt { return ProjectFnOpt{fn: composecli.WithConfigFileEnv} }

// DefaultConfigPath searches default compose.yaml paths on [LoadProject].
func DefaultConfigPath() ProjectFnOpt { return ProjectFnOpt{fn: composecli.WithDefaultConfigPath} }

// EnvironmentResolution enables or disables environment resolution on [LoadProject].
func EnvironmentResolution(v bool) ProjectFnOpt {
	if v {
		return ProjectFnOpt{fn: func(*composecli.ProjectOptions) error { return nil }}
	}
	return ProjectFnOpt{fn: composecli.WithoutEnvironmentResolution}
}

// LabelsResolution enables or disables label resolution on [LoadProject].
func LabelsResolution(v bool) ProjectFnOpt {
	if v {
		return ProjectFnOpt{fn: func(*composecli.ProjectOptions) error { return nil }}
	}
	return ProjectFnOpt{fn: composecli.WithoutLabelsResolution}
}

var (
	_ LoadProjectOption = Name("app")
	_ LoadProjectOption = File("compose.yaml")
	_ LoadProjectOption = ProjectDirectory(".")
	_ LoadProjectOption = EnvFile(".env")
	_ LoadProjectOption = Profile("dev")
	_ LoadProjectOption = Services("api")
	_ LoadProjectOption = AllResources(true)
	_ LoadProjectOption = Offline(true)
	_ LoadProjectOption = Compatibility(true)
	_ LoadProjectOption = InsecureRegistry("localhost:5000")
	_ LoadProjectOption = OnLoad(func(string, map[string]any) {})
	_ LoadProjectOption = OnUnsupported(func([]UnsupportedAttribute) {})
	_ LoadProjectOption = ProjectOption(nil)
	_ LoadProjectOption = Interpolation(false)
	_ LoadProjectOption = Normalization(false)
	_ LoadProjectOption = Consistency(false)
	_ LoadProjectOption = ResolvedPaths(true)
	_ LoadProjectOption = DiscardEnvFile()
	_ LoadProjectOption = OsEnv()
	_ LoadProjectOption = DotEnv()
	_ LoadProjectOption = ConfigFileEnv()
	_ LoadProjectOption = DefaultConfigPath()
	_ LoadProjectOption = EnvironmentResolution(false)
	_ LoadProjectOption = LabelsResolution(false)
)
