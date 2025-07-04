// this program runs and builds the image if it does not exist and then tags it with a label
// and then uses the image for a container the second time it runs
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aptd3v/go-contain/pkg/client"
	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/logs"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc/build"
	"github.com/aptd3v/go-contain/pkg/tools"
)

func main() {
	ctx := context.Background()
	// check if the image exists in the local docker daemon
	exists, err := ImageExists(ctx, "my-image", "thats-tagged")
	if err != nil {
		log.Fatalf("Error checking if image exists: %v", err)
	}

	// create a project with a service that uses the image
	project := create.NewProject("my-image-project")
	project.WithService(
		"my-image-service",

		// create a container with the image name and tag
		// notice that the image name and tag are not created just yet
		MyContainer("my-image", "thats-tagged"),

		// create an inline dockerfile that will be used to build the image
		WithInlineDockerfile("alpine", "latest"),

		// if the image does not exist, tag the image after building it
		WithTagImageIfNotExists(exists),
	)

	example := compose.NewCompose(project)

	// if the image does not exist, use the inline dockerfile
	err = example.Up(ctx, WithBuildIfNotExists(exists), up.WithDetach())
	if err != nil {
		log.Fatalf("Error upping project: %v", err)
	}
	err = example.Logs(ctx, logs.WithNoLogPrefix(), logs.WithTail(1))
	if err != nil {
		log.Fatalf("Error getting logs: %v", err)
	}

	// run docker image ls --filter label=my-image=thats-tagged to see the tagged image

}

// WithBuildIfNotExists sets the --build option in the compose up command if the image does not exist
func WithBuildIfNotExists(exists bool) compose.SetComposeUpOption {
	return tools.WhenTrue(!exists, up.WithBuild())
}

// WithTagImageIfNotExists tags the image with a label if it does not exist
func WithTagImageIfNotExists(exists bool) create.SetServiceConfig {
	return tools.WhenTrue(!exists, sc.WithBuild(
		build.WithLabels("my-image", "thats-tagged"),
	))
}

// ImageExists checks if the image exists in the local docker daemon
func ImageExists(ctx context.Context, imageName, tag string) (bool, error) {
	cli, err := client.NewClient()
	if err != nil {
		return false, err
	}
	// image inspect returns an error if the image does not exist
	_, err = cli.ImageInspect(ctx, fmt.Sprintf("%s:%s", imageName, tag))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// WithInlineDockerfile returns a string that can be used as an inline dockerfile
// within compose yaml
func WithInlineDockerfile(imageName, tag string) create.SetServiceConfig {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("FROM  %s:%s\n", imageName, tag))
	sb.WriteString("WORKDIR /app\n")
	sb.WriteString("RUN  echo \"Saving Hello, World!\" && \\\n")
	sb.WriteString("echo \"saved: Hello, World!\" > /app/hello.txt\n")
	sb.WriteString("CMD  [\"cat\", \"/app/hello.txt\"]\n")
	return sc.WithBuild(
		build.WithDockerfileInline(sb.String()),
	)
}

// MyContainer returns a container with the image name and tag
func MyContainer(imageName, tag string) *create.Container {
	return create.NewContainer("my-image-service-container").
		WithContainerConfig(
			cc.WithImagef("%s:%s", imageName, tag),
		)
}
