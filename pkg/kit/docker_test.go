package kit

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDockerfileBuilder(t *testing.T) {
	t.Parallel()
	df := NewDockerfile().
		From("alpine", "latest").
		FromAs("golang:1.22", "build").
		Arg("VERSION").
		Arg("APP", "web").
		Env("FOO", "bar").
		Copy("a.txt", "b.txt").
		Copy("file with space.txt", "/dest with space").
		Add("src", "dest").
		Add("src file", "dest dir").
		Entrypoint("/bin/sh", "-c").
		Expose("80").
		Label("k", "v").
		Onbuild("RUN echo hi").
		Workdir("/app").
		StopSignal("SIGTERM").
		User("root").
		Comment("note").
		Volumes("/data", "/logs").
		HealthCheck(Test("CMD", "true"), Timeout(time.Second), Interval(time.Second), Retries(1)).
		Run("apk add --no-cache curl").
		Run("echo still-run").
		RunArgs("--no-cache", "ok").
		Command("sleep", "infinity")

	if err := df.Validate(); err != nil {
		t.Fatal(err)
	}
	s := df.String()
	for _, want := range []string{
		"FROM alpine:latest",
		"FROM golang:1.22 AS build",
		`COPY ["file with space.txt", "/dest with space"]`,
		`ADD ["src file", "dest dir"]`,
		"RUN apk add --no-cache curl && \\",
		`CMD ["sleep", "infinity"]`,
		`ENTRYPOINT ["/bin/sh", "-c"]`,
		`VOLUME ["/data", "/logs"]`,
		"--interval=1s",
		"--timeout=1s",
		"--retries=1",
		`CMD ["true"]`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
	if strings.Contains(s, "--start-period=") {
		t.Fatalf("unexpected start-period in %s", s)
	}

	svc := NewService(Name("api"), Build(df))
	p := NewCompose("app").WithService(svc)
	yml, err := p.ExportYAML()
	if err != nil {
		t.Fatal(err)
	}
	text := string(yml)
	if !strings.Contains(text, "dockerfile_inline:") || !strings.Contains(text, "FROM alpine:latest") {
		t.Fatalf("export = %s", text)
	}

	path := filepath.Join(t.TempDir(), "Dockerfile")
	if err := df.Export(path, 0644); err != nil {
		t.Fatal(err)
	}
	got := ReadDockerfile(path)
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.String(), "FROM alpine:latest") {
		t.Fatalf("ReadDockerfile = %s", got.String())
	}
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	r, err := df.Archive(src)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(r)
	names := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		names[hdr.Name] = true
	}
	if !names["Dockerfile"] || !names["a.txt"] {
		t.Fatalf("archive names = %v", names)
	}
}

func TestDockerfileInstructions(t *testing.T) {
	t.Parallel()
	df := NewDockerFile().
		Syntax("docker/dockerfile:1").
		Escape(`\`).
		Check("error=true").
		From("scratch", "").
		FromPlatform("linux/amd64", "alpine", "3.20").
		FromAsPlatform("linux/arm64", "golang:1.22", "build").
		Copy("out", "/bin/app", CopyOpts{
			From:    "build",
			Chown:   "nobody:nobody",
			Chmod:   "755",
			Parents: true,
			Link:    true,
			Exclude: []string{"*.md"},
		}).
		Add("repo", "/src", AddOpts{Checksum: "sha256:abc", KeepGitDir: true, Link: true}).
		Shell("/bin/bash", "-c").
		EntrypointShell("nginx", "-g", "daemon off;").
		HealthCheckNone().
		Run("echo a").
		RunLayer("echo b").
		Run("echo c", RunOpts{Mount: []string{"type=cache,target=/cache"}, Network: "none", Security: "sandbox"}).
		Command("true")

	if err := df.Validate(); err != nil {
		t.Fatal(err)
	}
	s := df.String()
	if !strings.HasPrefix(s, "# syntax=docker/dockerfile:1\n# escape=\\\n# check=error=true\n") {
		t.Fatalf("prefix = %q", s[:min(80, len(s))])
	}
	for _, want := range []string{
		"FROM scratch\n",
		"FROM --platform=linux/amd64 alpine:3.20\n",
		"FROM --platform=linux/arm64 golang:1.22 AS build\n",
		"COPY --from=build --chown=nobody:nobody --chmod=755 --exclude=*.md --parents --link out /bin/app\n",
		"ADD --checksum=sha256:abc --keep-git-dir=true --link repo /src\n",
		`SHELL ["/bin/bash", "-c"]`,
		"ENTRYPOINT nginx -g daemon off;\n",
		"HEALTHCHECK NONE\n",
		"RUN echo a\nRUN echo b\n",
		"RUN --mount=type=cache,target=/cache --network=none --security=sandbox echo c\n",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
	if strings.Contains(s, "FROM scratch:") {
		t.Fatal("scratch should not have empty tag")
	}
	if strings.Contains(s, "echo a &&") {
		t.Fatal("RunLayer should not fold")
	}
}

func TestDockerfileErrorsAndFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		df   *Dockerfile
	}{
		{"second command", NewDockerfile().From("alpine", "latest").Command("true").Command("false")},
		{"runargs without run", NewDockerfile().RunArgs("x")},
		{"empty parse", ParseDockerfile("")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.df.Validate(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	df := NewDockerfile().From("%s", "%s").Format("alpine", "latest")
	if err := df.Validate(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(df.String(), "FROM alpine:latest") {
		t.Fatalf("format = %s", df.String())
	}
}

func TestBuildFileAndDockerfileConflict(t *testing.T) {
	t.Parallel()
	s := NewService(Name("api"), Build(DockerfilePath("Dockerfile"), NewDockerfile().From("alpine:latest")))
	if err := s.Validate(); err == nil {
		t.Fatal("expected File and Dockerfile conflict")
	}
}

func TestDockerfileHealthcheckShell(t *testing.T) {
	t.Parallel()
	df := NewDockerfile().HealthCheck(Test("CMD-SHELL", "curl -f http://localhost || exit 1"), Interval(5*time.Second))
	if err := df.Validate(); err != nil {
		t.Fatal(err)
	}
	want := "HEALTHCHECK --interval=5s \\\n\tCMD curl -f http://localhost || exit 1\n"
	if df.String() != want {
		t.Fatalf("got %q want %q", df.String(), want)
	}
}

func TestBuildImageUsesDockerfile(t *testing.T) {
	t.Parallel()
	api := &fakeEngine{}
	df := NewDockerfile().From("alpine:latest").Run("echo ok").Command("true")
	if _, err := BuildImage(context.Background(), api, df, Image("app:dev")); err != nil {
		t.Fatal(err)
	}
	if len(api.built.Tags) != 1 || api.built.Tags[0] != "app:dev" || api.built.Dockerfile != "Dockerfile" {
		t.Fatalf("built = %+v", api.built)
	}
}
