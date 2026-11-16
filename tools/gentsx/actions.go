package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeActions(root string, verbs []verb) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// %s\n\n", generatedBy)
	buf.WriteString("import type { DockerNode } from \"./docker-node.js\";\n")
	buf.WriteString("import { render } from \"./render.js\";\n")
	buf.WriteString("import type { Client, PlanAction, PlanDocument } from \"./plan.js\";\n")
	buf.WriteString("import type {\n")
	seenBag := map[string]bool{}
	var bags []string
	for _, v := range verbs {
		if v.bag == "" || seenBag[v.bag] {
			continue
		}
		seenBag[v.bag] = true
		bags = append(bags, v.bag)
	}
	sort.Strings(bags)
	for i, b := range bags {
		if i == len(bags)-1 {
			fmt.Fprintf(&buf, "\t%s,\n", b)
			continue
		}
		fmt.Fprintf(&buf, "\t%s,\n", b)
	}
	buf.WriteString("} from \"./generated-options.js\";\n\n")

	writeJSDoc(&buf, "", "Name of one session action.\n\n@remarks\nThis is the `op` of a plan action. Each name has a function of the same name in this file.")
	buf.WriteString("export type PlanOp =\n")
	for _, v := range verbs {
		fmt.Fprintf(&buf, "\t| %q\n", v.op)
	}
	buf.WriteString(";\n\n")

	for _, v := range verbs {
		writeVerb(&buf, v)
	}

	out := filepath.Join(root, "tsx", "src", "generated-actions.ts")
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d verbs)\n", out, len(verbs))
	return nil
}

func writeVerb(buf *bytes.Buffer, v verb) {
	args := []string{"cli: Client"}
	if v.needsYAML {
		args = append(args, "node: DockerNode")
	}
	for _, p := range v.positionals {
		args = append(args, p.name+": "+p.tsType)
	}
	if v.bag != "" {
		args = append(args, "options?: "+v.bag)
	}

	writeJSDoc(buf, "", verbDoc(v))
	fmt.Fprintf(buf, "export function %s(%s): PlanDocument {\n", v.tsName, strings.Join(args, ", "))
	fmt.Fprintf(buf, "\tconst action: PlanAction = { op: %q };\n", v.op)
	if v.bag != "" {
		buf.WriteString("\tif (options) {\n\t\taction.options = options;\n\t}\n")
	}
	for _, p := range v.positionals {
		fmt.Fprintf(buf, "\taction.%s = %s;\n", p.name, p.name)
	}
	buf.WriteString("\tconst doc: PlanDocument = { connect: cli.options, actions: [action] };\n")
	if v.needsYAML {
		buf.WriteString("\tdoc.yaml = render(node);\n")
	}
	buf.WriteString("\treturn doc;\n")
	buf.WriteString("}\n\n")
}

func verbDoc(v verb) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s builds a plan for kit %s.\n", v.tsName, v.goName)
	b.WriteString("\n@remarks\nGo ImportPlan runs the plan. A session sends it as one action.\n")
	b.WriteString("\n@param cli - Connected client. `connect({ fromEnv: true })` reads the local Docker host.\n")
	if v.needsYAML {
		b.WriteString("@param node - Compose, service, job, or container tree. Its YAML is rendered onto the plan.\n")
	}
	for _, p := range v.positionals {
		fmt.Fprintf(&b, "@param %s - %s\n", p.name, paramNote(p))
	}
	if v.bag != "" {
		fmt.Fprintf(&b, "@param options - %s. Omit it to keep the kit defaults.\n", v.bag)
	}
	b.WriteString("@returns The plan. Pass it to `run` or `session.call`.\n")
	if v.needsYAML {
		b.WriteString("@throws Error when the tree cannot be rendered.\n")
	}
	b.WriteString("\n@example\n```ts\n")
	fmt.Fprintf(&b, "%s(cli", v.tsName)
	if v.needsYAML {
		b.WriteString(", tree")
	}
	for _, p := range v.positionals {
		fmt.Fprintf(&b, ", %s", exampleArg(p))
	}
	if v.bag != "" {
		b.WriteString(", {}")
	}
	b.WriteString(");\n```")
	return b.String()
}

func paramNote(p leadParam) string {
	switch p.name {
	case "id":
		return "Container, image, network, volume, or other object id"
	case "ref", "refs":
		return "Image reference"
	case "service":
		return "Compose service name"
	case "name":
		return "Object name"
	case "port":
		return "Container port"
	case "repository":
		return "Registry repository"
	case "path":
		return "Path inside the container or on the host"
	default:
		return p.name + ", typed " + p.tsType
	}
}

func exampleArg(p leadParam) string {
	switch {
	case strings.Contains(p.tsType, "string[]"):
		return `["example"]`
	case strings.Contains(p.tsType, "string"):
		return `"example"`
	case strings.Contains(p.tsType, "number"):
		return "1"
	case strings.Contains(p.tsType, "boolean"):
		return "true"
	default:
		return "value"
	}
}
