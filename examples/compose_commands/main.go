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

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

const (
	projectName  = "compose-commands-example"
	appService   = "app"
	webService   = "web"
	redisService = "redis"
)

func inlineDockerfile(image, tag string) containerkit.BuildSpec {
	df := containerkit.NewDockerFile()
	df.From(image, tag)
	df.Workdir("/app")
	df.Run("echo \"Hello, World!\"")
	df.CommandExec("tail", "-f", "/dev/null")
	return df.WithInline()
}

func main() {
	project := containerkit.NewProject(projectName)
	project.WithNetwork("backend")
	project.WithService(appService, containerkit.NewContainer(appService),
		inlineDockerfile("alpine", "latest"),
	)
	project.WithService(webService, containerkit.NewContainer(webService).
		Image("nginx:alpine").
		Command("nginx", "-g", "daemon off;").
		PortBindings("tcp", "0.0.0.0", "9080", "80").
		Endpoint("backend"),
	)
	project.WithService(redisService, containerkit.NewContainer(redisService).
		Image("redis:7-alpine").
		Command("redis-server").
		Endpoint("backend"),
	)

	app := containerkit.NewCompose(project)
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
	if err := app.Build(ctx, &containerkit.ComposeBuild{Writer: os.Stdout, ServiceNames: []string{appService}}); err != nil {
		log.Fatalf("Build: %v", err)
	}

	fmt.Println("\nStarting stack...")
	if err := app.Up(ctx, &containerkit.Up{Detach: true, RemoveOrphans: true, Writer: os.Stdout}); err != nil {
		log.Fatalf("Up: %v", err)
	}

	fmt.Println("\n--- docker compose ps -a ---")
	if err := app.Ps(ctx, &containerkit.Ps{All: true, Writer: os.Stdout}); err != nil {
		log.Printf("Ps: %v", err)
	}

	fmt.Println("\n--- docker compose ps -a --format json (parsed) ---")
	var psOut bytes.Buffer
	if err := app.Ps(ctx, &containerkit.Ps{All: true, Format: "json", Writer: &psOut}); err != nil {
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
	if err := app.Stop(ctx, &containerkit.Stop{Timeout: &stopTimeout, ServiceNames: []string{webService}}); err != nil {
		log.Printf("Stop: %v", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("\n--- docker compose ps -a (after stop web) ---")
	if err := app.Ps(ctx, &containerkit.Ps{All: true, Writer: os.Stdout}); err != nil {
		log.Printf("Ps: %v", err)
	}

	fmt.Println("\nStarting web...")
	if err := app.Start(ctx, &containerkit.Start{ServiceNames: []string{webService}}); err != nil {
		log.Printf("Start: %v", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Println("\nRestarting redis...")
	restartTimeout := 5
	if err := app.Restart(ctx, &containerkit.Restart{NoDeps: true, Timeout: &restartTimeout, ServiceNames: []string{redisService}}); err != nil {
		log.Printf("Restart: %v", err)
	}

	fmt.Println("\n--- docker compose exec (non-interactive) ---")
	if err := app.Exec(ctx, &containerkit.Exec{
		Service: appService,
		Command: []string{"sh", "-c", "echo hello from compose exec"},
		NoTTY:   true,
		Writer:  os.Stdout,
	}); err != nil {
		log.Printf("Exec: %v", err)
	}

	fmt.Println("\n--- docker compose config --services ---")
	if err := app.Config(ctx, &containerkit.ComposeConfig{PrintServices: true, Writer: os.Stdout}); err != nil {
		log.Printf("Config: %v", err)
	}

	fmt.Println("\n--- docker compose images ---")
	if err := app.Images(ctx, &containerkit.Images{Writer: os.Stdout}); err != nil {
		log.Printf("Images: %v", err)
	}

	fmt.Println("\n--- docker compose top (app) ---")
	if err := app.Top(ctx, &containerkit.Top{ServiceNames: []string{appService}, Writer: os.Stdout}); err != nil {
		log.Printf("Top: %v", err)
	}

	fmt.Println("\n--- docker compose pause / unpause (web) ---")
	if err := app.Pause(ctx, &containerkit.Pause{ServiceNames: []string{webService}, Writer: os.Stdout}); err != nil {
		log.Printf("Pause: %v", err)
	}
	if err := app.Unpause(ctx, &containerkit.Unpause{ServiceNames: []string{webService}, Writer: os.Stdout}); err != nil {
		log.Printf("Unpause: %v", err)
	}

	fmt.Println("\n--- docker compose run (one-off) ---")
	if err := app.Run(ctx, &containerkit.Run{
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
		if err := app.Down(context.Background(), &containerkit.Down{RemoveOrphans: true, Writer: os.Stdout}); err != nil {
			log.Printf("Down: %v", err)
		}
	}()

	fmt.Println("\nStack running (Build, Ps, Stop, Start, Restart, Exec, Config, Images, Top, Pause, Run demonstrated). Press Ctrl+C to stop and remove.")
	<-ctx.Done()
}
