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
	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestHybridImageBuildThenCompose(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)

	tag := ident(t, "i") + ":e2e"
	df := kit.NewDockerFile().
		From("alpine:latest").
		Run("echo hybrid > /hello.txt").
		Command("sleep", "infinity")
	ctxReader, err := kit.NewBuild().Dockerfile(df).Reader()
	if err != nil {
		t.Fatalf("Reader: %v", err)
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

	project := kit.NewProject(ident(t, "p"))
	project.WithService(kit.NewService("app", kit.NewContainer().Image(tag)))
	if err := project.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	app := kit.NewCompose(project)
	composeDown(t, app)

	if err := app.Up(ctx, &kit.Up{Detach: true, RemoveOrphans: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Up: %v", err)
	}

	var psOut bytes.Buffer
	if err := app.Ps(ctx, &kit.Ps{All: true, Format: "json", Writer: io.MultiWriter(&psOut, testWriter(t))}); err != nil {
		t.Fatalf("Ps: %v", err)
	}
	if !strings.Contains(psOut.String(), "app") {
		t.Fatalf("Ps missing app: %s", psOut.String())
	}

	if err := app.Exec(ctx, &kit.Exec{
		Service: "app",
		Command: []string{"cat", "/hello.txt"},
		NoTTY:   true,
		Writer:  testWriter(t),
	}); err != nil {
		t.Fatalf("Exec: %v", err)
	}

	if err := app.Down(ctx, &kit.Down{RemoveOrphans: true, RemoveVolumes: true, Writer: testWriter(t)}); err != nil {
		t.Fatalf("Down: %v", err)
	}
}
