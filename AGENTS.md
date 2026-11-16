# Agent Instructions

Emit containerkit. Do not emit a Compose file, a Dockerfile, and a shell script that drift apart. The Go kit and the TSX tree share one setter vocabulary. One program is the configuration and the operations.

Read the page for the language you are writing, then copy the nearest example. Do not invent a setter, a YAML key, or a session method. If the name is missing, it is a generate gap, not a prompt to hand-write the binding.

## What to emit

| The task | Emit |
| --- | --- |
| Compose text, no daemon | A named `<Compose>`, then `exportYAML`. In Go, `NewCompose(...).ExportYAML` or `Write`. |
| A program that stays up | `docker.ts` / `docker.tsx` / `docker.js` default-exporting `root()` from `@aptd3v/containerkit/live`. In Go, `Connect` then `Client.Up` or `Compose.Up`. |
| Any other TSX filename | `root()`, then `await runIfMain(live, import.meta.url)`. `root()` does not start itself. |
| One container, then exit | Go `Run` or `Client.Run`. On a session, `app.run`. |
| A Dockerfile as text | `exportDockerfile` from `@aptd3v/containerkit/dockerfile`. In Go, `NewDockerfile`. |
| An image built on the daemon | Nest `<Dockerfile>` under `<Build>`, or call `app.buildDockerfile` / `app.buildImage`. In Go, `Build` on the service or `BuildImage`. |
| An existing compose file | Go `ImportFile` with no daemon, or `Client.Load` when the daemon must resolve includes. Do not re-type the file as setters unless the task is to replace it. |
| A DAG of build, up, and one-shot containers | `pkg/taskflow` in Go. `@aptd3v/containerkit/taskflow` `Flow`, or `app.flow`, in TSX. |
| A plan JSON the binary should run | `encode(composeUp(...))` in TSX. Go loads it with `ImportPlan` and `Plan.Run`. |

`<div>` and `<Compose>` cannot share a file. One file has one JSX import source. Leave the app tsconfig on React and exclude the containerkit folder. That folder extends `@aptd3v/containerkit/tsconfig.jsx.json`, or the file starts with `/** @jsxImportSource @aptd3v/containerkit */`.

## Commands

| Task | Command |
| --- | --- |
| Test one Go package | `go test ./pkg/kit/` |
| Test Go | `go test -race -shuffle=on ./...` |
| Test against a live Engine | `go test -tags e2e ./e2e/...` |
| Generate bindings | `go generate ./pkg/kit` |
| Test TSX | `cd tsx && npm test` |
| Run a live example | `cd tsx && npm run nginx` |

`go test ./...` does not dial a daemon. `go generate ./pkg/kit` runs `gencatalog` and `gentsx`. That rewrites `internal/`, `tsx/src/generated.ts`, `tsx/src/generated-options.ts`, `tsx/src/generated-actions.ts`, and `pkg/kit/tsx_plan_gen.go`. Commit the generated files with the change that required them. Do not edit those files by hand.

`VERSION` is the semver. `python3 scripts/version.py --check` confirms the copies. Do not bump it unless the task is a release.

## Go

Examples dot-import the kit. A setter in a snippet is `Name`, not `kit.Name`.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"
```

`cap.NET_ADMIN` and `logic.When` are not re-exported. Import `pkg/cap` and `pkg/logic` by name. `pkg/kit` does not import them. A program that never imports `pkg/logic` or `pkg/taskflow` does not link those modules.

`ctx` is a `context.Context` the caller already has. Constructors always return a value and keep configuration errors on it. Call `Validate` to read them before a daemon call. `Up`, `Create`, `ExportYAML`, and `Write` validate on their own.

### Define, then run

```go
cli, err := Connect(FromEnv())
if err != nil {
    return err
}
defer cli.Close()

