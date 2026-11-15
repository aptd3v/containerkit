package kit

import (
	"errors"
	"testing"
)

func TestServiceInheritsContainerName(t *testing.T) {
	t.Parallel()
	s := NewService(NewContainer(Name("foo-server"), Image("alpine:latest")))
	if s.Name() != "foo-server" {
		t.Fatalf("inherited name = %q", s.Name())
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestServiceGeneratesNameFromImage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{name: "tag", image: "nginx:latest", want: "nginx"},
		{name: "registry path", image: "ghcr.io/org/my.app:1.0", want: "my-app"},
		{name: "digest", image: "redis@sha256:abc", want: "redis"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewService(NewContainer(Image(tt.image)))
			if s.Name() != tt.want {
				t.Fatalf("generated name = %q, want %q", s.Name(), tt.want)
			}
		})
	}
}

func TestServiceGeneratesNameFromBuildTarget(t *testing.T) {
	t.Parallel()
	s := NewService(Build(Source("."), Target("prod")))
	if s.Name() != "prod" {
		t.Fatalf("generated name = %q", s.Name())
	}
}

func TestServiceExplicitNameWinsOverContainer(t *testing.T) {
	t.Parallel()
	s := NewService(
		NewContainer(Name("foo-server"), Image("alpine:latest")),
		Name("api"),
	)
	if s.Name() != "api" {
		t.Fatalf("name = %q", s.Name())
	}
}

func TestServiceAsServicePassesThrough(t *testing.T) {
	t.Parallel()
	c := NewContainer(Name("foo-server"), Image("alpine:latest"), Env("DEBUG", "1"))
	s := c.AsService()
	if s.Name() != "foo-server" || s.Image() != "alpine:latest" {
		t.Fatalf("as service = %s/%s", s.Name(), s.Image())
	}
}

func TestServiceValidateMissingImage(t *testing.T) {
	t.Parallel()
	s := NewService()
	err := s.Validate()
	if err == nil {
		t.Fatal("expected missing image")
	}
	if !errors.Is(err, errRequired) {
		t.Fatalf("Is required: %v", err)
	}
}
