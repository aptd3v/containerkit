package taskflow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	clienttypes "github.com/moby/moby/client"
)

func TestContainerStep(t *testing.T) {
	ctx := context.Background()
	eng := &scriptedEngine{logs: mux("hello\n"), id: "abc"}
	comp := &fakeComposer{}
	cli := newClient(eng, comp)
	app := sampleApp()
	app.Services()[0].Apply(kit.ContainerName("db-1"))
	f := New("ci", cli)
	up := f.Compose("up", app, Up())
	ctr := kit.NewContainer(kit.Image("app:ci"), kit.Command("true"))
	step := f.Container("test", ctr, nil)
	up.Precede(step)
	var logs bytes.Buffer
	var lines [][]byte
	if err := f.Run(ctx, &Run{Writer: &logs, Log: func(step string, p []byte) { lines = append(lines, append([]byte(nil), p...)) }}); err != nil {
		t.Fatal(err)
	}
	if step.ID() != "abc" || step.ExitCode() != 0 {
		t.Fatalf("id %s code %d", step.ID(), step.ExitCode())
	}
	if eng.created.Config == nil || eng.created.Config.Labels[LabelFlow] != "ci" || eng.created.Config.Labels[LabelTask] != "test" {
		t.Fatalf("labels %#v", eng.created.Config)
	}
	if !strings.Contains(eng.created.Config.Labels[LabelDepends], "db-1") {
		t.Fatalf("depends %#v", eng.created.Config.Labels)
	}
	if !bytes.Contains(logs.Bytes(), []byte("hello")) || len(lines) == 0 {
		t.Fatalf("logs %q lines %d", logs.Bytes(), len(lines))
	}
	yaml, err := app.ExportYAML()
	if err != nil || !bytes.Contains(yaml, []byte(LabelFlow)) {
		t.Fatalf("compose labels %v %s", err, yaml)
	}
}

func TestContainerFailures(t *testing.T) {
	ctx := context.Background()
	f := New("none", nil)
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	f.Build("b", kit.NewDockerfile().From("alpine", "latest"))
	if err := f.Run(ctx, discardRun()); !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}

	eng := &scriptedEngine{}
	f = New("job", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app"), kit.AutoRemove(true)), nil)
	if err := f.Run(ctx, discardRun()); err == nil || !strings.Contains(err.Error(), "autoremove") {
		t.Fatal(err)
	}
	for _, pol := range []kit.RestartPolicy{kit.Always, kit.UnlessStopped, kit.OnFailure} {
		f = New("job", newClient(&scriptedEngine{}, nil))
		f.Container("c", kit.NewContainer(kit.Image("app"), kit.Restart(kit.Policy(pol))), nil)
		if err := f.Run(ctx, discardRun()); err == nil {
			t.Fatal(pol)
		}
	}
	if err := rejectJobPolicy(false, "bogus"); err == nil {
		t.Fatal("bogus policy")
	}

	f = New("missing", newClient(&scriptedEngine{}, nil))
	f.Container("c", kit.NewContainer(), nil)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("missing image")
	}

	f = New("nilctr", newClient(&scriptedEngine{}, nil))
	f.Container("c", nil, nil)
	if err := f.Run(ctx, discardRun()); err == nil || !strings.Contains(err.Error(), "container is nil") {
		t.Fatal(err)
	}

	eng = &scriptedEngine{createErr: cerrdefs.ErrConflict}
	f = New("conflict", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Name("c"), kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); !errors.Is(err, kit.ErrConflict) || len(eng.removed) != 0 {
		t.Fatalf("conflict %v removed %v", err, eng.removed)
	}

	eng = &scriptedEngine{id: "busyid", startErr: errors.New("start"), removeErr: errors.New("busy")}
	f = New("rm", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil).Retry(1)
	if err := f.Run(ctx, discardRun()); err == nil || !strings.Contains(err.Error(), "busy") || len(eng.removed) != 1 || eng.removed[0] != "busyid" {
		t.Fatalf("retry remove %v %v", err, eng.removed)
	}

	eng = &scriptedEngine{id: "gone", startErr: errors.New("start"), removeErr: cerrdefs.ErrNotFound}
	f = New("notfound", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil).Retry(1)
	if err := f.Run(ctx, discardRun()); err == nil || !strings.Contains(err.Error(), "start") {
		t.Fatal(err)
	}

	eng = &scriptedEngine{createErr: errors.New("create")}
	f = New("create", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("create")
	}

	eng = &scriptedEngine{startErr: errors.New("start")}
	f = New("start", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("start")
	}

	eng = &scriptedEngine{wait: waitResult(2, "")}
	f = New("exit", newClient(eng, nil))
	step := f.Container("c", kit.NewContainer(kit.Image("app")), &ContainerRun{Logs: io.Discard})
	if err := f.Run(ctx, discardRun()); err == nil || step.ExitCode() != 2 {
		t.Fatalf("exit %v code %d", err, step.ExitCode())
	}

	eng = &scriptedEngine{wait: waitResult(0, "oom")}
	f = New("oom", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err == nil || !strings.Contains(err.Error(), "oom") {
		t.Fatal(err)
	}

	eng = &scriptedEngine{logsErr: errors.New("logs")}
	f = New("logs", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("logs")
	}

	eng = &scriptedEngine{logs: io.NopCloser(strings.NewReader("not multiplexed"))}
	f = New("badlogs", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("stdcopy")
	}
}

