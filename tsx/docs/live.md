# Live session

A containerkit program stays running. You do not write a file, execute it, and exit. `docker.ts` is the file that starts the session. It default-exports `root()`. The session stays open for the life of that process. Changing state renders the tree again and, when the Compose YAML changed, applies that project again. Containers whose YAML did not change keep running.

```tsx
import { Compose, Service } from "containerkit";
import { root } from "containerkit/live";

export default root({ fromEnv: true }, function App() {
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" />
    </Compose>
  );
});
```

`fromEnv: true` means "use the Docker this machine's `docker` command already uses." The client page explains the other ways to point at a daemon.

[`tsx/examples/docker.tsx`](../examples/docker.tsx) is that file as a program you can run. From `tsx/`, `npm run live` builds the runner and starts it. The process builds a small image, publishes `http://127.0.0.1:18080`, prints that address once the port is open, and stays open. `onExit: "down"` takes the project down on Ctrl+C. `LIVE_PORT` changes the host port.

Run it with a loader that compiles JSX. With the `tsx` package, that is `node --import tsx docker.ts`. `root()` starts the session when the file you ran is named `docker.ts`, `docker.tsx`, or `docker.js`. From another program, call `start()` in `containerkit/live`. `start()` loads `docker.ts` by default and also accepts `docker.tsx`. Any other file name throws. A second `root()` throws while the first session is open.

What the function returns is what gets applied:

- A `Compose`, `Service`, `Job`, or `Container` is applied with `composeUp`, detached, so the session does not follow the containers' logs.
- `null` applies nothing. The function is still waiting, or this program has no project right now.
- The same YAML as last time applies nothing. The containers keep running. The session compares the whole file, not individual fields.

`onExit: "down"` handles Ctrl+C and SIGTERM. `root` waits if an apply is already running, takes down the project it last brought up, and then closes the session. When this file is the program you ran, the process exits. You do not pass the tree to `composeDown`, and you do not register the signals. `close()` still just closes the session, so a caller that stops the root on purpose leaves the project running. Omit `onExit` and the containers stay up when the process exits.

`onUp` runs after an apply. `wait: true` waits until the containers are running, and until a healthcheck passes when the service has one, before `onUp` runs.

## Run once

An `async` function, or a function that takes the session, runs once. It does not render again. `await` finishes before `root` applies the tree you return.

```tsx
import { Compose, Service } from "containerkit";
import { root } from "containerkit/live";

export default root({ fromEnv: true, onExit: "down" }, async function App(session) {
  await session.pullImage("alpine:latest");
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" />
    </Compose>
  );
});
```

A function with no parameters is the render callback above. It runs again when state changes. Hooks (`useState`, `useMemo`, `useEffect`) belong to that callback. The once callback takes the session as an argument instead of calling `useSession()`.

When the once callback throws, or a command it awaits fails, `onError` runs. If this file is the program, the process then exits. The render callback stays open, so a later update can try again.

`exportYAML` is not an apply. It can run on any render, including a render that returns `null`, and it does not stop the session. It is documented below.

## Other files

The entry file calls `root()`. That file is named `docker.ts`, `docker.tsx`, or `docker.js`. Other files export components and functions. A function that talks to Docker takes the session. `await` runs the next command after the previous one returns.

```ts
import { isCKLoginRequiredError, type Session } from "containerkit/live";

export async function pullBase(app: Session, ref: string) {
  try {
    return await app.pullImage(ref);
  } catch (err) {
    if (!isCKLoginRequiredError(err)) throw err;
    await app.login({ serverAddress: "ghcr.io" });
    return app.pullImage(ref);
  }
}
```

`App` imports `pullBase` and calls it. `useSession()`, `useState()`, `useMemo()`, `useEffect()`, and `useResource()` throw if you call them outside the render of that root, including from a random module. A helper hook is fine when `App` calls it during render. `useContainerState` is another name for `useState`.

## Export YAML

`exportYAML` turns a `Compose` node into Compose file text. It does not contact Docker, it does not send a command, and it does not restart the session or the project. Call it on any render.

