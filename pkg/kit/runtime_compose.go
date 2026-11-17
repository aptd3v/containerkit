package kit

import (
	"context"
	"fmt"
	"strings"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

const (
	composeProjectLabel = "com.docker.compose.project"
	composeServiceLabel = "com.docker.compose.service"
	composeOneoffLabel  = "com.docker.compose.oneoff"
)

// stampComposeIdentity writes the labels Compose uses to find containers
// after create. The Compose CLI adds these while loading a file; a project
// built in kit does not, so start, kill, and down then report that the
// project has no containers and the next create hits a name conflict.
func stampComposeIdentity(project *composetypes.Project) {
	if project == nil {
		return
	}
	for name, svc := range project.Services {
		if svc.CustomLabels == nil {
			svc.CustomLabels = composetypes.Labels{}
		}
		setLabel(svc.CustomLabels, composeProjectLabel, project.Name)
		setLabel(svc.CustomLabels, composeServiceLabel, name)
		setLabel(svc.CustomLabels, composeOneoffLabel, "False")
		if project.WorkingDir != "" {
			setLabel(svc.CustomLabels, "com.docker.compose.project.working_dir", project.WorkingDir)
		}
		if len(project.ComposeFiles) > 0 {
			setLabel(svc.CustomLabels, "com.docker.compose.project.config_files", strings.Join(project.ComposeFiles, ","))
		}
		project.Services[name] = svc
	}
}

func setLabel(labels composetypes.Labels, key, value string) {
	if _, ok := labels[key]; ok {
		return
	}
	labels[key] = value
}

func (p *Compose) composeAPI(api Composer, setting string) error {
	if api == nil {
		return cfgErr("compose", setting, setting, fmt.Errorf("%w: compose client", errRequired))
	}
	return nil
}

func (p *Compose) loaded(api Composer, setting string) (*composetypes.Project, error) {
	if err := p.composeAPI(api, setting); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p.Project()
}

func upProject(ctx context.Context, api Composer, proj *composetypes.Project, name string, opts ...UpOption) error {
	if api == nil {
		return cfgErr("compose", name, "Up", fmt.Errorf("%w: compose client", errRequired))
	}
	var act upAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("up.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyUp(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	if log := adoptLog(ctx, &act.log); log != nil && act.start.consumer == nil && !act.detached {
		act.start.consumer = log
	}
	stampComposeIdentity(proj)
	return logAround(ctx, act.log, KindCompose, "up", name, func(ctx context.Context) error {
		if err := api.Up(ctx, proj, act.sdk()); err != nil {
			return cfgErr("compose", name, "Up", err)
		}
		return nil
	})
}

// Up brings this Compose file up through a [Composer]. It validates first.
func (p *Compose) Up(ctx context.Context, api Composer, opts ...UpOption) error {
	proj, err := p.loaded(api, "up")
	if err != nil {
		return err
	}
	return upProject(ctx, api, proj, p.Name(), opts...)
}

// Down tears this Compose file down through a [Composer].
func (p *Compose) Down(ctx context.Context, api Composer, opts ...DownOption) error {
	if err := p.composeAPI(api, "down"); err != nil {
		return err
	}
	var act downAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("down.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyDown(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	return logAround(ctx, act.log, KindCompose, "down", p.Name(), func(ctx context.Context) error {
		if err := api.Down(ctx, p.Name(), act.sdk()); err != nil {
			return cfgErr("compose", p.Name(), "Down", err)
		}
		return nil
	})
}

// Build builds images for this Compose file.
func (p *Compose) Build(ctx context.Context, api Composer, opts ...ComposeBuildOption) error {
	proj, err := p.loaded(api, "build")
	if err != nil {
		return err
	}
	var act composeBuildAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("build.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposeBuild(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	adoptLog(ctx, &act.log)
	stampComposeIdentity(proj)
	return logAround(ctx, act.log, KindBuild, "build", p.Name(), func(ctx context.Context) error {
		if err := api.Build(ctx, proj, act.sdk()); err != nil {
			return cfgErr("compose", p.Name(), "Build", err)
		}
		return nil
	})
}

// Create creates containers for this Compose file.
func (p *Compose) Create(ctx context.Context, api Composer, opts ...ComposeCreateOption) error {
	proj, err := p.loaded(api, "create")
	if err != nil {
		return err
	}
	var act createAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("create.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposeCreate(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	return logAround(ctx, act.log, KindCompose, "create", p.Name(), func(ctx context.Context) error {
		if err := api.Create(ctx, proj, act.sdk()); err != nil {
			return cfgErr("compose", "create", "Create", err)
		}
		return nil
	})
}

