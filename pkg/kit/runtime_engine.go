package kit

import (
	"context"
	"fmt"
)

func (c *Container) engineAPI(api Engine, setting string) error {
	if api == nil {
		return cfgErr("container", "engine", setting, fmt.Errorf("%w: engine client", errRequired))
	}
	return nil
}

func (c *Container) requireID(id, setting string) error {
	if id == "" {
		return cfgErr("container", "id", setting, fmt.Errorf("%w: container id", errRequired))
	}
	return nil
}

// Create sends the container to the Engine. Name() overrides the create-time name.
func (c *Container) Create(ctx context.Context, api Engine, opts ...EngineCreateOption) (Created, error) {
	if err := c.engineAPI(api, "Create"); err != nil {
		return Created{}, err
	}
	var act engineCreateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("create.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineCreate(&act))
	}
	if err := join(errs...); err != nil {
		return Created{}, wrapReceiver("container", err)
	}
	cfg, host, netcfg, err := c.EngineConfig()
	if err != nil {
		return Created{}, err
	}
	name, err := c.createName(act.name)
	if err != nil {
		return Created{}, err
	}
	res, err := api.ContainerCreate(ctx, ContainerCreateOptions{
		Config:           cfg,
		HostConfig:       host,
		NetworkingConfig: netcfg,
		Name:             name,
	})
	if err != nil {
		return Created{}, cfgErr("container", name, "Create", err)
	}
	return Created{ID: res.ID, Warnings: res.Warnings}, nil
}

func (c *Container) createName(override string) (string, error) {
	if override != "" {
		if err := validateName(override); err != nil {
			return "", wrapReceiver("container", wrapPath("create.name", err))
		}
		return override, nil
	}
	if c != nil && c.name.ok() {
		return c.name.val, nil
	}
	return "", cfgErr("container", "name", "Name", errRequired)
}

// CreateName returns the Engine create name from the container configuration.
func (c *Container) CreateName() (string, error) {
	return c.createName("")
}

// Start starts an Engine container by ID.
func (c *Container) Start(ctx context.Context, api Engine, id string, opts ...EngineStartOption) error {
	if err := c.engineAPI(api, "Start"); err != nil {
		return err
	}
	if err := c.requireID(id, "Start"); err != nil {
		return err
	}
	var act engineStartAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("start.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineStart(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("container", err)
	}
	if _, err := api.ContainerStart(ctx, id, act.sdk()); err != nil {
		return cfgErr("container", id, "Start", err)
	}
	return nil
}

// Stop stops an Engine container by ID.
func (c *Container) Stop(ctx context.Context, api Engine, id string, opts ...EngineStopOption) error {
	if err := c.engineAPI(api, "Stop"); err != nil {
		return err
	}
	if err := c.requireID(id, "Stop"); err != nil {
		return err
	}
	var act engineStopAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("stop.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineStop(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("container", err)
	}
	if _, err := api.ContainerStop(ctx, id, act.stopSDK()); err != nil {
		return cfgErr("container", id, "Stop", err)
	}
	return nil
}

// Restart restarts an Engine container by ID.
func (c *Container) Restart(ctx context.Context, api Engine, id string, opts ...EngineStopOption) error {
	if err := c.engineAPI(api, "Restart"); err != nil {
		return err
	}
	if err := c.requireID(id, "Restart"); err != nil {
		return err
	}
	var act engineStopAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("restart.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineStop(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("container", err)
	}
	if _, err := api.ContainerRestart(ctx, id, act.restartSDK()); err != nil {
		return cfgErr("container", id, "Restart", err)
	}
	return nil
}

// Remove removes an Engine container by ID.
func (c *Container) Remove(ctx context.Context, api Engine, id string, opts ...RemoveOption) error {
	if err := c.engineAPI(api, "Remove"); err != nil {
		return err
	}
	if err := c.requireID(id, "Remove"); err != nil {
		return err
	}
	var act removeAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("remove.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyRemove(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("container", err)
	}
	if _, err := api.ContainerRemove(ctx, id, act.engineSDK()); err != nil {
		return cfgErr("container", id, "Remove", err)
	}
	return nil
}

// Kill sends a signal to an Engine container by ID.
func (c *Container) Kill(ctx context.Context, api Engine, id string, opts ...KillOption) error {
	if err := c.engineAPI(api, "Kill"); err != nil {
		return err
	}
	if err := c.requireID(id, "Kill"); err != nil {
		return err
	}
	var act killAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("kill.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyKill(&act))
	}
	if err := join(errs...); err != nil {
		return wrapReceiver("container", err)
	}
	if _, err := api.ContainerKill(ctx, id, act.engineSDK()); err != nil {
		return cfgErr("container", id, "Kill", err)
	}
	return nil
}

// Pause pauses an Engine container by ID.
func (c *Container) Pause(ctx context.Context, api Engine, id string) error {
	if err := c.engineAPI(api, "Pause"); err != nil {
		return err
	}
	if err := c.requireID(id, "Pause"); err != nil {
		return err
	}
	if _, err := api.ContainerPause(ctx, id, ContainerPauseOptions{}); err != nil {
		return cfgErr("container", id, "Pause", err)
	}
	return nil
}

// Unpause unpauses an Engine container by ID.
func (c *Container) Unpause(ctx context.Context, api Engine, id string) error {
	if err := c.engineAPI(api, "Unpause"); err != nil {
		return err
	}
	if err := c.requireID(id, "Unpause"); err != nil {
		return err
	}
	if _, err := api.ContainerUnpause(ctx, id, ContainerUnpauseOptions{}); err != nil {
		return cfgErr("container", id, "Unpause", err)
	}
	return nil
}

// Inspect returns Engine inspect data for a container ID.
func (c *Container) Inspect(ctx context.Context, api Engine, id string, opts ...EngineInspectOption) (ContainerInspectResult, error) {
	if err := c.engineAPI(api, "Inspect"); err != nil {
		return ContainerInspectResult{}, err
	}
	if err := c.requireID(id, "Inspect"); err != nil {
		return ContainerInspectResult{}, err
	}
	var act engineInspectAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("inspect.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineInspect(&act))
	}
	if err := join(errs...); err != nil {
		return ContainerInspectResult{}, wrapReceiver("container", err)
	}
	res, err := api.ContainerInspect(ctx, id, act.sdk())
	if err != nil {
		return ContainerInspectResult{}, cfgErr("container", id, "Inspect", err)
	}
	return res, nil
}

// Logs returns Engine logs for a container ID.
func (c *Container) Logs(ctx context.Context, api Engine, id string, opts ...LogsOption) (ContainerLogsResult, error) {
	if err := c.engineAPI(api, "Logs"); err != nil {
		return nil, err
	}
	if err := c.requireID(id, "Logs"); err != nil {
		return nil, err
	}
	var act logsAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("logs.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyLogs(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("container", err)
	}
	res, err := api.ContainerLogs(ctx, id, act.engineSDK())
	if err != nil {
		return nil, cfgErr("container", id, "Logs", err)
	}
	return res, nil
}

// Wait waits for an Engine container to reach a condition.
func (c *Container) Wait(ctx context.Context, api Engine, id string, opts ...EngineWaitOption) ContainerWaitResult {
	if err := c.engineAPI(api, "Wait"); err != nil {
		return waitFailed(err)
	}
	if err := c.requireID(id, "Wait"); err != nil {
		return waitFailed(err)
	}
	var act engineWaitAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("wait.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineWait(&act))
	}
	if err := join(errs...); err != nil {
		return waitFailed(wrapReceiver("container", err))
	}
	return api.ContainerWait(ctx, id, act.sdk())
}
