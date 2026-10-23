package containerkit

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerFileBuilder(t *testing.T) {
	df := NewDockerFile().From("alpine", "latest").FromAs("golang:1.22", "build").Arg("VERSION").ArgKV("APP", "web").Env("FOO", "bar").Copy("a.txt", "b.txt").Copy("file with space.txt", "/dest with space").Add("src", "dest").Add("src file", "dest dir").Entrypoint("/bin/sh", "-c").Expose("80").Label("k", "v").Onbuild("RUN echo hi").Workdir("/app").StopSignal("SIGTERM").User("root").Comment("note").Volumes("/data", "/logs").Healthcheck(Health{Test: []string{"CMD", "true"}, Timeout: 1, Interval: 1, Retries: 1}).Run("apk add --no-cache curl").Run("echo still-run").RunArgs("--no-cache", "ok").CommandExec("sleep", "infinity")

	require.NoError(t, df.Validate())
	s := df.String()
	require.Contains(t, s, "FROM alpine:latest")
	require.Contains(t, s, "FROM golang:1.22 AS build")
	require.Contains(t, s, `COPY ["file with space.txt", "/dest with space"]`)
	require.Contains(t, s, `ADD ["src file", "dest dir"]`)
	require.Contains(t, s, "RUN apk add --no-cache curl && \\")
	require.Contains(t, s, "CMD [\"sleep\", \"infinity\"]")
	require.Contains(t, s, `ENTRYPOINT ["/bin/sh", "-c"]`)
	require.Contains(t, s, `VOLUME ["/data", "/logs"]`)
	require.Contains(t, s, "--interval=1s")
	require.Contains(t, s, "--timeout=1s")
	require.Contains(t, s, "--retries=1")
	require.NotContains(t, s, "--start-period=")
	require.Contains(t, s, `CMD ["true"]`)
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

func TestDockerFileInstructions(t *testing.T) {
	df := NewDockerFile().Syntax("docker/dockerfile:1").Escape(`\`).Check("error=true").From("scratch", "").FromPlatform("linux/amd64", "alpine", "3.20").FromAsPlatform("linux/arm64", "golang:1.22", "build").Copy("out", "/bin/app", CopyOpts{
		From:    "build",
		Chown:   "nobody:nobody",
		Chmod:   "755",
		Parents: true,
		Link:    true,
		Exclude: []string{"*.md"},
	}).Add("repo", "/src", AddOpts{
		Checksum:   "sha256:abc",
		KeepGitDir: true,
		Link:       true,
	}).Shell("/bin/bash", "-c").EntrypointShell("nginx", "-g", "daemon off;").HealthcheckNone().Run("echo a").RunLayer("echo b").Run("echo c", RunOpts{
		Mount:    []string{"type=cache,target=/cache"},
		Network:  "none",
		Security: "sandbox",
	}).CommandExec("true")

	require.NoError(t, df.Validate())
	s := df.String()
	require.True(t, strings.HasPrefix(s, "# syntax=docker/dockerfile:1\n# escape=\\\n# check=error=true\n"))
	require.Contains(t, s, "FROM scratch\n")
	require.NotContains(t, s, "FROM scratch:")
	require.Contains(t, s, "FROM --platform=linux/amd64 alpine:3.20\n")
	require.Contains(t, s, "FROM --platform=linux/arm64 golang:1.22 AS build\n")
	require.Contains(t, s, "COPY --from=build --chown=nobody:nobody --chmod=755 --exclude=*.md --parents --link out /bin/app\n")
	require.Contains(t, s, "ADD --checksum=sha256:abc --keep-git-dir=true --link repo /src\n")
	require.Contains(t, s, `SHELL ["/bin/bash", "-c"]`)
	require.Contains(t, s, "ENTRYPOINT nginx -g daemon off;\n")
	require.Contains(t, s, "HEALTHCHECK NONE\n")
	require.Contains(t, s, "RUN echo a\nRUN echo b\n")
	require.NotContains(t, s, "echo a &&")
	require.Contains(t, s, "RUN --mount=type=cache,target=/cache --network=none --security=sandbox echo c\n")
	require.Equal(t, `CMD ["true"]`+"\n", s[strings.LastIndex(s, "CMD "):])
}

func TestDockerFileHealthcheckShell(t *testing.T) {
	df := NewDockerFile().Healthcheck(Health{
		Test:     []string{"CMD-SHELL", "curl -f http://localhost || exit 1"},
		Interval: 5,
	})
	require.NoError(t, df.Validate())
	require.Equal(t,
		"HEALTHCHECK --interval=5s \\\n\tCMD curl -f http://localhost || exit 1\n",
		df.String(),
	)

	none := NewDockerFile().Healthcheck(Health{Test: []string{"NONE"}})
	require.NoError(t, none.Validate())
	require.Equal(t, "HEALTHCHECK NONE\n", none.String())

	bad := NewDockerFile().Healthcheck(Health{Timeout: -1, Test: []string{"CMD", "true"}})
	require.Error(t, bad.Validate())
	require.Empty(t, bad.String())
}

func TestDockerFileValidation(t *testing.T) {
	require.Error(t, NewDockerFile().From("scratch", "").Syntax("docker/dockerfile:1").Validate())
	require.Error(t, NewDockerFile().Syntax("a").Syntax("b").Validate())
	require.Error(t, NewDockerFile().Escape("x").Validate())
	require.Error(t, NewDockerFile().From("alpine:3", "latest").Validate())
	require.Equal(t, "CMD [\"true\"]\n", NewDockerFile().CommandExec("true").String())
	require.Equal(t, "ENTRYPOINT [\"sleep\"]\n", NewDockerFile().Entrypoint("sleep").String())
	require.Error(t, NewDockerFile().Volumes().Validate())
	require.Error(t, NewDockerFile().Copy("a", "b", CopyOpts{}, CopyOpts{From: "x"}).Validate())
	require.Error(t, NewDockerFile().Run("echo a", RunOpts{Network: "none"}, RunOpts{Network: "host"}).Validate())

	bad := NewDockerFile().RunArgs("x")
	project := NewProject("app").WithService("api", NewContainer().Image("alpine"), bad.WithInline())
	require.Error(t, project.Validate())
	svc, err := project.GetService("api")
	require.NoError(t, err)
	require.Nil(t, svc.Build)

	okDF := NewDockerFile().From("alpine", "latest").CommandExec("true")
	okProject := NewProject("app").WithService("api", NewContainer(), okDF.WithInline())
	require.NoError(t, okProject.Validate())
	okSvc, err := okProject.GetService("api")
	require.NoError(t, err)
	require.NotNil(t, okSvc.Build)
	require.Contains(t, okSvc.Build.DockerfileInline, "FROM alpine:latest")
	require.Contains(t, okSvc.Build.DockerfileInline, `CMD ["true"]`)
}

func TestDockerFileBuildContext(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "Dockerfile"), []byte("FROM scratch\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "noise.log"), []byte("no"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(src, "tmp"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "tmp", "secret"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "tmp", "keep.txt"), []byte("kept"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "keep.log"), []byte("yes"), 0644))
	script := filepath.Join(src, "script.sh")
	require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\n"), 0755))
	require.NoError(t, os.Chmod(script, 0755))
	require.NoError(t, os.Symlink("keep.txt", filepath.Join(src, "link.txt")))
	require.NoError(t, os.WriteFile(filepath.Join(src, ".dockerignore"), []byte("# comment\n*.log\n!keep.log\ntmp/\n!tmp/keep.txt\n"), 0644))

	df := NewDockerFile().From("alpine", "latest").CommandExec("true")
	reader, err := df.NewLocalBuildContext(src)
	require.NoError(t, err)

	entries := readTar(t, reader)
	require.Equal(t, "FROM alpine:latest\nCMD [\"true\"]\n", entries["Dockerfile"].body)
	require.Equal(t, "keep", entries["keep.txt"].body)
	require.Equal(t, "yes", entries["keep.log"].body)
	require.NotContains(t, entries, "noise.log")
	require.NotContains(t, entries, "tmp/secret")
	require.Equal(t, "kept", entries["tmp/keep.txt"].body)
	require.NotContains(t, entries, "tmp/")
	info, err := os.Stat(script)
	require.NoError(t, err)
	require.Equal(t, int64(info.Mode().Perm()), entries["script.sh"].header.Mode)
	link := entries["link.txt"].header
	require.Equal(t, byte(tar.TypeSymlink), link.Typeflag)
	require.Equal(t, "keep.txt", link.Linkname)

	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0644))
	_, err = df.NewLocalBuildContext(file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a directory")

	badIgnore := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(badIgnore, ".dockerignore"), []byte("[\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(badIgnore, "a.txt"), []byte("a"), 0644))
	_, err = df.NewLocalBuildContext(badIgnore)
	require.Error(t, err)
}

type tarEntry struct {
	header *tar.Header
	body   string
}

func readTar(t *testing.T, r io.Reader) map[string]tarEntry {
	t.Helper()
	out := map[string]tarEntry{}
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return out
		}
		require.NoError(t, err)
		copied := *header
		entry := tarEntry{header: &copied}
		if header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeSymlink {
			body, err := io.ReadAll(tr)
			require.NoError(t, err)
			entry.body = string(body)
		}
		out[header.Name] = entry
	}
}

type errString string

func (e errString) Error() string { return string(e) }
