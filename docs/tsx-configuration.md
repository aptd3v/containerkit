# A TSX plan, run from Go

A TSX module is a document this package loads. `exportYAML` on a named `Compose` is compose text, and [ImportYAML](import-and-export.md) reads it with no daemon. `encode(composeUp(...))` is plan JSON, and [ImportPlan](import-and-export.md) reads it. Neither call dials. [Plan.Run](https://pkg.go.dev/github.com/aptd3v/containerkit/pkg/kit#Plan.Run) is what dials.

This is not a React app and not a browser Docker client. One file has one JSX import source, so `<div>` and `<Compose>` do not share a file.

```tsx
/** @jsxImportSource containerkit */
import { Compose, Env, Publish, Service, composeUp, connect, encode, type DockerNode } from "containerkit";

export default function App(): DockerNode {
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest">
        <Publish hostPort={8080} containerPort={80} />
        <Env name="APP" value="1" />
      </Service>
    </Compose>
  );
}

const cli = connect({ fromEnv: true, timeout: "10s" });
export const encoded = encode(composeUp(cli, <App />, { detach: true }));
```

The same word is a different value in each place:

- `<Health timeout="5s" />` is the healthcheck probe.
- `connect({ timeout: "10s" })` is the Engine HTTP client.
- `composeUp(cli, node, { timeout: "30s" })` is the compose wait.

`cli` marks I/O. It is not serialized onto the node.

Go loads whichever document you kept:

```go
import (
    "context"

    . "github.com/aptd3v/containerkit/pkg/kit"
)

func upYAML(ctx context.Context, yaml []byte) error {
    compose := ImportYAML(yaml)
    cli, err := Connect(FromEnv())
    if err != nil {
        return err
    }
    defer cli.Close()
    return cli.Up(ctx, compose)
}

func upPlan(ctx context.Context, json []byte) error {
    return ImportPlan(json).Run(ctx)
}
```

Leave a React app's `jsxImportSource` on `react`. A containerkit file uses `/** @jsxImportSource containerkit */`, or a folder tsconfig that extends `containerkit/tsconfig.jsx.json` and is excluded from the app config. `jsx: "react-jsx"` is the transform's name. The language server uses the `JSX` namespace from that file's import source.

Components and option bags are generated from `pkg/kit` (`go run ./tools/gentsx`). Nested kit scopes are children. Scalars are props. `<Network>` and `<Volume>` under `<Compose>` are project resources. Under `<Service>` they are attachments. Dockerfile instructions are `containerkit/dockerfile`.

`run()` spawns the [runner](runner-binary.md) when `CONTAINERKIT_BIN` or `PATH` has the binary. The programs you can run are under `tsx/examples`: [nginx](../tsx/examples/nginx/docker.tsx), WordPress, a MongoDB replica set, and an interactive shell. From `tsx/`, `npm run nginx` builds the binary and starts the nginx session.
