package kit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit/errdefs"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestIgnoreOrphans(t *testing.T) {
	require.False(t, composeBool(""))
	require.True(t, composeBool("y"))
	require.True(t, composeBool("true"))

	p := NewProject("ok")
	p.WithService(NewService("web", NewContainer().Image("alpine:latest")))
	err := NewCompose(p).Up(context.Background(), &Up{IgnoreOrphans: true, RemoveOrphans: true, Writer: io.Discard})
	require.True(t, errdefs.IsComposeUpError(err))
	require.ErrorContains(t, err, "COMPOSE_IGNORE_ORPHANS")

	t.Setenv(composeIgnoreOrphans, "1")
	ignore, err := orphanMode(false, false, nil)
	require.NoError(t, err)
	require.True(t, ignore)
	ignore, err = orphanMode(false, false, map[string]string{composeIgnoreOrphans: "false"})
	require.NoError(t, err)
	require.False(t, ignore)
}

func TestComposeProject(t *testing.T) {
	p := NewProject("app")
	app := NewCompose(p)
	require.Same(t, p, app.Project())
	require.Nil(t, (*Compose)(nil).Project())
}

func TestProjectLoadOptions(t *testing.T) {
	app := NewCompose(NewProject("app"))
	app.Global(Global{ProjectDirectory: "/work", AllResources: true, Compatibility: true, EnvFile: []string{"a.env"}})
	var seen []UnsupportedAttribute
	opt := app.projectLoadOptions(&ProjectLoad{
		Name: "from-flag", Files: []string{"compose.yaml"}, EnvFiles: []string{"b.env"},
		Profiles: []string{"demo"}, Services: []string{"web"}, Offline: true,
		InsecureRegistries: []string{"localhost:5000"},
		OnUnsupported:      func(attrs []UnsupportedAttribute) { seen = attrs },
	})
	require.Equal(t, "from-flag", opt.ProjectName)
	require.Equal(t, "/work", opt.WorkingDir)
	require.Equal(t, []string{"a.env", "b.env"}, opt.EnvFiles)
	require.True(t, opt.All)
	require.True(t, opt.Compatibility)
	require.True(t, opt.Offline)
	require.Equal(t, []string{"localhost:5000"}, opt.OCI.InsecureRegistries)
	require.NotNil(t, opt.OnUnsupportedAttribute)
	opt.OnUnsupportedAttribute([]api.UnsupportedAttribute{{Service: "web", Path: "deploy.replicas", Reason: "swarm only"}})
	require.Equal(t, []UnsupportedAttribute{{Service: "web", Path: "deploy.replicas", Reason: "swarm only"}}, seen)

	_, err := (*Compose)(nil).LoadProject(context.Background(), nil)
	require.True(t, errdefs.IsComposeLoadError(err))
}

func TestVizIndentation(t *testing.T) {
	got, err := vizIndentation(nil, false)
	require.NoError(t, err)
	require.Equal(t, "\t", got)

	size := 2
	got, err = vizIndentation(&size, true)
	require.NoError(t, err)
	require.Equal(t, "  ", got)

	size = -1
	_, err = vizIndentation(&size, false)
	require.Error(t, err)
}

func TestWriteTransformers(t *testing.T) {
	var buf bytes.Buffer
	err := writeTransformers(&buf, nil, "table", true)
	require.NoError(t, err)
	require.Empty(t, buf.String())

	err = writeTransformers(&buf, nil, "table", false)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "IMAGE ID")
}

func TestDefaultWriterAndCancellation(t *testing.T) {
	require.Equal(t, io.Discard, defaultWriter(io.Discard))
	require.NotNil(t, defaultWriter(nil))

	require.NoError(t, handleContextCancellation(context.Background(), nil))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, handleContextCancellation(ctx, context.Canceled))
	require.NoError(t, handleContextCancellation(ctx, errors.New("wrapped")))

	dctx, dcancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer dcancel()
	<-dctx.Done()
	require.NoError(t, handleContextCancellation(dctx, context.DeadlineExceeded))
	require.NoError(t, handleContextCancellation(dctx, errors.New("late")))
	require.Error(t, handleContextCancellation(context.Background(), errors.New("real")))
}

