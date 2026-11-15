package taskflow

import (
	"context"
	"errors"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestComposeCommands(t *testing.T) {
	ctx := context.Background()
	app := sampleApp()
	commands := []struct {
		name string
		cmd  any
		op   string
	}{
		{"up", Up(kit.Detach(true), kit.Wait(true)), "up"},
		{"down", Down(kit.Volumes(true)), "down"},
		{"build", ComposeBuild(), "build"},
		{"create", Create(), "create"},
		{"start", Start(), "start"},
		{"stop", Stop(), "stop"},
		{"restart", Restart(), "restart"},
		{"pull", Pull(), "pull"},
		{"push", Push(), "push"},
		{"logs", Logs(), "logs"},
		{"ps", Ps(), "ps"},
		{"kill", Kill(), "kill"},
		{"rm", Rm(), "remove"},
		{"run", RunService("db"), "run"},
		{"exec", Exec("db"), "exec"},
		{"attach", Attach("db"), "attach"},
		{"copy", Copy(), "copy"},
		{"pause", Pause(), "pause"},
		{"unpause", Unpause(), "unpause"},
		{"top", Top("db"), "top"},
		{"events", Events(), "events"},
		{"port", Port("db", 5432), "port"},
		{"publish", PublishOCI("app"), "publish"},
		{"images", Images(), "images"},
		{"watch", Watch(), "watch"},
		{"viz", Viz(), "viz"},
		{"wait", Wait(), "wait"},
		{"scale", Scale(), "scale"},
		{"export", Export("db"), "export"},
		{"commit", Commit("db"), "commit"},
		{"volumes", VolumeList(), "volumes"},
		{"max", MaxConcurrency(2), "max"},
		{"dry", DryRun(true), "dry"},
	}
	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			comp := &fakeComposer{}
			f := New(tc.name, newClient(&scriptedEngine{}, comp))
			f.Compose(tc.name, app, tc.cmd)
			if err := f.Run(ctx, discardRun()); err != nil {
				t.Fatal(err)
			}
			if comp.op != tc.op {
				t.Fatalf("op %q", comp.op)
			}
		})
	}

	comp := &fakeComposer{code: 3}
	f := New("code", newClient(&scriptedEngine{}, comp))
	f.Compose("run", app, RunService("db"))
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("exit code")
	}
	comp = &fakeComposer{err: errors.New("down")}
	f = New("cerr", newClient(&scriptedEngine{}, comp))
	f.Compose("up", app, Up())
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("composer error")
	}

	f = New("fn", newClient(&scriptedEngine{}, &fakeComposer{}))
	f.Compose("fn", app, func(context.Context, *kit.Compose) error { return nil })
	f.Compose("fn3", app, func(context.Context, *kit.Compose, kit.Composer) error { return nil })
	if err := f.Run(ctx, discardRun()); err != nil {
		t.Fatal(err)
	}

	f = New("bad", nil)
	f.Compose("bad", app, "nope")
	f.Compose("zero", app, ComposeCmd{})
	f.Compose("nilcmd", app, nil)
	f.Compose("nilapp", nil, Up())
	step := f.Compose("emptyfn", app, Up())
	step.composeFn = nil
	if err := f.Run(ctx, discardRun()); err == nil {
		t.Fatal("expected compose errors")
	}
}

func TestExitCodeHelper(t *testing.T) {
	if err := exitCode(0, nil); err != nil {
		t.Fatal(err)
	}
	if err := exitCode(0, errors.New("x")); err == nil || exitCode(2, nil) == nil {
		t.Fatal("exit")
	}
}
