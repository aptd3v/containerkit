package kit

import "fmt"

var _ ServiceOption = (*Container)(nil)

// FromContainerOpt is the value returned by [FromContainer].
// It implements [ServiceOption].
type FromContainerOpt struct{ c *Container }

// FromContainer copies compatible container options onto a Compose service.
// Engine-only settings such as [AutoRemove] are reported as unsupported mappings.
// Passing a [*Container] directly to [NewService] is equivalent.
func FromContainer(c *Container) FromContainerOpt {
	return FromContainerOpt{c: c}
}

func (o FromContainerOpt) applyService(s *Service) error {
	return o.c.applyService(s)
}

func (c *Container) applyService(s *Service) error {
	if c == nil {
		return cfgErr("service", "container", "Container", fmt.Errorf("%w: container", errRequired))
	}
	s.errs = append(s.errs, c.errs...)
	existing := s.name
	copied := c.spec
	if existing.ok() {
		copied.name = existing
	}
	s.spec = copied
	if err := c.engineOnlyConflict(); err != nil {
		return wrapReceiver("service", err)
	}
	return nil
}

// AsService converts a container into a Compose service. Extra service options
// apply after the container is copied.
func (c *Container) AsService(opts ...ServiceOption) *Service {
	all := make([]ServiceOption, 0, 1+len(opts))
	all = append(all, c)
	all = append(all, opts...)
	return NewService(all...)
}
