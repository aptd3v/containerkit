package containerkit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerFileBuilder(t *testing.T) {
	df := NewDockerFile().
		From("alpine", "latest").
		FromAs("golang:1.22", "build").
		Arg("VERSION").
		ArgKV("APP", "web").
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
		Healthcheck(Health{Test: []string{"CMD", "true"}, Timeout: 1, Interval: 1, Retries: 1}).
		Run("apk add --no-cache curl").
		Run("echo still-run").
		RunArgs("--no-cache", "ok").
		CommandExec("sleep", "infinity")

	require.NoError(t, df.Validate())
	s := df.String()
	require.Contains(t, s, "FROM alpine:latest")
	require.Contains(t, s, "FROM golang:1.22 AS build")
	require.Contains(t, s, `COPY ["file with space.txt", "/dest with space"]`)
	require.Contains(t, s, `ADD ["src file", "dest dir"]`)
	require.Contains(t, s, "RUN apk add --no-cache curl && \\")
	require.Contains(t, s, "CMD [\"sleep\", \"infinity\"]")
	require.Contains(t, df.WithInline().DockerfileInline, "FROM alpine:latest")

	path := filepath.Join(t.TempDir(), "Dockerfile")
	require.NoError(t, df.Export(path, 0644))

	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("x"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(src, "sub"), 0755))
	r, err := df.NewLocalBuildContext(src)
	require.NoError(t, err)
	buf := make([]byte, 16)
	n, err := r.Read(buf)
	require.NoError(t, err)
	require.Greater(t, n, 0)
}

func TestDockerFileErrors(t *testing.T) {
	df := NewDockerFile().From("alpine", "latest").CommandExec("true").CommandExec("false")
	require.Error(t, df.Validate())
	require.Error(t, df.Export(filepath.Join(t.TempDir(), "D"), 0644))

	df2 := NewDockerFile().From("alpine", "latest").CommandShell("echo", "hi").CommandShell("echo", "again")
	require.Error(t, df2.Validate())

	df3 := NewDockerFile().RunArgs("x")
	require.Error(t, df3.Validate())
	require.Contains(t, df3.Validate().Error(), "runargs")

	df4 := NewDockerFile().From("alpine", "latest")
	_, err := df4.NewLocalBuildContext(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)

	bad := NewDockerFile()
	bad.errs = append(bad.errs, errString("boom"))
	_, err = bad.NewLocalBuildContext(t.TempDir())
	require.Error(t, err)

	df5 := NewDockerFile().From("%s", "%s").Format("alpine", "latest")
	require.Contains(t, df5.String(), "FROM alpine:latest")

	df6 := NewDockerFile().From("alpine", "latest").CommandShell("echo", "ok")
	require.NoError(t, df6.Validate())
	require.True(t, strings.HasPrefix(df6.String(), "FROM alpine:latest"))
}

type errString string

func (e errString) Error() string { return string(e) }
