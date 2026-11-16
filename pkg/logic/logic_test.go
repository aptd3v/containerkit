package logic_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/logic"
)

func TestGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		opts      []any
		wantName  string
		wantImage string
		wantErr   bool
	}{
		{
			name:      "name and image",
			opts:      []any{kit.Name("web"), kit.Image("alpine:latest")},
			wantName:  "web",
			wantImage: "alpine:latest",
		},
		{
			name:      "skips nil",
			opts:      []any{nil, kit.Name("ok"), nil, kit.Image("alpine:latest")},
			wantName:  "ok",
			wantImage: "alpine:latest",
		},
		{
			name:    "rejects option that does not apply",
			opts:    []any{kit.Name("web"), kit.Deploy(kit.Replicas(1))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := kit.NewContainer(logic.Group(tt.opts...))
			err := c.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if c.Name() != tt.wantName || c.Image() != tt.wantImage {
				t.Fatalf("got %s %s", c.Name(), c.Image())
			}
		})
	}
}

func TestWhen(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opt      kit.ContainerOption
		wantName string
	}{
		{
			name:     "when true",
			opt:      logic.When(true, kit.Name("on")),
			wantName: "on",
		},
		{
			name:     "when false skips",
			opt:      logic.When(false, kit.Name("off")),
			wantName: "",
		},
		{
			name:     "whenfn true",
			opt:      logic.WhenFn(func() bool { return true }, kit.Name("fn")),
			wantName: "fn",
		},
		{
			name:     "unless false applies",
			opt:      logic.Unless(false, kit.Name("unless")),
			wantName: "unless",
		},
		{
			name:     "unless true skips",
			opt:      logic.Unless(true, kit.Name("no")),
			wantName: "",
		},
		{
			name:     "whenelse true",
			opt:      logic.WhenElse(true, kit.Name("then"), kit.Name("else")),
			wantName: "then",
		},
		{
			name:     "whenelse false",
			opt:      logic.WhenElse(false, kit.Name("then"), kit.Name("else")),
			wantName: "else",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := kit.NewContainer(kit.Image("alpine:latest"), tt.opt)
			if err := c.Validate(); err != nil {
				t.Fatal(err)
			}
			if c.Name() != tt.wantName {
				t.Fatalf("name = %q", c.Name())
			}
		})
	}
}

func TestOnlyIf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		check    logic.Check
		wantName string
		wantErr  error
	}{
		{
			name:     "check true",
			check:    func() (bool, error) { return true, nil },
			wantName: "only",
		},
		{
			name:     "check false skips",
			check:    func() (bool, error) { return false, nil },
			wantName: "",
		},
		{
			name:    "check error",
			check:   func() (bool, error) { return false, errors.New("missing file") },
			wantErr: errors.New("missing file"),
		},
		{
			name:    "nil check",
			check:   nil,
			wantErr: errors.New("logic.onlyif: check is nil"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := kit.NewContainer(
				kit.Image("alpine:latest"),
				logic.OnlyIf(tt.check, kit.Name("only")),
			)
			err := c.Validate()
			if tt.wantErr != nil {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if c.Name() != tt.wantName {
				t.Fatalf("name = %q", c.Name())
			}
		})
	}
}

func TestEachMap(t *testing.T) {
	t.Parallel()
	c := kit.NewContainer(
		kit.Image("alpine:latest"),
		logic.Each([]string{"A", "B"}, func(i int, k string) any {
			return kit.Env(k, "1")
		}),
		logic.Map([]string{"C"}, func(k string) any {
			return kit.Env(k, "2")
		}),
		logic.EachKV(map[string]string{"D": "3", "E": "4"}, func(k, v string) any {
			return kit.Env(k, v)
		}),
	)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	cfg, _, _, err := c.EngineConfig()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"A=1", "B=1", "C=2", "D=3", "E=4"} {
		if !slices.Contains(cfg.Env, want) {
			t.Fatalf("env %v missing %s", cfg.Env, want)
		}
	}
}

func TestSwitch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		opt      kit.ContainerOption
		wantName string
	}{
		{
			name: "first matching case",
			opt: logic.Switch(
				logic.Case(false, kit.Name("a")),
				logic.Case(true, kit.Name("b")),
				logic.Default(kit.Name("c")),
			),
			wantName: "b",
		},
		{
			name: "default",
			opt: logic.Switch(
				logic.Case(false, kit.Name("a")),
				logic.Default(kit.Name("c")),
			),
			wantName: "c",
		},
		{
			name: "casefn",
			opt: logic.Switch(
				logic.CaseFn(func() bool { return false }, kit.Name("a")),
				logic.CaseFn(func() bool { return true }, kit.Name("d")),
			),
			wantName: "d",
		},
		{
			name: "on equality",
			opt: logic.Switch(
				logic.On("nginx", "caddy", kit.Name("caddy")),
				logic.On("nginx", "nginx", kit.Name("nginx")),
				logic.Default(kit.Name("envoy")),
			),
			wantName: "nginx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := kit.NewContainer(kit.Image("alpine:latest"), tt.opt)
			if err := c.Validate(); err != nil {
				t.Fatal(err)
			}
			if c.Name() != tt.wantName {
				t.Fatalf("name = %q", c.Name())
			}
		})
	}
}

func TestNestedGroupWhen(t *testing.T) {
	t.Parallel()
	prod := true
	c := kit.NewContainer(
		logic.Group(
			kit.Name("api"),
			kit.Image("alpine:latest"),
			logic.When(prod, kit.Env("CACHE", "1")),
			logic.Unless(prod, kit.Env("CACHE", "0")),
			kit.Health(logic.When(true, kit.Test("CMD-SHELL", "true"))),
		),
	)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	cfg, _, _, err := c.EngineConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(cfg.Env, "CACHE=1") {
		t.Fatalf("env = %v", cfg.Env)
	}
	if cfg.Healthcheck == nil || len(cfg.Healthcheck.Test) == 0 {
		t.Fatal("missing healthcheck")
	}
}
