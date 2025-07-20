// in this example we will create a container and run a command in it to demonstrate the use of the terminal exec attach
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aptd3v/go-contain/pkg/client"
	"github.com/aptd3v/go-contain/pkg/client/options/container/exec"
	"github.com/aptd3v/go-contain/pkg/client/options/container/execattach"
	"github.com/aptd3v/go-contain/pkg/client/options/container/remove"
	"github.com/aptd3v/go-contain/pkg/client/options/image/pull"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/tools"
)

func main() {
	ctx := context.Background()
	alpineContainer := create.NewContainer("exec", "example")
	alpineContainer.With(
		cc.WithImage("alpine:latest"),
		cc.WithCommand("tail", "-f", "/dev/null"),
	)

	cli, err := client.NewClient(client.FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	// Set up cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(cli, alpineContainer.Name)
	}()

	// Pull Ubuntu image
	if res, err := cli.ImagePull(ctx, "alpine:latest", pull.WithCurrentPlatform()); err != nil {
		log.Fatal(err)
	} else if _, err = io.Copy(os.Stdout, res); err != nil {
		log.Fatal(err)
	} else {
		defer res.Close()
	}
	defer cleanup(cli, alpineContainer.Name)

	// Create and start the container
	if _, err := cli.ContainerCreate(ctx, alpineContainer); err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	if err := cli.ContainerStart(ctx, alpineContainer.Name); err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	execCreate, err := cli.ContainerExecCreate(ctx, alpineContainer.Name, WithExecOptions())
	if err != nil {
		log.Fatal(err)
	}

	session, err := cli.ContainerExecAttachTerminal(ctx, execCreate.ID, execattach.WithTty())
	if err != nil {
		log.Fatal(err)
	}
	err = session.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("session closed running cleanup")
	defer session.Close()

}
func WithExecOptions() exec.SetContainerExecOption {
	return tools.Group(
		exec.WithAttachStderr(),
		exec.WithAttachStdin(),
		exec.WithAttachStdout(),
		exec.WithTty(),
		exec.WithPrivileged(),
		exec.WithCmd("/bin/sh"),
	)
}

func cleanup(client *client.Client, cName string) {
	ctx := context.Background()

	if err := client.ContainerStop(ctx, cName); err != nil {
		log.Printf("Failed to stop container: %v", err)
	}

	if err := client.ContainerRemove(ctx, cName, remove.WithForce()); err != nil {
		log.Printf("Failed to remove container: %v", err)
	}

	fmt.Println("Cleanup completed")
}
