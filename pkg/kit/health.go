package kit

import (
	"fmt"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/docker/docker/api/types/container"
)

// Health is a fluent container healthcheck.
// Pass it to Container.HealthCheck and Dockerfile.Healthcheck.
// Timeout/Interval/StartPeriod are seconds unless the matching *D method is used ("2s", "1m").
type Health struct {
	test         []string
	timeout      int
	interval     int
	startPeriod  int
	retries      int
	timeoutD     string
	intervalD    string
	startPeriodD string
}

// NewHealth starts an empty healthcheck.
func NewHealth() Health {
	return Health{}
}

// Test sets the healthcheck command, including the CMD / CMD-SHELL / NONE prefix.
func (h Health) Test(args ...string) Health {
	h.test = append([]string(nil), args...)
	return h
}

// CMD sets a CMD healthcheck (exec form).
func (h Health) CMD(args ...string) Health {
	return h.Test(append([]string{"CMD"}, args...)...)
}

// CMDShell sets a CMD-SHELL healthcheck.
func (h Health) CMDShell(cmd string) Health {
	return h.Test(append([]string{"CMD-SHELL"}, cmd)...)
}

// None disables the healthcheck (HEALTHCHECK NONE).
func (h Health) None() Health {
	return h.Test("NONE")
}

// Timeout sets the healthcheck timeout in seconds.
func (h Health) Timeout(seconds int) Health {
	h.timeout = seconds
	return h
}

// TimeoutD sets the healthcheck timeout from a duration string ("5s").
func (h Health) TimeoutD(d string) Health {
	h.timeoutD = d
	return h
}

// Interval sets the healthcheck interval in seconds.
func (h Health) Interval(seconds int) Health {
	h.interval = seconds
	return h
}

// IntervalD sets the healthcheck interval from a duration string ("10s").
func (h Health) IntervalD(d string) Health {
	h.intervalD = d
	return h
}

// StartPeriod sets the start-period in seconds.
func (h Health) StartPeriod(seconds int) Health {
	h.startPeriod = seconds
	return h
}

// StartPeriodD sets the start-period from a duration string ("0s").
func (h Health) StartPeriodD(d string) Health {
	h.startPeriodD = d
	return h
}

// Retries sets the number of consecutive failures needed to mark the container unhealthy.
func (h Health) Retries(n int) Health {
	h.retries = n
	return h
}

func (h Health) healthConfig() (*container.HealthConfig, error) {
	timeout, err := healthDuration(h.timeout, h.timeoutD, "timeout")
	if err != nil {
		return nil, err
	}
	interval, err := healthDuration(h.interval, h.intervalD, "interval")
	if err != nil {
		return nil, err
	}
	start, err := healthDuration(h.startPeriod, h.startPeriodD, "start period")
	if err != nil {
		return nil, err
	}
	if h.retries < 0 {
		return nil, errdefs.NewContainerConfigError("healthcheck", "retries must be greater than or equal to 0")
	}
	test := h.test
	if len(test) == 0 {
		test = []string{"NONE"}
	}
	return &container.HealthConfig{
		Test:        test,
		Timeout:     timeout,
		Interval:    interval,
		StartPeriod: start,
		Retries:     h.retries,
	}, nil
}

func healthDuration(seconds int, raw, field string) (time.Duration, error) {
	var duration time.Duration
	if raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return 0, errdefs.NewContainerConfigError("healthcheck", fmt.Sprintf("error parsing %s: %s", field, err))
		}
		duration = parsed
	} else {
		duration = time.Duration(seconds) * time.Second
	}
	if duration < 0 {
		return 0, errdefs.NewContainerConfigError("healthcheck", field+" must be non-negative")
	}
	return duration, nil
}
