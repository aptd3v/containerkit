package kit

import "fmt"

func (s *secretRef) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("secret", err)
	}
	s.source.assign(name)
	return nil
}

func (s *secretRef) setSource(v string) error {
	if v == "" {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: secret source", errEmpty))
	}
	s.source.assign(v)
	return nil
}

func (s *secretRef) setTarget(v string) error {
	s.target.assign(v)
	return nil
}

func (s *secretRef) setFile(p string) error {
	s.file.assign(p)
	return nil
}

func (s *secretRef) setDriver(d string) error {
	s.driver.assign(d)
	return nil
}

func (s *secretRef) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if s.labels == nil {
		s.labels = map[string]string{}
	}
	s.labels[k] = val
	return nil
}

func (c *fileConfigRef) setName(name string) error {
	if err := validateName(name); err != nil {
		return wrapPath("config", err)
	}
	c.source.assign(name)
	return nil
}

func (c *fileConfigRef) setSource(v string) error {
	if v == "" {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: config source", errEmpty))
	}
	c.source.assign(v)
	return nil
}

func (c *fileConfigRef) setTarget(v string) error {
	c.target.assign(v)
	return nil
}

func (c *fileConfigRef) setFile(p string) error {
	c.file.assign(p)
	return nil
}

func (c *fileConfigRef) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if c.labels == nil {
		c.labels = map[string]string{}
	}
	c.labels[k] = val
	return nil
}

// SecretOpt is the value returned by [Secret].
// It implements [BuildOption], [JobOption], and [ServiceOption].
type SecretOpt struct{ opts []SecretOption }

// Secret adds a Compose secret reference on a service/job.
func Secret(opts ...SecretOption) SecretOpt { return SecretOpt{opts: opts} }

func (o SecretOpt) applyService(s *Service) error { return o.apply(&s.secrets) }
func (o SecretOpt) applyJob(j *Job) error         { return o.apply(&j.secrets) }

func (o SecretOpt) apply(dst *[]secretRef) error {
	var s secretRef
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("secret.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applySecret(&s); err != nil {
			errs = append(errs, wrapPath("secret", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !s.source.ok() {
		return cfgErr("", "secret.source", "Name", fmt.Errorf("%w: secret source", errRequired))
	}
	*dst = append(*dst, s)
	return nil
}

// FileConfigOpt is the value returned by [FileConfig].
// It implements [JobOption] and [ServiceOption].
type FileConfigOpt struct{ opts []FileConfigOption }

// FileConfig adds a Compose config-file reference on a service/job.
// Named FileConfig because Config is the Engine portable config type.
func FileConfig(opts ...FileConfigOption) FileConfigOpt { return FileConfigOpt{opts: opts} }

func (o FileConfigOpt) applyService(s *Service) error { return o.apply(&s.configs) }
func (o FileConfigOpt) applyJob(j *Job) error         { return o.apply(&j.configs) }

func (o FileConfigOpt) apply(dst *[]fileConfigRef) error {
	var c fileConfigRef
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("config.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyFileConfig(&c); err != nil {
			errs = append(errs, wrapPath("config", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if !c.source.ok() {
		return cfgErr("", "config.source", "Name", fmt.Errorf("%w: config source", errRequired))
	}
	*dst = append(*dst, c)
	return nil
}
