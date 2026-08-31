// this example shows how to use the structs to create a service if you prefer to use them
package main

import (
	"log"
	"runtime"
	"time"

	"github.com/aptd3v/containerkit/pkg/containerkit"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	architecture = runtime.GOARCH
)

func main() {
	project := containerkit.NewProject("my-project")
	// Service created entirely using Docker SDK native structs and then mutated using containerkit methods
	alpine := MyAlpineBaseService(architecture)
	alpine.ExposedPort("tcp", "8080")
	if alpine.Config != nil && alpine.Config.Container != nil && alpine.Config.Container.Healthcheck != nil {
		alpine.Config.Container.Healthcheck.Test = append(alpine.Config.Container.Healthcheck.Test, "http://localhost:8080")
	}
	project.WithService("my-alpine-service", alpine)
	err := project.Export("./examples/structs/docker-compose.yml", 0644)
	if err != nil {
		log.Fatalf("Error exporting project: %v", err)
	}

}

func MyAlpineBaseService(architecture string) *containerkit.Container {
	return &containerkit.Container{
		Config: &containerkit.MergedConfig{
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
