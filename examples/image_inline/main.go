// Build an image if it doesn't exist, tag it, then run it. Subsequent runs
// reuse the image and Compose skips the build.
//
// The Engine client inspects the tag first. A second run reuses the image
// and Compose skips the build.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
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
	ctx := context.Background()
	if err := run(ctx, "echo 'Hello, World!' > hello.txt"); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cmd string) error {
	imageExists, err := ImageExists(ctx, taggedImage)
	if err != nil {
		return fmt.Errorf("image exists: %w", err)
	}
	project := kit.NewProject(projectName)
	project.WithService(
		kit.NewService(
			svcName,
			SimpleContainer(taggedImage),
		).
			Dockerfile(
				kit.NewDockerFile().
					From(baseImage).
					Label(labelKey, labelValue).
					Workdir(workdir).
					Run(cmd).
					Command("sleep", "5s"),
			))
	_, err = project.WriteTo(os.Stdout)
	if err != nil {
		return fmt.Errorf("write to stdout: %w", err)
	}

	app := kit.NewCompose(project)

	defer func() {
		if err := Down(ctx, app); err != nil {
			log.Printf("down: %v", err)
		}
	}()

	return Up(ctx, app, imageExists)
}

func Down(ctx context.Context, app *kit.Compose) error {
	return app.Down(ctx, &kit.Down{RemoveOrphans: true})
}

func Up(ctx context.Context, app *kit.Compose, exists bool) error {
	return app.Up(ctx, &kit.Up{
		Build:         !exists,
		RemoveOrphans: false,
		Writer:        os.Stdout,
	})
}
func SimpleContainer(image string) *kit.Container {
	return kit.NewContainer().Image(image)
}
func ImageExists(ctx context.Context, ref string) (bool, error) {
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionFromEnv())
	if err != nil {
		return false, err
	}
	defer cli.Close()
	img, err := cli.ImageInspect(ctx, ref, nil)
	if client.IsErrNotFound(err) {
		return false, nil
	}
	return img != nil, err
}
