package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeOptions(root string, setters map[string]*setter, bags map[string]*bagInfo) error {
	names := make([]string, 0, len(bags))
	for n := range bags {
		names = append(names, n)
	}
	sort.Strings(names)

	kindToBag := map[string]string{}
	for _, b := range bags {
		kind := jsIdent(strings.TrimSuffix(b.iface, "Option"))
		kindToBag[kind] = b.tsName
	}

	fieldsByBag := map[string][]*setter{}
	for _, s := range setters {
		for bag := range s.actionBags {
			fieldsByBag[bag] = append(fieldsByBag[bag], s)
		}
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// %s\n\n", generatedBy)
	buf.WriteString("export type Duration = number | string;\n\n")
	buf.WriteString("/** OCI platform. Same object image inspect returns. */\n")
	buf.WriteString("export type ImagePlatform = {\n")
	buf.WriteString("\tarchitecture: string;\n")
	buf.WriteString("\tos: string;\n")
	buf.WriteString("\t\"os.version\"?: string;\n")
	buf.WriteString("\t\"os.features\"?: string[];\n")
	buf.WriteString("\tvariant?: string;\n")
	buf.WriteString("};\n\n")

	leafTypes := map[string][]leadParam{}
	for _, s := range setters {
		if s.leaf && len(s.fields) > 0 {
			leafTypes[exportedIdent(s.kind)+"Input"] = s.fields
		}
	}
	leafNames := make([]string, 0, len(leafTypes))
	for n := range leafTypes {
		leafNames = append(leafNames, n)
	}
	sort.Strings(leafNames)
	for _, n := range leafNames {
		fmt.Fprintf(&buf, "export type %s = {\n", n)
		for _, f := range leafTypes[n] {
			fmt.Fprintf(&buf, "\t%s?: %s;\n", f.name, f.tsType)
		}
		buf.WriteString("};\n\n")
	}

	for _, name := range names {
		fs := fieldsByBag[name]
		sort.Slice(fs, func(i, j int) bool {
			return fs[i].prop < fs[j].prop || (fs[i].prop == fs[j].prop && fs[i].name < fs[j].name)
		})
		fmt.Fprintf(&buf, "export type %s = {\n", name)
		seen := map[string]bool{}
		for _, s := range fs {
			key, ts, ok := bagField(s, kindToBag)
			if !ok || key == "" || seen[key] {
				continue
			}
			seen[key] = true
			fmt.Fprintf(&buf, "\t%s?: %s;\n", key, ts)
		}
		buf.WriteString("};\n\n")
	}

	out := filepath.Join(root, "tsx", "src", "generated-options.ts")
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d option bags)\n", out, len(names))
	return nil
}

func bagField(s *setter, kindToBag map[string]string) (key, ts string, ok bool) {
	switch {
	case s.nested:
		key = jsIdent(s.name)
		bag := ""
		if s.nestedIface != "" {
			bag = bagTSName(s.nestedIface)
		}
		if bag == "" {
			bag = kindToBag[s.propScope]
		}
		if bag == "" {
			bag = exportedIdent(s.kind) + "Options"
		}
		if s.prop != "" {
			ts = fmt.Sprintf("({ %s?: %s } & %s) | Array<{ %s?: %s } & %s>", s.prop, s.tsType, bag, s.prop, s.tsType, bag)
		} else {
			ts = bag
		}
		return key, ts, true
	case s.leaf:
		key = jsIdent(s.name)
		input := exportedIdent(s.kind) + "Input"
		return key, input + " | " + input + "[]", true
	case s.jsonSkip:
		key = jsIdent(s.name)
		return key, "unknown", true
	case len(s.goParams) == 0:
		return jsIdent(s.name), "boolean", true
	default:
		key = jsIdent(s.name)
		ts = s.tsType
		if ts == "" {
			ts = "unknown"
		}
		return key, ts, true
	}
}
