// Package kit defines Compose projects and runs them through the
// Docker Compose SDK.
package kit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/bridge"
)

// Compose runs a Project through the Docker Compose SDK.
type Compose struct {
	project *Project
	global  Global
}

// Global holds Compose-wide settings applied to every command on this runner.
type Global struct {
	Progress         string
	ANSI             string
	Parallel         *int
	AllResources     bool
	EnvFile          []string
	Compatibility    bool
	ProjectDirectory string
	DryRun           bool
}

// NewCompose returns a Compose runner for the given project.
func NewCompose(project *Project) *Compose {
	return &Compose{
		project: project,
	}
}

// Global sets parent settings used by every command on this runner.
func (c *Compose) Global(g Global) *Compose {
	if c == nil {
		return nil
	}
	c.global = g
	return c
}

func defaultWriter(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

// Events streams compose events for the project.
// A nil opt watches every service and enables no profiles.
func (c *Compose) Events(ctx context.Context, opt *EventsOpt) (<-chan Events, <-chan error, error) {
	if opt == nil {
		opt = &EventsOpt{}
	}
	var services []string
	if opt.Service != "" {
		services = []string{opt.Service}
	}
	project, err := c.prepare("events", opt.Profiles, nil, selectKeep)
	if err != nil {
		return nil, nil, errdefs.NewComposeEventsError(err)
	}
	svc, closeSvc, err := c.open(io.Discard, nil, "")
	if err != nil {
		return nil, nil, errdefs.NewComposeEventsError(err)
	}
	eventsCh := make(chan Events, 1)
	errCh := make(chan error, 1)
	go func() {
		defer close(eventsCh)
		defer close(errCh)
		defer closeSvc()
		err := svc.Events(ctx, project.Name, api.EventsOptions{
			Services: services,
			Since:    opt.Since,
			Until:    opt.Until,
			Consumer: func(event api.Event) error {
				mapped := Events{
					Time:       event.Timestamp.Format(time.RFC3339Nano),
					Type:       "container",
					Action:     event.Status,
					ID:         event.Container,
					Service:    event.Service,
					Attributes: event.Attributes,
					Status:     event.Status,
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case eventsCh <- mapped:
					return nil
				}
			},
		})
		if handled := handleContextCancellation(ctx, err); handled != nil {
			errCh <- errdefs.NewComposeEventsError(handled)
		}
	}()
	return eventsCh, errCh, nil
}

// Kill stops service containers immediately.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Kill(ctx context.Context, opt *Kill) error {
	if opt == nil {
		opt = &Kill{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("kill", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeKillError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeKillError, func(svc api.Compose) error {
		return svc.Kill(ctx, project.Name, api.KillOptions{
			RemoveOrphans: opt.RemoveOrphans,
			Project:       project,
			Services:      opt.ServiceNames,
			Signal:        deref(opt.Signal),
		})
	})
}

// Up creates and starts the project.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Up(ctx context.Context, opt *Up) error {
	if opt == nil {
		opt = &Up{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if err := opt.validate(); err != nil {
		return errdefs.NewComposeUpError(err)
	}
	mode := selectIncludeDeps
	if opt.NoDeps {
		mode = selectIgnoreDeps
	}
	project, err := c.prepare("up", opt.Profiles, opt.ServiceNames, mode)
	if err != nil {
		return errdefs.NewComposeUpError(err)
	}
	if opt.Build {
		applyPullPolicy(project, "build")
	} else {
		applyPullPolicy(project, deref(opt.Pull))
	}
	applyScale(project, opt.Scale)
	ignore, err := orphanMode(opt.IgnoreOrphans, opt.RemoveOrphans, project.Environment)
	if err != nil {
		return errdefs.NewComposeUpError(err)
	}
	create := opt.createOptions(opt.ServiceNames, ignore)
	if opt.NoStart {
		return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeUpError, func(svc api.Compose) error {
			return svc.Create(ctx, project, create)
		})
	}
	var consumer api.LogConsumer
	if !opt.Detach && !opt.Wait {
		consumer = logConsumer(ctx, opt.Writer, !opt.NoColor, !opt.NoLogPrefix, opt.Timestamps)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeUpError, func(svc api.Compose) error {
		return svc.Up(ctx, project, api.UpOptions{
			Create: create,
			Start: api.StartOptions{
				Project:        project,
				Attach:         consumer,
				AttachTo:       opt.attachTo(project.ServiceNames()),
				OnExit:         opt.onExit(),
				ExitCodeFrom:   deref(opt.ExitCodeFrom),
				Wait:           opt.Wait,
				WaitTimeout:    durationOrZero(opt.WaitTimeout),
				Services:       opt.ServiceNames,
				Watch:          opt.Watch,
				NavigationMenu: opt.Menu,
			},
		})
	})
}

// Down stops and removes project resources.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Down(ctx context.Context, opt *Down) error {
	if opt == nil {
		opt = &Down{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	images := ""
	if opt.RemoveImage != nil {
		images = *opt.RemoveImage
		switch images {
		case "", "local", "all":
		default:
			return errdefs.NewComposeDownError(fmt.Errorf("invalid image prune mode %q", images))
		}
	}
	project, err := c.prepare("down", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeDownError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeDownError, func(svc api.Compose) error {
		return svc.Down(ctx, project.Name, api.DownOptions{
			RemoveOrphans: opt.RemoveOrphans,
			Project:       project,
			Timeout:       seconds(opt.Timeout),
			Images:        images,
			Volumes:       opt.RemoveVolumes,
			Services:      opt.ServiceNames,
		})
	})
}

// Logs writes service logs to Writer.
// A nil opt uses defaults (writer is os.Stdout).
func (c *Compose) Logs(ctx context.Context, opt *Logs) error {
	if opt == nil {
		opt = &Logs{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("logs", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeLogsError(err)
	}
	tail := ""
	if opt.Tail != nil {
		tail = fmt.Sprintf("%d", *opt.Tail)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeLogsError, func(svc api.Compose) error {
		return svc.Logs(ctx, project.Name, logConsumer(ctx, opt.Writer, !opt.NoColor, !opt.NoLogPrefix, opt.Timestamps), api.LogOptions{
			Project:    project,
			Index:      indexOrZero(opt.Index),
			Services:   opt.ServiceNames,
			Tail:       tail,
			Since:      opt.Since,
			Until:      opt.Until,
			Follow:     opt.Follow,
			Timestamps: opt.Timestamps,
		})
	})
}

// Ps lists project containers.
func (c *Compose) Ps(ctx context.Context, opt *Ps) error {
	if opt == nil {
		opt = &Ps{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("ps", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposePsError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposePsError, func(svc api.Compose) error {
		rows, err := svc.Ps(ctx, project.Name, api.PsOptions{
			Project:  project,
			All:      opt.All,
			Services: opt.ServiceNames,
		})
		if err != nil {
			return err
		}
		return writePs(opt.Writer, rows, opt)
	})
}

// Start starts existing project containers.
func (c *Compose) Start(ctx context.Context, opt *Start) error {
	if opt == nil {
		opt = &Start{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("start", opt.Profiles, opt.ServiceNames, selectIncludeDeps)
	if err != nil {
		return errdefs.NewComposeStartError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeStartError, func(svc api.Compose) error {
		return svc.Start(ctx, project.Name, api.StartOptions{
			Project:     project,
			Wait:        opt.Wait,
			WaitTimeout: durationOrZero(opt.WaitTimeout),
			Services:    opt.ServiceNames,
		})
	})
}

// Stop stops project containers.
func (c *Compose) Stop(ctx context.Context, opt *Stop) error {
	if opt == nil {
		opt = &Stop{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("stop", opt.Profiles, nil, selectKeep)
	if err != nil {
		return errdefs.NewComposeStopError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeStopError, func(svc api.Compose) error {
		return svc.Stop(ctx, project.Name, api.StopOptions{
			Project:  project,
			Timeout:  seconds(opt.Timeout),
			Services: opt.ServiceNames,
		})
	})
}

// Restart restarts project containers.
func (c *Compose) Restart(ctx context.Context, opt *Restart) error {
	if opt == nil {
		opt = &Restart{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	mode := selectIncludeDeps
	if opt.NoDeps {
		mode = selectIgnoreDeps
	}
	project, err := c.prepare("restart", opt.Profiles, opt.ServiceNames, mode)
	if err != nil {
		return errdefs.NewComposeRestartError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeRestartError, func(svc api.Compose) error {
		return svc.Restart(ctx, project.Name, api.RestartOptions{
			Project:  project,
			Timeout:  seconds(opt.Timeout),
			Services: opt.ServiceNames,
			NoDeps:   opt.NoDeps,
		})
	})
}

// Build builds or rebuilds services.
func (c *Compose) Build(ctx context.Context, opt *ComposeBuild) error {
	if opt == nil {
		opt = &ComposeBuild{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	mode := selectIgnoreDeps
	if opt.WithDependencies {
		mode = selectIncludeDeps
	}
	project, err := c.prepare("build", opt.Profiles, opt.ServiceNames, mode)
	if err != nil {
		return errdefs.NewComposeBuildError(err)
	}
	options, err := opt.apiBuild(opt.ServiceNames)
	if err != nil {
		return errdefs.NewComposeBuildError(err)
	}
	progress := ""
	if opt.Quiet {
		progress = "quiet"
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeBuildError, func(svc api.Compose) error {
		return svc.Build(ctx, project, options)
	}, progress)
}

// Pull pulls service images.
func (c *Compose) Pull(ctx context.Context, opt *Pull) error {
	if opt == nil {
		opt = &Pull{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	mode := selectIgnoreDeps
	if opt.IncludeDeps {
		mode = selectIncludeDeps
	}
	project, err := c.prepare("pull", opt.Profiles, opt.ServiceNames, mode)
	if err != nil {
		return errdefs.NewComposePullError(err)
	}
	applyPullPolicy(project, opt.Policy)
	progress := ""
	if opt.Quiet {
		progress = "quiet"
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposePullError, func(svc api.Compose) error {
		return svc.Pull(ctx, project, api.PullOptions{
			Quiet:           opt.Quiet,
			IgnoreFailures:  opt.IgnorePullFailures,
			IgnoreBuildable: opt.IgnoreBuildable,
		})
	}, progress)
}

// Exec runs a command in a running service container.
// Service and Command must be set on opt.
func (c *Compose) Exec(ctx context.Context, opt *Exec) error {
	if opt == nil {
		opt = &Exec{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return errdefs.NewComposeExecError(fmt.Errorf("service is required"))
	}
	if len(opt.Command) == 0 {
		return errdefs.NewComposeExecError(fmt.Errorf("command is required"))
	}
	project, err := c.prepare("exec", opt.Profiles, []string{opt.Service}, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeExecError(err)
	}
	return callErr(c, ctx, opt.Writer, opt.Stdin, errdefs.NewComposeExecError, func(svc api.Compose) error {
		code, err := svc.Exec(ctx, project.Name, api.RunOptions{
			Project:     project,
			Service:     opt.Service,
			Command:     opt.Command,
			Environment: opt.Env,
			WorkingDir:  opt.Workdir,
			User:        opt.User,
			Privileged:  opt.Privileged,
			Index:       indexOrZero(opt.Index),
			Detach:      opt.Detach,
			Tty:         !opt.NoTTY,
			Interactive: !opt.NoTTY && opt.Stdin != nil,
		})
		if err != nil {
			return err
		}
		return exitStatus(code)
	})
}

// Attach attaches to a running service container. Service is required.
func (c *Compose) Attach(ctx context.Context, opt *Attach) error {
	if opt == nil {
		opt = &Attach{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return errdefs.NewComposeAttachError(fmt.Errorf("service is required"))
	}
	project, err := c.prepare("attach", opt.Profiles, []string{opt.Service}, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeAttachError(err)
	}
	in := opt.Stdin
	if opt.NoStdin {
		in = nil
	}
	return callErr(c, ctx, opt.Writer, in, errdefs.NewComposeAttachError, func(svc api.Compose) error {
		return svc.Attach(ctx, project.Name, api.AttachOptions{
			Project:    project,
			Service:    opt.Service,
			Index:      indexOrZero(opt.Index),
			DetachKeys: opt.DetachKeys,
			NoStdin:    opt.NoStdin,
			Proxy:      !opt.NoSigProxy,
		})
	})
}

// Commit creates an image from a service container. Service is required.
func (c *Compose) Commit(ctx context.Context, opt *Commit) error {
	if opt == nil {
		opt = &Commit{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return errdefs.NewComposeCommitError(fmt.Errorf("service is required"))
	}
	changes, err := commitChanges(opt.Change)
	if err != nil {
		return errdefs.NewComposeCommitError(err)
	}
	project, err := c.prepare("commit", opt.Profiles, []string{opt.Service}, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeCommitError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeCommitError, func(svc api.Compose) error {
		return svc.Commit(ctx, project.Name, api.CommitOptions{
			Service:   opt.Service,
			Reference: opt.Repository,
			Pause:     !opt.NoPause,
			Comment:   opt.Message,
			Author:    opt.Author,
			Changes:   changes,
			Index:     indexOrZero(opt.Index),
		})
	})
}

// Config renders the project model.
func (c *Compose) Config(ctx context.Context, opt *ComposeConfig) error {
	if opt == nil {
		opt = &ComposeConfig{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("config", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeConfigError(err)
	}
	if err := c.renderConfig(opt.Writer, project, opt); err != nil {
		return errdefs.NewComposeConfigError(err)
	}
	_ = ctx
	return nil
}

// Cp copies between a service container and the local filesystem.
// Source and Dest are required.
func (c *Compose) Cp(ctx context.Context, opt *Cp) error {
	if opt == nil {
		opt = &Cp{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Source == "" || opt.Dest == "" {
		return errdefs.NewComposeCpError(fmt.Errorf("source and dest are required"))
	}
	project, err := c.prepare("cp", opt.Profiles, nil, selectKeep)
	if err != nil {
		return errdefs.NewComposeCpError(err)
	}
	return callErr(c, ctx, opt.Writer, opt.Stdin, errdefs.NewComposeCpError, func(svc api.Compose) error {
		return svc.Copy(ctx, project.Name, api.CopyOptions{
			Source:      opt.Source,
			Destination: opt.Dest,
			All:         opt.All,
			Index:       indexOrZero(opt.Index),
			FollowLink:  opt.FollowLink,
			CopyUIDGID:  opt.Archive,
		})
	})
}

// Create creates containers for services without starting them.
func (c *Compose) Create(ctx context.Context, opt *Create) error {
	if opt == nil {
		opt = &Create{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("create", opt.Profiles, opt.ServiceNames, selectIncludeDeps)
	if err != nil {
		return errdefs.NewComposeCreateError(err)
	}
	if opt.Build {
		applyPullPolicy(project, "build")
	}
	applyPullPolicy(project, opt.Pull)
	applyScale(project, opt.Scale)
	ignore, err := orphanMode(opt.IgnoreOrphans, opt.RemoveOrphans, project.Environment)
	if err != nil {
		return errdefs.NewComposeCreateError(err)
	}
	recreate, deps := recreate(opt.ForceRecreate, opt.NoRecreate, false)
	var build *api.BuildOptions
	if !opt.NoBuild {
		build = &api.BuildOptions{Quiet: opt.QuietPull, Services: opt.ServiceNames}
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeCreateError, func(svc api.Compose) error {
		return svc.Create(ctx, project, api.CreateOptions{
			Build:                build,
			Services:             opt.ServiceNames,
			RemoveOrphans:        opt.RemoveOrphans,
			IgnoreOrphans:        ignore,
			Recreate:             recreate,
			RecreateDependencies: deps,
			Inherit:              true,
			QuietPull:            opt.QuietPull,
		})
	})
}

// Export writes a service container filesystem as a tar archive. Service is required.
func (c *Compose) Export(ctx context.Context, opt *ComposeExport) error {
	if opt == nil {
		opt = &ComposeExport{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return errdefs.NewComposeExportError(fmt.Errorf("service is required"))
	}
	project, err := c.prepare("export", opt.Profiles, []string{opt.Service}, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeExportError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeExportError, func(svc api.Compose) error {
		return svc.Export(ctx, project.Name, api.ExportOptions{
			Service: opt.Service,
			Index:   indexOrZero(opt.Index),
			Output:  opt.Output,
		})
	})
}

// Images lists images used by project containers.
func (c *Compose) Images(ctx context.Context, opt *Images) error {
	if opt == nil {
		opt = &Images{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("images", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeImagesError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeImagesError, func(svc api.Compose) error {
		images, err := svc.Images(ctx, project.Name, api.ImagesOptions{Services: opt.ServiceNames})
		if err != nil {
			return err
		}
		return writeImages(opt.Writer, images, opt.Format, opt.Quiet)
	})
}

// Ls lists Compose projects on the engine.
func (c *Compose) Ls(ctx context.Context, opt *Ls) error {
	if opt == nil {
		opt = &Ls{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if _, err := c.prepare("ls", opt.Profiles, nil, selectKeep); err != nil {
		return errdefs.NewComposeLsError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeLsError, func(svc api.Compose) error {
		stacks, err := svc.List(ctx, api.ListOptions{All: opt.All})
		if err != nil {
			return err
		}
		return writeStacks(opt.Writer, stacks, opt)
	})
}

// Pause pauses services.
func (c *Compose) Pause(ctx context.Context, opt *Pause) error {
	if opt == nil {
		opt = &Pause{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("pause", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposePauseError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposePauseError, func(svc api.Compose) error {
		return svc.Pause(ctx, project.Name, api.PauseOptions{Services: opt.ServiceNames, Project: project})
	})
}

// Port prints the public address for a private port. Service and PrivatePort are required.
func (c *Compose) Port(ctx context.Context, opt *Port) error {
	if opt == nil {
		opt = &Port{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return errdefs.NewComposePortError(fmt.Errorf("service is required"))
	}
	if opt.PrivatePort <= 0 {
		return errdefs.NewComposePortError(fmt.Errorf("private port is required"))
	}
	project, err := c.prepare("port", opt.Profiles, []string{opt.Service}, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposePortError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposePortError, func(svc api.Compose) error {
		publishers, err := svc.Ports(ctx, project.Name, opt.Service, uint16(opt.PrivatePort), api.PortOptions{
			Protocol: opt.Protocol,
			Index:    indexOrZero(opt.Index),
		})
		if err != nil {
			return err
		}
		return writePublishers(opt.Writer, publishers, opt.Protocol)
	})
}

// Publish publishes the application image. Repository is required.
func (c *Compose) Publish(ctx context.Context, opt *Publish) error {
	if opt == nil {
		opt = &Publish{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Repository == "" {
		return errdefs.NewComposePublishError(fmt.Errorf("repository is required"))
	}
	oci := api.OCIVersion("")
	switch opt.OCIVersion {
	case "":
	case "1.0":
		oci = api.OCIVersion1_0
	case "1.1":
		oci = api.OCIVersion1_1
	default:
		return errdefs.NewComposePublishError(fmt.Errorf("oci version %q is not supported", opt.OCIVersion))
	}
	project, err := c.prepare("publish", opt.Profiles, nil, selectKeep)
	if err != nil {
		return errdefs.NewComposePublishError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposePublishError, func(svc api.Compose) error {
		return svc.Publish(ctx, project, opt.Repository, api.PublishOptions{
			ResolveImageDigests: opt.ResolveImageDigests,
			InsecureRegistry:    opt.InsecureRegistry,
			Application:         opt.App,
			WithEnvironment:     opt.WithEnv,
			OCIVersion:          oci,
		})
	})
}

// Push pushes service images.
func (c *Compose) Push(ctx context.Context, opt *Push) error {
	if opt == nil {
		opt = &Push{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	mode := selectIgnoreDeps
	if opt.IncludeDeps {
		mode = selectIncludeDeps
	}
	project, err := c.prepare("push", opt.Profiles, opt.ServiceNames, mode)
	if err != nil {
		return errdefs.NewComposePushError(err)
	}
	progress := ""
	if opt.Quiet {
		progress = "quiet"
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposePushError, func(svc api.Compose) error {
		return svc.Push(ctx, project, api.PushOptions{
			Quiet:          opt.Quiet,
			IgnoreFailures: opt.IgnorePushFailures,
		})
	}, progress)
}

// Rm removes stopped service containers.
func (c *Compose) Rm(ctx context.Context, opt *Rm) error {
	if opt == nil {
		opt = &Rm{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("rm", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeRmError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeRmError, func(svc api.Compose) error {
		return svc.Remove(ctx, project.Name, api.RemoveOptions{
			Project:  project,
			Stop:     opt.Stop,
			Volumes:  opt.Volumes,
			Force:    opt.Force,
			Services: opt.ServiceNames,
		})
	})
}

// Run runs a one-off command on a service. Service is required.
func (c *Compose) Run(ctx context.Context, opt *Run) error {
	if opt == nil {
		opt = &Run{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Service == "" {
		return errdefs.NewComposeRunError(fmt.Errorf("service is required"))
	}
	mode := selectIncludeDeps
	if opt.NoDeps {
		mode = selectIgnoreDeps
	}
	project, err := c.prepare("run", opt.Profiles, []string{opt.Service}, mode)
	if err != nil {
		return errdefs.NewComposeRunError(err)
	}
	applyPullPolicy(project, opt.Pull)
	ignore, err := orphanMode(opt.IgnoreOrphans, opt.RemoveOrphans, project.Environment)
	if err != nil {
		return errdefs.NewComposeRunError(err)
	}
	labels, err := labelsFrom(opt.Label)
	if err != nil {
		return errdefs.NewComposeRunError(err)
	}
	env := append([]string{}, opt.Env...)
	for _, file := range opt.EnvFromFile {
		lines, err := envFileLines(file)
		if err != nil {
			return errdefs.NewComposeRunError(err)
		}
		env = append(env, lines...)
	}
	interactive := !opt.NoTTY
	if opt.Interactive != nil {
		interactive = *opt.Interactive
	}
	var entry []string
	if opt.Entrypoint != "" {
		entry = []string{opt.Entrypoint}
	}
	var build *api.BuildOptions
	if opt.Build {
		build = &api.BuildOptions{Quiet: opt.QuietBuild || opt.Quiet, Services: []string{opt.Service}, Deps: !opt.NoDeps}
	}
	progress := ""
	if opt.Quiet {
		progress = "quiet"
	}
	return callErr(c, ctx, opt.Writer, opt.Stdin, errdefs.NewComposeRunError, func(svc api.Compose) error {
		code, err := svc.RunOneOffContainer(ctx, project, api.RunOptions{
			CreateOptions: api.CreateOptions{
				Build:         build,
				RemoveOrphans: opt.RemoveOrphans,
				IgnoreOrphans: ignore,
				QuietPull:     opt.QuietPull,
			},
			Project:           project,
			Name:              opt.Name,
			Service:           opt.Service,
			Command:           opt.Command,
			Entrypoint:        entry,
			Detach:            opt.Detach,
			AutoRemove:        opt.Rm,
			Tty:               !opt.NoTTY,
			Interactive:       interactive,
			WorkingDir:        opt.Workdir,
			User:              opt.User,
			Environment:       env,
			CapAdd:            opt.CapAdd,
			CapDrop:           opt.CapDrop,
			Labels:            labels,
			UseNetworkAliases: opt.UseAliases,
			NoDeps:            opt.NoDeps,
		})
		if err != nil {
			return err
		}
		return exitStatus(code)
	}, progress)
}

// Scale sets the number of replicas for services. Replicas is required.
func (c *Compose) Scale(ctx context.Context, opt *Scale) error {
	if opt == nil {
		opt = &Scale{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if len(opt.Replicas) == 0 {
		return errdefs.NewComposeScaleError(fmt.Errorf("at least one service replica is required"))
	}
	replicas := make(map[string]int, len(opt.Replicas))
	names := make([]string, 0, len(opt.Replicas))
	for _, replica := range opt.Replicas {
		if replica.Service == "" || replica.Num < 0 {
			return errdefs.NewComposeScaleError(fmt.Errorf("service name required and scale must be non-negative"))
		}
		replicas[replica.Service] = replica.Num
		names = append(names, replica.Service)
	}
	mode := selectIncludeDeps
	if opt.NoDeps {
		mode = selectIgnoreDeps
	}
	project, err := c.prepare("scale", opt.Profiles, names, mode)
	if err != nil {
		return errdefs.NewComposeScaleError(err)
	}
	applyScale(project, opt.Replicas)
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeScaleError, func(svc api.Compose) error {
		return svc.Scale(ctx, project, api.ScaleOptions{Replicas: replicas, Services: names})
	})
}

// Stats writes container resource usage. The Compose SDK has no stats method,
// so this reads one-shot or streamed stats from the engine for project containers.
func (c *Compose) Stats(ctx context.Context, opt *Stats) error {
	if opt == nil {
		opt = &Stats{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("stats", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeStatsError(err)
	}
	if err := c.containerStats(ctx, project, opt, opt.Writer); err != nil {
		return errdefs.NewComposeStatsError(handleContextCancellation(ctx, err))
	}
	return nil
}

// Top lists processes in project containers.
func (c *Compose) Top(ctx context.Context, opt *Top) error {
	if opt == nil {
		opt = &Top{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("top", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeTopError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeTopError, func(svc api.Compose) error {
		rows, err := svc.Top(ctx, project.Name, opt.ServiceNames)
		if err != nil {
			return err
		}
		return writeTop(opt.Writer, rows)
	})
}

// Unpause unpauses services.
func (c *Compose) Unpause(ctx context.Context, opt *Unpause) error {
	if opt == nil {
		opt = &Unpause{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("unpause", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeUnpauseError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeUnpauseError, func(svc api.Compose) error {
		return svc.UnPause(ctx, project.Name, api.PauseOptions{Services: opt.ServiceNames, Project: project})
	})
}

// Version writes the linked Compose SDK module version.
func (c *Compose) Version(ctx context.Context, opt *Version) error {
	if opt == nil {
		opt = &Version{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	version := composeModuleVersion()
	var err error
	switch {
	case strings.EqualFold(opt.Format, "json"):
		_, err = fmt.Fprintf(opt.Writer, "{\"version\":%q}\n", version)
	case opt.Short:
		_, err = fmt.Fprintln(opt.Writer, version)
	default:
		_, err = fmt.Fprintf(opt.Writer, "Docker Compose version %s\n", version)
	}
	if err != nil {
		return errdefs.NewComposeVersionError(err)
	}
	_ = ctx
	return nil
}

// Volumes lists project volumes.
func (c *Compose) Volumes(ctx context.Context, opt *Volumes) error {
	if opt == nil {
		opt = &Volumes{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	project, err := c.prepare("volumes", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeVolumesError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeVolumesError, func(svc api.Compose) error {
		vols, err := svc.Volumes(ctx, project.Name, api.VolumesOptions{Services: opt.ServiceNames})
		if err != nil {
			return err
		}
		if strings.EqualFold(opt.Format, "json") {
			return writeJSON(opt.Writer, vols)
		}
		for _, vol := range vols {
			line := vol.Name
			if opt.Quiet {
				line = vol.Name
			}
			if _, err := fmt.Fprintln(opt.Writer, line); err != nil {
				return err
			}
		}
		return nil
	})
}

// Wait blocks until a service container stops. ServiceNames is required.
func (c *Compose) Wait(ctx context.Context, opt *Wait) error {
	if opt == nil {
		opt = &Wait{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if len(opt.ServiceNames) == 0 {
		return errdefs.NewComposeWaitError(fmt.Errorf("at least one service is required"))
	}
	project, err := c.prepare("wait", opt.Profiles, opt.ServiceNames, selectIgnoreDeps)
	if err != nil {
		return errdefs.NewComposeWaitError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeWaitError, func(svc api.Compose) error {
		code, err := svc.Wait(ctx, project.Name, api.WaitOptions{
			Services:                   opt.ServiceNames,
			DownProjectOnContainerExit: opt.DownProject,
		})
		if err != nil {
			return err
		}
		return exitStatus(int(code))
	})
}

// Watch rebuilds services when build-context files change.
func (c *Compose) Watch(ctx context.Context, opt *Watch) error {
	if opt == nil {
		opt = &Watch{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if !opt.NoUp {
		if err := c.Up(ctx, &Up{
			Detach:       true,
			Profiles:     opt.Profiles,
			ServiceNames: opt.ServiceNames,
			Writer:       opt.Writer,
			QuietPull:    opt.Quiet,
		}); err != nil {
			return errdefs.NewComposeWatchError(err)
		}
	}
	project, err := c.prepare("watch", opt.Profiles, opt.ServiceNames, selectIncludeDeps)
	if err != nil {
		return errdefs.NewComposeWatchError(err)
	}
	progress := ""
	if opt.Quiet {
		progress = "quiet"
	}
	build := api.BuildOptions{Quiet: opt.Quiet, Services: opt.ServiceNames}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeWatchError, func(svc api.Compose) error {
		return svc.Watch(ctx, project, api.WatchOptions{
			Build:    &build,
			Prune:    !opt.NoPrune,
			Services: opt.ServiceNames,
			LogTo:    logConsumer(ctx, opt.Writer, !opt.Quiet, true, false),
		})
	}, progress)
}

// LoadProject loads a Compose project from files and returns it.
// Global compatibility, env files, working directory, all-resources, and dry-run apply.
func (c *Compose) LoadProject(ctx context.Context, opt *ProjectLoad) (*Project, error) {
	if c == nil {
		return nil, errdefs.NewComposeLoadError(fmt.Errorf("compose runner is required"))
	}
	if opt == nil {
		opt = &ProjectLoad{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	options := c.projectLoadOptions(opt)
	var loaded *types.Project
	err := callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeLoadError, func(svc api.Compose) error {
		project, err := svc.LoadProject(ctx, options)
		if err != nil {
			return err
		}
		loaded = project
		return nil
	}, "quiet")
	if err != nil {
		return nil, err
	}
	return &Project{wrapped: loaded}, nil
}

// Viz writes a Graphviz graph of the project.
func (c *Compose) Viz(ctx context.Context, opt *Viz) error {
	if opt == nil {
		opt = &Viz{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	indent, err := vizIndentation(opt.IndentationSize, opt.Spaces)
	if err != nil {
		return errdefs.NewComposeVizError(err)
	}
	project, err := c.prepare("viz", opt.Profiles, nil, selectKeep)
	if err != nil {
		return errdefs.NewComposeVizError(err)
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeVizError, func(svc api.Compose) error {
		graph, err := svc.Viz(ctx, project, api.VizOptions{
			IncludeNetworks:  opt.Networks,
			IncludePorts:     opt.Ports,
			IncludeImageName: opt.Image,
			Indentation:      indent,
		})
		if err != nil {
			return err
		}
		_, err = io.WriteString(opt.Writer, graph)
		return err
	}, "quiet")
}

// Generate writes a Compose model built from existing containers.
func (c *Compose) Generate(ctx context.Context, opt *Generate) error {
	if opt == nil {
		opt = &Generate{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if len(opt.Containers) == 0 {
		return errdefs.NewComposeGenerateError(fmt.Errorf("at least one container is required"))
	}
	format := opt.Format
	if format == "" {
		format = "yaml"
	}
	if format != "yaml" && format != "json" {
		return errdefs.NewComposeGenerateError(fmt.Errorf("unsupported format %q", format))
	}
	return callErr(c, ctx, opt.Writer, nil, errdefs.NewComposeGenerateError, func(svc api.Compose) error {
		project, err := svc.Generate(ctx, api.GenerateOptions{
			ProjectName: opt.Name,
			Containers:  opt.Containers,
		})
		if err != nil {
			return err
		}
		if opt.ProjectDir != "" {
			project.WorkingDir = opt.ProjectDir
		}
		var content []byte
		switch format {
		case "json":
			content, err = project.MarshalJSON()
		default:
			content, err = project.MarshalYAML()
		}
		if err != nil {
			return err
		}
		if len(content) == 0 || content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
		_, err = opt.Writer.Write(content)
		return err
	}, "quiet")
}

// BridgeConvert converts the project with a Compose bridge transformation.
func (c *Compose) BridgeConvert(ctx context.Context, opt *BridgeConvert) error {
	if opt == nil {
		opt = &BridgeConvert{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Output == "" {
		opt.Output = "out"
	}
	project, err := c.prepare("bridge", opt.Profiles, nil, selectKeep)
	if err != nil {
		return errdefs.NewComposeBridgeConvertError(err)
	}
	cli, closeCLI, err := c.engineCLI(opt.Writer, nil)
	if err != nil {
		return errdefs.NewComposeBridgeConvertError(err)
	}
	defer closeCLI()
	convert := bridge.ConvertOptions{
		Output:          opt.Output,
		Templates:       opt.Templates,
		Transformations: opt.Transformations,
	}
	if opt.Yes {
		convert.Confirm = func(string, bool) (bool, error) { return true, nil }
	}
	if err := bridge.Convert(ctx, cli, project, convert); err != nil {
		return errdefs.NewComposeBridgeConvertError(err)
	}
	return nil
}

// BridgeTransformationsList writes installed Compose bridge transformations.
func (c *Compose) BridgeTransformationsList(ctx context.Context, opt *BridgeTransformationsList) error {
	if opt == nil {
		opt = &BridgeTransformationsList{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	format := opt.Format
	if format == "" {
		format = "table"
	}
	if format != "table" && format != "json" {
		return errdefs.NewComposeBridgeTransformationsListError(fmt.Errorf("unsupported format %q", format))
	}
	cli, closeCLI, err := c.engineCLI(opt.Writer, nil)
	if err != nil {
		return errdefs.NewComposeBridgeTransformationsListError(err)
	}
	defer closeCLI()
	images, err := bridge.ListTransformers(ctx, cli)
	if err != nil {
		return errdefs.NewComposeBridgeTransformationsListError(err)
	}
	if err := writeTransformers(opt.Writer, images, format, opt.Quiet); err != nil {
		return errdefs.NewComposeBridgeTransformationsListError(err)
	}
	return nil
}

// BridgeTransformationsCreate copies a transformation image into Dest.
func (c *Compose) BridgeTransformationsCreate(ctx context.Context, opt *BridgeTransformationsCreate) error {
	if opt == nil {
		opt = &BridgeTransformationsCreate{}
	}
	opt.Writer = defaultWriter(opt.Writer)
	if opt.Dest == "" {
		return errdefs.NewComposeBridgeTransformationsCreateError(fmt.Errorf("dest is required"))
	}
	cli, closeCLI, err := c.engineCLI(opt.Writer, nil)
	if err != nil {
		return errdefs.NewComposeBridgeTransformationsCreateError(err)
	}
	defer closeCLI()
	if err := bridge.CreateTransformer(ctx, cli, bridge.CreateTransformerOptions{
		Dest: opt.Dest,
		From: opt.From,
	}); err != nil {
		return errdefs.NewComposeBridgeTransformationsCreateError(err)
	}
	return nil
}

func callErr[E error](c *Compose, ctx context.Context, out io.Writer, in io.Reader, wrap func(error) E, fn func(api.Compose) error, progress ...string) error {
	mode := ""
	if len(progress) > 0 {
		mode = progress[0]
	}
	svc, closeSvc, err := c.open(out, in, mode)
	if err != nil {
		return wrap(err)
	}
	defer closeSvc()
	err = handleContextCancellation(ctx, fn(svc))
	if err == nil {
		return nil
	}
	return wrap(err)
}

func durationOrZero(n *int) time.Duration {
	if n == nil {
		return 0
	}
	return time.Duration(*n) * time.Second
}

func writeJSON(w io.Writer, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}

func handleContextCancellation(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	if ctx.Err() != nil {
		return nil
	}
	return err
}
