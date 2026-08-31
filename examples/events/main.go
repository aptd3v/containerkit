// This example shows how to use the events API to get real-time updates about the state of a service.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aptd3v/containerkit/pkg/containerkit"
)

const (
	serviceName = "nginx"
)

func main() {
	project := containerkit.NewProject("events-project")
	project.WithService(serviceName, containerkit.NewContainer().
		Image("nginx:latest").
		PortBindings("tcp", "0.0.0.0", "8080", "80").
		HealthCheck(containerkit.Health{
			Test:         []string{"CMD-SHELL", "curl -f http://localhost:80 || exit 1"},
			IntervalD:    "2s",
			TimeoutD:     "5s",
			Retries:      3,
			StartPeriodD: "0s",
		}),
	)

	example := containerkit.NewCompose(project)
	ctx, cancel := context.WithCancel(context.Background())
	events, errCh, err := example.Events(ctx, serviceName)
	if err != nil {
		log.Fatalf("error starting events: %v", err)
	}
	defer cancel()

	go func() {
		for event := range events {
			fmt.Println("========================================")
			fmt.Printf("Event:\n")
			fmt.Printf("  Action  : %s\n", event.Action)
			fmt.Printf("  Time    : %s\n", event.Time)
			fmt.Printf("  Service : %s\n", event.Service)
			fmt.Printf("  ID      : %s\n", event.ID)
			fmt.Printf("  Type    : %s\n", event.Type)

			if len(event.Attributes) > 0 {
				fmt.Println("  Attributes:")
				for key, value := range event.Attributes {
					fmt.Printf("    - %s: %s\n", key, value)
				}
			} else {
				fmt.Println("  Attributes: <none>")
			}
			fmt.Println("========================================")
		}
	}()

	go func() {
		for err := range errCh {
			if err != nil {
				log.Fatalf("Events error: %v", err)
			}
		}
	}()

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("Cycle %d: Starting container...\n", i+1)
		time.Sleep(1 * time.Second)

		err := example.Up(context.Background(), &containerkit.Up{Detach: true, Writer: io.Discard})
		if err != nil {
			log.Fatalf("error executing 'up': %v", err)
		}

		time.Sleep(1 * time.Second)
		log.Printf("Cycle %d: Stopping container...\n", i+1)
		time.Sleep(1 * time.Second)

		err = example.Down(context.Background(), &containerkit.Down{RemoveOrphans: true, Writer: io.Discard})
		if err != nil {
			log.Fatalf("error executing 'down': %v", err)
		}
	}
}
