# GoFn Architecture and Design Patterns

## Core Architecture

### Parser Package (`parser/`)
- **Purpose**: Parses Go source files to extract `//gofn:` directives
- **Key Types**:
  - `StructInfo`: Describes parsed structs with directive information
  - `FuncInfo`: Describes parsed functions with directive information
  - `FieldInfo`, `ParamInfo`: Supporting data structures
- **Main Function**: `ParseDir(path)` - Scans directory for Go files and extracts directive info

### Generator Package (`generator/`)
- **Purpose**: Generates Go code based on parsed directive information
- **Main Function**: `GenerateFor(outDir, structs, funcs)` - Orchestrates code generation
- **Pattern**: Separate generation logic for structs vs functions
- **Output**: Writes `*_gen.go` files with generated helper code

### Monad Package (`monad/`)
- **Purpose**: Provides `Result[T]` type for error handling in pipelines
- **Key Types**:
  - `Result[T]`: Generic result type with value and error
  - `Pipeline[T]`: Wrapper for chaining operations
- **Functions**: `Ok()`, `Err()`, `Map()`, `AndThen()` for functional composition

## Design Patterns

### Code Generation Pattern
1. **Directive-based**: Uses special comments (`//gofn:directive`) to mark generation targets
2. **File-per-pattern**: Each directive type generates a separate `*_gen.go` file
3. **Timestamp-based**: Only regenerates when source is newer than generated file
4. **Non-intrusive**: Generated code exists alongside original, doesn't modify source

### Functional Programming Patterns
- **Currying**: Transforms multi-argument functions into chain of single-argument functions
- **Options Pattern**: Uses functional options for flexible struct configuration
- **Railway-Oriented Programming**: `Result[T]` type for explicit error handling
- **Composition**: Pipeline pattern composes multiple stages with error short-circuiting

### Interface Segregation
- **Record Pattern**: Private structs exposed through minimal public interfaces
- **Single Responsibility**: Each interface method corresponds to one struct field
- **Encapsulation**: Private fields accessed only through generated getter methods

## CLI Design
- **Simple Interface**: Two main flags (`-src`, `-out`)
- **Sensible Defaults**: Output defaults to source directory
- **Clear Feedback**: Reports generation status and file timestamps
- **Error Handling**: Proper exit codes for different error types