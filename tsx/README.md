# containerkit

Write Compose files, Dockerfiles, and container actions as one TypeScript framework. A program has four parts: a JSX tree, vocabulary values, a client, and commands. It is not a web UI and it does not open a Docker socket from the browser.

The same words (`name`, `timeout`, `Env`) show up in three places. Where you write them is the scope:

| Kind | TypeScript |
| --- | --- |
| Define (document) | JSX props / children on `DockerNode` |
| Nested define | `<Health timeout="5s" />` |
| Client | `connect({ host, timeout })` |
| Compose/Engine verbs | `await app.composeUp(node, { detach, timeout })` |

A module is a default-export function that returns a `DockerNode`. The type argument is the component and its props: `DockerNode<Compose<{ name?: string }>>`, or the same form for `Service`, `Job`, and `Container`. A Service node is not a Compose node. `exportYAML` prints a named `Compose` node as YAML. A program that stays up default-exports `root()` from `docker.ts` and calls `await app.composeUp(tree, { detach: true })`.

## Install

Requires Node 20 or newer.

```bash
npm install containerkit
```

`jsx: "react-jsx"` is the automatic JSX transform. The import source decides which runtime a file uses. React and containerkit can live in one project. They cannot share a file: each `.tsx` file has one import source, so `<div>` and `<Compose>` stay apart.

Leave the app tsconfig on React and exclude the containerkit folder:

```json
{ "exclude": ["infra"] }
```

Give that folder a config that extends `containerkit/tsconfig.jsx.json` (it sets `jsxImportSource` to `containerkit`):

```json
{
  "extends": ["../tsconfig.json", "containerkit/tsconfig.jsx.json"],
  "include": ["./**/*.ts", "./**/*.tsx"],
  "compilerOptions": { "noEmit": true }
}
```

A file can opt in without a second config. The first line overrides the project setting:

```tsx
/** @jsxImportSource containerkit */
```

A React file in a project pointed at containerkit opts back with `/** @jsxImportSource react */`.

`jsx-runtime` exports its own `JSX` namespace. If that namespace is missing, or `jsxImportSource` does not resolve, TypeScript looks up a global namespace named `React`, because `react-jsx` defaults the module specifier to `"react"`. Under containerkit's namespace, HTML tags are not intrinsic elements, so `<div>` is a type error. `JSX.Element` is intentionally unset so each component keeps its own `$$kind`.

