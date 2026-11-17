# containerkit docs

These pages are the TypeScript API. You write a Compose project or a Dockerfile as components. A file named `docker.ts` stays open, applies that project, and applies it again when the rendered YAML changes. `exportYAML` prints the file and does not start or stop anything.

These pages assume you know TypeScript and have used Compose: a `compose.yaml`, services, `docker compose up`. They do not assume you have seen containerkit before. Read them in this order.

| Page | Read it when |
| --- | --- |
| [Trees](trees.md) | You are writing the project: services, ports, volumes |
| [Live session](live.md) | You want that project to stay running, and to export YAML whenever you want |
| [Client](client.md) | You need to say which Docker daemon to use |
| [Commands](commands.md) | You want to pull an image, run a container, or do something other than apply the tree |
| [Dockerfiles](dockerfile.md) | A service is built from instructions, not from an image you already have |

Named values such as `cap`, `proto`, and `restartPolicy` come from `@aptd3v/containerkit/vocab`. Components, `exportYAML`, and `connect` stay on `@aptd3v/containerkit`. Hooks such as `useShutdown` come from `@aptd3v/containerkit/hooks`. The session, `root`, and `useState` come from `@aptd3v/containerkit/live`.

Install, the TypeScript setup, and sharing a repository with React are in the package [README](../README.md). [GUIDE.md](../GUIDE.md) is the longer list of prop names and command names.

Starting containers needs Docker, and the `containerkit` program on `PATH` (or `CONTAINERKIT_BIN`). `docker.ts` starts that program and leaves it running. `exportYAML` needs neither Docker nor the program. It only reads a `Compose` node.

The programs you can run, from `tsx/`, are `npm run nginx`, `npm run wordpress`, `npm run mongo-replica`, and `npm run terminal`.
