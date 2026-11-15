// Profiles hide services until the matching profile is active. message-chain
// echoes one paragraph, one line per service. kill-me stays up until SIGKILL.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	. "github.com/aptd3v/containerkit/pkg/kit"
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
	app := NewCompose("my-project").
		WithService(NewService(
			Name("never-service"),
			Image("alpine:latest"),
			Command("echo", "you wont see me"),
			Profile("never"),
		)).
		WithService(NewService(
			Name("never-ever-service"),
			Image("alpine:latest"),
			Command("echo", "you wont see me part II"),
			Profile("never-ever"),
		)).
		WithService(NewService(
			Name("kill-me-service"),
			Image("alpine:latest"),
			Command("tail", "-f", "/dev/null"),
			Profile("kill-me"),
		))

	words := strings.Split(paragraph, "\n")
	app.WithService(NewService(
		Name("service0"),
		Image("alpine:latest"),
		Command("echo", words[0]),
		Profile(selectedProfile),
	))
	for i := 1; i < len(words); i++ {
		app.WithService(NewService(
			Name(fmt.Sprintf("service%d", i)),
			Image("alpine:latest"),
			Command("echo", words[i]),
			Profile(selectedProfile),
			DependsOn(Name(fmt.Sprintf("service%d", i-1))),
		))
	}

	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	chain, err := withProfiles(app, selectedProfile)
	if err != nil {
		log.Fatal(err)
	}
	if err := chain.Up(context.Background(), cli.Composer(), RemoveOrphans(true)); err != nil {
		log.Fatalf("up %s: %v", selectedProfile, err)
	}

	killMe, err := withProfiles(app, "kill-me")
	if err != nil {
		log.Fatal(err)
	}
	if err := killMe.Up(context.Background(), cli.Composer(), Detach(true), RemoveOrphans(true)); err != nil {
		log.Fatalf("up kill-me: %v", err)
	}
	if err := killMe.Kill(context.Background(), cli.Composer(), StopSignal(SIGKILL), RemoveOrphans(true)); err != nil {
		log.Fatalf("kill kill-me: %v", err)
	}
}

// withProfiles keeps services that belong to one of profiles. Compose leaves
// a profiled service out until that profile is selected at load time.
func withProfiles(base *Compose, profiles ...string) (*Compose, error) {
	yml, err := base.ExportYAML()
	if err != nil {
		return nil, err
	}
	opts := make([]ImportOption, 0, len(profiles))
	for _, profile := range profiles {
		opts = append(opts, Profile(profile))
	}
	loaded := ImportYAML(yml, opts...)
	if err := loaded.Validate(); err != nil {
		return nil, err
	}
	return loaded, nil
}
