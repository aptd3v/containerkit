package kit

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	composeapi "github.com/aptd3v/containerkit/internal/compose/api/types"
	"github.com/aptd3v/containerkit/pkg/cascade"
	"github.com/aptd3v/containerkit/pkg/oci"
	"github.com/aptd3v/containerkit/pkg/progress"
)

func TestComposeUpSetters(t *testing.T) {
	t.Parallel()
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	if err := p.Up(context.Background(), api, Detach(true), Quiet(true), Wait(true), Timeout(5*time.Second), Services("api")); err != nil {
		t.Fatal(err)
	}
	if !api.upOpts.Create.QuietPull {
		t.Fatalf("quiet pull = %+v", api.upOpts.Create)
	}
	if !api.upOpts.Start.Wait {
		t.Fatalf("wait = %+v", api.upOpts.Start)
	}
	if api.upOpts.Start.WaitTimeout != 5*time.Second {
		t.Fatalf("wait timeout = %v", api.upOpts.Start.WaitTimeout)
	}
	if api.upOpts.Start.Attach != nil {
		t.Fatalf("detach should leave attach nil: %+v", api.upOpts.Start.Attach)
	}
	if len(api.upOpts.Create.Services) != 1 || api.upOpts.Create.Services[0] != "api" {
		t.Fatalf("services = %v", api.upOpts.Create.Services)
	}
}

func TestUpStampsComposeIdentity(t *testing.T) {
	t.Parallel()
	p := NewCompose("supabase", AddService(Name("imgproxy"), Image("darthsim/imgproxy:v3.30.1"), ContainerName("supabase-imgproxy")))
	api := &fakeComposer{}
	if err := p.Up(context.Background(), api, Detach(true)); err != nil {
		t.Fatal(err)
	}
	svc, ok := api.up.Services["imgproxy"]
	if !ok {
		t.Fatal("imgproxy missing from project passed to Up")
	}
	if svc.CustomLabels[composeProjectLabel] != "supabase" || svc.CustomLabels[composeServiceLabel] != "imgproxy" || svc.CustomLabels[composeOneoffLabel] != "False" {
		t.Fatalf("compose identity labels = %#v", svc.CustomLabels)
	}
}

func TestComposeDownBuildPullSetters(t *testing.T) {
	t.Parallel()
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	if err := p.Down(context.Background(), api, Volumes(true), RemoveOrphans(true), Timeout(3*time.Second), RemoveImages("local")); err != nil {
		t.Fatal(err)
	}
	if !api.downOpts.Volumes || !api.downOpts.RemoveOrphans || api.downOpts.Images != "local" {
		t.Fatalf("down = %+v", api.downOpts)
	}
	if api.downOpts.Timeout == nil || *api.downOpts.Timeout != 3*time.Second {
		t.Fatalf("down timeout = %v", api.downOpts.Timeout)
	}
	if err := p.Build(context.Background(), api, Quiet(true), Pull(true), NoCache(true), Services("api")); err != nil {
		t.Fatal(err)
	}
	if !api.built.Quiet || !api.built.Pull || !api.built.NoCache {
		t.Fatalf("build = %+v", api.built)
	}
	if err := p.Pull(context.Background(), api, Quiet(true), IgnoreFailures(true), IgnoreBuildable(true)); err != nil {
		t.Fatal(err)
	}
	if !api.pulled.Quiet || !api.pulled.IgnoreFailures || !api.pulled.IgnoreBuildable {
		t.Fatalf("pull = %+v", api.pulled)
	}
}

