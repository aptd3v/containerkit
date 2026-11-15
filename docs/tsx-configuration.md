# Declare config in TSX

Write a compose file as JSX. A module exports a function that returns a `DockerNode`. `render` at the edge emits compose-spec YAML. `connect` / `up` / `pullImage` / `listImages` encode client and verb setters as a plan JSON document. Go loads YAML with `ImportYAML` or the plan with `ImportPlan`, then `Connect` / `Up` / `Run`.

This is not a React app, not a browser Docker client, and not a workflow store.

```tsx
/** @jsxImportSource containerkit */
import { Compose, Service, Publish, Env, connect, encode, composeUp, type DockerNode } from "containerkit";

export default function App(): DockerNode {
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest">
        <Publish hostPort={8080} containerPort={80} />
        <Env name="APP">1</Env>
      </Service>
    </Compose>
  );
}

const cli = connect({ fromEnv: true, timeout: "10s" });
export const encoded = encode(composeUp(cli, <App />, { detach: true }));
```

Same word, three types:

- `<Health timeout="5s" />` — define / healthcheck
- `connect({ timeout: "10s" })` — Engine HTTP client
- `composeUp(cli, node, { timeout: "30s" })` — compose wait/stop

`cli` as a factory prop marks I/O. It is not serialized onto `DockerNode`.

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upYAML(ctx context.Context, yaml []byte) error {
    compose := ImportYAML(yaml)
    cli, err := Connect()
    if err != nil {
        return err
    }
    defer cli.Close()
    return cli.Up(ctx, compose)
}
```

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upPlan(ctx context.Context, json []byte) error {
    return ImportPlan(json).Run(ctx)
}
```

Leave a React app's `jsxImportSource` on `react`. Containerkit files use `/** @jsxImportSource containerkit */`, or a folder `tsconfig` that extends `containerkit/tsconfig.jsx.json` and is excluded from the app config. One file has one import source, so `<div>` and `<Compose>` cannot share a file. `jsx: "react-jsx"` is the transform's name; the language server uses the `JSX` namespace from that file's import source. If the source does not resolve, TypeScript falls back to a global namespace named `React`.

The TypeScript package is `containerkit` on npm. A checkout that has not been published yet can depend on the local build with `"containerkit": "file:../containerkit/tsx"` after `npm --prefix tsx run build`. Components and option bags are generated from `pkg/kit` (`go run ./tools/gentsx`). Nested kit scopes are children; scalars are props. `<Network>` / `<Volume>` under `<Compose>` are project resources; under `<Service>` they are attachments. Dockerfile instructions are `containerkit/dockerfile`: `<Dockerfile>` under `<Build>` renders `dockerfile_inline`, and `buildDockerfile` sends the same text as `parseDockerfile`.

`run()` spawns `containerkit` (`cmd/containerkit`) when `CONTAINERKIT_BIN` or `PATH` has the binary. See [runner-binary.md](runner-binary.md), [tsx/README.md](../tsx/README.md), and [tsx/examples/supabase.tsx](../tsx/examples/supabase.tsx).
