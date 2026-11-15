package kit_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aptd3v/containerkit/pkg/kit"
)

func Example() {
	container := kit.NewContainer(
		kit.Name("worker"),
		kit.Image("alpine:latest"),
	)

	myService := kit.NewService(
		kit.NewContainer(
			kit.Name("foo-server"),
			kit.Image("alpine:latest"),
		),
	)

	compose := kit.NewCompose("app").
		WithService(myService).
		WithService(kit.NewService(
			kit.Name("api"),
			kit.Image("alpine:latest"),
			kit.Publish(
				kit.Protocol(kit.TCP),
				kit.HostIP("0.0.0.0"),
				kit.HostPort(8080),
				kit.ContainerPort(80),
			),
		))

	fmt.Println(container.Name(), container.Image())
	fmt.Println(myService.Name(), myService.Image())
	fmt.Println(compose.Name(), compose.Services()[0].Name(), compose.Services()[1].Name())

	var buf bytes.Buffer
	if err := compose.Write(kit.Export{Writer: &buf}); err != nil {
		panic(err)
	}
	fmt.Println(bytes.Contains(buf.Bytes(), []byte("alpine:latest")))
	// Output:
	// worker alpine:latest
	// foo-server alpine:latest
	// app foo-server api
	// true
}

func ExampleImportYAML() {
	yml := []byte(`
name: app
services:
  api:
    image: alpine:latest
`)
	compose := kit.ImportYAML(yml)
	fmt.Println(compose.Name(), compose.Services()[0].Name(), compose.Services()[0].Image())
	// Output:
	// app api alpine:latest
}

func ExampleDockerfile() {
	df := kit.NewDockerfile().
		From("alpine", "latest").
		Run("apk add --no-cache curl").
		Command("sleep", "infinity")

	compose := kit.NewCompose("app").WithService(
		kit.NewService(kit.Name("api"), kit.Build(df)),
	)
	yml, err := compose.ExportYAML()
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(string(yml), "dockerfile_inline:"))
	fmt.Println(strings.Contains(string(yml), "FROM alpine:latest"))
	// Output:
	// true
	// true
}

func ExampleConnect() {
	cli := kit.NewClient(nil, nil)
	_ = cli.Engine()
	_ = cli.Composer()
}

func ExampleLogger() {
	var buf bytes.Buffer
	log := kit.NewLogger(kit.OutputStream(&buf), kit.Color(false), kit.Progress("plain"))
	log.Emit(kit.LogEvent{Kind: kit.KindPull, Op: "downloading", ID: "abc123", Current: 512, Total: 1024, Status: kit.StatusWorking})
	log.Emit(kit.LogEvent{Kind: kit.KindPull, Op: "pulled", ID: "alpine:latest", Status: kit.StatusDone})
	log.Emit(kit.LogEvent{Kind: kit.KindBuild, Op: "step", ID: "app", Detail: "FROM alpine", Status: kit.StatusWorking})
	log.Emit(kit.LogEvent{Kind: kit.KindCompose, Op: "up", ID: "app", Status: kit.StatusDone})
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		fmt.Println(strings.Contains(line, "pull") || strings.Contains(line, "build") || strings.Contains(line, "compose"))
	}
	// Output:
	// true
	// true
	// true
	// true
}
