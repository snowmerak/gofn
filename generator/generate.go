package generator

import (
	"fmt"
	"os"
	"time"

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

// shouldGenerate returns (generate, reason, error)
// If sourcePath is empty or not found, we allow generation.
// If outPath exists and its modtime >= src modtime, skip generation.
func shouldGenerate(sourcePath, outPath string) (bool, string, error) {
	if sourcePath == "" {
		return true, "no-source-info", nil
	}
	srcInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, "source-not-found", nil
		}
		return true, "stat-source-failed", err
	}
	outInfo, err := os.Stat(outPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, "no-generated-file", nil
		}
		return true, "stat-out-failed", err
	}
	// If generated file is newer or equal to source, skip
	if !outInfo.ModTime().Before(srcInfo.ModTime()) {
		return false, fmt.Sprintf("up-to-date (gen: %s >= src: %s)", outInfo.ModTime().Format(time.RFC3339), srcInfo.ModTime().Format(time.RFC3339)), nil
	}
	return true, "outdated", nil
}
