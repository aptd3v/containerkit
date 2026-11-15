// A CI-shaped task graph: build an image, bring a database up, run tests in
// parallel, retry a flaky step, then tag the image only when everything
// passed. Remove downs the database after List, including on failure.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
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
	cli, err := Connect(FromEnv())
	if err != nil {
		return err
	}
	defer cli.Close()

	flow := taskflow.New("ci", cli)
	addTasks(flow, cli)

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
	fmt.Printf("labeled containers: %d\n", len(listed.Items))
	return runErr
}

func addTasks(flow *taskflow.Flow, cli *Client) {
	df := NewDockerfile().
		From("alpine", "latest").
		Run("printf 'ok\\n' > /app").
		Command("cat", "/app")

	db := NewCompose(projectName).WithService(NewService(
		Name("db"),
		ContainerName("ck-taskflow-db"),
		Image("postgres:16-alpine"),
		Env("POSTGRES_PASSWORD", "ci"),
		Health(
			Test("pg_isready", "-U", "postgres"),
			Interval(2*time.Second),
			Timeout(3*time.Second),
			Retries(20),
		),
	))

	build := flow.Build("build", df, Tag(imageCI))
	upDB := flow.Compose("up-db", db, taskflow.Up(Detach(true), Wait(true)))
	tests := flow.Subflow("tests", func(sf *taskflow.Subflow) {
		sf.Container("unit", NewContainer(
			Name("ck-taskflow-unit"),
			Image(imageCI),
			Command("test", "-f", "/app"),
		), nil).Priority(taskflow.PriorityHigh)
		sf.Container("integration", NewContainer(
			Name("ck-taskflow-integ"),
			Image(imageCI),
			Command("test", "-f", "/app"),
		), nil).Retry(2)
		sf.Container("lint", NewContainer(
			Name("ck-taskflow-lint"),
			Image(imageCI),
			Command("true"),
		), &taskflow.ContainerRun{Logs: os.Stdout}).Priority(taskflow.PriorityLow)
	})
	release := flow.If("release", func(context.Context) bool {
		return os.Getenv("RELEASE") == "1"
	})
	tag := flow.Task("tag", func(ctx context.Context) error {
		_, err := TagImage(ctx, cli.Engine(), Source(imageCI), Target(imageRelease))
		return err
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
