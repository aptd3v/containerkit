// this example shows how to use profiles to start a service with a specific profile
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

var (
	paragraph = `containerkit
is a Go library that provides a programmatic and composable interface
for defining, running, and managing Docker containers and Compose projects.
It abstracts both the Docker SDK and Docker Compose into a unified API.
In this example, we'll use profiles and demonstrate how to start a service
with a specific profile.
`

	selectedProfile = "message-chain"
)

func main() {
	project := containerkit.NewProject("my-project")
	project.WithService("never-service",
		containerkit.NewContainer().
			Image("alpine:latest").
			Command("echo", "you wont see me"),
		containerkit.Profiles("never"),
	)
	project.WithService("never-ever-service",
		containerkit.NewContainer().
			Image("alpine:latest").
			Command("echo", "you wont see me part II"),
		containerkit.Profiles("never-ever"),
	)
	project.WithService("kill-me-service",
		containerkit.NewContainer().
			Image("alpine:latest").
			Command("tail", "-f", "/dev/null"),
		containerkit.Profiles("kill-me"),
	)

	for i, word := range strings.Split(paragraph, "\n") {
		serviceName := fmt.Sprintf("service%d", i)
		extras := []any{containerkit.Profiles(selectedProfile)}
		if i > 0 {
			extras = append(extras, containerkit.DependsOn(fmt.Sprintf("service%d", i-1)))
		}
		project.WithService(serviceName,
			containerkit.NewContainer().
				Image("alpine:latest").
				Command("echo", word),
			extras...,
		)
	}
	example := containerkit.NewCompose(project)
	err := example.Up(
		context.Background(),
		&containerkit.Up{
			Profiles:      []string{selectedProfile},
			NoLogPrefix:   true,
			RemoveOrphans: true,
		},
	)

	if err != nil {
		log.Fatalf("error executing example 'up' with profile 'tail': %v", err)
	}
	err = example.Up(context.Background(),
		&containerkit.Up{
			Profiles:      []string{"kill-me"},
			NoLogPrefix:   true,
			RemoveOrphans: true,
			Detach:        true,
		},
	)
	if err != nil {
		log.Fatalf("error executing example 'up' with profile 'kill-me': %v", err)
	}

	sig := "SIGKILL"
	err = example.Kill(context.Background(),
		&containerkit.Kill{
			Signal:        &sig,
			RemoveOrphans: true,
			Profiles:      []string{"kill-me"},
		},
	)
	if err != nil {
		log.Fatalf("error executing example 'kill' with profile 'kill-me': %v", err)
	}
}
