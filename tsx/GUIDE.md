# containerkit guide

containerkit is a TypeScript framework for Compose files, Dockerfiles, and the actions that run them. You describe a stack as JSX. A live session in `docker.ts` runs the actions and stays open. `exportYAML` prints a named `Compose` node as Compose YAML without stopping that session.

This is not a web UI. JSX here builds a document, not HTML. Setup for TypeScript, including a project that also uses React, is in [README.md](README.md).

The pages under [docs](docs/README.md) go further than this guide:

| Page | What it covers |
| --- | --- |
| [Client](docs/client.md) | `connect`, host, TLS, timeouts, one client across many calls |
| [Commands](docs/commands.md) | Argument positions and worked calls |
| [Trees](docs/trees.md) | `DockerNode<Component<{ props }>>`, children, `render` |
| [Dockerfiles](docs/dockerfile.md) | Instruction components, `CMD` forms, `buildDockerfile` |
| [Live session](docs/live.md) | `docker.ts`, state that restarts a stack, the `.containerkit/memo` cache, error checks |

## Shape of a program

```tsx
/** @jsxImportSource containerkit */
import {
  Compose, Service, Publish, Restart,
  exportYAML,
  type DockerNode,
} from "containerkit";
import { cap, proto, restartPolicy } from "containerkit/vocab";

export function stack(): DockerNode<Compose<{ name?: string }>> {
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" capAdd={cap.NET_BIND_SERVICE}>
        <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
        <Restart policy={restartPolicy.Always} />
      </Service>
    </Compose>
  );
}

export const yaml = exportYAML(stack());
```

`exportYAML` reads the `Compose` node and returns the file text. The node must be `Compose` and it must have `name`. Nothing is started, stopped, or restarted. A session that is already running keeps running. Running the stack is the live session in [Live session](docs/live.md).

`connect` does not open a connection by itself. It records how the runner should connect. Do not put `cli` on a `DockerNode`. Pass it only to commands.

## Several calls, one client

`connect` once. Pass that `cli` into each command, then `run` the plan. The tree function stays synchronous and returns a `DockerNode`. Calls that talk to the daemon sit beside it.

```ts
import { connect, run, listImages, pullImage, composeUp, composePs, composeStop } from "containerkit";

const cli = connect({ fromEnv: true, timeout: "10s" });

const images = await run(listImages(cli, {
  all: true,
  filter: { term: "reference", values: ["alpine*"] },
}));
await run(pullImage(cli, "alpine:latest", { platform: "linux/arm64" }));
await run(composeUp(cli, stack(), { detach: true, wait: true, timeout: "30s", services: ["api"] }));
await run(composePs(cli, stack(), { services: "api" }));
await run(composeStop(cli, stack(), { services: "api", timeout: "10s" }));
```

`images` is the runner's result for that plan. `filter` is `{ term, values }`. `values` is a string or a list. Host, TLS, and timeout live on `cli`. `detach`, `wait`, and `services` live on the command. Field reference and more calls are in [Client](docs/client.md) and [Commands](docs/commands.md).

## The tree

A module exports a function that returns a `DockerNode`. The type argument is the component and its props, the same shape as a React element type: `DockerNode<Compose<{ name?: string }>>`, or `DockerNode<Service<{ name?: string; image?: string }>>`, and the same form for `Job` and `Container`. With no argument, `DockerNode` is any of those four roots. A `Service` node is not a `Compose` node, and `image` is not a Compose prop. A lone `Service` or `Container` renders as a one-service Compose file. `Compose({ name: "app" })` and `jsx(Compose, { name: "app" })` return that compose node, so a `Service` value cannot be returned from `DockerNode<Compose<{ name?: string }>>`. How children, parents, and `render` fit together is in [Trees](docs/trees.md).

Children are nested scopes. Props are scalars. `Env` and `Label` take `name` because JSX reserves `key`.

Under `<Compose>`, `Network`, `Volume`, `Secret`, and `FileConfig` declare project resources. Under `<Service>`, the same components attach that resource to the service. `render` chooses from the parent.