func TestComposeCommandWithoutDaemon(t *testing.T) {
	empty := NewCompose(NewProject("empty"))
	ctx := context.Background()
	w := io.Discard

	require.Error(t, empty.Up(ctx, &Up{Writer: w}))
	require.Error(t, empty.Down(ctx, &Down{Writer: w}))
	require.Error(t, empty.Logs(ctx, &Logs{Writer: w}))
	require.Error(t, empty.Kill(ctx, &Kill{Writer: w}))
	require.Error(t, empty.Ps(ctx, &Ps{Writer: w}))
	require.Error(t, empty.Start(ctx, &Start{Writer: w}))
	require.Error(t, empty.Stop(ctx, &Stop{Writer: w}))
	require.Error(t, empty.Restart(ctx, &Restart{Writer: w}))
	require.Error(t, empty.Build(ctx, &ComposeBuild{Writer: w}))
	require.Error(t, empty.Pull(ctx, &Pull{Writer: w}))
	require.True(t, errdefs.IsComposeExecError(empty.Exec(ctx, nil)))
	require.True(t, errdefs.IsComposeExecError(empty.Exec(ctx, &Exec{Writer: w, Service: "a"})))
	require.True(t, errdefs.IsComposeAttachError(empty.Attach(ctx, nil)))
	require.True(t, errdefs.IsComposeCommitError(empty.Commit(ctx, nil)))
	require.True(t, errdefs.IsComposeCpError(empty.Cp(ctx, nil)))
	require.True(t, errdefs.IsComposeExportError(empty.Export(ctx, nil)))
	require.True(t, errdefs.IsComposePortError(empty.Port(ctx, nil)))
	require.True(t, errdefs.IsComposePortError(empty.Port(ctx, &Port{Writer: w, Service: "web"})))
	require.True(t, errdefs.IsComposePublishError(empty.Publish(ctx, nil)))
	require.True(t, errdefs.IsComposeRunError(empty.Run(ctx, nil)))
	require.True(t, errdefs.IsComposeScaleError(empty.Scale(ctx, nil)))
	require.True(t, errdefs.IsComposeScaleError(empty.Scale(ctx, &Scale{Writer: w, Replicas: []UpScale{{Service: "", Num: 1}}})))
	require.True(t, errdefs.IsComposeWaitError(empty.Wait(ctx, nil)))
	require.Error(t, empty.Create(ctx, &Create{Writer: w}))
	require.Error(t, empty.Images(ctx, &Images{Writer: w}))
	require.Error(t, empty.Pause(ctx, &Pause{Writer: w}))
	require.Error(t, empty.Unpause(ctx, &Unpause{Writer: w}))
	require.Error(t, empty.Top(ctx, &Top{Writer: w}))
	require.Error(t, empty.Stats(ctx, &Stats{Writer: w, NoStream: true}))
	require.Error(t, empty.Push(ctx, &Push{Writer: w}))
	require.Error(t, empty.Rm(ctx, &Rm{Writer: w, Force: true}))
	require.Error(t, empty.Volumes(ctx, &Volumes{Writer: w}))
	require.Error(t, empty.Watch(ctx, &Watch{Writer: w}))
	require.Error(t, empty.Viz(ctx, &Viz{Writer: w}))
	require.Error(t, empty.BridgeConvert(ctx, &BridgeConvert{Writer: w}))
	require.True(t, errdefs.IsComposeGenerateError(empty.Generate(ctx, nil)))
	require.True(t, errdefs.IsComposeGenerateError(empty.Generate(ctx, &Generate{Writer: w, Format: "xml", Containers: []string{"web"}})))
	require.True(t, errdefs.IsComposeBridgeTransformationsListError(empty.BridgeTransformationsList(ctx, &BridgeTransformationsList{Writer: w, Format: "xml"})))
	require.True(t, errdefs.IsComposeBridgeTransformationsCreateError(empty.BridgeTransformationsCreate(ctx, nil)))
	size := -1
	require.True(t, errdefs.IsComposeVizError(empty.Viz(ctx, &Viz{Writer: w, IndentationSize: &size})))
	_, _, err := empty.Events(ctx, &EventsOpt{Service: "web"})
	require.Error(t, err)

	envFile := filepath.Join(t.TempDir(), "stack.env")
	require.NoError(t, os.WriteFile(envFile, []byte("FROM_FILE=1\n"), 0o644))

	p := NewProject("ok")
	p.WithService(NewService("web", NewContainer().Image("alpine:latest")))
	app := NewCompose(p)
	parallel := 2
	app.Global(Global{
		Progress: "plain", ANSI: "never", Parallel: &parallel,
		AllResources: true, EnvFile: []string{envFile}, Compatibility: true,
		ProjectDirectory: "/work", DryRun: true,
	})
	project, err := app.prepare("ps", nil, nil, selectKeep)
	require.NoError(t, err)
	require.Equal(t, "/work", project.WorkingDir)
	require.Equal(t, "1", project.Environment["FROM_FILE"])
	require.Contains(t, fmt.Sprintf("%T", progressProcessor("plain", "never", io.Discard)), "plainWriter")
	require.Contains(t, fmt.Sprintf("%T", progressProcessor("quiet", "", io.Discard)), "quiet")

	_, err = app.prepare("up", []string{"demo"}, nil, selectKeep)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Profiles")

	p2 := NewProject("prof")
	p2.WithService(NewService("web", NewContainer().Image("alpine:latest")).Profiles("demo"))
	app2 := NewCompose(p2)
	_, err = app2.prepare("up", nil, nil, selectKeep)
	require.Error(t, err)
	require.Contains(t, err.Error(), "profiles found")

	var cfg bytes.Buffer
	require.NoError(t, app.Config(ctx, &ComposeConfig{PrintServices: true, Writer: &cfg}))
	require.Equal(t, "web\n", cfg.String())
}

