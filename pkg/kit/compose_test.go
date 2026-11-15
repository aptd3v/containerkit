package kit

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestComposeFluentWithServiceChain(t *testing.T) {
	t.Parallel()
	myService := NewService(NewContainer(Name("foo-server"), Image("alpine:latest")))
	p := NewCompose("app").
		WithService(myService).
		WithService(NewService(Image("nginx:latest"))).
		WithNetwork(Name("front"), Driver("bridge")).
		WithVolume(Name("dbdata"), Driver("local")).
		WithSecret(Name("dbpass"), FromFile("./dbpass.txt")).
		WithFile(Name("nginx"), FromFile("./nginx.conf"))
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Name() != "app" {
		t.Fatalf("name = %q", p.Name())
	}
	svcs := p.Services()
	if len(svcs) != 2 || svcs[0].Name() != "foo-server" || svcs[1].Name() != "nginx" {
		t.Fatalf("services = %+v", svcs)
	}
	yml, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(yml)
	for _, want := range []string{
		"name: app",
		"foo-server:",
		"nginx:",
		"front:",
		"dbdata:",
		"dbpass:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
}

func TestDefaultNetworkUsesProjectEngineName(t *testing.T) {
	t.Parallel()
	p := NewCompose("supabase").WithNetwork(Name("default"), Driver("bridge"))
	proj, err := p.Project()
	if err != nil {
		t.Fatal(err)
	}
	if proj.Networks["default"].Name != "supabase_default" {
		t.Fatalf("engine network = %q", proj.Networks["default"].Name)
	}
	front := NewCompose("app").WithNetwork(Name("front"))
	named, err := front.Project()
	if err != nil {
		t.Fatal(err)
	}
	if named.Networks["front"].Name != "front" {
		t.Fatalf("engine network = %q", named.Networks["front"].Name)
	}
}

func TestComposeUniquifiesGeneratedServiceNames(t *testing.T) {
	t.Parallel()
	p := NewCompose("app").
		WithService(NewService(Image("nginx:latest"))).
		WithService(NewService(Image("nginx:1.27")))
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	svcs := p.Services()
	if svcs[0].Name() != "nginx" || svcs[1].Name() != "nginx-2" {
		t.Fatalf("names = %q %q", svcs[0].Name(), svcs[1].Name())
	}
}

func TestComposeDuplicateExplicitServiceName(t *testing.T) {
	t.Parallel()
	p := NewCompose("app").
		WithService(NewService(Name("api"), Image("alpine:latest"))).
		WithService(NewService(Name("api"), Image("nginx:latest")))
	err := p.Validate()
	if err == nil {
		t.Fatal("expected duplicate name")
	}
	if !errors.Is(err, errInvalid) {
		t.Fatalf("Is invalid: %v", err)
	}
}

func TestComposeUpRuntime(t *testing.T) {
	t.Parallel()
	p := NewCompose("app").
		WithService(NewService(Name("api"), Image("alpine:latest")))
	api := &fakeComposer{}
	if err := p.Up(context.Background(), api); err != nil {
		t.Fatal(err)
	}
	if api.up == nil || api.up.Name != "app" {
		t.Fatalf("up compose = %+v", api.up)
	}
}

func TestComposeValidateBeforeUp(t *testing.T) {
	t.Parallel()
	p := NewCompose("").WithService(NewService(Image("alpine:latest")))
	err := p.Up(context.Background(), &fakeComposer{})
	if err == nil {
		t.Fatal("expected missing compose name")
	}
	if !errors.Is(err, errRequired) {
		t.Fatalf("Is required: %v", err)
	}
}
