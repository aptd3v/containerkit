package kit

import "fmt"

func (c *Container) setStdinOnce(v bool) error {
	c.stdinOnce.assign(v)
	return nil
}

func (c *Container) setArgsEscaped(v bool) error {
	c.argsEscaped.assign(v)
	return nil
}

func (c *Container) setPublishAll(v bool) error {
	c.publishAll.assign(v)
	return nil
}

func (c *Container) setIDFile(p string) error {
	if p == "" {
		return cfgErr("", "container_id_file", "ContainerIDFile", fmt.Errorf("%w: path", errEmpty))
	}
	c.idFile.assign(p)
	return nil
}

// StdinOnceOpt is the value returned by [StdinOnce].
// It implements [ContainerOption].
type StdinOnceOpt struct{ v bool }

// StdinOnce closes stdin after one attached client disconnects. Engine-only.
func StdinOnce(v bool) StdinOnceOpt                      { return StdinOnceOpt{v: v} }
func (o StdinOnceOpt) applyContainer(c *Container) error { return c.setStdinOnce(o.v) }

// ArgsEscapedOpt is the value returned by [ArgsEscaped].
// It implements [ContainerOption].
type ArgsEscapedOpt struct{ v bool }

// ArgsEscaped marks the command as already escaped (Windows). Engine-only.
func ArgsEscaped(v bool) ArgsEscapedOpt                    { return ArgsEscapedOpt{v: v} }
func (o ArgsEscapedOpt) applyContainer(c *Container) error { return c.setArgsEscaped(o.v) }

// AttachStdin attaches stdin. Engine-only.
func AttachStdin(v bool) AttachStdinOpt { return AttachStdinOpt{v: v} }

// AttachStdinOpt is the value returned by [AttachStdin].
// It implements [ContainerOption], [EngineAttachOption], and [EngineExecCreateOption].
type AttachStdinOpt struct{ v bool }

// AttachStdoutOpt is the value returned by [AttachStdout].
// It implements [ContainerOption], [EngineAttachOption], [EngineExecCreateOption], and
// [LogsOption].
type AttachStdoutOpt struct{ v bool }

// AttachStderrOpt is the value returned by [AttachStderr].
// It implements [ContainerOption], [EngineAttachOption], [EngineExecCreateOption], and
// [LogsOption].
type AttachStderrOpt struct{ v bool }

func (o AttachStdinOpt) applyContainer(c *Container) error {
	c.attachStdin.assign(o.v)
	return nil
}
func (o AttachStdinOpt) applyEngineAttachOption(a *engineAttachAct) error {
	a.opts.Stdin = o.v
	return nil
}
func (o AttachStdinOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.AttachStdin = o.v
	return nil
}

// AttachStdout attaches stdout. Engine-only.
func AttachStdout(v bool) AttachStdoutOpt { return AttachStdoutOpt{v: v} }
func (o AttachStdoutOpt) applyContainer(c *Container) error {
	c.attachStdout.assign(o.v)
	return nil
}
func (o AttachStdoutOpt) applyEngineAttachOption(a *engineAttachAct) error {
	a.opts.Stdout = o.v
	return nil
}
func (o AttachStdoutOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.AttachStdout = o.v
	return nil
}

// AttachStderr attaches stderr. Engine-only.
func AttachStderr(v bool) AttachStderrOpt { return AttachStderrOpt{v: v} }
func (o AttachStderrOpt) applyContainer(c *Container) error {
	c.attachStderr.assign(o.v)
	return nil
}
func (o AttachStderrOpt) applyEngineAttachOption(a *engineAttachAct) error {
	a.opts.Stderr = o.v
	return nil
}
func (o AttachStderrOpt) applyEngineExecCreateOption(a *engineExecCreateAct) error {
	a.opts.AttachStderr = o.v
	return nil
}

// PublishAllOpt is the value returned by [PublishAll].
// It implements [ContainerOption].
type PublishAllOpt struct{ v bool }

// PublishAll publishes all exposed ports on random host ports. Engine-only.
func PublishAll(v bool) PublishAllOpt                     { return PublishAllOpt{v: v} }
func (o PublishAllOpt) applyContainer(c *Container) error { return c.setPublishAll(o.v) }

