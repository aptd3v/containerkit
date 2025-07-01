// This is a simple example of how to use go-contain to create a simple project.
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

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
)

const (
	ProjectName   = "simple-project"
	ContainerName = "simple-container"
	ServiceName   = "simple"
)

func main() {
	project := create.NewProject(ProjectName)
	project.WithService(ServiceName, AlpineContainer("latest"))

	ctx := context.Background()
	app := compose.NewCompose(project)
	if err := app.Up(ctx, up.WithWriter(os.Stdout)); err != nil {
		log.Fatal(err)
	}
}

func AlpineContainer(tag string) *create.Container {
	simple := create.NewContainer(ContainerName)
	simple.WithContainerConfig(
		cc.WithImagef("alpine:%s", tag),
		cc.WithCommand("echo", "hello, world"),
	)
	return simple
}
