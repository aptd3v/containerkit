package kit_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestNewEngineClientSettersCompileWithoutEngineSDK(t *testing.T) {
	t.Parallel()
	var (
		_ kit.ClientOption              = kit.FromEnv()
		_ kit.ClientOption              = kit.Host("unix:///var/run/docker.sock")
		_ kit.ClientOption              = kit.HostFromEnv()
		_ kit.ClientOption              = kit.Timeout(2 * time.Second)
		_ kit.ClientOption              = kit.UserAgent("containerkit")
		_ kit.ClientOption              = kit.HTTPHeader("X-Test", "1")
		_ kit.ClientOption              = kit.Scheme("http")
		_ kit.ClientOption              = kit.TLS(kit.CA("ca.pem"), kit.Cert("cert.pem"), kit.Key("key.pem"))
		_ kit.ClientOption              = kit.TLSFromEnv()
		_ kit.ClientOption              = kit.APIVersion("1.52")
		_ kit.ClientOption              = kit.APIVersionFromEnv()
		_ kit.ClientOption              = kit.HTTPClient(http.DefaultClient)
		_ kit.ClientOption              = kit.DialContext(nil)
		_ kit.ClientOption              = kit.HTTPRequestHook(func(*http.Request) error { return nil })
		_ kit.ClientOption              = kit.HTTPResponseHook(func(*http.Response) {})
		_ kit.ClientOption              = kit.NegotiateAPIVersion()
		_ kit.ClientOption              = kit.TraceProvider(nil)
		_ kit.ClientOption              = kit.TraceOptions()
		_ kit.EngineContainerListOption = kit.All(true)
		_ kit.EngineContainerListOption = kit.Filter("status", "running")
		_ kit.EngineContainerListOption = kit.Limit(5)
		_ kit.EngineImageInspectOption  = kit.Manifests(true)
		_ kit.EngineServiceCreateOption = kit.Name("web")
		_ kit.EngineServiceCreateOption = kit.Image("alpine:latest")
		_ kit.EngineServiceCreateOption = kit.QueryRegistry(true)
		_ kit.EngineSecretCreateOption  = kit.Data([]byte("s"))
		_ kit.EngineWaitOption          = kit.WaitFor(kit.NotRunning)
		_ kit.EngineWaitOption          = kit.WaitFor(kit.NextExit)
		_ kit.EngineWaitOption          = kit.WaitFor(kit.Removed)
		_ kit.EngineSwarmInitOption     = kit.ListenAddr("0.0.0.0:2377")
		_ kit.EngineSwarmInitOption     = kit.AdvertiseAddr("10.0.0.11:2377")
		_ kit.EngineSwarmInitOption     = kit.DataPathAddr("10.0.0.11")
		_ kit.EngineSwarmInitOption     = kit.ForceNewCluster(true)
		_ kit.EngineSwarmJoinOption     = kit.JoinToken("token")
		_ kit.EngineSwarmJoinOption     = kit.RemoteAddrs("10.0.0.11:2377")
		_ kit.EngineSwarmJoinOption     = kit.ListenAddr("0.0.0.0:2377")
		_ kit.EngineSwarmJoinOption     = kit.AdvertiseAddr("10.0.0.12:2377")
	)
	c := kit.NewContainer(kit.Name("worker"), kit.Image("alpine:latest"))
	_ = func(ctx context.Context, api kit.Engine) (kit.Created, error) {
		return c.Create(ctx, api, kit.Name("worker"))
	}
	_ = func(ctx context.Context, api kit.Engine, id string) kit.ContainerWaitResult {
		return c.Wait(ctx, api, id, kit.WaitFor(kit.NotRunning))
	}
	_ = func(ctx context.Context, api kit.Engine) (kit.ImagePullResponse, error) {
		return kit.PullImage(ctx, api, "alpine:latest", kit.All(true), kit.Platform("linux/amd64"))
	}
	_ = func(ctx context.Context, api kit.Engine) (kit.ContainerListResult, error) {
		return kit.ListContainers(ctx, api, kit.All(true), kit.Filter("status", "running"))
	}
	_ = func(ctx context.Context, api kit.Engine) (kit.ImageInspectResult, error) {
		return kit.InspectImage(ctx, api, "alpine:latest", kit.Manifests(true))
	}
	_ = func(ctx context.Context, api kit.Engine) (kit.ServiceCreateResult, error) {
		return kit.CreateSwarmService(ctx, api, kit.Name("web"), kit.Image("alpine:latest"), kit.Command("sleep", "infinity"), kit.Replicas(1))
	}
}

func TestNewEngineConstructsClient(t *testing.T) {
	t.Parallel()
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{}"))}, nil
	})
	api, err := kit.NewEngine(
		kit.Host("http://localhost"),
		kit.HTTPClient(&http.Client{Transport: transport}),
		kit.Timeout(time.Second),
		kit.UserAgent("containerkit-test"),
		kit.HTTPHeader("X-Test", "1"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := api.Close(); err != nil {
			t.Error(err)
		}
	})
}

func TestNewEngineNilOption(t *testing.T) {
	t.Parallel()
	_, err := kit.NewEngine(kit.Host("http://localhost"), nil)
	if err == nil {
		t.Fatal("expected nil option")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
