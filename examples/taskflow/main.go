// A CI-shaped task graph: build an image, bring a database up, run tests in
// parallel, retry a flaky step, then tag the image only when everything
// passed. Remove downs the database after List, including on failure.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/taskflow"
)

const (
	imageCI      = "ck-taskflow-ci:latest"
	imageRelease = "ck-taskflow-ci:release"
	projectName  = "ck-taskflow-db"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	flow := taskflow.New("ci", cli)
	addTasks(flow, cli, newConfig())

	if err := flow.Dump(os.Stdout); err != nil {
		return err
	}
	defer flow.Remove(context.Background())

	runErr := flow.Run(ctx, &taskflow.Run{
		Writer:  os.Stdout,
		Profile: os.Stdout,
		Trace:   os.Stdout,
		Watch:   watch,
	})

	listed, err := flow.List(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("labeled containers: %d\n", len(listed))
	return runErr
}

// config is the kit values the graph runs. It does not know about task order.
type config struct {
	build       kit.Build
	db          *kit.Compose
	unit        *kit.Container
	integration *kit.Container
	lint        *kit.Container
}

func newConfig() config {
	df := kit.NewDockerFile().
		From("alpine:latest").
		Run("printf 'ok\\n' > /app").
		Command("cat", "/app")

	project := kit.NewProject(projectName)
	project.WithService(kit.NewService("db",
		kit.NewContainer("ck-taskflow-db").
			Image("postgres:16-alpine").
			Env("POSTGRES_PASSWORD", "ci").
			HealthCheck(kit.NewHealth().
				CMD("pg_isready", "-U", "postgres").
				IntervalD("2s").
				TimeoutD("3s").
				Retries(20)),
	))

	test := func(name string, cmd ...string) *kit.Container {
		return kit.NewContainer(name).Image(imageCI).Command(cmd...)
	}
	return config{
		build:       kit.NewBuild().Dockerfile(df),
		db:          kit.NewCompose(project),
		unit:        test("ck-taskflow-unit", "test", "-f", "/app"),
		integration: test("ck-taskflow-integ", "test", "-f", "/app"),
		lint:        test("ck-taskflow-lint", "true"),
	}
}

func addTasks(flow *taskflow.Flow, cli *client.Client, cfg config) {
	build := flow.Build("build", cfg.build, &client.ImageBuild{Tags: []string{imageCI}})
	upDB := flow.Compose("up-db", cfg.db, &kit.Up{Detach: true, Wait: true})
	tests := flow.Subflow("tests", func(sf *taskflow.Subflow) {
		sf.Container("unit", cfg.unit, nil).Priority(taskflow.PriorityHigh)
		sf.Container("integration", cfg.integration, nil).Retry(2)
		sf.Container("lint", cfg.lint, &taskflow.ContainerRun{Logs: os.Stdout}).
			Priority(taskflow.PriorityLow)
	})
	release := flow.If("release", func(context.Context) bool {
		return os.Getenv("RELEASE") == "1"
	})
	tag := flow.Task("tag", func(ctx context.Context) error {
		return cli.ImageTag(ctx, imageCI, imageRelease)
	})

	build.Precede(upDB)
	upDB.Precede(tests)
	tests.Precede(release)
	release.Then(tag)
}

func watch(e taskflow.Event) {
	switch e.Kind {
	case taskflow.EventRetry:
		fmt.Printf("[%s] attempt %d failed: %v\n", e.Name, e.Attempt, e.Err)
	case taskflow.EventSkip, taskflow.EventError:
		fmt.Printf("[%s] %s %v\n", e.Name, e.Kind, e.Err)
	}
}
