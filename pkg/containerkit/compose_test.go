package containerkit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
	require.True(t, IsComposeExecError(empty.Exec(ctx, nil)))
	require.True(t, IsComposeExecError(empty.Exec(ctx, &Exec{Writer: w, Service: "a"})))
	require.True(t, IsComposeAttachError(empty.Attach(ctx, nil)))
	require.True(t, IsComposeCommitError(empty.Commit(ctx, nil)))
	require.True(t, IsComposeCpError(empty.Cp(ctx, nil)))
	require.True(t, IsComposeExportError(empty.Export(ctx, nil)))
	require.True(t, IsComposePortError(empty.Port(ctx, nil)))
	require.True(t, IsComposePortError(empty.Port(ctx, &Port{Writer: w, Service: "web"})))
	require.True(t, IsComposePublishError(empty.Publish(ctx, nil)))
	require.True(t, IsComposeRunError(empty.Run(ctx, nil)))
	require.True(t, IsComposeScaleError(empty.Scale(ctx, nil)))
	require.True(t, IsComposeScaleError(empty.Scale(ctx, &Scale{Writer: w, Replicas: []UpScale{{Service: "", Num: 1}}})))
	require.True(t, IsComposeWaitError(empty.Wait(ctx, nil)))
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
	_, _, err := empty.Events(ctx, "web")
	require.Error(t, err)

	p := NewProject("ok")
	p.WithService("web", NewContainer().Image("alpine:latest"))
	app := NewCompose(p)
	cmd, err := app.command(ctx, w, []string{"ps"}, nil, nil)
	require.NoError(t, err)
	require.Contains(t, cmd.Args, "compose")
	require.Contains(t, cmd.Args, "-f")

	_, err = app.command(ctx, w, []string{"up"}, []string{"demo"}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Profiles")

	p2 := NewProject("prof")
	p2.WithService("web", NewContainer().Image("alpine:latest"), Profiles("demo"))
	app2 := NewCompose(p2)
	_, err = app2.command(ctx, w, []string{"up"}, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "profiles found")

	cmd, err = app2.command(ctx, w, []string{"exec"}, []string{"demo"}, strings.NewReader("stdin"))
	require.NoError(t, err)
	require.NotNil(t, cmd.Stdin)
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
		require.True(t, IsComposeEventsError(err))
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
