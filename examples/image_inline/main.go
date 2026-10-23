// Build a tagged image from an inline Dockerfile, then run it.
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
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

const imageRef = "hello-inline:example"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionFromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()

	img, err := cli.ImageInspect(ctx, imageRef, nil)
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}

	df := containerkit.NewDockerFile().
		From("alpine", "latest").
		Label("example", "image-inline").
		Workdir("/app").
		Run(`printf 'Hello, World!\n' > hello.txt`).
		CommandExec("cat", "hello.txt")

	project := containerkit.NewProject("image-inline")
	project.WithService("hello",
		containerkit.NewContainer().Image(imageRef),
		df.WithInline(),
	)
	if err := project.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	app := containerkit.NewCompose(project)
	defer func() {
		if err := app.Down(context.Background(), &containerkit.Down{RemoveOrphans: true}); err != nil {
			log.Printf("down: %v", err)
		}
	}()

	if img == nil {
		fmt.Print(df.String())
	} else {
		fmt.Printf("reusing %s\n", imageRef)
	}

	return app.Up(ctx, &containerkit.Up{
		Build:         img == nil,
		RemoveOrphans: true,
		Writer:        os.Stdout,
	})
}
