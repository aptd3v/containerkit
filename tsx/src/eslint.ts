import type { ESLint, Linter, Rule } from "eslint";
import {
	allowedChildren,
	composeComponents,
	dockerfileComponents,
	exclusiveChildProps,
	singleChildren,
} from "./generated.js";

// An editor hint for a tag written inline. A child held in a variable, returned
// from a helper, or spread onto the parent is invisible here; element() still
// rejects it when the tree is built.

type AstNode = { type: string; [key: string]: unknown };

type ModuleName = "compose" | "dockerfile" | "runtime";

type Binding = { module: ModuleName; exportName: string } | { module: ModuleName; namespace: true };

export type NestingProblem = { node: AstNode; message: string };

const skipKeys = new Set(["parent", "loc", "range", "start", "end", "tokens", "comments"]);

function isNode(value: unknown): value is AstNode {
	return typeof value === "object" && value !== null && typeof (value as { type?: unknown }).type === "string";
}

function moduleOf(source: string): ModuleName | undefined {
	const bare = source.replace(/\.js$/, "");
	if (bare === "containerkit") {
		return "compose";
	}
	if (bare === "containerkit/dockerfile") {
		return "dockerfile";
	}
	if (bare === "containerkit/jsx-runtime" || bare === "containerkit/jsx-dev-runtime") {
		return "runtime";
	}
	return undefined;
}

function ident(node: AstNode | undefined): string | undefined {
	if (node?.type === "JSXIdentifier" && typeof node.name === "string") {
		return node.name;
	}
	return undefined;
}

function bindings(program: AstNode): Map<string, Binding> {
	const locals = new Map<string, Binding>();
	for (const stmt of (program.body as AstNode[] | undefined) ?? []) {
		if (stmt.type !== "ImportDeclaration" || typeof stmt.source !== "object" || stmt.source === null) {
			continue;
		}
		const source = stmt.source as AstNode;
		if (typeof source.value !== "string") {
			continue;
		}
		const mod = moduleOf(source.value);
		if (!mod) {
			continue;
		}
		for (const spec of (stmt.specifiers as AstNode[] | undefined) ?? []) {
			const local = spec.local as AstNode | undefined;
			if (!local || typeof local.name !== "string") {
				continue;
			}
			if (spec.type === "ImportSpecifier") {
				const imported = spec.imported as AstNode | undefined;
				const exportName = imported && typeof imported.name === "string" ? imported.name : local.name;
				locals.set(local.name, { module: mod, exportName });
			} else if (spec.type === "ImportNamespaceSpecifier") {
				locals.set(local.name, { module: mod, namespace: true });
			}
		}
	}
	return locals;
}

type Tag = { node: AstNode; kind: string; text: string };

function resolveTag(name: AstNode, locals: Map<string, Binding>): { kind: string; text: string } | "fragment" | undefined {
	const root = name.type === "JSXMemberExpression" ? (name.object as AstNode) : name;
	const member = name.type === "JSXMemberExpression" ? ident(name.property as AstNode) : undefined;
	const rootName = ident(root);
	if (!rootName) {
		return undefined;
	}
	const binding = locals.get(rootName);
	if (!binding) {
		return undefined;
	}
	if ("namespace" in binding) {
		if (!member) {
			return undefined;
		}
		return kindOf(binding.module, member, `${rootName}.${member}`);
	}
	if (member) {
		return undefined;
	}
	return kindOf(binding.module, binding.exportName, rootName);
}

function kindOf(mod: ModuleName, exportName: string, text: string): { kind: string; text: string } | "fragment" | undefined {
	if (exportName === "Fragment" && (mod === "compose" || mod === "runtime")) {
		return "fragment";
	}
	const table = mod === "dockerfile" ? dockerfileComponents : composeComponents;
	if (mod === "runtime") {
		return undefined;
	}
	const kind = table[exportName];
	if (!kind) {
		return undefined;
	}
	return { kind, text };
}

function childTags(nodes: AstNode[] | undefined, locals: Map<string, Binding>): Tag[] {
	const out: Tag[] = [];
	const take = (node: AstNode | null | undefined): void => {
		if (!node) {
			return;
		}
		if (node.type === "JSXFragment" || isFragmentElement(node, locals)) {
			childTags(node.children as AstNode[] | undefined, locals).forEach((tag) => out.push(tag));
			return;
		}
		if (node.type === "JSXElement") {
			const resolved = resolveTag((node.openingElement as AstNode).name as AstNode, locals);
			if (resolved && resolved !== "fragment") {
				out.push({ node, kind: resolved.kind, text: resolved.text });
			}
			return;
		}
		if (node.type === "JSXExpressionContainer") {
			takeExpr(node.expression as AstNode | undefined);
		}
	};
	const takeExpr = (expr: AstNode | null | undefined): void => {
		if (!expr || expr.type === "JSXEmptyExpression") {
			return;
		}
		if (expr.type === "JSXElement" || expr.type === "JSXFragment") {
			take(expr);
			return;
		}
		if (expr.type === "LogicalExpression") {
			takeExpr(expr.left as AstNode);
			takeExpr(expr.right as AstNode);
			return;
		}
		if (expr.type === "ConditionalExpression") {
			takeExpr(expr.consequent as AstNode);
			takeExpr(expr.alternate as AstNode);
			return;
		}
		if (expr.type === "ArrayExpression") {
			for (const el of (expr.elements as Array<AstNode | null> | undefined) ?? []) {
				takeExpr(el ?? undefined);
			}
			return;
		}
		if (expr.type === "ParenthesizedExpression") {
			takeExpr(expr.expression as AstNode);
			return;
		}
		if (expr.type === "SequenceExpression") {
			for (const item of (expr.expressions as AstNode[] | undefined) ?? []) {
				takeExpr(item);
			}
		}
	};
	for (const node of nodes ?? []) {
		take(node);
	}
	return out;
}