The node has to be `Compose`, and `name` has to contain something other than spaces. A `Service`, `Job`, or `Container` is not an export root. Put it inside a named `Compose`.

```tsx
import { Compose, Service, exportYAML } from "containerkit";
import { root, useState } from "containerkit/live";

export default root({ fromEnv: true }, function App() {
  const [replicas, setReplicas] = useState(1);
  const tree = (
    <Compose name="app">
      <Service name="api" image="alpine:latest" replicas={replicas} />
    </Compose>
  );
  const yaml = exportYAML(tree);
  return tree;
});
```

`yaml` is the file text for this render. The containers that are already running stay running until the returned tree's YAML actually changes. `setReplicas(2)` renders again, so `exportYAML` runs again, and the project is applied because the file changed. A render whose file matches what is already running exports the text and leaves the containers alone.

You can also call `exportYAML` outside `docker.ts`, on any `Compose` node you already built. That call never opens a session.

## State

`useState` stores a value across renders. Setting it runs `App` again.

```tsx
import { Compose, Service } from "containerkit";
import { root, useState } from "containerkit/live";

export default root({ fromEnv: true }, function App() {
  const [replicas, setReplicas] = useState(1);
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" replicas={replicas} />
    </Compose>
  );
});
```

`setReplicas(2)` changes the file from one replica to two, so the project is applied again. A state update that leaves the YAML the same does not apply anything.

## Calls that must not repeat

`useMemo(fn, deps)` saves what `fn` returns so a later render does not do that work again. Use it for a pull, a login, a `containerRun` that should happen once, a list, or a file read. The dependency list is the array after `fn`. The saved value can be anything, including a node.

```tsx
import { Compose, Service } from "containerkit";
import { root, useMemo, useSession, useState } from "containerkit/live";
import { pullBase } from "./images";

export default root({ fromEnv: true }, function App() {
  const app = useSession();
  const [replicas, setReplicas] = useState(1);
  const image = useMemo(() => pullBase(app, "alpine:latest"), ["alpine:latest"]);
  const seeded = useMemo(
    () => app.containerRun({
      image: "alpine:latest",
      command: ["echo", "ok"],
    }),
    ["alpine:latest"],
  );
  if (!image || !seeded) return null;
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" replicas={replicas} />
    </Compose>
  );
});
```

`setReplicas(2)` renders again, uses both saved results, and applies the project with two replicas. `pullBase` and `containerRun` do not run again. Changing a dependency, such as the image name, finds no saved result for that list, runs that call once, and applies the project with the new tree.

The saved result is a file, `.containerkit/memo/<hash>.json`, in the working directory. The hash is the text of `fn` plus the dependency list. The next process, and the next `root()`, reads the file and skips `fn` when both still match. Closing the session does not delete the files. Add `.containerkit/` to `.gitignore`. It is a local cache, not part of the project.

If `fn` returns a promise, `useMemo` returns `undefined` on that render. Return `null` from `App`. The session waits for the promise and does not apply a project yet. When the promise succeeds, the value is written to disk and `App` renders again with the value filled in. That second tree is the first apply.

This is the end of `App`. `app` is `useSession()`. `image` is the image name you are building. `latestRun` is a function you wrote: it takes the session and returns a promise of the shell command to put in `RUN`.

```tsx
import { Build, Compose, Service } from "containerkit";
import { Dockerfile, From, Run } from "containerkit/dockerfile";
import { useMemo } from "containerkit/live";

const cmd = useMemo(() => latestRun(app, image), [image]);
if (!cmd) return null;
return (
  <Compose name="app">
    <Service name="api" image="app:dev">
      <Build>
        <Dockerfile>
          <From image="alpine" tag="latest" />
          <Run cmd={cmd} />
        </Dockerfile>
      </Build>
    </Service>
  </Compose>
);
```

The first render starts that promise and returns `null`. The Dockerfile is not applied until `cmd` exists, because a Dockerfile on its own is not a project. It has to sit under `<Build>` inside the `Compose` you return.

