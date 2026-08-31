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

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func main() {
	ctx := context.Background()
	alpineContainer := containerkit.NewContainer("exec", "example").
		Image("alpine:latest").
		Command("tail", "-f", "/dev/null")

	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
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

	if res, err := cli.ImagePull(ctx, "alpine:latest", &client.ImagePull{CurrentPlatform: true}); err != nil {
		log.Fatal(err)
	} else if _, err = io.Copy(os.Stdout, res); err != nil {
		log.Fatal(err)
	} else {
		defer res.Close()
	}
	defer cleanup(cli, alpineContainer.Name)

	if _, err := cli.ContainerCreate(ctx, alpineContainer); err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	if err := cli.ContainerStart(ctx, alpineContainer.Name, nil); err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	execCreate, err := cli.ContainerExecCreate(ctx, alpineContainer.Name, &client.Exec{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          true,
		Command:      []string{"/bin/sh"},
	})
	if err != nil {
		log.Fatal(err)
	}

	session, err := cli.ContainerExecAttachTerminal(ctx, execCreate.ID, &client.ExecAttach{Tty: true})
	if err != nil {
		log.Fatal(err)
	}
	monitor := session.MonitorSize()
	go func() {
		for size := range monitor {
			err := cli.ContainerExecResize(ctx, execCreate.ID, &client.Resize{Width: size.Width, Height: size.Height})
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	err = session.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("session closed running cleanup")
	defer session.Close()

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
	os.Exit(0)
}