function isFragmentElement(node: AstNode, locals: Map<string, Binding>): boolean {
	if (node.type !== "JSXElement") {
		return false;
	}
	const opening = node.openingElement as AstNode | undefined;
	if (!opening) {
		return false;
	}
	return resolveTag(opening.name as AstNode, locals) === "fragment";
}

function propSet(opening: AstNode, name: string): boolean {
	for (const attr of (opening.attributes as AstNode[] | undefined) ?? []) {
		if (attr.type !== "JSXAttribute") {
			continue;
		}
		const id = attr.name as AstNode | undefined;
		if (!id || id.name !== name) {
			continue;
		}
		if (attr.value === null || attr.value === undefined) {
			return true;
		}
		const value = attr.value as AstNode;
		if (value.type === "Literal") {
			return value.value !== "";
		}
		return true;
	}
	return false;
}

function alreadyMessage(message: string, child: string, parent: string): string {
	if (/^<[^>]+> is already set inside <[^>]+>$/.test(message)) {
		return `<${child}> is already set inside <${parent}>`;
	}
	return message;
}

type ResolvedTag = { kind: string; text: string };
type SingleGroup = { kinds: readonly string[]; message: string };

function reportNode(node: AstNode): AstNode {
	return (node.openingElement as AstNode) ?? node;
}

function invalidChild(child: Tag, parent: ResolvedTag): NestingProblem {
	return {
		node: reportNode(child.node),
		message: `<${child.text}> is not valid inside <${parent.text}>`,
	};
}

function duplicateChild(group: SingleGroup, child: Tag, parent: ResolvedTag): NestingProblem {
	return {
		node: reportNode(child.node),
		message: alreadyMessage(group.message, child.text, parent.text),
	};
}

function singleChildProblems(parent: ResolvedTag, kids: Tag[]): NestingProblem[] {
	const allowed = allowedChildren[parent.kind];
	if (!allowed) {
		return [];
	}
	const groups = singleChildren[parent.kind] ?? [];
	const seen = new Set<number>();
	const problems: NestingProblem[] = [];
	for (const child of kids) {
		if (!allowed.includes(child.kind)) {
			problems.push(invalidChild(child, parent));
			continue;
		}
		const problem = repeatedChild(groups, seen, child, parent);
		if (problem) {
			problems.push(problem);
		}
	}
	return problems;
}

function repeatedChild(groups: readonly SingleGroup[], seen: Set<number>, child: Tag, parent: ResolvedTag): NestingProblem | undefined {
	for (let i = 0; i < groups.length; i++) {
		const group = groups[i];
		if (!group?.kinds.includes(child.kind)) {
			continue;
		}
		if (seen.has(i)) {
			return duplicateChild(group, child, parent);
		}
		seen.add(i);
		return undefined;
	}
	return undefined;
}

function exclusiveProblem(parent: ResolvedTag, opening: AstNode, kids: Tag[]): NestingProblem | undefined {
	const exclusive = exclusiveChildProps[parent.kind];
	if (!exclusive) {
		return undefined;
	}
	const hasChild = kids.some((child) => child.kind === exclusive.child);
	const hasProp = exclusive.props.some((name) => propSet(opening, name));
	if (!hasChild || !hasProp) {
		return undefined;
	}
	return { node: opening, message: exclusive.message };
}

function elementProblems(node: AstNode, locals: Map<string, Binding>): NestingProblem[] {
	const opening = node.openingElement as AstNode;
	const parent = resolveTag(opening.name as AstNode, locals);
	if (!parent || parent === "fragment") {
		return [];
	}
	const kids = childTags(node.children as AstNode[] | undefined, locals);
	const problems = singleChildProblems(parent, kids);
	const exclusive = exclusiveProblem(parent, opening, kids);
	if (exclusive) {
		problems.push(exclusive);
	}
	return problems;
}

function walk(node: AstNode, visit: (node: AstNode) => void): void {
	for (const [key, value] of Object.entries(node)) {
		if (skipKeys.has(key)) {
			continue;
		}
		if (Array.isArray(value)) {
			for (const item of value) {
				if (isNode(item)) {
					visit(item);
				}
			}
			continue;
		}
		if (isNode(value)) {
			visit(value);
		}
	}
}

export function nestingProblems(program: AstNode): NestingProblem[] {
	const locals = bindings(program);
	const problems: NestingProblem[] = [];
	const visit = (node: AstNode): void => {
		if (node.type === "JSXElement") {
			problems.push(...elementProblems(node, locals));
		}
		walk(node, visit);
	};
	visit(program);
	return problems;
}

const nesting: Rule.RuleModule = {
	meta: {
		type: "problem",
		docs: {
			description: "Disallow a containerkit child the parent does not accept.",
		},
		schema: [],
		messages: {
			nesting: "{{text}}",
		},
	},
	create(context) {
		return {
			Program(node) {
				for (const problem of nestingProblems(node as unknown as AstNode)) {
					context.report({
						node: problem.node as unknown as Rule.Node,
						messageId: "nesting",
						data: { text: problem.message },
					});
				}
			},
		};
	},
};

const plugin: ESLint.Plugin = {
	meta: { name: "containerkit" },
	rules: { nesting },
};

const recommended: Linter.Config[] = [
	{
		files: ["**/*.{jsx,tsx}"],
		plugins: { containerkit: plugin },
		rules: { "containerkit/nesting": "error" },
	},
];

plugin.configs = { recommended };

export default plugin;
export { plugin, recommended };
