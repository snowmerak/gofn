package generator

import (
	"fmt"
	"go/format"
	"strings"
	"unicode"

	"github.com/snowmerak/gofn/parser"
)

// small helpers
func paramsForFields(fields []parser.FieldInfo) string {
	parts := []string{}
	for _, f := range fields {
		name := f.Name
		if name == "" {
			name = "_f"
		}
		parts = append(parts, fmt.Sprintf("%s %s", strings.ToLower(name), f.Type))
	}
	return strings.Join(parts, ", ")
}

func valuesForFields(fields []parser.FieldInfo) string {
	parts := []string{}
	for _, f := range fields {
		name := f.Name
		if name == "" {
			name = "_f"
		}
		parts = append(parts, strings.ToLower(name)+": "+strings.ToLower(name))
	}
	return strings.Join(parts, ", ")
}

func generateCurriedFunc(f parser.FuncInfo) string {
	var b strings.Builder
	n := len(f.Params)
	resCount := len(f.Results)

	// helper to build remaining nested type starting at index i
	remainingType := func(i int) string {
		var sb strings.Builder
		for j := i; j < n; j++ {
			sb.WriteString("func(")
			// if this param is variadic, it should be represented with ellipsis
			ptype := f.Params[j].Type
			sb.WriteString(paramName(f.Params[j], j))
			sb.WriteString(" ")
			sb.WriteString(ptype)
			sb.WriteString(") ")
		}
		// append result types
		if resCount == 1 {
			sb.WriteString(f.Results[0].Type)
		} else if resCount > 1 {
			// multiple results: (t1, t2, ...)
			parts := []string{}
			for _, r := range f.Results {
				parts = append(parts, r.Type)
			}
			sb.WriteString("(" + strings.Join(parts, ", ") + ")")
		}
		return sb.String()
	}

	b.WriteString("// Generated curried wrapper for " + f.Name + "\n")
	// exported wrapper name (capitalize original name then append Curried)
	wrapperName := exportName(f.Name) + "Curried"

	// Top-level signature
	if n == 0 {
		// no params: just return original result directly
		if resCount == 0 {
			b.WriteString("func " + wrapperName + "() {")
		} else {
			b.WriteString("func " + wrapperName + "() " + f.Results[0].Type + " {")
		}
		b.WriteString("\n    ")
		if resCount == 0 {
			b.WriteString(f.Name + "()\n")
		} else {
			b.WriteString("return " + f.Name + "()\n")
		}
		b.WriteString("}\n")
		return b.String()
	}

	// signature: func NameCurried() <nested type>
	b.WriteString("func " + wrapperName + "() " + remainingType(0) + " {\n")

	// body: produce nested "return func(...) <remaining> {" lines
	for i := 0; i < n; i++ {
		indent := strings.Repeat("    ", i+1)
		b.WriteString(indent + "return func(")
		// if this param is variadic (starts with ...), keep the ellipsis in the type
		ptype := f.Params[i].Type
		b.WriteString(paramName(f.Params[i], i) + " " + ptype + ") ")
		// remaining return type after this param
		rem := remainingType(i + 1)
		if rem != "" {
			b.WriteString(rem)
		}
		b.WriteString(" {\n")
	}

	// innermost: call original function
	innIndent := strings.Repeat("    ", n+1)
	if resCount == 0 {
		b.WriteString(innIndent + f.Name + "(")
	} else {
		b.WriteString(innIndent + "return " + f.Name + "(")
	}
	// arguments are parameter names p0..pn-1
	args := []string{}
	for i := 0; i < n; i++ {
		// if param type is variadic (starts with ...), expand when forwarding: use 'arg...' in call
		pname := paramName(f.Params[i], i)
		if strings.HasPrefix(f.Params[i].Type, "...") {
			args = append(args, pname+"...")
		} else {
			args = append(args, pname)
		}
	}
	b.WriteString(strings.Join(args, ", ") + ")\n")

	// close braces
	for i := n - 1; i >= 0; i-- {
		indent := strings.Repeat("    ", i+1)
		b.WriteString(indent + "}\n")
	}

	// close outer function
	b.WriteString("}\n")

	return b.String()
}

func paramName(p parser.ParamInfo, i int) string {
	if p.Name != "" {
		return p.Name
	}
	return fmt.Sprintf("p%d", i)
}

func formatSource(src []byte) ([]byte, error) {
	out, err := format.Source(src)
	if err != nil {
		// return original with error so caller can decide
		return src, fmt.Errorf("gofn: format error: %w", err)
	}
	return out, nil
}

func normalizeDirective(d string) string {
	// keep alnum and replace others with underscore, and lowercase
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(d)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func exportName(s string) string {
	if s == "" {
		return s
	}
	rs := []rune(s)
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

func fieldParamName(field string, i int) string {
	if field != "" {
		// if field already starts with lowercase, use as-is; otherwise lowercase first rune
		rs := []rune(field)
		rs[0] = unicode.ToLower(rs[0])
		return string(rs)
	}
	return fmt.Sprintf("f%d", i)
}

func isPrivateIdent(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsLower(r)
}
