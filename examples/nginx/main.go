// Build nginx-example:latest from the files in this directory, then bring the
// service up on port 8080 with a develop watch.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

func main() {
	df := nginxDockerfile()
	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	ctx := context.Background()
	if err := BuildImage(ctx, cli.Engine(),
		df,
		Source("./examples/nginx"),
		Image("nginx-example:latest"),
	); err != nil {
		log.Fatal(err)
	}

	app := NewCompose("nginx-example").WithService(NewService(
		Name("nginx"),
		Image("nginx-example:latest"),
		Command("nginx", "-g", "daemon off;"),
		Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(8080), ContainerPort(80)),
		Health(
			Test("CMD-SHELL", "curl -f http://localhost:80 || exit 1"),
			Interval(10*time.Second),
			Timeout(5*time.Second),
			Retries(3),
		),
		Memory(100<<20),
		Develop(WatchFiles(
			Action(Sync),
			Path("./src"),
			Target("/app"),
			Ignore("tmp"),
		)),
		Build(Source("./examples/nginx"), df),
	))
	if err := app.Up(ctx, cli.Composer(), Detach(true), RemoveOrphans(true)); err != nil {
		log.Fatal(err)
	}
	fmt.Println("app started at http://localhost:8080")
}

func nginxDockerfile() *Dockerfile {
	return NewDockerfile().
		From("nginx", "latest").
		Copy("nginx.conf", "/etc/nginx/nginx.conf").
		Copy("index.html", "/usr/share/nginx/html/index.html").
		Expose("80").
		Command("nginx", "-g", "daemon off;")
}
