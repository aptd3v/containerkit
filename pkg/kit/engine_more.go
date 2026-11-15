package kit

import (
	"context"
	"fmt"
	"io"
)

// EnginePluginCreateOption configures [CreatePlugin].
type EnginePluginCreateOption interface {
	applyEnginePluginCreateOption(*enginePluginCreateAct) error
}

type enginePluginCreateAct struct {
	in   io.Reader
	opts PluginCreateOptions
}

func (o ContentOpt) applyEnginePluginCreateOption(a *enginePluginCreateAct) error {
	a.in = o.r
	return nil
}
func (o NameOpt) applyEnginePluginCreateOption(a *enginePluginCreateAct) error {
	a.opts.RepoName = o.name
	return nil
}

// CreatePlugin wraps Engine.PluginCreate.
func CreatePlugin(ctx context.Context, api Engine, opts ...EnginePluginCreateOption) (PluginCreateResult, error) {
	if err := engineAPI(api, "CreatePlugin"); err != nil {
		return PluginCreateResult{}, err
	}
	var act enginePluginCreateAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("createplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginCreateOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginCreate(ctx, act.in, act.opts)
	if err != nil {
		return PluginCreateResult{}, cfgErr("engine", "createplugin", "CreatePlugin", err)
	}
	return out, nil
}

// EnginePluginInstallOption configures [InstallPlugin].
type EnginePluginInstallOption interface {
	applyEnginePluginInstallOption(*enginePluginInstallAct) error
}

type enginePluginInstallAct struct{ opts PluginInstallOptions }

func (o RegistryAuthOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.RegistryAuth = o.s
	return nil
}
func (o NameOpt) applyEnginePluginInstallOption(a *enginePluginInstallAct) error {
	a.opts.RemoteRef = o.name
	return nil
}

// InstallPlugin wraps Engine.PluginInstall.
func InstallPlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginInstallOption) (PluginInstallResult, error) {
	if err := engineAPI(api, "InstallPlugin"); err != nil {
		return PluginInstallResult{}, err
	}
	var act enginePluginInstallAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("installplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginInstallOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginInstallResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginInstall(ctx, name, act.opts)
	if err != nil {
		return PluginInstallResult{}, cfgErr("engine", "installplugin", "InstallPlugin", err)
	}
	return out, nil
}

// EnginePluginPushOption configures [PushPlugin].
type EnginePluginPushOption interface {
	applyEnginePluginPushOption(*enginePluginPushAct) error
}

type enginePluginPushAct struct{ opts PluginPushOptions }

func (o RegistryAuthOpt) applyEnginePluginPushOption(a *enginePluginPushAct) error {
	a.opts.RegistryAuth = o.s
	return nil
}

// PushPlugin wraps Engine.PluginPush.
func PushPlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginPushOption) (PluginPushResult, error) {
	if err := engineAPI(api, "PushPlugin"); err != nil {
		return PluginPushResult{}, err
	}
	var act enginePluginPushAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("pushplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginPushOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginPushResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginPush(ctx, name, act.opts)
	if err != nil {
		return PluginPushResult{}, cfgErr("engine", "pushplugin", "PushPlugin", err)
	}
	return out, nil
}

// EnginePluginSetOption configures [SetPlugin].
type EnginePluginSetOption interface {
	applyEnginePluginSetOption(*enginePluginSetAct) error
}

type enginePluginSetAct struct{ opts PluginSetOptions }

// ArgsOpt is the value returned by [Args].
// It implements [EnginePluginInstallOption], [EnginePluginSetOption], and
// [EnginePluginUpgradeOption].
type ArgsOpt struct{ args []string }

// Args sets plugin arguments for [SetPlugin], [InstallPlugin], and [UpgradePlugin].
func Args(args ...string) ArgsOpt { return ArgsOpt{args: args} }

func (o ArgsOpt) applyEnginePluginSetOption(a *enginePluginSetAct) error {
	a.opts.Args = append(a.opts.Args, o.args...)
	return nil
}

// SetPlugin wraps Engine.PluginSet.
func SetPlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginSetOption) (PluginSetResult, error) {
	if err := engineAPI(api, "SetPlugin"); err != nil {
		return PluginSetResult{}, err
	}
	var act enginePluginSetAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("setplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginSetOption(&act))
	}
	if err := join(errs...); err != nil {
		return PluginSetResult{}, wrapReceiver("engine", err)
	}
	out, err := api.PluginSet(ctx, name, act.opts)
	if err != nil {
		return PluginSetResult{}, cfgErr("engine", "setplugin", "SetPlugin", err)
	}
	return out, nil
}

// EnginePluginUpgradeOption configures [UpgradePlugin].
type EnginePluginUpgradeOption interface {
	applyEnginePluginUpgradeOption(*enginePluginUpgradeAct) error
}

type enginePluginUpgradeAct struct{ opts PluginUpgradeOptions }

func (o RegistryAuthOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.RegistryAuth = o.s
	return nil
}
func (o NameOpt) applyEnginePluginUpgradeOption(a *enginePluginUpgradeAct) error {
	a.opts.RemoteRef = o.name
	return nil
}

