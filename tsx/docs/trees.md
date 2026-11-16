# Trees

A Compose file is a tree. The project is the root. Services sit inside it. Ports, environment variables, and volume mounts sit inside a service. In containerkit that tree is TypeScript, written with components.

```tsx
import { Compose, Service, Publish, exportYAML } from "containerkit";

const stack = (
  <Compose name="app">
    <Service name="api" image="alpine:latest">
      <Publish hostPort={8080} containerPort={80} />
    </Service>
  </Compose>
);

export const yaml = exportYAML(stack);
```

`yaml` is the Compose file as text. Nothing is started. `exportYAML` accepts a `Compose` node whose `name` contains something other than spaces, and it rejects anything else. How a running program uses the same tree is in [Live session](live.md).

`<Compose name="app">` is a function call. The component is `Compose`, and `name` is the project name. `<Service>` is one service. `image` is the image that service runs. `<Publish>` is the port mapping nested under that service: host port 8080 forwards to port 80 in the container.

A value you can return as the whole project is a `DockerNode`. That is a `Compose`, a `Service`, a `Job`, or a `Container`. `Network`, `Volume`, `Publish`, `Env`, and the other components are children. They are real nodes, and they are not a whole project by themselves.

`Job` is a one-off task in the same project: it runs and finishes, rather than staying up like a service. `Container` describes a single container with the same kind of props as a service. [Live session](live.md) can apply any of those four. `exportYAML` still requires the named `Compose`, so put a service, job, or container inside one when you want the file.

## Props and children

A simple setting is a prop. A nested scope is a child. `Env` and `Label` use `name` because JSX already uses `key`.

```tsx
import { cap, proto } from "containerkit/vocab";

const databaseURL = "postgres://app@db:5432/app";

<Compose name="app">
  <Service name="api" image="alpine:latest" capAdd={cap.NET_ADMIN}>
    <Publish hostPort={8080} containerPort={80} protocol={proto.TCP} />
    <Env name="APP">1</Env>
    <Label name="com.harbor.role">notes</Label>
    <Env name="DATABASE_URL" value={databaseURL} />
  </Service>
</Compose>
```

`Env`, `Label`, and `BuildArg` each store one string. A literal is the child text. A string that already lives in code is the `value` prop. `<Label name="com.harbor.role">notes</Label>` is the same node as `value="notes"`. `<BuildArg name="APP_ENV">{mode}</BuildArg>` is the same node as `<Env name="APP_ENV" value={mode} />`, and inside `<Build>` that node is a build arg. Each component preserves the string it receives from JSX, including `"false"`, `"0"`, and spaces. A multiline literal can have its whitespace normalized by the compiler before the component sees it. An empty string is `value=""`.

Shell and exec forms, heredocs, inline file content, and named options such as `Annotation`, `Sysctl`, and `DriverOpt` stay on their props.

`cap.NET_ADMIN` and `proto.TCP` come from `containerkit/vocab`. They are the legal words for a setting, so a typo is a type error. You can also pass the string the setting expects, such as `"NET_ADMIN"` or `"tcp"`.

`false`, `null`, and `undefined` children are dropped. `{prod && <Deploy replicas={2} />}` omits the deploy block when `prod` is false. An array from `.map()` becomes one child per item.

A function that returns a node is a component. Props are the argument:

```tsx
function API(props: { tag: string }) {
  return <Service name="api" image={`alpine:${props.tag}`} />;
}

<Compose name="app">
  <API tag="3.20" />
</Compose>
```

Calling the component as a function builds the same node. This is useful in a `.ts` file, where there is no JSX:

```ts
import { Compose, Service } from "containerkit";

const stack = Compose({
  name: "app",
  children: Service({ name: "api", image: "alpine:latest" }),
});
```

## The same tag means two things

`<Network>`, `<Volume>`, `<Secret>`, and `<FileConfig>` declare a project resource when they sit under `<Compose>`. The same tags attach that resource to one service when they sit under `<Service>`.

```tsx
import { netDriver } from "containerkit/vocab";

<Compose name="app">
  <Network name="front" driver={netDriver.Bridge} subnet="10.1.0.0/16" gateway="10.1.0.1" enableIPv6={false} />
  <Volume name="data" />
  <Service name="api" image="alpine:latest">
    <Network name="front" />
    <Volume source="data" target="/data" />
  </Service>
</Compose>
```

The first `<Network>` creates a network named `front` for the project, with a subnet. The second attaches the `api` service to that network. The first `<Volume>` creates a named volume. The second mounts it at `/data` in the service.

`createNetwork` is a different thing. It is a command that creates a network on the daemon directly, not a block in this file. Its arguments are in [Commands](commands.md).

## Saying which node a function returns

Most functions do not need a return annotation. Add one when you want TypeScript to reject the wrong component.

```ts
import { Compose, Service, type DockerNode } from "containerkit";

export function stack(): DockerNode<Compose<{ name?: string }>> {
  return (
    <Compose name="app">
      <Service name="api" image="alpine:latest" />
    </Compose>
  );
}
```

`DockerNode<Compose<{ name?: string }>>` means "a Compose node, and the props I am promising are `name`". Returning a `Service` from that function is a type error. `stack().props.image` is a type error because `image` belongs to `Service`, not to this promise.

The same form works for a service:

```ts
function api(): DockerNode<Service<{ name?: string; image?: string }>> {
  return <Service name="api" image="alpine:latest" />;
}
```

On that annotation, `props.image` is `string | undefined`. A prop you left out of the `{ ... }`, such as `command`, is rejected on `props` even though `<Service command={...}>` is legal at the tag. With no type argument, `DockerNode` means any of the four roots.

JSX checks each attribute. `<Compose image="alpine" />` is a type error because `image` is not a Compose prop. `<div />` is a type error in a containerkit file because these components are not HTML. A React file and a containerkit file stay separate. The package README shows the `tsconfig` for that split. Once the containerkit folder extends `containerkit/tsconfig.jsx.json`, you do not put a JSX comment at the top of each file. An optional ESLint config, `containerkit/eslint`, marks an illegal inline tag. The package README shows how a project turns it on. Building the tree still throws when the tag is hidden from the linter.

## Printing the file

```ts
import { render } from "containerkit";

const yaml = render(stack());
```

`render` returns Compose YAML and does not talk to Docker. It also accepts a lone `Service`, `Job`, or `Container` and wraps that node as a one-service file. `exportYAML` is the call to use when you want the file from a project: it requires `Compose` and `name`, and it is what you call from a live program without disturbing that program. Commands such as `composeUp` render the tree for you. You do not call `render` first.
