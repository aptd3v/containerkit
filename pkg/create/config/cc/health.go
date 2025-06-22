package cc

import (
	"fmt"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/docker/docker/api/types/container"
)

// WithHealthCheck sets the health check for the container
// parameters:
//   - healthCheckFns: the health check functions to set
func WithHealthCheck(setters ...health.SetHealthcheckConfig) create.SetContainerConfig {
	return func(opt *create.ContainerConfig) error {
		if opt.Healthcheck == nil {
			opt.Healthcheck = &container.HealthConfig{}
		}
		for _, setter := range setters {
			if setter == nil {
				continue
			}
			if err := setter(opt.Healthcheck); err != nil {
				return create.NewContainerConfigError("healthcheck", fmt.Sprintf("failed to set health check: %s", err))
			}
		}
		return nil
	}
}

// WithTCPHealthCheck checks if a TCP port is open using netcat.
// it has a default start period of 5 seconds, a timeout of 5 seconds, a retries of 3,
// and an interval of the parameter interval
// parameters:
//   - host: the host to check
//   - port: the port to check
//   - interval: the interval in seconds between health checks
//
// "CMD-SHELL", "nc -z "+host+" "+port+" || exit 1"
func WithTCPHealthCheck(host string, port string, interval int) create.SetContainerConfig {
	cmd := fmt.Sprintf("nc -z %s %s || exit 1", host, port)
	return WithHealthCheck(
		health.WithHealthCheckTest("CMD-SHELL", cmd),
		health.WithHealthCheckStartPeriod(5),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(5),
		health.WithHealthCheckRetries(3),
	)
}

// WithHTTPHealthCheck allows setting method and advanced HTTP health checking.
// it has a default start period of 5 seconds, a timeout of 10 seconds, a retries of 5,
// and an interval of the parameter interval
// parameters:
//   - url: the URL to check
//   - method: the HTTP method to use
//   - interval: the interval in seconds between health checks
func WithHTTPHealthCheck(url, method string, interval int) create.SetContainerConfig {
	cmd := fmt.Sprintf(`curl -X %s -f "%s" || exit 1`, method, url)
	return WithHealthCheck(
		health.WithHealthCheckTest("CMD-SHELL", cmd),
		health.WithHealthCheckStartPeriod(5),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(10),
		health.WithHealthCheckRetries(5),
	)
}

// WithCurlHealthCheck sets a health check that uses curl to check if the container is healthy
// it has a default start period of 5 seconds, a timeout of 10 seconds, a retries of 5,
// and an interval of the parameter interval
// parameters:
//   - url: the URL to check
//   - interval: the interval in seconds between health checks
//
// "CMD-SHELL", "curl -f "+url+" || exit 1"
func WithCurlHealthCheck(url string, interval int) create.SetContainerConfig {
	return WithHealthCheck(
		health.WithHealthCheckTest("CMD-SHELL", `curl -f "`+url+`" || exit 1`),
		health.WithHealthCheckStartPeriod(5),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(10),
		health.WithHealthCheckRetries(5),
	)
}

// WithCommandHealthCheck executes an arbitrary shell command.
// it has a default start period of 5 seconds, a timeout of 5 seconds, a retries of 3,
// and an interval of the parameter interval
// parameters:
//   - cmds: the commands ran in the health check
//   - interval: the interval in seconds between health checks
//
// "CMD-SHELL", cmds...
func WithCommandHealthCheck(interval int, cmds ...string) create.SetContainerConfig {
	return WithHealthCheck(
		health.WithHealthCheckTest(append([]string{"CMD-SHELL"}, cmds...)...),
		health.WithHealthCheckStartPeriod(5),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(5),
		health.WithHealthCheckRetries(3),
	)
}

// WithDockerSocketHealthCheck ensures /var/run/docker.sock exists.
// it has a default start period of 5 seconds, a timeout of 5 seconds, a retries of 3,
// and an interval of the parameter interval
// parameters:
//   - interval: the interval in seconds between health checks
//
// "CMD-SHELL", "[ -S /var/run/docker.sock ] || exit 1"
func WithDockerSocketHealthCheck(interval int) create.SetContainerConfig {
	cmd := `[ -S /var/run/docker.sock ] || exit 1`
	return WithHealthCheck(
		health.WithHealthCheckTest("CMD-SHELL", cmd),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(5),
		health.WithHealthCheckStartPeriod(5),
		health.WithHealthCheckRetries(3),
	)
}

// WithFileExistsHealthCheck checks if a specific file exists.
// it has a default start period of 5 seconds, a timeout of 5 seconds, a retries of 3,
// and an interval of the parameter interval
// parameters:
//   - filePath: the file path to check
//   - interval: the interval in seconds between health checks
//
// "CMD-SHELL", "[ -f "+filePath+" ] || exit 1"
func WithFileExistsHealthCheck(filePath string, interval int) create.SetContainerConfig {
	cmd := fmt.Sprintf(`[ -f "%s" ] || exit 1`, filePath)
	return WithHealthCheck(
		health.WithHealthCheckTest("CMD-SHELL", cmd),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(5),
		health.WithHealthCheckRetries(3),
	)
}

// WithLogFileContainsHealthCheck waits for a pattern in a log file.
// it has a default start period of 5 seconds, a timeout of 5 seconds, a retries of 3,
// and an interval of the parameter interval
// parameters:
//   - logFile: the log file to check
//   - pattern: the pattern to check for
//   - interval: the interval in seconds between health checks
//
// "CMD-SHELL", "grep -q "+pattern+" "+logFile+" || exit 1"
func WithLogFileContainsHealthCheck(logFile, pattern string, interval int) create.SetContainerConfig {
	cmd := fmt.Sprintf(`grep -q "%s" "%s" || exit 1`, pattern, logFile)
	return WithHealthCheck(
		health.WithHealthCheckTest("CMD-SHELL", cmd),
		health.WithHealthCheckInterval(interval),
		health.WithHealthCheckTimeout(5),
		health.WithHealthCheckRetries(3),
	)
}

// WithDisabledHealthCheck disables the health check by setting it to NONE.
func WithDisabledHealthCheck() create.SetContainerConfig {
	return WithHealthCheck(
		health.WithHealthCheckTest("NONE"),
	)
}
