# gofn

Small tool to parse `//gofn:` directives and generate helper code (constructors, curried wrappers).

Usage:

1. Add `//gofn:` comments above structs or functions.
2. Run the CLI: `go run ./cmd/gofn -src=. -out=./gen`
