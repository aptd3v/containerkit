package kit

import (
	"fmt"
	"time"

	"github.com/aptd3v/containerkit/pkg/endpoint"
	"github.com/aptd3v/containerkit/pkg/failact"
	"github.com/aptd3v/containerkit/pkg/order"
)

// ExternalLinkOpt is the value returned by [ExternalLink].
// It implements [ServiceOption].
type ExternalLinkOpt struct{ name string }

// ExternalLink adds a Compose external_links entry. Service only.
func ExternalLink(name string) ExternalLinkOpt { return ExternalLinkOpt{name: name} }

func (o ExternalLinkOpt) applyService(s *Service) error {
	if o.name == "" {
		return cfgErr("", "external_links", "ExternalLink", fmt.Errorf("%w: link", errEmpty))
	}
	s.externalLinks = append(s.externalLinks, o.name)
	return nil
}

// PullRefreshAfterOpt is the value returned by [PullRefreshAfter].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type PullRefreshAfterOpt struct{ s string }

// PullRefreshAfter sets Compose pull_refresh_after (duration such as "24h").
func PullRefreshAfter(s string) PullRefreshAfterOpt { return PullRefreshAfterOpt{s: s} }

func (o PullRefreshAfterOpt) applyService(s *Service) error {
	s.pullRefreshAfter.assign(o.s)
	return nil
}
func (o PullRefreshAfterOpt) applyJob(j *Job) error {
	j.pullRefreshAfter.assign(o.s)
	return nil
}
func (o PullRefreshAfterOpt) applyContainer(c *Container) error {
	c.pullRefreshAfter.assign(o.s)
	return nil
}

// DevelopOpt is the value returned by [Develop].
// It implements [ServiceOption].
type DevelopOpt struct{ opts []DevelopOption }

// Develop sets Compose develop.watch. Service only.
func Develop(opts ...DevelopOption) DevelopOpt { return DevelopOpt{opts: opts} }