func TestComposeRemainingActions(t *testing.T) {
	t.Parallel()
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	ctx := context.Background()
	if err := p.Create(ctx, api, Recreate("force"), Inherit(true)); err != nil {
		t.Fatal(err)
	}
	if api.created.Recreate != "force" || !api.created.Inherit {
		t.Fatalf("create = %+v", api.created)
	}
	for _, fn := range []func() error{
		func() error { return p.Start(ctx, api, Wait(true), Services("api")) },
		func() error { return p.Stop(ctx, api, Timeout(time.Second), Services("api")) },
		func() error { return p.Restart(ctx, api, NoDeps(true)) },
		func() error { return p.Push(ctx, api, Quiet(true), ImageMandatory(true)) },
		func() error { return p.Logs(ctx, api, fakeLogConsumer{}, Follow(true), Tail("all"), Services("api")) },
		func() error { return p.Kill(ctx, api, StopSignal("SIGTERM"), All(true)) },
		func() error { return p.Remove(ctx, api, Force(true), Stop(true), Volumes(true)) },
		func() error { return p.Pause(ctx, api, Services("api")) },
		func() error { return p.Unpause(ctx, api, Services("api")) },
		func() error { return p.Copy(ctx, api, CopyPath("/src", "/dst"), FollowLink(true)) },
		func() error { return p.Attach(ctx, api, "api", Index(1), DetachKeys("ctrl-p,ctrl-q")) },
		func() error {
			return p.Events(ctx, api, Since("1h"), EventConsumer(func(composeapi.Event) error { return nil }))
		},
		func() error { return p.Watch(ctx, api, Prune(true), Services("api")) },
		func() error { return p.Scale(ctx, api, Services("api")) },
		func() error {
			_, err := p.Wait(ctx, api, DownOnExit(true), Services("api"))
			return err
		},
		func() error { return p.Export(ctx, api, "api", OutputFile("out.tar"), Index(1)) },
		func() error {
			return p.Commit(ctx, api, "api", Pause(true), Author("a"), Comment("c"), Reference("img:tag"))
		},
		func() error { return p.PublishOCI(ctx, api, "repo", AssumeYes(true), ResolveDigests(true)) },
		func() error {
			_, err := p.Viz(ctx, api, IncludeNetworks(true), IncludePorts(true), IncludeImage(true), Indent("  "))
			return err
		},
		func() error {
			_, err := p.Images(ctx, api, Services("api"))
			return err
		},
		func() error {
			_, err := p.Ps(ctx, api, All(true), Services("api"))
			return err
		},
		func() error {
			_, err := p.VolumeList(ctx, api, Services("api"))
			return err
		},
		func() error {
			_, _, err := p.Port(ctx, api, "api", 80, Protocol(TCP), Index(1))
			return err
		},
		func() error {
			_, err := p.Top(ctx, api, "api")
			return err
		},
		func() error {
			_, err := p.Run(ctx, api, "api", Command("echo", "hi"), Detach(true), Name("once"))
			return err
		},
		func() error {
			_, err := p.Exec(ctx, api, "api", Command("true"), User("root"))
			return err
		},
		func() error { return p.MaxConcurrency(api, 4) },
	} {
		if err := fn(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.DryRunMode(ctx, api, true); err != nil {
		t.Fatal(err)
	}
	if _, err := List(ctx, api, All(true)); err != nil {
		t.Fatal(err)
	}
	proj, err := Generate(ctx, api, Name("from"), Containers("abc"))
	if err != nil {
		t.Fatal(err)
	}
	if proj == nil || proj.Name != "from" || len(api.gen.Containers) != 1 {
		t.Fatalf("generate = %+v opts=%+v", proj, api.gen)
	}
}

func TestComposeMethodMissingSetters(t *testing.T) {
	t.Parallel()
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	ctx := context.Background()

	if err := p.Build(ctx, api,
		SSH("default", "/id"),
		Env("FOO", "bar"),
		Attestations(true),
		Provenance("mode=min"),
		SBOM("mode=min"),
		Progress(progress.Plain),
		Push(true),
		Deps(true),
		Builder("default"),
		Print(true),
		Check(true),
		OutputStream(io.Discard),
		Memory(64),
	); err != nil {
		t.Fatal(err)
	}
	if len(api.built.SSHs) != 1 || api.built.SSHs[0].ID != "default" || api.built.SSHs[0].Path != "/id" {
		t.Fatalf("build ssh = %+v", api.built.SSHs)
	}
	if api.built.Args == nil || api.built.Args["FOO"] == nil || *api.built.Args["FOO"] != "bar" {
		t.Fatalf("build args = %+v", api.built.Args)
	}
	if !api.built.Attestations || api.built.Provenance != "mode=min" || api.built.SBOM != "mode=min" {
		t.Fatalf("build attest = %+v", api.built)
	}
	if api.built.Progress != string(progress.Plain) || !api.built.Push || !api.built.Deps || api.built.Builder != "default" || !api.built.Print || !api.built.Check || api.built.Memory != 64 || api.built.Out != io.Discard {
		t.Fatalf("build flags = %+v", api.built)
	}

	if err := p.Create(ctx, api, Pull(true), NoCache(true), SSH("agent", "")); err != nil {
		t.Fatal(err)
	}
	if api.created.Build == nil || !api.created.Build.Pull || !api.created.Build.NoCache || len(api.created.Build.SSHs) != 1 {
		t.Fatalf("create build = %+v", api.created.Build)
	}

	if err := p.Up(ctx, api, OnExit(cascade.Fail), NavigationMenu(true), Attestations(true), Env("UP", "1")); err != nil {
		t.Fatal(err)
	}
	if api.upOpts.Start.OnExit != composeapi.Cascade(cascade.Fail) || !api.upOpts.Start.NavigationMenu {
		t.Fatalf("up start = %+v", api.upOpts.Start)
	}
	if api.upOpts.Create.Build == nil || !api.upOpts.Create.Build.Attestations || api.upOpts.Create.Build.Args["UP"] == nil {
		t.Fatalf("up build = %+v", api.upOpts.Create.Build)
	}

	if err := p.Start(ctx, api, OnExit(cascade.Stop), NavigationMenu(true)); err != nil {
		t.Fatal(err)
	}
	if api.started.OnExit != composeapi.Cascade(cascade.Stop) || !api.started.NavigationMenu {
		t.Fatalf("start = %+v", api.started)
	}

	if err := p.Watch(ctx, api, Pull(true), Quiet(true), SSH("default", ""), Attestations(true)); err != nil {
		t.Fatal(err)
	}
	if api.watched.Build == nil || !api.watched.Build.Pull || !api.watched.Build.Quiet || !api.watched.Build.Attestations || len(api.watched.Build.SSHs) != 1 {
		t.Fatalf("watch build = %+v", api.watched.Build)
	}

	if err := p.PublishOCI(ctx, api, "repo", OCI(oci.V1_1)); err != nil {
		t.Fatal(err)
	}
	if api.published.OCIVersion != composeapi.OCIVersion(oci.V1_1) {
		t.Fatalf("publish oci = %q", api.published.OCIVersion)
	}

	if err := p.Commit(ctx, api, "api", Change("CMD /bin/true"), Change("EXPOSE 80")); err != nil {
		t.Fatal(err)
	}
	if got := api.committed.Changes.GetSlice(); len(got) != 2 || got[0] != "CMD /bin/true" || got[1] != "EXPOSE 80" {
		t.Fatalf("commit changes = %v", got)
	}

	if _, err := p.Run(ctx, api, "api", Recreate("force"), RecreateDependencies("never"), RemoveOrphans(true), IgnoreOrphans(true), AssumeYes(true), Timeout(2*time.Second), Pull(true)); err != nil {
		t.Fatal(err)
	}
	if api.ran.Recreate != "force" || api.ran.RecreateDependencies != "never" || !api.ran.RemoveOrphans || !api.ran.IgnoreOrphans {
		t.Fatalf("run create = %+v", api.ran.CreateOptions)
	}
	if api.ran.Timeout == nil || *api.ran.Timeout != 2*time.Second {
		t.Fatalf("run timeout = %v", api.ran.Timeout)
	}
	if api.ran.Build == nil || !api.ran.Build.Pull {
		t.Fatalf("run build = %+v", api.ran.Build)
	}

	if err := p.Commit(ctx, api, "api", Change("")); err == nil {
		t.Fatal("expected empty Change error")
	}
	if err := p.Build(ctx, api, SSH("", "")); err == nil {
		t.Fatal("expected empty SSH id error")
	}
}

func TestEngineCreateNameOverrideAndLifecycle(t *testing.T) {
	t.Parallel()
	c := NewContainer(Name("worker"), Image("alpine:latest"))
	api := &fakeEngine{}
	ctx := context.Background()
	created, err := c.Create(ctx, api, Name("override"))
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "abc" || api.created.Name != "override" {
		t.Fatalf("created = %+v cfg=%+v", created, api.created)
	}
	if err := c.Stop(ctx, api, created.ID, Timeout(2*time.Second), StopSignal("SIGTERM")); err != nil {
		t.Fatal(err)
	}
	if api.stopped.Signal != "SIGTERM" || api.stopped.Timeout == nil || *api.stopped.Timeout != 2 {
		t.Fatalf("stop = %+v", api.stopped)
	}
	if err := c.Remove(ctx, api, created.ID, Force(true), Volumes(true), Links(true)); err != nil {
		t.Fatal(err)
	}
	if !api.removed.Force || !api.removed.RemoveVolumes || !api.removed.RemoveLinks {
		t.Fatalf("remove = %+v", api.removed)
	}
	if err := c.Restart(ctx, api, created.ID, Timeout(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := c.Kill(ctx, api, created.ID, StopSignal("SIGKILL")); err != nil {
		t.Fatal(err)
	}
	if err := c.Pause(ctx, api, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := c.Unpause(ctx, api, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Inspect(ctx, api, created.ID, IncludeSize(true)); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Logs(ctx, api, created.ID, Follow(true), Tail("10"), Timestamps(true)); err != nil {
		t.Fatal(err)
	}
	_ = c.Wait(ctx, api, created.ID, WaitFor(NotRunning))
}

func TestEngineImageNetworkVolumeSetters(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	ctx := context.Background()
	if _, err := PullImage(ctx, api, "alpine:latest", All(true), Platform("linux/amd64"), RegistryAuth("xyz")); err != nil {
		t.Fatal(err)
	}
	if api.pullRef != "alpine:latest" || !api.pulled.All || api.pulled.RegistryAuth != "xyz" || len(api.pulled.Platforms) != 1 {
		t.Fatalf("pull = %s %+v", api.pullRef, api.pulled)
	}
	if _, err := BuildImage(ctx, api, Tar(bytes.NewReader(nil)), Image("app:latest"), Quiet(true), NoCache(true), Pull(true), Force(true), Memory(64), DockerfilePath("Dockerfile"), Target("release"), Label("a", "b"), Env("V", "1"), Platform("linux/arm64")); err != nil {
		t.Fatal(err)
	}
	if len(api.built.Tags) != 1 || api.built.Tags[0] != "app:latest" {
		t.Fatalf("build tags = %+v", api.built.Tags)
	}
	if !api.built.SuppressOutput || !api.built.NoCache || !api.built.PullParent || !api.built.ForceRemove || api.built.Memory != 64 || api.built.Dockerfile != "Dockerfile" || api.built.Target != "release" {
		t.Fatalf("build = %+v", api.built)
	}
	if _, err := CreateNetwork(ctx, api, "front", Driver("bridge"), Internal(true), Attachable(true), Label("k", "v"), DriverOpt("opt", "1")); err != nil {
		t.Fatal(err)
	}
	if api.netName != "front" || api.network.Driver != "bridge" || !api.network.Internal || !api.network.Attachable || api.network.Labels["k"] != "v" || api.network.Options["opt"] != "1" {
		t.Fatalf("network = %s %+v", api.netName, api.network)
	}
	if _, err := CreateVolume(ctx, api, Name("data"), Driver("local"), Label("k", "v"), DriverOpt("opt", "1")); err != nil {
		t.Fatal(err)
	}
	if api.volume.Name != "data" || api.volume.Driver != "local" || api.volume.Labels["k"] != "v" || api.volume.DriverOpts["opt"] != "1" {
		t.Fatalf("volume = %+v", api.volume)
	}
}

func TestRuntimeNilClientAndNilOption(t *testing.T) {
	t.Parallel()
	p := NewCompose("app", AddService(Name("api"), Image("alpine:latest")))
	c := NewContainer(Name("worker"), Image("alpine:latest"))
	ctx := context.Background()
	if err := p.Up(ctx, nil, Detach(true)); err == nil {
		t.Fatal("expected missing compose client")
	}
	if _, err := c.Create(ctx, nil, Name("x")); err == nil {
		t.Fatal("expected missing engine client")
	}
	api := &fakeComposer{}
	if err := p.Up(ctx, api, nil); err == nil {
		t.Fatal("expected nil option")
	}
}
