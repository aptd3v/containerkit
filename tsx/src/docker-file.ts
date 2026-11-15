import { resolveScalar } from "./env.js";
import { element, isReactElement, reactCompiledJSX, type Child, type Node } from "./node.js";

// Dockerfile TSX. Instruction names match the Dockerfile, so this module is a
// separate entry from the compose components (Env, Label, Expose, Volume).

export type DockerfileProps = {
	syntax?: string;
	escape?: string;
	check?: string;
	children?: Child;
};

export type DockerfileNode = Node<"dockerfile", Omit<DockerfileProps, "children">>;

export function Dockerfile(props: DockerfileProps): DockerfileNode {
	return element("dockerfile", props);
}

type SrcDest = { src: string; dest: string };

export type FromProps = { image: string; tag?: string; platform?: string; as?: string; children?: never };
export type ArgProps =
	| { name: string; default?: string; children?: never }
	| { name: string; default?: never; children: string };
export type EnvProps =
	| { name: string; value: string; children?: never }
	| { name: string; value?: never; children: string };
export type CopyProps = SrcDest & {
	from?: string;
	chown?: string;
	chmod?: string;
	link?: boolean;
	exclude?: string | readonly string[];
	parents?: boolean;
	children?: never;
};
export type AddProps = SrcDest & {
	chown?: string;
	chmod?: string;
	checksum?: string;
	keepGitDir?: boolean;
	link?: boolean;
	exclude?: string | readonly string[];
	children?: never;
};
export type ExecProps = { executable: string; args?: readonly string[]; children?: never };
export type ExposeProps = { port: string; children?: never };
export type LabelProps =
	| { name: string; value: string; children?: never }
	| { name: string; value?: never; children: string };
export type OnbuildProps = { command: string; children?: never };
export type WorkdirProps =
	| { path: string; children?: never }
	| { path?: never; children: string };
export type StopSignalProps = { signal: string; children?: never };
export type UserProps =
	| { name: string; children?: never }
	| { name?: never; children: string };
export type CommentProps = { text: string; children?: never };
export type VolumeProps = { paths: readonly string[]; children?: never };
export type HealthCheckProps = {
	test: readonly string[];
	interval?: string;
	timeout?: string;
	startPeriod?: string;
	retries?: number;
	children?: never;
};
export type RunProps = {
	cmd: string;
	mount?: string | readonly string[];
	network?: string;
	security?: string;
	children?: never;
};
export type RunArgsProps = { args: readonly string[]; children?: never };
export type DirectiveProps = { value: string; children?: never };

export const From = instr<"dfFrom", FromProps>("dfFrom");
export function Arg(props: ArgProps): Node<"dfArg", { name: string; default?: string }> {
	return element("dfArg", resolveScalar(props, "default", { tag: "arg", required: false }) as { name: string; default?: string });
}
export function Env(props: EnvProps): Node<"dfEnv", { name: string; value: string }> {
	return element("dfEnv", resolveScalar(props, "value", { tag: "env", required: true }) as { name: string; value: string });
}
export const Copy = instr<"dfCopy", CopyProps>("dfCopy");
export const Add = instr<"dfAdd", AddProps>("dfAdd");
export const Entrypoint = instr<"dfEntrypoint", ExecProps>("dfEntrypoint");
export const EntrypointShell = instr<"dfEntrypointShell", ExecProps>("dfEntrypointShell");
export const Expose = instr<"dfExpose", ExposeProps>("dfExpose");
export function Label(props: LabelProps): Node<"dfLabel", { name: string; value: string }> {
	return element("dfLabel", resolveScalar(props, "value", { tag: "label", required: true }) as { name: string; value: string });
}
export const Onbuild = instr<"dfOnbuild", OnbuildProps>("dfOnbuild");
export function Workdir(props: WorkdirProps): Node<"dfWorkdir", { path: string }> {
	return element("dfWorkdir", resolveScalar(props, "path", { tag: "workdir", required: true }) as { path: string });
}
export const StopSignal = instr<"dfStopSignal", StopSignalProps>("dfStopSignal");
export function User(props: UserProps): Node<"dfUser", { name: string }> {
	return element("dfUser", resolveScalar(props, "name", { tag: "user", required: true }) as { name: string });
}
export const Comment = instr<"dfComment", CommentProps>("dfComment");
export const Volume = instr<"dfVolume", VolumeProps>("dfVolume");
export const HealthCheck = instr<"dfHealth", HealthCheckProps>("dfHealth");
export const HealthCheckNone = instr<"dfHealthNone", Record<string, never>>("dfHealthNone");
export const Shell = instr<"dfShell", ExecProps>("dfShell");
export const Command = instr<"dfCommand", ExecProps>("dfCommand");
export const CommandShell = instr<"dfCommandShell", ExecProps>("dfCommandShell");
export const Run = instr<"dfRun", RunProps>("dfRun");
export const RunLayer = instr<"dfRunLayer", RunProps>("dfRunLayer");
export const RunArgs = instr<"dfRunArgs", RunArgsProps>("dfRunArgs");
export const Syntax = instr<"dfSyntax", DirectiveProps>("dfSyntax");
export const Escape = instr<"dfEscape", DirectiveProps>("dfEscape");
export const Check = instr<"dfCheck", DirectiveProps>("dfCheck");

function instr<K extends string, P extends Record<string, unknown>>(kind: K) {
	return function instruction(props: P): Node<K, Omit<P, "children">> {
		return element(kind, props);
	};
}

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