| Component | Role |
| --- | --- |
| `Compose` | Project root. `name`, and resource children |
| `Service` | A service. `name`, `image`, and most container fields |
| `Job` | A one-shot service |
| `Container` | A single container, or the container fields nested under a service |
| `Publish` | A published port. `hostPort`, `containerPort`, `protocol`, `mode` |
| `Expose` | An exposed port |
| `Volume` | A project volume or a service mount. `kind`, `source`, `target` |
| `Network` | A project network or a service attachment. `driver` |
| `Env`, `Label` | A name/value pair |
| `EnvFile` | An env file |
| `DependsOn` | Startup order. `condition` is `depend.Started`, `depend.Healthy`, or `depend.CompletedSuccessfully` |
| `Health` | Healthcheck. `test` starts with `"CMD"`, `"CMD-SHELL"`, or `"NONE"` |
| `Restart` | Restart policy. `policy` is a `restartPolicy` value |
| `Logging` | Log driver. `driver` is a `logDriver` value |
| `Build` | Image build. Nest a `<Dockerfile>` or set `dockerfilePath` |
| `Deploy` | Swarm deploy. `mode` is a `svcmode` value |
| `Update`, `Rollback` | Deploy update config. `failureAction` is `failact`, `order` is `order` |
| `Placement`, `Reserve` | Deploy constraints and resources |
| `Device`, `GPU`, `Ulimit` | Devices and limits |
| `Secret`, `FileConfig` | Secrets and configs |
| `ExtraHost`, `Sysctl`, `Annotation` | Extra hosts, sysctls, annotations |
| `Include` | Another Compose file |
| `Model` | A model attachment |
| `WatchFiles` | Develop watch. `action` is a `watchAction` value |
| `Filter`, `DriverOpt`, `IPAMOpt`, `StorageOpt`, `AuxAddress` | Option pairs on the parent that asks for them |
| `SSH`, `AdditionalContext` | Build extras |
| `TLS`, `HTTPHeader` | Client connection details on `connect` |
| `ConsoleSize`, `CopyPath`, `Discrete`, `ProviderOption`, `PreStart`, `PostStart`, `PreStop` | Narrow fields on the parent that accepts them |

## Dockerfile

Import instruction components from `containerkit/dockerfile`. Those names (`Env`, `Volume`, `Command`) are the Dockerfile forms, not the Compose components.

`exportDockerfile(file)` returns Dockerfile text. Nest the file under `<Build>` and it is sent as `dockerfile_inline`. `buildDockerfile(cli, file, { image })` builds that text with `buildImage`.

| Component | Writes |
| --- | --- |
| `From` | `FROM`. `image`, optional `tag`, `platform`, `as` |
| `Arg` | `ARG`. `name`, optional `default` |
| `Env` | `ENV` |
| `Copy`, `Add` | `COPY` / `ADD`. `src`, `dest`, plus flags such as `from`, `chown`, `chmod` |
| `Run` | `RUN`. Consecutive `Run`s with no flags fold into one layer |
| `RunLayer` | `RUN` that always starts a new layer |
| `RunArgs` | Extra arguments appended to the previous `RUN` |
| `Command` | `CMD` in exec form: `CMD ["server", "--port", "80"]` |
| `CommandShell` | `CMD` in shell form: `CMD server --port 80` |
| `Entrypoint`, `EntrypointShell` | `ENTRYPOINT`, exec or shell |
| `Shell` | `SHELL` in exec form |
| `Expose`, `Label`, `Volume`, `User`, `Workdir`, `StopSignal`, `Onbuild`, `Comment` | The matching instruction |
| `HealthCheck` | `HEALTHCHECK`. `test` starts with `"CMD"`, `"CMD-SHELL"`, or `"NONE"` |
| `HealthCheckNone` | `HEALTHCHECK NONE` |
| `Syntax`, `Escape`, `Check` | Parser directives. They must come before instructions |

A file can contain one `CMD`. A second `Command` or `CommandShell` throws. `CMD-SHELL` is only the healthcheck test prefix. It is written as a shell-form `CMD`, not as the word `CMD-SHELL`. Stages, layers, and the three ways to hand the text to a build are in [Dockerfiles](docs/dockerfile.md).

## Vocabulary

Import a group from `containerkit/vocab`, or import `vocab` and use `vocab.cap.NET_ADMIN`. The strings below are the values those props expect.