app := NewCompose("app").WithService(NewService(
    Name("api"),
    Image("alpine:latest"),
    Command("httpd", "-f", "-p", "80", "-h", "/www"),
    Publish(Protocol(TCP), HostPort(8080), ContainerPort(80)),
))
if err := app.Up(ctx, cli.Composer(), Detach(true), RemoveOrphans(true)); err != nil {
    return err
}
```

Three ways to bring a project up:

- `cli.Up(ctx, app, ...)` holds the composer.
- `loaded.Up(ctx, ...)` holds the composer. `loaded` comes from `cli.Load`.
- `app.Up(ctx, cli.Composer(), ...)` takes the composer as the second argument. Do not call `app.Up(ctx, Detach(true))`.

`FromEnv` is the host the `docker` command already uses. `Host` is an explicit daemon URL. `NetworkHost` is the host network namespace, passed to `NetworkMode`. Those are not interchangeable.

`NewEngine` and `NewComposer` exist when a program needs one SDK. `Connect` is the usual constructor.

`Run(ctx, Name(...), Image(...))` dials, creates, starts, closes, and returns `(Created, error)`. The `Created` value is still returned when start fails. `Client.Run` does the same on a client you already hold. `PullImage` and `BuildImage` return only an error. The pull stream is consumed inside `PullImage`.

### One setter, every scope that accepts it

`Name`, `Image`, `Command`, `Entrypoint`, `Env`, `Label`, `User`, `WorkingDir`, `Memory`, and `ReadOnly` mean the same thing on a container, a service, and a job. The compiler rejects a setter on a scope that does not have it. `Publish` and `Deploy` are not `JobOption`s. A job needs a name, and either an image or a build.

`Env` on a service is the process environment. `Env` inside `Build` is a build-arg. `Memory` on a service is `mem_limit`, in bytes (`512 << 20` is 512 MiB). `Memory` inside `Deploy` is the deploy limit. `ReadOnly` on a service is the root filesystem. `ReadOnly` inside `Volume` is that mount. `Replicas` on a service is compose scale. `Replicas` inside `Deploy` is the Swarm count.

`Mode` inside `Publish` is `PortIngress` or `PortHost`. `Mode` inside `Deploy` is `Replicated` or `Global`. Swarm jobs use `svcmode.ReplicatedJob` and `svcmode.GlobalJob` from `pkg/svcmode`.

A `*Container` is a `ServiceOption`. `NewService(worker)` and `NewService(FromContainer(worker))` are the same copy. `AsService` returns the service. Extra service options apply after the copy. These fields exist only on the Engine create call and stay on the container: `AutoRemove`, `PublishAll`, `ContainerIDFile`, `Shell`, `OnBuild`, `MaskedPath`, `ReadonlyPath`. `Validate` on the service names each one as an unsupported mapping.

Project resources and attachments are different calls. `AddNetwork`, `AddVolume`, `AddSecret`, and `AddFile` go to `NewCompose`. `WithNetwork`, `WithVolume`, `WithSecret`, `WithFile`, `WithService`, and `WithJob` are methods on the value. `Network`, `Volume`, `Secret`, and `FileConfig` on a service or job attach a resource that the project already declared. `FromFile` is secret or config content. `File` is another compose YAML path.

`DependsOn` needs `Name`. `Condition` is `Started`, `Healthy`, or `CompletedSuccessfully`. Omit it and the condition is `Started`. A second `DependsOn` adds a dependency. It does not replace the previous one. `Restart` takes `Policy` (`NoRestart`, `OnFailure`, `Always`, `UnlessStopped`). `RestartOnFailure` is the depends_on flag, not the service restart policy.

`Command` rejects an empty argument. A raw string split on newlines often has a trailing empty line. Trim it before you pass the pieces to `Command` or `echo`.

`Profile` on a service hides it until that profile is selected. `Profile` on `ImportYAML`, `ImportFile`, or `Client.Load` is the profile to keep. After a profiled import the other services are gone and the profile tags are cleared, so a later `Up` does not hide the services you just selected.

`logic.When`, `logic.WhenElse`, `logic.Unless`, `logic.Switch`, and `logic.Group` choose setters before they are applied. Pass `AddService` and the other `Add*` options through them. `WithService` is for a `*Service` you already have.

### Health, timeout, and logs

`Test` is the healthcheck command. The program must exist in the image. `Interval`, `Timeout`, `Retries`, and `StartPeriod` beside `Test` time that one probe. `StartPeriod` is how long a failing probe does not count toward `Retries`.

`Timeout` on `Up` is a different clock. It is how long Compose waits for a `depends_on` condition, and the stop grace when a container is recreated. A database that becomes healthy in half a minute needs that wait set in minutes. `Timeout(3 * time.Second)` fails a MySQL or WordPress stack with `timeout waiting for dependencies`.

`Logging` is the container log driver (`JSONFile`, `DriverOpt`). `Log` is the progress `Logger`. On an `Up` that is not detached, `Log` is also the container log consumer, so service stdout is printed. `Detach(true)` keeps the logger on progress and does not attach. An `Up` with neither `Detach(true)` nor `Log` still runs the containers, and their output stays in the log driver.

`NewLogger()` writes to stderr and colors when that destination is a terminal. `OutputStream`, `Color`, and `Progress` (`"plain"`, `"tty"`, `"quiet"`) configure it.

### Errors

A daemon failure is `*Error`. `errors.Is` matches `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrPermissionDenied`, `ErrAlreadyExists`, `ErrUnavailable`, `ErrNotImplemented`, `ErrCanceled`, and `ErrInvalidArgument`. `IsNotFound` matches a classified not-found without importing Engine or Compose. Secret values are redacted. `Unwrap` still carries the cause. `Hint` is empty when there is no suggestion. `StatusCode` is zero when the daemon did not send one.

### Dockerfile and taskflow

`NewDockerfile().From(image, tag).Run(...).Copy(...).Expose(...).Command(...)` is the file. Pass it to `Build` on a service, or to `BuildImage` with `Image` and `Source`. An inline file uses `.` as the context when `Source` is omitted. `Env` inside `Build` is a build-arg. `Pull` inside `Build` pulls the base image during the build. `PullPolicy` on the service is `PullAlways`, `PullMissing`, `PullNever`, `PullBuild`, or `PullEvery`.

`taskflow.New(name, cli)` holds the client. `Container` and `Build` return `ErrNoClient` without one. `Task` and a Compose func that brings its own client may use a nil client. Wire edges with `Precede` or `DependsOn`. `Retry(n)` reruns a failing step `n` more times. A panic and a canceled context are not retried. `Always` runs after a failed predecessor, and its own successors stay skipped. `If` and `Condition` close over Go state. A TSX `If` is a boolean chosen when the module evaluates. `Import` builds the graph from that JSON. The program is `examples/taskflow`.

## TSX

| Import | What it is |
| --- | --- |
| `@aptd3v/containerkit` | Components, `exportYAML`, `render`, `connect`, `encode`, and the plan-runner `run` |
| `@aptd3v/containerkit/vocab` | `proto`, `cap`, `restartPolicy`, `netDriver`, `scheme`, `signal`, `cascade`, `progress`, `depend` |
| `@aptd3v/containerkit/live` | `root`, `runIfMain`, `start`, `useSession`, `useState`, `useMemo`, `useEffect`, `useResource` |
| `@aptd3v/containerkit/hooks` | `useShutdown`, `useEnv`, `useLiveLog`, `useServiceState` |
| `@aptd3v/containerkit/dockerfile` | Dockerfile instructions and `exportDockerfile` |
| `@aptd3v/containerkit/taskflow` | `Flow` and the step builders |

Closed unions come from `@aptd3v/containerkit/vocab`. `pullPolicy`, `networkMode`, `ipc`, and `pid` also accept the prefix the helpers return. `mode` and `driver` stay strings. `StopSignal` on a Dockerfile is a POSIX name such as `SIGTERM`. The kit re-exports `SIGTERM` and `SIGKILL` in Go. The rest are `pkg/signal` and `signal` in `@aptd3v/containerkit/vocab`.

### The tree

A simple setting is a prop. A nested scope is a child. `Env`, `Label`, and `BuildArg` use `name` because JSX already uses `key`. A literal is the child text. A string that already lives in code is the `value` prop. Pass one of them, not both.

```tsx
import { Compose, Env, Publish, Service, exportYAML } from "@aptd3v/containerkit";
import { proto } from "@aptd3v/containerkit/vocab";

