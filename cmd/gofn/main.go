package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowmerak/gofn/generator"
	"github.com/snowmerak/gofn/parser"
)

func main() {
	src := flag.String("src", ".", "source directory to scan")
	out := flag.String("out", "", "output directory for generated code (defaults to src)")
	flag.Parse()
	absSrc, _ := filepath.Abs(*src)
	if *out == "" {
		*out = absSrc
	}
	structs, funcs, err := parser.ParseDir(absSrc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse error:", err)
		os.Exit(2)
	}

	if err := generator.GenerateFor(*out, structs, funcs); err != nil {
		fmt.Fprintln(os.Stderr, "generate error:", err)
		os.Exit(3)
	}

	fmt.Println("generated to", *out)
}
