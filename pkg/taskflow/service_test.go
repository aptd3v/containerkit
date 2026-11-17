package taskflow

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
	cerrdefs "github.com/containerd/errdefs"
)

func TestServiceOutlivesTheChain(t *testing.T) {
	comp := &fakeComposer{}
	f := New("scope", newClient(&scriptedEngine{}, comp))
	db := f.Service("db", sampleApp(), kit.Timeout(time.Second), kit.Wait(false), kit.Detach(false))
	migrate := f.Task("migrate", func(context.Context) error {
		if comp.ups != 1 || comp.downs != 0 {
			return errors.New("db not ready")
		}
		return nil
	}).DependsOn(db)
	f.Task("integration-test", func(context.Context) error {
		if comp.downs != 0 {
			return errors.New("down too early")
		}
		return errors.New("test failed")
	}).DependsOn(migrate)
	err := f.Run(context.Background(), discardRun())
	if err == nil || !strings.Contains(err.Error(), "test failed") {
		t.Fatal(err)
	}
	if comp.downs != 1 || len(comp.downOpts) != 1 || comp.downOpts[0].Volumes {
		t.Fatalf("down %+v count %d", comp.downOpts, comp.downs)
	}
	if !comp.upOpts.Start.Wait || comp.upOpts.Start.Attach != nil {
		t.Fatalf("up %+v", comp.upOpts.Start)
	}
}

func TestServiceDownJoinsTheTestError(t *testing.T) {
	comp := &fakeComposer{downErr: errors.New("down broke")}
	f := New("both", newClient(&scriptedEngine{}, comp))
	db := f.Service("db", sampleApp())
	f.Task("integration-test", func(context.Context) error { return errors.New("test failed") }).DependsOn(db)
	err := f.Run(context.Background(), discardRun())
	if err == nil || !strings.Contains(err.Error(), "test failed") || !strings.Contains(err.Error(), "down broke") {
		t.Fatal(err)
	}
}

func TestServiceDownNotFoundIsSuccess(t *testing.T) {
	comp := &fakeComposer{downErr: cerrdefs.ErrNotFound}
	f := New("gone", newClient(&scriptedEngine{}, comp))
	f.Service("db", sampleApp())
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
	if comp.downs != 1 {
		t.Fatal(comp.downs)
	}
}

func TestServiceDownOnCanceledRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	comp := &fakeComposer{}
	f := New("stop", newClient(&scriptedEngine{}, comp))
	db := f.Service("db", sampleApp())
	started := make(chan struct{})
	f.Task("block", func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}).DependsOn(db)
	go func() {
		<-started
		cancel()
	}()
	err := f.Run(ctx, discardRun())
	if err == nil || comp.downs != 1 {
		t.Fatalf("downs %d err %v", comp.downs, err)
	}
	if !comp.downLive || !comp.downDeadline {
		t.Fatalf("down live %v deadline %v", comp.downLive, comp.downDeadline)
	}
}

func TestSubflowServiceDownsBeforeParent(t *testing.T) {
	comp := &fakeComposer{}
	f := New("parent", newClient(&scriptedEngine{}, comp))
	sub := f.Subflow("inner", func(sf *Subflow) {
		db := sf.Service("db", sampleApp())
		sf.Task("use", func(context.Context) error {
			if comp.downs != 0 {
				return errors.New("down inside subflow")
			}
			return nil
		}).DependsOn(db)
	})
	f.Task("after", func(context.Context) error {
		if comp.downs != 1 {
			return errors.New("parent before down")
		}
		return nil
	}).DependsOn(sub)
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
}

func TestTwoServicesDownWithoutWaitingOnEachOther(t *testing.T) {
	comp := &fakeComposer{downErr: errors.New("down")}
	f := New("pair", newClient(&scriptedEngine{}, comp))
	f.Service("a", kit.NewCompose("a").WithService(kit.NewService(kit.Name("a"), kit.Image("alpine"))))
	f.Service("b", kit.NewCompose("b").WithService(kit.NewService(kit.Name("b"), kit.Image("alpine"))))
	f.Task("work", func(context.Context) error { return nil })
	err := f.Run(context.Background(), discardRun())
	if err == nil || comp.downs != 2 {
		t.Fatalf("downs %d err %v", comp.downs, err)
	}
}

func TestServiceRequiresAProjectAndAClient(t *testing.T) {
	f := New("nilapp", newClient(&scriptedEngine{}, &fakeComposer{}))
	f.Service("db", nil)
	err := f.Run(context.Background(), discardRun())
	if err == nil || !strings.Contains(err.Error(), "compose is nil") {
		t.Fatal(err)
	}
	f = New("nocli", nil)
	f.Service("db", sampleApp())
	err = f.Run(context.Background(), discardRun())
	if err == nil || !errors.Is(err, ErrNoClient) {
		t.Fatal(err)
	}
}

func TestImportService(t *testing.T) {
	if err := prepare(stepDoc{Op: "service", YAML: "services: ["}); err == nil {
		t.Fatal("bad yaml")
	}
	_, err := Import([]byte(`{"name":"app","steps":[{"name":"db","op":"service"}]}`), nil)
	if err == nil || !strings.Contains(err.Error(), "yaml required") {
		t.Fatal(err)
	}
	_, err = Import([]byte(`{"name":"app","steps":[{"name":"db","op":"service","yaml":"name: app\nservices:\n  db:\n    image: postgres:16\n","options":{"detach":"no"}}]}`), nil)
	if err == nil {
		t.Fatal("bad options")
	}
	comp := &fakeComposer{}
	doc := `{"name":"app","steps":[{"name":"db","op":"service","yaml":"name: app\nservices:\n  db:\n    image: postgres:16\n","options":{"timeout":"1s","wait":false}}]}`
	f, err := Import([]byte(doc), newClient(&scriptedEngine{}, comp))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(context.Background(), discardRun()); err != nil {
		t.Fatal(err)
	}
	if _, err := makeStep(New("badopt", nil).scope(), stepDoc{Name: "db", Op: "service", YAML: appYAML, Options: map[string]any{"detach": "no"}}); err == nil {
		t.Fatal("makeStep options")
	}
	if comp.ups != 1 || comp.downs != 1 || !comp.upOpts.Start.Wait {
		t.Fatalf("up %d down %d wait %v", comp.ups, comp.downs, comp.upOpts.Start.Wait)
	}
}
