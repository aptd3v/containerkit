# go-contain

>**go-contain** brings declarative, dynamic Docker Compose to Go. Programmatically define, extend, and orchestrate multi-container environments with the flexibility of code, all while staying fully compatible with existing YAML workflows.

![Go Version](https://img.shields.io/badge/go-1.23.0-blue)
[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/go-contain.svg)](https://pkg.go.dev/github.com/aptd3v/go-contain)
[![Go Report Card](https://goreportcard.com/badge/github.com/aptd3v/go-contain)](https://goreportcard.com/report/github.com/aptd3v/go-contain)

## Features

* Support for Docker Compose commands `up`, `down`, `logs`, (more coming soon!)
* Declarative container/service creation with chainable options
* Native Go option setters for containers, networks, volumes, and health checks etc.
* IDE-friendly
* Designed for automation, CI/CD pipelines, and advanced dev environments

---

## Why go-contain?

While Docker Compose YAML files work great for simple, static configurations, **go-contain** unlocks the full power of programmatic infrastructure definition. Here's why you might choose go-contain over traditional approaches:

### **Programmatic Infrastructure Control**
```go
// Generate infrastructure from data, APIs, configs - A real pain with static YAML

//// Generate a unique environment for each microservice from a config object.
func setupEnvironment(envConfig EnvironmentConfig) *create.Project {
    project := create.NewProject(envConfig.Name)
    
    // Generate services from database records, API responses, etc.
    for _, service := range envConfig.Services {
        replicas := envConfig.GetReplicas(service.Name)
        
        for i := 0; i < replicas; i++ {
            project.WithService(fmt.Sprintf("%s-%d", service.Name, i),
                create.NewContainer(service.Name).
                    WithContainerConfig(
                        cc.WithImagef("%s:%s", service.Image, envConfig.Version),
                        cc.WithEnv("INSTANCE_ID", strconv.Itoa(i)),
                        cc.WithEnv("ENVIRONMENT", envConfig.Environment),
                    ).
                    WithHostConfig(
                        hc.WithPortBindings("tcp", "0.0.0.0", 
                            strconv.Itoa(8080+i), "8080"),
                    ),
            )
        }
    }
    return project
}

// Call with live data from your application
envConfig := fetchEnvironmentFromAPI()
project := setupEnvironment(envConfig)
compose.NewCompose(project).Up(context.Background())
```
```yaml
# Docker Compose scaling creates IDENTICAL containers - no per-instance customization
version: '3.8'
services:
  api:
    image: myapp:v1.2.3
    environment:
      - ENVIRONMENT=production
      # All scaled instances get the SAME environment variables
      # No way to give each replica different INSTANCE_ID or ports
    ports:
      - "8080:8080"  # Port conflicts when scaling!

# docker compose up --scale api=3
# ↑ Creates 3 identical containers, but:
# - All have the same environment variables
# - Port binding conflicts (all try to bind to 8080)
# - No way to customize individual instances
```

### **Dynamic & Conditional Configuration**
```go
// Environment-based logic, loops, and conditionals
for _, env := range []string{"dev", "staging", "prod"} {
    project.WithService(fmt.Sprintf("api-%s", env),
    	create.NewContainer().
            WithContainerConfig(
                cc.WithImagef("myapp:%s", env),
                tools.WhenTrue(env == "prod", 
                    cc.WithEnv("CACHE_ENABLED", "true"),
                ),
            ),
    )
}

// Static configuration requires multiple files or templating
// No native support for conditionals or loops
```

###  **Code Reusability & Composition**
```go
// Create reusable components and patterns
func DatabaseContainer(name, version string)  *create.Container {
    return create.NewContainer().
        WithContainerConfig(
            cc.WithImagef("postgres:%s", version),
            cc.WithEnv("POSTGRES_DB", name),
        ).
        WithHostConfig(
            hc.WithPortBindings("tcp", "0.0.0.0", "5432", "5432"),
        )
}

func RedisContainer() *create.Container {
    return create.NewContainer().
        WithContainerConfig(
            cc.WithImage("redis:7-alpine"),
        ).
        WithHostConfig(
            hc.WithPortBindings("tcp", "0.0.0.0", "6379", "6379"),
        )
}

// Microservices architecture - each service gets its own database
project.WithService("user-service-db", DatabaseContainer("users", "latest"))
project.WithService("user-service-cache", RedisContainer())
project.WithService("order-service-db", DatabaseContainer("orders", "latest"))
project.WithService("order-service-cache", RedisContainer())
```

### **Perfect for Automation & CI/CD**
```go
// Integrate with existing Go tools and workflows
func DeployEnvironment(ctx context.Context, env string, replicas int) error {
    project := create.NewProject(fmt.Sprintf("app-%s", env))
    
    // Build services programmatically based on parameters
    for i := 0; i < replicas; i++ {
        project.WithService(fmt.Sprintf("worker-%d", i), 
            // ... configure based on env and replica count
        )
    }
    
    compose := compose.NewCompose(project)
    return compose.Up(ctx, up.WithDetach())
}
```

### Portable Container Configuration
In `go-contain` the underlying docker sdk is also wrapped as well, allowing you to use the same configuration for docker client control, and compose.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aptd3v/go-contain/pkg/client"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	// Reuse the same config to create a container using the Docker SDK
	resp, err := cli.ContainerCreate(
		context.Background(),
		MySimpleContainer("latest"),
	)
	if err != nil {
		log.Fatalf("Error creating container: %v", err)
	}
	fmt.Println(resp.ID) //container id

	//create a compose project with the same container configuration
	project := create.NewProject("my-project")

	project.WithService("simple-service", MySimpleContainer("latest"))

	err = project.Export("./docker-compose.yml", 0644)
	if err != nil {
		log.Fatalf("Error exporting project: %v", err)
	}

}

func MySimpleContainer(tag string) *create.Container {
	return create.NewContainer().
		WithContainerConfig(
			cc.WithImagef("alpine:%s", tag),
			cc.WithCommand("echo", "hello world"),
		)
}

```


### **Leverage Go's Ecosystem**
- **Testing**: Write unit tests for your infrastructure code
- **Debugging**: Use Go's debugging tools and error handling
- **Libraries**: Integrate with any Go package (HTTP clients, databases, etc.)
- **Tooling**: Build CLIs, APIs, and automation around your containers

### **Still Docker Compose Compatible**
```go
// Export to standard YAML when needed
if err := project.Export("./docker-compose.yaml", 0644); err != nil {
    log.Fatal(err)
}
// Now use with: docker compose up -d
```

**go-contain** gives you the best of both worlds: the flexibility and power of Go programming with full compatibility with the Docker Compose ecosystem.

---

## Prerequisites

- **Go**: 1.23+
- **Docker**: 28.2.0+ with Docker Compose v2.37.0
- **Operating System**: Linux, macOS, or Windows

---

## Installation

```bash
go get github.com/aptd3v/go-contain@latest
```

---

## Quick Start

Get up and running in 30 seconds:

```bash
# Create a new Go module
mkdir my-containers && cd my-containers
go mod init my-containers
go get github.com/aptd3v/go-contain@latest
```
## Create main.go
```go
package main

import (
	"context"
	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
)

func main() {
	project := create.NewProject("hello-world")
	project.WithService("hello-service", 
		create.NewContainer().
			WithContainerConfig(
				cc.WithImage("alpine:latest"),
				cc.WithCommand("echo", "Hello from go-contain!"),
			),
	)
	
	compose.NewCompose(project).Up(context.Background())
}
```
## Run it
```bash
go run main.go
```

## Basic Usage

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
)

func main() {

	project := create.NewProject("my-app")

	project.WithService("my-service",
		create.NewContainer("my-container").
			WithContainerConfig(
				cc.WithImage("alpine:latest"),
				cc.WithCommand("tail", "-f", "/dev/null"),
			),
	)
	//export yaml if desired. (not needed)
	if err := project.Export("./docker-compose.yaml", 0644); err != nil {
		log.Fatal(err)
	}
	//create a new compose instance
	compose := compose.NewCompose(project)

	//execute the up command
	err := compose.Up(
		context.Background(),
		up.WithWriter(os.Stdout),
		up.WithRemoveOrphans(),
		up.WithDetach(),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

---

## Declarative Container Configuration

Each setter type is defined in its own package

```go
project.WithService("api",
	create.NewContainer("my-api-container").
        WithContainerConfig(
			//cc == container config
            cc.WithImagef("ubuntu:%s", tag)
        ).
        WithHostConfig(
			// hc == host config
            hc.WithPortBindings("tcp", "0.0.0.0", "8080", "80"),
        ).
        WithNetworkConfig(
			// nc == network config
            nc.WithEndpoint("my-network"),
        ).
        WithPlatformConfig(
			//pc == platform config
            pc.WithArchitecture("amd64"),
        ),
)
```
## Or use underlying docker SDK structs if desired

Check out [`examples/structs`](./examples/structs) to see how using both can be useful.
```go
project.WithService("api", &create.Container{
		Config: &create.MergedConfig{
			Container: &container.Config{
				Image: fmt.Sprintf("ubuntu:%s", tag),
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
			},
			Network: &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{
					"my-network": {
						Aliases: []string{"my-api-container"},
					}},
			},
			Platform: &ocispec.Platform{
				Architecture: "amd64",
			},
		},
	})
```
---

## tools Package: Declarative Logic for Setters

The `tools` package provides composable helpers for conditional configuration. These are useful when flags, environment variables, or dynamic inputs control what options get applied.

### Highlights

* `tools.WhenTrue(...)` – Apply setters only if a boolean is true
* `tools.WhenTrueFn(...)` – Like above, but accepts  predicate closure `func() bool`
* `tools.OnlyIf(...)` – Apply a setter only if a runtime check passes `func () (bool, error)`
* `tools.Group(...)` – Combine multiple setters into one `func[T any, O ~func(T) error](fns ...O) O`
* `tools.And(...)`, `tools.Or(...)` – Compose multiple predicate closures

### Example

```go
package main

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
	"github.com/aptd3v/go-contain/pkg/create/config/nc"
	"github.com/aptd3v/go-contain/pkg/create/config/sc"
	"github.com/aptd3v/go-contain/pkg/tools"
)

