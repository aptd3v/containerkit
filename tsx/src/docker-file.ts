import { resolveScalar } from "./env.js";
import { element, isReactElement, reactCompiledJSX, type Child, type Node } from "./node.js";
import type { SignalName } from "./vocab.js";

// Dockerfile TSX. Instruction names match the Dockerfile, so this module is a
// separate entry from the compose components (Env, Label, Expose, Volume).

/**
 * Props for {@link Dockerfile}.
 *
 * `syntax`, `escape`, and `check` are parser directives (`# syntax=`, `# escape=`, `# check=`).
 * They are written before every instruction. The same lines can be child elements
 * ({@link Syntax}, {@link Escape}, {@link Check}) when those elements stay ahead of the first instruction.
 */
export type DockerfileProps = {
	/**
	 * Frontend to pin, written as `# syntax=`.
	 * Leave this unset to use the Dockerfile frontend bundled with the builder.
	 * Example: `docker/dockerfile:1`.
	 */
	syntax?: string;
	/** Escape character, backslash or backtick, written as `# escape=`. */
	escape?: string;
	/** Build-check configuration, written as `# check=`. */
	check?: string;
	/** Instruction elements. A `<Build>` accepts one Dockerfile. */
	children?: Child;
};

/** A Dockerfile tree. Render it with {@link exportDockerfile}, nest it under `<Build>`, or pass it to `buildDockerfile`. */
export type DockerfileNode = Node<"dockerfile", Omit<DockerfileProps, "children">>;

/**
 * Root of a Dockerfile.
 *
 * Nest one `<Dockerfile>` under `<Build>` inside a Compose service so the build stores the
 * text inline. {@link exportDockerfile} returns that text and does not build the image.
 * `exportYAML` accepts the Dockerfile only as part of a Compose project.
 *
 * @example
 * ```tsx
 * <Dockerfile>
 *   <From image="alpine" tag="latest" />
 *   <Run cmd="echo ok" />
 * </Dockerfile>
 * ```
 */
export function Dockerfile(props: DockerfileProps): DockerfileNode {
	return element("dockerfile", props);
}

/** Source and destination for {@link Copy} and {@link Add}. A path that contains a space is written inside a JSON array. */
type SrcDest = {
	/** Source path in the build context. For {@link Copy}, an absolute path when `from` names another stage. */
	src: string;
	/** Destination path in the image. */
	dest: string;
};

/** Props for {@link From}. */
export type FromProps = {
	/**
	 * Repository (`node`), a reference that already includes a tag (`golang:1.22`),
	 * or the name of an earlier stage. When `tag` is set, `image` is the repository name and must not contain `:`.
	 */
	image: string;
	/** Tag appended as `:tag`. Omit when `image` already includes a tag or names a stage. */
	tag?: string;
	/** Target platform, written as `--platform`. Example: `linux/arm64`. */
	platform?: string;
	/** Stage name, written as `AS`. Later {@link Copy} `from` and {@link From} `image` values refer to it. */
	as?: string;
	children?: never;
};

/**
 * Props for {@link Arg}.
 *
 * Omit `default` and the child to declare the name with no value (`ARG NAME`).
 * An empty-string default writes `ARG NAME=`. A string child is the default;
 * a value computed in code belongs on `default`.
 */
export type ArgProps =
	| {
			/** Build-arg name. */
			name: string;
			/** Default value. Omit the property to declare the arg with no default. `""` writes `ARG NAME=`. */
			default?: string;
			children?: never;
	  }
	| {
			/** Build-arg name. */
			name: string;
			default?: never;
			/** Default value. An empty string writes `ARG NAME=`. */
			children: string;
	  };

/**
 * Props for {@link Env}. Written as `ENV NAME=value`.
 * A string child is the value. A value computed in code belongs on `value`.
 */
export type EnvProps =
	| {
			/** Variable name. */
			name: string;
			/** Variable value. */
			value: string;
			children?: never;
	  }
	| {
			/** Variable name. */
			name: string;
			value?: never;
			/** Variable value. */
			children: string;
	  };