func TestUpPrepareOmitsDependents(t *testing.T) {
	p := NewProject("rs")
	p.WithService(NewService("db-0", NewContainer().Image("mongo")))
	p.WithService(NewService("db-1", NewContainer().Image("mongo")).DependsOn("db-0"))
	p.WithService(NewService("ui", NewContainer().Image("mongo-express")).DependsOnHealthy("db-0"))
	app := NewCompose(p)
	project, err := app.prepare("up", nil, []string{"db-0", "db-1"}, selectIncludeDeps)
	require.NoError(t, err)
	require.Contains(t, project.Services, "db-0")
	require.Contains(t, project.Services, "db-1")
	require.NotContains(t, project.Services, "ui")
}

func TestEventsWriter(t *testing.T) {
	ctx := context.Background()
	ch := make(chan Events, 4)
	errCh := make(chan error, 4)
	w := newEventsWriter(ctx, ch, errCh)

	n, err := w.Write([]byte(`{"action":"create","service":"web","type":"container"}` + "\nnot-json\npartial"))
	require.NoError(t, err)
	require.Greater(t, n, 0)

	select {
	case e := <-ch:
		require.Equal(t, "create", e.Action)
		require.Equal(t, "web", e.Service)
	case <-time.After(time.Second):
		t.Fatal("expected event")
	}
	select {
	case err := <-errCh:
		require.True(t, errdefs.IsComposeEventsError(err))
	case <-time.After(time.Second):
		t.Fatal("expected parse error")
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	cw := newEventsWriter(canceled, make(chan Events), make(chan error, 1))
	_, err = cw.Write([]byte("{\"action\":\"x\"}\n"))
	require.Error(t, err)

	cw2 := newEventsWriter(canceled, make(chan Events, 1), make(chan error))
	_, err = cw2.Write([]byte("not-json\n"))
	require.Error(t, err)

	_ = bytes.Buffer{}
}
