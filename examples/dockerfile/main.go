// Write a multi-stage Dockerfile with NewDockerFile, build it with the
// Engine client, and serve it.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

const (
	imageRef       = "dockerfile-example:latest"
	containerName  = "dockerfile-example"
	hostPort       = "8088"
	dockerfilePath = "./examples/dockerfile/Dockerfile"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	df := kit.NewDockerFile().
		Syntax("docker/dockerfile:1").
		Comment("stage a static page, then serve it with busybox httpd").
		FromAs("alpine:latest", "build").
		Workdir("/out").
		Run(`printf 'Hello from containerkit\n' > index.html`).
		Run("chmod 644 index.html").
		From("alpine:latest").
		Run("apk add --no-cache busybox-extras").
		Copy("/out/index.html", "/www/index.html", kit.CopyOpts{From: "build"}).
		Label("example", "dockerfile").
		Expose("8080").
		User("nobody").
		Healthcheck(kit.NewHealth().
			CMDShell("wget -qO- http://127.0.0.1:8080/ >/dev/null").
			IntervalD("5s").
			TimeoutD("2s").
			Retries(3)).
		Entrypoint("/usr/sbin/httpd", "-f", "-p", "8080", "-h", "/www")
	if err := df.Validate(); err != nil {
		return fmt.Errorf("dockerfile: %w", err)
	}
	if err := df.Export(dockerfilePath, 0644); err != nil {
		return fmt.Errorf("export: %w", err)
	}
	fmt.Print(df.String())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionFromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()

	buildCtx, err := kit.NewBuild().Dockerfile(df).Reader()
	if err != nil {
		return fmt.Errorf("build context: %w", err)
	}
	buildOpt := &client.ImageBuild{}
	buildOpt.Tags = []string{imageRef}
	buildOpt.Remove = true
	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpt)
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		return fmt.Errorf("build output: %w", err)
	}

	ctr := kit.NewContainer(containerName).
		Image(imageRef).
		PortBinding(types.PortBinding{
			Protocol:      types.TCP,
			HostIP:        "0.0.0.0",
			HostPort:      hostPort,
			ContainerPort: "8080",
		})
	if err := ctr.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if err := cli.ContainerRemove(ctx, containerName, &client.Remove{Force: true}); err != nil && !client.IsErrNotFound(err) {
		return fmt.Errorf("remove: %w", err)
	}
	created, err := cli.ContainerCreate(ctx, ctr)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer func() {
		cleanup := context.Background()
		if err := cli.ContainerStop(cleanup, created.ID, nil); err != nil {
			log.Printf("stop: %v", err)
		}
		if err := cli.ContainerRemove(cleanup, created.ID, &client.Remove{Force: true}); err != nil {
			log.Printf("remove: %v", err)
		}
	}()
	if err := cli.ContainerStart(ctx, created.ID, nil); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	fmt.Printf("serving http://localhost:%s  (Ctrl+C to stop)\n", hostPort)
	<-ctx.Done()
	return nil
}
