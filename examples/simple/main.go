// This is a simple example of how to use containerkit to create a simple project.
//
// this is the equivalent of the following docker-compose.yml file:
//
//	name: simple-project
//	services:
//
//	simple:
//	  container_name: simple-container
//	  command:
//	    - echo
//	    - hello, world
//	  image: alpine:latest
//
// and the running docker compose up
package main

import (
	"context"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

const (
	ProjectName = "simple-project"
	ServiceName = "simple"
)

func main() {
	project := containerkit.NewProject(ProjectName)
	project.WithService(ServiceName, AlpineContainer("latest"))

	ctx := context.Background()
	app := containerkit.NewCompose(project)
	if err := app.Up(ctx, &containerkit.Up{Writer: os.Stdout}); err != nil {
		log.Fatal(err)
	}
}

func AlpineContainer(tag string) *containerkit.Container {
	return containerkit.NewContainer().
		Imagef("alpine:%s", tag).
		Command("echo", "hello, world")
}
