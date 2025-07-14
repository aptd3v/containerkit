// this example shows how to use profiles to start a service with a specific profile
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
	"github.com/aptd3v/go-contain/pkg/tools"
)

var (
	paragraph = `go-contain
is a Go library that provides a programmatic and composable interface
for defining, running, and managing Docker containers and Compose projects.
It abstracts both the Docker SDK and Docker Compose into a unified API.
In this example, we'll use profiles and demonstrate how to start a service
with a specific profile.
`

	selectedProfile = "message-chain"
)

func main() {
	project := create.NewProject("my-project")
	project.WithService("never-service", create.NewContainer("never-service").
		WithContainerConfig(
			cc.WithImage("alpine:latest"),
			cc.WithCommand("echo", "you wont see me"),
		),
		sc.WithProfiles("never"),
	)
	project.WithService("never-ever-service", create.NewContainer("never-ever-service").
		WithContainerConfig(
			cc.WithImage("alpine:latest"),
			cc.WithCommand("echo", "you wont see me part II"),
		),
		sc.WithProfiles("never-ever"),
	)

	for i, word := range strings.Split(paragraph, "\n") {
		serviceName := fmt.Sprintf("service%d", i)
		project.WithService(serviceName,
			create.NewContainer().
				WithContainerConfig(
					cc.WithImage("alpine:latest"),
					cc.WithCommand("echo", word),
				),
			sc.WithProfiles(selectedProfile),
			WithDependencyChain(i, "service%d", i-1),
		)
	}
	example := compose.NewCompose(project)
	err := example.Up(
		context.Background(),
		up.WithProfiles(selectedProfile),
		up.WithNoLogPrefix(),
		up.WithRemoveOrphans(),
	)

	if err != nil {
		log.Fatalf("error executing example 'up' with profile 'tail': %v", err)
	}

}

func WithDependencyChain(index int, format string, a ...any) create.SetServiceConfig {
	return tools.WhenTrue(index > 0, sc.WithDependsOn(fmt.Sprintf(format, a...)))
}
