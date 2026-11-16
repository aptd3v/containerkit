# Go kit

This folder is the Go API. Snippets dot-import the kit, so a setter is `Name`, not `kit.Name`.

```go
import . "github.com/aptd3v/containerkit/pkg/kit"
```

A name that is not re-exported (`cap.NET_ADMIN`, `logic.When`) stays a normal import. `ctx` in a snippet is a `context.Context` the caller already has.

Read Define, then Run. The last two pages are the `containerkit` binary, and how a TSX file becomes a document this package loads.

## Define

- [A service](define-a-service.md): `NewContainer`, `NewService`, `NewCompose`, `AsService`, `Validate`.
- [Environment and command](environment-and-command.md): `Command`, `Entrypoint`, `Env`, `Label`, `User`, `WorkingDir`.
- [Publish a port](publish-a-port.md): `Publish` and `Expose`.
- [Mount a volume](mount-a-volume.md): bind, named volume, tmpfs, and a project volume.
- [Networks](networks.md): attachments, project networks, aliases, and `NetworkMode`.
- [Secrets and configs](secrets-and-configs.md): `Secret`, `FileConfig`, `FromFile`.
- [Healthcheck](healthcheck.md): `Test`, `Interval`, `Timeout`, `Retries`, `StartPeriod`.
- [Restart and depends_on](restart-and-depends.md): `Restart`, `DependsOn`, `Condition`.
- [Build and Dockerfile](build-and-dockerfile.md): `Build`, `NewDockerfile`.
- [Resources and deploy](resources-and-deploy.md): `Memory`, `Deploy`, `Replicas`.
- [Capabilities](capabilities.md): `CapAdd`, `Privileged`, `ReadOnly`.
- [Jobs](jobs.md): `NewJob`, `WithJob`.
- [Conditional stacks](conditional-stacks.md): `logic.When`, `logic.Switch`, `AddService`.
- [Same word, different scope](same-word-different-scope.md): names that mean one thing, and names that only look the same.

## Run

- [Bring a compose file up](bring-a-compose-file-up.md): `Connect`, `Load`, `Up`.
- [Run a container](run-a-container.md): `Run`, `Create`, `PullImage`, `BuildImage`.
- [Import and export](import-and-export.md): `ImportYAML`, `ImportFile`, `ExportYAML`, `Write`.
- [Logging](logging.md): container `Logging` and the progress `Logger`.
- [Errors](errors.md): `Validate` and `IsNotFound`.
- [Taskflow](taskflow.md): a DAG of container, build, and Compose steps.

## The binary

- [Runner binary](runner-binary.md): `containerkit version`, `validate`, `exec`, `session`, and `terminal`.
- [A TSX plan, run from Go](tsx-configuration.md): JSX becomes YAML or plan JSON. `ImportYAML` and `ImportPlan` load it.
