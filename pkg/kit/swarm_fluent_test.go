package kit

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/avail"
	"github.com/aptd3v/containerkit/pkg/role"
	"github.com/aptd3v/containerkit/pkg/svcmode"
)

func TestCreateSwarmServiceFluentSetters(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	ctx := context.Background()
	out, err := CreateSwarmService(ctx, api,
		Name("web"),
		Image("alpine:3.21"),
		Command("sleep", "infinity"),
		Env("FOO", "bar"),
		Label("app", "containerkit"),
		Replicas(3),
		RegistryAuth("auth"),
		QueryRegistry(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "svc1" {
		t.Fatalf("id = %q", out.ID)
	}
	spec := api.serviceCreate.Spec
	if spec.Name != "web" {
		t.Fatalf("name = %q", spec.Name)
	}
	if spec.TaskTemplate.ContainerSpec == nil || spec.TaskTemplate.ContainerSpec.Image != "alpine:3.21" {
		t.Fatalf("image = %+v", spec.TaskTemplate.ContainerSpec)
	}
	if got := spec.TaskTemplate.ContainerSpec.Command; len(got) != 2 || got[0] != "sleep" {
		t.Fatalf("command = %#v", got)
	}
	if got := spec.TaskTemplate.ContainerSpec.Env; len(got) != 1 || got[0] != "FOO=bar" {
		t.Fatalf("env = %#v", got)
	}
	if spec.Labels["app"] != "containerkit" {
		t.Fatalf("labels = %#v", spec.Labels)
	}
	if spec.Mode.Replicated == nil || spec.Mode.Replicated.Replicas == nil || *spec.Mode.Replicated.Replicas != 3 {
		t.Fatalf("replicas = %+v", spec.Mode.Replicated)
	}
	if api.serviceCreate.EncodedRegistryAuth != "auth" || !api.serviceCreate.QueryRegistry {
		t.Fatalf("create opts = %+v", api.serviceCreate)
	}
}

func TestCreateSwarmServiceModeAndNil(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := &fakeEngine{}
	if _, err := CreateSwarmService(ctx, api, Mode(svcmode.Global)); err != nil {
		t.Fatal(err)
	}
	if api.serviceCreate.Spec.Mode.Global == nil {
		t.Fatalf("mode = %+v", api.serviceCreate.Spec.Mode)
	}
	if _, err := CreateSwarmService(ctx, api, nil); err == nil {
		t.Fatal("expected nil option")
	}
	if _, err := CreateSwarmService(ctx, nil, Name("web")); err == nil {
		t.Fatal("expected nil engine")
	}
	if _, err := CreateSwarmService(ctx, api, Mode("nope")); err == nil {
		t.Fatal("expected invalid mode")
	}
}

func TestUpdateSwarmServiceFluentSetters(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	if _, err := UpdateSwarmService(context.Background(), api, "svc1",
		Name("web"),
		Image("alpine:3.21"),
		VersionIndex(7),
		RegistryAuthFrom("spec"),
		ServiceRollback("previous"),
		QueryRegistry(true),
	); err != nil {
		t.Fatal(err)
	}
	if api.serviceUpdate.Version.Index != 7 || api.serviceUpdate.Rollback != "previous" {
		t.Fatalf("update = %+v", api.serviceUpdate)
	}
	if api.serviceUpdate.Spec.Name != "web" {
		t.Fatalf("name = %q", api.serviceUpdate.Spec.Name)
	}
}

func TestCreateEngineSecretAndConfigFluent(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	ctx := context.Background()
	if _, err := CreateEngineSecret(ctx, api, Name("tok"), Label("k", "v"), Data([]byte("secret"))); err != nil {
		t.Fatal(err)
	}
	if api.secretCreate.Spec.Name != "tok" || string(api.secretCreate.Spec.Data) != "secret" {
		t.Fatalf("secret = %+v", api.secretCreate.Spec)
	}
	if _, err := CreateEngineConfig(ctx, api, Name("cfg"), Content(strings.NewReader("body"))); err != nil {
		t.Fatal(err)
	}
	if api.configCreate.Spec.Name != "cfg" || string(api.configCreate.Spec.Data) != "body" {
		t.Fatalf("config = %+v", api.configCreate.Spec)
	}
	if _, err := CreateEngineSecret(ctx, api, Content(nil)); err == nil {
		t.Fatal("expected nil content")
	}
}

func TestUpdateNodeAndVolumeFluent(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	ctx := context.Background()
	if _, err := UpdateNode(ctx, api, "n1",
		VersionIndex(2),
		Name("worker"),
		Role(role.Worker),
		Availability(avail.Drain),
		Label("rack", "a"),
	); err != nil {
		t.Fatal(err)
	}
	if api.nodeUpdate.Version.Index != 2 || api.nodeUpdate.Spec.Role != NodeRoleWorker {
		t.Fatalf("node = %+v", api.nodeUpdate)
	}
	if api.nodeUpdate.Spec.Availability != NodeAvailability(avail.Drain) {
		t.Fatalf("avail = %q", api.nodeUpdate.Spec.Availability)
	}
	spec := &ClusterVolumeSpec{}
	if _, err := UpdateVolume(ctx, api, "vol1", VersionIndex(3), ClusterSpec(spec)); err != nil {
		t.Fatal(err)
	}
	if api.volumeUpdate.Version.Index != 3 || api.volumeUpdate.Spec != spec {
		t.Fatalf("volume = %+v", api.volumeUpdate)
	}
}

func TestDialerWrapsEngine(t *testing.T) {
	t.Parallel()
	d, err := Dialer(&fakeEngine{})
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("nil dialer")
	}
	if _, err := Dialer(nil); err == nil {
		t.Fatal("expected nil engine")
	}
	if _, err := DialHijack(context.Background(), nil, "/", "tcp", nil); err == nil {
		t.Fatal("expected nil engine")
	}
}

func TestInspectImageFluentOptions(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	var raw bytes.Buffer
	if _, err := InspectImage(context.Background(), api, "alpine:latest", Manifests(true), RawResponse(&raw)); err != nil {
		t.Fatal(err)
	}
	if _, err := InspectImage(context.Background(), api, "", Manifests(true)); err == nil {
		t.Fatal("expected empty ref")
	}
	if _, err := InspectImage(context.Background(), api, "alpine:latest", nil); err == nil {
		t.Fatal("expected nil option")
	}
	if _, err := LoadImage(context.Background(), api, strings.NewReader("tar"), Quiet(true)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadImage(context.Background(), api, nil); err == nil {
		t.Fatal("expected nil reader")
	}
}
