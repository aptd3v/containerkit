# containerkit

Write Compose files, Dockerfiles, and container actions as one TypeScript framework. A program has four parts: a JSX tree, vocabulary values, a client, and commands. It is not a web UI and it does not open a Docker socket from the browser.

This package is the TypeScript surface. The Go library is [containerkit](https://github.com/aptd3v/containerkit). Components and session methods are generated from that kit, from pinned Moby and Compose SDK versions. Updating the pins and regenerating picks up new upstream fields.

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
npm install @aptd3v/containerkit
```

The package includes the runner. The `containerkit` command selects the binary for this machine.

`jsx: "react-jsx"` is the automatic JSX transform. The import source decides which runtime a file uses. React and containerkit can live in one project. They cannot share a file: each `.tsx` file has one import source, so `<div>` and `<Compose>` stay apart.

Leave the app tsconfig on React and exclude the containerkit folder:

```json
{ "exclude": ["infra"] }
```

Give that folder a config that extends `@aptd3v/containerkit/tsconfig.jsx.json` (it sets `jsxImportSource` to `@aptd3v/containerkit`):

```json
{
  "extends": ["../tsconfig.json", "@aptd3v/containerkit/tsconfig.jsx.json"],
  "include": ["./**/*.ts", "./**/*.tsx"],
  "compilerOptions": { "noEmit": true }
}
```

A file can opt in without a second config. The first line overrides the project setting:

```tsx
/** @jsxImportSource @aptd3v/containerkit */
```

A React file in a project pointed at `@aptd3v/containerkit` opts back with `/** @jsxImportSource react */`.

`jsx-runtime` exports its own `JSX` namespace. If that namespace is missing, or `jsxImportSource` does not resolve, TypeScript looks up a global namespace named `React`, because `react-jsx` defaults the module specifier to `"react"`. Under containerkit's namespace, HTML tags are not intrinsic elements, so `<div>` is a type error. `JSX.Element` is intentionally unset so each component keeps its own `$$kind`.

Put `containerkit` on `PATH`, or set `CONTAINERKIT_BIN` to the binary. [Runner binary](https://github.com/aptd3v/containerkit/blob/main/docs/runner-binary.md).

## ESLint

TypeScript checks props and the import source. It does not check which component may contain which child, so this typechecks:

```tsx
<Service name="api" image="alpine">
  <Dockerfile>
    <From image="nginx" tag="alpine" />
  </Dockerfile>
</Service>
```

Building that tree throws `containerkit/tsx: <dockerfile> is not valid inside <service>`. The names in the throw are the internal kinds, and the throw happens when the tree is built, not while you are editing. The `containerkit/nesting` rule marks the tag you wrote: `<Dockerfile> is not valid inside <Service>`. It also marks a second child where the parent accepts one, such as a second `<Health>` on a service.

The rule only sees a tag written inline. A child kept in a variable, returned from a helper, or spread onto the parent is invisible to it, and building the tree still throws. Installing `@aptd3v/containerkit` does not enable the rule.

ESLint 10 is an optional peer. `typescript-eslint` supplies the parser that reads TSX. Both are dev dependencies:

```bash
npm install --save-dev eslint typescript-eslint
```

`recommended` turns `containerkit/nesting` on for `**/*.{jsx,tsx}`. The parser config has to cover those files:

```js
import { defineConfig } from "eslint/config";
import tseslint from "typescript-eslint";
import { recommended } from "@aptd3v/containerkit/eslint";

export default defineConfig(
  { files: ["**/*.{jsx,tsx}"], languageOptions: { parser: tseslint.parser } },
  ...recommended,
);
```

## Example

```tsx
/** @jsxImportSource @aptd3v/containerkit */
import {
  Compose, Service, Publish, Volume, Health, Env, Network, Deploy, Restart,
  exportYAML,
  type DockerNode,
} from "@aptd3v/containerkit";
import { cap, netDriver, proto, restartPolicy } from "@aptd3v/containerkit/vocab";

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

## Programs you can run

These programs are in the git repository. They are not part of the npm package. Clone the repo, `cd tsx`, and run the script. Docker has to be running.

