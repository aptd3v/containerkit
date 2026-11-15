package kit

import (
	"context"
	"fmt"
	"io"

	swarmtypes "github.com/aptd3v/containerkit/internal/engine/swarm/types"
	"github.com/aptd3v/containerkit/pkg/svcmode"
)

// EngineServiceCreateOption configures [CreateSwarmService].
type EngineServiceCreateOption interface {
	applyEngineServiceCreateOption(*engineServiceCreateAct) error
}

type engineServiceCreateAct struct{ opts ServiceCreateOptions }

// EngineServiceUpdateOption configures [UpdateSwarmService].
type EngineServiceUpdateOption interface {
	applyEngineServiceUpdateOption(*engineServiceUpdateAct) error
}

type engineServiceUpdateAct struct{ opts ServiceUpdateOptions }

// EngineSecretCreateOption configures [CreateEngineSecret].
type EngineSecretCreateOption interface {
	applyEngineSecretCreateOption(*engineSecretCreateAct) error
}

type engineSecretCreateAct struct{ opts SecretCreateOptions }

// EngineSecretUpdateOption configures [UpdateEngineSecret].
type EngineSecretUpdateOption interface {
	applyEngineSecretUpdateOption(*engineSecretUpdateAct) error
}

type engineSecretUpdateAct struct{ opts SecretUpdateOptions }

// EngineConfigCreateOption configures [CreateEngineConfig].
type EngineConfigCreateOption interface {
	applyEngineConfigCreateOption(*engineConfigCreateAct) error
}

type engineConfigCreateAct struct{ opts ConfigCreateOptions }

// EngineConfigUpdateOption configures [UpdateEngineConfig].
type EngineConfigUpdateOption interface {
	applyEngineConfigUpdateOption(*engineConfigUpdateAct) error
}

type engineConfigUpdateAct struct{ opts ConfigUpdateOptions }

// EngineNodeUpdateOption configures [UpdateNode].
type EngineNodeUpdateOption interface {
	applyEngineNodeUpdateOption(*engineNodeUpdateAct) error
}

type engineNodeUpdateAct struct{ opts NodeUpdateOptions }

// EngineVolumeUpdateOption configures [UpdateVolume].
type EngineVolumeUpdateOption interface {
	applyEngineVolumeUpdateOption(*engineVolumeUpdateAct) error
}

type engineVolumeUpdateAct struct{ opts VolumeUpdateOptions }

func swarmTaskContainer(spec *ServiceSpec) *ContainerSpec {
	if spec.TaskTemplate.ContainerSpec == nil {
		spec.TaskTemplate.ContainerSpec = &ContainerSpec{}
	}
	return spec.TaskTemplate.ContainerSpec
}

func setAnnotationLabel(labels *map[string]string, key, value string) error {
	if key == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if *labels == nil {
		*labels = map[string]string{}
	}
	(*labels)[key] = value
	return nil
}

func setSwarmServiceMode(spec *ServiceSpec, m string) error {
	switch svcmode.Mode(m) {
	case svcmode.Global:
		spec.Mode = swarmtypes.ServiceMode{Global: &swarmtypes.GlobalService{}}
	case svcmode.Replicated:
		spec.Mode = swarmtypes.ServiceMode{Replicated: &swarmtypes.ReplicatedService{}}
	case svcmode.ReplicatedJob:
		spec.Mode = swarmtypes.ServiceMode{ReplicatedJob: &swarmtypes.ReplicatedJob{}}
	case svcmode.GlobalJob:
		spec.Mode = swarmtypes.ServiceMode{GlobalJob: &swarmtypes.GlobalJob{}}
	default:
		return cfgErr("", "mode", "Mode", fmt.Errorf("%w: swarm service mode %q", errInvalid, m))
	}
	return nil
}

func (o NameOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}
func (o NameOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}
func (o NameOpt) applyEngineSecretCreateOption(a *engineSecretCreateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}
func (o NameOpt) applyEngineSecretUpdateOption(a *engineSecretUpdateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}
func (o NameOpt) applyEngineConfigCreateOption(a *engineConfigCreateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}
func (o NameOpt) applyEngineConfigUpdateOption(a *engineConfigUpdateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}
func (o NameOpt) applyEngineNodeUpdateOption(a *engineNodeUpdateAct) error {
	a.opts.Spec.Name = o.name
	return nil
}

