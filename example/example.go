package main

import (
	"errors"
	"fmt"

	"github.com/snowmerak/gofn/monad"
)

//go:generate go run github.com/snowmerak/gofn/cmd/gofn -src=. -out=.

// Example definitions with gofn directives. The runnable demo is at the bottom (runExamples).

//gofn:record
type person struct {
	name string
	age  int
}

//gofn:optional
type Config struct {
	Host string
	Port int
}

// 필수 인자를 받는 생성자와 옵션 기반 생성자(WithX helpers)는
// gofn 실행 시 생성됩니다.

//gofn:curried
func add(a int, b int) int {
	return a + b
}

// 추가 예제 함수들: variadic 및 multi-result (curried 래퍼는 gofn으로 생성)
//
//gofn:curried
func Concat(prefix string, parts ...string) string {
	out := prefix
	for _, p := range parts {
		out += p
	}
	return out
}

//gofn:curried
func DivMod(a, b int) (int, int) {
	return a / b, a % b
}

//gofn:pipeline
type anyPipe struct {
	first  int64
	second string
	third  float32
	fourth bool
}

//gofn:match
type Address struct {
	Street string
	City   string
	Zip    string
}

// Demo: exercise all generated helpers.
func main() {
	// record: exported interface + constructor + getters
	p := NewPerson("alice", 30)
	fmt.Println("record:", p.Name(), p.Age())

	// optional: functional options constructor
	cfg := NewConfigWithOptions(
		WithHost("localhost"),
		WithPort(8080),
	)
	fmt.Println("optional:", cfg.Host, cfg.Port)

	// curried: simple, variadic, and multi-result
	sum := AddCurried()(1)(2)
	fmt.Println("curried add:", sum)

	s := ConcatCurried()("pre-")("a", "b")
	fmt.Println("curried concat:", s)

	q, r := DivModCurried()(10)(3)
	fmt.Println("curried divmod:", q, r)

	// pipeline: compose stages with Result short-circuiting
	f1 := func(x int64) monad.Result[string] { return monad.Ok(fmt.Sprint(x)) }
	f2 := func(s string) monad.Result[float32] { return monad.Ok(float32(len(s))) }
	f3 := func(f float32) monad.Result[bool] { return monad.Ok(f > 0) }

	pipe := AnyPipeComposer(f1, f2, f3)
	ok, err := pipe(42).Unwrap()
	fmt.Println("pipeline ok:", ok, "err:", err)

	// show short-circuit on error
	f2Err := func(string) monad.Result[float32] { return monad.Err[float32](errors.New("boom")) }
	pipeErr := AnyPipeComposer(f1, f2Err, f3)
	ok2, err2 := pipeErr(42).Unwrap()
	fmt.Println("pipeline error:", ok2, "err!=nil:", err2 != nil)

	// match: pattern matching for Address
	addr := Address{
		Street: "123 Main St",
		City:   "Seoul",
		Zip:    "12345",
	}

	fmt.Println("match examples:")

	// Basic pattern matching
	addr.Match().
		When(monad.S("123 Main St"), monad.S("Seoul"), monad.S("12345"), func(a Address) {
			fmt.Println("  exact match: perfect address!")
		}).
		When(monad.W[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
			fmt.Println("  Seoul address")
		}).
		When(monad.W[string](), monad.W[string](), monad.S("12345"), func(a Address) {
			fmt.Println("  zip code 12345")
		}).
		Default(func(a Address) {
			fmt.Println("  other address")
		})

	// Pattern matching with guard conditions
	addr.Match().
		WhenGuard(monad.W[string](), monad.S("Seoul"), monad.W[string](),
			func(a Address) bool { return len(a.Street) > 10 },
			func(a Address) {
				fmt.Println("  Seoul address with long street name")
			}).
		Default(func(a Address) {
			fmt.Println("  other or short street")
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

	fmt.Println("  description:", description)

	// Complex matching example
	addressType := MatchAddressReturn[string](addr).
		WhenGuard(monad.W[string](), monad.S("Seoul"), monad.W[string](),
			func(a Address) bool { return a.Street[0] >= '0' && a.Street[0] <= '9' },
			func(a Address) string { return "Seoul numbered street" }).
		WhenGuard(monad.W[string](), monad.W[string](), monad.W[string](),
			func(a Address) bool { return len(a.Zip) == 5 },
			func(a Address) string { return "Standard 5-digit zip" }).
		DefaultWith(func(a Address) string {
			return fmt.Sprintf("Custom: %s, %s %s", a.Street, a.City, a.Zip)
		})

	fmt.Println("  type:", addressType)
}
