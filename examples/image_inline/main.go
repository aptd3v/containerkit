// this program runs and builds the image if it does not exist and then tags it with a label
// and then uses the image for a container the second time it runs
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func main() {
	ctx := context.Background()
	exists, err := ImageExists(ctx, "my-image", "thats-tagged")
	if err != nil {
		log.Fatalf("Error checking if image exists: %v", err)
	}

	project := containerkit.NewProject("my-image-project")
	project.WithService(
		"my-image-service",
		MyContainer("my-image", "thats-tagged"),
		inlineDockerfile("alpine", "latest", !exists),
	)

	example := containerkit.NewCompose(project)

	err = example.Up(ctx, &containerkit.Up{Build: !exists, Detach: true})
	if err != nil {
		log.Fatalf("Error upping project: %v", err)
	}
	err = example.Logs(ctx, &containerkit.Logs{NoLogPrefix: true, Tail: intPtr(1)})
	if err != nil {
		log.Fatalf("Error getting logs: %v", err)
	}
}

func inlineDockerfile(image, tag string, addLabel bool) containerkit.BuildSpec {
	df := containerkit.NewDockerFile()
	df.From(image, tag)
	df.Workdir("/app")
	df.Run("echo \"Saving Hello, World!\"")
	df.Run("echo \"saved: Hello, World!\" > /app/hello.txt")
	df.CommandExec("cat", "/app/hello.txt")
	spec := df.WithInline()
	if addLabel {
		spec.Labels = map[string]string{"my-image": "thats-tagged"}
	}
	return spec
}

func ImageExists(ctx context.Context, imageName, tag string) (bool, error) {
	host := os.Getenv("DOCKER_HOST")
	var opts []client.SetClientOption
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}
	cli, err := client.NewClient(opts...)
	if err != nil {
		return false, err
	}

	res, err := cli.ImageInspect(ctx, fmt.Sprintf("%s:%s", imageName, tag))
	if err != nil {
		return false, nil
	}
	fmt.Println("Found image", strings.Join(res.RepoTags, ","))
	return true, nil
}

func MyContainer(imageName, tag string) *containerkit.Container {
	return containerkit.NewContainer().Imagef("%s:%s", imageName, tag)
}

func intPtr(i int) *int { return &i }