func (o ImageOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	swarmTaskContainer(&a.opts.Spec).Image = o.image
	return nil
}
func (o ImageOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	swarmTaskContainer(&a.opts.Spec).Image = o.image
	return nil
}

func (o CommandOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	swarmTaskContainer(&a.opts.Spec).Command = append([]string(nil), o.cmd...)
	return nil
}
func (o CommandOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	swarmTaskContainer(&a.opts.Spec).Command = append([]string(nil), o.cmd...)
	return nil
}

func (o EnvOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	if o.key == "" {
		return secretErr("", "env", "Env", fmt.Errorf("%w: env key", errEmpty))
	}
	cs := swarmTaskContainer(&a.opts.Spec)
	cs.Env = append(cs.Env, o.key+"="+o.value)
	return nil
}
func (o EnvOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	if o.key == "" {
		return secretErr("", "env", "Env", fmt.Errorf("%w: env key", errEmpty))
	}
	cs := swarmTaskContainer(&a.opts.Spec)
	cs.Env = append(cs.Env, o.key+"="+o.value)
	return nil
}

func (o LabelOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}
func (o LabelOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}
func (o LabelOpt) applyEngineSecretCreateOption(a *engineSecretCreateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}
func (o LabelOpt) applyEngineSecretUpdateOption(a *engineSecretUpdateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}
func (o LabelOpt) applyEngineConfigCreateOption(a *engineConfigCreateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}
func (o LabelOpt) applyEngineConfigUpdateOption(a *engineConfigUpdateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}
func (o LabelOpt) applyEngineNodeUpdateOption(a *engineNodeUpdateAct) error {
	return setAnnotationLabel(&a.opts.Spec.Labels, o.key, o.value)
}

func (o ReplicasOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	if o.n < 0 {
		return cfgErr("", "replicas", "Replicas", fmt.Errorf("%w: replicas", errInvalid))
	}
	n := uint64(o.n)
	a.opts.Spec.Mode = swarmtypes.ServiceMode{Replicated: &swarmtypes.ReplicatedService{Replicas: &n}}
	return nil
}
func (o ReplicasOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	if o.n < 0 {
		return cfgErr("", "replicas", "Replicas", fmt.Errorf("%w: replicas", errInvalid))
	}
	n := uint64(o.n)
	a.opts.Spec.Mode = swarmtypes.ServiceMode{Replicated: &swarmtypes.ReplicatedService{Replicas: &n}}
	return nil
}

func (o ModeOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	return setSwarmServiceMode(&a.opts.Spec, o.m)
}
func (o ModeOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	return setSwarmServiceMode(&a.opts.Spec, o.m)
}

func (o RegistryAuthOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	a.opts.EncodedRegistryAuth = o.s
	return nil
}
func (o RegistryAuthOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	a.opts.EncodedRegistryAuth = o.s
	return nil
}

func (o ContentOpt) applyEngineSecretCreateOption(a *engineSecretCreateAct) error {
	b, err := readPayload(o.r)
	if err != nil {
		return err
	}
	a.opts.Spec.Data = b
	return nil
}
func (o ContentOpt) applyEngineSecretUpdateOption(a *engineSecretUpdateAct) error {
	b, err := readPayload(o.r)
	if err != nil {
		return err
	}
	a.opts.Spec.Data = b
	return nil
}
func (o ContentOpt) applyEngineConfigCreateOption(a *engineConfigCreateAct) error {
	b, err := readPayload(o.r)
	if err != nil {
		return err
	}
	a.opts.Spec.Data = b
	return nil
}
func (o ContentOpt) applyEngineConfigUpdateOption(a *engineConfigUpdateAct) error {
	b, err := readPayload(o.r)
	if err != nil {
		return err
	}
	a.opts.Spec.Data = b
	return nil
}

func readPayload(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, cfgErr("", "content", "Content", fmt.Errorf("%w: reader", errRequired))
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, cfgErr("", "content", "Content", err)
	}
	return append([]byte(nil), b...), nil
}