// Start starts services in this Compose file.
func (p *Compose) Start(ctx context.Context, api Composer, opts ...ComposeStartOption) error {
	if err := p.composeAPI(api, "start"); err != nil {
		return err
	}
	var act startAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("start.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposeStart(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	if log := adoptLog(ctx, &act.log); log != nil && act.consumer == nil {
		act.consumer = log
	}
	return logAround(ctx, act.log, KindCompose, "start", p.Name(), func(ctx context.Context) error {
		if err := api.Start(ctx, p.Name(), act.sdk()); err != nil {
			return cfgErr("compose", "start", "Start", err)
		}
		return nil
	})
}

// Stop stops services in this Compose file.
func (p *Compose) Stop(ctx context.Context, api Composer, opts ...ComposeStopOption) error {
	if err := p.composeAPI(api, "stop"); err != nil {
		return err
	}
	var act stopAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("stop.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposeStop(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	return logAround(ctx, act.log, KindCompose, "stop", p.Name(), func(ctx context.Context) error {
		if err := api.Stop(ctx, p.Name(), act.sdk()); err != nil {
			return cfgErr("compose", "stop", "Stop", err)
		}
		return nil
	})
}

// Restart restarts services in this Compose file.
func (p *Compose) Restart(ctx context.Context, api Composer, opts ...ComposeRestartOption) error {
	if err := p.composeAPI(api, "restart"); err != nil {
		return err
	}
	var act restartAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("restart.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposeRestart(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	return logAround(ctx, act.log, KindCompose, "restart", p.Name(), func(ctx context.Context) error {
		if err := api.Restart(ctx, p.Name(), act.sdk()); err != nil {
			return cfgErr("compose", "restart", "Restart", err)
		}
		return nil
	})
}

// Pull pulls images for this Compose file.
func (p *Compose) Pull(ctx context.Context, api Composer, opts ...PullOption) error {
	proj, err := p.loaded(api, "pull")
	if err != nil {
		return err
	}
	var act pullAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("pull.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyPull(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	return logAround(ctx, act.log, KindPull, "pull", p.Name(), func(ctx context.Context) error {
		if err := api.Pull(ctx, proj, act.sdk()); err != nil {
			return cfgErr("compose", p.Name(), "Pull", err)
		}
		return nil
	})
}

// Push pushes images for this Compose file.
func (p *Compose) Push(ctx context.Context, api Composer, opts ...PushOption) error {
	proj, err := p.loaded(api, "push")
	if err != nil {
		return err
	}
	var act pushAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("push.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyPush(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	return logAround(ctx, act.log, KindPush, "push", p.Name(), func(ctx context.Context) error {
		if err := api.Push(ctx, proj, act.sdk()); err != nil {
			return cfgErr("compose", p.Name(), "Push", err)
		}
		return nil
	})
}

