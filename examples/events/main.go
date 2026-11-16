// Stream compose events while nginx is started and stopped three times.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

const serviceName = "nginx"

func main() {
	app := NewCompose("events-project").WithService(NewService(
		Name(serviceName),
		Image("nginx:latest"),
		Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(8080), ContainerPort(80)),
		Health(
			Test("CMD-SHELL", "curl -f http://localhost:80 || exit 1"),
			Interval(2*time.Second),
			Timeout(5*time.Second),
			Retries(3),
		),
	))

	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := app.Events(ctx, cli.Composer(), Services(serviceName), EventConsumer(printEvent))
		if err != nil && ctx.Err() == nil {
			log.Printf("events: %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		log.Printf("Cycle %d: Starting container...\n", i+1)
		time.Sleep(time.Second)
		if err := app.Up(context.Background(), cli.Composer(), Detach(true), OutputStream(io.Discard)); err != nil {
			log.Fatalf("up: %v", err)
		}
		time.Sleep(time.Second)
		log.Printf("Cycle %d: Stopping container...\n", i+1)
		time.Sleep(time.Second)
		if err := app.Down(context.Background(), cli.Composer(), RemoveOrphans(true)); err != nil {
			log.Fatalf("down: %v", err)
		}
	}
}

func printEvent(event Event) error {
	fmt.Println("========================================")
	fmt.Printf("Event:\n")
	fmt.Printf("  Status    : %s\n", event.Status)
	fmt.Printf("  Time      : %s\n", event.Timestamp.Format(time.RFC3339))
	fmt.Printf("  Service   : %s\n", event.Service)
	fmt.Printf("  Container : %s\n", event.Container)
	if len(event.Attributes) > 0 {
		fmt.Println("  Attributes:")
		for key, value := range event.Attributes {
			fmt.Printf("    - %s: %s\n", key, value)
		}
	} else {
		fmt.Println("  Attributes: <none>")
	}
	fmt.Println("========================================")
	return nil
}
