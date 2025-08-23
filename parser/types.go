package parser

import "go/token"

// FieldInfo describes a struct field
type FieldInfo struct {
	Name string
	Type string
	Tag  string
}

// StructInfo describes a parsed struct and its gofn directive (if any)
type StructInfo struct {
	Package   string
	Name      string
	Fields    []FieldInfo
	Directive string // raw value after //gofn:
	Pos       token.Position
}

// ParamInfo describes a function parameter or result
type ParamInfo struct {
	Name string
	Type string
}

// FuncInfo describes a parsed function and its gofn directive (if any)
type FuncInfo struct {
	Package   string
	Name      string
	Params    []ParamInfo
	Results   []ParamInfo
	Directive string
	Pos       token.Position
}
