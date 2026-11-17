# Taskflow

[taskflow](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/taskflow) is an optional DAG over the containers, builds, and Compose projects the kit already runs. `pkg/kit` does not import it. A program that never imports `taskflow` never downloads go-taskflow. The `containerkit` binary does, so a plan document with a `flow` field can run.

```go
import (
    . "github.com/aptd3v/containerkit/pkg/kit"
    "github.com/aptd3v/containerkit/pkg/taskflow"
)
```

The runnable program is [examples/taskflow](../examples/taskflow).

## Graph

`New` names the graph and holds a `*Client`. Nil is enough for `Task` steps and for a Compose func that brings its own client. `Container` and `Build` return `ErrNoClient` without one. Add steps, wire edges, then call `Run` once.

```go
flow := taskflow.New("ci", cli)

df := NewDockerfile().From("alpine", "latest").Command("true")
db := NewCompose("ci").WithService(NewService(
    Name("db"),
    Image("postgres:16-alpine"),
))

build := flow.Build("build", df, Tag("app:ci"))
up := flow.Compose("up-db", db, taskflow.Up(Detach(true), Wait(true)))
test := flow.Container("test", NewContainer(
    Image("app:ci"),
    Command("true"),
), nil)

build.Precede(up)
up.Precede(test)
err := flow.Run(ctx, &taskflow.Run{Writer: os.Stdout})
```

`a.Precede(b)` is the same edge as `b.DependsOn(a)`. Both return the task, so `test.DependsOn(up.DependsOn(build))` is one chain. Independent steps run together. A failed step skips its successors. `Always` still runs, and its own successors stay skipped. `Retry(n)` reruns a failing step `n` more times. `SkipIfPresent` on a Build step inspects every tag before building. All present finishes the step done, with reason `already_present`, and does not build, so the next step still runs. A missing tag builds once with the original options. Any other inspect error fails the step and does not build, whichever tag is first. Zero tags is an error before any inspect, and that configuration error is not retried. `SkipIfPresent` on any other step fails the step once and does not run it. A panic and a canceled context are not retried. A cycle is an error from `Run` and from `Import`, and the error names the tasks. An edge between two flows is an error from `Run`, and the error names both flows and both tasks.

`If` and `Condition` close over Go state while the graph is running. `Then` and `Else` are the branches. `Precede` order is the condition index. `Subflow` nests the same constructors. The function passed to `Subflow` runs when that subflow starts.

`Service` takes the same project as `Compose`. It calls `Up` and returns when the project is ready. Caller up options are applied first, then `Detach(true)` and `Wait(true)`. Turning either off is `Compose`. A `Flow.Service` stays up until every other task in the flow has finished. A `Subflow.Service` stays up until every other task in that subflow has finished, then downs, then the parent continues. The down is not a successor of the service's direct dependents, so `db → migrate → integration-test` keeps the database up for the test. Two services in one scope tear down without waiting on each other. `Down` runs once, without a volume wipe. A not-found result is success. Cancellation does not skip that down: it uses a bounded context from `context.Background`. A killed process does not call `Remove`.

`Volume(logical)` returns the physical name to mount, allocated when it is called, before `Run` creates a run id. The logical name is not the engine name. Two calls get two volumes. `Run` creates the physical volume labeled `containerkit.flow` and `containerkit.run`. `Remove` deletes it only after this run's containers are gone, and only when the run label matches. A volume still in use returns that error. A pre-created volume is never that physical name, so it stays. A flow document stores the logical names. A container mount source `volume:<logical>` is rewritten to the physical name on import.

`Dump` writes DOT. After `Run`, `List` is the containers labeled `containerkit.flow` and `containerkit.run`. Before a run id exists, `List` is empty. `Remove` downs every Compose project the graph used and removes its volumes, then deletes the container ids this run created and any leftover container labeled with this run, then removes this run's private volumes. The package `List` is every container with that flow label. `Remove` does not run if the process is killed.

A container step is one shot. A non-zero exit fails the step. `Always`, `UnlessStopped`, `OnFailure`, and `AutoRemove` are rejected on that container. The step writes `Name` and the `containerkit.*` labels onto it. An omitted name is unique to that run. A name that already exists is a create conflict, and nothing was removed first. A retry removes the container id that step created.

`Status` is the terminal outcome. It stays zero until the step finishes, including while a retry is in progress. `Err` and `ExitCode` stay zero when the step was skipped. A child suppressed by invalid-subflow validation never starts, retains zero status, and emits no start or done event. The subflow reports the error.

## The same document from TSX

`@aptd3v/containerkit/taskflow` `Flow` builds the JSON document. It accepts `volumes`, the logical names. `volumeSource` emits the `volume:<logical>` mount source Import rewrites. `Service` emits `op: "service"` with the Compose YAML and up options. `ComposeStep` uses the op to select the options type and the required fields. An inline options object is checked against that type, so `volumes` on an inline `composeUp` is rejected, `composePort` requires `port`, and `composePublishOCI` requires `repository`. A predeclared options value is checked structurally, so extra properties can remain when its type overlaps. [Import](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/taskflow#Import) turns that JSON into this graph, and `Run` executes it. A TSX `If` is a boolean chosen when the module evaluates. `Task`, `If` with a Go predicate, and `Condition` with a Go predicate stay in Go. They are not in the document.
