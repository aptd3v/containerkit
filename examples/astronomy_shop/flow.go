// Package main: taskflow DAG: Build(kit-probe) -> Compose Up -> probe -> print URLs.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/taskflow"
)

func storeURL(cfg *Config) string {
	return fmt.Sprintf("http://127.0.0.1:%s/", cfg.EnvoyPort)
}

func printURLs(w io.Writer, cfg *Config, profiles []string) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Astronomy Shop is up:")
	fmt.Fprintf(w, "  Store:          %s\n", storeURL(cfg))
	fmt.Fprintf(w, "  Load generator: http://127.0.0.1:%s/\n", cfg.LocustWebPort)
	if slices.Contains(profiles, profileObservability) {
		fmt.Fprintf(w, "  Grafana:        http://127.0.0.1:%s/grafana/\n", cfg.EnvoyPort)
		fmt.Fprintf(w, "  Jaeger:         http://127.0.0.1:%s/jaeger/ui\n", cfg.JaegerUIPort)
		fmt.Fprintf(w, "  Prometheus:     http://127.0.0.1:%s/\n", cfg.PrometheusPort)
	}
	fmt.Fprintln(w, "")
}

// runFlow validates, optionally exports, then runs Build -> Up(Wait) -> probe.
func runFlow(ctx context.Context, cfg *Config, project *kit.Project, profiles []string, skipBuild bool) error {
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	app := kit.NewCompose(project)
	waitTimeout := 900
	up := &kit.Up{
		Detach:        true,
		Wait:          true,
		WaitTimeout:   &waitTimeout,
		Profiles:      profiles,
		RemoveOrphans: true,
		Writer:        os.Stdout,
	}

	flow := taskflow.New("astronomy-shop", cli)
	var build *taskflow.Task
	if !skipBuild {
		build = flow.Build("build-kit-probe", kitProbeBuild(), &client.ImageBuild{
			Tags: []string{probeImage},
		})
	}
	upTask := flow.Compose("compose-up", app, up)
	probe := flow.Container("kit-probe", kitProbeContainer(cfg), &taskflow.ContainerRun{Logs: os.Stdout}).
		Retry(5)
	urls := flow.Task("print-urls", func(context.Context) error {
		printURLs(os.Stdout, cfg, profiles)
		return nil
	})

	if build != nil {
		build.Precede(upTask)
	}
	upTask.Precede(probe)
	probe.Precede(urls)

	if err := flow.Dump(os.Stdout); err != nil {
		return err
	}

	tearDown := func() {
		_ = app.Down(context.Background(), &kit.Down{
			RemoveOrphans: true,
			RemoveVolumes: true,
			Profiles:      profiles,
			Writer:        os.Stdout,
		})
		_ = flow.Remove(context.Background())
	}

	err = flow.Run(ctx, &taskflow.Run{
		Writer: os.Stdout,
		Watch: func(e taskflow.Event) {
			switch e.Kind {
			case taskflow.EventRetry:
				fmt.Printf("[%s] attempt %d failed: %v\n", e.Name, e.Attempt, e.Err)
			case taskflow.EventSkip, taskflow.EventError:
				fmt.Printf("[%s] %s %v\n", e.Name, e.Kind, e.Err)
			}
		},
	})
	if err != nil {
		tearDown()
		return err
	}
	fmt.Fprintln(os.Stdout, "stack is up; Ctrl+C to tear down")
	waitInterrupt()
	tearDown()
	return nil
}

func waitInterrupt() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