### cap

Linux capabilities for `capAdd` and `capDrop`.

`ALL`, `CHOWN`, `DAC_OVERRIDE`, `DAC_READ_SEARCH`, `FOWNER`, `FSETID`, `KILL`, `SETGID`, `SETUID`, `SETPCAP`, `LINUX_IMMUTABLE`, `NET_BIND_SERVICE`, `NET_BROADCAST`, `NET_ADMIN`, `NET_RAW`, `IPC_LOCK`, `IPC_OWNER`, `SYS_MODULE`, `SYS_RAWIO`, `SYS_CHROOT`, `SYS_PTRACE`, `SYS_PACCT`, `SYS_ADMIN`, `SYS_BOOT`, `SYS_NICE`, `SYS_RESOURCE`, `SYS_TIME`, `SYS_TTY_CONFIG`, `MKNOD`, `LEASE`, `AUDIT_WRITE`, `AUDIT_CONTROL`, `SETFCAP`, `MAC_OVERRIDE`, `MAC_ADMIN`, `SYSLOG`, `WAKE_ALARM`, `BLOCK_SUSPEND`, `AUDIT_READ`, `PERFMON`, `BPF`, `CHECKPOINT_RESTORE`.

### proto

`Publish` and `Expose` protocol.

| Name | Value |
| --- | --- |
| `TCP` | `tcp` |
| `UDP` | `udp` |
| `SCTP` | `sctp` |

### restartPolicy

`<Restart policy={...}>`.

| Name | Value |
| --- | --- |
| `No` | `no` |
| `OnFailure` | `on-failure` |
| `Always` | `always` |
| `UnlessStopped` | `unless-stopped` |

`OnFailure` plus `maxRetries` renders as `on-failure:<n>`.

### pullPolicy

Service `pullPolicy`.

| Name | Value |
| --- | --- |
| `Always` | `always` |
| `Never` | `never` |
| `Missing` | `missing` |
| `IfNotPresent` | `if_not_present` |
| `Build` | `build` |
| `Refresh` | `refresh` |
| `Daily` | `daily` |
| `Weekly` | `weekly` |

`pullPolicy.every("24h")` is `every_24h`.

### netmode

Service or container `networkMode`.

| Name | Value |
| --- | --- |
| `Bridge` | `bridge` |
| `Host` | `host` |
| `None` | `none` |
| `Default` | `default` |

`netmode.container(id)` is `container:<id>`. `netmode.service(name)` is `service:<name>`.

### netDriver

Network `driver`: `Bridge` `bridge`, `Host` `host`, `None` `none`, `Overlay` `overlay`, `Macvlan` `macvlan`, `IPvlan` `ipvlan`.

### logDriver

Logging `driver`: `JSONFile` `json-file`, `Local` `local`, `Syslog` `syslog`, `Journald` `journald`, `GELF` `gelf`, `Fluentd` `fluentd`, `AWSLogs` `awslogs`, `Splunk` `splunk`, `ETWLogs` `etwlogs`, `GCPLogs` `gcplogs`, `None` `none`.

### portmode

`Publish` `mode`: `Ingress` `ingress`, `Host` `host`.

### vol

Volume `kind`: `Bind` `bind`, `Volume` `volume`, `Tmpfs` `tmpfs`, `Pipe` `npipe`, `Cluster` `cluster`, `Image` `image`.

### propagate

Volume `propagation`: `Private` `private`, `RPrivate` `rprivate`, `Shared` `shared`, `RShared` `rshared`, `Slave` `slave`, `RSlave` `rslave`.

### selinux

Volume `seLinux`. `Shared` is `z`. `Private` is `Z`.

### recursive

Volume `recursive`: `Enabled` `enabled`, `Disabled` `disabled`, `Writable` `writable`, `ReadOnly` `readonly`.

### Namespaces

| Group | Values |
| --- | --- |
| `ipc` | `None` `none`, `Private` `private`, `Shareable` `shareable`, `Host` `host`. `ipc.container(id)` joins another container |
| `pidns` | `Host` `host`. `pidns.container(id)` joins another container |
| `cgroupns` | `Host` `host`, `Private` `private` |
| `userns` | `Host` `host` |
| `uts` | `Host` `host` |
| `isolation` | `Default` `default`, `Process` `process`, `HyperV` `hyperv` |