// DataOpt is the value returned by [Data].
// It implements [EngineConfigCreateOption], [EngineConfigUpdateOption],
// [EngineSecretCreateOption], and [EngineSecretUpdateOption].
type DataOpt struct{ b []byte }

// Data sets secret or config payload bytes for swarm create/update.
func Data(b []byte) DataOpt { return DataOpt{b: append([]byte(nil), b...)} }

func (o DataOpt) applyEngineSecretCreateOption(a *engineSecretCreateAct) error {
	a.opts.Spec.Data = append([]byte(nil), o.b...)
	return nil
}
func (o DataOpt) applyEngineSecretUpdateOption(a *engineSecretUpdateAct) error {
	a.opts.Spec.Data = append([]byte(nil), o.b...)
	return nil
}
func (o DataOpt) applyEngineConfigCreateOption(a *engineConfigCreateAct) error {
	a.opts.Spec.Data = append([]byte(nil), o.b...)
	return nil
}
func (o DataOpt) applyEngineConfigUpdateOption(a *engineConfigUpdateAct) error {
	a.opts.Spec.Data = append([]byte(nil), o.b...)
	return nil
}

// QueryRegistryOpt is the value returned by [QueryRegistry].
// It implements [EngineServiceCreateOption] and [EngineServiceUpdateOption].
type QueryRegistryOpt struct{ v bool }

// QueryRegistry contacts the registry to resolve image digest on swarm
// service create/update.
func QueryRegistry(v bool) QueryRegistryOpt { return QueryRegistryOpt{v: v} }

func (o QueryRegistryOpt) applyEngineServiceCreateOption(a *engineServiceCreateAct) error {
	a.opts.QueryRegistry = o.v
	return nil
}
func (o QueryRegistryOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	a.opts.QueryRegistry = o.v
	return nil
}

// VersionIndexOpt is the value returned by [VersionIndex].
// It implements [EngineConfigUpdateOption], [EngineNodeUpdateOption],
// [EngineSecretUpdateOption], [EngineServiceUpdateOption], and [EngineVolumeUpdateOption].
type VersionIndexOpt struct{ n uint64 }

// VersionIndex sets the object version index for swarm update operations.
func VersionIndex(n uint64) VersionIndexOpt { return VersionIndexOpt{n: n} }

func (o VersionIndexOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	a.opts.Version.Index = o.n
	return nil
}
func (o VersionIndexOpt) applyEngineSecretUpdateOption(a *engineSecretUpdateAct) error {
	a.opts.Version.Index = o.n
	return nil
}
func (o VersionIndexOpt) applyEngineConfigUpdateOption(a *engineConfigUpdateAct) error {
	a.opts.Version.Index = o.n
	return nil
}
func (o VersionIndexOpt) applyEngineNodeUpdateOption(a *engineNodeUpdateAct) error {
	a.opts.Version.Index = o.n
	return nil
}
func (o VersionIndexOpt) applyEngineVolumeUpdateOption(a *engineVolumeUpdateAct) error {
	a.opts.Version.Index = o.n
	return nil
}

// RegistryAuthFromOpt is the value returned by [RegistryAuthFrom].
// It implements [EngineServiceUpdateOption].
type RegistryAuthFromOpt struct{ src RegistryAuthSource }

// RegistryAuthFrom selects where swarm service update reads registry auth
// ("spec" or "previous-spec").
func RegistryAuthFrom[T ~string](src T) RegistryAuthFromOpt {
	return RegistryAuthFromOpt{src: RegistryAuthSource(src)}
}

func (o RegistryAuthFromOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	a.opts.RegistryAuthFrom = o.src
	return nil
}

// ServiceRollbackOpt is the value returned by [ServiceRollback].
// It implements [EngineServiceUpdateOption].
type ServiceRollbackOpt struct{ v string }

// ServiceRollback requests a server-side swarm service rollback ("previous" or "none").
func ServiceRollback(v string) ServiceRollbackOpt { return ServiceRollbackOpt{v: v} }

func (o ServiceRollbackOpt) applyEngineServiceUpdateOption(a *engineServiceUpdateAct) error {
	a.opts.Rollback = o.v
	return nil
}

