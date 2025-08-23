package generator

import (
	"os"

	"github.com/snowmerak/gofn/parser"
)

// GenerateFor orchestrates generation for structs and funcs
func GenerateFor(outDir string, structs []parser.StructInfo, funcs []parser.FuncInfo) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	if err := generateStructs(outDir, structs); err != nil {
		return err
	}
	if err := generateFuncs(outDir, funcs); err != nil {
		return err
	}
	return nil
}
