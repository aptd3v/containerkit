package kit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestOptionValueTypesExported(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	pkg := pkgs["kit"]
	if pkg == nil {
		t.Fatal("package kit not found")
	}

	optionMethods := map[string]string{}
	docs := map[string]string{}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				doc := ts.Doc
				if doc == nil {
					doc = gd.Doc
				}
				if doc != nil {
					docs[ts.Name.Name] = doc.Text()
				}
				iface, ok := ts.Type.(*ast.InterfaceType)
				if !ok || !strings.HasSuffix(ts.Name.Name, "Option") || iface.Methods == nil {
					continue
				}
				for _, m := range iface.Methods.List {
					if len(m.Names) == 0 {
						continue
					}
					optionMethods[m.Names[0].Name] = ts.Name.Name
				}
			}
		}
	}
	if len(optionMethods) == 0 {
		t.Fatal("no option interfaces found")
	}

	skip := map[string]bool{
		"Grouped":    true,
		"Container":  true,
		"Dockerfile": true,
		"spec":       true,
	}

	implementers := map[string][]string{}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 || fd.Name == nil {
				continue
			}
			iface, ok := optionMethods[fd.Name.Name]
			if !ok {
				continue
			}
			recv := receiverTypeName(fd.Recv.List[0].Type)
			if recv == "" || skip[recv] || strings.HasSuffix(recv, "Act") {
				continue
			}
			implementers[recv] = append(implementers[recv], iface)
		}
	}
	if len(implementers) == 0 {
		t.Fatal("no setter option types found")
	}

	var unexported []string
	var undocumented []string
	for name, ifaces := range implementers {
		r, _ := utf8.DecodeRuneInString(name)
		if !unicode.IsUpper(r) {
			unexported = append(unexported, name+" ("+strings.Join(unique(ifaces), ", ")+")")
			continue
		}
		doc := docs[name]
		if !strings.HasPrefix(doc, name+" ") {
			undocumented = append(undocumented, name)
		}
	}
	if len(unexported) > 0 {
		t.Errorf("unexported setter option types (export them for pkg.go.dev Implemented by):\n  %s", strings.Join(unexported, "\n  "))
	}
	if len(undocumented) > 0 {
		t.Errorf("exported setter option types missing godoc starting with the type name:\n  %s", strings.Join(undocumented, "\n  "))
	}
}

func receiverTypeName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return receiverTypeName(x.X)
	default:
		return ""
	}
}

func unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
