package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

func main() {
	df, err := WithDockerContext("./examples/nginx")
	if err != nil {
		log.Fatal(err)
	}
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionFromEnv())
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
	project := kit.NewProject("nginx-example")
	project.WithService(kit.NewService("nginx", kit.NewContainer().
		Image("nginx-example:latest").
		Command("nginx", "-g", "daemon off;").
		PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "8080", ContainerPort: "80"}).
		HealthCheck(kit.NewHealth().
			CMDShell("curl -f http://localhost:80 || exit 1").
			IntervalD("10s").
			TimeoutD("5s").
			StartPeriodD("0s").
			Retries(3)).
		MemoryLimit(100*types.MiB),
	).Develop(kit.Develop{
		Action: kit.WatchActionSync,
		Path:   "./src",
		Target: "/app",
		Ignore: []string{"tmp"},
	}).Build(kit.NewBuild().
		Context(".").
		DockerfilePath("Dockerfile").
		Dockerfile(kit.NewDockerFile().
			From("nginx:latest").
			Copy("nginx.conf", "/etc/nginx/nginx.conf").
			Copy("index.html", "/usr/share/nginx/html/index.html").
			Expose("80").
			Command("nginx", "-g", "daemon off;"))))
	app := kit.NewCompose(project)
	err = app.Up(context.Background(), &kit.Up{Detach: true, RemoveOrphans: true})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("app started at http://localhost:8080")
}
func WithDockerContext(path string) (io.Reader, error) {
	return kit.NewBuild().
		Context(path).
		Dockerfile(kit.NewDockerFile().
			From("nginx:latest").
			Copy("nginx.conf", "/etc/nginx/nginx.conf").
			Copy("index.html", "/usr/share/nginx/html/index.html")).
		Reader()
}
