package kit

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStandaloneContainerEngineConfig(t *testing.T) {
	t.Parallel()
	c := NewContainer(
		Name("worker"),
		Image("alpine:latest"),
		Command("sleep", "infinity"),
		Env("TOKEN", "s3cret"),
		Privileged(false),
		Memory(0),
		Publish(
			Protocol(TCP),
			HostIP("0.0.0.0"),
			HostPort(8080),
			ContainerPort(80),
		),
		Health(
			Test("CMD", "true"),
			Interval(10*time.Second),
			Timeout(2*time.Second),
			Retries(3),
		),
		CapAdd(ALL),
	)
	cfg, host, netcfg, err := c.EngineConfig()
	if err != nil {
		t.Fatal(err)
	}
	if netcfg == nil {
		t.Fatal("expected networking config")
	}
	if cfg.Image != "alpine:latest" || cfg.Hostname != "" {
		t.Fatalf("image/hostname = %q %q", cfg.Image, cfg.Hostname)
	}
	if host.Privileged {
		t.Fatal("explicit false privileged should stay false")
	}
	if host.Memory != 0 {
		t.Fatalf("explicit zero memory = %d", host.Memory)
	}
	if len(host.PortBindings) != 1 {
		t.Fatalf("port bindings = %v", host.PortBindings)
	}
	if cfg.Healthcheck == nil || cfg.Healthcheck.Retries != 3 {
		t.Fatalf("health = %+v", cfg.Healthcheck)
	}
	if len(host.CapAdd) != 1 || host.CapAdd[0] != string(ALL) {
		t.Fatalf("capadd = %v", host.CapAdd)
	}
	name, err := c.CreateName()
	if err != nil {
		t.Fatal(err)
	}
	if name != "worker" {
		t.Fatalf("create name = %q", name)
	}
}

func TestComposeServiceAndExport(t *testing.T) {
	t.Parallel()
	p := NewCompose(
		"app",
		AddService(
			Name("api"),
			Image("alpine:latest"),
			Publish(
				Protocol(TCP),
				HostPort(8080),
				ContainerPort(80),
			),
			DependsOn(
				Name("db"),
				Condition(Healthy),
			),
		),
		AddService(
			Name("db"),
			Image("postgres:16"),
		),
	)
	yml, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(yml)
	for _, want := range []string{"name: app", "api:", "db:", "image: alpine:latest", "target: 80", "published: \"8080\"", "condition: service_healthy"} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
	var buf bytes.Buffer
	if err := p.Write(Export{Writer: &buf}); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("write produced no yaml")
	}
}

func TestReuseOptionsAcrossContainerAndService(t *testing.T) {
	t.Parallel()
	c := NewContainer(
		Name("worker"),
		Image("alpine:latest"),
		Env("DEBUG", "1"),
		Publish(ContainerPort(80), HostPort(8080), Protocol(TCP)),
	)
	s := NewService(
		Name("worker"),
		Image("alpine:latest"),
		Env("DEBUG", "1"),
		Publish(ContainerPort(80), HostPort(8080), Protocol(TCP)),
	)
	from := NewService(FromContainer(c), Name("worker"))
	if s.Image() != from.Image() || s.Name() != from.Name() {
		t.Fatalf("service %s/%s vs from %s/%s", s.Name(), s.Image(), from.Name(), from.Image())
	}
	cfg, err := from.ComposeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Ports) != 1 || cfg.Ports[0].Target != 80 {
		t.Fatalf("ports = %+v", cfg.Ports)
	}
}

func TestIncompatibleAutoRemoveMapping(t *testing.T) {
	t.Parallel()
	c := NewContainer(Name("w"), Image("alpine:latest"), AutoRemove(true))
	s := NewService(FromContainer(c), Name("w"))
	if s == nil {
		t.Fatal("constructor must return a service")
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("expected unsupported mapping")
	}
	if !errors.Is(err, errUnsupported) {
		t.Fatalf("Is unsupported: %v", err)
	}
	var ce *Error
	if !errors.As(err, &ce) {
		t.Fatalf("As Error: %v", err)
	}
	if ce.Setting != "AutoRemove" || ce.Receiver != "service" {
		t.Fatalf("error identity = %+v (%s)", ce, err)
	}
}

func TestNestedErrorPropagation(t *testing.T) {
	t.Parallel()
	p := NewCompose(
		"app",
		AddService(
			Name("api"),
			Image("alpine:latest"),
			Publish(
				Protocol(TCP),
				HostPort(8080),
			),
		),
	)
	if p == nil {
		t.Fatal("constructor must return a value")
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("expected missing container port")
	}
	var ce *Error
	if !errors.As(err, &ce) {
		t.Fatalf("As Error: %v", err)
	}
	if !strings.Contains(err.Error(), "publish") || !strings.Contains(err.Error(), "ContainerPort") {
		t.Fatalf("path missing: %v", err)
	}
	if !errors.Is(err, errRequired) {
		t.Fatalf("Is required: %v", err)
	}
}

func TestEnvErrorRedactsSecret(t *testing.T) {
	t.Parallel()
	c := NewContainer(Image("alpine:latest"), Env("", "super-secret-value"))
	err := c.Validate()
	if err == nil {
		t.Fatal("expected env error")
	}
	if strings.Contains(err.Error(), "super-secret-value") {
		t.Fatalf("secret leaked: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid secret value") {
		t.Fatalf("expected redacted secret error: %v", err)
	}
}

func TestCrossFieldRestartOnFailure(t *testing.T) {
	t.Parallel()
	c := NewContainer(
		Image("alpine:latest"),
		Restart(Policy(OnFailure), MaxRetries(5)),
	)
	_, host, _, err := c.EngineConfig()
	if err != nil {
		t.Fatal(err)
	}
	if string(host.RestartPolicy.Name) != string(OnFailure) || host.RestartPolicy.MaximumRetryCount != 5 {
		t.Fatalf("restart = %+v", host.RestartPolicy)
	}
}

func TestConstructorAlwaysReturnsValue(t *testing.T) {
	t.Parallel()
	c := NewContainer(Name("x"))
	if c == nil {
		t.Fatal("NewContainer must return a container")
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected missing image")
	}
	p := NewCompose("")
	if p == nil {
		t.Fatal("NewCompose must return a value")
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected missing name")
	}
}
