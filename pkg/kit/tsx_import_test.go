package kit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestTSXRenderedYAMLImport(t *testing.T) {
	t.Parallel()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	path := filepath.Join(filepath.Dir(file), "..", "..", "tsx", "testdata", "example.yaml")
	yml, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p := ImportYAML(yml)
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Name() != "app" {
		t.Fatalf("name = %q", p.Name())
	}
	svcs := p.Services()
	if len(svcs) != 1 || svcs[0].Name() != "api" || svcs[0].Image() != "alpine:latest" {
		t.Fatalf("services = %+v", svcs)
	}
	cfg, err := svcs[0].ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Ports) != 1 || cfg.Ports[0].Published != "8080" || cfg.Ports[0].Target != 80 {
		t.Fatalf("ports = %+v", cfg.Ports)
	}
	if cfg.Environment["APP"] == nil || *cfg.Environment["APP"] != "1" {
		t.Fatalf("env = %+v", cfg.Environment)
	}
	if cfg.HealthCheck == nil || cfg.HealthCheck.Retries == nil || *cfg.HealthCheck.Retries != 3 {
		t.Fatalf("health = %+v", cfg.HealthCheck)
	}
	if cfg.Build == nil || cfg.Build.Dockerfile != "Dockerfile" {
		t.Fatalf("build = %+v", cfg.Build)
	}
	out, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, want := range []string{"name: app", "api:", "alpine:latest", "front:", "data:", "APP:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
}