/**
 * Props for {@link Copy}. Copies from the build context, or from another stage when `from` is set.
 */
export type CopyProps = SrcDest & {
	/** Earlier stage name (the `as` of a {@link From}), written as `--from`. */
	from?: string;
	/** `--chown` value, such as `node:node`. */
	chown?: string;
	/** `--chmod` value, such as `755`. */
	chmod?: string;
	/** Write `--link`, so the copy can reuse a cached layer. */
	link?: boolean;
	/** One or more `--exclude` patterns. An empty pattern throws. */
	exclude?: string | readonly string[];
	/** Write `--parents`, preserving the source directory structure under `dest`. */
	parents?: boolean;
	children?: never;
};

/**
 * Props for {@link Add}. Copies local files, fetches a URL, and can unpack an archive.
 */
export type AddProps = SrcDest & {
	/** `--chown` value, such as `node:node`. */
	chown?: string;
	/** `--chmod` value, such as `755`. */
	chmod?: string;
	/** `--checksum` for a remote source. */
	checksum?: string;
	/** Write `--keep-git-dir=true` when `src` is a Git URL. */
	keepGitDir?: boolean;
	/** Write `--link`, so the add can reuse a cached layer. */
	link?: boolean;
	/** One or more `--exclude` patterns. An empty pattern throws. */
	exclude?: string | readonly string[];
	children?: never;
};

/** Program and arguments shared by the exec-form and shell-form instructions. */
export type ExecProps = {
	/** Program to run. */
	executable: string;
	/** Arguments after `executable`. */
	args?: readonly string[];
	children?: never;
};

/** Props for {@link Expose}. */
export type ExposeProps = {
	/** Port, or `port/protocol` such as `8080/tcp`. */
	port: string;
	children?: never;
};

/**
 * Props for {@link Label}. Written as `LABEL name=value`.
 * A string child is the value. A value computed in code belongs on `value`.
 */
export type LabelProps =
	| {
			/** Label key. */
			name: string;
			/** Label value. */
			value: string;
			children?: never;
	  }
	| {
			/** Label key. */
			name: string;
			value?: never;
			/** Label value. */
			children: string;
	  };

/** Props for {@link Onbuild}. */
export type OnbuildProps = {
	/** Dockerfile instruction deferred until an image uses this one as its base. Example: `RUN echo hi`. */
	command: string;
	children?: never;
};

/**
 * Props for {@link Workdir}.
 * A string child is the path. A path computed in code belongs on `path`.
 */
export type WorkdirProps =
	| {
			/** Working directory inside the image. */
			path: string;
			children?: never;
	  }
	| {
			path?: never;
			/** Working directory inside the image. */
			children: string;
	  };

/** Props for {@link StopSignal}. */
export type StopSignalProps = {
	/** POSIX signal name sent to the container's main process, such as `SIGTERM`. */
	signal: SignalName;
	children?: never;
};

/**
 * Props for {@link User}.
 * A string child is the user. A value computed in code belongs on `name`.
 */
export type UserProps =
	| {
			/** User, uid, or `uid:gid`, such as `node` or `1000:1000`. */
			name: string;
			children?: never;
	  }
	| {
			name?: never;
			/** User, uid, or `uid:gid`, such as `node` or `1000:1000`. */
			children: string;
	  };

/** Props for {@link Comment}. The text is one line; a line break throws. */
export type CommentProps = {
	/** Comment text, without the leading `#`. */
	text: string;
	children?: never;
};

/** Props for {@link Volume}. Paths are written as a JSON array. */
export type VolumeProps = {
	/** Mount points created in the image. */
	paths: readonly string[];
	children?: never;
};

/**
 * Props for {@link HealthCheck}.
 *
 * `test` starts with `CMD`, `CMD-SHELL`, or `NONE`.
 * `CMD` is exec form: each following string is one argument, written as `CMD ["…"]`.
 * `CMD-SHELL` is two strings, the prefix and the shell command, written as `CMD <command>`.
 * `NONE` disables the check, the same line as {@link HealthCheckNone}.
 */
