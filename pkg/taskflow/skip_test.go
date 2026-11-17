package taskflow

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	cerrdefs "github.com/containerd/errdefs"
)

func alpineFile() *kit.Dockerfile {
	return kit.NewDockerfile().From("alpine", "latest").Command("true")
}

func buildBody() io.ReadCloser {
	return io.NopCloser(strings.NewReader("{\"stream\":\"ok\\n\"}\n"))
}

func TestSkipIfPresentHitRunsSmoke(t *testing.T) {
	ctx := context.Background()
	eng := &scriptedEngine{images: map[string]error{"app:ci": nil, "base:1": nil}}
	f := New("hit", newClient(eng, nil))
	build := f.Build("build", alpineFile(), kit.Tag("app:ci"), kit.Tag("base:1"), kit.NoCache(true)).SkipIfPresent()
	smoke := 0
	f.Task("smoke", func(context.Context) error {
		smoke++
		return nil
	}).DependsOn(build)
	var dones []Event
	var skips int
	err := f.Run(ctx, &Run{Writer: io.Discard, Watch: func(e Event) {
		if e.Name != "build" {
			return
		}
		if e.Kind == EventDone {
			dones = append(dones, e)
		}
		if e.Kind == EventSkip {
			skips++
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	if smoke != 1 {
		t.Fatalf("smoke ran %d", smoke)
	}
	if eng.builds != 0 {
		t.Fatalf("builds %d", eng.builds)
	}
	if skips != 0 || len(dones) != 1 || dones[0].Reason != ReasonAlreadyPresent {
		t.Fatalf("events skips=%d dones=%+v", skips, dones)
	}
	if build.Status() != StatusDone || build.Err() != nil {
		t.Fatalf("status %s err %v", build.Status(), build.Err())
	}
	if !slices.Equal(eng.inspects, []string{"app:ci", "base:1"}) {
		t.Fatalf("inspects %v", eng.inspects)
	}
}

func TestSkipIfPresentBuildsOnceWhenATagIsMissing(t *testing.T) {
	ctx := context.Background()
	eng := &scriptedEngine{
		buildBody: buildBody(),
		images:    map[string]error{"app:ci": cerrdefs.ErrNotFound, "base:1": nil},
	}
	f := New("miss", newClient(eng, nil))
	build := f.Build("build", alpineFile(), kit.Tag("app:ci"), kit.Tag("base:1"), kit.NoCache(true)).SkipIfPresent()
	smoke := 0
	f.Task("smoke", func(context.Context) error {
		smoke++
		return nil
	}).DependsOn(build)
	var reason string
	if err := f.Run(ctx, &Run{Writer: io.Discard, Watch: func(e Event) {
		if e.Name == "build" && e.Kind == EventDone {
			reason = e.Reason
		}
	}}); err != nil {
		t.Fatal(err)
	}
	if smoke != 1 || eng.builds != 1 || reason != "" {
		t.Fatalf("smoke %d builds %d reason %q", smoke, eng.builds, reason)
	}
	if !eng.built.NoCache || !slices.Equal(eng.built.Tags, []string{"app:ci", "base:1"}) {
		t.Fatalf("built %+v", eng.built)
	}
	if !slices.Equal(eng.inspects, []string{"app:ci", "base:1"}) {
		t.Fatalf("inspects %v", eng.inspects)
	}
	if build.Status() != StatusDone {
		t.Fatal(build.Status())
	}
}

func TestSkipIfPresentInspectErrorIgnoresTagOrder(t *testing.T) {
	orders := [][]string{{"gone", "boom"}, {"boom", "gone"}}
	for _, tags := range orders {
		eng := &scriptedEngine{images: map[string]error{
			"gone": cerrdefs.ErrNotFound,
			"boom": errors.New("boom"),
		}}
		f := New("ord", newClient(eng, nil))
		opts := []kit.EngineBuildOption{alpineFile()}
		for _, tag := range tags {
			opts = append(opts, kit.Tag(tag))
		}
		build := f.Build("build", opts...).SkipIfPresent()
		smoke := 0
		f.Task("smoke", func(context.Context) error {
			smoke++
			return nil
		}).DependsOn(build)
		err := f.Run(context.Background(), &Run{Writer: io.Discard})
		if err == nil || !strings.Contains(err.Error(), "boom") {
			t.Fatal(err)
		}
		if smoke != 0 || eng.builds != 0 {
			t.Fatalf("order %v smoke %d builds %d", tags, smoke, eng.builds)
		}
		if !slices.Equal(eng.inspects, tags) {
			t.Fatalf("order %v inspects %v", tags, eng.inspects)
		}
		if build.Status() != StatusError {
			t.Fatal(build.Status())
		}
	}
}

func TestSkipIfPresentKeepsTheFirstInspectError(t *testing.T) {
	eng := &scriptedEngine{images: map[string]error{
		"a": errors.New("first"),
		"b": errors.New("second"),
	}}
	f := New("hard", newClient(eng, nil))
	f.Build("build", alpineFile(), kit.Tag("a"), kit.Tag("b")).SkipIfPresent()
	err := f.Run(context.Background(), &Run{Writer: io.Discard})
	if err == nil || !strings.Contains(err.Error(), "first") || strings.Contains(err.Error(), "second") {
		t.Fatal(err)
	}
	if eng.builds != 0 || !slices.Equal(eng.inspects, []string{"a", "b"}) {
		t.Fatalf("builds %d inspects %v", eng.builds, eng.inspects)
	}
}

func TestSkipIfPresentRejectsMissingTagsBeforeInspect(t *testing.T) {
	eng := &scriptedEngine{}
	f := New("none", newClient(eng, nil))
	f.Build("build", alpineFile()).SkipIfPresent()
	err := f.Run(context.Background(), &Run{Writer: io.Discard})
	if err == nil || !strings.Contains(err.Error(), "no tag") {
		t.Fatal(err)
	}
	if eng.builds != 0 || len(eng.inspects) != 0 {
		t.Fatalf("builds %d inspects %v", eng.builds, eng.inspects)
	}

	eng = &scriptedEngine{}
	f = New("empty", newClient(eng, nil))
	f.Build("build", alpineFile(), kit.Tag("")).SkipIfPresent()
	err = f.Run(context.Background(), &Run{Writer: io.Discard})
	if err == nil || strings.Contains(err.Error(), "no tag") {
		t.Fatal(err)
	}
	if eng.builds != 0 || len(eng.inspects) != 0 {
		t.Fatalf("empty tag builds %d inspects %v", eng.builds, eng.inspects)
	}

	eng = &scriptedEngine{}
	f = New("nilopt", newClient(eng, nil))
	var opt kit.EngineBuildOption
	f.Build("build", alpineFile(), opt).SkipIfPresent()
	err = f.Run(context.Background(), &Run{Writer: io.Discard})
	if err == nil {
		t.Fatal("nil option")
	}
	if eng.builds != 0 || len(eng.inspects) != 0 {
		t.Fatalf("nil option builds %d inspects %v", eng.builds, eng.inspects)
	}

	eng = &scriptedEngine{}
	f = New("retry", newClient(eng, nil))
	f.Build("build", alpineFile()).SkipIfPresent().Retry(2)
	var retries int
	err = f.Run(context.Background(), &Run{Writer: io.Discard, Watch: func(e Event) {
		if e.Kind == EventRetry {
			retries++
		}
	}})
	if err == nil || !strings.Contains(err.Error(), "no tag") || retries != 0 {
		t.Fatalf("retries %d err %v", retries, err)
	}
}

func TestSkipIfPresentRejectsNonBuild(t *testing.T) {
	eng := &scriptedEngine{}
	f := New("job", newClient(eng, nil))
	step := f.Container("job", kit.NewContainer(kit.Image("app")), nil).SkipIfPresent().SkipIfPresent().Retry(2)
	step.noteConfig(nil)
	var retries int
	err := f.Run(context.Background(), &Run{Writer: io.Discard, Watch: func(e Event) {
		if e.Kind == EventRetry {
			retries++
		}
	}})
	if err == nil || !strings.Contains(err.Error(), "build step") || retries != 0 || eng.seq != 0 {
		t.Fatalf("seq %d retries %d err %v", eng.seq, retries, err)
	}
	if step.Status() != StatusError {
		t.Fatal(step.Status())
	}
	_, err = Import([]byte(`{"name":"x","steps":[{"name":"job","op":"containerRun","skipIfPresent":true,"options":{"image":"app"}}]}`), nil)
	if err == nil || !strings.Contains(err.Error(), "buildImage") {
		t.Fatal(err)
	}
}

func TestImportSkipIfPresent(t *testing.T) {
	doc := `{
	  "name": "hit",
	  "steps": [
	    {"name": "build", "op": "buildImage", "skipIfPresent": true, "options": {"parseDockerfile": "FROM alpine:latest\nCMD [\"true\"]\n", "tag": "app:ci"}},
	    {"name": "smoke", "op": "tagImage", "options": {"source": "app:ci", "target": "app:ci"}, "dependsOn": ["build"]}
	  ]
	}`
	eng := &scriptedEngine{images: map[string]error{"app:ci": nil}}
	f, err := Import([]byte(doc), newClient(eng, &fakeComposer{}))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(context.Background(), &Run{Writer: io.Discard}); err != nil {
		t.Fatal(err)
	}
	if eng.builds != 0 || eng.tags != 1 {
		t.Fatalf("builds %d tags %d", eng.builds, eng.tags)
	}
}
