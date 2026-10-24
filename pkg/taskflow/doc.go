// Package taskflow orders kit and client work as a task graph.
//
// Import it the same way as kit:
//
//	import "github.com/aptd3v/containerkit/pkg/taskflow"
//
// A program that already has a *kit.Container, a kit.Build, or a *kit.Compose
// can hand those values to a Flow instead of sequencing them in main. Edges
// are Precede (or its inverse, DependsOn). A failed step skips its
// successors; Always opts a step out of that for teardown. Retry reruns a
// failing step. If is a boolean branch with Then and Else. Condition is an
// indexed branch that also runs after failures, for retry cycles. A Subflow
// groups steps and counts as failed when any step inside it failed. Priority
// only orders the scheduler queue.
//
//	flow := taskflow.New("ci", cli)
//	build := flow.Build("build", kit.NewBuild().Dockerfile(df), &client.ImageBuild{Tags: []string{"app:ci"}})
//	up := flow.Compose("up-db", db, &kit.Up{Detach: true, Wait: true})
//	test := flow.Container("test", kit.NewContainer().Image("app:ci"), nil).Retry(2)
//	build.Precede(up)
//	up.Precede(test)
//	defer flow.Remove(context.Background())
//	err := flow.Run(ctx, &taskflow.Run{Profile: os.Stdout})
//
// Dump writes DOT. Run.Profile and Run.Trace receive profiler output when the
// graph finishes. Run.Watch is step lifecycle; Run.Log is step-scoped build,
// container, and compose output. Container and Compose steps stamp
// containerkit.flow / task / depends / precedes labels. List is docker ps -a
// for that label; call it before Remove. Remove downs Compose projects the
// graph referenced, then deletes leftover labeled containers.
// pkg/kit and pkg/client do not import this package; programs that never
// import taskflow never download go-taskflow.
package taskflow
