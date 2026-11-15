// Create an Alpine container and attach an interactive shell to it.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	. "github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/terminal"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctr := NewContainer(
		Name("exec-example"),
		Image("alpine:latest"),
		Command("tail", "-f", "/dev/null"),
	)
	cli, err := Connect(FromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()
	defer cleanup(cli, "exec-example")

	if _, err := InspectImage(ctx, cli.Engine(), "alpine:latest"); err != nil {
		if !IsNotFound(err) {
			return err
		}
		fmt.Fprintln(os.Stderr, "pulling alpine:latest")
		if err := PullImage(ctx, cli.Engine(), "alpine:latest", Platform(fmt.Sprintf("linux/%s", runtime.GOARCH))); err != nil {
			return err
		}
	}

	created, err := ctr.Create(ctx, cli.Engine())
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	if err := ctr.Start(ctx, cli.Engine(), created.ID); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	exec, err := CreateExec(ctx, cli.Engine(), created.ID,
		Tty(true),
		AttachStdin(true),
		AttachStdout(true),
		AttachStderr(true),
		ConsoleSize(24, 80),
		Command("/bin/sh"),
	)
	if err != nil {
		return err
	}
	attached, err := AttachExec(ctx, cli.Engine(), exec.ID, Tty(true), ConsoleSize(24, 80))
	if err != nil {
		return err
	}
	session, err := terminal.NewSession(os.Stdin, os.Stdout, attached.Conn, attached.Reader)
	if err != nil {
		attached.Close()
		return err
	}
	defer session.Close()

	err = session.Start(ctx, func(sz terminal.Size) error {
		return ResizeExecTerminal(ctx, cli.Engine(), exec.ID, sz)
	})
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