func main() {
	enableDebug := true //imagine this is a flag from your cli or something
	isLinux := runtime.GOOS == "linux"
	project := create.NewProject("conditional-env")
	envVars := tools.Group(
		cc.WithEnv("MYSQL_ROOT_PASSWORD", "password"),
		cc.WithEnv("MYSQL_DATABASE", "mydb"),
		cc.WithEnv("MYSQL_USER", "myuser"),
		cc.WithEnv("MYSQL_PASSWORD", "mypassword"),

		tools.WhenTrueFn(
			tools.Or(enableDebug, os.Getenv("NODE_ENV") == "development"),
			cc.WithEnv("DEBUG", "true"),
		),
	)

	project.WithService("express",
		create.NewContainer().
			WithContainerConfig(
				cc.WithImage("node:latest"),
				cc.WithCommand("npm", "start"),
				envVars,
			).
			WithHostConfig(
				tools.WhenTrueElse(isLinux,//if
					hc.WithRWNamedVolumeMount("node-data", "/app"),//true 
					hc.WithVolumeBinds("./:/app/:rw"),//else 
				),
			).
			WithNetworkConfig(
				nc.WithEndpoint("express-network"),
			),
		// service level configuration
		tools.OnlyIf(EnvFileExists(".ThisFileDoesNotExist.env"),
			sc.WithEnvFile(".ThisFileDoesNotExist.env"),
		),
	).
		WithNetwork("express-network").
		WithVolume("node-data")

	compose := compose.NewCompose(project)

	if err := compose.Up(context.Background()); err != nil {
		// will output .ThisFileDoesNotExist.env: no such file or directory
		log.Fatal(err)

	}
}
// CheckClosure is just a func() (bool, error)
func EnvFileExists(name string) tools.CheckClosure {
	return func() (bool, error) {
		_, err := os.Stat(name)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}
```



### Examples

Explore examples in the [`examples/`](./examples) directory:



### Getting Help

- **API Docs**: [pkg.go.dev/github.com/aptd3v/go-contain](https://pkg.go.dev/github.com/aptd3v/go-contain)
- **Issues**: [GitHub Issues](https://github.com/aptd3v/go-contain/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aptd3v/go-contain/discussions)


## Roadmap

### Current Features 
- ✅ Core Compose commands: `up`, `down`, `logs`
- ✅ Container, network, and volume service configuration 
- ✅ Conditional logic with `tools` package
- ✅ YAML export for compatibility

### In Development  
- Additional Compose commands: `restart`, `stop`, `start`, `ps`
- Enhanced Docker SDK client features
- Image registry authentication helpers
- More comprehensive test coverage

### Ideas & Suggestions
Have ideas for go-contain? We'd love to hear them! Open an [issue](https://github.com/aptd3v/go-contain/issues) or start a [discussion](https://github.com/aptd3v/go-contain/discussions).


### License

MIT License. See [LICENSE](./LICENSE) for details.


### Contributions

Contributions, feedback, and issues are welcome! Fork the repo and submit a PR or open an issue with your idea.