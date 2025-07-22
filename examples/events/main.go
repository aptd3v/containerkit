// This example shows how to use the events API to get real-time updates about the state of a service.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/down"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
)

const (
	serviceName = "nginx"
)

func main() {
	project := create.NewProject("events-project")
	project.WithService(serviceName, create.NewContainer().
		With(
			cc.WithImage("nginx:latest"),
			hc.WithPortBindings("tcp", "0.0.0.0", "8080", "80"),
			cc.WithHealthCheck(
				health.WithTest("CMD-SHELL", "curl -f http://localhost:80 || exit 1"),
				health.WithInterval("2s"),
				health.WithTimeout("5s"),
				health.WithRetries(3),
				health.WithStartPeriod("0s"),
			),
		),
	)

	example := compose.NewCompose(project)
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

		err := example.Up(context.Background(), up.WithDetach(), up.WithWriter(io.Discard))
		if err != nil {
			log.Fatalf("error executing 'up': %v", err)
		}

		time.Sleep(1 * time.Second)
		log.Printf("Cycle %d: Stopping container...\n", i+1)
		time.Sleep(1 * time.Second)

		err = example.Down(context.Background(), down.WithRemoveOrphans(), down.WithWriter(io.Discard))
		if err != nil {
			log.Fatalf("error executing 'down': %v", err)
		}
	}
}
