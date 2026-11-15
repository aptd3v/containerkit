// Build, up, ps, stop, start, restart, exec, images, top, pause, run, then down.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

const (
	projectName  = "compose-commands-example"
	appService   = "app"
	webService   = "web"
	redisService = "redis"
)

func main() {
	app := NewCompose(projectName, AddNetwork(Name("backend"), Driver("bridge"))).
		WithService(NewService(
			Name(appService),
			Build(inlineDockerfile("alpine", "latest")),
			Network(Name("backend")),
		)).
		WithService(NewService(
			Name(webService),
			Image("nginx:alpine"),
			Command("nginx", "-g", "daemon off;"),
			Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(9080), ContainerPort(80)),
			Network(Name("backend")),
		)).
		WithService(NewService(
			Name(redisService),
			Image("redis:7-alpine"),
			Command("redis-server"),
			Network(Name("backend")),
		))

	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()
	defer func() {
		fmt.Println("\nTearing down...")
		if err := app.Down(context.Background(), cli.Composer(), RemoveOrphans(true), Log(NewLogger(OutputStream(os.Stdout)))); err != nil {
			log.Printf("Down: %v", err)
		}
	}()

	api := cli.Composer()
	fmt.Println("Building app service...")
	if err := app.Build(ctx, api, Services(appService), OutputStream(os.Stdout)); err != nil {
		log.Fatalf("Build: %v", err)
	}

	fmt.Println("\nStarting stack...")
	if err := app.Up(ctx, api, Detach(true), RemoveOrphans(true), OutputStream(os.Stdout)); err != nil {
		log.Fatalf("Up: %v", err)
	}

	fmt.Println("\n--- docker compose ps -a ---")
	printPs(ctx, app, api)

	fmt.Println("\n--- docker compose ps -a (json) ---")
	printPsJSON(ctx, app, api)

	fmt.Println("\nStopping web...")
	if err := app.Stop(ctx, api, Services(webService), Timeout(5*time.Second)); err != nil {
		log.Printf("Stop: %v", err)
	}
	time.Sleep(time.Second)

	fmt.Println("\n--- docker compose ps -a (after stop web) ---")
	printPs(ctx, app, api)

	fmt.Println("\nStarting web...")
	if err := app.Start(ctx, api, Services(webService)); err != nil {
		log.Printf("Start: %v", err)
	}
	time.Sleep(time.Second)

	fmt.Println("\nRestarting redis...")
	if err := app.Restart(ctx, api, Services(redisService), NoDeps(true), Timeout(5*time.Second)); err != nil {
		log.Printf("Restart: %v", err)
	}

	fmt.Println("\n--- docker compose exec (non-interactive) ---")
	code, err := app.Exec(ctx, api, appService, Command("sh", "-c", "echo hello from compose exec"), Tty(false))
	if err != nil {
		log.Printf("Exec: %v", err)
	} else {
		fmt.Printf("exec exit %d\n", code)
	}

	fmt.Println("\n--- services ---")
	for _, svc := range app.Services() {
		fmt.Println(svc.Name())
	}

	fmt.Println("\n--- docker compose images ---")
	images, err := app.Images(ctx, api)
	if err != nil {
		log.Printf("Images: %v", err)
	} else {
		for name, image := range images {
			fmt.Printf("%s  %s:%s  %s\n", name, image.Repository, image.Tag, image.ID)
		}
	}

	fmt.Println("\n--- docker compose top (app) ---")
	tops, err := app.Top(ctx, api, appService)
	if err != nil {
		log.Printf("Top: %v", err)
	} else {
		for _, top := range tops {
			fmt.Printf("%s %s\n", top.Service, top.Name)
			fmt.Println(top.Titles)
			for _, proc := range top.Processes {
				fmt.Println(proc)
			}
		}
	}

	fmt.Println("\n--- docker compose pause / unpause (web) ---")
	if err := app.Pause(ctx, api, Services(webService)); err != nil {
		log.Printf("Pause: %v", err)
	}
	if err := app.Unpause(ctx, api, Services(webService)); err != nil {
		log.Printf("Unpause: %v", err)
	}

	fmt.Println("\n--- docker compose run (one-off) ---")
	code, err = app.Run(ctx, api, appService,
		Command("echo", "hello from compose run"),
		Tty(false),
		AutoRemove(true),
		NoDeps(true),
	)
	if err != nil {
		log.Printf("Run: %v", err)
	} else {
		fmt.Printf("run exit %d\n", code)
	}

	fmt.Println("\nStack running. Press Ctrl+C to stop and remove.")
	<-ctx.Done()
}

func inlineDockerfile(image, tag string) *Dockerfile {
	return NewDockerfile().
		From(image, tag).
		Workdir("/app").
		Run(`echo "Hello, World!"`).
		Command("tail", "-f", "/dev/null")
}

func printPs(ctx context.Context, app *Compose, api Composer) {
	rows, err := app.Ps(ctx, api, All(true))
	if err != nil {
		log.Printf("Ps: %v", err)
		return
	}
	fmt.Println("SERVICE  STATE  NAME  ID")
	for _, row := range rows {
		fmt.Printf("%s  %s  %s  %s\n", row.Service, row.State, row.Name, row.ID)
	}
}

func printPsJSON(ctx context.Context, app *Compose, api Composer) {
	rows, err := app.Ps(ctx, api, All(true))
	if err != nil {
		log.Printf("Ps: %v", err)
		return
	}
	out, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		log.Printf("Ps json: %v", err)
		return
	}
	fmt.Println(string(out))
}