func (o DevelopOpt) applyService(s *Service) error {
	var d develop
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("develop.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyDevelop(&d); err != nil {
			errs = append(errs, wrapPath("develop", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	s.develop.assign(d)
	return nil
}

// WatchFilesOpt is the value returned by [WatchFiles].
// It implements [DevelopOption].
type WatchFilesOpt struct{ opts []WatchTriggerOption }

// WatchFiles adds a develop.watch trigger inside [Develop].
func WatchFiles(opts ...WatchTriggerOption) WatchFilesOpt { return WatchFilesOpt{opts: opts} }

func (o WatchFilesOpt) applyDevelop(d *develop) error {
	var t watchTrigger
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("watch.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyWatchTrigger(&t); err != nil {
			errs = append(errs, wrapPath("watch", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !t.path.ok() {
		return cfgErr("", "watch.path", "Path", fmt.Errorf("%w: watch path", errRequired))
	}
	d.watch = append(d.watch, t)
	return nil
}

func (o PathOpt) applyWatchTrigger(t *watchTrigger) error {
	if o.p == "" {
		return cfgErr("", "path", "Path", fmt.Errorf("%w: watch path", errEmpty))
	}
	t.path.assign(o.p)
	return nil
}

func (o TargetOpt) applyWatchTrigger(t *watchTrigger) error {
	t.target.assign(o.s)
	return nil
}

// ActionOpt is the value returned by [Action].
// It implements [WatchTriggerOption].
type ActionOpt struct{ a WatchAction }

// Action sets a develop.watch action inside [WatchFiles].
func Action(a WatchAction) ActionOpt { return ActionOpt{a: a} }

func (o ActionOpt) applyWatchTrigger(t *watchTrigger) error {
	t.action.assign(o.a)
	return nil
}

// IncludeOpt is the value returned by [Include].
// It implements [WatchTriggerOption].
type IncludeOpt struct{ p string }

// Include adds a develop.watch include glob inside [WatchFiles].
func Include(p string) IncludeOpt { return IncludeOpt{p: p} }

func (o IncludeOpt) applyWatchTrigger(t *watchTrigger) error {
	if o.p == "" {
		return cfgErr("", "include", "Include", fmt.Errorf("%w: include", errEmpty))
	}
	t.include = append(t.include, o.p)
	return nil
}

// IgnoreOpt is the value returned by [Ignore].
// It implements [WatchTriggerOption].
type IgnoreOpt struct{ p string }

// Ignore adds a develop.watch ignore glob inside [WatchFiles].
func Ignore(p string) IgnoreOpt { return IgnoreOpt{p: p} }

func (o IgnoreOpt) applyWatchTrigger(t *watchTrigger) error {
	if o.p == "" {
		return cfgErr("", "ignore", "Ignore", fmt.Errorf("%w: ignore", errEmpty))
	}
	t.ignore = append(t.ignore, o.p)
	return nil
}

// ModelOpt is the value returned by [Model].
// It implements [JobOption] and [ServiceOption].
type ModelOpt struct{ opts []ModelOption }

// Model adds a Compose models mapping. Service and job.
func Model(opts ...ModelOption) ModelOpt { return ModelOpt{opts: opts} }

func (o ModelOpt) applyService(s *Service) error { return o.apply(&s.spec) }
func (o ModelOpt) applyJob(j *Job) error         { return o.apply(&j.spec) }

func (o ModelOpt) apply(s *spec) error {
	var m serviceModel
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("model.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyModel(&m); err != nil {
			errs = append(errs, wrapPath("models", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !m.name.ok() {
		return cfgErr("", "models.name", "Name", fmt.Errorf("%w: model name", errRequired))
	}
	s.models = append(s.models, m)
	return nil
}

func (o NameOpt) applyModel(m *serviceModel) error {
	if err := validateName(o.name); err != nil {
		return wrapPath("models", err)
	}
	m.name.assign(o.name)
	return nil
}

// EndpointVarOpt is the value returned by [EndpointVar].
// It implements [ModelOption].
type EndpointVarOpt struct{ s string }

// EndpointVar sets models.*.endpoint_var inside [Model].
func EndpointVar(s string) EndpointVarOpt { return EndpointVarOpt{s: s} }

func (o EndpointVarOpt) applyModel(m *serviceModel) error {
	m.endpointVar.assign(o.s)
	return nil
}

// ModelVarOpt is the value returned by [ModelVar].
// It implements [ModelOption].
type ModelVarOpt struct{ s string }

// ModelVar sets models.*.model_var inside [Model].
func ModelVar(s string) ModelVarOpt { return ModelVarOpt{s: s} }

func (o ModelVarOpt) applyModel(m *serviceModel) error {
	m.modelVar.assign(o.s)
	return nil
}

// EntitlementOpt is the value returned by [Entitlement].
// It implements [BuildOption].
type EntitlementOpt struct{ s string }

// Entitlement adds a build entitlement inside [Build].
func Entitlement(s string) EntitlementOpt { return EntitlementOpt{s: s} }

func (o EntitlementOpt) applyBuild(b *build) error {
	if o.s == "" {
		return cfgErr("", "entitlements", "Entitlement", fmt.Errorf("%w: entitlement", errEmpty))
	}
	b.entitlements = append(b.entitlements, o.s)
	return nil
}

// ProvenanceOpt is the value returned by [Provenance].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption], [UpOption], and
// [WatchOption].
type ProvenanceOpt struct{ s string }

// Provenance sets build provenance inside [Build] or [Compose.Build].
func Provenance(s string) ProvenanceOpt { return ProvenanceOpt{s: s} }

func (o ProvenanceOpt) applyBuild(b *build) error {
	b.provenance.assign(o.s)
	return nil
}

// SBOMOpt is the value returned by [SBOM].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption], [UpOption], and
// [WatchOption].
type SBOMOpt struct{ s string }

// SBOM sets build SBOM attestation inside [Build] or [Compose.Build].
func SBOM(s string) SBOMOpt { return SBOMOpt{s: s} }

func (o SBOMOpt) applyBuild(b *build) error {
	b.sbom.assign(o.s)
	return nil
}

// SSHOpt is the value returned by [SSH].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption], [UpOption], and
// [WatchOption].
type SSHOpt struct{ id, path string }

// SSH adds a build SSH mount inside [Build] or [Compose.Build] (id, optional path).
func SSH(id, path string) SSHOpt { return SSHOpt{id: id, path: path} }

func (o SSHOpt) applyBuild(b *build) error {
	if o.id == "" {
		return cfgErr("", "ssh", "SSH", fmt.Errorf("%w: ssh id", errEmpty))
	}
	b.ssh = append(b.ssh, sshKey{id: o.id, path: o.path})
	return nil
}

// CacheFromOpt is the value returned by [CacheFrom].
// It implements [BuildOption].
type CacheFromOpt struct{ s string }

// CacheFrom adds a build cache_from entry inside [Build].
func CacheFrom(s string) CacheFromOpt { return CacheFromOpt{s: s} }

func (o CacheFromOpt) applyBuild(b *build) error {
	if o.s == "" {
		return cfgErr("", "cache_from", "CacheFrom", fmt.Errorf("%w: cache from", errEmpty))
	}
	b.cacheFrom = append(b.cacheFrom, o.s)
	return nil
}

// CacheToOpt is the value returned by [CacheTo].
// It implements [BuildOption].
type CacheToOpt struct{ s string }

// CacheTo adds a build cache_to entry inside [Build].
func CacheTo(s string) CacheToOpt { return CacheToOpt{s: s} }

func (o CacheToOpt) applyBuild(b *build) error {
	if o.s == "" {
		return cfgErr("", "cache_to", "CacheTo", fmt.Errorf("%w: cache to", errEmpty))
	}
	b.cacheTo = append(b.cacheTo, o.s)
	return nil
}

// NoCacheFilterOpt is the value returned by [NoCacheFilter].
// It implements [BuildOption].
type NoCacheFilterOpt struct{ s string }

// NoCacheFilter adds a build no_cache_filter stage inside [Build].
func NoCacheFilter(s string) NoCacheFilterOpt { return NoCacheFilterOpt{s: s} }

func (o NoCacheFilterOpt) applyBuild(b *build) error {
	if o.s == "" {
		return cfgErr("", "no_cache_filter", "NoCacheFilter", fmt.Errorf("%w: filter", errEmpty))
	}
	b.noCacheFilter = append(b.noCacheFilter, o.s)
	return nil
}

// AdditionalContextOpt is the value returned by [AdditionalContext].
// It implements [BuildOption].
type AdditionalContextOpt struct{ name, value string }

// AdditionalContext adds a named additional build context inside [Build].
func AdditionalContext(name, value string) AdditionalContextOpt {
	return AdditionalContextOpt{name: name, value: value}
}

func (o AdditionalContextOpt) applyBuild(b *build) error {
	if o.name == "" {
		return cfgErr("", "additional_contexts", "AdditionalContext", fmt.Errorf("%w: context name", errEmpty))
	}
	if b.extraContexts == nil {
		b.extraContexts = map[string]string{}
	}
	b.extraContexts[o.name] = o.value
	return nil
}

func (o SecretOpt) applyBuild(b *build) error {
	return o.apply(&b.secrets)
}

func (o UlimitOpt) applyBuild(b *build) error {
	var u ulimit
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("ulimit.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyUlimit(&u); err != nil {
			errs = append(errs, wrapPath("ulimit", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !u.name.ok() {
		return cfgErr("", "ulimit.name", "Name", fmt.Errorf("%w: ulimit name", errRequired))
	}
	b.ulimits = append(b.ulimits, u)
	return nil
}

func (o PlatformOpt) applyBuild(b *build) error {
	if o.p == "" {
		return cfgErr("", "platform", "Platform", fmt.Errorf("%w: platform", errEmpty))
	}
	b.platforms = append(b.platforms, o.p)
	return nil
}

// IsolationOpt is the value returned by [Isolation].
// It implements [BuildOption], [ContainerOption], [JobOption], and [ServiceOption].
type IsolationOpt struct{ m string }

func (o IsolationOpt) applyContainer(c *Container) error { return c.setIsolation(o.m) }
func (o IsolationOpt) applyService(s *Service) error     { return s.setIsolation(o.m) }
func (o IsolationOpt) applyJob(j *Job) error             { return j.setIsolation(o.m) }
func (o IsolationOpt) applyBuild(b *build) error {
	b.isolation.assign(o.m)
	return nil
}

// ExtraHostOpt is the value returned by [ExtraHost].
// It implements [BuildOption], [ContainerOption], [JobOption], and [ServiceOption].
type ExtraHostOpt struct{ hostname, ip string }

func (o ExtraHostOpt) applyContainer(c *Container) error {
	return addExtraHost(&c.spec, o.hostname, o.ip)
}
func (o ExtraHostOpt) applyService(s *Service) error {
	return addExtraHost(&s.spec, o.hostname, o.ip)
}
func (o ExtraHostOpt) applyJob(j *Job) error { return addExtraHost(&j.spec, o.hostname, o.ip) }
func (o ExtraHostOpt) applyBuild(b *build) error {
	if err := extraHostErr(o.hostname, o.ip); err != nil {
		return err
	}
	b.extraHosts = append(b.extraHosts, extraHost{hostname: o.hostname, ip: o.ip})
	return nil
}

// NetworkModeOpt is the value returned by [NetworkMode].
// It implements [BuildOption], [ContainerOption], [JobOption], and [ServiceOption].
type NetworkModeOpt struct{ m string }

func (o NetworkModeOpt) applyContainer(c *Container) error { return c.setNetworkMode(o.m) }
func (o NetworkModeOpt) applyService(s *Service) error     { return s.setNetworkMode(o.m) }
func (o NetworkModeOpt) applyJob(j *Job) error             { return j.setNetworkMode(o.m) }
func (o NetworkModeOpt) applyBuild(b *build) error {
	b.network.assign(o.m)
	return nil
}

func extraHostErr(hostname, ip string) error {
	if hostname == "" {
		return cfgErr("", "extra_host", "ExtraHost", fmt.Errorf("%w: hostname", errEmpty))
	}
	if ip == "" {
		return cfgErr("", "extra_host", "ExtraHost", fmt.Errorf("%w: ip", errEmpty))
	}
	return nil
}

func addExtraHost(s *spec, hostname, ip string) error {
	if err := extraHostErr(hostname, ip); err != nil {
		return err
	}
	s.extraHosts = append(s.extraHosts, extraHost{hostname: hostname, ip: ip})
	return nil
}

// UpdateOpt is the value returned by [Update] and [Rollback].
// It implements [DeployOption].
type UpdateOpt struct {
	opts []UpdateOption
	kind string
}

// Update sets deploy.update_config inside [Deploy].
func Update(opts ...UpdateOption) UpdateOpt { return UpdateOpt{opts: opts, kind: "update"} }

// Rollback sets deploy.rollback_config inside [Deploy].
func Rollback(opts ...UpdateOption) UpdateOpt { return UpdateOpt{opts: opts, kind: "rollback"} }

func (o UpdateOpt) applyDeploy(d *deploy) error {
	var u updateCfg
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("%s.option[%d]", o.kind, i), "option", errNilOption))
			continue
		}
		if err := opt.applyUpdate(&u); err != nil {
			errs = append(errs, wrapPath(o.kind, err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if o.kind == "rollback" {
		d.rollback.assign(u)
	} else {
		d.update.assign(u)
	}
	return nil
}

// ParallelismOpt is the value returned by [Parallelism].
// It implements [UpdateOption].
type ParallelismOpt struct{ n uint64 }

// Parallelism sets update/rollback parallelism.
func Parallelism(n uint64) ParallelismOpt { return ParallelismOpt{n: n} }

func (o ParallelismOpt) applyUpdate(u *updateCfg) error {
	u.parallelism.assign(o.n)
	return nil
}

// DelayOpt is the value returned by [Delay].
// It implements [RestartOption] and [UpdateOption].
type DelayOpt struct{ d time.Duration }

// Delay sets update delay or deploy restart_policy delay.
func Delay(d time.Duration) DelayOpt { return DelayOpt{d: d} }

func (o DelayOpt) applyUpdate(u *updateCfg) error {
	u.delay.assign(o.d)
	return nil
}
func (o DelayOpt) applyRestart(r *restart) error {
	r.delay.assign(o.d)
	return nil
}

// FailureActionOpt is the value returned by [FailureAction].
// It implements [UpdateOption].
type FailureActionOpt struct{ s string }

// FailureAction sets update/rollback failure_action.
func FailureAction(s failact.Action) FailureActionOpt { return FailureActionOpt{s: string(s)} }

func (o FailureActionOpt) applyUpdate(u *updateCfg) error {
	u.failureAction.assign(o.s)
	return nil
}

// MonitorOpt is the value returned by [Monitor].
// It implements [UpdateOption].
type MonitorOpt struct{ d time.Duration }

// Monitor sets update/rollback monitor duration.
func Monitor(d time.Duration) MonitorOpt { return MonitorOpt{d: d} }

func (o MonitorOpt) applyUpdate(u *updateCfg) error {
	u.monitor.assign(o.d)
	return nil
}

// MaxFailureRatioOpt is the value returned by [MaxFailureRatio].
// It implements [UpdateOption].
type MaxFailureRatioOpt struct{ n float32 }

// MaxFailureRatio sets update/rollback max_failure_ratio.
func MaxFailureRatio(n float32) MaxFailureRatioOpt { return MaxFailureRatioOpt{n: n} }

func (o MaxFailureRatioOpt) applyUpdate(u *updateCfg) error {
	u.maxFailureRatio.assign(o.n)
	return nil
}

// OrderOpt is the value returned by [Order].
// It implements [UpdateOption].
type OrderOpt struct{ s string }

// Order sets update/rollback order (start-first / stop-first).
func Order(s order.Order) OrderOpt { return OrderOpt{s: string(s)} }

func (o OrderOpt) applyUpdate(u *updateCfg) error {
	u.order.assign(o.s)
	return nil
}

// WindowOpt is the value returned by [Window].
// It implements [RestartOption].
type WindowOpt struct{ d time.Duration }

// Window sets deploy restart_policy.window inside [Restart].
func Window(d time.Duration) WindowOpt { return WindowOpt{d: d} }

func (o WindowOpt) applyRestart(r *restart) error {
	r.window.assign(o.d)
	return nil
}

func (o RestartOpt) applyDeploy(d *deploy) error {
	var r restart
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("restart.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyRestart(&r); err != nil {
			errs = append(errs, wrapPath("restart", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	d.restart.assign(r)
	return nil
}

// PlacementOpt is the value returned by [Placement].
// It implements [DeployOption].
type PlacementOpt struct{ opts []PlacementOption }

// Placement sets deploy.placement inside [Deploy].
func Placement(opts ...PlacementOption) PlacementOpt { return PlacementOpt{opts: opts} }

func (o PlacementOpt) applyDeploy(d *deploy) error {
	var p placement
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("placement.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyPlacement(&p); err != nil {
			errs = append(errs, wrapPath("placement", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	d.placement.assign(p)
	return nil
}

// ConstraintOpt is the value returned by [Constraint].
// It implements [PlacementOption].
type ConstraintOpt struct{ s string }

// Constraint adds a placement constraint inside [Placement].
func Constraint(s string) ConstraintOpt { return ConstraintOpt{s: s} }

func (o ConstraintOpt) applyPlacement(p *placement) error {
	if o.s == "" {
		return cfgErr("", "constraint", "Constraint", fmt.Errorf("%w: constraint", errEmpty))
	}
	p.constraints = append(p.constraints, o.s)
	return nil
}

// SpreadOpt is the value returned by [Spread].
// It implements [PlacementOption].
type SpreadOpt struct{ s string }

// Spread adds a placement spread preference inside [Placement].
func Spread(s string) SpreadOpt { return SpreadOpt{s: s} }

func (o SpreadOpt) applyPlacement(p *placement) error {
	if o.s == "" {
		return cfgErr("", "spread", "Spread", fmt.Errorf("%w: spread", errEmpty))
	}
	p.spreads = append(p.spreads, o.s)
	return nil
}

// MaxReplicasOpt is the value returned by [MaxReplicas].
// It implements [PlacementOption].
type MaxReplicasOpt struct{ n uint64 }

// MaxReplicas sets placement max_replicas_per_node inside [Placement].
func MaxReplicas(n uint64) MaxReplicasOpt { return MaxReplicasOpt{n: n} }

func (o MaxReplicasOpt) applyPlacement(p *placement) error {
	p.maxReplicas.assign(o.n)
	return nil
}

// EndpointModeOpt is the value returned by [EndpointMode].
// It implements [DeployOption].
type EndpointModeOpt struct{ s string }

// EndpointMode sets deploy.endpoint_mode inside [Deploy].
func EndpointMode(s endpoint.Mode) EndpointModeOpt { return EndpointModeOpt{s: string(s)} }

func (o EndpointModeOpt) applyDeploy(d *deploy) error {
	d.endpointMode.assign(o.s)
	return nil
}

// ReserveOpt is the value returned by [Reserve].
// It implements [DeployOption].
type ReserveOpt struct{ opts []ReserveOption }

// Reserve sets deploy.resources.reservations inside [Deploy].
func Reserve(opts ...ReserveOption) ReserveOpt { return ReserveOpt{opts: opts} }

func (o ReserveOpt) applyDeploy(d *deploy) error {
	var r resource
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("reservations.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyReserve(&r); err != nil {
			errs = append(errs, wrapPath("reservations", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	d.reserve.assign(r)
	return nil
}

func (o MemoryOpt) applyReserve(r *resource) error {
	if o.n < 0 {
		return cfgErr("", "memory", "Memory", fmt.Errorf("%w: reservation memory", errInvalid))
	}
	r.memory.assign(o.n)
	return nil
}

func (o CPUsOpt) applyReserve(r *resource) error {
	if o.n < 0 {
		return cfgErr("", "cpus", "CPUs", fmt.Errorf("%w: reservation cpus", errInvalid))
	}
	r.cpus.assign(o.n)
	return nil
}

// EnableIPv4Opt is the value returned by [EnableIPv4].
// It implements [ComposeNetOption].
type EnableIPv4Opt struct{ v bool }

// EnableIPv4 sets Compose network enable_ipv4.
func EnableIPv4(v bool) EnableIPv4Opt { return EnableIPv4Opt{v: v} }

func (o EnableIPv4Opt) applyComposeNet(n *projectNet) error {
	n.ipv4.assign(o.v)
	return nil
}

// EnableIPv6Opt is the value returned by [EnableIPv6].
// It implements [ComposeNetOption].
type EnableIPv6Opt struct{ v bool }

// EnableIPv6 sets Compose network enable_ipv6.
func EnableIPv6(v bool) EnableIPv6Opt { return EnableIPv6Opt{v: v} }

func (o EnableIPv6Opt) applyComposeNet(n *projectNet) error {
	n.ipv6.assign(o.v)
	return nil
}

// ContextSizeOpt is the value returned by [ContextSize].
// It implements [ComposeModelOption].
type ContextSizeOpt struct{ n int }

// ContextSize sets a top-level Compose model's context_size inside [AddModel].
func ContextSize(n int) ContextSizeOpt { return ContextSizeOpt{n: n} }

func (o ContextSizeOpt) applyComposeModel(m *projectModel) error {
	m.contextSize.assign(o.n)
	return nil
}

// RuntimeFlagOpt is the value returned by [RuntimeFlag].
// It implements [ComposeModelOption].
type RuntimeFlagOpt struct{ s string }

// RuntimeFlag adds a top-level Compose model runtime flag inside [AddModel].
func RuntimeFlag(s string) RuntimeFlagOpt { return RuntimeFlagOpt{s: s} }

func (o RuntimeFlagOpt) applyComposeModel(m *projectModel) error {
	if o.s == "" {
		return cfgErr("", "runtime_flags", "RuntimeFlag", fmt.Errorf("%w: flag", errEmpty))
	}
	m.runtimeFlags = append(m.runtimeFlags, o.s)
	return nil
}

// AddModelOpt is the value returned by [AddModel].
// It implements [ComposeOption].
type AddModelOpt struct{ opts []ComposeModelOption }

// AddModel defines a project-level Compose model.
func AddModel(opts ...ComposeModelOption) AddModelOpt { return AddModelOpt{opts: opts} }

func (o AddModelOpt) applyCompose(p *Compose) error {
	var m projectModel
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("model.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyComposeModel(&m); err != nil {
			errs = append(errs, wrapPath("model", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !m.name.ok() {
		return cfgErr("", "model.name", "Name", fmt.Errorf("%w: model name", errRequired))
	}
	p.models = append(p.models, m)
	return nil
}

var (
	_ ServiceOption        = ExternalLink("db")
	_ ServiceOption        = Develop(WatchFiles(Path("./src"), Action(Sync), Target("/app")))
	_ ServiceOption        = Model(Name("llm"), EndpointVar("MODEL_URL"))
	_ BuildOption          = Entitlement("network.host")
	_ BuildOption          = SSH("default", "")
	_ BuildOption          = CacheFrom("type=registry")
	_ DeployOption         = Update(Parallelism(1), Delay(0), FailureAction("pause"))
	_ DeployOption         = Placement(Constraint("node.role==manager"), Spread("node.labels.zone"))
	_ DeployOption         = Reserve(Memory(1<<20), CPUs(0.1))
	_ DeployOption         = Restart(Policy(OnFailure), MaxRetries(3))
	_ ComposeNetOption     = EnableIPv4(true)
	_ ComposeNetOption     = EnableIPv6(true)
	_ ComposeOption        = AddModel(Name("llm"), Image("ai/mistral"))
	_ ComposeModelOption   = ContextSize(4096)
	_ WatchTriggerOption   = Path("./src")
	_ EngineStatPathOption = Path("/etc/hosts")
)