// UpgradePlugin wraps Engine.PluginUpgrade.
func UpgradePlugin(ctx context.Context, api Engine, name string, opts ...EnginePluginUpgradeOption) (PluginUpgradeResult, error) {
	if err := engineAPI(api, "UpgradePlugin"); err != nil {
		return nil, err
	}
	var act enginePluginUpgradeAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("upgradeplugin.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePluginUpgradeOption(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	out, err := api.PluginUpgrade(ctx, name, act.opts)
	if err != nil {
		return nil, cfgErr("engine", "upgradeplugin", "UpgradePlugin", err)
	}
	return out, nil
}

// EngineExecResizeOption configures [ResizeExec].
type EngineExecResizeOption interface {
	applyEngineExecResizeOption(*engineExecResizeAct) error
}

type engineExecResizeAct struct{ opts ExecResizeOptions }

func (o ConsoleSizeOpt) applyEngineExecResizeOption(a *engineExecResizeAct) error {
	a.opts.Height = o.h
	a.opts.Width = o.w
	return nil
}

// ResizeExec wraps Engine.ExecResize.
func ResizeExec(ctx context.Context, api Engine, id string, opts ...EngineExecResizeOption) (ExecResizeResult, error) {
	if err := engineAPI(api, "ResizeExec"); err != nil {
		return ExecResizeResult{}, err
	}
	var act engineExecResizeAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("resizeexec.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineExecResizeOption(&act))
	}
	if err := join(errs...); err != nil {
		return ExecResizeResult{}, wrapReceiver("engine", err)
	}
	return api.ExecResize(ctx, id, act.opts)
}

// EngineVersion wraps Engine.ServerVersion (GET /version).
func EngineVersion(ctx context.Context, api Engine) (ServerVersionResult, error) {
	if err := engineAPI(api, "EngineVersion"); err != nil {
		return ServerVersionResult{}, err
	}
	return api.ServerVersion(ctx, ServerVersionOptions{})
}

// ClientVersion returns the API version this Engine client is using.
func ClientVersion(api Engine) (string, error) {
	if err := engineAPI(api, "ClientVersion"); err != nil {
		return "", err
	}
	return api.ClientVersion(), nil
}

// DaemonHost returns the Engine daemon host the client is connected to.
func DaemonHost(api Engine) (string, error) {
	if err := engineAPI(api, "DaemonHost"); err != nil {
		return "", err
	}
	return api.DaemonHost(), nil
}

// EngineServiceLogsOption configures [SwarmServiceLogs].
type EngineServiceLogsOption interface {
	applyEngineServiceLogsOption(*engineServiceLogsAct) error
}

type engineServiceLogsAct struct{ opts ServiceLogsOptions }

// SwarmServiceLogs wraps Engine.ServiceLogs.
func SwarmServiceLogs(ctx context.Context, api Engine, id string, opts ...EngineServiceLogsOption) (ServiceLogsResult, error) {
	if err := engineAPI(api, "SwarmServiceLogs"); err != nil {
		return nil, err
	}
	var act engineServiceLogsAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("servicelogs.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineServiceLogsOption(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	return api.ServiceLogs(ctx, id, act.opts)
}

// EngineTaskLogsOption configures [TaskLogs].
type EngineTaskLogsOption interface {
	applyEngineTaskLogsOption(*engineTaskLogsAct) error
}

type engineTaskLogsAct struct{ opts TaskLogsOptions }

func (o FollowOpt) applyEngineServiceLogsOption(a *engineServiceLogsAct) error {
	a.opts.Follow = o.v
	return nil
}
func (o FollowOpt) applyEngineTaskLogsOption(a *engineTaskLogsAct) error {
	a.opts.Follow = o.v
	return nil
}
func (o TailOpt) applyEngineServiceLogsOption(a *engineServiceLogsAct) error {
	a.opts.Tail = o.n
	return nil
}
func (o TailOpt) applyEngineTaskLogsOption(a *engineTaskLogsAct) error {
	a.opts.Tail = o.n
	return nil
}
func (o SinceOpt) applyEngineServiceLogsOption(a *engineServiceLogsAct) error {
	a.opts.Since = o.s
	return nil
}
func (o SinceOpt) applyEngineTaskLogsOption(a *engineTaskLogsAct) error {
	a.opts.Since = o.s
	return nil
}
func (o UntilOpt) applyEngineServiceLogsOption(a *engineServiceLogsAct) error {
	a.opts.Until = o.s
	return nil
}
func (o UntilOpt) applyEngineTaskLogsOption(a *engineTaskLogsAct) error {
	a.opts.Until = o.s
	return nil
}
func (o TimestampsOpt) applyEngineServiceLogsOption(a *engineServiceLogsAct) error {
	a.opts.Timestamps = o.v
	return nil
}
func (o TimestampsOpt) applyEngineTaskLogsOption(a *engineTaskLogsAct) error {
	a.opts.Timestamps = o.v
	return nil
}
func (o DetailsOpt) applyEngineServiceLogsOption(a *engineServiceLogsAct) error {
	a.opts.Details = o.v
	return nil
}
func (o DetailsOpt) applyEngineTaskLogsOption(a *engineTaskLogsAct) error {
	a.opts.Details = o.v
	return nil
}

// TaskLogs wraps Engine.TaskLogs.
func TaskLogs(ctx context.Context, api Engine, id string, opts ...EngineTaskLogsOption) (TaskLogsResult, error) {
	if err := engineAPI(api, "TaskLogs"); err != nil {
		return nil, err
	}
	var act engineTaskLogsAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("tasklogs.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineTaskLogsOption(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	return api.TaskLogs(ctx, id, act.opts)
}
