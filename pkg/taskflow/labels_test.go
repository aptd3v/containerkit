package taskflow

import (
	"context"
	"errors"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
)

func TestListAndRemove(t *testing.T) {
	ctx := context.Background()
	if _, err := List(ctx, nil, "ci"); !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}
	if _, err := List(ctx, kit.NewClient(nil, nil), "ci"); !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}
	eng := &scriptedEngine{id: "made"}
	eng.list.Items = []container.Summary{{ID: "left"}}
	comp := &fakeComposer{}
	cli := newClient(eng, comp)
	app := sampleApp()
	f := New("ci", cli)
	f.Compose("up", app, Up())
	f.Compose("again", app, Up())
	step := f.Container("c", kit.NewContainer(kit.Image("app"), kit.Name("job")), nil)
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	listed, err := f.List(ctx)
	if err != nil || len(listed.Items) != 1 {
		t.Fatalf("list %v %#v", err, listed.Items)
	}
	if err := f.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	if comp.downs != 1 {
		t.Fatalf("downs %d", comp.downs)
	}
	if step.ID() != "made" {
		t.Fatal(step.ID())
	}

	eng = &scriptedEngine{listErr: errors.New("list")}
	f = New("list", newClient(eng, &fakeComposer{}))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if err := f.Remove(ctx); err == nil {
		t.Fatal("list error")
	}

	eng = &scriptedEngine{}
	f = New("busy", newClient(eng, &fakeComposer{}))
	f.Container("c", kit.NewContainer(kit.Image("app")), nil)
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	eng.removeErr = errors.New("busy")
	if err := f.Remove(ctx); err == nil {
		t.Fatal("remove error")
	}

	eng = &scriptedEngine{}
	eng.list.Items = []container.Summary{{ID: "gone"}}
	f = New("gone", newClient(eng, &fakeComposer{}))
	f.mu.Lock()
	f.created = []string{"", "kept"}
	f.mu.Unlock()
	eng.removeErr = cerrdefs.ErrNotFound
	if err := f.Remove(ctx); err != nil {
		t.Fatal(err)
	}

	f = New("nocli", nil)
	f.Compose("up", sampleApp(), func(context.Context, *kit.Compose) error { return nil })
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if err := f.Remove(ctx); !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}

	bare := New("bare", nil)
	if err := bare.Remove(ctx); err != nil {
		t.Fatal(err)
	}
	ran := New("ran", nil)
	ran.Task("a", func(context.Context) error { return nil })
	if err := ran.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	if _, err := ran.List(ctx); !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}
	comp = &fakeComposer{}
	f = New("down", newClient(&scriptedEngine{}, comp))
	f.Compose("up", sampleApp(), Up())
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}
	comp.err = errors.New("down")
	if err := f.Remove(ctx); err == nil {
		t.Fatal("down error")
	}
}