export type HealthCheckProps = {
	/** Check command. The first element is `CMD`, `CMD-SHELL`, or `NONE`. */
	test: readonly string[];
	/** `--interval` duration, such as `30s`. */
	interval?: string;
	/** `--timeout` duration, such as `3s`. */
	timeout?: string;
	/** `--start-period` duration, such as `15s`. */
	startPeriod?: string;
	/** `--retries`. Omitted when missing or not greater than zero. */
	retries?: number;
	children?: never;
};

/**
 * Props for {@link Run} and {@link RunLayer}. The command is shell form.
 *
 * A {@link Run} with no `mount`, `network`, or `security` is appended with `&&` when the
 * previous instruction was also a run, so those calls share one layer.
 * A run that sets any of those flags starts its own `RUN`. {@link RunLayer} always starts a new `RUN`.
 */
export type RunProps = {
	/** Shell command. */
	cmd: string;
	/** One or more `--mount` values, such as `type=cache,id=pnpm,target=/pnpm/store`. An empty mount throws. */
	mount?: string | readonly string[];
	/** `--network` value, such as `host` or `none`. */
	network?: string;
	/** `--security` value, such as `insecure`. */
	security?: string;
	children?: never;
};

/**
 * Props for {@link RunArgs}.
 * The previous instruction must be a {@link Run} or {@link RunLayer}. Each argument is a continuation line on that `RUN`.
 */
export type RunArgsProps = {
	/** Arguments appended to the previous `RUN`. At least one is required. */
	args: readonly string[];
	children?: never;
};

/** Value for a parser directive. The directive is written before every instruction. */
export type DirectiveProps = {
	/** Directive value, without the `# name=` prefix. An empty value throws. */
	value: string;
	children?: never;
};

/**
 * `FROM` instruction. Starts a build stage.
 *
 * @example
 * ```tsx
 * <From image="node" tag="22" as="build" platform="linux/arm64" />
 * <From image="build" as="runtime" />
 * ```
 */
export const From = instr<"dfFrom", FromProps>("dfFrom");

/**
 * `ARG` instruction. Declares a build arg, with an optional default.
 *
 * `<Arg name="APP_ENV" />` writes `ARG APP_ENV`. `<Arg name="APP_ENV" default="" />` writes `ARG APP_ENV=`.
 *
 * @example
 * ```tsx
 * <Arg name="APP_VERSION" />
 * <Arg name="APP_ENV">production</Arg>
 * ```
 */
export function Arg(props: ArgProps): Node<"dfArg", { name: string; default?: string }> {
	return element("dfArg", resolveScalar(props, "default", { tag: "arg", required: false }) as { name: string; default?: string });
}

/**
 * `ENV` instruction, written as `ENV NAME=value`.
 *
 * @example
 * ```tsx
 * <Env name="NODE_ENV">production</Env>
 * <Env name="APP_NAME" value={name} />
 * ```
 */
export function Env(props: EnvProps): Node<"dfEnv", { name: string; value: string }> {
	return element("dfEnv", resolveScalar(props, "value", { tag: "env", required: true }) as { name: string; value: string });
}

/**
 * `COPY` instruction. Use `from` to copy out of an earlier stage.
 *
 * @example
 * ```tsx
 * <Copy from="build" src="/src/dist" dest="/app/dist" chown="node:node" link />
 * ```
 */
export const Copy = instr<"dfCopy", CopyProps>("dfCopy");

/**
 * `ADD` instruction. Copies files, fetches a URL, or unpacks a local archive.
 *
 * @example
 * ```tsx
 * <Add src="https://example.com/app.tgz" dest="/tmp/app.tgz" checksum="sha256:…" />
 * ```
 */
export const Add = instr<"dfAdd", AddProps>("dfAdd");

/**
 * `ENTRYPOINT` in exec form: `ENTRYPOINT ["executable", "arg"]`.
 * The process is started directly. {@link EntrypointShell} writes a shell command line.
 *
 * @example
 * ```tsx
 * <Entrypoint executable="/usr/bin/tini" args={["--"]} />
 * ```
 */