// Logs streams service logs.
// A nil consumer uses Log or the logger on ctx. It is an error when neither is set.
// A consumer other than that logger is teed, so both receive each line.
func (p *Compose) Logs(ctx context.Context, api Composer, consumer composeapi.LogConsumer, opts ...LogsOption) error {
	if err := p.composeAPI(api, "logs"); err != nil {
		return err
	}
	var act logsAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("logs.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyLogs(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	adoptLog(ctx, &act.log)
	if consumer == nil && act.log != nil {
		consumer = act.log
	}
	if consumer == nil {
		return cfgErr("compose", "logs", "Logs", fmt.Errorf("%w: log consumer", errRequired))
	}
	if act.log != nil && consumer != act.log {
		consumer = teeLogConsumer{first: consumer, second: act.log}
	}
	proj, err := p.Project()
	if err != nil {
		proj = nil
	}
	if err := api.Logs(ctx, p.Name(), consumer, act.composeSDK(proj)); err != nil {
		return cfgErr("compose", "logs", "Logs", err)
	}
	return nil
}

// Ps lists containers for this Compose file.
func (p *Compose) Ps(ctx context.Context, api Composer, opts ...PsOption) ([]composeapi.ContainerSummary, error) {
	if err := p.composeAPI(api, "ps"); err != nil {
		return nil, err
	}
	var act psAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("ps.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyPs(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	out, err := api.Ps(ctx, p.Name(), act.sdk(proj))
	if err != nil {
		return nil, cfgErr("compose", "ps", "Ps", err)
	}
	return out, nil
}

// Kill sends a signal to services in this Compose file.
func (p *Compose) Kill(ctx context.Context, api Composer, opts ...KillOption) error {
	if err := p.composeAPI(api, "kill"); err != nil {
		return err
	}
	var act killAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("kill.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyKill(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	if err := api.Kill(ctx, p.Name(), act.composeSDK(proj)); err != nil {
		return cfgErr("compose", "kill", "Kill", err)
	}
	return nil
}

// Remove removes stopped service containers.
func (p *Compose) Remove(ctx context.Context, api Composer, opts ...RemoveOption) error {
	if err := p.composeAPI(api, "remove"); err != nil {
		return err
	}
	var act removeAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("remove.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyRemove(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	if err := api.Remove(ctx, p.Name(), act.composeSDK(proj)); err != nil {
		return cfgErr("compose", "remove", "Remove", err)
	}
	return nil
}

// Run runs a one-off command on a service.
func (p *Compose) Run(ctx context.Context, api Composer, service string, opts ...RunOption) (int, error) {
	proj, err := p.loaded(api, "run")
	if err != nil {
		return 0, err
	}
	var act runAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("run.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyRun(&act))
	}
	if err := join(errs...); err != nil {
		return 0, wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	code, err := api.RunOneOffContainer(ctx, proj, act.sdk(proj, service))
	if err != nil {
		return 0, cfgErr("compose", "run", "Run", err)
	}
	return code, nil
}

// Exec runs a command in a running service container.
func (p *Compose) Exec(ctx context.Context, api Composer, service string, opts ...RunOption) (int, error) {
	if err := p.composeAPI(api, "exec"); err != nil {
		return 0, err
	}
	var act runAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("exec.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyRun(&act))
	}
	if err := join(errs...); err != nil {
		return 0, wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	code, err := api.Exec(ctx, p.Name(), act.sdk(proj, service))
	if err != nil {
		return 0, cfgErr("compose", "exec", "Exec", err)
	}
	return code, nil
}

// Attach attaches to a running service container.
func (p *Compose) Attach(ctx context.Context, api Composer, service string, opts ...AttachOption) error {
	if err := p.composeAPI(api, "attach"); err != nil {
		return err
	}
	var act attachAct
	act.service = service
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("attach.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyAttach(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	if err := api.Attach(ctx, p.Name(), act.sdk(proj)); err != nil {
		return cfgErr("compose", "attach", "Attach", err)
	}
	return nil
}

// Copy copies files between a service container and the local filesystem.
func (p *Compose) Copy(ctx context.Context, api Composer, opts ...CopyOption) error {
	if err := p.composeAPI(api, "copy"); err != nil {
		return err
	}
	var act copyAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("copy.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyCopy(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	if err := api.Copy(ctx, p.Name(), act.sdk()); err != nil {
		return cfgErr("compose", "copy", "Copy", err)
	}
	return nil
}

// Pause pauses services.
func (p *Compose) Pause(ctx context.Context, api Composer, opts ...PauseOption) error {
	if err := p.composeAPI(api, "pause"); err != nil {
		return err
	}
	var act pauseAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("pause.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyPause(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	if err := api.Pause(ctx, p.Name(), act.sdk(proj)); err != nil {
		return cfgErr("compose", "pause", "Pause", err)
	}
	return nil
}

// Unpause unpauses services.
func (p *Compose) Unpause(ctx context.Context, api Composer, opts ...PauseOption) error {
	if err := p.composeAPI(api, "unpause"); err != nil {
		return err
	}
	var act pauseAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("unpause.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyPause(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	proj, _ := p.Project()
	if err := api.UnPause(ctx, p.Name(), act.sdk(proj)); err != nil {
		return cfgErr("compose", "unpause", "Unpause", err)
	}
	return nil
}

// Top shows running processes.
func (p *Compose) Top(ctx context.Context, api Composer, services ...string) ([]composeapi.ContainerProcSummary, error) {
	if err := p.composeAPI(api, "top"); err != nil {
		return nil, err
	}
	out, err := api.Top(ctx, p.Name(), services)
	if err != nil {
		return nil, cfgErr("compose", "top", "Top", err)
	}
	return out, nil
}

// Events streams service events.
func (p *Compose) Events(ctx context.Context, api Composer, opts ...EventsOption) error {
	if err := p.composeAPI(api, "events"); err != nil {
		return err
	}
	var act eventsAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("events.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEvents(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	if err := api.Events(ctx, p.Name(), act.sdk()); err != nil {
		return cfgErr("compose", "events", "Events", err)
	}
	return nil
}

// Port looks up a published port for a service.
func (p *Compose) Port(ctx context.Context, api Composer, service string, port uint16, opts ...ComposePortOption) (string, int, error) {
	if err := p.composeAPI(api, "port"); err != nil {
		return "", 0, err
	}
	var act composePortAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("port.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyComposePort(&act))
	}
	if err := join(errs...); err != nil {
		return "", 0, wrapReceiver("compose", err)
	}
	pubs, err := api.Ports(ctx, p.Name(), service, port, act.sdk())
	if err != nil {
		return "", 0, cfgErr("compose", "port", "Port", err)
	}
	if len(pubs) == 0 {
		return "", 0, nil
	}
	return pubs[0].URL, pubs[0].PublishedPort, nil
}

// PublishOCI publishes the Compose application as an OCI artifact.
// Port mappings use the [Publish] setter, not this method.
func (p *Compose) PublishOCI(ctx context.Context, api Composer, repository string, opts ...PublishOption) error {
	proj, err := p.loaded(api, "publish")
	if err != nil {
		return err
	}
	var act publishAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("publish.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyPublish(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	if err := api.Publish(ctx, proj, repository, act.sdk()); err != nil {
		return cfgErr("compose", "publish", "Publish", err)
	}
	return nil
}

// Images lists images used by this Compose file.
func (p *Compose) Images(ctx context.Context, api Composer, opts ...ImagesOption) (map[string]composeapi.ImageSummary, error) {
	if err := p.composeAPI(api, "images"); err != nil {
		return nil, err
	}
	var act imagesAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("images.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyImages(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	out, err := api.Images(ctx, p.Name(), act.sdk())
	if err != nil {
		return nil, cfgErr("compose", "images", "Images", err)
	}
	return out, nil
}

// Watch watches development context and rebuilds/syncs on change.
func (p *Compose) Watch(ctx context.Context, api Composer, opts ...WatchOption) error {
	proj, err := p.loaded(api, "watch")
	if err != nil {
		return err
	}
	var act watchAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("watch.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyWatch(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	adoptLog(ctx, &act.log)
	stampComposeIdentity(proj)
	return logAround(ctx, act.log, KindCompose, "watch", p.Name(), func(ctx context.Context) error {
		if err := api.Watch(ctx, proj, act.sdk()); err != nil {
			return cfgErr("compose", "watch", "Watch", err)
		}
		return nil
	})
}

// Viz returns a graphviz graph of services.
func (p *Compose) Viz(ctx context.Context, api Composer, opts ...VizOption) (string, error) {
	proj, err := p.loaded(api, "viz")
	if err != nil {
		return "", err
	}
	var act vizAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("viz.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyViz(&act))
	}
	if err := join(errs...); err != nil {
		return "", wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	out, err := api.Viz(ctx, proj, act.sdk())
	if err != nil {
		return "", cfgErr("compose", "viz", "Viz", err)
	}
	return out, nil
}

// Wait blocks until a service container exits.
func (p *Compose) Wait(ctx context.Context, api Composer, opts ...WaitOption) (int64, error) {
	if err := p.composeAPI(api, "wait"); err != nil {
		return 0, err
	}
	var act waitAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("wait.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyWait(&act))
	}
	if err := join(errs...); err != nil {
		return 0, wrapReceiver("compose", err)
	}
	code, err := api.Wait(ctx, p.Name(), act.sdk())
	if err != nil {
		return 0, cfgErr("compose", "wait", "Wait", err)
	}
	return code, nil
}

// Scale changes replica counts.
func (p *Compose) Scale(ctx context.Context, api Composer, opts ...ScaleOption) error {
	proj, err := p.loaded(api, "scale")
	if err != nil {
		return err
	}
	var act scaleAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("scale.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyScale(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	stampComposeIdentity(proj)
	if err := api.Scale(ctx, proj, act.sdk()); err != nil {
		return cfgErr("compose", "scale", "Scale", err)
	}
	return nil
}

// Export writes a service container filesystem as a tar archive.
func (p *Compose) Export(ctx context.Context, api Composer, service string, opts ...ExportOption) error {
	if err := p.composeAPI(api, "export"); err != nil {
		return err
	}
	var act exportAct
	act.service = service
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("export.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyExport(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	if err := api.Export(ctx, p.Name(), act.sdk()); err != nil {
		return cfgErr("compose", "export", "Export", err)
	}
	return nil
}

// Commit creates an image from a service container.
func (p *Compose) Commit(ctx context.Context, api Composer, service string, opts ...CommitOption) error {
	if err := p.composeAPI(api, "commit"); err != nil {
		return err
	}
	var act commitAct
	act.service = service
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("commit.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyCommit(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("compose", err)
	}
	if err := api.Commit(ctx, p.Name(), act.sdk()); err != nil {
		return cfgErr("compose", "commit", "Commit", err)
	}
	return nil
}

// VolumeList lists volumes used by this Compose file.
func (p *Compose) VolumeList(ctx context.Context, api Composer, opts ...VolumesListOption) ([]composeapi.VolumesSummary, error) {
	if err := p.composeAPI(api, "volumes"); err != nil {
		return nil, err
	}
	var act volumesListAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("volumes.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyVolumesList(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	out, err := api.Volumes(ctx, p.Name(), act.sdk())
	if err != nil {
		return nil, cfgErr("compose", "volumes", "VolumeList", err)
	}
	return out, nil
}

// Generate builds a Compose project from existing Engine containers.
func Generate(ctx context.Context, api Composer, opts ...GenerateOption) (*composetypes.Project, error) {
	if api == nil {
		return nil, cfgErr("compose", "generate", "Generate", fmt.Errorf("%w: compose client", errRequired))
	}
	var act generateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("generate.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyGenerate(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	out, err := api.Generate(ctx, act.sdk())
	if err != nil {
		return nil, cfgErr("compose", "generate", "Generate", err)
	}
	return out, nil
}

// MaxConcurrency limits concurrent Engine operations when the Composer
// implementation supports it. The Compose v5 SDK sets concurrency at
// construction time; this is a no-op on that service.
func (p *Compose) MaxConcurrency(api Composer, n int) error {
	if err := p.composeAPI(api, "maxConcurrency"); err != nil {
		return err
	}
	if s, ok := api.(interface{ MaxConcurrency(int) }); ok {
		s.MaxConcurrency(n)
	}
	return nil
}

// DryRunMode enables or disables dry-run when the Composer implementation
// supports it. The Compose v5 SDK selects dry-run at construction time.
func (p *Compose) DryRunMode(ctx context.Context, api Composer, dry bool) (context.Context, error) {
	if err := p.composeAPI(api, "dryRun"); err != nil {
		return ctx, err
	}
	if s, ok := api.(interface {
		DryRunMode(context.Context, bool) (context.Context, error)
	}); ok {
		out, err := s.DryRunMode(ctx, dry)
		if err != nil {
			return ctx, cfgErr("compose", "dryRun", "DryRunMode", err)
		}
		return out, nil
	}
	return ctx, nil
}

// List lists Compose stacks.
func List(ctx context.Context, api Composer, opts ...ListOption) ([]composeapi.Stack, error) {
	if api == nil {
		return nil, cfgErr("compose", "list", "List", fmt.Errorf("%w: compose client", errRequired))
	}
	var act listAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("compose", fmt.Sprintf("list.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyList(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("compose", err)
	}
	out, err := api.List(ctx, act.sdk())
	if err != nil {
		return nil, cfgErr("compose", "list", "List", err)
	}
	return out, nil
}
