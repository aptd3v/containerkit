package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aptd3v/go-contain/pkg/client"
	"github.com/aptd3v/go-contain/pkg/client/options/image/build"
	"github.com/aptd3v/go-contain/pkg/compose"
	"github.com/aptd3v/go-contain/pkg/compose/options/up"
	"github.com/aptd3v/go-contain/pkg/create"
	"github.com/aptd3v/go-contain/pkg/create/config/cc"
	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/create/config/hc"
)

func main() {
	df, err := WithDockerContext("./examples/nginx")
	if err != nil {
		log.Fatal(err)
	}
	cli, err := client.NewClient(client.FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	resp, err := cli.ImageBuild(context.Background(), df, build.WithTags("nginx-example:latest"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	project := create.NewProject("nginx-example")
	project.WithService("nginx", create.NewContainer().
		With(
			cc.WithImage("nginx-example:latest"),
			cc.WithCommand("nginx", "-g", "daemon off;"),
			hc.WithPortBindings("tcp", "0.0.0.0", "8080", "80"),
			cc.WithHealthCheck(
				health.WithTest("CMD-SHELL", "curl -f http://localhost:80 || exit 1"),
				health.WithInterval("10s"),
				health.WithTimeout("5s"),
				health.WithStartPeriod("0s"),
				health.WithRetries(3),
			),
			hc.WithMemoryLimit("100MiB"),
		),
	)
	app := compose.NewCompose(project)
	err = app.Up(context.Background(), up.WithDetach(), up.WithRemoveOrphans())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("app started at http://localhost:8080")
}
func WithDockerContext(path string) (io.Reader, error) {
	df := create.NewDockerFile()
	df.From("nginx", "latest")
	df.Copy("nginx.conf", "/etc/nginx/nginx.conf")
	df.Copy("index.html", "/usr/share/nginx/html/index.html")

	return df.NewLocalBuildContext(path)
}