// RoleOpt is the value returned by [Role].
// It implements [EngineNodeUpdateOption].
type RoleOpt struct{ r NodeRole }

// Role sets a swarm node role on [UpdateNode]. Known values live in package role.
func Role[T ~string](r T) RoleOpt { return RoleOpt{r: NodeRole(r)} }

func (o RoleOpt) applyEngineNodeUpdateOption(a *engineNodeUpdateAct) error {
	a.opts.Spec.Role = o.r
	return nil
}

// AvailabilityOpt is the value returned by [Availability].
// It implements [EngineNodeUpdateOption].
type AvailabilityOpt struct{ a NodeAvailability }

// Availability sets swarm node availability on [UpdateNode]. Known values live in package avail.
func Availability[T ~string](a T) AvailabilityOpt {
	return AvailabilityOpt{a: NodeAvailability(a)}
}

func (o AvailabilityOpt) applyEngineNodeUpdateOption(a *engineNodeUpdateAct) error {
	a.opts.Spec.Availability = o.a
	return nil
}

// ClusterSpecOpt is the value returned by [ClusterSpec].
// It implements [EngineVolumeUpdateOption].
type ClusterSpecOpt struct{ spec *ClusterVolumeSpec }

// ClusterSpec sets the cluster volume spec on [UpdateVolume].
func ClusterSpec(spec *ClusterVolumeSpec) ClusterSpecOpt { return ClusterSpecOpt{spec: spec} }

func (o ClusterSpecOpt) applyEngineVolumeUpdateOption(a *engineVolumeUpdateAct) error {
	a.opts.Spec = o.spec
	return nil
}

