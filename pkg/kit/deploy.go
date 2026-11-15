package kit

import "fmt"

func (d *deploy) setMemory(n int64) error {
	if n < 0 {
		return cfgErr("", "memory", "Memory", fmt.Errorf("%w: deploy memory", errInvalid))
	}
	d.memory.assign(n)
	return nil
}

func (d *deploy) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if d.labels == nil {
		d.labels = map[string]string{}
	}
	d.labels[k] = val
	return nil
}

// ReplicasOpt is the value returned by [Replicas].
// It implements [DeployOption], [EngineServiceCreateOption], [EngineServiceUpdateOption],
// [ScaleOption], and [ServiceOption].
type ReplicasOpt struct{ n int }

// Replicas sets deploy replicas inside [Deploy], or service scale at the outer scope.
func Replicas(n int) ReplicasOpt { return ReplicasOpt{n: n} }

func (o ReplicasOpt) applyDeploy(d *deploy) error {
	if o.n < 0 {
		return cfgErr("", "replicas", "Replicas", fmt.Errorf("%w: replicas", errInvalid))
	}
	d.replicas.assign(o.n)
	return nil
}

func (o ReplicasOpt) applyService(s *Service) error {
	if o.n < 0 {
		return cfgErr("", "scale", "Replicas", fmt.Errorf("%w: scale", errInvalid))
	}
	s.scale.assign(o.n)
	return nil
}

func (o ReplicasOpt) applyScale(a *scaleAct) error {
	if o.n < 0 {
		return cfgErr("", "replicas", "Replicas", fmt.Errorf("%w: replicas", errInvalid))
	}
	a.replicas = o.n
	a.replicasSet = true
	return nil
}

// ModeOpt is the value returned by [Mode].
// It implements [DeployOption], [EngineServiceCreateOption], [EngineServiceUpdateOption], and
// [PortOption].
type ModeOpt struct{ m string }

// Mode sets deploy mode inside [Deploy] (replicated, global, replicated-job, global-job)
// or Compose ports.mode inside [Publish] (ingress, host). Known values live in
// packages svcmode and portmode.
func Mode[T ~string](m T) ModeOpt { return ModeOpt{m: string(m)} }

func (o ModeOpt) applyDeploy(d *deploy) error {
	d.mode.assign(o.m)
	return nil
}

// DeployOpt is the value returned by [Deploy].
// It implements [ServiceOption].
type DeployOpt struct{ opts []DeployOption }

// Deploy sets Compose deploy configuration. Service only.
func Deploy(opts ...DeployOption) DeployOpt { return DeployOpt{opts: opts} }

func (o DeployOpt) applyService(s *Service) error {
	var d deploy
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("deploy.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyDeploy(&d); err != nil {
			errs = append(errs, wrapPath("deploy", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	s.deploy.assign(d)
	return nil
}

// ProfileOpt is the value returned by [Profile].
// It implements [ImportOption], [LoadProjectOption], and [ServiceOption].
type ProfileOpt struct{ p string }

// Profile adds a Compose profile on a service or [LoadProject].
func Profile(p string) ProfileOpt { return ProfileOpt{p: p} }

func (o ProfileOpt) applyService(s *Service) error {
	if o.p == "" {
		return cfgErr("", "profile", "Profile", fmt.Errorf("%w: profile", errEmpty))
	}
	s.profiles = append(s.profiles, o.p)
	return nil
}

// ContainerNameOpt is the value returned by [ContainerName].
// It implements [ServiceOption].
type ContainerNameOpt struct{ n string }

// ContainerName sets Compose container_name. Service only.
func ContainerName(n string) ContainerNameOpt { return ContainerNameOpt{n: n} }

func (o ContainerNameOpt) applyService(s *Service) error {
	if err := validateName(o.n); err != nil {
		return wrapPath("container_name", err)
	}
	s.containerName.assign(o.n)
	return nil
}
