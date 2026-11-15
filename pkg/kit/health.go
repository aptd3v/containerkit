package kit

import (
	"fmt"
	"time"
)

func (h *health) setTest(cmd []string) error {
	if len(cmd) == 0 {
		return cfgErr("", "test", "Test", fmt.Errorf("%w: health test", errEmpty))
	}
	h.test.assign(append([]string(nil), cmd...))
	return nil
}

func (h *health) setInterval(d time.Duration) error {
	if d < 0 {
		return cfgErr("", "interval", "Interval", fmt.Errorf("%w: interval", errInvalid))
	}
	h.interval.assign(d)
	return nil
}

func (h *health) setTimeout(d time.Duration) error {
	if d < 0 {
		return cfgErr("", "timeout", "Timeout", fmt.Errorf("%w: timeout", errInvalid))
	}
	h.timeout.assign(d)
	return nil
}

func (h *health) setRetries(n int) error {
	if n < 0 {
		return cfgErr("", "retries", "Retries", fmt.Errorf("%w: retries", errInvalid))
	}
	h.retries.assign(n)
	return nil
}

func (h *health) setStartPeriod(d time.Duration) error {
	if d < 0 {
		return cfgErr("", "start_period", "StartPeriod", fmt.Errorf("%w: start period", errInvalid))
	}
	h.startPeriod.assign(d)
	return nil
}

func (h *health) setDisabled(v bool) error {
	h.disabled.assign(v)
	return nil
}

// TestOpt is the value returned by [Test].
// It implements [HealthOption].
type TestOpt struct{ cmd []string }

// IntervalOpt is the value returned by [Interval].
// It implements [HealthOption].
type IntervalOpt struct{ d time.Duration }

// TimeoutOpt is the value returned by [Timeout].
// It implements [ClientOption], [ComposeCreateOption], [ComposeRestartOption],
// [ComposeStartOption], [ComposeStopOption], [ComposerClientOption], [DownOption],
// [EnginePluginEnableOption], [EngineStopOption], [HealthOption], [RunOption], and [UpOption].
type TimeoutOpt struct{ d time.Duration }

// RetriesOpt is the value returned by [Retries].
// It implements [HealthOption].
type RetriesOpt struct{ n int }

// StartPeriodOpt is the value returned by [StartPeriod].
// It implements [HealthOption].
type StartPeriodOpt struct{ d time.Duration }

// DisabledOpt is the value returned by [Disabled].
// It implements [EnginePluginInstallOption], [EnginePluginUpgradeOption], and [HealthOption].
type DisabledOpt struct{ v bool }

// Test sets the healthcheck command in a [Health] scope. Extra strings are
// command arguments, not interval or retries; those are [Interval], [Timeout],
// and [Retries].
func Test(cmd ...string) TestOpt { return TestOpt{cmd: cmd} }

func (o TestOpt) applyHealth(h *health) error { return h.setTest(o.cmd) }

// Interval sets the healthcheck interval in a [Health] scope. Zero is explicit.
func Interval(d time.Duration) IntervalOpt { return IntervalOpt{d: d} }

func (o IntervalOpt) applyHealth(h *health) error { return h.setInterval(o.d) }

// Timeout sets a duration for healthchecks, compose wait/stop, Engine stop,
// and [EnablePlugin] (seconds). Zero is explicit.
func Timeout(d time.Duration) TimeoutOpt { return TimeoutOpt{d: d} }

func (o TimeoutOpt) applyHealth(h *health) error { return h.setTimeout(o.d) }

// Retries sets healthcheck retries in a [Health] scope. Zero is explicit.
func Retries(n int) RetriesOpt { return RetriesOpt{n: n} }

func (o RetriesOpt) applyHealth(h *health) error { return h.setRetries(o.n) }

// StartPeriod sets the healthcheck start period in a [Health] scope.
func StartPeriod(d time.Duration) StartPeriodOpt { return StartPeriodOpt{d: d} }

func (o StartPeriodOpt) applyHealth(h *health) error { return h.setStartPeriod(o.d) }

// Disabled disables a healthcheck, or leaves a plugin disabled after
// [InstallPlugin] / [UpgradePlugin].
func Disabled(v bool) DisabledOpt { return DisabledOpt{v: v} }

func (o DisabledOpt) applyHealth(h *health) error { return h.setDisabled(o.v) }

// HealthOpt is the value returned by [Health].
// It implements [ContainerOption] and [ServiceOption].
type HealthOpt struct{ opts []HealthOption }

// Health configures a healthcheck with nested setters.
func Health(opts ...HealthOption) HealthOpt { return HealthOpt{opts: opts} }

func (o HealthOpt) applyContainer(c *Container) error { return o.apply(&c.spec) }
func (o HealthOpt) applyService(s *Service) error     { return o.apply(&s.spec) }

func (o HealthOpt) apply(s *spec) error {
	var h health
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = append(errs, cfgErr("", fmt.Sprintf("health.option[%d]", i), "option", errNilOption))
			continue
		}
		if err := opt.applyHealth(&h); err != nil {
			errs = append(errs, wrapPath("health", err))
		}
	}
	if err := join(errs...); err != nil {
		return err
	}
	return wrapPath("health", s.setHealth(h))
}
