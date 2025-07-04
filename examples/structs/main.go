// this example shows how to use the structs to create a service if you prefer to use them
package main

import (
	"log"
	"runtime"
	"time"

	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/tools"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	architecture = runtime.GOARCH
)

func main() {
	project := create.NewProject("my-project")
	// Service created entirely using Docker SDK native structs and then mutated using go-contain option setters
	alpine := MyAlpineBaseService(architecture)
	alpine.WithContainerConfig(
		WithConfigOverride("http://localhost:8080", "8080"),
	)
	project.WithService("my-alpine-service", alpine)
	err := project.Export("./examples/structs/docker-compose.yml", 0644)
	if err != nil {
		log.Fatalf("Error exporting project: %v", err)
	}

}

func WithConfigOverride(healthCheck string, port string) create.SetContainerConfig {
	return tools.Group(
		cc.WithHealthCheck(
			// append to the health check test
			health.WithTest(healthCheck),
		),
		cc.WithExposedPort("tcp", port),
	)
}

func MyAlpineBaseService(architecture string) *create.Container {
	return &create.Container{
		Config: &create.MergedConfig{
			Container: &container.Config{
				Image: "alpine",
				Cmd:   []string{"tail", "-f", "/dev/null"},
				Env: []string{
					"ENV1=value1",
					"ENV2=value2",
				},
				Healthcheck: &container.HealthConfig{
					Test:     []string{"CMD", "curl", "-f"},
					Interval: 10 * time.Second,
					Timeout:  5 * time.Second,
					Retries:  3,
				},
			},
			Host: &container.HostConfig{
				PortBindings: nat.PortMap{
					"8080/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "8080",
						},
					},
				},
				LogConfig: container.LogConfig{
					Type: "json-file",
					Config: map[string]string{
						"max-file": "3",
						"max-size": "10m",
					},
				},
			},
			Network: &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					"my-network": {
						Aliases: []string{"my-alpine-service"},
					},
				},
			},
			Platform: &ocispec.Platform{
				Architecture: architecture,
			},
		},
	}

}
