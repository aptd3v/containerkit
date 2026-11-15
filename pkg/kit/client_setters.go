package kit

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/aptd3v/containerkit/pkg/scheme"
	engineclient "github.com/moby/moby/client"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// FromEnvOpt is the value returned by [FromEnv].
// It implements [ClientOption] and [ComposerClientOption].
type FromEnvOpt struct{}

// FromEnv configures the Engine client from DOCKER_HOST, DOCKER_API_VERSION,
// DOCKER_CERT_PATH, and DOCKER_TLS_VERIFY.
func FromEnv() FromEnvOpt { return FromEnvOpt{} }

func (FromEnvOpt) applyClient(a *clientAct) error {
	a.add(engineclient.FromEnv)
	return nil
}

// HostOpt is the value returned by [Host].
// It implements [ClientOption] and [ComposerClientOption].
type HostOpt struct{ host string }

// Host sets the Engine daemon URL (for example unix:///var/run/docker.sock).
func Host(host string) HostOpt { return HostOpt{host: host} }

func (o HostOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithHost(o.host))
	return nil
}

// HostFromEnvOpt is the value returned by [HostFromEnv].
// It implements [ClientOption] and [ComposerClientOption].
type HostFromEnvOpt struct{}

// HostFromEnv sets the daemon URL from DOCKER_HOST.
func HostFromEnv() HostFromEnvOpt { return HostFromEnvOpt{} }

func (HostFromEnvOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithHostFromEnv())
	return nil
}

// HTTPClientOpt is the value returned by [HTTPClient].
// It implements [ClientOption] and [ComposerClientOption].
type HTTPClientOpt struct{ c *http.Client }

// HTTPClient sets the HTTP client used to reach the daemon.
func HTTPClient(c *http.Client) HTTPClientOpt { return HTTPClientOpt{c: c} }

func (o HTTPClientOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithHTTPClient(o.c))
	return nil
}

func (o TimeoutOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithTimeout(o.d))
	return nil
}

// UserAgentOpt is the value returned by [UserAgent].
// It implements [ClientOption] and [ComposerClientOption].
type UserAgentOpt struct{ ua string }

// UserAgent sets the HTTP User-Agent header.
func UserAgent(ua string) UserAgentOpt { return UserAgentOpt{ua: ua} }

func (o UserAgentOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithUserAgent(o.ua))
	return nil
}

// HTTPHeaderOpt is the value returned by [HTTPHeader].
// It implements [ClientOption] and [ComposerClientOption].
type HTTPHeaderOpt struct{ key, value string }

// HTTPHeader adds a custom HTTP header on Engine requests.
func HTTPHeader(key, value string) HTTPHeaderOpt { return HTTPHeaderOpt{key: key, value: value} }

func (o HTTPHeaderOpt) applyClient(a *clientAct) error {
	if a.headers == nil {
		a.headers = map[string]string{}
	}
	a.headers[o.key] = o.value
	return nil
}

// SchemeOpt is the value returned by [Scheme].
// It implements [ClientOption] and [ComposerClientOption].
type SchemeOpt struct{ s string }

// Scheme sets the Engine client URL scheme. Known values live in package scheme.
func Scheme(s scheme.Mode) SchemeOpt { return SchemeOpt{s: string(s)} }

func (o SchemeOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithScheme(o.s))
	return nil
}

// TLSOpt is the value returned by [TLS].
// It implements [ClientOption] and [ComposerClientOption].
type TLSOpt struct{ opts []TLSOption }

// TLS configures client TLS. Nested setters are [CA], [Cert], and [Key].
func TLS(opts ...TLSOption) TLSOpt { return TLSOpt{opts: opts} }

func (o TLSOpt) applyClient(a *clientAct) error {
	var errs []error
	for i, opt := range o.opts {
		if opt == nil {
			errs = addErr(errs, cfgErr("engine", fmt.Sprintf("tls.option[%d]", i), "option", errNilOption))
			continue
		}
		errs = addErr(errs, opt.applyTLS(&a.tls))
	}
	a.tlsSet = true
	return join(errs...)
}

// CAOpt is the value returned by [CA].
// It implements [TLSOption].
type CAOpt struct{ path string }

// CA sets the TLS CA certificate path inside [TLS].
func CA(path string) CAOpt { return CAOpt{path: path} }

func (o CAOpt) applyTLS(t *tlsAct) error { t.ca = o.path; return nil }

// CertOpt is the value returned by [Cert].
// It implements [TLSOption].
type CertOpt struct{ path string }

// Cert sets the TLS client certificate path inside [TLS].
func Cert(path string) CertOpt { return CertOpt{path: path} }

func (o CertOpt) applyTLS(t *tlsAct) error { t.cert = o.path; return nil }

// KeyOpt is the value returned by [Key].
// It implements [TLSOption].
type KeyOpt struct{ path string }

// Key sets the TLS client key path inside [TLS].
func Key(path string) KeyOpt { return KeyOpt{path: path} }

func (o KeyOpt) applyTLS(t *tlsAct) error { t.key = o.path; return nil }

// TLSFromEnvOpt is the value returned by [TLSFromEnv].
// It implements [ClientOption] and [ComposerClientOption].
type TLSFromEnvOpt struct{}

// TLSFromEnv loads TLS files from DOCKER_CERT_PATH and DOCKER_TLS_VERIFY.
func TLSFromEnv() TLSFromEnvOpt { return TLSFromEnvOpt{} }

func (TLSFromEnvOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithTLSClientConfigFromEnv())
	return nil
}

// APIVersionOpt is the value returned by [APIVersion].
// It implements [ClientOption] and [ComposerClientOption].
type APIVersionOpt struct{ v string }