### signal

Stop and kill signals: `SIGHUP`, `SIGINT`, `SIGQUIT`, `SIGILL`, `SIGTRAP`, `SIGABRT`, `SIGBUS`, `SIGFPE`, `SIGKILL`, `SIGUSR1`, `SIGSEGV`, `SIGUSR2`, `SIGPIPE`, `SIGALRM`, `SIGTERM`, `SIGSTKFLT`, `SIGCHLD`, `SIGCONT`, `SIGSTOP`, `SIGTSTP`, `SIGTTIN`, `SIGTTOU`, `SIGURG`, `SIGXCPU`, `SIGXFSZ`, `SIGVTALRM`, `SIGPROF`, `SIGWINCH`, `SIGIO`, `SIGPWR`, `SIGSYS`.

Use `stopSignal={signal.SIGTERM}` on a service, or `kill(cli, tree, { stopSignal: signal.SIGKILL })`.

### depend

`DependsOn` `condition`: `Started` `service_started`, `Healthy` `service_healthy`, `CompletedSuccessfully` `service_completed_successfully`.

### Deploy

| Group | Values |
| --- | --- |
| `svcmode` | `Replicated` `replicated`, `Global` `global`, `ReplicatedJob` `replicated-job`, `GlobalJob` `global-job` |
| `endpoint` | `VIP` `vip`, `DNSRR` `dnsrr` |
| `failact` | `Continue` `continue`, `Pause` `pause`, `Rollback` `rollback` |
| `order` | `StartFirst` `start-first`, `StopFirst` `stop-first` |

### cascade

`onExit` for `composeUp` and `composeStart`. These are numbers, not strings.

| Name | Value |
| --- | --- |
| `Ignore` | `0` |
| `Stop` | `1` |
| `Fail` | `2` |

### progress

`composeUp` option `progress`: `Auto` `auto`, `Plain` `plain`, `TTY` `tty`, `Quiet` `quiet`, `JSON` `json`.

### scheme

`connect({ scheme: scheme.Unix })`. Also `Npipe` `npipe`, `TCP` `tcp`, `HTTP` `http`, `HTTPS` `https`, `SSH` `ssh`, `FD` `fd`.

### Swarm node

`avail`: `Active` `active`, `Pause` `pause`, `Drain` `drain`.

`role`: `Manager` `manager`, `Worker` `worker`.

### df

`diskUsage` type filter: `Container` `container`, `Image` `image`, `Volume` `volume`, `BuildCache` `build-cache`.

### oci

`publishOCI` option `oci`: `V1_0` `1.0`, `V1_1` `1.1`.

### waitFor

`wait` condition: `NotRunning` `not-running`, `NextExit` `next-exit`, `Removed` `removed`.

### watchAction

`WatchFiles` `action`: `Sync` `sync`, `Rebuild` `rebuild`, `SyncRestart` `sync+restart`, `SyncExec` `sync+exec`, `Restart` `restart`.

## Client

```ts
const cli = connect({
  fromEnv: true,
  host: "unix:///var/run/docker.sock",
  scheme: scheme.Unix,
  timeout: "10s",
});
```

Useful `connect` fields: `fromEnv`, `host`, `hostFromEnv`, `scheme`, `timeout`, `apiVersion`, `negotiateAPIVersion`, `tls` (`{ ca, cert, key }` paths), `tlsFromEnv`, `dryRun`, `maxConcurrency`, `userAgent`, `httpHeader`. `timeout` is a number of seconds or a duration string such as `"10s"`.

```ts
const remote = connect({
  host: "tcp://192.0.2.10:2376",
  scheme: scheme.TCP,
  timeout: "15s",
  tls: { ca: "/certs/ca.pem", cert: "/certs/cert.pem", key: "/certs/key.pem" },
});
```

`dryRun: true` asks every command that receives that client to skip changes. Every field, and the rule that `host` cannot be passed to `composeUp`, is in [Client](docs/client.md).

## Commands

