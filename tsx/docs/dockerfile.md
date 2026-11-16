# Dockerfiles

A Dockerfile is a second tree. Its components live in `containerkit/dockerfile`, not in `"containerkit"`, because both modules have an `Env`, a `Volume`, a `Command`, and an `Expose`, and they mean different things. The compose `Env` is an environment entry on a service. The Dockerfile `Env` is an `ENV` instruction. Import the one you mean.

```tsx
import { Build, Compose, Service } from "containerkit";
import { Command, Dockerfile, From, Run } from "containerkit/dockerfile";

const image = (
  <Dockerfile>
    <From image="alpine" tag="latest" />
    <Run cmd="echo ok" />
    <Command executable="/out/app" args={["--port", "8080"]} />
  </Dockerfile>
);

const stack = (
  <Compose name="app">
    <Service name="api">
      <Build>{image}</Build>
    </Service>
  </Compose>
);
```

`image` is not a project. `exportYAML` will not accept it. Nest it under `<Build>` inside a named `<Compose>`, as `stack` does, and the live session applies that project like any other tree. [nginx/docker.tsx](../examples/nginx/docker.tsx) is that shape as a program: the Dockerfile is built from the files next to the Go nginx example, and the service publishes port 8080. The build block stores the Dockerfile text inline. A build accepts one Dockerfile. It rejects a second source when `dockerfilePath`, `parseDockerfile`, or `readDockerfile` is already set on that build.

`exportDockerfile`, from `containerkit/dockerfile`, returns the Dockerfile as text and does not build it:

```dockerfile
FROM alpine:latest
RUN echo ok
CMD ["/out/app", "--port", "8080"]
```

A second `<Command>` or `<CommandShell>` in one file throws. One Dockerfile has one `CMD`.

## Scalar instructions

`Env`, `Label`, `Arg`, `Workdir`, and `User` preserve the string they receive from JSX. A literal is the child. A string from code stays on the prop (`value`, `default`, `path`, or `name`). A multiline literal can have its whitespace normalized by the compiler before the component sees it.

```tsx
<Env name="PORT">8080</Env>
<Label name="com.harbor.role">notes</Label>
<Arg name="APP_ENV">production</Arg>
<Arg name="APP_VERSION" />
<Workdir>/app</Workdir>
<User>1000:1000</User>
```

`<Arg name="APP_ENV" />` declares the argument and stores no default. `<Arg name="APP_ENV" default="" />` and `<Arg name="APP_ENV">{""}</Arg>` store an empty default. Those are different nodes, and they export as `ARG APP_ENV` and `ARG APP_ENV=`.

`Run`, `Command`, and `Entrypoint` stay on `cmd`, `executable`, and `args`. Heredocs, inline file content, and Compose options such as `Annotation`, `Sysctl`, and `DriverOpt` are unchanged.

## Exec form and shell form

`<Command>` writes the exec form, a JSON array. The program is executed directly, with no shell:

```dockerfile
CMD ["/out/app", "--port", "8080"]
```

`<CommandShell>` writes the shell form. The props are still `executable` and `args`. The text is one line the shell will interpret:

```dockerfile
CMD /out/app --port 8080
```

`<Entrypoint>` and `<EntrypointShell>` are the same split for `ENTRYPOINT`. `<Shell>` writes `SHELL` in exec form, which changes the shell later shell-form instructions use.

## Healthchecks

```tsx
import { HealthCheck, HealthCheckNone } from "containerkit/dockerfile";

<HealthCheck test={["CMD", "true"]} interval="5s" timeout="3s" retries={3} />
```

The `test` array starts with `"CMD"`, `"CMD-SHELL"`, or `"NONE"`.

`"CMD"` is exec form: each following string is one argument. `"CMD-SHELL"` is two strings, the prefix and the shell command. The file shows that command in shell form. The words `CMD-SHELL` are how you ask for that form. They are not written into the file as the command. `"NONE"` disables the healthcheck. `<HealthCheckNone />` writes that disable on its own.

## RUN layers

Two `<Run>` instructions in a row, with no flags, become one `RUN` layer:

```tsx
<Run cmd="npm ci" />
<Run cmd="npm run build" />
```

```dockerfile
RUN npm ci && \
	npm run build
```

`<RunLayer>` starts a new layer when the instruction before it was also a run. Anything else already ends the chain, including `<From>`, so a run after a new stage is `<Run>`:

```tsx
<Run cmd="npm ci" />
<RunLayer cmd="npm run build" />
```

```dockerfile
RUN npm ci
RUN npm run build
```

`<RunArgs args={["--extra"]} />` appends arguments to the previous `RUN`. `Run` flags are `mount`, `network`, and `security`. A `<Run>` that sets any of them stays its own layer, because those flags apply to that instruction only.

## Three ways to build

1. Nest `<Dockerfile>` under `<Build>` in the Compose tree, as the first example does. Returning that tree from `docker.ts` builds it when the project is applied.
2. `buildDockerfile` builds the tree now, through the session, without making it the project you return:

```ts
await app.buildDockerfile(image, {
  image: "app:dev",
  target: "runtime",
  platform: "linux/arm64",
  pull: true,
});
```

`image` on the options is the name to tag. `target` selects a stage. `platform` selects the architecture. `pull` pulls the `FROM` image first. `noCache` ignores the build cache. `tag` is an extra tag.

3. `buildImage` points at a Dockerfile that is already a file on disk. `source` is the build context, the directory sent to the build. `dockerfilePath` is the file inside that context. `target` selects a stage.

```ts
await app.buildImage({
  image: "app:dev",
  source: ".",
  dockerfilePath: "Dockerfile",
  target: "runtime",
});
```

## Instructions

`<Escape>` and `<Check>` are parser directives. They are written before the instructions. `<Syntax>` is the optional `# syntax=` line. Leave it unset. BuildKit then uses the Dockerfile frontend bundled with the builder, which is what current Docker already does. Set it only to pin an external frontend, such as `<Syntax value="docker/dockerfile:1" />` or `<Dockerfile syntax="docker/dockerfile:1-labs">`.

A file with two stages names the first one and starts a second `FROM`:

```tsx
import { Command, Dockerfile, From, Run } from "containerkit/dockerfile";

const image = (
  <Dockerfile>
    <From image="node" tag="22" as="build" />
    <Run cmd="npm run build" />
    <From image="alpine" tag="latest" />
    <Run cmd="apk add --no-cache ca-certificates" />
    <Command executable="/out/app" args={["--port", "8080"]} />
  </Dockerfile>
);
```

`From` takes `image`, and optional `tag`, `platform`, and `as`. `Arg` takes `name` and an optional `default`. `Copy` and `Add` take `src` and `dest`. `Copy` also takes `from`, `chown`, `chmod`, `link`, `exclude`, and `parents`. `Add` adds `checksum` and `keepGitDir`. `from` on `Copy` names an earlier stage, the `as` on a `<From>`.
