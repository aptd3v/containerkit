//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func TestHybridImageBuildThenCompose(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)

	tag := ident(t, "i") + ":e2e"
	df := containerkit.NewDockerFile().
		From("alpine", "latest").
		Run("echo hybrid > /hello.txt").
		CommandExec("sleep", "infinity")
	ctxReader, err := df.NewLocalBuildContext(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalBuildContext: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	buildOpt := &client.ImageBuild{}
	buildOpt.Tags = []string{tag}
	buildOpt.Remove = true
	resp, err := cli.ImageBuild(ctx, ctxReader, buildOpt)
	if err != nil {
		t.Fatalf("ImageBuild: %v", err)
	}
	drainBuild(t, resp.Body)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, tag, &client.ImageRemove{Force: true, PruneChildren: true})
	})

	project := containerkit.NewProject(ident(t, "p"))
	project.WithService("app", containerkit.NewContainer().Image(tag))
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := containerkit.NewCompose(project)
	composeDown(t, app)

	if err := app.Up(ctx, &containerkit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}

	var psOut bytes.Buffer
	if err := app.Ps(ctx, &containerkit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t))}); err != nil {
		t.Fatalf("Ps: %v", err)
	}
	if !strings.Contains(psOut.String(), "app") {
		t.Fatalf("Ps missing app: %s", psOut.String())
	}

	if err := app.Exec(ctx, &containerkit.Exec{
		Service: "app",
		Command: []string{"cat", "/hello.txt"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	if err := app.Down(ctx, &containerkit.Down{RemoveOrphans: true, RemoveVolumes: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Down: %v", err)
	}
}