A live program calls a method on the session: `await app.composeUp(tree, { detach: true })`. `await` runs the next line after that result. Worked calls are in [Commands](docs/commands.md). There is no plan file.

```ts
await app.composeUp(tree, { detach: true });
```

A test can replace the program `run` starts. Session methods are not redirected by `setInvoke`.

```ts
import { binaryInvoke, setInvoke } from "containerkit";
setInvoke(binaryInvoke("./dist/containerkit"));
```

`exportYAML` is the Compose file. It does not stop the session.

### Project

These take `(cli, tree, options?)` unless noted.

| Command | What it does | Options you usually set |
| --- | --- | --- |
| `composeUp` | Create and start the project | `detach`, `wait`, `timeout`, `services`, `pull`, `removeOrphans`, `onExit`, `progress`, `recreate`, `noCache` |
| `composeDown` | Stop and remove the project | `volumes`, `removeOrphans`, `removeImages`, `timeout`, `services` |
| `composeStart` | Start existing services | `services`, `wait`, `timeout`, `onExit`, `attachTo` |
| `composeStop` | Stop services | `services`, `timeout` |
| `composeRestart` | Restart services | `services`, `timeout`, `noDeps` |
| `composePause`, `composeUnpause` | Pause or unpause | `services` |
| `composeKill` | Kill services | `services`, `stopSignal`, `all` |
| `composePs` | List project containers | `services`, `all` |
| `composePull` | Pull service images | `ignoreBuildable`, `ignoreFailures`, `quiet` |
| `composePush` | Push service images | `ignoreFailures`, `quiet` |
| `composeBuild` | Build services | `dockerfilePath`, `tag`, `target`, `noCache`, `pull`, `platform`, `memory` |
| `composeCreate` | Create containers without starting | project create options |
| `composeRemove` | Remove services | `services`, `stop`, `volumes`, `force` |
| `composeScale` | Scale services | scale options on the tree |
| `composeWait` | Wait until services match a condition | `services`, `downOnExit` |
| `composeWatch` | Rebuild and sync on file changes | watch options |
| `composeEvents` | Stream project events | event filters |
| `composeImages` | List images used by the project | |
| `composeList` | List project resources | `all` |
| `composeVolumeList` | List project volumes | |
| `composePort` | `(cli, tree, service, port, options?)` published address of one port | `protocol`, `index` |
| `composeCopy` | Copy files into the project | copy paths |
| `composeGenerate` | Render the resolved project | |
| `composeViz` | Graph the project | |
| `composeLoadProject` | Load a tree as a project | `file`, `name`, `offline`, `interpolation` |
| `composePublishOCI` | `(cli, tree, repository, options?)` publish as an OCI artifact | `oci`, `application`, `withEnvironment` |
| `composeAttach` | `(cli, tree, service, options?)` attach to a service | |
| `composeCommit` | `(cli, tree, service, options?)` commit a service container | |
| `composeExec` | `(cli, tree, service, options?)` exec in a service | |
| `composeRun` | `(cli, tree, service, options?)` run a one-off service command | |
| `composeExport` | `(cli, tree, service, options?)` export a service filesystem | |

`onExit` takes `cascade.Ignore`, `cascade.Stop`, or `cascade.Fail`. `progress` takes a `progress` value. `timeout` is a number of seconds or a duration string such as `"30s"`.

### Containers

`run(cli, options)` and `containerRun(cli, options)` start a container from options rather than from a tree. The rest take an id.

| Command | Role |
| --- | --- |
| `listContainers` | List containers. `all`, `filter`, `limit`, `since` |
| `create` | Create by name. Options are `{ name }`. One-shot container fields (`image`, `command`, `publish`, `env`) belong on `containerRun` |
| `start` | Start |
| `stop` | Stop. `timeout` |
| `restart` | Restart |
| `kill` | Kill. `stopSignal` |
| `remove` | Remove. `force`, `volumes` |
| `renameContainer` | Rename |
| `updateContainer` | Update resources |
| `logs` | Container logs |
| `statsContainer` | Live stats |
| `topContainer` | Process list |
| `wait` | Wait. Condition is `waitFor` |
| `resizeContainer` | Resize the TTY |
| `copyTo`, `copyFrom` | Copy files in or out |
| `statPath` | Stat a path in the container |
| `attachContainer` | Attach stdio |
| `commitContainer` | Commit to an image |
| `createExec`, `startExec`, `attachExec`, `resizeExec` | Exec sessions |
| `createCheckpoint`, `listCheckpoints`, `removeCheckpoint` | Checkpoints |

