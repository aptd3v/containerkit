// This example demonstrates the compose CLI commands: Build, Ps, Start, Stop, and Restart.
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

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/build"
	"github.com/aptd3v/go-contain/pkg/compose/options/down"
	"github.com/aptd3v/go-contain/pkg/compose/options/exec"
	"github.com/aptd3v/go-contain/pkg/compose/options/ps"
	"github.com/aptd3v/go-contain/pkg/compose/options/restart"
	"github.com/aptd3v/go-contain/pkg/compose/options/start"
	"github.com/aptd3v/go-contain/pkg/compose/options/stop"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/aptd3v/go-contain/pkg/create/config/nc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
)

const (
	projectName  = "compose-commands-example"
	appService   = "app"
	webService   = "web"
	redisService = "redis"
)

// WithInlineDockerfile returns a string that can be used as an inline dockerfile
// within compose yaml
func WithInlineDockerfile(image, tag string) create.SetServiceConfig {
	df := create.NewDockerFile()
	df.From(image, tag)
	df.Workdir("/app")
	df.Run("echo \"Hello, World!\"")
	df.CommandExec("tail", "-f", "/dev/null")
	return sc.WithBuild(df.WithInline())
}

func main() {
	project := create.NewProject(projectName)
	project.WithNetwork("backend")
	// app is built from the local Dockerfile in this directory
	project.WithService(appService, create.NewContainer(appService),
		WithInlineDockerfile("alpine", "latest"),
	)
	project.WithService(webService, create.NewContainer(webService).
		WithContainerConfig(
			cc.WithImage("nginx:alpine"),
			cc.WithCommand("nginx", "-g", "daemon off;"),
		).
		WithHostConfig(
			hc.WithPortBindings("tcp", "0.0.0.0", "9080", "80"),
		).
		WithNetworkConfig(nc.WithEndpoint("backend")),
	)
	project.WithService(redisService, create.NewContainer(redisService).
		WithContainerConfig(
			cc.WithImage("redis:7-alpine"),
			cc.WithCommand("redis-server"),
		).
		WithNetworkConfig(nc.WithEndpoint("backend")),
	)

	app := compose.NewCompose(project)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Build the app service
	fmt.Println("Building app service...")
	//change the context to the directory of the Dockerfile
	if err := app.Build(ctx, build.WithWriter(os.Stdout), build.WithServiceNames(appService)); err != nil {
		log.Fatalf("Build: %v", err)
	}

	// Up (detached)
	fmt.Println("\nStarting stack...")
	if err := app.Up(ctx, up.WithDetach(), up.WithRemoveOrphans(), up.WithWriter(os.Stdout)); err != nil {
		log.Fatalf("Up: %v", err)
	}

	// Ps — list all containers (table)
	fmt.Println("\n--- docker compose ps -a ---")
	if err := app.Ps(ctx, ps.WithAll(), ps.WithWriter(os.Stdout)); err != nil {
		log.Printf("Ps: %v", err)
	}

	// Ps — same as JSON, capture and parse
	fmt.Println("\n--- docker compose ps -a --format json (parsed) ---")
	var psOut bytes.Buffer
	if err := app.Ps(ctx, ps.WithAll(), ps.WithFormat("json"), ps.WithWriter(&psOut)); err != nil {
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
			// Some versions output one JSON object per line (NDJSON)
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

	// Stop web only
	fmt.Println("\nStopping web...")
	if err := app.Stop(ctx, stop.WithTimeout(5), stop.WithServiceNames(webService)); err != nil {
		log.Printf("Stop: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Ps again
	fmt.Println("\n--- docker compose ps -a (after stop web) ---")
	if err := app.Ps(ctx, ps.WithAll(), ps.WithWriter(os.Stdout)); err != nil {
		log.Printf("Ps: %v", err)
	}

	// Start web
	fmt.Println("\nStarting web...")
	if err := app.Start(ctx, start.WithServiceNames(webService)); err != nil {
		log.Printf("Start: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Restart redis (no deps)
	fmt.Println("\nRestarting redis...")
	if err := app.Restart(ctx, restart.WithNoDeps(), restart.WithTimeout(5), restart.WithServiceNames(redisService)); err != nil {
		log.Printf("Restart: %v", err)
	}

	// Exec — non-interactive: run a command in the app service (no TTY, script-friendly)
	fmt.Println("\n--- docker compose exec (non-interactive) ---")
	if err := app.Exec(ctx,
		exec.WithService(appService),
		exec.WithCommand("sh", "-c", "echo hello from compose exec"),
		exec.WithNoTTY(),
		exec.WithWriter(os.Stdout),
	); err != nil {
		log.Printf("Exec: %v", err)
	}

	// Down on exit
	defer func() {
		fmt.Println("\nTearing down...")
		if err := app.Down(context.Background(), down.WithRemoveOrphans(), down.WithWriter(os.Stdout)); err != nil {
			log.Printf("Down: %v", err)
		}
	}()

	fmt.Println("\nStack running (Build, Ps, Stop, Start, Restart, Exec demonstrated). Press Ctrl+C to stop and remove.")
	<-ctx.Done()
}
