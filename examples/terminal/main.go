// Build an image with its own user, start it under a hostname, and attach a shell.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

const (
	terminalImage = "terminal-example:local"
	terminalName  = "exec-example"
	terminalUser  = "dev"
	terminalHost  = "terminal"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cli, err := Connect(FromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()
	defer cleanup(cli, terminalName)

	image := NewDockerfile().
		From("ubuntu:24.04").
		Run("useradd --create-home --shell /bin/bash "+terminalUser).
		User(terminalUser).
		Workdir("/home/"+terminalUser).
		Command("tail", "-f", "/dev/null")
	logger := NewLogger(OutputStream(os.Stdout))
	if err := BuildImage(ctx, cli.Engine(), image, Image(terminalImage), Log(logger)); err != nil {
		return fmt.Errorf("build image: %w", err)
	}

	ctr := NewContainer(
		Name(terminalName),
		Image(terminalImage),
		User(terminalUser),
		Hostname(terminalHost),
	)
	created, err := ctr.Create(ctx, cli.Engine())
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	if err := ctr.Start(ctx, cli.Engine(), created.ID); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	err = StartExecTerminal(ctx, cli.Engine(), created.ID, os.Stdin, os.Stdout,
		User(terminalUser),
		Command("/bin/bash"),
	)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	fmt.Println("session closed running cleanup")
	return nil
}

func cleanup(cli *Client, name string) {
	ctx := context.Background()
	ctr := NewContainer(Name(name))
	if err := ctr.Stop(ctx, cli.Engine(), name); err != nil && !IsNotFound(err) {
		log.Printf("Failed to stop container: %v", err)
	}
	if err := ctr.Remove(ctx, cli.Engine(), name, Force(true)); err != nil && !IsNotFound(err) {
		log.Printf("Failed to remove container: %v", err)
	}
	fmt.Println("Cleanup completed")
}
