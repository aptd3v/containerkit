package client

import (
	"context"
	"net"
	"net/http"
	"time"

	sdk "github.com/moby/moby/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// Client is a wrapper around the Docker Engine API client.
type Client struct {
	wrapped *sdk.Client
}

// SetClientOption is a configuration option to initialize a [Client].
type SetClientOption = sdk.Opt

// NewClient initializes a client. Nil options are skipped.
func NewClient(setters ...SetClientOption) (*Client, error) {
	ops := make([]sdk.Opt, 0, len(setters))
	for _, setter := range setters {
		if setter == nil {
			continue
		}
		ops = append(ops, setter)
	}
	wrapped, err := sdk.New(ops...)
	if err != nil {
		return nil, err
	}
	return &Client{wrapped: wrapped}, nil
}

// Unwrap returns the underlying Engine API client.
func (c *Client) Unwrap() *sdk.Client {
	return c.wrapped
}

// Close closes idle connections used by the client.
func (c *Client) Close() error {
	return c.wrapped.Close()
}

// ClientVersion returns the API version this client sends.
func (c *Client) ClientVersion() string {
	return c.wrapped.ClientVersion()
}

// DaemonHost returns the daemon address configured on the client.
func (c *Client) DaemonHost() string {
	return c.wrapped.DaemonHost()
}

// Dialer returns a dial function for a raw stream to the daemon.
func (c *Client) Dialer() func(context.Context) (net.Conn, error) {
	return c.wrapped.Dialer()
}

// DialHijack opens a hijacked connection. proto is the Upgrade protocol, such as "tcp".
// meta is copied onto the request headers.
func (c *Client) DialHijack(ctx context.Context, rawURL, proto string, meta map[string][]string) (net.Conn, error) {
	return c.wrapped.DialHijack(ctx, rawURL, proto, meta)
}

// WithHTTPClient overrides the client's HTTP client with the specified one.
func WithHTTPClient(httpClient *http.Client) SetClientOption {
	return sdk.WithHTTPClient(httpClient)
}

// FromEnv configures the client with values from environment variables. It
// is the equivalent of using the [WithTLSClientConfigFromEnv], [WithHostFromEnv],
// and [WithVersionFromEnv] options.
//
// FromEnv uses the following environment variables:
//
//   - DOCKER_HOST to set the URL to the docker server.
//   - DOCKER_API_VERSION to set the version of the API to use, leave empty for latest.
//   - DOCKER_CERT_PATH to specify the directory from which to load the TLS
//     certificates ("ca.pem", "cert.pem", "key.pem").
//   - DOCKER_TLS_VERIFY to enable or disable TLS verification (off by default).
func FromEnv() SetClientOption {
	return sdk.FromEnv
}

// WithDialContext applies the dialer to the client transport. This can be
// used to set the Timeout and KeepAlive settings of the client. It returns
// an error if the client does not have an [http.Transport] configured.
func WithDialContext(dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) SetClientOption {
	return sdk.WithDialContext(dialContext)
}

// WithHost overrides the client host with the specified one.
func WithHost(host string) SetClientOption {
	return sdk.WithHost(host)
}

// WithHostFromEnv overrides the client host with the host specified in the
// DOCKER_HOST environment variable. If DOCKER_HOST is not set, or set to an
// empty value, the host is not modified.
func WithHostFromEnv() SetClientOption {
	return sdk.WithHostFromEnv()
}

// WithTimeout configures the time limit for requests made by the HTTP client.
func WithTimeout(timeout time.Duration) SetClientOption {
	return sdk.WithTimeout(timeout)
}

// WithUserAgent configures the User-Agent header to use for HTTP requests.
// It overrides any User-Agent set in headers. When set to an empty string,
// the User-Agent header is removed, and no header is sent.
func WithUserAgent(userAgent string) SetClientOption {
	return sdk.WithUserAgent(userAgent)
}

// WithHTTPHeaders appends custom HTTP headers to the client's default headers.
// It does not allow for built-in headers (such as "User-Agent", if set) to
// be overridden. Also see [WithUserAgent].
func WithHTTPHeaders(headers map[string]string) SetClientOption {
	return sdk.WithHTTPHeaders(headers)
}

// WithScheme overrides the client scheme with the specified one.
func WithScheme(scheme string) SetClientOption {
	return sdk.WithScheme(scheme)
}

// WithTLSClientConfig applies a TLS config to the client transport.
func WithTLSClientConfig(cacertPath, certPath, keyPath string) SetClientOption {
	return sdk.WithTLSClientConfig(cacertPath, certPath, keyPath)
}

// WithTLSClientConfigFromEnv configures the client's TLS settings with the
// settings in the DOCKER_CERT_PATH and DOCKER_TLS_VERIFY environment variables.
// If DOCKER_CERT_PATH is not set or empty, TLS configuration is not modified.
func WithTLSClientConfigFromEnv() SetClientOption {
	return sdk.WithTLSClientConfigFromEnv()
}

// WithAPIVersion overrides the client version with the specified one. If an empty
// version is provided, the value is ignored to allow version negotiation
// (see [WithAPIVersionNegotiation]).
func WithAPIVersion(version string) SetClientOption {
	return sdk.WithAPIVersion(version)
}

// WithAPIVersionNegotiation leaves the API version unset so the client
// negotiates with the daemon.
func WithAPIVersionNegotiation() SetClientOption {
	return sdk.WithAPIVersion("")
}

// WithAPIVersionFromEnv overrides the client version with the version specified in
// the DOCKER_API_VERSION environment variable. If DOCKER_API_VERSION is not set,
// or set to an empty value, the version is not modified.
func WithAPIVersionFromEnv() SetClientOption {
	return sdk.WithAPIVersionFromEnv()
}

// WithTraceProvider sets the trace provider for the client.
// If this is not set then the global trace provider will be used.
func WithTraceProvider(provider trace.TracerProvider) SetClientOption {
	return sdk.WithTraceProvider(provider)
}

// WithTraceOptions sets tracing span options for the client.
func WithTraceOptions(opts ...otelhttp.Option) SetClientOption {
	return sdk.WithTraceOptions(opts...)
}
