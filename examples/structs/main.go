// The same Alpine service the old structs example described, written with
// setters. ExportYAML writes the compose file; nothing talks to the daemon.
package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

func main() {
	app := NewCompose("my-project",
		AddNetwork(Name("my-network"), Driver("bridge")),
	).WithService(NewService(
		Name("my-alpine-service"),
		Image("alpine"),
		Command("tail", "-f", "/dev/null"),
		Env("ENV1", "value1"),
		Env("ENV2", "value2"),
		Platform(fmt.Sprintf("linux/%s", runtime.GOARCH)),
		Publish(
			Protocol(TCP),
			HostIP("0.0.0.0"),
			HostPort(8080),
			ContainerPort(8080),
		),
		Expose(8080, TCP),
		Health(
			Test("CMD", "curl", "-f", "http://localhost:8080"),
			Interval(10*time.Second),
			Timeout(5*time.Second),
			Retries(3),
		),
		Logging(
			Driver(JSONFile),
			DriverOpt("max-file", "3"),
			DriverOpt("max-size", "10m"),
		),
		Network(Name("my-network"), Alias("my-alpine-service")),
	))

	yml, err := app.ExportYAML()
	if err != nil {
		log.Fatalf("export: %v", err)
	}
	if err := os.WriteFile("./examples/structs/docker-compose.yml", yml, 0644); err != nil {
		log.Fatalf("write: %v", err)
	}
}
