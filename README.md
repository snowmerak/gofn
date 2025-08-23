# gofn

gofn is a tiny Go code generator. It scans your source for `//gofn:` directives and emits helper code next to your files.

Highlights
- Understands `//gofn:` on types and functions.
- Record-style constructors with an exported interface (for private structs).
- Functional Options helpers (`optional`).
- Curried wrappers for functions (`curried`) — supports variadic and multi-result.
- Pipeline composers (`pipeline`) that compose stage functions `func(Ti) monad.Result[T{i+1}]` into `func(T1) monad.Result[Tn]` with early error short-circuiting.
- Skips generation when output is up-to-date.

Install / Run

```fish
# Dev run (no install):
go run ./cmd/gofn -src=. -out=.

# Install the CLI then run:
go install ./cmd/gofn
gofn -src . -out .
```

Generated files
- Named `<TypeOrFuncName>_<directive>_gen.go`, e.g. `anyPipe_pipeline_gen.go`.
- Written to `-out` (default: current directory). Files are formatted and only rewritten when stale.

Directives and examples

1) record — private struct → exported interface + constructor + getters

Input
```go
//gofn:record
type person struct { // must be private
		name string
		age  int
}
```

Output (summary)
```go
type Person interface {
		Name() string
		Age() int
}

func NewPerson(name string, age int) Person { ... }
```

Usage
```go
p := NewPerson("alice", 30)
fmt.Println(p.Name(), p.Age())
```

2) optional — functional options for a struct

Input
```go
//gofn:optional
type Config struct {
		Host string
		Port int
}
```

Output (summary)
```go
type ConfigOption func(*Config)

func WithHost(host string) ConfigOption { return func(c *Config) { c.Host = host } }
func WithPort(port int)   ConfigOption { return func(c *Config) { c.Port = port } }

func NewConfigWithOptions(opts ...ConfigOption) Config { ... }
```

Usage
```go
cfg := NewConfigWithOptions(
	WithHost("localhost"),
	WithPort(8080),
)
```

3) curried — exported curried wrappers for functions

Input
```go
//gofn:curried
func Concat(prefix string, parts ...string) string { /* ... */ }

//gofn:curried
func DivMod(a, b int) (int, int) { /* ... */ }
```

Output (summary)
```go
func ConcatCurried() func(string) func(...string) string { ... }
func DivModCurried() func(int) func(int) (int, int) { ... }
```

Usage
```go
s := ConcatCurried()("pre-")("a", "b") // "pre-ab"
q, r := DivModCurried()(10)(3)            // 3, 1
```

4) pipeline — compose stage functions with Result short‑circuiting

Input
```go
//gofn:pipeline
type anyPipe struct {
		first  int64
		second string
		third  float32
		fourth bool
}
```

Output (signature)
```go
func AnyPipeComposer(
	f1 func(int64)  monad.Result[string],
	f2 func(string) monad.Result[float32],
	f3 func(float32) monad.Result[bool],
) func(int64) monad.Result[bool]
```

Usage
```go
f1 := func(x int64) monad.Result[string]   { return monad.Ok(fmt.Sprint(x)) }
f2 := func(s string) monad.Result[float32] { return monad.Ok(float32(len(s))) }
f3 := func(f float32) monad.Result[bool]   { return monad.Ok(f > 0) }

pipe := AnyPipeComposer(f1, f2, f3)
ok, err := pipe(42).Unwrap()  // ok: true, err: nil
```

Notes
- record: enforced only for private struct names and private fields; otherwise skipped.
- curried: variadic parameters are forwarded using `arg...`; multi‑result functions are supported.
- pipeline: short‑circuits on the first error; depends on this repo’s `monad` package (`Result[T]`, `Ok`, `Err`, `Unwrap`).
- Up‑to‑date skip: generation compares source vs. output modification times.

Try it

```fish
# Generate
go run ./cmd/gofn -src=. -out=.

# Build everything
go build ./...
```

License
MIT

