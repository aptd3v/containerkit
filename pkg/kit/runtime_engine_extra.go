package kit

import (
	"context"
	"fmt"
)

func engineAPI(api Engine, setting string) error {
	if api == nil {
		return cfgErr("engine", "engine", setting, fmt.Errorf("%w: engine client", errRequired))
	}
	return nil
}

// PullImage pulls an image through Engine.
func PullImage(ctx context.Context, api Engine, ref string, opts ...EnginePullOption) (ImagePullResponse, error) {
	if err := engineAPI(api, "PullImage"); err != nil {
		return nil, err
	}
	if ref == "" {
		return nil, cfgErr("engine", "ref", "PullImage", fmt.Errorf("%w: image reference", errRequired))
	}
	var act enginePullAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("pull.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEnginePull(&act))
	}
	if err := join(errs...); err != nil {
		return nil, wrapReceiver("engine", err)
	}
	adoptLog(ctx, &act.log)
	emitPhase(act.log, KindPull, "pulling", ref, StatusWorking, "")
	out, err := api.ImagePull(ctx, ref, act.sdk())
	if err != nil {
		failPhase(act.log, KindPull, "pulling", ref, err)
		return nil, cfgErr("engine", ref, "PullImage", err)
	}
	if err := act.log.drainJSON(ctx, out, KindPull, ref); err != nil {
		failPhase(act.log, KindPull, "pulling", ref, err)
		return out, cfgErr("engine", ref, "PullImage", err)
	}
	finishPhase(act.log, KindPull, "pulled", ref)
	return out, nil
}

// BuildImage builds an image through Engine.
// Supply a [Dockerfile], [Source] context directory, and/or [Tar] archive.
func BuildImage(ctx context.Context, api Engine, opts ...EngineBuildOption) (ImageBuildResult, error) {
	if err := engineAPI(api, "BuildImage"); err != nil {
		return ImageBuildResult{}, err
	}
	var act engineBuildAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("build.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineBuild(&act))
	}
	if err := join(errs...); err != nil {
		return ImageBuildResult{}, wrapReceiver("engine", err)
	}
	id := "image"
	if len(act.tags) > 0 {
		id = act.tags[0]
	}
	emitPhase(act.log, KindBuild, "building", id, StatusWorking, "")
	buildContext, err := act.reader()
	if err != nil {
		failPhase(act.log, KindBuild, "building", id, err)
		return ImageBuildResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ImageBuild(ctx, buildContext, act.sdk())
	if err != nil {
		failPhase(act.log, KindBuild, "building", id, err)
		return ImageBuildResult{}, cfgErr("engine", id, "BuildImage", err)
	}
	if err := act.log.drainJSON(ctx, out.Body, KindBuild, id); err != nil {
		failPhase(act.log, KindBuild, "building", id, err)
		return out, cfgErr("engine", id, "BuildImage", err)
	}
	finishPhase(act.log, KindBuild, "built", id)
	return out, nil
}

// CreateNetwork creates an Engine network.
func CreateNetwork(ctx context.Context, api Engine, name string, opts ...EngineNetworkOption) (NetworkCreateResult, error) {
	if err := engineAPI(api, "CreateNetwork"); err != nil {
		return NetworkCreateResult{}, err
	}
	if name == "" {
		return NetworkCreateResult{}, cfgErr("engine", "name", "CreateNetwork", fmt.Errorf("%w: network name", errRequired))
	}
	var act engineNetworkAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("network.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineNetwork(&act))
	}
	if err := join(errs...); err != nil {
		return NetworkCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NetworkCreate(ctx, name, act.sdk())
	if err != nil {
		return NetworkCreateResult{}, cfgErr("engine", "network", "CreateNetwork", err)
	}
	return out, nil
}

// CreateVolume creates an Engine volume.
func CreateVolume(ctx context.Context, api Engine, opts ...EngineVolumeOption) (VolumeCreateResult, error) {
	if err := engineAPI(api, "CreateVolume"); err != nil {
		return VolumeCreateResult{}, err
	}
	var act engineVolumeAct
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("volume.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyEngineVolume(&act))
	}
	if err := join(errs...); err != nil {
		return VolumeCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.VolumeCreate(ctx, act.sdk())
	if err != nil {
		return VolumeCreateResult{}, cfgErr("engine", "volume", "CreateVolume", err)
	}
	return out, nil
}
