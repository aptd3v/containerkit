package kit

import "fmt"

// Service is a Compose service: shared container settings plus Compose-only fields.
type Service struct {
	spec
	depends       []depend
	build         field[build]
	deploy        field[deploy]
	scale         field[int]
	profiles      []string
	envFiles      []envFileRef
	secrets       []secretRef
	configs       []fileConfigRef
	pullPolicy    field[string]
	containerName field[string]
	labelFiles    []string
	useAPISocket  field[bool]
	attach        field[bool]
	extends       field[string]
	extendsFile   field[string]
	credFile      field[string]
	credRegistry  field[string]
	credConfig    field[string]
	provider      field[string]
	providerOpts  map[string][]string
	preStart      []spec
	postStart     []serviceHook
	preStop       []serviceHook
	develop       field[develop]
	externalLinks []string
	errs          []error
	nameGenerated bool
}

// NewService applies typed options and always returns a service.
// A [*Container] is a [ServiceOption]: NewService(NewContainer(Name("foo-server"), ...), ...).
// The service name is taken from the container when present, otherwise generated
// from the image. Call [Service.Validate] to inspect accumulated errors.
func NewService(opts ...ServiceOption) *Service {
	s := &Service{}
	for i, opt := range opts {
		if opt == nil {
			s.errs = addErr(s.errs, cfgErr("service", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyService(s); err != nil {
			s.errs = addErr(s.errs, wrapReceiver("service", err))
		}
	}
	s.ensureName()
	return s
}

func (s *Service) setName(name string) error {
	s.nameGenerated = false
	return s.spec.setName(name)
}

func (s *Service) ensureName() {
	if s == nil || s.name.ok() {
		return
	}
	if n := generateServiceName(s); n != "" {
		s.name.assign(n)
		s.nameGenerated = true
	}
}

func (s *Service) validate() error {
	if s == nil {
		return cfgErr("service", ".", "", errRequired)
	}
	if !s.name.ok() {
		return cfgErr("service", "name", "Name", fmt.Errorf("%w: name", errRequired))
	}
	if !s.image.ok() && !s.build.ok() {
		return cfgErr("service", "image", "Image", fmt.Errorf("%w: image or build", errRequired))
	}
	return nil
}

// Validate reports accumulated option errors and missing required fields.
func (s *Service) Validate() error {
	if s == nil {
		return cfgErr("service", ".", "", errRequired)
	}
	return join(append(append([]error(nil), s.errs...), s.validate())...)
}

// Name returns the service name.
func (s *Service) Name() string {
	if s == nil || !s.name.ok() {
		return ""
	}
	return s.name.val
}

// Image returns the service image.
func (s *Service) Image() string {
	if s == nil || !s.image.ok() {
		return ""
	}
	return s.image.val
}

// ContainerName returns the Compose container_name, or empty if omitted.
func (s *Service) ContainerName() string {
	if s == nil || !s.containerName.ok() {
		return ""
	}
	return s.containerName.val
}

// Apply adds options to an existing service. Errors join the service's
// accumulated set and are also returned.
func (s *Service) Apply(opts ...ServiceOption) error {
	if s == nil {
		return cfgErr("service", ".", "Apply", fmt.Errorf("%w: service", errRequired))
	}
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("service", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyService(s); err != nil {
			errs = addErr(errs, wrapReceiver("service", err))
		}
	}
	s.errs = append(s.errs, errs...)
	return join(errs...)
}

func (d *depend) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("depends_on", err)
	}
	d.name.assign(name)
	return nil
}

func (d *depend) setCondition(c DependCondition) error {
	switch c {
	case Started, Healthy, CompletedSuccessfully, "":
		d.condition.assign(c)
		return nil
	default:
		return cfgErr("", "condition", "Condition", fmt.Errorf("%w: condition %q", errInvalid, c))
	}
}

func (d *depend) setRestart(v bool) error {
	d.restart.assign(v)
	return nil
}

// ConditionOpt is the value returned by [Condition].
// It implements [DependOption].
type ConditionOpt struct{ c DependCondition }

// RestartOnFailureOpt is the value returned by [RestartOnFailure].
// It implements [DependOption].
type RestartOnFailureOpt struct{ v bool }

// Condition sets a depends_on condition in a [DependsOn] scope.
func Condition(c DependCondition) ConditionOpt { return ConditionOpt{c: c} }

func (o ConditionOpt) applyDepend(d *depend) error { return d.setCondition(o.c) }

// RestartOnFailure marks depends_on restart in a [DependsOn] scope.
func RestartOnFailure(v bool) RestartOnFailureOpt { return RestartOnFailureOpt{v: v} }

func (o RestartOnFailureOpt) applyDepend(d *depend) error { return d.setRestart(o.v) }

// DependsOnOpt is the value returned by [DependsOn].
// It implements [JobOption] and [ServiceOption].
type DependsOnOpt struct{ opts []DependOption }

// DependsOn adds a service dependency with nested setters (Name, Condition).
func DependsOn(opts ...DependOption) DependsOnOpt { return DependsOnOpt{opts: opts} }

func (o DependsOnOpt) applyService(s *Service) error { return o.apply(&s.depends) }
func (o DependsOnOpt) applyJob(j *Job) error         { return o.apply(&j.depends) }

func (o DependsOnOpt) apply(dst *[]depend) error {
	var d depend
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("depends_on.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyDepend(&d); err != nil {
			errs = append(errs, wrapPath("depends_on", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !d.name.ok() {
		return cfgErr("", "depends_on.name", "Name", fmt.Errorf("%w: depends_on name", errRequired))
	}
	*dst = append(*dst, d)
	return nil
}
