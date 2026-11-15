package kit

import "fmt"

func (b *build) setSource(s string) error {
	if s == "" {
		return cfgErr("", "source", "Source", fmt.Errorf("%w: build context", errEmpty))
	}
	b.source.assign(s)
	return nil
}

func (b *build) setTarget(t string) error {
	b.target.assign(t)
	return nil
}

func (b *build) setDockerfile(p string) error {
	if p == "" {
		return cfgErr("", "file", "File", fmt.Errorf("%w: dockerfile", errEmpty))
	}
	b.dockerfile.assign(p)
	return nil
}

func (b *build) setPrivileged(v bool) error {
	b.privileged.assign(v)
	return nil
}

func (b *build) setShmSize(n int64) error {
	if n < 0 {
		return cfgErr("", "memory", "Memory", fmt.Errorf("%w: build shm size", errInvalid))
	}
	b.shmSize.assign(n)
	return nil
}

func (b *build) addArg(key, value string) error {
	if key == "" {
		return secretErr("", "env", "Env", fmt.Errorf("%w: build arg key", errEmpty))
	}
	b.args = append(b.args, envVar{key: key, value: field[string]{set: true, val: value}})
	return nil
}

func (b *build) addLabel(k, val string) error {
	if k == "" {
		return cfgErr("", "label", "Label", fmt.Errorf("%w: label key", errEmpty))
	}
	if b.labels == nil {
		b.labels = map[string]string{}
	}
	b.labels[k] = val
	return nil
}

// PullOpt is the value returned by [Pull].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption],
// [EngineBuildOption], [RunOption], [UpOption], and [WatchOption].
type PullOpt struct{ v bool }

// Pull requests a fresh base image inside [Build].
func Pull(v bool) PullOpt { return PullOpt{v: v} }

func (o PullOpt) applyBuild(b *build) error {
	b.pull.assign(o.v)
	return nil
}

// NoCacheOpt is the value returned by [NoCache].
// It implements [BuildOption], [ComposeBuildOption], [ComposeCreateOption],
// [EngineBuildOption], [RunOption], [UpOption], and [WatchOption].
type NoCacheOpt struct{ v bool }

// NoCache disables layer cache inside [Build].
func NoCache(v bool) NoCacheOpt { return NoCacheOpt{v: v} }

func (o NoCacheOpt) applyBuild(b *build) error {
	b.noCache.assign(o.v)
	return nil
}

// TagOpt is the value returned by [Tag].
// It implements [BuildOption] and [EngineBuildOption].
type TagOpt struct{ t string }

// Tag adds an image tag inside [Build].
func Tag(t string) TagOpt { return TagOpt{t: t} }

func (o TagOpt) applyBuild(b *build) error {
	if o.t == "" {
		return cfgErr("", "tag", "Tag", fmt.Errorf("%w: tag", errEmpty))
	}
	b.tags = append(b.tags, o.t)
	return nil
}

// BuildOpt is the value returned by [Build].
// It implements [JobOption] and [ServiceOption].
type BuildOpt struct{ opts []BuildOption }

// Build sets Compose build configuration. Service and job only.
func Build(opts ...BuildOption) BuildOpt { return BuildOpt{opts: opts} }

func (o BuildOpt) applyService(s *Service) error { return o.apply(s) }
func (o BuildOpt) applyJob(j *Job) error {
	var b build
	if err := o.collect(&b); err != nil {
		return err
	}
	j.build.assign(b)
	return nil
}

func (o BuildOpt) apply(s *Service) error {
	var b build
	if err := o.collect(&b); err != nil {
		return err
	}
	s.build.assign(b)
	return nil
}

func (o BuildOpt) collect(b *build) error {
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("build.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyBuild(b); err != nil {
			errs = append(errs, wrapPath("build", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	if b.inline.ok() && b.dockerfile.ok() {
		return cfgErr("", "build.file", "File", fmt.Errorf("%w: File and Dockerfile cannot both be set", errInvalid))
	}
	if !b.source.ok() {
		if b.inline.ok() {
			b.source.assign(".")
		} else {
			return cfgErr("", "build.source", "Source", fmt.Errorf("%w: build context", errRequired))
		}
	}
	return nil
}

// PullPolicyOpt is the value returned by [PullPolicy].
// It implements [JobOption] and [ServiceOption].
type PullPolicyOpt struct{ p PullMode }

// PullPolicy sets Compose pull_policy. Service and job only.
func PullPolicy(p PullMode) PullPolicyOpt { return PullPolicyOpt{p: p} }

func (o PullPolicyOpt) applyService(s *Service) error {
	s.pullPolicy.assign(string(o.p))
	return nil
}
func (o PullPolicyOpt) applyJob(j *Job) error {
	j.pullPolicy.assign(string(o.p))
	return nil
}

// EnvFileOpt is the value returned by [EnvFile].
// It implements [IncludeOption], [JobOption], [LoadProjectOption], and [ServiceOption].
type EnvFileOpt struct {
	path string
	opts []EnvFileOption
}

// EnvFile adds a Compose env_file path. Service, job, [AddInclude], and [LoadProject].
func EnvFile(path string, opts ...EnvFileOption) EnvFileOpt {
	return EnvFileOpt{path: path, opts: opts}
}

func (o EnvFileOpt) applyService(s *Service) error {
	ref, err := o.ref()
	if err != nil {
		return err
	}
	s.envFiles = append(s.envFiles, ref)
	return nil
}
func (o EnvFileOpt) applyJob(j *Job) error {
	ref, err := o.ref()
	if err != nil {
		return err
	}
	j.envFiles = append(j.envFiles, ref)
	return nil
}

func (o EnvFileOpt) applyInclude(inc *includeRef) error {
	if o.path == "" {
		return cfgErr("", "env_file", "EnvFile", fmt.Errorf("%w: env file", errEmpty))
	}
	inc.envFiles = append(inc.envFiles, o.path)
	return nil
}

func (o EnvFileOpt) ref() (envFileRef, error) {
	if o.path == "" {
		return envFileRef{}, cfgErr("", "env_file", "EnvFile", fmt.Errorf("%w: env file", errEmpty))
	}
	ref := envFileRef{}
	ref.path.assign(o.path)
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("env_file.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyEnvFile(&ref); err != nil {
			errs = append(errs, wrapPath("env_file", err))
		}
	}
	if err := join(errs...); err != nil {
		return envFileRef{}, err
	}
	return ref, nil
}
