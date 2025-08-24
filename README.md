# gofn

gofn is a powerful Go code generator that scans your source for `//gofn:` directives and emits functional programming helpers next to your files.

## Features

- **Record types**: Private structs → exported interfaces with constructors and getters
- **Optional configurations**: Functional options pattern for flexible struct initialization
- **Curried functions**: Transform regular functions into curried versions
- **Pipeline composition**: Compose stage functions with Result short-circuiting and advanced error handling
- **Pattern matching**: Rust-style pattern matching for structs with Option types
- **Smart generation**: Skips generation when output is up-to-date

## Installation

```bash
# Install the CLI
go install github.com/snowmerak/gofn/cmd/gofn@latest

# Or run directly from source
go run github.com/snowmerak/gofn/cmd/gofn -src=. -out=.
```

## Usage

```bash
# Generate code for current directory
gofn -src . -out .

# Or use go generate
go generate ./...
```

Generated files are named `<TypeOrFuncName>_<directive>_gen.go` and are automatically formatted.

## Directives

### 1. `//gofn:record` - Immutable Records

Transform private structs into immutable records with exported interfaces.

**Input:**
```go
//gofn:record
type person struct {
    name string
    age  int
}
```

**Generated:**
```go
type Person interface {
    Name() string
    Age() int
}

func NewPerson(name string, age int) Person {
    return &person{name: name, age: age}
}

func (p *person) Name() string { return p.name }
func (p *person) Age() int     { return p.age }
```

**Usage:**
```go
p := NewPerson("Alice", 30)
fmt.Println(p.Name(), p.Age()) // Alice 30
```

### 2. `//gofn:optional` - Functional Options

Generate functional options pattern for flexible struct initialization.

**Input:**
```go
//gofn:optional
type Config struct {
    Host string
    Port int
    Debug bool
}
```

**Generated:**
```go
type ConfigOption func(*Config)

func WithHost(host string) ConfigOption {
    return func(c *Config) { c.Host = host }
}

func WithPort(port int) ConfigOption {
    return func(c *Config) { c.Port = port }
}

func WithDebug(debug bool) ConfigOption {
    return func(c *Config) { c.Debug = debug }
}

func NewConfigWithOptions(opts ...ConfigOption) Config {
    config := Config{}
    for _, opt := range opts {
        opt(&config)
    }
    return config
}
```

**Usage:**
```go
cfg := NewConfigWithOptions(
    WithHost("localhost"),
    WithPort(8080),
    WithDebug(true),
)
```

### 3. `//gofn:curried` - Curried Functions

Transform regular functions into curried versions for partial application.

**Input:**
```go
//gofn:curried
func add(a int, b int) int {
    return a + b
}

//gofn:curried
func Concat(prefix string, parts ...string) string {
    result := prefix
    for _, part := range parts {
        result += part
    }
    return result
}

//gofn:curried
func DivMod(a, b int) (int, int) {
    return a / b, a % b
}
```

**Generated:**
```go
func AddCurried() func(int) func(int) int {
    return func(a int) func(int) int {
        return func(b int) int {
            return add(a, b)
        }
    }
}

func ConcatCurried() func(string) func(...string) string { /* ... */ }
func DivModCurried() func(int) func(int) (int, int) { /* ... */ }
```

**Usage:**
```go
// Partial application
addFive := AddCurried()(5)
result := addFive(3) // 8

// Direct usage
sum := AddCurried()(10)(20) // 30
text := ConcatCurried()("Hello, ")("World", "!") // "Hello, World!"
quotient, remainder := DivModCurried()(10)(3) // 3, 1
```

### 4. `//gofn:pipeline` - Pipeline Composition

Compose stage functions with automatic error short-circuiting using Result types. Includes advanced error handling capabilities.

**Input:**
```go
//gofn:pipeline
type DataPipeline struct {
    input  string
    parsed int
    result float64
}
```

**Generated:**
```go
// Basic composer - original functionality
func DataPipelineComposer(
    f1 func(string) monad.Result[int],
    f2 func(int) monad.Result[float64],
) func(string) monad.Result[float64] {
    return func(input string) monad.Result[float64] {
        v1, err := f1(input).Unwrap()
        if err != nil { return monad.Err[float64](err) }
        return f2(v1)
    }
}

// Enhanced composer with error handling
func DataPipelineComposerWithErrorHandler(
    f1 func(string) monad.Result[int],
    f2 func(int) monad.Result[float64],
    errorHandler func(int, error) monad.Result[float64],
) func(string) monad.Result[float64] {
    return func(input string) monad.Result[float64] {
        v1, err := f1(input).Unwrap()
        if err != nil {
            return errorHandler(1, err) // Stage 1 error
        }
        result := f2(v1)
        if !result.IsOk() {
            _, err := result.Unwrap()
            return errorHandler(2, err) // Stage 2 error
        }
        return result
    }
}

// Helper functions for common error handling patterns
func DataPipelineWithFallback(fallbackValue float64) func(int, error) monad.Result[float64] {
    return func(stageIndex int, err error) monad.Result[float64] {
        return monad.Ok(fallbackValue)
    }
}

func DataPipelineWithLogging(logger func(int, error)) func(int, error) monad.Result[float64] {
    return func(stageIndex int, err error) monad.Result[float64] {
        logger(stageIndex, err)
        return monad.Err[float64](err)
    }
}
```

