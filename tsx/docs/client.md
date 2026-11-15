# Client

A client is the answer to "which Docker?". `connect` stores that answer. It does not open a connection.

```ts
import { connect } from "containerkit";

const cli = connect({ fromEnv: true, timeout: "10s" });
```

In `docker.ts` you usually do not call `connect`. `root({ fromEnv: true }, App)` stores the same kind of settings, and `useSession()` is `app`. A command is a method on that session:

```ts
await app.pullImage("alpine:latest");
```

`await` continues after the pull finishes. `connect` is for a script that does one thing and exits, shown at the end of [Commands](commands.md). The usual program is the live session.

Keep the client out of the tree. A `Compose` node becomes YAML. A socket path, a timeout, and a certificate path are not service settings, so they are never props of `<Compose>` or `<Service>`.

## The Docker you already use

```ts
const cli = connect({ fromEnv: true });
```

`fromEnv: true` reads the host and the certificates from the environment the process was started in. That is the same place the `docker` command looks. On `root`, pass it as `root({ fromEnv: true }, App)`.

Three narrower switches read one piece and leave the rest alone:

| Field | What it reads |
| --- | --- |
| `hostFromEnv` | Only the host |
| `tlsFromEnv` | Only the certificates |
| `apiVersionFromEnv` | Only the API version |

## A specific daemon

```ts
import { connect } from "containerkit";
import { scheme } from "containerkit/vocab";

const cli = connect({
  host: "unix:///var/run/docker.sock",
  scheme: scheme.Unix,
  timeout: "10s",
});
```

`host` is the daemon URL. `timeout` is how long to wait while connecting: a number of seconds, or a duration string such as `"10s"`. `scheme` is how that URL is spoken:

| Name | Meaning |
| --- | --- |
| `scheme.Unix` | A Unix socket, the usual local Docker |
| `scheme.Npipe` | A named pipe, the usual local Docker on Windows |
| `scheme.TCP` | A TCP address |
| `scheme.HTTP` | HTTP |
| `scheme.HTTPS` | HTTPS |
| `scheme.SSH` | SSH |
| `scheme.FD` | An already-open file descriptor |

A daemon on another machine, with certificate files:

```ts
const cli = connect({
  host: "tcp://192.0.2.10:2376",
  scheme: scheme.TCP,
  timeout: "15s",
  tls: {
    ca: "/certs/ca.pem",
    cert: "/certs/cert.pem",
    key: "/certs/key.pem",
  },
});
```

`tls` is three file paths: `ca`, `cert`, and `key`.

## API version

```ts
const cli = connect({
  fromEnv: true,
  apiVersion: "1.47",
  negotiateAPIVersion: true,
});
```

`apiVersion` pins the Docker API version as a string. `negotiateAPIVersion` asks the session to agree on a version with the daemon. One `ping` call can negotiate for itself by passing `{ negotiateAPIVersion: true }` or `{ forceNegotiate: true }`.

## Settings you can write down

`connect` keeps values that can be written as data: strings, numbers, booleans, and plain objects. A function, an open socket, or a stream is not a client setting. Leave those out. Command output comes back from the method you awaited, or from the value `run` returns in a script that does one thing and exits.

| Field | Type | Role |
| --- | --- | --- |
| `fromEnv` | `boolean` | Host and certificates come from the environment |
| `host` | `string` | Daemon URL |
| `hostFromEnv` | `boolean` | Read only the host from the environment |
| `scheme` | `string` | One of the `scheme` values above |
| `timeout` | `number \| string` | How long to wait while connecting |
| `tls` | `{ ca?, cert?, key? }` | Certificate file paths |
| `tlsFromEnv` | `boolean` | Read only the certificates from the environment |
| `apiVersion` | `string` | Pinned API version |
| `apiVersionFromEnv` | `boolean` | Read only the API version from the environment |
| `negotiateAPIVersion` | `boolean` | Agree on an API version with the daemon |
| `dryRun` | `boolean` | Ask every command using this client to skip changes |
| `maxConcurrency` | `number` | How many operations in one description may run at once |
| `userAgent` | `string` | Sent with requests |
| `httpHeader` | `{ name?, value? }` or a list | Extra request headers |

`dryRun` applies to every command that receives this client.

## One client, many calls

`root` stores the settings once. Every command in `App` is a method on `app`. `useSession()` is legal only while `App` is rendering, so this function is `App`, not a loose module.

```tsx
import { Compose, Service } from "containerkit";
import { root, useMemo, useSession } from "containerkit/live";

export default root({ fromEnv: true }, function App() {
  const app = useSession();
  const ready = useMemo(() => app.pullImage("alpine:latest"), ["alpine:latest"]);
  if (!ready) return null;
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" />
    </Compose>
  );
});
```

`listImages`, `composePs`, and `composeStop` are methods on that same `app`. Call them from a function `App` calls, or from `useMemo` when the result has to survive the next render:

```ts
import { Compose, Service } from "containerkit";
import type { Session } from "containerkit/live";

function stack() {
  return Compose({
    name: "app",
    children: Service({ name: "api", image: "alpine:latest" }),
  });
}

export async function inspectApi(app: Session) {
  const images = await app.listImages({ all: true });
  const running = await app.composePs(stack(), { services: "api" });
  await app.composeStop(stack(), { services: "api", timeout: "10s" });
  return { images, running };
}
```

`images` is the value `listImages` returned. `App` calls `inspectApi(app)` and passes the session in. `app` is not a prop of `<Compose>`.

## A client setting is not a command setting

`connect` accepts host, timeout, and certificates. `composeUp` accepts `detach`, `wait`, and its own `timeout`, which is how long that up may take. TypeScript rejects `detach` on `connect` and rejects `host` on `composeUp`. The two objects are different on purpose, so a connection setting and a command setting cannot be swapped by accident.
