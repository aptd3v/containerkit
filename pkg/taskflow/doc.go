// Package taskflow is an optional DAG runner for containerkit.
//
// Hand the same *kit.Container, kit.Build, and *kit.Compose values you already
// have to a [Flow], plus ordinary Go functions. kit and client do not import
// this package; a program that never imports taskflow never downloads
// go-taskflow.
//
//	import "github.com/aptd3v/containerkit/pkg/taskflow"
//
// # Graph
//
// New names the graph and holds an optional *client.Client. Add steps, wire
// edges, then call Run once. Independent steps whose predecessors are done run
// concurrently. Run.Concurrency is the worker pool (0 means GOMAXPROCS, raised
// above the number of subflows).
//
//	flow := taskflow.New("ci", cli)
//	build := flow.Build("build", kit.NewBuild().Dockerfile(df), &client.ImageBuild{Tags: []string{"app:ci"}})
//	up := flow.Compose("up-db", db, &kit.Up{Detach: true, Wait: true})
//	test := flow.Container("test", kit.NewContainer().Image("app:ci"), nil).Retry(2)
//	test.DependsOn(up.DependsOn(build))
//	defer flow.Remove(ctx)
//	err := flow.Run(ctx, &taskflow.Run{Writer: os.Stdout})
//
// cli may be nil when every step is Task or Compose. Container and Build need a
// client ([ErrNoClient] otherwise).
//
// # Steps
//
// Flow and [Subflow] share these constructors. Every one returns *Task.
//
//   - Task(name, fn) — fn(ctx) error
//   - Build(name, kit.Build, *client.ImageBuild) — Engine image build
//   - Container(name, *kit.Container, *ContainerRun) — remove, create, start, wait for exit
//   - Compose(name, *kit.Compose, cmd) — *kit.Up, *kit.Logs, … or func(ctx, app) error
//   - Subflow(name, instantiate) — nested graph; instantiate runs when the subflow starts
//   - If(name, pred) — boolean branch (Then / Else)
//   - Condition(name, pred) — indexed branch (Precede order)
//
// Container is a one-shot job. A non-zero exit fails the step. Restart policies
// (always, unless-stopped, on-failure) and AutoRemove are rejected: this step
// records the first exit, then copies logs. Long-running services belong in
// Compose with *kit.Up; flaky jobs use Retry. The step writes Name and
// containerkit.* labels onto the *kit.Container you pass.
//
// Compose cmd is a *kit option. The struct is copied; a nil Writer becomes
// Run.Writer. Use a func for Events, LoadProject, or options that depend on
// earlier steps.
//
// # Edges
//
// a.Precede(b) is b.DependsOn(a): A then B. Priority only orders the ready
// queue; it does not override edges.
//
// # Failure
//
// A failed or skipped step skips its successors. Siblings keep running. Run
// joins every step error.
//
// Always runs a step even when a predecessor failed (teardown inside the
// graph). Retry(n) reruns a failing step n more times. Panics and a canceled
// context are not retried.
//
// If is skipped when a predecessor failed, and then neither branch runs.
// Condition still runs after a failed predecessor, which is what makes a retry
// cycle work: a non-condition node is the first entry, and the Condition sits
// on the back edge. For retrying one step, Retry is simpler than a cycle.
//
// Prefer defer flow.Remove for end-of-run teardown so List can still see the
// stack. Use a *kit.Down step when the graph itself must bring the project
// down (and Always if that must happen after a failure).
//
// # Observe
//
// Dump writes DOT. Run.Profile is flamegraph text; Run.Trace is Chrome Trace
// JSON. Run.Watch is step lifecycle (start, done, skip, retry, error). Run.Log
// is step-scoped bytes (build stream, container logs, compose progress).
//
// Container and Compose steps stamp containerkit.flow / task / depends /
// precedes. List is docker ps -a for that flow label; call it before Remove.
// Remove Downs every *kit.Compose the graph referenced, then deletes leftover
// labeled containers.
//
// The full guide is the README in this directory. A runnable CI graph is
// examples/taskflow.
package taskflow
