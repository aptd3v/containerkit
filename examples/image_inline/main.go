// Build an image when the tag is missing, then bring the service up.
// A later run finds the tag and skips the build.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

const (
	baseImage   = "alpine:latest"
	taggedImage = "hello-inline:v1"
	svcName     = "hello-image-inline"
	projectName = "image-inline"
	workdir     = "/app"
	labelKey    = "example"
	labelValue  = "image-inline"
)

func main() {
	if err := run(context.Background(), "echo 'Hello, World!' > hello.txt"); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cmd string) error {
	cli, err := Connect(FromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()

	exists, err := imageExists(ctx, cli, taggedImage)
	if err != nil {
		return fmt.Errorf("image exists: %w", err)
	}

	df := NewDockerfile().
		From(baseImage).
		Label(labelKey, labelValue).
		Workdir(workdir).
		Run(cmd).
		Command("sleep", "5s")
	app := NewCompose(projectName).WithService(NewService(
		Name(svcName),
		Image(taggedImage),
		Build(df),
	))
	if err := app.Write(Export{Writer: os.Stdout}); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	defer func() {
		if err := app.Down(context.Background(), cli.Composer(), RemoveOrphans(true)); err != nil {
			log.Printf("down: %v", err)
		}
	}()

	if !exists {
		if err := app.Build(ctx, cli.Composer(), OutputStream(os.Stdout)); err != nil {
			return fmt.Errorf("build: %w", err)
		}
	}
	return app.Up(ctx, cli.Composer(), OutputStream(os.Stdout))
}

func imageExists(ctx context.Context, cli *Client, ref string) (bool, error) {
	_, err := InspectImage(ctx, cli.Engine(), ref)
	if IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}