func TestContainerOwnership(t *testing.T) {
	ctx := context.Background()
	eng := &scriptedEngine{}
	cli := newClient(eng, nil)
	one := New("one", cli)
	one.Container("migrate", kit.NewContainer(kit.Image("app")), nil)
	if err := one.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	firstName := eng.created.Name
	if !strings.HasPrefix(firstName, "one-migrate-") || firstName == "migrate" {
		t.Fatal(firstName)
	}
	eng.removed = nil
	two := New("two", cli)
	two.Container("migrate", kit.NewContainer(kit.Image("app")), nil)
	if err := two.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	for _, id := range eng.removed {
		if id == firstName || id == "ctr-1" {
			t.Fatalf("removed %s from %v", id, eng.removed)
		}
	}
	if eng.created.Name == firstName || !strings.HasPrefix(eng.created.Name, "two-migrate-") {
		t.Fatal(eng.created.Name)
	}

	eng = &scriptedEngine{id: "made", startErr: errors.New("start")}
	f := New("job", newClient(eng, nil))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil).Retry(1)
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("start")
	}
	if len(eng.removed) != 1 || eng.removed[0] != "made" {
		t.Fatalf("retry removes %v", eng.removed)
	}

	eng = &scriptedEngine{}
	f = New("edge", newClient(eng, nil))
	earlier := f.Container("a", kit.NewContainer(kit.Image("app")), nil)
	f.Container("b", kit.NewContainer(kit.Image("app")), nil)
	earlier.Precede(f.taskList()[1])
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	cfg, _, _, err := earlier.container.EngineConfig()
	if err != nil || !strings.Contains(cfg.Labels[LabelPrecedes], "edge-b-") {
		t.Fatalf("precedes %v %v", err, cfg)
	}
}

func TestRunSweep(t *testing.T) {
	ctx := context.Background()
	eng := &scriptedEngine{}
	cli := newClient(eng, nil)
	early := New("ci", cli)
	if _, err := early.List(ctx); err != nil {
		t.Fatal(err)
	}
	if err := early.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	if eng.lists != 0 {
		t.Fatalf("lists before run %d", eng.lists)
	}

	a := New("ci", cli)
	stepA := a.Container("a", kit.NewContainer(kit.Image("app")), nil)
	b := New("ci", cli)
	stepB := b.Container("b", kit.NewContainer(kit.Image("app")), nil)
	if err := a.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if err := b.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if stepA.ID() == "" || stepA.ID() == stepB.ID() {
		t.Fatalf("ids %s %s", stepA.ID(), stepB.ID())
	}
	var labels map[string]string
	for _, c := range eng.live {
		if c.id == stepA.ID() {
			labels = c.labels
		}
	}
	if labels[LabelFlow] != "ci" || labels[LabelRun] == "" {
		t.Fatalf("labels %#v", labels)
	}
	eng.live = append(eng.live, scriptedContainer{id: "left", labels: map[string]string{
		LabelFlow: labels[LabelFlow],
		LabelRun:  labels[LabelRun],
	}})
	if err := a.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	left := map[string]bool{}
	for _, c := range eng.live {
		left[c.id] = true
	}
	if left[stepA.ID()] || left["left"] || !left[stepB.ID()] {
		t.Fatalf("live %#v removed %v", left, eng.removed)
	}
}

func TestSkippedContainer(t *testing.T) {
	ctx := context.Background()
	eng := &scriptedEngine{}
	f := New("skip", newClient(eng, nil))
	bad := f.Task("bad", func(context.Context) error { return errors.New("no") })
	step := f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	bad.Precede(step)
	var saw Status
	err := f.Run(ctx, &Run{Writer: io.Discard, Watch: func(e Event) {
		if e.Kind == EventSkip {
			saw = e.Task.Status()
		}
	}})
	if err == nil || step.Status() != StatusSkip || step.Err() != nil || step.ExitCode() != 0 || saw != StatusSkip {
		t.Fatalf("skip %v status %s err %v code %d saw %s", err, step.Status(), step.Err(), step.ExitCode(), saw)
	}
	if eng.created.Name != "" || len(eng.live) != 0 {
		t.Fatalf("created %#v live %d", eng.created, len(eng.live))
	}
}

