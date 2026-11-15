package taskflow

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/moby/moby/api/types/container"
)

func TestCoverageGaps(t *testing.T) {
	ctx := context.Background()
	discardLogs{}.Log("a", "b")
	discardLogs{}.Err("a", "b")
	discardLogs{}.Status("a", "b")

	f := New("gaps", nil)
	f.scope().subflow("bad", func(*Subflow) error { return errors.New("instantiate") })
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("instantiate")
	}

	f = New("idx", nil)
	cond := f.Condition("c", nil)
	if cond.fallbackIndex() != 0 {
		t.Fatal("fallback")
	}
	f.ctx = nil
	if f.context() == nil {
		t.Fatal("context")
	}
	f.trackID("")
	cond.setErr(nil)
	f = New("nilopt", nil)
	f.Task("a", func(context.Context) error { return nil })
	if err := f.Run(ctx, nil); err != nil {
		t.Fatal(err)
	}

	log := kit.NewFrameLogger(func(kit.LogEvent) {})
	f = New("kinds", nil)
	f.mu.Lock()
	f.ctx = log.Context(ctx)
	f.mu.Unlock()
	f.emit(Event{Name: "a", Kind: EventDone})
	f.emit(Event{Name: "a", Kind: EventSkip})
	f.emit(Event{Name: "a", Kind: EventRetry})
	f.emit(Event{Name: "a", Kind: EventError})
	f.emitLog("a", []byte("line\n"))

	if _, err := kit.ExecFlow(ctx, nil, json.RawMessage(`{`)); err == nil {
		t.Fatal("exec import")
	}
	if err := wireBranch(f.scope(), f.Task("t", nil), json.RawMessage(`{`), true, nil, nil); err == nil {
		t.Fatal("branch")
	}
	if err := wireBranch(f.scope(), f.Task("gate", nil), json.RawMessage(`["missing"]`), true, nil, nil); err == nil || !strings.Contains(err.Error(), "unknown step") {
		t.Fatal(err)
	}
	if _, _, err := parseBranchElements([]json.RawMessage{nil}); err == nil || !strings.Contains(err.Error(), "empty branch element") {
		t.Fatal(err)
	}
	if err := importSubflow("inner", nil)(nil); err == nil || !strings.Contains(err.Error(), "scope is not ready") {
		t.Fatal(err)
	}

	when := true
	off := false
	parent := f.Task("parent", nil)
	if _, err := addSteps(f.scope(), []stepDoc{
		{Name: "gate", Op: "if", When: &when, Then: json.RawMessage(`["other"]`)},
		{Name: "other", Op: "if", When: &off, DependsOn: []string{"parent"}, Then: json.RawMessage(`["parent"]`), Else: json.RawMessage(`[{"name":"deep","op":"if","when":true,"then":[{"name":"deeper","op":"if","when":false}]}]`)},
	}, map[string]*Task{"parent": parent}); err != nil {
		t.Fatal(err)
	}
	thenDoc := `{"steps":[{"name":"build","op":"if","when":true,"then":[{"name":"bad","op":"buildImage","options":{"noCache":"nope"}}]}]}`
	elseDoc := `{"steps":[{"name":"box","op":"if","when":true,"else":[{"name":"bad","op":"containerRun","options":{"autoRemove":"nope"}}]}]}`
	if _, err := Import([]byte(thenDoc), nil); err == nil {
		t.Fatal("then decode")
	}
	if _, err := Import([]byte(elseDoc), nil); err == nil {
		t.Fatal("else decode")
	}
	stopped, stop := context.WithCancel(ctx)
	stop()
	if _, err := readWait(stopped, make(chan container.WaitResponse)); err == nil {
		t.Fatal("readWait cancel")
	}

	if _, err := List(nil, newClient(&scriptedEngine{}, nil), "ci"); err != nil {
		t.Fatal(err)
	}
	var none *Task
	if none.alreadyCreated() || none.containerName() != "" {
		t.Fatal("nil task")
	}
	plain := &Task{name: "plain", kind: kindTask}
	if plain.containerName() != "" {
		t.Fatal("not a container")
	}
	named := &Task{name: "job", kind: kindContainer, container: kit.NewContainer(kit.Name("job"), kit.Image("app"))}
	if named.containerName() != "job" {
		t.Fatal(named.containerName())
	}
	unnamed := &Task{name: "step", kind: kindContainer, container: kit.NewContainer(kit.Image("app"))}
	if unnamed.containerName() != "step" {
		t.Fatal(unnamed.containerName())
	}
	blank := &Task{kind: kindContainer, container: kit.NewContainer(kit.Image("app"))}
	if blank.containerName() != "" {
		t.Fatal("blank")
	}

	eng := &scriptedEngine{}
	cli := newClient(eng, &fakeComposer{})
	flow := New("labels", cli)
	first := flow.Container("first", kit.NewContainer(kit.Image("app"), kit.Name("first")), nil)
	mid := flow.Task("mid", func(context.Context) error { return nil })
	last := flow.Container("last", kit.NewContainer(kit.Image("app"), kit.Name("last")), nil)
	first.Precede(mid)
	mid.Precede(last)
	first.preds = append(first.preds, nil, first)
	if err := flow.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	cfg, _, _, err := first.container.EngineConfig()
	if err != nil || cfg.Labels[LabelPrecedes] == "" {
		t.Fatalf("precedes %#v err %v", cfg, err)
	}

	eng = &scriptedEngine{}
	eng.list.Items = []container.Summary{{ID: "left"}}
	flow = New("left", newClient(eng, &fakeComposer{}))
	flow.trackCompose(nil)
	flow.mu.Lock()
	flow.compose = append(flow.compose, nil)
	flow.mu.Unlock()
	flow.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := flow.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	eng.removeErr = errors.New("busy")
	if err := flow.Remove(ctx); err == nil {
		t.Fatal("leftover remove")
	}

	flow = New("space", newClient(&scriptedEngine{}, nil))
	flow.Container("bad name", kit.NewContainer(kit.Image("app")), nil)
	if err := flow.Run(ctx, discardRun()); err == nil {
		t.Fatal("bad name")
	}

	errCh := make(chan error, 1)
	errCh <- nil
	closed := make(chan container.WaitResponse)
	close(closed)
	if _, err := waitStatus(ctx, kit.ContainerWaitResult{Error: errCh, Result: closed}); err == nil {
		t.Fatal("closed follow")
	}
}