// CreateSwarmService wraps Engine.ServiceCreate.
func CreateSwarmService(ctx context.Context, api Engine, opts ...EngineServiceCreateOption) (ServiceCreateResult, error) {
	if err := engineAPI(api, "CreateSwarmService"); err != nil {
		return ServiceCreateResult{}, err
	}
	var act engineServiceCreateAct
	if err := applyEngineOptList("createswarmservice", opts, func(o EngineServiceCreateOption) error {
		return o.applyEngineServiceCreateOption(&act)
	}); err != nil {
		return ServiceCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ServiceCreate(ctx, act.opts)
	if err != nil {
		return ServiceCreateResult{}, cfgErr("engine", "createswarmservice", "CreateSwarmService", err)
	}
	return out, nil
}

// UpdateSwarmService wraps Engine.ServiceUpdate.
func UpdateSwarmService(ctx context.Context, api Engine, id string, opts ...EngineServiceUpdateOption) (ServiceUpdateResult, error) {
	if err := engineAPI(api, "UpdateSwarmService"); err != nil {
		return ServiceUpdateResult{}, err
	}
	var act engineServiceUpdateAct
	if err := applyEngineOptList("updateswarmservice", opts, func(o EngineServiceUpdateOption) error {
		return o.applyEngineServiceUpdateOption(&act)
	}); err != nil {
		return ServiceUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ServiceUpdate(ctx, id, act.opts)
	if err != nil {
		return ServiceUpdateResult{}, cfgErr("engine", "updateswarmservice", "UpdateSwarmService", err)
	}
	return out, nil
}

// CreateEngineSecret wraps Engine.SecretCreate.
func CreateEngineSecret(ctx context.Context, api Engine, opts ...EngineSecretCreateOption) (SecretCreateResult, error) {
	if err := engineAPI(api, "CreateEngineSecret"); err != nil {
		return SecretCreateResult{}, err
	}
	var act engineSecretCreateAct
	if err := applyEngineOptList("createenginesecret", opts, func(o EngineSecretCreateOption) error {
		return o.applyEngineSecretCreateOption(&act)
	}); err != nil {
		return SecretCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SecretCreate(ctx, act.opts)
	if err != nil {
		return SecretCreateResult{}, cfgErr("engine", "createenginesecret", "CreateEngineSecret", err)
	}
	return out, nil
}

// UpdateEngineSecret wraps Engine.SecretUpdate.
func UpdateEngineSecret(ctx context.Context, api Engine, id string, opts ...EngineSecretUpdateOption) (SecretUpdateResult, error) {
	if err := engineAPI(api, "UpdateEngineSecret"); err != nil {
		return SecretUpdateResult{}, err
	}
	var act engineSecretUpdateAct
	if err := applyEngineOptList("updateenginesecret", opts, func(o EngineSecretUpdateOption) error {
		return o.applyEngineSecretUpdateOption(&act)
	}); err != nil {
		return SecretUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.SecretUpdate(ctx, id, act.opts)
	if err != nil {
		return SecretUpdateResult{}, cfgErr("engine", "updateenginesecret", "UpdateEngineSecret", err)
	}
	return out, nil
}

// CreateEngineConfig wraps Engine.ConfigCreate.
func CreateEngineConfig(ctx context.Context, api Engine, opts ...EngineConfigCreateOption) (ConfigCreateResult, error) {
	if err := engineAPI(api, "CreateEngineConfig"); err != nil {
		return ConfigCreateResult{}, err
	}
	var act engineConfigCreateAct
	if err := applyEngineOptList("createengineconfig", opts, func(o EngineConfigCreateOption) error {
		return o.applyEngineConfigCreateOption(&act)
	}); err != nil {
		return ConfigCreateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ConfigCreate(ctx, act.opts)
	if err != nil {
		return ConfigCreateResult{}, cfgErr("engine", "createengineconfig", "CreateEngineConfig", err)
	}
	return out, nil
}

// UpdateEngineConfig wraps Engine.ConfigUpdate.
func UpdateEngineConfig(ctx context.Context, api Engine, id string, opts ...EngineConfigUpdateOption) (ConfigUpdateResult, error) {
	if err := engineAPI(api, "UpdateEngineConfig"); err != nil {
		return ConfigUpdateResult{}, err
	}
	var act engineConfigUpdateAct
	if err := applyEngineOptList("updateengineconfig", opts, func(o EngineConfigUpdateOption) error {
		return o.applyEngineConfigUpdateOption(&act)
	}); err != nil {
		return ConfigUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.ConfigUpdate(ctx, id, act.opts)
	if err != nil {
		return ConfigUpdateResult{}, cfgErr("engine", "updateengineconfig", "UpdateEngineConfig", err)
	}
	return out, nil
}

// UpdateNode wraps Engine.NodeUpdate.
func UpdateNode(ctx context.Context, api Engine, id string, opts ...EngineNodeUpdateOption) (NodeUpdateResult, error) {
	if err := engineAPI(api, "UpdateNode"); err != nil {
		return NodeUpdateResult{}, err
	}
	var act engineNodeUpdateAct
	if err := applyEngineOptList("updatenode", opts, func(o EngineNodeUpdateOption) error {
		return o.applyEngineNodeUpdateOption(&act)
	}); err != nil {
		return NodeUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.NodeUpdate(ctx, id, act.opts)
	if err != nil {
		return NodeUpdateResult{}, cfgErr("engine", "updatenode", "UpdateNode", err)
	}
	return out, nil
}

// UpdateVolume wraps Engine.VolumeUpdate.
func UpdateVolume(ctx context.Context, api Engine, id string, opts ...EngineVolumeUpdateOption) (VolumeUpdateResult, error) {
	if err := engineAPI(api, "UpdateVolume"); err != nil {
		return VolumeUpdateResult{}, err
	}
	var act engineVolumeUpdateAct
	if err := applyEngineOptList("updatevolume", opts, func(o EngineVolumeUpdateOption) error {
		return o.applyEngineVolumeUpdateOption(&act)
	}); err != nil {
		return VolumeUpdateResult{}, wrapReceiver("engine", err)
	}
	out, err := api.VolumeUpdate(ctx, id, act.opts)
	if err != nil {
		return VolumeUpdateResult{}, cfgErr("engine", "updatevolume", "UpdateVolume", err)
	}
	return out, nil
}

var (
	_ EngineServiceCreateOption = Name("")
	_ EngineServiceCreateOption = Image("")
	_ EngineServiceCreateOption = Command()
	_ EngineServiceCreateOption = Replicas(1)
	_ EngineServiceCreateOption = QueryRegistry(true)
	_ EngineSecretCreateOption  = Data(nil)
	_ EngineNodeUpdateOption    = VersionIndex(0)
)
