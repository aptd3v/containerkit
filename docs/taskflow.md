# Taskflow

[taskflow](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/taskflow) is an optional DAG over the same containers, builds, and Compose projects the kit already runs. `pkg/kit` does not import it. A program that never imports `taskflow` never downloads go-taskflow. The `containerkit` binary does, so a plan document with a `flow` field runs.

```go
import (
    . "github.com/aptd3v/containerkit/pkg/kit"
    "github.com/aptd3v/containerkit/pkg/taskflow"
)
```

Runnable demo: [examples/taskflow](../examples/taskflow).

## Graph

`New` names the graph and holds a `*Client`. Nil is fine for `Task` steps and Compose funcs. `Container` and `Build` return `ErrNoClient` without one. Add steps, wire edges, then `Run` once.

```go
flow := taskflow.New("ci", cli)
build := flow.Build("build", df, Tag("app:ci"))
up := flow.Compose("up-db", db, taskflow.Up(Detach(true), Wait(true)))
test := flow.Container("test", NewContainer(Image("app:ci"), Command("go", "test", "./...")), nil)

test.DependsOn(up.DependsOn(build))
err := flow.Run(ctx, &taskflow.Run{Writer: os.Stdout})
```

`a.Precede(b)` is the same edge as `b.DependsOn(a)`. Independent steps run together. A failed step skips its successors. `Always` still runs, and its own successors stay skipped. `Retry(n)` reruns a failing step `n` more times. A panic and a canceled context are not retried.

`If` and `Condition` close over Go state while the graph is running. `Then` is the true branch. `Precede` order is the condition index. `Subflow` nests the same constructors.

`Dump` writes DOT. `List` is the containers labeled `containerkit.flow`. `Remove` downs every Compose project the graph used, then deletes leftover labeled containers.

## TypeScript

`containerkit/taskflow` builds the same document. `Flow` puts a `flow` field on the plan next to `connect`. `run` sends that JSON. Go `Import` builds the graph and `Run`s it. A TSX `If` is a boolean chosen while the module evaluates, the same moment `logic.When` decides. `Task`, `If(pred func)`, and `Condition(pred func)` stay Go-only.

```ts
import { Compose, Service, connect, run } from "containerkit";
import { Build, ComposeStep, Container, Flow, If, Tag } from "containerkit/taskflow";

const cli = connect({ fromEnv: true });
const db = Compose({ name: "app", children: Service({ name: "db", image: "postgres:16" }) });

const doc = Flow(cli, {
  name: "ci",
  steps: [
    Build({ name: "build", image: "app:ci", source: "." }),
    ComposeStep({ name: "up-db", node: db, op: "composeUp", options: { detach: true, wait: true }, dependsOn: ["build"] }),
    Container({ name: "test", image: "app:ci", command: ["go", "test", "./..."], dependsOn: ["up-db"] }),
    If({ name: "release", when: process.env.RELEASE === "1", dependsOn: ["test"], then: [Tag({ name: "tag", image: "app:ci" })] }),
  ],
});
await run(doc);
```

Edges are step names (`dependsOn`, `precede`). `retry`, `always`, and `priority` (`high`, `normal`, `low`) are props on a step. `then` and `else` are step arrays. `Subflow` is a nested `steps` array baked in before `run`. Compose steps store YAML from `render` and an options object such as `UpOptions`.

A live session exposes the same helper as `session.flow`. Step events arrive on `session.on("log")` and `session.on("status")`.

[tsx/examples/taskflow.tsx](../tsx/examples/taskflow.tsx) is a release graph on top of a Dockerfile and a Compose project. `npm run taskflow` in `tsx/` prints the plan document. The image build overlaps Postgres. Lint waits for that image, then overlaps the database. Migrate retries. A failed check skips the tag. Down waits only for migrate, because unit and lint do not use Postgres, and still runs when migrate fails.
