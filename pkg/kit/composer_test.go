package kit_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func TestNewComposerNilOption(t *testing.T) {
	t.Parallel()
	_, err := kit.NewComposer(kit.Host("http://localhost"), nil)
	if err == nil {
		t.Fatal("expected nil option")
	}
}

func TestNewComposerConstructsService(t *testing.T) {
	t.Parallel()
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})
	log := kit.NewLogger(kit.OutputStream(io.Discard), kit.Color(false), kit.Progress("quiet"))
	c, err := kit.NewComposer(
		kit.Host("http://localhost"),
		kit.HTTPClient(&http.Client{Transport: transport}),
		kit.Log(log),
		kit.MaxConcurrency(2),
		kit.OutputStream(io.Discard),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil composer")
	}
	var _ kit.Composer = c
}

func TestNewComposerInstallsLogBridge(t *testing.T) {
	t.Parallel()
	transport := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})
	var events []kit.LogEvent
	log := kit.NewFrameLogger(func(e kit.LogEvent) { events = append(events, e) })
	c, err := kit.NewComposer(
		kit.Host("http://localhost"),
		kit.HTTPClient(&http.Client{Transport: transport}),
		kit.Log(log),
	)
	if err != nil {
		t.Fatal(err)
	}
	app := kit.NewCompose("app").WithService(kit.NewService(kit.Name("api"), kit.Image("alpine:latest")))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Detach keeps the call from attaching logs. The context has no logger, so
	// the working "up" line is Start on the bridge installed by NewComposer.
	_ = app.Up(ctx, c, kit.Detach(true))
	var working int
	for _, e := range events {
		if e.Op == "up" && e.Status == kit.StatusWorking && e.ID == "" {
			working++
		}
	}
	if working != 1 {
		t.Fatalf("bridge start events = %d in %#v", working, events)
	}
}

func TestNewComposerSettersCompile(t *testing.T) {
	t.Parallel()
	var (
		_ kit.ComposerClientOption = kit.FromEnv()
		_ kit.ComposerClientOption = kit.Host("http://localhost")
		_ kit.ComposerClientOption = kit.Log(kit.NewLogger())
		_ kit.ComposerClientOption = kit.MaxConcurrency(4)
		_ kit.ComposerClientOption = kit.DryRun(true)
		_ kit.ComposerClientOption = kit.OutputStream(io.Discard)
	)
}