**Basic Usage:**
```go
parseStr := func(s string) monad.Result[int] {
    if val, err := strconv.Atoi(s); err != nil {
        return monad.Err[int](err)
    } else {
        return monad.Ok(val)
    }
}

toFloat := func(i int) monad.Result[float64] {
    return monad.Ok(float64(i) * 1.5)
}

// Basic pipeline composition
pipeline := DataPipelineComposer(parseStr, toFloat)
result, err := pipeline("42").Unwrap() // 63.0, nil
```

**Advanced Error Handling:**
```go
// Pipeline with fallback on any error
fallbackHandler := DataPipelineWithFallback(0.0)
pipelineWithFallback := DataPipelineComposerWithErrorHandler(parseStr, toFloat, fallbackHandler)
result, _ := pipelineWithFallback("invalid").Unwrap() // 0.0, nil (recovered)

// Pipeline with logging
logHandler := DataPipelineWithLogging(func(stage int, err error) {
    fmt.Printf("Stage %d failed: %v
", stage, err)
})
pipelineWithLogging := DataPipelineComposerWithErrorHandler(parseStr, toFloat, logHandler)

// Custom error handling per stage
customHandler := func(stageIndex int, err error) monad.Result[float64] {
    switch stageIndex {
    case 1: // parseStr failed
        fmt.Printf("Parse error: %v, using default
", err)
        return monad.Ok(0.0) // Provide default value
    case 2: // toFloat failed  
        fmt.Printf("Conversion error: %v
", err)
        return monad.Err[float64](err) // Propagate error
    default:
        return monad.Err[float64](err)
    }
}
pipelineWithCustom := DataPipelineComposerWithErrorHandler(parseStr, toFloat, customHandler)
```

**Error Handler Features:**
- **Stage Index**: Know exactly which stage failed (1, 2, 3, ...)
- **Error Recovery**: Return a recovery value or transform the error
- **Stage-specific Logic**: Handle different stages with different strategies
- **Composability**: Chain multiple error handling strategies

### 5. `//gofn:match` - Pattern Matching

Generate Rust-style pattern matching for structs with Option types for wildcards.

**Input:**
```go
//gofn:match
type Address struct {
    Street string
    City   string
    Zip    string
}
```

**Generated:**
```go
type AddressMatcher struct {
    target Address
    matched bool
}

func (m *AddressMatcher) When(
    street monad.Option[string],
    city monad.Option[string], 
    zip monad.Option[string],
    action func(Address),
) *AddressMatcher {
    if m.matched { return m }
    
    if street.Match(m.target.Street) &&
       city.Match(m.target.City) &&
       zip.Match(m.target.Zip) {
        action(m.target)
        m.matched = true
    }
    return m
}

func (m *AddressMatcher) WhenGuard(
    street monad.Option[string],
    city monad.Option[string],
    zip monad.Option[string],
    guard func(Address) bool,
    action func(Address),
) *AddressMatcher { /* ... */ }

func (m *AddressMatcher) Default(action func(Address)) { /* ... */ }

func (a Address) Match() *AddressMatcher {
    return &AddressMatcher{target: a}
}
```

**Usage:**
```go
addr := Address{
    Street: "123 Main St",
    City:   "Seoul", 
    Zip:    "12345",
}

// Basic pattern matching
addr.Match().
    When(monad.S("123 Main St"), monad.S("Seoul"), monad.S("12345"), func(a Address) {
        fmt.Println("Exact match!")
    }).
    When(monad.W[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
        fmt.Println("Any Seoul address")
    }).
    When(monad.W[string](), monad.W[string](), monad.S("12345"), func(a Address) {
        fmt.Println("Zip code 12345")
    }).
    Default(func(a Address) {
        fmt.Println("Other address")
    })

// Pattern matching with guards
addr.Match().
    WhenGuard(monad.W[string](), monad.S("Seoul"), monad.W[string](),
        func(a Address) bool { return len(a.Street) > 10 },
        func(a Address) {
            fmt.Println("Seoul address with long street name")
        }).
    Default(func(a Address) {
        fmt.Println("Other pattern")
    })

// Pattern matching with return values
description := MatchAddressReturn[string](addr).
    When(monad.S("123 Main St"), monad.W[string](), monad.W[string](), func(a Address) string {
        return "Main Street address in " + a.City
    }).
    When(monad.W[string](), monad.S("Seoul"), monad.W[string](), func(a Address) string {
        return "Seoul: " + a.Street
    }).
    When(monad.W[string](), monad.W[string](), monad.S("12345"), func(a Address) string {
        return "Zip 12345: " + a.Street + ", " + a.City
    }).
    Default("Unknown address pattern")

fmt.Println(description) // "Main Street address in Seoul"
```