export const Entrypoint = instr<"dfEntrypoint", ExecProps>("dfEntrypoint");

/**
 * `ENTRYPOINT` in shell form: `ENTRYPOINT executable arg`.
 * The shell interprets the line. {@link Entrypoint} writes a JSON array.
 */
export const EntrypointShell = instr<"dfEntrypointShell", ExecProps>("dfEntrypointShell");

/**
 * `EXPOSE` instruction. Documents the ports the image listens on.
 *
 * @example
 * ```tsx
 * <Expose port="8080" />
 * ```
 */
export const Expose = instr<"dfExpose", ExposeProps>("dfExpose");

/**
 * `LABEL` instruction, written as `LABEL name=value`.
 *
 * @example
 * ```tsx
 * <Label name="org.opencontainers.image.title">ledger</Label>
 * ```
 */
export function Label(props: LabelProps): Node<"dfLabel", { name: string; value: string }> {
	return element("dfLabel", resolveScalar(props, "value", { tag: "label", required: true }) as { name: string; value: string });
}

/**
 * `ONBUILD` instruction. `command` runs when a later image uses this one in `FROM`.
 *
 * @example
 * ```tsx
 * <Onbuild command="RUN echo hi" />
 * ```
 */
export const Onbuild = instr<"dfOnbuild", OnbuildProps>("dfOnbuild");

/**
 * `WORKDIR` instruction. Later instructions run in this directory.
 *
 * @example
 * ```tsx
 * <Workdir>/app</Workdir>
 * ```
 */
export function Workdir(props: WorkdirProps): Node<"dfWorkdir", { path: string }> {
	return element("dfWorkdir", resolveScalar(props, "path", { tag: "workdir", required: true }) as { path: string });
}

/**
 * `STOPSIGNAL` instruction. Sets the signal used to stop a container started from this image.
 *
 * @example
 * ```tsx
 * <StopSignal signal="SIGTERM" />
 * ```
 */
export const StopSignal = instr<"dfStopSignal", StopSignalProps>("dfStopSignal");

/**
 * `USER` instruction. Later instructions and the default process run as this user.
 *
 * @example
 * ```tsx
 * <User>node</User>
 * ```
 */
export function User(props: UserProps): Node<"dfUser", { name: string }> {
	return element("dfUser", resolveScalar(props, "name", { tag: "user", required: true }) as { name: string });
}

/**
 * A `#` comment in the Dockerfile. The text is a single line.
 *
 * @example
 * ```tsx
 * <Comment text="runtime stage" />
 * ```
 */
export const Comment = instr<"dfComment", CommentProps>("dfComment");

/**
 * `VOLUME` instruction. Creates the mount points listed in `paths`.
 *
 * @example
 * ```tsx
 * <Volume paths={["/var/lib/ledger"]} />
 * ```
 */
export const Volume = instr<"dfVolume", VolumeProps>("dfVolume");

/**
 * `HEALTHCHECK` instruction. `test` begins with `CMD`, `CMD-SHELL`, or `NONE`.
 *
 * @example
 * ```tsx
 * <HealthCheck test={["CMD", "true"]} interval="30s" timeout="3s" retries={3} />
 * ```
 */
export const HealthCheck = instr<"dfHealth", HealthCheckProps>("dfHealth");

/**
 * `HEALTHCHECK NONE`. Disables a healthcheck inherited from the base image.
 *
 * @example
 * ```tsx
 * <HealthCheckNone />
 * ```
 */
export const HealthCheckNone = instr<"dfHealthNone", Record<string, never>>("dfHealthNone");

/**
 * `SHELL` in exec form. Selects the shell that later shell-form instructions use.
 *
 * @example
 * ```tsx
 * <Shell executable="/bin/bash" args={["-c"]} />
 * ```
 */
export const Shell = instr<"dfShell", ExecProps>("dfShell");

/**
 * `CMD` in exec form: `CMD ["executable", "arg"]`.
 * The process is started directly. A Dockerfile has one `CMD`; a second {@link Command} or {@link CommandShell} throws.
 *
 * @example
 * ```tsx
 * <Command executable="node" args={["/app/dist/server.js"]} />
 * ```
 */
