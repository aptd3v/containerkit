package kit

import "fmt"

// Container is a standalone Engine container configuration that also converts
// to a Compose service when the settings map.
type Container struct {
	spec
	autoRemove    field[bool]
	stdinOnce     field[bool]
	argsEscaped   field[bool]
	attachStdin   field[bool]
	attachStdout  field[bool]
	attachStderr  field[bool]
	publishAll    field[bool]
	idFile        field[string]
	shell         []string
	onBuild       []string
	maskedPaths   []string
	readonlyPaths []string
	errs          []error
}

// NewContainer applies typed options and always returns a container.
// Option and validation errors accumulate; call [Container.Validate] to inspect them.
func NewContainer(opts ...ContainerOption) *Container {
	c := &Container{}
	for i, opt := range opts {
		if opt == nil {
			c.errs = addErr(c.errs, cfgErr("container", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyContainer(c); err != nil {
			c.errs = addErr(c.errs, wrapReceiver("container", err))
		}
	}
	return c
}

func (c *Container) setAutoRemove(v bool) error {
	c.autoRemove.assign(v)
	return nil
}

func (c *Container) validate() error {
	if c == nil {
		return cfgErr("container", ".", "", errRequired)
	}
	if !c.image.ok() {
		return cfgErr("container", "image", "Image", fmt.Errorf("%w: image", errRequired))
	}
	return nil
}

// Validate reports accumulated option errors and missing required fields.
func (c *Container) Validate() error {
	if c == nil {
		return cfgErr("container", ".", "", errRequired)
	}
	return join(append(append([]error(nil), c.errs...), c.validate())...)
}

// Name returns the configured container name, or empty if omitted.
func (c *Container) Name() string {
	if c == nil || !c.name.ok() {
		return ""
	}
	return c.name.val
}

// Image returns the configured image, or empty if omitted.
func (c *Container) Image() string {
	if c == nil || !c.image.ok() {
		return ""
	}
	return c.image.val
}

// Apply adds options to an existing container. Errors join the container's
// accumulated set and are also returned.
func (c *Container) Apply(opts ...ContainerOption) error {
	if c == nil {
		return cfgErr("container", ".", "Apply", fmt.Errorf("%w: container", errRequired))
	}
	var errs []error
	for i, opt := range opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("container", fmt.Sprintf("option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyContainer(c); err != nil {
			errs = addErr(errs, wrapReceiver("container", err))
		}
	}
	c.errs = append(c.errs, errs...)
	return join(errs...)
}