// APIVersion pins the Engine API version and disables negotiation.
func APIVersion(v string) APIVersionOpt { return APIVersionOpt{v: v} }

func (o APIVersionOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithAPIVersion(o.v))
	return nil
}

// APIVersionFromEnvOpt is the value returned by [APIVersionFromEnv].
// It implements [ClientOption] and [ComposerClientOption].
type APIVersionFromEnvOpt struct{}

// APIVersionFromEnv pins the Engine API version from DOCKER_API_VERSION.
func APIVersionFromEnv() APIVersionFromEnvOpt { return APIVersionFromEnvOpt{} }

func (APIVersionFromEnvOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithAPIVersionFromEnv())
	return nil
}

// DialContextOpt is the value returned by [DialContext].
// It implements [ClientOption] and [ComposerClientOption].
type DialContextOpt struct {
	fn func(ctx context.Context, network, addr string) (net.Conn, error)
}

// DialContext sets the dialer used by the Engine HTTP transport.
func DialContext(fn func(ctx context.Context, network, addr string) (net.Conn, error)) DialContextOpt {
	return DialContextOpt{fn: fn}
}

func (o DialContextOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithDialContext(o.fn))
	return nil
}

// HTTPRequestHookOpt is the value returned by [HTTPRequestHook].
// It implements [ClientOption] and [ComposerClientOption].
type HTTPRequestHookOpt struct {
	fn func(*http.Request) error
}

// HTTPRequestHook runs before each Engine HTTP request.
func HTTPRequestHook(fn func(*http.Request) error) HTTPRequestHookOpt {
	return HTTPRequestHookOpt{fn: fn}
}

func (o HTTPRequestHookOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithHTTPRequestHook(o.fn))
	return nil
}

// NegotiateAPIVersionOpt is the value returned by [NegotiateAPIVersion].
// It implements [ClientOption], [ComposerClientOption], and [EnginePingOption].
type NegotiateAPIVersionOpt struct{}

// NegotiateAPIVersion enables Engine API version negotiation on [NewEngine]
// or [Ping]. The Engine SDK treats the client option as a no-op because
// negotiation is now the default; on [Ping] it sets PingOptions.NegotiateAPIVersion.
func NegotiateAPIVersion() NegotiateAPIVersionOpt { return NegotiateAPIVersionOpt{} }

func (NegotiateAPIVersionOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithAPIVersionNegotiation())
	return nil
}

func (NegotiateAPIVersionOpt) applyEnginePingOption(a *enginePingAct) error {
	a.opts.NegotiateAPIVersion = true
	return nil
}

// TraceProviderOpt is the value returned by [TraceProvider].
// It implements [ClientOption] and [ComposerClientOption].
type TraceProviderOpt struct{ p trace.TracerProvider }

// TraceProvider sets the OpenTelemetry tracer provider for Engine HTTP spans.
func TraceProvider(p trace.TracerProvider) TraceProviderOpt { return TraceProviderOpt{p: p} }

func (o TraceProviderOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithTraceProvider(o.p))
	return nil
}

// TraceOptionsOpt is the value returned by [TraceOptions].
// It implements [ClientOption] and [ComposerClientOption].
type TraceOptionsOpt struct{ opts []otelhttp.Option }

// TraceOptions adds OpenTelemetry HTTP span options to the Engine client.
func TraceOptions(opts ...otelhttp.Option) TraceOptionsOpt { return TraceOptionsOpt{opts: opts} }

func (o TraceOptionsOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithTraceOptions(o.opts...))
	return nil
}

// HTTPResponseHookOpt is the value returned by [HTTPResponseHook].
// It implements [ClientOption] and [ComposerClientOption].
type HTTPResponseHookOpt struct {
	fn func(*http.Response)
}

// HTTPResponseHook runs after each Engine HTTP response.
func HTTPResponseHook(fn func(*http.Response)) HTTPResponseHookOpt {
	return HTTPResponseHookOpt{fn: fn}
}

func (o HTTPResponseHookOpt) applyClient(a *clientAct) error {
	a.add(engineclient.WithHTTPResponseHook(o.fn))
	return nil
}

func (FromEnvOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, FromEnvOpt{})
}
func (o HostOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (HostFromEnvOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, HostFromEnvOpt{})
}
func (o HTTPClientOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o TimeoutOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o UserAgentOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o HTTPHeaderOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o SchemeOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o TLSOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (TLSFromEnvOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, TLSFromEnvOpt{})
}
func (o APIVersionOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (APIVersionFromEnvOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, APIVersionFromEnvOpt{})
}
func (o DialContextOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o HTTPRequestHookOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (NegotiateAPIVersionOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, NegotiateAPIVersionOpt{})
}
func (o TraceProviderOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o TraceOptionsOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}
func (o HTTPResponseHookOpt) applyComposerClient(a *composerAct) error {
	return applyAsComposerClient(a, o)
}

var (
	_ ClientOption         = Timeout(0)
	_ ClientOption         = FromEnv()
	_ ClientOption         = Host("")
	_ ClientOption         = NegotiateAPIVersion()
	_ ClientOption         = TraceProvider(nil)
	_ ClientOption         = TraceOptions()
	_ ComposerClientOption = Timeout(0)
	_ ComposerClientOption = FromEnv()
	_ ComposerClientOption = Host("")
	_ ComposerClientOption = MaxConcurrency(0)
	_ ComposerClientOption = DryRun(true)
	_ TLSOption            = CA("")
	_ TLSOption            = Cert("")
	_ TLSOption            = Key("")
)