export const Command = instr<"dfCommand", ExecProps>("dfCommand");

/**
 * `CMD` in shell form: `CMD executable arg`.
 * The shell interprets the line. A Dockerfile has one `CMD`; a second {@link Command} or {@link CommandShell} throws.
 *
 * @example
 * ```tsx
 * <CommandShell executable="/out/app" args={["--port", "8080"]} />
 * ```
 */
export const CommandShell = instr<"dfCommandShell", ExecProps>("dfCommandShell");

/**
 * `RUN` in shell form.
 *
 * Consecutive `<Run>` elements with no flags become one layer, joined with `&&`.
 * A `mount`, `network`, or `security` flag starts a new `RUN`.
 * Use {@link RunLayer} to start a new layer on purpose, and {@link RunArgs} to append arguments.
 *
 * @example
 * ```tsx
 * <Run cmd="pnpm install --frozen-lockfile" mount="type=cache,id=pnpm,target=/pnpm/store" />
 * ```
 */
export const Run = instr<"dfRun", RunProps>("dfRun");

/**
 * `RUN` that starts a new layer even when the previous instruction was a run.
 * A `<Run>` after a `<From>` is already a new layer; use `<RunLayer>` when the previous line was itself a run.
 *
 * @example
 * ```tsx
 * <Run cmd="npm ci" />
 * <RunLayer cmd="npm run build" />
 * ```
 */
export const RunLayer = instr<"dfRunLayer", RunProps>("dfRunLayer");

/**
 * Appends arguments to the previous {@link Run} or {@link RunLayer}, each on a continuation line.
 * Throws when no run precedes it, or when `args` is empty.
 *
 * @example
 * ```tsx
 * <Run cmd="apk add" />
 * <RunArgs args={["--no-cache", "curl"]} />
 * ```
 */
export const RunArgs = instr<"dfRunArgs", RunArgsProps>("dfRunArgs");

/**
 * `# syntax=` parser directive. Place it before any instruction, or set {@link DockerfileProps.syntax}.
 * Leave it unset to use the frontend bundled with the builder. Set it to pin an external frontend.
 *
 * @example
 * ```tsx
 * <Syntax value="docker/dockerfile:1" />
 * ```
 */
export const Syntax = instr<"dfSyntax", DirectiveProps>("dfSyntax");

/**
 * `# escape=` parser directive. `value` is `\` or a backtick.
 * Place it before any instruction, or set {@link DockerfileProps.escape}.
 *
 * @example
 * ```tsx
 * <Escape value="`" />
 * ```
 */
export const Escape = instr<"dfEscape", DirectiveProps>("dfEscape");

/**
 * `# check=` parser directive. Place it before any instruction, or set {@link DockerfileProps.check}.
 *
 * @example
 * ```tsx
 * <Check value="skip=JSONArgsRecommended" />
 * ```
 */
export const Check = instr<"dfCheck", DirectiveProps>("dfCheck");

function instr<K extends string, P extends Record<string, unknown>>(kind: K) {
	return function instruction(props: P): Node<K, Omit<P, "children">> {
		return element(kind, props);
	};
}

/**
 * Renders a {@link Dockerfile} tree to Dockerfile text. Does not build the image.
 *
 * Throws when `node` is not a Dockerfile, when a parser directive follows an instruction,
 * when an instruction value is empty or contains a line break, or when `CMD` is set more than once.
 *
 * @param node - A node produced by {@link Dockerfile}.
 * @returns The Dockerfile, including the trailing newline on each instruction.
 */
