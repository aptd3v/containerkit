// this example shows how to use profiles to start a service with a specific profile
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
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
	project := kit.NewProject("my-project")
	project.WithService(kit.NewService("never-service",
		kit.NewContainer().
			Image("alpine:latest").
			Command("echo", "you wont see me"),
	).Profiles("never"))
	project.WithService(kit.NewService("never-ever-service",
		kit.NewContainer().
			Image("alpine:latest").
			Command("echo", "you wont see me part II"),
	).Profiles("never-ever"))
	project.WithService(kit.NewService("kill-me-service",
		kit.NewContainer().
			Image("alpine:latest").
			Command("tail", "-f", "/dev/null"),
	).Profiles("kill-me"))

	words := strings.Split(paragraph, "\n")
	project.WithService(kit.NewService("service0",
		kit.NewContainer().Image("alpine:latest").Command("echo", words[0]),
	).Profiles(selectedProfile))
	for i := 1; i < len(words); i++ {
		project.WithService(kit.NewService(fmt.Sprintf("service%d", i),
			kit.NewContainer().Image("alpine:latest").Command("echo", words[i]),
		).Profiles(selectedProfile).DependsOn(fmt.Sprintf("service%d", i-1)))
	}
	example := kit.NewCompose(project)
	err := example.Up(
		context.Background(),
		&kit.Up{
			Profiles:      []string{selectedProfile},
			NoLogPrefix:   true,
			RemoveOrphans: true,
		},
	)

	if err != nil {
		log.Fatalf("error executing example 'up' with profile 'tail': %v", err)
	}
	err = example.Up(context.Background(),
		&kit.Up{
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
		&kit.Kill{
			Signal:        &sig,
			RemoveOrphans: true,
			Profiles:      []string{"kill-me"},
		},
	)
	if err != nil {
		log.Fatalf("error executing example 'kill' with profile 'kill-me': %v", err)
	}
}
