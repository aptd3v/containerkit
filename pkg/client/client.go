package client

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/docker/docker/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// Client is a wrapper around the docker client.
type Client struct {
	wrapped *client.Client
}

// SetClientOption is a configuration option to initialize a [Client].
type SetClientOption func(*client.Client) error

func NewClient(setters ...SetClientOption) (*Client, error) {
	opt := []client.Opt{}

	for _, setter := range setters {
		if setter == nil {
			continue
		}
		// Convert the SetClientOption to a client.Opt.
		opt = append(opt, func(c *client.Client) error {
			return setter(c)
		})
	}
	wrapped, err := client.NewClientWithOpts(opt...)
	if err != nil {
		return nil, err
	}

	return &Client{wrapped: wrapped}, nil
}

// Unwrap returns the underlying client.Client
func (c *Client) Unwrap() *client.Client {
	return c.wrapped
}

// WithHTTPClient overrides the client's HTTP client with the specified one.
func WithHTTPClient(httpClient *http.Client) SetClientOption {
	return func(c *client.Client) error {
		return client.WithHTTPClient(httpClient)(c)
	}
}

// FromEnv configures the client with values from environment variables. It
// is the equivalent of using the [WithTLSClientConfigFromEnv], [WithHostFromEnv],
// and [WithVersionFromEnv] options.
//
// FromEnv uses the following environment variables:
//
//   - DOCKER_HOST ([EnvOverrideHost]) to set the URL to the docker server.
//   - DOCKER_API_VERSION ([EnvOverrideAPIVersion]) to set the version of the
//     API to use, leave empty for latest.
//   - DOCKER_CERT_PATH ([EnvOverrideCertPath]) to specify the directory from
//     which to load the TLS certificates ("ca.pem", "cert.pem", "key.pem').
//   - DOCKER_TLS_VERIFY ([EnvTLSVerify]) to enable or disable TLS verification
//     (off by default).
func FromEnv() SetClientOption {
	return func(c *client.Client) error {
		return client.FromEnv(c)
	}
}

// WithDialContext applies the dialer to the client transport. This can be
// used to set the Timeout and KeepAlive settings of the client. It returns
// an error if the client does not have a [http.Transport] configured.
func WithDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) SetClientOption {
	return func(c *client.Client) error {
		return client.WithDialContext(dialContext)(c)
	}
}

// WithHost overrides the client host with the specified one.
func WithHost(host string) SetClientOption {
	return func(c *client.Client) error {
		return client.WithHost(host)(c)
	}
}

// WithHostFromEnv overrides the client host with the host specified in the
// DOCKER_HOST ([EnvOverrideHost]) environment variable. If DOCKER_HOST is not set,
// or set to an empty value, the host is not modified.
func WithHostFromEnv() SetClientOption {
	return func(c *client.Client) error {
		return client.WithHostFromEnv()(c)
	}
}

// WithTimeout configures the time limit for requests made by the HTTP client.
func WithTimeout(timeout time.Duration) SetClientOption {
	return func(c *client.Client) error {
		return client.WithTimeout(timeout)(c)
	}
}

// WithUserAgent configures the User-Agent header to use for HTTP requests.
// It overrides any User-Agent set in headers. When set to an empty string,
// the User-Agent header is removed, and no header is sent.
func WithUserAgent(userAgent string) SetClientOption {
	return func(c *client.Client) error {
		return client.WithUserAgent(userAgent)(c)
	}
}

// WithHTTPHeaders appends custom HTTP headers to the client's default headers.
// It does not allow for built-in headers (such as "User-Agent", if set) to
// be overridden. Also see [WithUserAgent].
func WithHTTPHeaders(headers map[string]string) SetClientOption {
	return func(c *client.Client) error {
		return client.WithHTTPHeaders(headers)(c)
	}
}

// WithScheme overrides the client scheme with the specified one.
func WithScheme(scheme string) SetClientOption {
	return func(c *client.Client) error {
		return client.WithScheme(scheme)(c)
	}
}

// WithTLSClientConfig applies a TLS config to the client transport.
func WithTLSClientConfig(cacertPath, certPath, keyPath string) SetClientOption {
	return func(c *client.Client) error {
		return client.WithTLSClientConfig(cacertPath, certPath, keyPath)(c)

	}
}

// WithTLSClientConfigFromEnv configures the client's TLS settings with the
// settings in the DOCKER_CERT_PATH ([EnvOverrideCertPath]) and DOCKER_TLS_VERIFY
// ([EnvTLSVerify]) environment variables. If DOCKER_CERT_PATH is not set or empty,
// TLS configuration is not modified.
//
// WithTLSClientConfigFromEnv uses the following environment variables:
//
//   - DOCKER_CERT_PATH ([EnvOverrideCertPath]) to specify the directory from
//     which to load the TLS certificates ("ca.pem", "cert.pem", "key.pem").
//   - DOCKER_TLS_VERIFY ([EnvTLSVerify]) to enable or disable TLS verification
//     (off by default).
func WithTLSClientConfigFromEnv() SetClientOption {
	return func(c *client.Client) error {
		return client.WithTLSClientConfigFromEnv()(c)
	}
}

// WithVersion overrides the client version with the specified one. If an empty
// version is provided, the value is ignored to allow version negotiation
// (see [WithAPIVersionNegotiation]).
func WithVersion(version string) SetClientOption {
	return func(c *client.Client) error {
		return client.WithVersion(version)(c)
	}
}

// WithVersionFromEnv overrides the client version with the version specified in
// the DOCKER_API_VERSION ([EnvOverrideAPIVersion]) environment variable.
// If DOCKER_API_VERSION is not set, or set to an empty value, the version
// is not modified.
func WithVersionFromEnv() SetClientOption {
	return func(c *client.Client) error {
		return client.WithVersionFromEnv()(c)
	}
}

// WithAPIVersionNegotiation enables automatic API version negotiation for the client.
// With this option enabled, the client automatically negotiates the API version
// to use when making requests. API version negotiation is performed on the first
// request; subsequent requests do not re-negotiate.
func WithAPIVersionNegotiation() SetClientOption {
	return func(c *client.Client) error {
		return client.WithAPIVersionNegotiation()(c)
	}
}

// WithTraceProvider sets the trace provider for the client.
// If this is not set then the global trace provider will be used.
func WithTraceProvider(provider trace.TracerProvider) SetClientOption {
	return func(c *client.Client) error {
		return client.WithTraceProvider(provider)(c)
	}
}

// WithTraceOptions sets tracing span options for the client.
func WithTraceOptions(opts ...otelhttp.Option) SetClientOption {
	return func(c *client.Client) error {
		return client.WithTraceOptions(opts...)(c)
	}
}
