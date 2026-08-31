//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func TestDockerFileWriterRealBuild(t *testing.T) {
	t.Parallel()
	cli := newCLI(t)
	pullAlpine(t, cli)

	src := t.TempDir()
	mustWrite := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, name), []byte(body), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	mustWrite("hello.txt", "hello-from-copy")
	mustWrite("my file.txt", "spaced-copy")
	mustWrite("extra.txt", "from-add")

	df := containerkit.NewDockerFile().
		Comment("writer e2e for %s").
		FromAs("%s:%s", "base").
		From("%s", "%s").
		Format("alpine", "alpine", "latest", "alpine", "latest").
		Arg("MSG").
		ArgKV("APP", "e2e").
		Env("APP", "e2e").
		Env("GREETING", "hello").
		Label("e2e.writer", "1").
		Workdir("/app").
		Copy("hello.txt", "/app/hello.txt").
		Copy("my file.txt", "/app/my file.txt").
		Add("extra.txt", "/app/extra.txt").
		Run("echo run-ok > /app/run.txt").
		Run("true").
		RunArgs("true").
		Expose("8080").
		Onbuild("RUN true").
		StopSignal("SIGTERM").
		Volumes("/data").
		Healthcheck(containerkit.Health{
			Test:     []string{"CMD", "true"},
			Timeout:  2,
			Interval: 5,
			Retries:  1,
		}).
		User("nobody").
		CommandExec("sleep", "infinity")
	if err := df.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	fmt.Fprintln(testWriter(t), df.String())
	exported := filepath.Join(src, "exported.Dockerfile")
	if err := df.Export(exported, 0644); err != nil {
		t.Fatalf("Export: %v", err)
	}

	ctxReader, err := df.NewLocalBuildContext(src)
	if err != nil {
		t.Fatalf("NewLocalBuildContext: %v", err)
	}

	tag := ident(t, "i") + ":e2e"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	opt := &client.ImageBuild{}
	opt.Tags = []string{tag}
	opt.Remove = true
	resp, err := cli.ImageBuild(ctx, ctxReader, opt)
	if err != nil {
		t.Fatalf("ImageBuild: %v", err)
	}
	drainBuild(t, resp.Body)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = cli.ImageRemove(c, tag, &client.ImageRemove{Force: true, PruneChildren: true})
	})

	img, err := cli.ImageInspect(ctx, tag)
	if err != nil {
		t.Fatalf("ImageInspect: %v", err)
	}
	if img.Config == nil {
		t.Fatal("ImageInspect: nil Config")
	}
	cfg := img.Config
	if !containsEnv(cfg.Env, "APP=e2e") || !containsEnv(cfg.Env, "GREETING=hello") {
		t.Fatalf("env %v", cfg.Env)
	}
	if cfg.Labels["e2e.writer"] != "1" {
		t.Fatalf("label e2e.writer=%q", cfg.Labels["e2e.writer"])
	}
	if cfg.WorkingDir != "/app" {
		t.Fatalf("workdir %q", cfg.WorkingDir)
	}
	if cfg.User != "nobody" {
		t.Fatalf("user %q", cfg.User)
	}
	if cfg.StopSignal != "SIGTERM" {
		t.Fatalf("stopsignal %q", cfg.StopSignal)
	}
	if _, ok := cfg.ExposedPorts["8080/tcp"]; !ok {
		t.Fatalf("exposed ports %v", cfg.ExposedPorts)
	}
	if _, ok := cfg.Volumes["/data"]; !ok {
		t.Fatalf("volumes %v", cfg.Volumes)
	}
	if cfg.Healthcheck == nil || len(cfg.Healthcheck.Test) == 0 {
		t.Fatal("missing healthcheck")
	}
	if len(cfg.OnBuild) == 0 {
		t.Fatalf("onbuild %v", cfg.OnBuild)
	}

	name := ident(t, "c")
	ctr := containerkit.NewContainer(name).Image(tag)
	if err := ctr.Validate(); err != nil {
		t.Fatalf("container Validate: %v", err)
	}
	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	removeContainer(t, cli, created.ID)
	if err := cli.ContainerStart(ctx, created.ID, nil); err != nil {
		t.Fatalf("ContainerStart: %v", err)
	}

	out := execOutput(t, cli, created.ID,
		"sh", "-c", `cat /app/hello.txt; echo; cat "/app/my file.txt"; echo; cat /app/extra.txt; echo; cat /app/run.txt`,
	)
	for _, want := range []string{"hello-from-copy", "spaced-copy", "from-add", "run-ok"} {
		if !strings.Contains(out, want) {
			t.Fatalf("exec output %q, want %q", out, want)
		}
	}
}

func containsEnv(env []string, kv string) bool {
	for _, e := range env {
		if e == kv {
			return true
		}
	}
	return false
}

func execOutput(t *testing.T, cli *client.Client, id string, cmd ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	execResp, err := cli.ContainerExecCreate(ctx, id, &client.Exec{
		AttachStdout: true,
		AttachStderr: true,
		User:         "root",
		Command:      cmd,
	})
	if err != nil {
		t.Fatalf("ContainerExecCreate: %v", err)
	}
	hij, err := cli.ContainerExecAttach(ctx, execResp.ID, &client.ExecAttach{})
	if err != nil {
		t.Fatalf("ContainerExecAttach: %v", err)
	}
	out := drainExec(t, hij.Reader)
	hij.Close()
	return out
}