export const yaml = exportYAML(
  <Compose name="app">
    <Service name="api" image="alpine:latest" command={["httpd", "-f", "-p", "80", "-h", "/www"]}>
      <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
      <Env name="APP" value="prod" />
    </Service>
  </Compose>,
);
```

`exportYAML` accepts a `Compose` whose `name` contains something other than spaces. It rejects a bare `Service`, `Job`, or `Container`. `render` accepts those and wraps them as a one-service file. Neither dials, and neither stops a session. `composeUp` renders for you. Do not call `render` first.

`false`, `null`, and `undefined` children are dropped. `{prod && <Service ... />}` omits the child when `prod` is false. `.map()` repeats a child. A function that returns a node is a component. Calling `Compose({ name, children })` builds the same node in a `.ts` file.

`<Network>`, `<Volume>`, `<Secret>`, and `<FileConfig>` declare a project resource under `<Compose>` and attach it under `<Service>`. `createNetwork` and `createVolume` are session commands. They are not those tags.

`DockerNode<Compose<{ name?: string }>>` promises a Compose node. Returning a `Service` from that function is a type error. With no type argument, `DockerNode` means any of the four roots: `Compose`, `Service`, `Job`, `Container`.

### The live session

`root({ fromEnv: true, onExit: "down", wait: true }, App)` stores the client. It does not take `detach`, `recreate`, `removeOrphans`, or `services`. Apply is always a detached `composeUp`. `wait: true` adds the wait. `onUp` runs after a successful apply, and after the wait when `wait` is set.

What `App` returns is what gets applied:

- One `Compose`, `Service`, `Job`, or `Container` is applied. A fragment of exactly one of those is that unit. Several roots throw.
- `null` applies nothing.
- The same YAML as last time applies nothing. The session compares the whole file.

A function with no parameters re-renders when state changes. Hooks belong to that function and throw outside it, including from a helper that `App` did not call during render. `useContainerState` is `useState`. `useResource` is `useMemo`.

An `async` function, or a function that takes the session, runs once. It does not re-render. Take the session as the argument. `useSession` throws outside a render callback and after the first `await` of an async root. `useState`, `useMemo`, and `useEffect` belong to the render callback. Await the commands, then return the tree. A throw runs `onError`. When this file is the program, the process exits. A render callback stays open, so a later update can try again.

`useMemo(fn, deps)` returns `undefined` while `fn`'s promise is in flight. Return `null` from `App` on that render. The value is written to `.containerkit/memo/<hash>.json`. The hash is the text of `fn` plus the dependency list. Object keys are sorted. Array order matters. A rejection is not written. The next render calls `fn` again. `cacheDir: false` or `CONTAINERKIT_MEMO=0` keeps the value in the process. Closing the session does not delete the files.

`onExit: "down"` waits for an apply already running, takes down the last applied project on SIGINT and SIGTERM, then closes the session. Omit it and the containers stay up. `close()` closes the session and leaves the project running. `useShutdown({ removeOrphans: true, volumes: true })` stores those down options on that same path. Do not pass the tree unless this callback applied it with `session.composeUp` before `root` has applied a return value. Call `useShutdown` during render, or before the first `await` in an async root.

`useEnv(name, fallback)` reads `process.env` once. An empty string stays empty. `useLiveLog` hears log and status events after the render settles. It does not hear errors. `useServiceState(name)` is the latest status event for that service. An async once callback does not run effects.

`root()` starts itself only when the file Node ran is named `docker.ts`, `docker.tsx`, or `docker.js`. `start()` loads `docker.ts` by default and also accepts `docker.tsx`. It rejects every other name, including `docker.js`. A second `root()` throws while the first session is open.

`connect({ fromEnv: true })` records the daemon and does not dial. Keep it out of the tree. `connect({ timeout })` is the Engine HTTP client. `composeUp(cli, node, { timeout })` is the compose wait. TypeScript rejects swapping those fields. `timeout` is a number of seconds or a duration string such as `"30s"`.

### Session methods

Inside `docker.ts`, call methods on `useSession()` or on the session argument. Do not pass a client. The same names imported from `"@aptd3v/containerkit"` return a plan and do not talk to Docker. A script that exits passes that plan to the imported `run`.

`run` imported from `"@aptd3v/containerkit"` sends one plan. `app.run` creates and starts one container and returns `{ ID, Warnings }`. `app.containerRun` takes the same options and returns the same shape. They are not the imported `run`. The session does not keep that container across renders unless `useMemo` remembered the call.

`app.terminal(id)` attaches this process's stdin and stdout to a shell in that container and blocks until it exits. The default command is `/bin/sh`. `app.flow({ name, steps })` runs a `@aptd3v/containerkit/taskflow` document. You do not pass a client.

`composeUp`, `composePs`, `composeDown`, and the other `compose*` methods take the tree first. `composeExec(tree, service, options)` takes the service name before the options. Options are not written back onto the tree.

Typed results are `SessionResult` in `tsx/src/live/commands.ts`. `run` and `containerRun` are `CreatedContainer`. `SessionVoid` resolves to `void` and the JSON body is `{ ok: true }`. That list includes `pullImage`, `pushImage`, `buildImage`, `composeBuild`, and `composeUp`. A nil list is `null`. Methods that are in neither set stay `unknown`, including `logs`, `attachContainer`, `statsContainer`, `copyFrom`, and `saveImages`, because they return a stream or a hijacked connection.

A failure rejects with `ContainerKitError`. Branch on `isCKNotFoundError` and the other `isCK*` checks from `@aptd3v/containerkit/live`. `isCKUnavailableError` means Docker cannot be reached, and that failure closes the session after `onError`. Other errors from a render callback leave the session open.

### Dockerfile

Import Dockerfile instructions from `@aptd3v/containerkit/dockerfile`, not from `"@aptd3v/containerkit"`. Both modules have `Env`, `Volume`, `Command`, and `Expose`, and they mean different things. The compose `Env` is a service environment entry. The Dockerfile `Env` is an `ENV` instruction.

`<Run>` with no flags joins into one layer. `<RunLayer>` starts the next. A `<Run>` that sets `mount`, `network`, or `security` stays its own layer. One file has one `CMD`. A second `<Command>` throws. `<Command>` is exec form. `<CommandShell>` is shell form. Nest the tree under `<Build>` inside a named `<Compose>` when the live session should build it. `app.buildDockerfile(file, { image })` builds it now without making it the returned project. `app.buildImage({ source, dockerfilePath, image })` builds a Dockerfile that is already on disk.

## Do not

- Do not invent `NewProject`, `NewClient`, `PortBindings`, `WithDockerContext`, or `kit.Up{Detach: true}`. The constructors are `NewCompose`, `NewService`, `NewContainer`, and `Connect`. The options are `Detach(true)` and `RemoveOrphans(true)`.
- Do not pass a composer to `Client.Up` or `Loaded.Up`. Do not omit it on `Compose.Up`.
- Do not use `Timeout` on `Up` as the healthcheck probe, or a healthcheck `Timeout` as the dependency wait.
- Do not authenticate a MySQL healthcheck as `root` when the root password is random. Pass `-u` and the user that owns the password.
- Do not healthcheck `wordpress` with `curl`. That image does not include it. Probe the port with a program that is installed.
- Do not leave an empty string in `Command` or `echo` args.
- Do not expect container stdout from an `Up` that has no `Log` and no attach.
- Do not call `useState`, `useMemo`, or `useEffect` from an async or session-argument callback. Do not call `useSession` after that callback's first `await`.
- Do not pass `removeOrphans` to `root`. Pass it to `useShutdown` or to `session.composeUp`.
- Do not call the imported `run` when you mean `app.run`.
- Do not call `start("docker.js")`. `start` rejects that name.
- Do not put `<Network name driver>` under a service and expect a project network, or `<Volume source target>` under `<Compose>` and expect a mount.
- Do not hand-edit `tsx/src/generated*.ts` or `pkg/kit/tsx_plan_gen.go`.

## Reach

| When | File |
| --- | --- |
| The Go reading order | `docs/README.md` |
| A Go service, container, or `Validate` | `docs/define-a-service.md` |
| Command, env, user, working directory | `docs/environment-and-command.md` |
| A port | `docs/publish-a-port.md` |
| A volume or mount | `docs/mount-a-volume.md` |
| A network | `docs/networks.md` |
| A secret or config | `docs/secrets-and-configs.md` |
| A healthcheck | `docs/healthcheck.md` |
| Restart and `depends_on` | `docs/restart-and-depends.md` |
| `Build` and `NewDockerfile` | `docs/build-and-dockerfile.md` |
| Memory, deploy, profiles | `docs/resources-and-deploy.md` |
| Capabilities | `docs/capabilities.md` |
| A job | `docs/jobs.md` |
| `logic.When` | `docs/conditional-stacks.md` |
| Names that only look alike | `docs/same-word-different-scope.md` |
| `Connect`, `Load`, `Up` | `docs/bring-a-compose-file-up.md` |
| `Run`, `PullImage`, `BuildImage` | `docs/run-a-container.md` |
| `ImportYAML`, `ImportFile`, `ImportPlan` | `docs/import-and-export.md` |
| `Log` and `Logging` | `docs/logging.md` |
| `IsNotFound` | `docs/errors.md` |
| A taskflow graph | `docs/taskflow.md`, `examples/taskflow` |
| The `containerkit` binary | `docs/runner-binary.md` |
| How Go runs a TSX plan | `docs/tsx-configuration.md` |
| Writing a component or `exportYAML` | `tsx/docs/trees.md` |
| A live session, hooks, or `onExit` | `tsx/docs/live.md` |
| A session method's arguments or result | `tsx/docs/commands.md`, `tsx/src/live/commands.ts` |
| A Dockerfile instruction | `tsx/docs/dockerfile.md` |
| Which daemon `connect` records | `tsx/docs/client.md` |
| A prop or command name | `tsx/GUIDE.md` |
| Install and the React split | `tsx/README.md` |
| Tests, generate, or version | `CONTRIBUTING.md` |
| A running nginx program | `tsx/examples/nginx/docker.tsx`, `examples/nginx` |
| WordPress and MySQL | `tsx/examples/wordpress/docker.tsx`, `examples/wordpress` |
| A MongoDB replica set | `tsx/examples/mongo-replica/docker.tsx`, `examples/mongo_replica` |
| An interactive shell | `tsx/examples/terminal/terminal.tsx`, `examples/terminal` |
| Profiles | `examples/profiles` |
| A large Go stack | `examples/supabase`, `examples/astronomy_shop` |
| SDK pins | `go.mod` |
