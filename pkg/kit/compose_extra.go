package kit

import "fmt"

// LabelFileOpt is the value returned by [LabelFile].
// It implements [JobOption] and [ServiceOption].
type LabelFileOpt struct{ path string }

// LabelFile adds a Compose label_file path. Service and job only.
func LabelFile(path string) LabelFileOpt { return LabelFileOpt{path: path} }

func (o LabelFileOpt) applyService(s *Service) error {
	if o.path == "" {
		return cfgErr("", "label_file", "LabelFile", fmt.Errorf("%w: label file", errEmpty))
	}
	s.labelFiles = append(s.labelFiles, o.path)
	return nil
}
func (o LabelFileOpt) applyJob(j *Job) error {
	if o.path == "" {
		return cfgErr("", "label_file", "LabelFile", fmt.Errorf("%w: label file", errEmpty))
	}
	j.labelFiles = append(j.labelFiles, o.path)
	return nil
}

// UseAPISocketOpt is the value returned by [UseAPISocket].
// It implements [JobOption] and [ServiceOption].
type UseAPISocketOpt struct{ v bool }

// UseAPISocket forwards the Docker socket into the service. Compose-only.
func UseAPISocket(v bool) UseAPISocketOpt { return UseAPISocketOpt{v: v} }
func (o UseAPISocketOpt) applyService(s *Service) error {
	s.useAPISocket.assign(o.v)
	return nil
}
func (o UseAPISocketOpt) applyJob(j *Job) error {
	j.useAPISocket.assign(o.v)
	return nil
}

// AttachOpt is the value returned by [Attach].
// It implements [RunOption] and [ServiceOption].
type AttachOpt struct{ v bool }

// Attach sets Compose attach. Service only.
func Attach(v bool) AttachOpt { return AttachOpt{v: v} }
func (o AttachOpt) applyService(s *Service) error {
	s.attach.assign(o.v)
	return nil
}

// ProviderOpt is the value returned by [Provider].
// It implements [ServiceOption].
type ProviderOpt struct{ t string }

// Provider sets Compose provider.type. Service only.
func Provider(typ string) ProviderOpt { return ProviderOpt{t: typ} }
func (o ProviderOpt) applyService(s *Service) error {
	if o.t == "" {
		return cfgErr("", "provider", "Provider", fmt.Errorf("%w: provider type", errEmpty))
	}
	s.provider.assign(o.t)
	return nil
}

// PostStartOpt is the value returned by [PostStart].
// It implements [ServiceOption].
type PostStartOpt struct{ opts []HookOption }

// PostStart adds a Compose post_start hook. Service only.
func PostStart(opts ...HookOption) PostStartOpt { return PostStartOpt{opts: opts} }

func (o PostStartOpt) applyService(s *Service) error {
	h, err := collectHook("post_start", o.opts)
	if err != nil {
		return err
	}
	s.postStart = append(s.postStart, h)
	return nil
}

// PreStopOpt is the value returned by [PreStop].
// It implements [ServiceOption].
type PreStopOpt struct{ opts []HookOption }

// PreStop adds a Compose pre_stop hook. Service only.
func PreStop(opts ...HookOption) PreStopOpt { return PreStopOpt{opts: opts} }

func (o PreStopOpt) applyService(s *Service) error {
	h, err := collectHook("pre_stop", o.opts)
	if err != nil {
		return err
	}
	s.preStop = append(s.preStop, h)
	return nil
}

func collectHook(kind string, opts []HookOption) (serviceHook, error) {
	var h serviceHook
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("%s.option[%d]", kind, i), "option", errNilOption))
			continue
		}
		if err := opt.applyHook(&h); err != nil {
			errs = append(errs, wrapPath(kind, err))
		}
	}
	if err := join(errs...); err != nil {
		return serviceHook{}, err
	}
	if len(h.cmd) == 0 {
		return serviceHook{}, cfgErr("", kind, "Command", fmt.Errorf("%w: command", errEmpty))
	}
	return h, nil
}

// PreStartOpt is the value returned by [PreStart].
// It implements [ServiceOption].
type PreStartOpt struct{ opts []ContainerOption }

// PreStart adds a Compose pre_start hook using the same container vocabulary.
func PreStart(opts ...ContainerOption) PreStartOpt { return PreStartOpt{opts: opts} }
func (o PreStartOpt) applyService(s *Service) error {
	c := &Container{}
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("pre_start.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyContainer(c); err != nil {
			errs = append(errs, wrapPath("pre_start", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	s.preStart = append(s.preStart, c.spec)
	return nil
}

// CredentialFileOpt is the value returned by [CredentialFile].
// It implements [ServiceOption].
type CredentialFileOpt struct{ p string }

// CredentialRegistryOpt is the value returned by [CredentialRegistry].
// It implements [ServiceOption].
type CredentialRegistryOpt struct{ p string }

// CredentialFile sets Compose credential_spec.file. Service only.
func CredentialFile(path string) CredentialFileOpt { return CredentialFileOpt{p: path} }
func (o CredentialFileOpt) applyService(s *Service) error {
	if o.p == "" {
		return cfgErr("", "credential_spec.file", "CredentialFile", fmt.Errorf("%w: path", errEmpty))
	}
	s.credFile.assign(o.p)
	return nil
}

// CredentialRegistry sets Compose credential_spec.registry. Service only.
func CredentialRegistry(p string) CredentialRegistryOpt { return CredentialRegistryOpt{p: p} }
func (o CredentialRegistryOpt) applyService(s *Service) error {
	if o.p == "" {
		return cfgErr("", "credential_spec.registry", "CredentialRegistry", fmt.Errorf("%w: registry", errEmpty))
	}
	s.credRegistry.assign(o.p)
	return nil
}

// ExtendsOpt is the value returned by [Extends].
// It implements [ServiceOption].
type ExtendsOpt struct{ n string }

// Extends sets Compose extends.service. Service only.
func Extends(name string) ExtendsOpt { return ExtendsOpt{n: name} }
func (o ExtendsOpt) applyService(s *Service) error {
	if err := validateName(o.n); err != nil {
		return wrapPath("extends", err)
	}
	s.extends.assign(o.n)
	return nil
}
