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
	args = append(args, "options?: "+v.bag)

	fmt.Fprintf(buf, "/** %s encodes kit %s as a plan document. Go ImportPlan / ExecJSON runs it. */\n", v.tsName, v.goName)
	fmt.Fprintf(buf, "export function %s(%s): PlanDocument {\n", v.tsName, strings.Join(args, ", "))
	fmt.Fprintf(buf, "\tconst action: PlanAction = { op: %q };\n", v.op)
	buf.WriteString("\tif (options) {\n\t\taction.options = options;\n\t}\n")
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
