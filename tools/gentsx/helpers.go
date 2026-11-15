package main

import (
	"go/ast"
	"strings"
)

var bagAlias = map[string]string{
	"ComposerClientOption":      "ConnectOptions",
	"EnginePullOption":          "PullImageOptions",
	"EngineImageListOption":     "ListImagesOptions",
	"EngineBuildOption":         "BuildImageOptions",
	"EngineContainerListOption": "ListContainersOptions",
}

func bagTSName(iface string) string {
	if a := bagAlias[iface]; a != "" {
		return a
	}
	if strings.HasSuffix(iface, "Option") {
		return strings.TrimSuffix(iface, "Option") + "Options"
	}
	return iface + "Options"
}

func collectGoParams(fields *ast.FieldList) []goParam {
	if fields == nil {
		return nil
	}
	var out []goParam
	for _, f := range fields.List {
		typ := f.Type
		variadic := false
		if ell, ok := typ.(*ast.Ellipsis); ok {
			variadic = true
			typ = ell.Elt
		}
		names := f.Names
		if len(names) == 0 {
			out = append(out, goParam{
				name:     "value",
				tsType:   tsType(typ),
				goType:   typeName(typ),
				variadic: variadic,
				skip:     isSkipExpr(typ),
			})
			continue
		}
		for _, id := range names {
			out = append(out, goParam{
				name:     jsxPropName(id.Name),
				tsType:   tsType(typ),
				goType:   typeName(typ),
				variadic: variadic,
				skip:     isSkipExpr(typ),
			})
		}
	}
	return out
}

func paramsSkipJSON(ps []goParam) bool {
	for _, p := range ps {
		if p.skip {
			return true
		}
	}
	return false
}

func isSkipExpr(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.Ident:
		n := x.Name
		switch n {
		case "string", "bool", "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "byte", "rune",
			"float32", "float64":
			return false
		}
		if strings.HasSuffix(n, "Fn") || strings.HasSuffix(n, "Func") ||
			strings.HasSuffix(n, "Listener") || strings.HasSuffix(n, "Consumer") ||
			strings.HasSuffix(n, "Handler") {
			return true
		}
		return false
	case *ast.FuncType, *ast.ChanType, *ast.InterfaceType, *ast.MapType:
		return true
	case *ast.StarExpr:
		n := typeName(x.X)
		if n == "Logger" || strings.Contains(n, ".") {
			return true
		}
		return true
	case *ast.SelectorExpr:
		if id, ok := x.X.(*ast.Ident); ok {
			if id.Name == "time" && x.Sel.Name == "Duration" {
				return false
			}
			switch x.Sel.Name {
			case "Mode", "Proto", "Capability", "Policy":
				return false
			case "Action":
				return id.Name == "cascade"
			}
			switch id.Name {
			case "http", "net", "io", "trace", "otelhttp", "context", "engineclient", "composeapi", "ocispec":
				return true
			}
		}
		return true
	case *ast.ArrayType:
		return isSkipExpr(x.Elt)
	case *ast.Ellipsis:
		return isSkipExpr(x.Elt)
	default:
		return false
	}
}

func rewriteGenericParams(s *setter, tparams *ast.FieldList) {
	if s == nil || tparams == nil {
		return
	}
	names := map[string]bool{}
	for _, f := range tparams.List {
		for _, id := range f.Names {
			names[id.Name] = true
		}
	}
	for i, p := range s.goParams {
		if names[p.goType] {
			s.goParams[i].goType = "string"
		}
	}
}

func exportedIdent(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] = r[0] - 'a' + 'A'
	}
	return string(r)
}
