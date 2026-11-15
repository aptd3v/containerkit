# Commands

In `docker.ts`, a command is a method on the session. `await` runs the next line after the previous result.

```ts
await app.pullImage("alpine:latest");
await app.login({ serverAddress: "ghcr.io" });
await app.composeUp(tree, { detach: true });
```

`app` is `useSession()`, or an argument typed `Session`. The session already has the client, so you do not pass one. Building that session is [Live session](live.md). Choosing the daemon is [Client](client.md).

The same names are also functions. `pullImage(cli, ref)` returns a description and does not talk to Docker. A script that does one thing and exits passes that description to `run`. Inside the session, use the method. The full prop list is in [GUIDE.md](../GUIDE.md).

Compose methods start with `compose`. Methods for one container, image, network, or the host keep a short name: `start`, `pullImage`, `createNetwork`.

`timeout` is a number of seconds or a duration string such as `"30s"`. Where a method takes `services`, pass one service name or a list of names. A value such as `proto.TCP`, `netDriver.Bridge`, or `oci.V1_1` comes from `containerkit/vocab`. [GUIDE.md](../GUIDE.md) lists each group. [Trees](trees.md) shows `proto` and `cap` on a component.

## Where each argument goes

| You write | What you pass after the name |
| --- | --- |
| `app.composeUp(tree, { detach: true })` | The tree, then options. The tree becomes the Compose file |
| `app.composeExec(tree, "api", { command: ["sh"] })` | The tree, the service name, the options |
| `app.composePort(tree, "api", 80, { protocol: "tcp" })` | The tree, the service name, the port |
| `app.composePublishOCI(tree, "ghcr.io/acme/app", { oci: oci.V1_1 })` | The tree, the repository |
| `app.stop(id, { timeout: "10s" })` | A container id |
| `app.pullImage("alpine:latest", { platform: "linux/arm64" })` | An image reference |
| `app.createNetwork("front", { driver: netDriver.Bridge })` | A network name |
| `app.load("compose.yaml", { name: "app" })` | A file path |
| `app.saveImages(["alpine:latest"])` | A list of image references |
| `app.searchImages("alpine")` | A search term |
| `app.listImages({ all: true })`, `app.ping()` | Options, or nothing |

Options stay on the command. They are not written back onto the tree. `detach` does not become a prop of `<Compose>`.

## Project

These commands take the tree, the same `Compose` (or service, job, or container) you would return from `App`.

```ts
import { cascade, progress, proto } from "containerkit/vocab";

await app.composeUp(tree, {
  detach: true,
  wait: true,
  timeout: "30s",
  services: ["api"],
  pull: true,
  removeOrphans: true,
  noCache: false,
  onExit: cascade.Ignore,
  progress: progress.Plain,
});

await app.composePs(tree, { services: "api", all: true });

const published = await app.composePort(tree, "api", 80, {
  protocol: proto.TCP,
  index: 0,
});

await app.composeStop(tree, { services: "api", timeout: "10s" });
await app.composeDown(tree, { volumes: true, removeOrphans: true, timeout: "20s" });
```

`composeUp` applies the project. The session does this for you when `App` returns a tree whose YAML changed. Call it yourself when you want those options, such as `wait` or a service list.

`onExit` is what to do when a container exits: `cascade.Ignore`, `cascade.Stop`, or `cascade.Fail`. `progress` is how progress is reported: `progress.Auto`, `progress.Plain`, `progress.TTY`, `progress.Quiet`, or `progress.JSON`.

`composeDown` removes the project. `volumes: true` also removes named volumes. `removeImages` is `"local"` or `"all"`. `composeStart` takes `services`, `wait`, `timeout`, `onExit`, and `attachTo`. `composeRestart` adds `noDeps`. `composeKill` takes `services`, `stopSignal` (`signal.SIGKILL`), and `all`.

`composeBuild` builds services that have a build block. Options include `dockerfilePath`, `tag`, `target`, `noCache`, `pull`, `platform`, and `memory`. `composePull` takes `ignoreBuildable`, `ignoreFailures`, and `quiet`. `composePush` takes `ignoreFailures` and `quiet`.

`composeWait` waits until the named services match. Options are `services` and `downOnExit`. `composeExec`, `composeRun`, `composeAttach`, `composeCommit`, and `composeExport` take the tree, the service name, then options: `await app.composeExec(tree, "api", { command: ["sh"] })`.

`await app.composeLoadProject(tree, { name, file, offline, interpolation })` reads the tree as a project. It does not start containers. `await app.load(path, options)` reads a Compose file from disk. The path is the first argument. The same load options apply.

## Containers

`containerRun` describes one container from options: `image`, `command`, `env`, `publish`, `volume`, `restart`, and the other container fields. It does not take a tree. The session does not keep it running across renders unless you remember the call with `useMemo`.

```ts
import { cap, proto, restartPolicy } from "containerkit/vocab";

await app.containerRun({
  image: "alpine:latest",
  name: "once",
  command: ["echo", "ok"],
  capAdd: cap.NET_ADMIN,
  publish: { hostPort: 8080, containerPort: 80, protocol: proto.TCP },
  restart: { policy: restartPolicy.No },
});
```

A container you already have is addressed by id:

```ts
import { signal, waitFor } from "containerkit/vocab";

const id = "container-id";

await app.start(id, { checkpointDir: "/checkpoints" });
await app.logs(id, { follow: true, tail: "100", timestamps: true, attachStdout: true });
await app.wait(id, { waitFor: waitFor.NextExit });
await app.stop(id, { timeout: "10s", stopSignal: signal.SIGTERM });
await app.kill(id, { stopSignal: signal.SIGKILL });
await app.remove(id, { force: true, volumes: true });
```

