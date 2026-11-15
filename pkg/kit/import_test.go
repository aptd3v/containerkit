package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	composetypes "github.com/aptd3v/containerkit/internal/compose/types"
)

func TestImportYAMLBasic(t *testing.T) {
	t.Parallel()
	yml := []byte(`
name: demo
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      FOO: bar
`)
	p := ImportYAML(yml)
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Name() != "demo" {
		t.Fatalf("name = %q", p.Name())
	}
	svcs := p.Services()
	if len(svcs) != 1 || svcs[0].Name() != "web" || svcs[0].Image() != "nginx:latest" {
		t.Fatalf("services = %+v", svcs)
	}
	cfg, err := svcs[0].ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Ports) != 1 || cfg.Ports[0].Published != "8080" || cfg.Ports[0].Target != 80 {
		t.Fatalf("ports = %+v", cfg.Ports)
	}
	if cfg.Environment["FOO"] == nil || *cfg.Environment["FOO"] != "bar" {
		t.Fatalf("env = %+v", cfg.Environment)
	}
}

func TestImportYAMLRoundTrip(t *testing.T) {
	t.Parallel()
	orig := NewCompose(
		"round",
		AddNetwork(Name("front"), Driver("bridge")),
		AddVolume(Name("data"), Driver("local")),
		AddService(
			Name("api"),
			Image("alpine:latest"),
			Command("sleep", "infinity"),
			Publish(HostPort(8080), ContainerPort(80)),
			Env("APP", "1"),
		),
	)
	yml, err := orig.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	loaded := ImportYAML(yml)
	if err := loaded.Validate(); err != nil {
		t.Fatal(err)
	}
	out, err := loaded.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{
		"name: round",
		"api:",
		"alpine:latest",
		"front:",
		"data:",
		"APP:",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("round-trip missing %q in:\n%s", want, text)
		}
	}
}

func TestImportYAMLInterpolation(t *testing.T) {
	t.Parallel()
	yml := []byte(`
name: interp
services:
  api:
    image: ${IMG}
`)
	p := ImportYAML(yml, Env("IMG", "alpine:3.20"))
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Services()[0].Image() != "alpine:3.20" {
		t.Fatalf("image = %q", p.Services()[0].Image())
	}
}

func TestImportFileAndNameOverride(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(path, []byte(`
name: fromfile
services:
  db:
    image: postgres:16
`), 0o600); err != nil {
		t.Fatal(err)
	}
	p := ImportFile(path, Name("renamed"), ProjectDirectory(dir))
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Name() != "renamed" {
		t.Fatalf("name = %q", p.Name())
	}
	if len(p.Services()) != 1 || p.Services()[0].Name() != "db" {
		t.Fatalf("services = %+v", p.Services())
	}
}

func TestImportYAMLEmpty(t *testing.T) {
	t.Parallel()
	p := ImportYAML(nil)
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestImportRejectsPortAndExposeOverflow(t *testing.T) {
	t.Parallel()
	_, err := portFromConfig(composetypes.ServicePortConfig{Target: 70000, Published: "8080"})
	if err == nil {
		t.Fatal("expected container port overflow")
	}
	p, err := portFromConfig(composetypes.ServicePortConfig{Target: 80, Published: "8080"})
	if err != nil || !p.containerPort.ok() || p.containerPort.val != 80 {
		t.Fatalf("port = %+v err %v", p, err)
	}
	yml := []byte(`
name: demo
services:
  web:
    image: nginx:latest
    expose:
      - "99999"
`)
	loaded := ImportYAML(yml)
	if err := loaded.Validate(); err == nil {
		t.Fatal("expected expose overflow")
	}
}

func TestImportInvalidYAML(t *testing.T) {
	t.Parallel()
	p := ImportYAML([]byte(":::not yaml"))
	if err := p.Validate(); err == nil {
		t.Fatal("expected error")
	}
}
