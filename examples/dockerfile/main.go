// Write a multi-stage Dockerfile, build it, and serve the page on port 8088.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

const (
	imageRef       = "dockerfile-example:latest"
	containerName  = "dockerfile-example"
	hostPort       = 8088
	dockerfilePath = "./examples/dockerfile/Dockerfile"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	df := NewDockerfile().
		Syntax("docker/dockerfile:1").
		Comment("stage a static page, then serve it with busybox httpd").
		FromAs("alpine:latest", "build").
		Workdir("/out").
		Run(`printf 'Hello from containerkit\n' > index.html`).
		Run("chmod 644 index.html").
		From("alpine", "latest").
		Run("apk add --no-cache busybox-extras").
		Copy("/out/index.html", "/www/index.html", CopyOpts{From: "build"}).
		Label("example", "dockerfile").
		Expose("8080").
		User("nobody").
		HealthCheck(
			Test("CMD-SHELL", "wget -qO- http://127.0.0.1:8080/ >/dev/null"),
			Interval(5*time.Second),
			Timeout(2*time.Second),
			Retries(3),
		).
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

	cli, err := Connect(FromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := BuildImage(ctx, cli.Engine(), df, Image(imageRef), Force(true)); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	web := NewContainer(
		Name(containerName),
		Image(imageRef),
		Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(hostPort), ContainerPort(8080)),
	)
	if err := web.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	if err := web.Remove(ctx, cli.Engine(), containerName, Force(true)); err != nil && !IsNotFound(err) {
		return fmt.Errorf("remove: %w", err)
	}
	created, err := web.Create(ctx, cli.Engine())
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer func() {
		cleanup := context.Background()
		if err := web.Stop(cleanup, cli.Engine(), created.ID); err != nil && !IsNotFound(err) {
			log.Printf("stop: %v", err)
		}
		if err := web.Remove(cleanup, cli.Engine(), created.ID, Force(true)); err != nil && !IsNotFound(err) {
			log.Printf("remove: %v", err)
		}
	}()
	if err := web.Start(ctx, cli.Engine(), created.ID); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	fmt.Printf("serving http://localhost:%d  (Ctrl+C to stop)\n", hostPort)
	<-ctx.Done()
	return nil
}
