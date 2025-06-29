# go-contain

>**go-contain** brings declarative, dynamic Docker Compose to Go — programmatically define, extend, and orchestrate multi-container environments with the flexibility of code, all while staying fully compatible with existing YAML workflows.

![Go Version](https://img.shields.io/badge/go-1.23.0-blue)
[![Go Reference](https://pkg.go.dev/badge/github.com/aptd3v/go-contain.svg)](https://pkg.go.dev/github.com/aptd3v/go-contain)

## 🚀 Features

* Support for Docker Compose commands `up`, `down`, `logs`, (more coming soon!)
* Declarative container/service creation with chainable options
* Native Go option setters for containers, networks, volumes, and health checks etc.
* IDE-friendly
* Designed for automation, CI/CD pipelines, and advanced dev environments

---

## 📦 Installation

```bash
go get github.com/aptd3v/go-contain@latest
```

---

## 🛠️ Basic Usage

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

## 🔧 Declarative Container Configuration

Each setter type is defined in its own package

```go
project.WithService("api",
	create.NewContainer("my-api-container").
        WithContainerConfig(
            cc.WithImagef("ubuntu:%s", tag)
        ).
        WithHostConfig(
            hc.WithPortBindings("tcp", "0.0.0.0", "8080", "80"),
        ).
        WithNetworkConfig(
            nc.WithEndpoint("my-network"),
        ).
        WithPlatformConfig(
            pc.WithArchitecture("amd64"),
        ),
)
```

---

## 🧰 tools Package: Declarative Logic for Setters

The `tools` package provides composable helpers for conditional configuration. These are useful when flags, environment variables, or dynamic inputs control what options get applied.

### ✅ Highlights

* `tools.WhenTrue(...)` – Apply setters only if a boolean is true
* `tools.WhenTrueFn(...)` – Like above, but accepts  predicate closure `func() bool`
* `tools.OnlyIf(...)` – Apply a setter only if a runtime check passes `func () (bool, error)`
* `tools.Group(...)` – Combine multiple setters into one `func[T any, O ~func(T) error](fns ...O) O`
* `tools.And(...)`, `tools.Or(...)` – Compose multiple predicate closures

### 📦 Example

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
		create.NewContainer("node:latest").
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
		// will output ThisFileDoesNotExist.env: no such file or directory
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

---

## 🧪 Advanced Patterns

* Programmatically build services, networks, and volumes using loops
* Reuse options via functional composition
* Create declarative DSLs for internal infrastructure automation

---

## 📁 Project Structure (Current)

```bash
├── examples
│   └── ... #examples
├── go.mod
├── go.sum
├── LICENSE
├── main.go
├── pkg
│   ├── client # docker sdk client wrapper
│   │   ├── auth
│   │   │   └── auth.go # image registry auth helpers (untested for now)
│   │   │
│   │   │
│   │   ├── options # docker sdk client [action] option setters
│   │   │   ├── container
│   │   │   │   └── ... #container option setters
│   │   │   ├── image
│   │   │   │   └── ... #image option setters
│   │   │   ├── network
│   │   │   │   └── ... #network option setters
│   │   │   └── volume
│   │   │       └── ... #volume option setters
│   │   │
│   │   │
│   │   └── response # wrapped client response types
│   │       
│   │   
│   │   
│   ├── compose # compose cli wrapper
│   │   │ 
│   │   └── options 
│   │       └── ... # compose cli option setters
│   │ 
│   │ 
│   ├── create # create container and compose projects/services
│   │   └── config
│   │       ├── cc 
│   │       │   └── ... # container config setters
│   │       │
│   │       ├── hc # container host config setters
│   │       │   └── ...
│   │       │
│   │       ├── nc # container network config setters
│   │       │   └── ...
│   │       │
│   │       ├── pc # container platform config setters
│   │       │   └── ...
│   │       │
│   │       └── sc 
│   │           └── ... # compose service setters
│   │       
│   └── tools # various helpers
└── README.md # this file
```

---

## 📝 YAML Export & Compatibility

**go-contain** lets you export your programmatically defined Compose projects as standard Docker Compose YAML files. These exported YAML files are fully compatible with the traditional Docker Compose CLI and ecosystem.

This means you can:

* Use `docker compose up`, `docker compose down`, and other Docker Compose commands directly on the exported YAML.
* Share the exported YAML with teams or CI pipelines that rely on standard Docker Compose workflows.


Example:

```go
if err := project.Export("./docker-compose.yaml", 0644); err != nil {
	log.Fatal(err)
}
```

You can then run:

```bash
docker compose up -d
```

to start your services exactly as defined by your Go code.

This design ensures maximum flexibility and compatibility, letting you leverage the power of Go while staying aligned with Docker Compose standards.



## 📄 License

MIT License. See [LICENSE](./LICENSE) for details.

---

## 🤝 Contributions

Contributions, feedback, and issues are welcome! Fork the repo and submit a PR or open an issue with your idea.