**Pattern Matching Helpers:**
- `monad.S[T](value)` - Match specific value (Some)
- `monad.N[T]()` - Explicit absence (None - doesn't match actual values)
- `monad.W[T]()` - Match any value (Wildcard - universal pattern)

**Understanding None vs Wildcard:**
```go
// Example with empty string
emptyAddr := Address{Street: "", City: "Seoul", Zip: "12345"}

emptyAddr.Match().
    When(monad.N[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
        // This won't match - None doesn't match empty string
        fmt.Println("None pattern")
    }).
    When(monad.S(""), monad.S("Seoul"), monad.W[string](), func(a Address) {
        // This matches - Some("") explicitly matches empty string
        fmt.Println("Explicit empty string match")
    }).
    When(monad.W[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
        // This would also match - Wildcard matches everything
        fmt.Println("Wildcard catches all")
    })
```

## Complete Example

```go
package main

import (
    "fmt"
    "strconv"
    "github.com/snowmerak/gofn/monad"
)

//go:generate gofn -src=. -out=.

//gofn:record
type user struct {
    name  string
    email string
    age   int
}

//gofn:optional
type ServerConfig struct {
    Host string
    Port int
    SSL  bool
}

//gofn:curried
func greet(prefix string, name string) string {
    return prefix + " " + name + "!"
}

//gofn:pipeline
type ProcessingPipeline struct {
    input   string
    parsed  int
    result  float64
}

//gofn:match
type LoginAttempt struct {
    Username string
    Success  bool
    IP       string
}

func main() {
    // Record usage
    u := NewUser("Alice", "alice@example.com", 25)
    fmt.Printf("User: %s (%s), Age: %d\n", u.Name(), u.Email(), u.Age())
    
    // Optional configuration
    cfg := NewServerConfigWithOptions(
        WithHost("0.0.0.0"),
        WithPort(443),
        WithSSL(true),
    )
    fmt.Printf("Server: %s:%d (SSL: %v)\n", cfg.Host, cfg.Port, cfg.SSL)
    
    // Curried functions
    sayHello := GreetCurried()("Hello")
    fmt.Println(sayHello("World")) // "Hello World!"
    
    // Pipeline with error handling
    parseInput := func(s string) monad.Result[int] {
        if val, err := strconv.Atoi(s); err != nil {
            return monad.Err[int](err)
        } else {
            return monad.Ok(val)
        }
    }
    
    processData := func(i int) monad.Result[float64] {
        if i < 0 {
            return monad.Err[float64](fmt.Errorf("negative input: %d", i))
        }
        return monad.Ok(float64(i) * 2.5)
    }
    
    // Pipeline with fallback error handling
    fallbackHandler := ProcessingPipelineWithFallback(-1.0)
    pipeline := ProcessingPipelineComposerWithErrorHandler(parseInput, processData, fallbackHandler)
    
    result, _ := pipeline("42").Unwrap()
    fmt.Printf("Pipeline result: %.1f
", result) // "Pipeline result: 105.0"
    
    result2, _ := pipeline("invalid").Unwrap()
    fmt.Printf("Pipeline fallback: %.1f
", result2) // "Pipeline fallback: -1.0"
    
    // Pattern matching
    attempt := LoginAttempt{
        Username: "admin",
        Success:  false,
        IP:       "192.168.1.100",
    }
    
    attempt.Match().
        When(monad.S("admin"), monad.S(false), monad.W[string](), func(l LoginAttempt) {
            fmt.Printf("Failed admin login from %s - ALERT!\n", l.IP)
        }).
        When(monad.W[string](), monad.S(true), monad.W[string](), func(l LoginAttempt) {
            fmt.Printf("Successful login: %s\n", l.Username)
        }).
        Default(func(l LoginAttempt) {
            fmt.Printf("Regular failed login: %s\n", l.Username)
        })
}
```

## Requirements

- Go 1.25+ (for generics support)
- The generated code depends on the `monad` package in this repository

## License

MIT License - see [LICENSE](LICENSE) file for details.