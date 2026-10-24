// in this example we will create a container and run a command in it to demonstrate the use of the terminal exec attach
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/client/terminal"
	"github.com/aptd3v/containerkit/pkg/kit"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	alpineContainer := kit.NewContainer("exec", "example").
		Image("alpine:latest").
		Command("tail", "-f", "/dev/null")

	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionFromEnv())
	if err != nil {
		return err
	}
	defer cleanup(cli, alpineContainer.Name)

	if _, err := cli.ImageInspect(ctx, "alpine:latest", nil); err != nil {
		if !client.IsErrNotFound(err) {
			return err
		}
		fmt.Fprintln(os.Stderr, "pulling alpine:latest")
		pulled, err := cli.ImagePull(ctx, "alpine:latest", &client.ImagePull{Platforms: []ocispec.Platform{{OS: "linux", Architecture: runtime.GOARCH}}})
		if err != nil {
			return err
		}
		if _, err := io.Copy(os.Stdout, pulled); err != nil {
			pulled.Close()
			return err
		}
		if err := pulled.Close(); err != nil {
			return err
		}
	}

	if _, err := cli.ContainerCreate(ctx, alpineContainer); err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	if err := cli.ContainerStart(ctx, alpineContainer.Name, nil); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	execCreate, err := cli.ContainerExecCreate(ctx, alpineContainer.Name, &client.Exec{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		TTY:          true,
		Cmd:          []string{"/bin/sh"},
	})
	if err != nil {
		return err
	}

	session, err := cli.ContainerExecAttachTerminal(ctx, execCreate.ID, &client.ExecAttach{TTY: true})
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Start(ctx, func(sz terminal.Size) error {
		return cli.ContainerExecResize(ctx, execCreate.ID, &client.Resize{
			Width:  sz.Width,
			Height: sz.Height,
		})
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	fmt.Println("session closed running cleanup")
	return nil
}

func cleanup(cli *client.Client, cName string) {
	ctx := context.Background()

	if err := cli.ContainerStop(ctx, cName, nil); err != nil {
		log.Printf("Failed to stop container: %v", err)
	}

	if err := cli.ContainerRemove(ctx, cName, &client.Remove{Force: true}); err != nil {
		log.Printf("Failed to remove container: %v", err)
	}

	fmt.Println("Cleanup completed")
}