// ContainerIDFileOpt is the value returned by [ContainerIDFile].
// It implements [ContainerOption].
type ContainerIDFileOpt struct{ p string }

// ContainerIDFile writes the container ID to a file. Engine-only.
func ContainerIDFile(p string) ContainerIDFileOpt              { return ContainerIDFileOpt{p: p} }
func (o ContainerIDFileOpt) applyContainer(c *Container) error { return c.setIDFile(o.p) }

// ShellOpt is the value returned by [Shell].
// It implements [ContainerOption].
type ShellOpt struct{ cmd []string }

// Shell sets the shell used for shell-form RUN/CMD/ENTRYPOINT. Engine-only.
func Shell(cmd ...string) ShellOpt { return ShellOpt{cmd: cmd} }
func (o ShellOpt) applyContainer(c *Container) error {
	if len(o.cmd) == 0 {
		return cfgErr("", "shell", "Shell", fmt.Errorf("%w: shell", errEmpty))
	}
	c.shell = append([]string(nil), o.cmd...)
	return nil
}

// OnBuildOpt is the value returned by [OnBuild].
// It implements [ContainerOption].
type OnBuildOpt struct{ args []string }

// OnBuild adds image ONBUILD metadata. Engine-only.
func OnBuild(args ...string) OnBuildOpt { return OnBuildOpt{args: args} }
func (o OnBuildOpt) applyContainer(c *Container) error {
	c.onBuild = append(c.onBuild, o.args...)
	return nil
}

// MaskedPathOpt is the value returned by [MaskedPath].
// It implements [ContainerOption].
type MaskedPathOpt struct{ p string }

// MaskedPath overrides a masked path. Engine-only.
func MaskedPath(p string) MaskedPathOpt { return MaskedPathOpt{p: p} }
func (o MaskedPathOpt) applyContainer(c *Container) error {
	if o.p == "" {
		return cfgErr("", "masked_path", "MaskedPath", fmt.Errorf("%w: path", errEmpty))
	}
	c.maskedPaths = append(c.maskedPaths, o.p)
	return nil
}

// ReadonlyPathOpt is the value returned by [ReadonlyPath].
// It implements [ContainerOption].
type ReadonlyPathOpt struct{ p string }

// ReadonlyPath overrides a readonly path. Engine-only.
func ReadonlyPath(p string) ReadonlyPathOpt { return ReadonlyPathOpt{p: p} }
func (o ReadonlyPathOpt) applyContainer(c *Container) error {
	if o.p == "" {
		return cfgErr("", "readonly_path", "ReadonlyPath", fmt.Errorf("%w: path", errEmpty))
	}
	c.readonlyPaths = append(c.readonlyPaths, o.p)
	return nil
}

func (c *Container) engineOnlyConflict() error {
	if c.autoRemove.ok() {
		return cfgErr("", "auto_remove", "AutoRemove", fmt.Errorf("%w: AutoRemove has no Compose equivalent", errUnsupported))
	}
	if c.publishAll.ok() && c.publishAll.val {
		return cfgErr("", "publish_all", "PublishAll", fmt.Errorf("%w: PublishAll has no Compose equivalent", errUnsupported))
	}
	if c.idFile.ok() {
		return cfgErr("", "container_id_file", "ContainerIDFile", fmt.Errorf("%w: ContainerIDFile has no Compose equivalent", errUnsupported))
	}
	if len(c.shell) > 0 {
		return cfgErr("", "shell", "Shell", fmt.Errorf("%w: Shell has no Compose equivalent", errUnsupported))
	}
	if len(c.onBuild) > 0 {
		return cfgErr("", "on_build", "OnBuild", fmt.Errorf("%w: OnBuild has no Compose equivalent", errUnsupported))
	}
	if len(c.maskedPaths) > 0 {
		return cfgErr("", "masked_path", "MaskedPath", fmt.Errorf("%w: MaskedPath has no Compose equivalent", errUnsupported))
	}
	if len(c.readonlyPaths) > 0 {
		return cfgErr("", "readonly_path", "ReadonlyPath", fmt.Errorf("%w: ReadonlyPath has no Compose equivalent", errUnsupported))
	}
	return nil
}