| Script | Program |
| --- | --- |
| `npm run nginx` | [nginx](https://github.com/aptd3v/containerkit/blob/main/tsx/examples/nginx/docker.tsx) builds a site and publishes `http://127.0.0.1:8080` |
| `npm run wordpress` | [WordPress](https://github.com/aptd3v/containerkit/blob/main/tsx/examples/wordpress/docker.tsx) runs WordPress and MySQL |
| `npm run mongo-replica` | [MongoDB](https://github.com/aptd3v/containerkit/blob/main/tsx/examples/mongo-replica/docker.tsx) starts eight members, initiates the set, then adds Mongo Express |
| `npm run terminal` | [terminal](https://github.com/aptd3v/containerkit/blob/main/tsx/examples/terminal/terminal.tsx) attaches a shell to an Alpine container |

The first three are render callbacks: the function returns the tree, and a later state change applies again. `terminal` is `async` because attaching the shell has to be awaited. Ctrl+C on the first three takes the project down.

`false`, `null`, and `undefined` children are dropped. `.map()` repeats a child. A function that returns a node is a component. `&&` and props are ordinary JavaScript.

A standalone `<Container>` or `<Service>` renders a one-service compose file.

## Dockerfile

Instruction components live in `@aptd3v/containerkit/dockerfile` so names like `Env` and `Volume` are the Dockerfile instructions, not the compose ones. `exportDockerfile` emits Dockerfile text. `<Run>` joins into one layer. `<RunLayer>` starts the next. A `<Dockerfile>` child of `<Build>` becomes `dockerfile_inline`. `buildDockerfile` puts that text on `buildImage` as `parseDockerfile`. [`examples/nginx/docker.tsx`](https://github.com/aptd3v/containerkit/blob/main/tsx/examples/nginx/docker.tsx) builds that kind of file and publishes the site.

```tsx
/** @jsxImportSource @aptd3v/containerkit */
import { Build, Compose, Service, connect, buildDockerfile } from "@aptd3v/containerkit";
import { Command, Dockerfile, From, Run } from "@aptd3v/containerkit/dockerfile";

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

Import these from `@aptd3v/containerkit/vocab`. They are the legal words for a prop or a command option. Pass the object instead of a raw string. Every group is also on `vocab`, which lists them all. `pullPolicy`, `restartPolicy`, `waitFor`, and `watchAction` use the longer name so they are not the `pull`, `restart`, `wait`, and `watch` commands.

```ts
import { cap, proto, restartPolicy } from "@aptd3v/containerkit/vocab";
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

Swarm: `initSwarm`, `joinSwarm`, `leaveSwarm`, `unlockSwarm`, `updateSwarm`, `inspectSwarm`, `createSwarmService`, `listSwarmServices`, `inspectSwarmService`, `updateSwarmService`, `swarmServiceLogs`, `listNodes`, `inspectNode`, `updateNode`, `removeNode`, `listTasks`, `inspectTask`, `taskLogs`, `createEngineConfig`, `listConfigs`, `inspectConfig`, `updateEngineConfig`, `createEngineSecret`, `listSecrets`, `inspectSecret`, `updateEngineSecret`.

Plugins: `createPlugin`, `installPlugin`, `inspectPlugin`, `enablePlugin`, `disablePlugin`, `listPlugins`, `setPlugin`, `upgradePlugin`, `pushPlugin`, `removePlugin`.

Host: `ping`, `login`, `diskUsage`, `inspect`, `listenEvents`, `load`, `pruneBuildCache`, `pruneContainers`.

`composeUp` is the project command. Signatures and the options you will actually set are in [GUIDE.md](GUIDE.md).

## It stays open

`docker.ts` default-exports `root()` from `@aptd3v/containerkit/live`. The session stays up for the life of the process. A later render applies the project only when the rendered YAML changes. `exportYAML` prints a named `Compose` node as YAML at any time. Exporting the file does not stop the session. [`examples/nginx/docker.tsx`](https://github.com/aptd3v/containerkit/blob/main/tsx/examples/nginx/docker.tsx) is a running copy: from `tsx/`, `npm run nginx` builds an nginx image, serves `http://127.0.0.1:8080`, and `onExit: "down"` takes the project down on Ctrl+C.

```tsx
import { Compose, Service, Publish, Health } from "@aptd3v/containerkit";
import { proto } from "@aptd3v/containerkit/vocab";
import { root } from "@aptd3v/containerkit/live";

export default root({ fromEnv: true }, function App() {
  return (
    <Compose name="app">
      <Service
        name="api"
        image="alpine:latest"
        command={["sh", "-c", "mkdir -p /www && echo ok > /www/index.html && exec httpd -f -p 80 -h /www"]}
      >
        <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
        <Health test={["CMD", "wget", "-q", "-O", "/dev/null", "http://127.0.0.1/"]} interval="5s" retries={3} />
      </Service>
    </Compose>
  );
});
```

Commands are methods on that session. `await app.pullImage("alpine:latest")` runs, then the next line runs. `useMemo` writes the result to `.containerkit/memo` and skips the call the next time the same inputs come back. See [Live session](docs/live.md).
