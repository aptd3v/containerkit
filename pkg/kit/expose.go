package kit

import "fmt"

// ExposeOpt is the value returned by [Expose].
// It implements [ContainerOption], [JobOption], and [ServiceOption].
type ExposeOpt struct {
	port  uint16
	proto Proto
}

// Expose publishes a port without a host binding (Engine ExposedPorts / Compose expose).
func Expose(port uint16, proto ...Proto) ExposeOpt {
	p := TCP
	if len(proto) > 0 {
		p = proto[0]
	}
	return ExposeOpt{port: port, proto: p}
}

func (o ExposeOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o ExposeOpt) applyService(s *Service) error     { return o.apply(&s.spec) }
func (o ExposeOpt) applyJob(j *Job) error             { return o.apply(&j.spec) }

func (o ExposeOpt) apply(s *spec) error {
	if o.port == 0 {
		return cfgErr("", "expose.port", "Expose", fmt.Errorf("%w: expose port", errRequired))
	}
	return s.addExpose(exposePort{port: o.port, protocol: o.proto})
}
