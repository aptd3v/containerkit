package main

import (
	"go/ast"
	"sort"
	"strings"
)

var verbRename = map[string]string{
	"Compose.Run":    "composeRun",
	"Compose.Exec":   "composeExec",
	"Compose.Export": "composeExport",
	"Client.Run":     "containerRun",
	"Client.Load":    "load",
	"export":         "composeExport",
	"import":         "composeImport",
	"delete":         "deleteResource",
	"debugger":       "debugAction",
}

var tsReserved = map[string]bool{
	"break": true, "case": true, "catch": true, "class": true, "const": true,
	"continue": true, "debugger": true, "default": true, "delete": true,
	"do": true, "else": true, "enum": true, "export": true, "extends": true,
	"false": true, "finally": true, "for": true, "function": true, "if": true,
	"import": true, "in": true, "instanceof": true, "new": true, "null": true,
	"return": true, "super": true, "switch": true, "this": true, "throw": true,
	"true": true, "try": true, "typeof": true, "var": true, "void": true,
	"while": true, "with": true, "yield": true, "await": true, "let": true,
	"static": true, "implements": true, "interface": true, "package": true,
	"private": true, "protected": true, "public": true,
}

func parseVerbs(files []*ast.File, bags map[string]*bagInfo) []verb {
	ifaceBag := map[string]string{}
	for _, b := range bags {
		ifaceBag[b.iface] = b.tsName
	}

	byName := map[string]verb{}
	for _, file := range files {
		for _, decl := range file.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok || d.Name == nil || !d.Name.IsExported() || d.Type.Params == nil {
				continue
			}
			recv := ""
			if d.Recv != nil {
				recv = recvName(d.Recv)
			}
			if recv == "Grouped" || recv == "Loaded" || recv == "Plan" || recv == "Dockerfile" {
				continue
			}
			v, ok := classifyVerb(d, recv, ifaceBag)
			if !ok {
				continue
			}
			if existing, dup := byName[v.tsName]; dup {
				kept, extra := existing, v
				if verbScore(v) > verbScore(existing) {
					kept, extra = v, existing
				}
				byName[kept.tsName] = kept
				if dropComposeDuplicate(kept, extra) {
					continue
				}
				alt := disambiguate(extra)
				if alt == kept.tsName {
					alt += "Alt"
				}
				if _, taken := byName[alt]; taken {
					alt += "2"
				}
				extra.tsName = alt
				extra.op = alt
				byName[alt] = extra
				continue
			}
			byName[v.tsName] = v
		}
	}
	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]verb, 0, len(names))
	for _, n := range names {
		out = append(out, byName[n])
	}
	return out
}

func classifyVerb(d *ast.FuncDecl, recv string, ifaceBag map[string]string) (verb, bool) {
	params := d.Type.Params.List
	var bag, iface string
	var positionals []leadParam
	var slots []callSlot
	needsYAML := recv == "Compose" || (recv == "Client" && hasComposeParam(params))
	engine := false
	sawCtx := false

	for _, f := range params {
		typ := f.Type
		ell, variadic := typ.(*ast.Ellipsis)
		if variadic {
			typ = ell.Elt
		}
		tn := typeName(typ)
		switch tn {
		case "Context", "context.Context":
			sawCtx = true
			slots = append(slots, callSlot{kind: "ctx"})
			continue
		case "Engine":
			engine = true
			slots = append(slots, callSlot{kind: "engine"})
			continue
		case "Composer":
			needsYAML = true
			slots = append(slots, callSlot{kind: "composer"})
			continue
		case "Compose":
			needsYAML = true
			slots = append(slots, callSlot{kind: "compose"})
			continue
		case "Client":
			continue
		}
		if variadic && strings.HasSuffix(tn, "Option") {
			if b, ok := ifaceBag[tn]; ok {
				bag = b
				iface = tn
			}
			slots = append(slots, callSlot{kind: "opts"})
			continue
		}
		if skipVerbParam(tn) {
			return verb{}, false
		}
		n := "value"
		if len(f.Names) > 0 {
			n = jsIdent(f.Names[0].Name)
		}
		positionals = append(positionals, leadParam{name: n, tsType: tsType(typ), goType: tn})
		slots = append(slots, callSlot{kind: "pos", name: n, goType: tn})
	}

	bare := false
	if bag == "" {
		if !bareInspect(d.Name.Name) {
			return verb{}, false
		}
		bare = true
	}
	if !sawCtx && !engine {
		return verb{}, false
	}
	if recv == "Compose" && !sawCtx {
		return verb{}, false
	}

	goName := d.Name.Name
	key := goName
	if recv != "" {
		key = recv + "." + goName
	}
	tsName := jsIdent(goName)
	if r := verbRename[key]; r != "" {
		tsName = r
	} else if r := verbRename[tsName]; r != "" {
		tsName = r
	}
	if tsReserved[tsName] {
		tsName = "compose" + exportedIdent(tsName)
	}
	tsName = composeCommandName(tsName, needsYAML)
	return verb{
		tsName:      tsName,
		op:          tsName,
		goName:      goName,
		bag:         bag,
		iface:       iface,
		recv:        recv,
		needsYAML:   needsYAML,
		engine:      engine && recv == "",
		bare:        bare,
		positionals: positionals,
		slots:       slots,
		results:     resultTypes(d),
	}, true
}

// bareInspect is an Engine read that returns JSON and takes no option bag.
// The generator otherwise drops a command that has nothing to decode.
func bareInspect(name string) bool {
	switch name {
	case "InspectSwarm", "InspectPlugin", "InspectNode", "InspectSwarmService",
		"InspectTask", "InspectSecret", "InspectConfig":
		return true
	default:
		return false
	}
}

func verbScore(v verb) int {
	score := 0
	if v.needsYAML {
		score += 10
	}
	switch v.recv {
	case "Client":
		score += 5
	case "Compose":
		score += 3
	case "Container":
		score += 1
	}
	return score
}

// composeCommandName prefixes a compose-project command. Engine and container
// commands are left alone. Names that already start with compose stay put.
func composeCommandName(name string, needsYAML bool) string {
	if !needsYAML || strings.HasPrefix(name, "compose") {
		return name
	}
	return "compose" + exportedIdent(name)
}

func dropComposeDuplicate(kept, extra verb) bool {
	// Client.Up takes the compose file and calls Compose.Up. Keeping both
	// emits composeUp and composeUpCompose for one action.
	return kept.recv == "Client" && extra.recv == "Compose" && kept.goName == extra.goName
}

func disambiguate(v verb) string {
	switch v.recv {
	case "":
		if v.engine {
			return v.tsName + "Engine"
		}
		return v.tsName + "Fn"
	default:
		return v.tsName + v.recv
	}
}

func resultTypes(d *ast.FuncDecl) []string {
	if d.Type.Results == nil {
		return nil
	}
	var out []string
	for _, f := range d.Type.Results.List {
		n := len(f.Names)
		if n == 0 {
			n = 1
		}
		tn := typeName(f.Type)
		for i := 0; i < n; i++ {
			out = append(out, tn)
		}
	}
	return out
}

func hasComposeParam(params []*ast.Field) bool {
	for _, f := range params {
		if typeName(f.Type) == "Compose" {
			return true
		}
	}
	return false
}

func skipVerbParam(tn string) bool {
	switch {
	case tn == "", strings.Contains(tn, "."):
		if tn == "context.Context" {
			return false
		}
		return tn != "time.Duration"
	case tn == "Logger", tn == "LogConsumer", tn == "ImageImportSource":
		return true
	default:
		return false
	}
}
