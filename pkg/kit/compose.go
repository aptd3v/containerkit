package kit

import "fmt"

// Compose is a Compose file: name, services, jobs, networks, volumes, secrets, and configs.
type Compose struct {
	name     field[string]
	services []*Service
	jobs     []*Job
	networks []projectNet
	volumes  []projectVol
	secrets  []projectSecret
	configs  []projectFile
	models   []projectModel
	includes []includeRef
	errs     []error
}

// NewCompose applies typed options and always returns a Compose file named name.
// Use fluent [Compose.WithService] / [Compose.WithNetwork] for top-level Compose YAML.
// Call [Compose.Validate] before [Compose.Up].
func NewCompose(name string, opts ...ComposeOption) *Compose {
	p := &Compose{}
	if err := p.setName(name); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	for i, opt := range opts {
		if opt == nil {
			p.errs = addErr(p.errs, cfgErr("compose", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyCompose(p); err != nil {
			p.errs = addErr(p.errs, wrapReceiver("compose", err))
		}
	}
	return p
}

func (p *Compose) setName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	p.name.assign(name)
	return nil
}

func (p *Compose) addService(s *Service) error {
	if s == nil {
		return cfgErr("compose", "service", "Service", fmt.Errorf("%w: service", errRequired))
	}
	for _, existing := range p.services {
		if existing == s {
			return cfgErr("compose", "service", "Name", fmt.Errorf("%w: duplicate service %q", errInvalid, s.Name()))
		}
	}
	s.ensureName()
	if s.Name() != "" && p.hasService(s.Name()) {
		if s.nameGenerated {
			s.name.assign(uniqueName(s.Name(), p.hasService))
		} else {
			return cfgErr("compose", "service", "Name", fmt.Errorf("%w: duplicate service %q", errInvalid, s.Name()))
		}
	}
	p.services = append(p.services, s)
	return nil
}

func (p *Compose) addJob(j *Job) error {
	if j == nil {
		return cfgErr("compose", "job", "Job", fmt.Errorf("%w: job", errRequired))
	}
	j.ensureName()
	for _, existing := range p.jobs {
		if existing.Name() == j.Name() && j.Name() != "" {
			return cfgErr("compose", "job", "Name", fmt.Errorf("%w: duplicate job %q", errInvalid, j.Name()))
		}
	}
	p.jobs = append(p.jobs, j)
	return nil
}

func (p *Compose) hasService(name string) bool {
	for _, s := range p.services {
		if s.Name() == name {
			return true
		}
	}
	return false
}

func (p *Compose) validate() error {
	if p == nil {
		return cfgErr("compose", ".", "", errRequired)
	}
	if !p.name.ok() {
		return cfgErr("compose", "name", "Name", fmt.Errorf("%w: name", errRequired))
	}
	return nil
}

// Validate reports accumulated option errors, nested service/job errors, and
// missing required fields.
func (p *Compose) Validate() error {
	if p == nil {
		return cfgErr("compose", ".", "", errRequired)
	}
	errs := append([]error(nil), p.errs...)
	if err := p.validate(); err != nil {
		errs = addErr(errs, err)
	}
	for _, s := range p.services {
		if err := s.Validate(); err != nil {
			errs = addErr(errs, wrapPath("service."+s.Name(), err))
		}
	}
	for _, j := range p.jobs {
		if err := j.Validate(); err != nil {
			errs = addErr(errs, wrapPath("job."+j.Name(), err))
		}
	}
	return join(errs...)
}

// Name returns the Compose name (top-level YAML `name`).
func (p *Compose) Name() string {
	if p == nil || !p.name.ok() {
		return ""
	}
	return p.name.val
}

// WithName sets the top-level YAML `name`.
func (p *Compose) WithName(name string) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := p.setName(name); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	return p
}

// WithService appends a service and returns Compose for chaining.
func (p *Compose) WithService(s *Service) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if s == nil {
		p.errs = addErr(p.errs, cfgErr("compose", "service", "WithService", fmt.Errorf("%w: service", errRequired)))
		return p
	}
	if err := p.addService(s); err != nil {
		p.errs = addErr(p.errs, err)
	}
	return p
}

// WithJob appends a job and returns Compose for chaining.
func (p *Compose) WithJob(j *Job) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := p.addJob(j); err != nil {
		p.errs = addErr(p.errs, err)
	}
	return p
}

// WithNetwork defines a top-level Compose network.
func (p *Compose) WithNetwork(opts ...ComposeNetOption) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := (AddNetworkOpt{opts: opts}).applyCompose(p); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	return p
}

// WithVolume defines a top-level Compose volume.
func (p *Compose) WithVolume(opts ...ComposeVolOption) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := (AddVolumeOpt{opts: opts}).applyCompose(p); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	return p
}

// WithSecret defines a top-level Compose secret.
func (p *Compose) WithSecret(opts ...ComposeSecretOption) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := (AddSecretOpt{opts: opts}).applyCompose(p); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	return p
}

// WithFile defines a top-level Compose config object.
func (p *Compose) WithFile(opts ...ComposeFileOption) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := (AddFileOpt{opts: opts}).applyCompose(p); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	return p
}

// WithModel defines a top-level Compose model.
func (p *Compose) WithModel(opts ...ComposeModelOption) *Compose {
	if p == nil {
		p = &Compose{}
	}
	if err := (AddModelOpt{opts: opts}).applyCompose(p); err != nil {
		p.errs = addErr(p.errs, wrapReceiver("compose", err))
	}
	return p
}

// Services returns the configured services in definition order.
func (p *Compose) Services() []*Service {
	if p == nil {
		return nil
	}
	out := make([]*Service, len(p.services))
	copy(out, p.services)
	return out
}

// Jobs returns the configured jobs in definition order.
func (p *Compose) Jobs() []*Job {
	if p == nil {
		return nil
	}
	out := make([]*Job, len(p.jobs))
	copy(out, p.jobs)
	return out
}

// AddServiceOpt is the value returned by [AddService].
// It implements [ComposeOption].
type AddServiceOpt struct{ opts []ServiceOption }

// AddService adds a nested service via the option vocabulary so [logic.Group]
// can carry one. Callers building a file should use [Compose.WithService].
func AddService(opts ...ServiceOption) AddServiceOpt { return AddServiceOpt{opts: opts} }

func (o AddServiceOpt) applyCompose(p *Compose) error {
	s := NewService(o.opts...)
	if err := p.addService(s); err != nil {
		return err
	}
	return nil
}

// AddJobOpt is the value returned by [AddJob].
// It implements [ComposeOption].
type AddJobOpt struct{ opts []JobOption }

// AddJob adds a nested job to Compose.
func AddJob(opts ...JobOption) AddJobOpt { return AddJobOpt{opts: opts} }

func (o AddJobOpt) applyCompose(p *Compose) error {
	return p.addJob(NewJob(o.opts...))
}