func TestWaitStatus(t *testing.T) {
	ctx := context.Background()
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{}); err == nil {
		t.Fatal("empty")
	}
	cancel, stop := context.WithCancel(ctx)
	stop()
	if _, err := waitStatus(cancel, kit.ContainerWaitResult{}); err == nil {
		t.Fatal("canceled empty")
	}
	errCh := make(chan error, 1)
	errCh <- errors.New("wait")
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{Error: errCh}); err == nil {
		t.Fatal("error channel")
	}
	closed := make(chan error)
	close(closed)
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{Error: closed}); err == nil {
		t.Fatal("closed error")
	}
	openErr := make(chan error)
	ready := make(chan container.WaitResponse, 1)
	ready <- container.WaitResponse{StatusCode: 0}
	code, err := waitStatus(ctx, kit.ContainerWaitResult{Error: openErr, Result: ready})
	if err != nil || code != 0 {
		t.Fatalf("result while error stays open: %v %d", err, code)
	}
	// The nil send is the only ready case. The result follows it, so select
	// takes the error and readWait receives the exit.
	nilErr := make(chan error)
	followed := make(chan container.WaitResponse)
	go func() {
		nilErr <- nil
		followed <- container.WaitResponse{StatusCode: 4}
	}()
	code, err = waitStatus(ctx, kit.ContainerWaitResult{Error: nilErr, Result: followed})
	if err != nil || code != 4 {
		t.Fatalf("nil error then result: %v %d", err, code)
	}
	closedRes := make(chan container.WaitResponse)
	close(closedRes)
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{Result: closedRes}); err == nil {
		t.Fatal("closed result")
	}
	stopCtx, stop := context.WithCancel(ctx)
	stop()
	errCh = make(chan error, 1)
	errCh <- nil
	if _, err := waitStatus(stopCtx, kit.ContainerWaitResult{Error: errCh, Result: make(chan container.WaitResponse)}); err == nil {
		t.Fatal("cancel after empty error")
	}
	block := make(chan container.WaitResponse)
	if _, err := waitStatus(stopCtx, kit.ContainerWaitResult{Result: block}); err == nil {
		t.Fatal("cancel on result")
	}
	if _, err := waitStatus(cancel, kit.ContainerWaitResult{Error: make(chan error)}); err == nil {
		t.Fatal("cancel while only the error channel is open")
	}
	if _, err := waitStatus(cancel, kit.ContainerWaitResult{Error: make(chan error), Result: make(chan container.WaitResponse)}); err == nil {
		t.Fatal("cancel while both channels are open")
	}
	engineErr := make(chan error, 1)
	engineErr <- errors.New("engine")
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{Error: engineErr, Result: make(chan container.WaitResponse)}); err == nil || err.Error() != "engine" {
		t.Fatalf("error beside an unread result: %v", err)
	}
	closedBeside := make(chan container.WaitResponse)
	close(closedBeside)
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{Error: make(chan error), Result: closedBeside}); err == nil || err.Error() != "wait: result channel closed" {
		t.Fatalf("closed result beside an open error: %v", err)
	}
	if msg := exitMessage(nil); msg != "" {
		t.Fatal(msg)
	}
}

func TestBuildStep(t *testing.T) {
	ctx := context.Background()
	body := io.NopCloser(strings.NewReader("{\"stream\":\"built\\n\"}\n"))
	eng := &scriptedEngine{buildBody: body}
	f := New("build", newClient(eng, nil))
	f.Build("b", kit.NewDockerfile().From("alpine", "latest").Command("true"), kit.Tag("app:ci"))
	var buf bytes.Buffer
	if err := f.Run(ctx, &Run{Writer: &buf}); err != nil {
		t.Fatal(err)
	}
	eng = &scriptedEngine{buildErr: errors.New("build")}
	f = New("build-err", newClient(eng, nil))
	f.Build("b", kit.NewDockerfile().From("alpine", "latest"))
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("build error")
	}
}

func TestStepWriter(t *testing.T) {
	f := New("w", nil)
	f.mu.Lock()
	f.writer = nil
	f.mu.Unlock()
	w := stepWriter(f, "s", nil)
	if _, err := w.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	_ = clienttypes.ContainerCreateOptions{}
}