Strings and numbers in the dependency list match by value. Keys in an object are sorted, so `{ a: "alpine", b: 1 }` and `{ b: 1, a: "alpine" }` are one file. Order in the array still matters: `[image, cmd]` is not `[cmd, image]`. A rejected promise is not written. The next render calls `fn` again.

`root({ cacheDir: "/var/cache/app" })` stores the files somewhere else. `root({ cacheDir: false })` or `CONTAINERKIT_MEMO=0` keeps the saved result in that process only. `useResource` is `useMemo`.

The node you return is not part of that cache. Replica state changes the tree, so the project is applied again, and every saved result whose dependencies are unchanged stays on disk.

## After the apply

`useEffect` runs after a render that has nothing left to wait for. A function it returns runs before that effect runs again, and when the session closes. Inside `App`, after `const app = useSession()`, a log listener looks like this:

```tsx
useEffect(() => app.on("log", (event) => {
  console.log(event.source, event.line);
}), []);
```

`app.on("log" | "status" | "error", handler)` returns a function that removes the handler. Call that function from the effect cleanup if you do not want the listener to last until the session closes. `app.on` already returns that function, so returning it from the effect is enough, as above.

`await app.pullImage(ref)` pulls that image on the open session. A container that runs once and exits does not apply your tree. The node you `return` does. Wrap the calls you want to keep in `useMemo`, and return the node from `App`.

## Errors

A failed command rejects with `ContainerKitError`. Check it with the function for that failure. `message` has secrets removed. `hint` is a short suggestion when there is one. `status` is a numeric code when the daemon sent one. `receiver` names the area that failed, such as a container or a network. `path` is the setting or file involved, when there is one. `exitCode` is the container's exit status.

```ts
import { isCKNotFoundError } from "containerkit/live";

if (isCKNotFoundError(err)) {
  // that image, container, network, or volume is not there
}
```

A missing image, container, network, or volume is `isCKNotFoundError`. The daemon reports one not-found failure for all of them.

Branch inside the function you passed to `useMemo`. Hooks are tracked by the order `App` calls them, so call the same hooks on every render. A hook that runs only on the error path makes the next render throw.

An error thrown from `App` or from an effect is passed to `root({ onError })` and emitted as an `"error"` event. The session stays open so you can retry. `isCKUnavailableError` means Docker cannot be reached. That failure closes the session after `onError`. Files under `.containerkit/memo` stay.

| Check | What happened |
| --- | --- |
| `isCKNotFoundError` | The image, container, network, volume, or other object is not there |
| `isCKAlreadyExistsError` | It is already there |
| `isCKConflictError` | The request clashes with the current state |
| `isCKUnauthorizedError` | The daemon rejected the credentials |
| `isCKPermissionDeniedError` | The credentials are known and are not allowed to do this |
| `isCKLoginRequiredError` | A registry needs a login before the pull or push |
| `isCKLoginFailedError` | The login was attempted and did not succeed |
| `isCKInvalidArgumentError` | A value in the command is not acceptable |
| `isCKParsingFailedError` | The YAML, Dockerfile, or other text could not be parsed |
| `isCKInvalidProjectError` | The project description itself is not valid |
| `isCKFailedPreconditionError` | This call is legal only after something else has happened |
| `isCKOutOfRangeError` | A number or index is outside the range the call accepts |
| `isCKUnavailableError` | Docker cannot be reached. The session closes |
| `isCKUnsupportedError` | This daemon or this call does not support the operation |
| `isCKUnsupportedFlagError` | A flag on the call is not supported here |
| `isCKNotImplementedError` | The operation is not implemented |
| `isCKIncompatibleError` | The daemon and the request do not agree |
| `isCKDisabledError` | The feature is turned off |
| `isCKWrongContextTypeError` | The command was aimed at the wrong kind of object |
| `isCKCanceledError` | The call was canceled |
| `isCKAbortedError` | The call was aborted |
| `isCKNotModifiedError` | Nothing changed |
| `isCKResourceExhaustedError` | The daemon is out of space, memory, or a similar limit |
| `isCKDataLossError` | The result was lost |
| `isCKInternalError` | The failure is inside the daemon or the session |
| `isCKUnknownError` | The failure did not fit another check |
