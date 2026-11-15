package kit

import (
	"bytes"
	"context"
	"testing"
)

func TestClientNilLoad(t *testing.T) {
	t.Parallel()
	if _, err := (*Client)(nil).Load(context.Background(), "compose.yaml"); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientLoadEmptyPath(t *testing.T) {
	t.Parallel()
	c := NewClient(&fakeEngine{}, &fakeComposer{})
	if _, err := c.Load(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestClientLoadAndUp(t *testing.T) {
	t.Parallel()
	api := &fakeComposer{}
	c := NewClient(&fakeEngine{}, api)
	loaded, err := c.Load(context.Background(), "compose.yaml", Name("app"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Project == nil || loaded.Project.Name != "app" {
		t.Fatalf("project = %+v", loaded.Project)
	}
	if api.loaded.ConfigPaths[0] != "compose.yaml" {
		t.Fatalf("loaded files = %#v", api.loaded.ConfigPaths)
	}
	if err := loaded.Up(context.Background(), Detach(true)); err != nil {
		t.Fatal(err)
	}
	if api.up == nil || api.up.Name != "app" {
		t.Fatalf("up = %+v", api.up)
	}
}

func TestClientUpCompose(t *testing.T) {
	t.Parallel()
	api := &fakeComposer{}
	c := NewClient(&fakeEngine{}, api)
	p := NewCompose("app").WithService(NewService(Name("api"), Image("alpine:latest")))
	if err := c.Up(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if api.up == nil || api.up.Name != "app" {
		t.Fatalf("up = %+v", api.up)
	}
}

func TestClientRun(t *testing.T) {
	t.Parallel()
	eng := &fakeEngine{}
	c := NewClient(eng, &fakeComposer{})
	if err := c.Run(context.Background(), Name("web"), Image("nginx")); err != nil {
		t.Fatal(err)
	}
	if eng.started != "abc" {
		t.Fatalf("started = %q", eng.started)
	}
}

func TestClientCloseNil(t *testing.T) {
	t.Parallel()
	if err := (*Client)(nil).Close(); err != nil {
		t.Fatal(err)
	}
	c := NewClient(&fakeEngine{}, &fakeComposer{})
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWriteValidates(t *testing.T) {
	t.Parallel()
	p := NewCompose("app").WithService(NewService())
	if err := p.Write(Export{Writer: &bytes.Buffer{}}); err == nil {
		t.Fatal("expected validate error")
	}
	if _, err := p.ExportYAML(); err == nil {
		t.Fatal("expected validate error")
	}
}

func TestFromFileEmpty(t *testing.T) {
	t.Parallel()
	p := NewCompose("app").WithSecret(Name("tok"), FromFile(""))
	if err := p.Validate(); err == nil {
		t.Fatal("expected empty from-file")
	}
}

func TestMainPackageConstants(t *testing.T) {
	t.Parallel()
	c := NewContainer(
		Name("web"),
		Image("nginx"),
		NetworkMode(Bridge),
		IPC(IPCShareable),
		StopSignal(SIGTERM),
		Restart(Policy(Always)),
		Publish(ContainerPort(80), Protocol(TCP), Mode(PortHost)),
		Logging(Driver(JSONFile)),
	)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	s := NewService(c, Deploy(Mode(Replicated), EndpointMode(EndpointVIP)))
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}
