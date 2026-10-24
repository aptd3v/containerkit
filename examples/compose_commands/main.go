// This example demonstrates compose CLI commands: Build, Ps, Start, Stop, Restart,
// Exec, Config, Images, Top, Pause, Unpause, and Run.
//
// It builds the app service, brings the stack up (Up), lists containers (Ps),
// then runs Stop, Start, and Restart before tearing down on Ctrl+C (Down).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

const (
	projectName  = "compose-commands-example"
	appService   = "app"
	webService   = "web"
	redisService = "redis"
)

func inlineDockerfile(image, tag string) *kit.Dockerfile {
	return kit.NewDockerFile().
		From(image, tag).
		Workdir("/app").
		Run("echo \"Hello, World!\"").
		Command("tail", "-f", "/dev/null")
}

func main() {
	project := kit.NewProject(projectName)
	project.WithNetwork("backend")
	project.WithService(kit.NewService(appService, kit.NewContainer(appService)).
		Dockerfile(inlineDockerfile("alpine", "latest")),
	)
	project.WithService(kit.NewService(webService, kit.NewContainer(webService).
		Image("nginx:alpine").
		Command("nginx", "-g", "daemon off;").
		PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "9080", ContainerPort: "80"}).
		Endpoint("backend"),
	))
	project.WithService(kit.NewService(redisService, kit.NewContainer(redisService).
		Image("redis:7-alpine").
		Command("redis-server").
		Endpoint("backend"),
	))

	app := kit.NewCompose(project)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	fmt.Println("Building app service...")
	if err := app.Build(ctx, &kit.ComposeBuild{Writer: os.Stdout, ServiceNames: []string{appService}}); err != nil {
		log.Fatalf("Build: %v", err)
	}

	fmt.Println("\nStarting stack...")
	if err := app.Up(ctx, &kit.Up{Detach: true, RemoveOrphans: true, Writer: os.Stdout}); err != nil {
		log.Fatalf("Up: %v", err)
	}

	fmt.Println("\n--- docker compose ps -a ---")
	if err := app.Ps(ctx, &kit.Ps{All: true, Writer: os.Stdout}); err != nil {
		log.Printf("Ps: %v", err)
	}

	fmt.Println("\n--- docker compose ps -a --format json (parsed) ---")
	var psOut bytes.Buffer
	if err := app.Ps(ctx, &kit.Ps{All: true, Format: "json", Writer: &psOut}); err != nil {
		log.Printf("Ps (json): %v", err)
	} else {
		var entries []struct {
			ID      string `json:"ID"`
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
		}
		raw := psOut.Bytes()
		if err := json.Unmarshal(raw, &entries); err != nil {
			for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
				if len(line) == 0 {
					continue
				}
				var e struct {
					ID      string `json:"ID"`
					Name    string `json:"Name"`
					Service string `json:"Service"`
					State   string `json:"State"`
				}
				if err := json.Unmarshal(line, &e); err != nil {
					log.Printf("Ps (json line): %v", err)
					continue
				}
				entries = append(entries, e)
			}
		}
		fmt.Println("SERVICE  STATE    NAME")
		for _, e := range entries {
			fmt.Printf("%-8s  %-8s  %-8s  %s\n", e.ID, e.Service, e.State, e.Name)
		}
	}

	fmt.Println("\nStopping web...")
	stopTimeout := 5
	if err := app.Stop(ctx, &kit.Stop{Timeout: &stopTimeout, ServiceNames: []string{webService}}); err != nil {
		log.Printf("Stop: %v", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("\n--- docker compose ps -a (after stop web) ---")
	if err := app.Ps(ctx, &kit.Ps{All: true, Writer: os.Stdout}); err != nil {
		log.Printf("Ps: %v", err)
	}

	fmt.Println("\nStarting web...")
	if err := app.Start(ctx, &kit.Start{ServiceNames: []string{webService}}); err != nil {
		log.Printf("Start: %v", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("\nRestarting redis...")
	restartTimeout := 5
	if err := app.Restart(ctx, &kit.Restart{NoDeps: true, Timeout: &restartTimeout, ServiceNames: []string{redisService}}); err != nil {
		log.Printf("Restart: %v", err)
	}

	fmt.Println("\n--- docker compose exec (non-interactive) ---")
	if err := app.Exec(ctx, &kit.Exec{
		Service: appService,
		Command: []string{"sh", "-c", "echo hello from compose exec"},
		NoTTY:   true,
		Writer:  os.Stdout,
	}); err != nil {
		log.Printf("Exec: %v", err)
	}

	fmt.Println("\n--- docker compose config --services ---")
	if err := app.Config(ctx, &kit.ComposeConfig{PrintServices: true, Writer: os.Stdout}); err != nil {
		log.Printf("Config: %v", err)
	}

	fmt.Println("\n--- docker compose images ---")
	if err := app.Images(ctx, &kit.Images{Writer: os.Stdout}); err != nil {
		log.Printf("Images: %v", err)
	}

	fmt.Println("\n--- docker compose top (app) ---")
	if err := app.Top(ctx, &kit.Top{ServiceNames: []string{appService}, Writer: os.Stdout}); err != nil {
		log.Printf("Top: %v", err)
	}

	fmt.Println("\n--- docker compose pause / unpause (web) ---")
	if err := app.Pause(ctx, &kit.Pause{ServiceNames: []string{webService}, Writer: os.Stdout}); err != nil {
		log.Printf("Pause: %v", err)
	}
	if err := app.Unpause(ctx, &kit.Unpause{ServiceNames: []string{webService}, Writer: os.Stdout}); err != nil {
		log.Printf("Unpause: %v", err)
	}

	fmt.Println("\n--- docker compose run (one-off) ---")
	if err := app.Run(ctx, &kit.Run{
		Service: appService,
		Command: []string{"echo", "hello from compose run"},
		NoTTY:   true,
		Rm:      true,
		NoDeps:  true,
		Writer:  os.Stdout,
	}); err != nil {
		log.Printf("Run: %v", err)
	}

	defer func() {
		fmt.Println("\nTearing down...")
		if err := app.Down(context.Background(), &kit.Down{RemoveOrphans: true, Writer: os.Stdout}); err != nil {
			log.Printf("Down: %v", err)
		}
	}()

	fmt.Println("\nStack running (Build, Ps, Stop, Start, Restart, Exec, Config, Images, Top, Pause, Run demonstrated). Press Ctrl+C to stop and remove.")
	<-ctx.Done()
}
