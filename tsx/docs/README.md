# containerkit docs

containerkit is a TypeScript library for Docker Compose projects and Dockerfiles. You write the project as components. A file named `docker.ts` stays open, applies that project to Docker, and applies it again when your state changes. You can print the Compose YAML at any moment without starting or stopping anything.

These pages assume you know TypeScript and have used Compose as a user: a `compose.yaml`, services, `docker compose up`. They do not assume you have seen containerkit before. Read them in this order.

| Page | Read it when |
| --- | --- |
| [Trees](trees.md) | You are writing the project: services, ports, volumes |
| [Live session](live.md) | You want that project to stay running, and to export YAML whenever you want |
| [Client](client.md) | You need to say which Docker daemon to use |
| [Commands](commands.md) | You want to pull an image, list containers, or do something other than apply the tree |
| [Dockerfiles](dockerfile.md) | A service is built from instructions, not from an image you already have |

Named values such as `cap`, `proto`, and `restartPolicy` come from `containerkit/vocab`. Components, `exportYAML`, and `connect` stay on `containerkit`.

Install, the TypeScript setup, and sharing a repository with React are in the package [README](../README.md). [GUIDE.md](../GUIDE.md) is a longer list of prop names and command names once the pages above are familiar.

Starting containers needs Docker running, and the `containerkit` program on your `PATH` (or `CONTAINERKIT_BIN` pointing at it). `docker.ts` starts that program and leaves it running. You do not write a plan file, and you do not stop and start that program to export YAML or to change the project. `exportYAML` needs neither Docker nor the program. It only reads a `Compose` node.