export function exportDockerfile(node: Node): string {
	if (isReactElement(node)) {
		throw new Error(reactCompiledJSX);
	}
	if (node.$$kind !== "dockerfile") {
		throw new Error(`containerkit: exportDockerfile requires a Dockerfile node, got <${node.$$kind}>`);
	}
	let text = "";
	let body = false;
	let lastRun = false;
	let cmdSet = false;

	const directive = (name: string, value: string) => {
		if (body) {
			fail(`${name} directive must precede instructions`);
		}
		token(value, `${name} directive`);
		if (value.trim() === "") {
			fail(`${name} directive is empty`);
		}
		text += `# ${name}=${value}\n`;
	};

	const prop = (key: string) => (typeof node.props[key] === "string" ? (node.props[key] as string) : "");
	if (prop("syntax")) {
		directive("syntax", prop("syntax"));
	}
	if (prop("escape")) {
		escapeChar(prop("escape"));
		directive("escape", prop("escape"));
	}
	if (prop("check")) {
		directive("check", prop("check"));
	}

	const commit = (line: string, run: boolean) => {
		body = true;
		lastRun = run;
		text += line;
	};

	for (const child of node.children) {
		const p = child.props;
		switch (child.$$kind) {
			case "dfSyntax":
				directive("syntax", str(p.value, "syntax directive"));
				break;
			case "dfEscape": {
				const value = str(p.value, "escape directive");
				escapeChar(value);
				directive("escape", value);
				break;
			}
			case "dfCheck":
				directive("check", str(p.value, "check directive"));
				break;
			case "dfFrom":
				commit(fromLine(p), false);
				break;
			case "dfArg":
				commit(argLine(p), false);
				break;
			case "dfEnv":
				commit(`ENV ${token(str(p.name, "env name"), "env name")}=${token(str(p.value, "env value"), "env value")}\n`, false);
				break;
			case "dfCopy":
				commit(fileLine("COPY", copyFlags(p), p), false);
				break;
			case "dfAdd":
				commit(fileLine("ADD", addFlags(p), p), false);
				break;
			case "dfEntrypoint":
				commit(execLine("ENTRYPOINT", p), false);
				break;
			case "dfEntrypointShell":
				commit(shellLine("ENTRYPOINT", p), false);
				break;
			case "dfExpose":
				commit(`EXPOSE ${token(str(p.port, "expose port"), "expose port")}\n`, false);
				break;
			case "dfLabel":
				commit(`LABEL ${token(str(p.name, "label name"), "label name")}=${token(str(p.value, "label value"), "label value")}\n`, false);
				break;
			case "dfOnbuild":
				commit(`ONBUILD ${token(str(p.command, "onbuild command"), "onbuild")}\n`, false);
				break;
			case "dfWorkdir":
				commit(`WORKDIR ${token(str(p.path, "workdir path"), "workdir")}\n`, false);
				break;
			case "dfStopSignal":
				commit(`STOPSIGNAL ${token(str(p.signal, "stopsignal"), "stopsignal")}\n`, false);
				break;
			case "dfUser":
				commit(`USER ${token(str(p.name, "user"), "user")}\n`, false);
				break;
			case "dfComment":
				commit(`# ${token(str(p.text, "comment"), "comment")}\n`, false);
				break;
			case "dfVolume":
				commit(`VOLUME ${jsonArray(list(p.paths, "volume path"))}\n`, false);
				break;
			case "dfHealth":
				commit(healthLine(p), false);
				break;
			case "dfHealthNone":
				commit("HEALTHCHECK NONE\n", false);
				break;
			case "dfShell":
				commit(execLine("SHELL", p), false);
				break;
			case "dfCommand":
				if (cmdSet) {
					fail("command has already been set");
				}
				cmdSet = true;
				commit(execLine("CMD", p), false);
				break;
			case "dfCommandShell":
				if (cmdSet) {
					fail("command has already been set");
				}
				cmdSet = true;
				commit(shellLine("CMD", p), false);
				break;
			case "dfRun":
				text = runLine(text, p, false, lastRun);
				body = true;
				lastRun = true;
				break;
			case "dfRunLayer":
				text = runLine(text, p, true, lastRun);
				body = true;
				lastRun = true;
				break;
			case "dfRunArgs":
				text = runArgsLine(text, p, lastRun);
				body = true;
				lastRun = true;
				break;
			default:
				fail(`<${child.$$kind}> is not a Dockerfile instruction`);
		}
	}
	return text;
}