`start` also takes `checkpoint`. `stop` and `restart` take `timeout` and `stopSignal`. `logs` takes `follow`, `tail`, `since`, `until`, `timestamps`, `details`, `attachStdout`, and `attachStderr`. `wait` takes `waitFor`: `waitFor.NotRunning`, `waitFor.NextExit`, or `waitFor.Removed`.

`await app.listContainers({ all: true, limit: 20, since: id, filter: { term: "status", values: ["running"] } })` lists containers. A filter is `{ term, values }`. `values` is a string or a list. Pass one filter or an array of them.

`await app.createExec(id, { command: ["sh"], tty: true, user: "root", attachStdout: true })` prepares a command inside a running container. `startExec`, `attachExec`, and `resizeExec` continue that session. `copyTo` and `copyFrom` move files. `statPath` reads a path inside the container.

## Images

```ts
import { Command, Dockerfile, From } from "containerkit/dockerfile";

await app.pullImage("alpine:latest", { platform: "linux/arm64", all: false });
await app.listImages({ all: true, filter: { term: "reference", values: ["alpine*"] } });
await app.inspectImage("alpine:latest");

await app.buildImage({
  image: "app:dev",
  source: ".",
  dockerfilePath: "Dockerfile",
  target: "runtime",
  pull: true,
  noCache: false,
  platform: "linux/arm64",
});

const file = (
  <Dockerfile>
    <From image="alpine" tag="latest" />
    <Command executable="sleep" args={["infinity"]} />
  </Dockerfile>
);
await app.buildDockerfile(file, { image: "app:dev", tag: "app:dev" });

await app.pushImage("ghcr.io/acme/app:dev");
await app.removeImage("app:dev");
```

`buildDockerfile` turns the Dockerfile tree into the build and builds it. `buildImage` points at a Dockerfile that is already on disk. `source` is the build context. `target` selects a stage. The Dockerfile page explains the tree.

`tagImage`, `historyImage`, `searchImages`, `saveImages`, `attestImage`, and `pruneImages` take a reference, a search term, or a list of references, then options. `await app.pruneImages({ all: true, filter })` drops unused images.

## Networks and volumes

A `<Network>` or `<Volume>` in the tree is part of the Compose project. These commands create the same kinds of objects directly on the daemon, without a project file.

```ts
import { netDriver } from "containerkit/vocab";

await app.createNetwork("front", {
  driver: netDriver.Bridge,
  attachable: true,
  internal: false,
  label: { name: "app", value: "demo" },
});
await app.connectNetwork("front", { containerID: id });
await app.listNetworks({ filter: { term: "name", values: ["front"] } });
await app.pruneNetworks();

await app.createVolume({ name: "data", driver: "local" });
await app.listEngineVolumes();
await app.removeVolume("data", { force: true });
await app.pruneVolumes({ all: true });
```

`createNetwork` options are `driver`, `attachable`, `internal`, `label`, and `driverOpt`. Subnet, gateway, and `enableIPv6` belong on a `<Network>` in the tree, not on this call. `disconnectNetwork` takes `containerID` and `force`. `inspectNetwork` takes `verbose`.

## The host, Swarm, and plugins

```ts
import { avail, df, role } from "containerkit/vocab";

await app.ping({ negotiateAPIVersion: true });
await app.login({ serverAddress: "ghcr.io", username: "acme" });
await app.diskUsage({ images: true, volumes: true, buildCache: true, usage: df.Image });
await app.listenEvents({ since: "2026-01-01T00:00:00Z", filter: { term: "type", values: ["container"] } });
await app.inspect(id);

await app.initSwarm({ advertiseAddr: "192.0.2.10:2377" });
await app.listNodes();
await app.updateNode(nodeId, { availability: avail.Drain, role: role.Worker });

await app.installPlugin("plugin-name", { acceptAllPermissions: true });
await app.enablePlugin("plugin-name", { timeout: "30s" });
await app.listPlugins();
```

`diskUsage` uses `images`, `volumes`, and `buildCache` to choose which kinds to include. `usage` narrows the report to one kind: `df.Container`, `df.Image`, `df.Volume`, or `df.BuildCache`. `updateNode` also takes `name` and `versionIndex`.

Swarm services, tasks, configs, and secrets follow the same shape. An id, when the method has one, is the first argument. The names are `createSwarmService`, `listSwarmServices`, `updateSwarmService`, `swarmServiceLogs`, `listTasks`, `taskLogs`, `createEngineConfig`, `listConfigs`, `updateEngineConfig`, `createEngineSecret`, `listSecrets`, and `updateEngineSecret`.

`leaveSwarm`, `joinSwarm`, `unlockSwarm`, and `updateSwarm` take options: `await app.leaveSwarm()`. Plugin methods that name a plugin (`installPlugin`, `enablePlugin`, `disablePlugin`, `setPlugin`, `upgradePlugin`, `pushPlugin`, `removePlugin`) take the plugin name, then options. `createPlugin` and `listPlugins` take options.

`pruneContainers` and `pruneBuildCache` take options. `pruneBuildCache` accepts `all` and `filter`.

## One command, then exit

`run` sends one description and returns the result, without a live session. A test or a short script can use it. A program that stays up uses `await app.pullImage(...)` and the other session methods.

```ts
import { connect, pullImage, run } from "containerkit";

const cli = connect({ fromEnv: true });
await run(pullImage(cli, "alpine:latest"));
```

`setInvoke` replaces what `run` does, which is how a test avoids starting the `containerkit` program. Session methods are not redirected by `setInvoke`.

```ts
import { connect, pullImage, run, setInvoke } from "containerkit";

const cli = connect({ fromEnv: true });
setInvoke(async (_op, description) => description);
await run(pullImage(cli, "alpine:latest"));
setInvoke(undefined);
```