### Images

| Command | Arguments | Notes |
| --- | --- | --- |
| `buildImage` | `(cli, options?)` | `image`, `dockerfilePath`, `parseDockerfile`, `target`, `tag`, `noCache`, `pull`, `platform` |
| `buildDockerfile` | `(cli, file, options?)` | Same as `buildImage`, with the Dockerfile tree rendered into `parseDockerfile` |
| `pullImage` | `(cli, ref, options?)` | `platform`, `all` |
| `pushImage` | `(cli, ref, options?)` | |
| `listImages` | `(cli, options?)` | `all`, `filter` |
| `inspectImage` | `(cli, ref, options?)` | |
| `removeImage` | `(cli, ref, options?)` | |
| `tagImage` | `(cli, options?)` | source and target tags |
| `historyImage` | `(cli, ref, options?)` | |
| `searchImages` | `(cli, term, options?)` | |
| `saveImages` | `(cli, refs, options?)` | |
| `attestImage` | `(cli, ref, options?)` | |
| `pruneImages` | `(cli, options?)` | `all`, `filter` |

### Networks

`createNetwork(cli, name, options)`, `listNetworks`, `inspectNetwork(cli, name)`, `connectNetwork(cli, name, { containerID })`, `disconnectNetwork(cli, name, { containerID, force })`, `pruneNetworks`. Create options are `driver` (`netDriver`), `attachable`, `internal`, `label`, and `driverOpt`. Subnet, gateway, and `enableIPv6` are `<Network>` props in the tree.

### Volumes

`createVolume(cli, options)`, `listEngineVolumes`, `removeVolume(cli, id)`, `updateVolume(cli, id)`, `pruneVolumes`. Create options include `driver` and `name`.

### Swarm

| Command | Arguments |
| --- | --- |
| `initSwarm`, `joinSwarm`, `leaveSwarm`, `unlockSwarm`, `updateSwarm` | `(cli, options?)` |
| `createSwarmService` | `(cli, options?)` |
| `listSwarmServices` | `(cli, options?)` |
| `updateSwarmService`, `swarmServiceLogs` | `(cli, id, options?)` |
| `listNodes` | `(cli, options?)` |
| `updateNode`, `removeNode` | `(cli, id, options?)` |
| `listTasks` | `(cli, options?)` |
| `taskLogs` | `(cli, id, options?)` |
| `createEngineConfig`, `listConfigs` | `(cli, options?)` |
| `updateEngineConfig` | `(cli, id, options?)` |
| `createEngineSecret`, `listSecrets` | `(cli, options?)` |
| `updateEngineSecret` | `(cli, id, options?)` |

Node availability is `avail`. Node role is `role`.

### Plugins

`createPlugin`, `installPlugin(cli, name)`, `enablePlugin(cli, name)`, `disablePlugin(cli, name)`, `listPlugins`, `setPlugin(cli, name)`, `upgradePlugin(cli, name)`, `pushPlugin(cli, name)`, `removePlugin(cli, name)`.

### Host

| Command | Role |
| --- | --- |
| `ping` | Check that the daemon answers |
| `login` | Registry login |
| `diskUsage` | Disk usage. `usage` takes `df.Container`, `df.Image`, `df.Volume`, or `df.BuildCache`. `images`, `volumes`, and `buildCache` are booleans |
| `inspect` | Inspect an object by id |
| `listenEvents` | Daemon event stream |
| `load` | Load a Compose file from `path` |
| `pruneBuildCache` | Remove build cache |
| `pruneContainers` | Remove stopped containers |

## Logs

A command result is the value the method returns. A log line arrives on the open session through `app.on("log", ...)`, described in [Live session](docs/live.md). `docker.ts` default-exports `root()`. `useMemo` writes the call result to `.containerkit/memo` and returns the node from that same function. The next process reads the file and skips the call.