function runLine(text: string, p: Record<string, unknown>, fresh: boolean, lastRun: boolean): string {
	const cmd = token(str(p.cmd, "run command"), "run");
	const flags = runFlags(p);
	if (lastRun && !fresh && flags === "") {
		return `${text.replace(/\n$/, "")} && \\\n\t${cmd}\n`;
	}
	return text + `RUN ${flags}${cmd}\n`;
}

function runArgsLine(text: string, p: Record<string, unknown>, lastRun: boolean): string {
	if (!lastRun) {
		fail("runargs was called but no run command was called before it");
	}
	const args = list(p.args, "runargs argument");
	if (args.length === 0) {
		fail("runargs requires at least one argument");
	}
	for (const arg of args) {
		token(arg, "runargs");
	}
	return `${text.replace(/\n$/, "")} \\\n\t${args.join(" \\\n\t")}\n`;
}

function fromLine(p: Record<string, unknown>): string {
	const image = str(p.image, "from image").trim();
	const tag = opt(p.tag).trim();
	const platform = opt(p.platform).trim();
	const alias = opt(p.as).trim();
	if (image === "") {
		fail("from image is empty");
	}
	if (tag !== "" && image.includes(":")) {
		fail("from tag must be passed separately from the image");
	}
	for (const part of [platform, image, tag, alias]) {
		if (/[ \t\r\n]/.test(part)) {
			fail("from contains whitespace");
		}
	}
	let line = "FROM ";
	if (platform !== "") {
		line += `--platform=${platform} `;
	}
	line += image;
	if (tag !== "") {
		line += `:${tag}`;
	}
	if (alias !== "") {
		line += ` AS ${alias}`;
	}
	return line + "\n";
}

function argLine(p: Record<string, unknown>): string {
	const name = token(str(p.name, "arg name"), "arg name");
	if (!("default" in p) || p.default === undefined) {
		return `ARG ${name}\n`;
	}
	if (p.default === "") {
		return `ARG ${name}=\n`;
	}
	return `ARG ${name}=${token(String(p.default), "arg default")}\n`;
}

function copyFlags(p: Record<string, unknown>): string {
	let flags = "";
	flags += flag("--from", opt(p.from));
	flags += flag("--chown", opt(p.chown));
	flags += flag("--chmod", opt(p.chmod));
	for (const pattern of many(p.exclude)) {
		if (pattern === "") {
			fail("--exclude is empty");
		}
		flags += flag("--exclude", pattern);
	}
	if (p.parents === true) {
		flags += "--parents ";
	}
	if (p.link === true) {
		flags += "--link ";
	}
	return flags;
}

function addFlags(p: Record<string, unknown>): string {
	let flags = "";
	flags += flag("--chown", opt(p.chown));
	flags += flag("--chmod", opt(p.chmod));
	flags += flag("--checksum", opt(p.checksum));
	if (p.keepGitDir === true) {
		flags += "--keep-git-dir=true ";
	}
	for (const pattern of many(p.exclude)) {
		if (pattern === "") {
			fail("--exclude is empty");
		}
		flags += flag("--exclude", pattern);
	}
	if (p.link === true) {
		flags += "--link ";
	}
	return flags;
}

function runFlags(p: Record<string, unknown>): string {
	let flags = "";
	for (const mount of many(p.mount)) {
		if (mount === "") {
			fail("--mount is empty");
		}
		flags += flag("--mount", mount);
	}
	flags += flag("--network", opt(p.network));
	flags += flag("--security", opt(p.security));
	return flags;
}

function fileLine(instrName: string, flags: string, p: Record<string, unknown>): string {
	const src = str(p.src, `${instrName.toLowerCase()} source`).trim();
	const dest = str(p.dest, `${instrName.toLowerCase()} destination`).trim();
	token(src, `${instrName.toLowerCase()} source`);
	token(dest, `${instrName.toLowerCase()} destination`);
	if (src.includes(" ") || dest.includes(" ")) {
		return `${instrName} ${flags}${jsonArray([src, dest])}\n`;
	}
	return `${instrName} ${flags}${src} ${dest}\n`;
}

