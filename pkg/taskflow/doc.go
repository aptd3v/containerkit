// Package taskflow is an optional DAG runner for containerkit.
//
// Hand the same *kit.Container and *kit.Compose values you already have to a
// [Flow], plus ordinary Go functions. pkg/kit does not import this package; a
// program that never imports taskflow never downloads go-taskflow. The
// containerkit binary blank-imports it so a plan document with a flow field
// can run.
//
//	import "github.com/aptd3v/containerkit/pkg/taskflow"
//
// New names the graph and holds an optional *kit.Client. Add steps, wire
// edges, then call Run once. Independent steps whose predecessors are done run
// concurrently. Run.Concurrency is the worker pool (0 means GOMAXPROCS, raised
// above the number of subflows).
//
// cli may be nil when every step is Task or a Compose func that brings its
// own client. Container and Build need a client ([ErrNoClient] otherwise).
//
// Flow and [Subflow] share the step constructors. Every one returns *Task.
//
//   - Task(name, fn) — fn(ctx) error. Go only.
//   - Build(name, ...EngineBuildOption) — Engine image build
//   - Container(name, *kit.Container, *ContainerRun) — remove, create, start, wait
//   - Compose(name, *kit.Compose, cmd) — [Up], [Down], … or a func
//   - Subflow(name, instantiate) — nested graph; instantiate runs when the subflow starts
//   - If(name, pred) — boolean branch (Then / Else). Go only.
//   - Condition(name, pred) — indexed branch (Precede order). Go only.
//
// [Import] builds the same graph from the JSON document the TSX Flow helper
// emits. Task, If, and Condition callbacks are not in that document: a TSX If
// is a boolean chosen when the module renders.
//
// Container is a one-shot job. A non-zero exit fails the step. Restart
// policies (always, unless-stopped, on-failure) and AutoRemove are rejected.
// The step writes Name and the containerkit.* labels onto the container.
//
// A failed or skipped step skips its successors. Always runs a step even when
// a predecessor failed. Retry(n) reruns a failing step n more times. Panics
// and a canceled context are not retried. If is skipped when a predecessor
// failed, and then neither branch runs. Condition still runs after a failed
// predecessor.
//
// Dump writes DOT. List is docker ps -a for the flow label. Remove downs every
// Compose project the graph referenced, then deletes leftover labeled
// containers.
package taskflow
