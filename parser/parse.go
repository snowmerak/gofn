package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// ParseDir scans a directory for Go files and returns structs and funcs with //gofn: directives
func ParseDir(dir string) ([]StructInfo, []FuncInfo, error) {
	fset := token.NewFileSet()
	var structs []StructInfo
	var funcs []FuncInfo

	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, nil, err
	}

	for _, f := range files {
		src, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, nil, err
		}
		file, err := parser.ParseFile(fset, f, src, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}

		pkg := file.Name.Name

		// comments are inspected per-declaration below using x.Doc on nodes

		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if st, ok := x.Type.(*ast.StructType); ok {
					pos := fset.Position(x.Pos())
					dir := ""
					// try to find preceding comment for the type
					if x.Doc != nil {
						for _, c := range x.Doc.List {
							txt := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
							if strings.HasPrefix(txt, "gofn:") {
								dir = strings.TrimSpace(strings.TrimPrefix(txt, "gofn:"))
								break
							}
						}
					}
					// If TypeSpec.Doc is empty, the comment may be attached to the enclosing GenDecl
					if dir == "" {
						// search file declarations to find the GenDecl that contains this TypeSpec
						for _, decl := range file.Decls {
							gd, ok := decl.(*ast.GenDecl)
							if !ok || gd.Doc == nil {
								continue
							}
							for _, spec := range gd.Specs {
								if ts, ok := spec.(*ast.TypeSpec); ok && ts == x {
									for _, c := range gd.Doc.List {
										txt := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
										if strings.HasPrefix(txt, "gofn:") {
											dir = strings.TrimSpace(strings.TrimPrefix(txt, "gofn:"))
											break
										}
									}
								}
								if dir != "" {
									break
								}
							}
							if dir != "" {
								break
							}
						}
					}
					fields := []FieldInfo{}
					for _, f := range st.Fields.List {
						t := exprString(f.Type)
						tag := ""
						if f.Tag != nil {
							tag = strings.Trim(f.Tag.Value, "`\"")
						}
						if len(f.Names) == 0 {
							fields = append(fields, FieldInfo{Name: "", Type: t, Tag: tag})
						} else {
							for _, nm := range f.Names {
								fields = append(fields, FieldInfo{Name: nm.Name, Type: t, Tag: tag})
							}
						}
					}
					structs = append(structs, StructInfo{Package: pkg, Name: x.Name.Name, Fields: fields, Directive: dir, Pos: pos})
				}
			case *ast.FuncDecl:
				pos := fset.Position(x.Pos())
				dir := ""
				if x.Doc != nil {
					for _, c := range x.Doc.List {
						txt := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
						if strings.HasPrefix(txt, "gofn:") {
							dir = strings.TrimSpace(strings.TrimPrefix(txt, "gofn:"))
							break
						}
					}
				}
				params := []ParamInfo{}
				if x.Type.Params != nil {
					for _, p := range x.Type.Params.List {
						t := exprString(p.Type)
						if len(p.Names) == 0 {
							params = append(params, ParamInfo{Name: "", Type: t})
						} else {
							for _, n := range p.Names {
								params = append(params, ParamInfo{Name: n.Name, Type: t})
							}
						}
					}
				}
				results := []ParamInfo{}
				if x.Type.Results != nil {
					for _, r := range x.Type.Results.List {
						t := exprString(r.Type)
						if len(r.Names) == 0 {
							results = append(results, ParamInfo{Name: "", Type: t})
						} else {
							for _, n := range r.Names {
								results = append(results, ParamInfo{Name: n.Name, Type: t})
							}
						}
					}
				}
				funcs = append(funcs, FuncInfo{Package: pkg, Name: x.Name.Name, Params: params, Results: results, Directive: dir, Pos: pos})
			}
			return true
		})
	}

	return structs, funcs, nil
}

// exprString renders a limited set of expr types to string for type names
func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.Ellipsis:
		return "..." + exprString(t.Elt)
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.MapType:
		return "map[" + exprString(t.Key) + "]" + exprString(t.Value)
	case *ast.FuncType:
		return "func"
	default:
		return "<unknown>"
	}
}
