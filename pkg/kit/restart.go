package kit

import "fmt"

func (r *restart) setPolicy(p RestartPolicy) error {
	switch p {
	case NoRestart, OnFailure, Always, UnlessStopped:
		r.policy.assign(p)
		return nil
	default:
		return cfgErr("", "policy", "Policy", fmt.Errorf("%w: restart policy %q", errInvalid, p))
	}
}

func (r *restart) setMaxRetries(n int) error {
	if n < 0 {
		return cfgErr("", "max_retries", "MaxRetries", fmt.Errorf("%w: max retries", errInvalid))
	}
	r.maxRetries.assign(n)
	return nil
}

// PolicyOpt is the value returned by [Policy].
// It implements [RestartOption].
type PolicyOpt struct{ p RestartPolicy }

// MaxRetriesOpt is the value returned by [MaxRetries].
// It implements [RestartOption].
type MaxRetriesOpt struct{ n int }

// Policy sets the restart policy name in a [Restart] scope.
func Policy(p RestartPolicy) PolicyOpt { return PolicyOpt{p: p} }

func (o PolicyOpt) applyRestart(r *restart) error { return r.setPolicy(o.p) }

// MaxRetries sets on-failure retry count in a [Restart] scope. Zero is explicit.
func MaxRetries(n int) MaxRetriesOpt { return MaxRetriesOpt{n: n} }

func (o MaxRetriesOpt) applyRestart(r *restart) error { return r.setMaxRetries(o.n) }

// RestartOpt is the value returned by [Restart].
// It implements [ContainerOption], [DeployOption], and [ServiceOption].
type RestartOpt struct{ opts []RestartOption }

// Restart configures the restart policy with nested setters.
func Restart(opts ...RestartOption) RestartOpt { return RestartOpt{opts: opts} }

func (o RestartOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o RestartOpt) applyService(s *Service) error     { return o.apply(&s.spec) }

func (o RestartOpt) apply(s *spec) error {
	var r restart
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("restart.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyRestart(&r); err != nil {
			errs = append(errs, wrapPath("restart", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	return wrapPath("restart", s.setRestart(r))
}
