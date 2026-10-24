//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

func TestComposeEchoUp(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := kit.NewProject(name)
	project.WithService(kit.NewService("echo", kit.NewContainer().
		Image(alpine).
		Command("echo", "hello-e2e"),
	))
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := kit.NewCompose(project)
	composeDown(t, app)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := app.Up(ctx, &kit.Up{Writer: testWriter(t), RemoveOrphans: true}); err != nil {
		t.Fatalf("Up: %v", err)
	}
}

func TestComposeLifecycle(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := kit.NewProject(name)
	project.WithService(kit.NewService("app", kit.NewContainer().
		Image(alpine).
		Command("sleep", "infinity").
		PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "127.0.0.1", HostPort: "0", ContainerPort: "80"}),
	))
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := kit.NewCompose(project)
	composeDown(t, app)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := app.Up(ctx, &kit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}

	if err := app.Ps(ctx, &kit.Ps{All: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Ps: %v", err)
	}

	var psOut bytes.Buffer
	if err := app.Ps(ctx, &kit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t))}); err != nil {
		t.Fatalf("Ps json: %v", err)
	}
	if !strings.Contains(psOut.String(), "app") && !strings.Contains(psOut.String(), `"State"`) {
		t.Fatalf("Ps json missing service: %s", psOut.String())
	}

	if err := app.Exec(ctx, &kit.Exec{
		Service: "app",
		Command: []string{"sh", "-c", "echo compose-ok"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	stopTimeout := 5
	if err := app.Stop(ctx, &kit.Stop{Timeout: &stopTimeout, ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	time.Sleep(time.Second)
	if err := app.Start(ctx, &kit.Start{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	restartTimeout := 5
	if err := app.Restart(ctx, &kit.Restart{Timeout: &restartTimeout, ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Restart: %v", err)
	}

	if err := app.Config(ctx, &kit.ComposeConfig{Quiet: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Config: %v", err)
	}
	if err := app.Images(ctx, &kit.Images{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Images: %v", err)
	}
	if err := app.Top(ctx, &kit.Top{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Top: %v", err)
	}
	if err := app.Stats(ctx, &kit.Stats{NoStream: true, ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if err := app.Pause(ctx, &kit.Pause{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if err := app.Unpause(ctx, &kit.Unpause{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Unpause: %v", err)
	}
	if err := app.Port(ctx, &kit.Port{Service: "app", PrivatePort: 80, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Port: %v", err)
	}
	if err := app.Ls(ctx, &kit.Ls{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Ls: %v", err)
	}
	if err := app.Version(ctx, &kit.Version{Short: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Version: %v", err)
	}
	if err := app.Volumes(ctx, &kit.Volumes{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Volumes: %v", err)
	}
	if err := app.Run(ctx, &kit.Run{
		Service: "app",
		Command: []string{"echo", "run-ok"},
		NoTTY:   true,
		Rm:      true,
		NoDeps:  true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	hostFile := filepath.Join(t.TempDir(), "hostname")
	if err := app.Cp(ctx, &kit.Cp{Source: "app:/etc/hostname", Dest: hostFile, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Cp: %v", err)
	}

	if err := app.Logs(ctx, &kit.Logs{Tail: intPtr(20), Writer: testWriter(t)}); err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if err := app.Pull(ctx, &kit.Pull{Policy: "missing", Quiet: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if err := app.Kill(ctx, &kit.Kill{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if err := app.Rm(ctx, &kit.Rm{Force: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Rm: %v", err)
	}

	if err := app.Down(ctx, &kit.Down{RemoveOrphans: true, RemoveVolumes: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Down: %v", err)
	}
}

func intPtr(v int) *int { return &v }

func TestComposeProfiles(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := kit.NewProject(name)
	project.WithService(kit.NewService("keep",
		kit.NewContainer().Image(alpine).Command("sleep", "infinity"),
	).Profiles("demo"))
	project.WithService(kit.NewService("skip",
		kit.NewContainer().Image(alpine).Command("sleep", "infinity"),
	).Profiles("other"))
	app := kit.NewCompose(project)
	composeDown(t, app, "demo")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err := app.Up(ctx, &kit.Up{Detach: true, Writer: testWriter(t)})
	if err == nil {
		t.Fatal("Up without Profiles: want error")
	}
	if !errdefs.IsComposeUpError(err) {
		t.Fatalf("Up without Profiles: want ComposeUpError, got %v", err)
	}
	if !strings.Contains(err.Error(), "Profiles") {
		t.Fatalf("Up without Profiles: unexpected error: %v", err)
	}

	if err := app.Up(ctx, &kit.Up{
		Detach:   true,
		Profiles: []string{"demo"},
		Writer:   testWriter(t),
	}); err != nil {
		t.Fatalf("Up with Profiles: %v", err)
	}

	var psOut bytes.Buffer
	if err := app.Ps(ctx, &kit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t)), Profiles: []string{"demo"}}); err != nil {
		t.Fatalf("Ps: %v", err)
	}
	raw := psOut.String()
	if !strings.Contains(raw, "keep") {
		t.Fatalf("expected keep service in ps: %s", raw)
	}
	if strings.Contains(raw, `"Service":"skip"`) || strings.Contains(raw, `"Service": "skip"`) {
		t.Fatalf("skip service should not start: %s", raw)
	}
}

func TestComposeInlineBuild(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	img := ident(t, "i") + ":e2e"
	df := kit.NewDockerFile().From("alpine:latest").Command("sleep", "infinity")
	project := kit.NewProject(name)
	project.WithService(kit.NewService("app",
		kit.NewContainer().Image(img),
	).Dockerfile(df))
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := kit.NewCompose(project)
	composeDown(t, app)
	cli := newCLI(t)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, img, &client.ImageRemove{Force: true, PruneChildren: true})
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := app.Build(ctx, &kit.ComposeBuild{Writer: testWriter(t), ServiceNames: []string{"app"}}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := app.Up(ctx, &kit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if err := app.Exec(ctx, &kit.Exec{
		Service: "app",
		Command: []string{"echo", "inline-ok"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec: %v", err)
	}
}

func TestComposeEvents(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := kit.NewProject(name)
	project.WithService(kit.NewService("web", kit.NewContainer().
		Image(alpine).
		Command("sleep", "infinity"),
	))
	app := kit.NewCompose(project)
	composeDown(t, app)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, errCh, err := app.Events(ctx, nil)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}

	got := make(chan string, 32)
	go func() {
		for e := range evCh {
			fmt.Fprintf(testWriter(t), "event action=%s service=%s type=%s id=%s\n", e.Action, e.Service, e.Type, e.ID)
			got <- e.Action
		}
	}()
	go func() {
		for range errCh {
		}
	}()

	upCtx, upCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer upCancel()
	if err := app.Up(upCtx, &kit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}

	deadline := time.After(30 * time.Second)
	seen := map[string]bool{}
	for !seen["create"] || !seen["start"] {
		select {
		case a := <-got:
			seen[a] = true
		case <-deadline:
			t.Fatalf("Events: want create and start, got %v", seen)
		}
	}

	if err := app.Down(upCtx, &kit.Down{RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Down: %v", err)
	}
	cancel()
}

func TestComposeStackNetworkVolumeHealthScaleWait(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)

	name := ident(t, "p")
	netName := ident(t, "n")
	volName := ident(t, "v")
	project := kit.NewProject(name)
	project.WithNetwork(netName)
	project.WithVolume(volName)
	project.WithService(kit.NewService("db",
		kit.NewContainer().
			Image(alpine).
			Command("sleep", "infinity").
			HealthCheck(kit.NewHealth().CMDShell("true").Interval(1).Timeout(1).Retries(5)).
			RWNamedVolumeMount(volName, "/data").
			Endpoint(netName),
	))
	project.WithService(kit.NewService("app",
		kit.NewContainer().
			Image(alpine).
			Command("sleep", "infinity").
			RWNamedVolumeMount(volName, "/data").
			Endpoint(netName),
	).DependsOnHealthy("db"))
	project.WithService(kit.NewService("oneshot",
		kit.NewContainer().
			Image(alpine).
			Command("sleep", "infinity").
			RestartPolicyNever().
			Endpoint(netName),
	))
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	app := kit.NewCompose(project)
	composeDown(t, app)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	waitTimeout := 90
	if err := app.Up(ctx, &kit.Up{
		Wait:          true,
		WaitTimeout:   &waitTimeout,
		RemoveOrphans: true,
		Writer:        testWriter(t),
	}); err != nil {
		t.Fatalf("Up wait: %v", err)
	}

	if err := app.Exec(ctx, &kit.Exec{
		Service: "db",
		Command: []string{"sh", "-c", "echo shared-ok > /data/ping"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec db write: %v", err)
	}
	var volOut bytes.Buffer
	if err := app.Exec(ctx, &kit.Exec{
		Service: "app",
		Command: []string{"cat", "/data/ping"},
		NoTTY:   true,
		Writer:  io.MultiWriter(&volOut, testWriter(t)),
	}); err != nil {
		t.Fatalf("Exec app read: %v", err)
	}
	if !strings.Contains(volOut.String(), "shared-ok") {
		t.Fatalf("volume not shared: %s", volOut.String())
	}

	var netOut bytes.Buffer
	if err := app.Exec(ctx, &kit.Exec{
		Service: "app",
		Command: []string{"getent", "hosts", "db"},
		NoTTY:   true,
		Writer:  io.MultiWriter(&netOut, testWriter(t)),
	}); err != nil {
		t.Fatalf("Exec getent db: %v", err)
	}
	if !strings.Contains(netOut.String(), "db") {
		t.Fatalf("network dns missing db: %s", netOut.String())
	}

	if err := app.Scale(ctx, &kit.Scale{
		Replicas: []kit.UpScale{{Service: "app", Num: 2}},
		Writer:   testWriter(t),
	}); err != nil {
		t.Fatalf("Scale: %v", err)
	}
	var psOut bytes.Buffer
	if err := app.Ps(ctx, &kit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t))}); err != nil {
		t.Fatalf("Ps after scale: %v", err)
	}
	if n := countPSService(psOut.String(), "app"); n < 2 {
		t.Fatalf("Scale: want 2 app replicas, got %d in %s", n, psOut.String())
	}

	stopErr := make(chan error, 1)
	go func() {
		time.Sleep(time.Second)
		stopErr <- app.Stop(ctx, &kit.Stop{ServiceNames: []string{"oneshot"}, Writer: testWriter(t)})
	}()
	waitErr := app.Wait(ctx, &kit.Wait{ServiceNames: []string{"oneshot"}, Writer: testWriter(t)})
	if err := <-stopErr; err != nil {
		t.Fatalf("Stop oneshot: %v", err)
	}
	if waitErr != nil && !strings.Contains(waitErr.Error(), "exit status") {
		t.Fatalf("Wait oneshot: %v", waitErr)
	}
}

func TestComposeLoadProjectUp(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)

	name := ident(t, "p")
	src := kit.NewProject(name)
	src.WithService(kit.NewService("web", kit.NewContainer().
		Image(alpine).
		Command("sleep", "infinity"),
	))
	if err := src.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "compose.yaml")
	if err := src.Export(yamlPath, 0o644); err != nil {
		t.Fatalf("Export: %v", err)
	}
	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read compose.yaml: %v", err)
	}
	fmt.Fprintln(testWriter(t), string(raw))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	loader := kit.NewCompose(kit.NewProject("loader"))
	loaded, err := loader.LoadProject(ctx, &kit.ProjectLoad{
		Name:       name,
		Files:      []string{yamlPath},
		WorkingDir: dir,
		Writer:     testWriter(t),
	})
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if _, err := loaded.GetService("web"); err != nil {
		t.Fatalf("loaded web: %v", err)
	}

	app := kit.NewCompose(loaded)
	composeDown(t, app)
	if err := app.Up(ctx, &kit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up loaded: %v", err)
	}
	if err := app.Exec(ctx, &kit.Exec{
		Service: "web",
		Command: []string{"echo", "loaded-ok"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec loaded: %v", err)
	}
}

func countPSService(raw, service string) int {
	return strings.Count(raw, `"Service":"`+service+`"`) + strings.Count(raw, `"Service": "`+service+`"`)
}