function execLine(instrName: string, p: Record<string, unknown>): string {
	const executable = str(p.executable, `${instrName.toLowerCase()} executable`);
	const args = many(p.args);
	return `${instrName} ${jsonArray([executable, ...args])}\n`;
}

function shellLine(instrName: string, p: Record<string, unknown>): string {
	const executable = token(str(p.executable, `${instrName.toLowerCase()} executable`).trim(), `${instrName.toLowerCase()} executable`);
	const args = many(p.args).map((arg, i) => token(arg, `${instrName.toLowerCase()} arg[${i}]`));
	return `${instrName} ${[executable, ...args].join(" ")}\n`;
}

function healthLine(p: Record<string, unknown>): string {
	const test = list(p.test, "healthcheck test");
	const cmd = healthCmd(test);
	if (cmd === "NONE") {
		return "HEALTHCHECK NONE\n";
	}
	const flags: string[] = [];
	if (opt(p.interval)) {
		flags.push(`--interval=${opt(p.interval)}`);
	}
	if (opt(p.timeout)) {
		flags.push(`--timeout=${opt(p.timeout)}`);
	}
	if (opt(p.startPeriod)) {
		flags.push(`--start-period=${opt(p.startPeriod)}`);
	}
	if (typeof p.retries === "number" && p.retries > 0) {
		flags.push(`--retries=${p.retries}`);
	}
	if (flags.length === 0) {
		return `HEALTHCHECK ${cmd}\n`;
	}
	return `HEALTHCHECK ${flags.join(" ")} \\\n\t${cmd}\n`;
}

function healthCmd(test: string[]): string {
	if (test.length === 0) {
		fail("healthcheck test is empty");
	}
	switch (test[0]) {
		case "NONE":
			return "NONE";
		case "CMD-SHELL":
			if (test.length !== 2 || test[1].trim() === "") {
				fail("healthcheck shell command is empty");
			}
			token(test[1], "healthcheck shell command");
			return `CMD ${test[1]}`;
		case "CMD":
			if (test.length < 2 || test[1].trim() === "") {
				fail("healthcheck command is empty");
			}
			return `CMD ${jsonArray(test.slice(1))}`;
		default:
			fail("healthcheck test must start with CMD, CMD-SHELL, or NONE");
	}
}

function jsonArray(args: readonly string[]): string {
	if (args.length === 0) {
		fail("json array is empty");
	}
	return `[${args.map((arg) => JSON.stringify(arg)).join(", ")}]`;
}

function flag(name: string, value: string): string {
	if (value === "") {
		return "";
	}
	if (/[ \t\r\n]/.test(value)) {
		fail(`${name} contains whitespace`);
	}
	return `${name}=${value} `;
}

function escapeChar(value: string): void {
	if (value !== "\\" && value !== "`") {
		fail("escape directive must be \\ or `");
	}
}

function token(value: string, what: string): string {
	if (value.trim() === "") {
		fail(`${what} is empty`);
	}
	for (let i = 0; i < value.length; i++) {
		const c = value.charCodeAt(i);
		if (c === 10 || c === 13) {
			fail(`${what} contains line break`);
		}
		if (c === 0 || c === 0x7f || (c < 0x20 && c !== 9)) {
			fail(`${what} contains control character`);
		}
	}
	return value;
}

function str(v: unknown, what: string): string {
	if (typeof v !== "string") {
		fail(`${what} is empty`);
	}
	return v;
}

function opt(v: unknown): string {
	return typeof v === "string" ? v : "";
}

function many(v: unknown): string[] {
	if (v == null) {
		return [];
	}
	if (typeof v === "string") {
		return [v];
	}
	if (Array.isArray(v)) {
		return v.map((item) => String(item));
	}
	return [];
}

function list(v: unknown, what: string): string[] {
	const items = many(v);
	for (const item of items) {
		if (item.trim() === "") {
			fail(`${what} is empty`);
		}
	}
	return items;
}

function fail(msg: string): never {
	throw new Error(`containerkit/tsx: ${msg}`);
}
