//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func TestComposeEchoUp(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := containerkit.NewProject(name)
	project.WithService("echo", containerkit.NewContainer().
		Image(alpine).
		Command("echo", "hello-e2e"),
	)
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := containerkit.NewCompose(project)
	composeDown(t, app)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := app.Up(ctx, &containerkit.Up{Writer: testWriter(t), RemoveOrphans: true}); err != nil {
		t.Fatalf("Up: %v", err)
	}
}

func TestComposeLifecycle(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := containerkit.NewProject(name)
	project.WithService("app", containerkit.NewContainer().
		Image(alpine).
		Command("sleep", "infinity").
		PortBindings("tcp", "127.0.0.1", "0", "80"),
	)
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := containerkit.NewCompose(project)
	composeDown(t, app)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := app.Up(ctx, &containerkit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}

	if err := app.Ps(ctx, &containerkit.Ps{All: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Ps: %v", err)
	}

	var psOut bytes.Buffer
	if err := app.Ps(ctx, &containerkit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t))}); err != nil {
		t.Fatalf("Ps json: %v", err)
	}
	if !strings.Contains(psOut.String(), "app") && !strings.Contains(psOut.String(), `"State"`) {
		t.Fatalf("Ps json missing service: %s", psOut.String())
	}

	if err := app.Exec(ctx, &containerkit.Exec{
		Service: "app",
		Command: []string{"sh", "-c", "echo compose-ok"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	stopTimeout := 5
	if err := app.Stop(ctx, &containerkit.Stop{Timeout: &stopTimeout, ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	time.Sleep(time.Second)
	if err := app.Start(ctx, &containerkit.Start{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	restartTimeout := 5
	if err := app.Restart(ctx, &containerkit.Restart{Timeout: &restartTimeout, ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Restart: %v", err)
	}

	if err := app.Config(ctx, &containerkit.ComposeConfig{Quiet: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Config: %v", err)
	}
	if err := app.Images(ctx, &containerkit.Images{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Images: %v", err)
	}
	if err := app.Top(ctx, &containerkit.Top{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Top: %v", err)
	}
	if err := app.Stats(ctx, &containerkit.Stats{NoStream: true, ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if err := app.Pause(ctx, &containerkit.Pause{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if err := app.Unpause(ctx, &containerkit.Unpause{ServiceNames: []string{"app"}, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Unpause: %v", err)
	}
	if err := app.Port(ctx, &containerkit.Port{Service: "app", PrivatePort: 80, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Port: %v", err)
	}
	if err := app.Ls(ctx, &containerkit.Ls{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Ls: %v", err)
	}
	if err := app.Version(ctx, &containerkit.Version{Short: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Version: %v", err)
	}
	if err := app.Volumes(ctx, &containerkit.Volumes{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Volumes: %v", err)
	}
	if err := app.Run(ctx, &containerkit.Run{
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
	if err := app.Cp(ctx, &containerkit.Cp{Source: "app:/etc/hostname", Dest: hostFile, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Cp: %v", err)
	}

	if err := app.Logs(ctx, &containerkit.Logs{Tail: intPtr(20), Writer: testWriter(t)}); err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if err := app.Pull(ctx, &containerkit.Pull{Policy: "missing", Quiet: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if err := app.Kill(ctx, &containerkit.Kill{Writer: testWriter(t)}); err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if err := app.Rm(ctx, &containerkit.Rm{Force: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Rm: %v", err)
	}

	if err := app.Down(ctx, &containerkit.Down{RemoveOrphans: true, RemoveVolumes: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Down: %v", err)
	}
}

func intPtr(v int) *int { return &v }

func TestComposeProfiles(t *testing.T) {
	t.Parallel()
	name := ident(t, "p")
	project := containerkit.NewProject(name)
	project.WithService("keep",
		containerkit.NewContainer().Image(alpine).Command("sleep", "infinity"),
		containerkit.Profiles("demo"),
	)
	project.WithService("skip",
		containerkit.NewContainer().Image(alpine).Command("sleep", "infinity"),
		containerkit.Profiles("other"),
	)
	app := containerkit.NewCompose(project)
	composeDown(t, app, "demo")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	err := app.Up(ctx, &containerkit.Up{Detach: true, Writer: testWriter(t)})
	if err == nil {
		t.Fatal("Up without Profiles: want error")
	}
	if !containerkit.IsComposeUpError(err) {
		t.Fatalf("Up without Profiles: want ComposeUpError, got %v", err)
	}
	if !strings.Contains(err.Error(), "Profiles") {
		t.Fatalf("Up without Profiles: unexpected error: %v", err)
	}

	if err := app.Up(ctx, &containerkit.Up{
		Detach:   true,
		Profiles: []string{"demo"},
		Writer:   testWriter(t),
	}); err != nil {
		t.Fatalf("Up with Profiles: %v", err)
	}

	var psOut bytes.Buffer
	if err := app.Ps(ctx, &containerkit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t)), Profiles: []string{"demo"}}); err != nil {
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
	df := containerkit.NewDockerFile().From("alpine", "latest").CommandExec("sleep", "infinity")
	project := containerkit.NewProject(name)
	project.WithService("app",
		containerkit.NewContainer().Image(img),
		df.WithInline(),
	)
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := containerkit.NewCompose(project)
	composeDown(t, app)
	cli := newCLI(t)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, img, &client.ImageRemove{Force: true, PruneChildren: true})
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := app.Build(ctx, &containerkit.ComposeBuild{Writer: testWriter(t), ServiceNames: []string{"app"}}); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := app.Up(ctx, &containerkit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if err := app.Exec(ctx, &containerkit.Exec{
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
	project := containerkit.NewProject(name)
	project.WithService("web", containerkit.NewContainer().
		Image(alpine).
		Command("sleep", "infinity"),
	)
	app := containerkit.NewCompose(project)
	composeDown(t, app)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	evCh, errCh, err := app.Events(ctx, "")
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
	if err := app.Up(upCtx, &containerkit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
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

	if err := app.Down(upCtx, &containerkit.Down{RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Down: %v", err)
	}
	cancel()
}
