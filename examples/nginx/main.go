package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

func main() {
	df, err := WithDockerContext("./examples/nginx")
	if err != nil {
		log.Fatal(err)
	}
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	buildOpt := &client.ImageBuild{}
	buildOpt.Tags = []string{"nginx-example:latest"}
	resp, err := cli.ImageBuild(context.Background(), df, buildOpt)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	project := containerkit.NewProject("nginx-example")
	project.WithService("nginx", containerkit.NewContainer().
		Image("nginx-example:latest").
		Command("nginx", "-g", "daemon off;").
		PortBindings("tcp", "0.0.0.0", "8080", "80").
		HealthCheck(containerkit.Health{
			Test:         []string{"CMD-SHELL", "curl -f http://localhost:80 || exit 1"},
			IntervalD:    "10s",
			TimeoutD:     "5s",
			StartPeriodD: "0s",
			Retries:      3,
		}).
		MemoryLimitString("100MiB"),
	)
	app := containerkit.NewCompose(project)
	err = app.Up(context.Background(), &containerkit.Up{Detach: true, RemoveOrphans: true})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("app started at http://localhost:8080")
}
func WithDockerContext(path string) (io.Reader, error) {
	df := containerkit.NewDockerFile()
	df.From("nginx", "latest")
	df.Copy("nginx.conf", "/etc/nginx/nginx.conf")
	df.Copy("index.html", "/usr/share/nginx/html/index.html")

	return df.NewLocalBuildContext(path)
}