Put `containerkit` on `PATH`, or set `CONTAINERKIT_BIN` to the binary. [Runner binary](https://github.com/aptd3v/containerkit/blob/main/docs/runner-binary.md).

## Example

```tsx
/** @jsxImportSource containerkit */
import {
  Compose, Service, Publish, Volume, Health, Env, Network, Deploy, Restart,
  exportYAML,
  type DockerNode,
} from "containerkit";
import { cap, netDriver, proto, restartPolicy } from "containerkit/vocab";

const prod = process.env.NODE_ENV === "production";

export default function App(): DockerNode<Compose<{ name?: string }>> {
  return (
    <Compose name="app">
      <Network name="front" driver={netDriver.Bridge} />
      <Volume name="data" />
      <Service name="api" image="alpine:latest" capAdd={cap.NET_ADMIN}>
        <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
        <Env name="APP">1</Env>
        <Volume source="data" target="/data" />
        <Restart policy={restartPolicy.UnlessStopped} />
        <Health test={["CMD", "true"]} interval="5s" retries={3} />
        {prod && <Deploy replicas={2} />}
      </Service>
    </Compose>
  );
}

export const yaml = exportYAML(<App />);
```

`exportYAML` needs a `Compose` node with a `name`. It does not start or stop anything. A session is `useSession()` inside `docker.ts`. Never put that session on a `DockerNode`.

Control flow is ordinary JavaScript: `&&`, `.map()`, function components and props.

A standalone `<Container>` or `<Service>` renders a one-service compose file.

## Dockerfile

Instruction components live in `containerkit/dockerfile` so names like `Env` and `Volume` are the Dockerfile instructions, not the compose ones. `exportDockerfile` emits Dockerfile text. A `<Dockerfile>` child of `<Build>` becomes `dockerfile_inline`. `buildDockerfile` puts that text on `buildImage` as `parseDockerfile`.

```tsx
/** @jsxImportSource containerkit */
import { Build, Compose, Service, connect, buildDockerfile } from "containerkit";
import { Command, Dockerfile, From, Run } from "containerkit/dockerfile";

const image = (
  <Dockerfile>
    <From image="alpine" tag="latest" />
    <Run cmd="echo ok" />
    <Command executable="sleep" args={["infinity"]} />
  </Dockerfile>
);

export const compose = (
  <Compose name="app">
    <Service name="api">
      <Build>{image}</Build>
    </Service>
  </Compose>
);

export const plan = buildDockerfile(connect({ fromEnv: true }), image, { image: "app:dev" });
```

## Parent is the scope

`<Network>`, `<Volume>`, `<Secret>`, and `<FileConfig>` mean an attachment under a service and a project resource under `<Compose>`. `render` switches on the parent kind.

Scalars are props. Nested scopes are children. `Env` / `Label` use `name` (JSX `key` is reserved). Each component's `kind` prop is the node kind.

## Vocabulary

Import these from `containerkit/vocab`. They are the legal words for a prop or a command option. Pass the object instead of a raw string. Every group is also on `vocab`, which lists them all. `pullPolicy`, `restartPolicy`, `waitFor`, and `watchAction` use the longer name so they are not the `pull`, `restart`, `wait`, and `watch` commands.

```ts
import { cap, proto, restartPolicy } from "containerkit/vocab";
```

| Export | Use for |
| --- | --- |
| `cap` | Linux capabilities (`cap.NET_ADMIN`, `cap.ALL`) |
| `proto` | Port protocol: `TCP`, `UDP`, `SCTP` |
| `restartPolicy` | `No`, `OnFailure`, `Always`, `UnlessStopped` |
| `pullPolicy` | Image pull policy, plus `pullPolicy.every("24h")` |
| `netmode` | `networkMode`: `Bridge`, `Host`, `None`, `container(id)`, `service(name)` |
| `netDriver` | Network `driver`: bridge, host, none, overlay, macvlan, ipvlan |
| `logDriver` | Logging `driver`, including `JSONFile` |
| `portmode` | Published port `mode`: `Ingress`, `Host` |
| `vol` | Volume `kind`: bind, volume, tmpfs, npipe, cluster, image |
| `propagate` | Bind `propagation` |
| `selinux` | Bind relabel: `Shared` (`z`), `Private` (`Z`) |
| `recursive` | Bind `recursive` |
| `ipc`, `pidns`, `userns`, `uts`, `cgroupns` | Namespace modes. `ipc.container(id)` and `pidns.container(id)` join another container |
| `isolation` | `default`, `process`, `hyperv` |
| `signal` | Stop and kill signals (`signal.SIGTERM`) |
| `depend` | `DependsOn` condition |
| `svcmode`, `endpoint`, `failact`, `order` | Deploy mode, endpoint mode, update failure action, update order |
| `cascade` | `onExit` for `composeUp` and `composeStart`: `Ignore` `0`, `Stop` `1`, `Fail` `2` |
| `progress` | `composeUp` progress mode |
| `scheme` | Client URL scheme on `connect` |
| `avail`, `role` | Swarm node availability and role |
| `df` | `diskUsage` type filter |
| `oci` | `publishOCI` manifest version |
| `waitFor` | `wait` condition |
| `watchAction` | Develop watch `action` |

The value behind each name, and which prop it belongs on, is in [GUIDE.md](GUIDE.md). Client settings, command argument positions, trees, and Dockerfiles are in [docs](docs/README.md).

## Commands

A command is a method on the session. Options are the last argument (`await app.composeUp(tree, { detach: true })`). Compose commands start with `compose`. Container and engine commands do not.

Project commands take the tree: `composeUp`, `composeDown`, `composeStart`, `composeStop`, `composeRestart`, `composePause`, `composeUnpause`, `composeKill`, `composePs`, `composePull`, `composePush`, `composeBuild`, `composeCreate`, `composeRemove`, `composeScale`, `composeWait`, `composeWatch`, `composeEvents`, `composeImages`, `composeList`, `composePort`, `composeCopy`, `composeGenerate`, `composeViz`, `composeVolumeList`, `composeLoadProject`, `composePublishOCI`, `composeAttach`, `composeCommit`, `composeExec`, `composeExport`, `composeRun`.

Container commands take an id: `create`, `start`, `stop`, `restart`, `kill`, `remove`, `renameContainer`, `updateContainer`, `logs`, `statsContainer`, `topContainer`, `wait`, `resizeContainer`, `copyFrom`, `copyTo`, `statPath`, `attachContainer`, `commitContainer`, `listContainers`, `containerRun`, `run`. Exec: `createExec`, `startExec`, `attachExec`, `resizeExec`. Checkpoints: `createCheckpoint`, `listCheckpoints`, `removeCheckpoint`.

Images: `buildImage`, `buildDockerfile`, `pullImage`, `pushImage`, `listImages`, `inspectImage`, `removeImage`, `tagImage`, `historyImage`, `searchImages`, `saveImages`, `attestImage`, `pruneImages`.

Networks: `createNetwork`, `listNetworks`, `inspectNetwork`, `connectNetwork`, `disconnectNetwork`, `pruneNetworks`.

Volumes: `createVolume`, `listEngineVolumes`, `removeVolume`, `updateVolume`, `pruneVolumes`.

Swarm: `initSwarm`, `joinSwarm`, `leaveSwarm`, `unlockSwarm`, `updateSwarm`, `createSwarmService`, `listSwarmServices`, `updateSwarmService`, `swarmServiceLogs`, `listNodes`, `updateNode`, `removeNode`, `listTasks`, `taskLogs`, `createEngineConfig`, `listConfigs`, `updateEngineConfig`, `createEngineSecret`, `listSecrets`, `updateEngineSecret`.

Plugins: `createPlugin`, `installPlugin`, `enablePlugin`, `disablePlugin`, `listPlugins`, `setPlugin`, `upgradePlugin`, `pushPlugin`, `removePlugin`.

Host: `ping`, `login`, `diskUsage`, `inspect`, `listenEvents`, `load`, `pruneBuildCache`, `pruneContainers`.

`composeUp` is the project command. Signatures and the options you will actually set are in [GUIDE.md](GUIDE.md).

## Run

A project stays live. `docker.ts` default-exports `root()` from `containerkit/live`. Commands are methods on that session: `await app.pullImage("alpine:latest")`. `exportYAML` prints a named `Compose` node as YAML at any time. It does not stop or start the session. See [Live session](docs/live.md).
