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

//gofn:reactive
type Counter struct {
	Value int
	Name  string
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

	// NEW: Error handling with custom handlers
	fmt.Println("pipeline with error handlers:")
	
	// Error handler that provides fallback value
	fallbackHandler := AnyPipeWithFallback(false) // fallback to false
	pipeWithFallback := AnyPipeComposerWithErrorHandler(f1, f2Err, f3, fallbackHandler)
	okFallback, errFallback := pipeWithFallback(42).Unwrap()
	fmt.Println("  with fallback:", okFallback, "err:", errFallback)

	// Error handler that logs and propagates
	logHandler := AnyPipeWithLogging(func(stage int, err error) {
		fmt.Printf("  [LOG] Stage %d failed: %v\n", stage, err)
	})
	pipeWithLog := AnyPipeComposerWithErrorHandler(f1, f2Err, f3, logHandler)
	okLog, errLog := pipeWithLog(42).Unwrap()
	fmt.Println("  with logging:", okLog, "err!=nil:", errLog != nil)

	// Custom error handler for specific stages
	customHandler := func(stageIndex int, err error) monad.Result[bool] {
		switch stageIndex {
		case 2: // f2 stage - provide recovery value
			fmt.Printf("  [CUSTOM] Recovering from stage %d error: %v\n", stageIndex, err)
			return monad.Ok(true) // recover with true
		default:
			fmt.Printf("  [CUSTOM] Cannot recover from stage %d error: %v\n", stageIndex, err)
			return monad.Err[bool](err)
		}
	}
	pipeWithCustom := AnyPipeComposerWithErrorHandler(f1, f2Err, f3, customHandler)
	okCustom, errCustom := pipeWithCustom(42).Unwrap()
	fmt.Println("  with custom handler:", okCustom, "err:", errCustom)

	// match: pattern matching for Address
	addr := Address{
		Street: "123 Main St",
		City:   "Seoul",
		Zip:    "12345",
	}

	fmt.Println("match examples:")

	// Basic pattern matching - S for Some, W for Wildcard, N for None
	addr.Match().
		When(monad.S("123 Main St"), monad.S("Seoul"), monad.S("12345"), func(a Address) {
			fmt.Println("  exact match: perfect address!")
		}).
		When(monad.W[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
			fmt.Println("  Seoul address (any street, any zip)")
		}).
		When(monad.W[string](), monad.W[string](), monad.S("12345"), func(a Address) {
			fmt.Println("  zip code 12345 (any street, any city)")
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

	// Complex matching example with wildcards
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

	// reactive: reactive programming with subscriptions
	fmt.Println("reactive examples:")

	counter := NewReactiveCounter(Counter{Value: 0, Name: "MainCounter"})

	// Subscribe to changes
	subID1 := counter.Subscribe(func(old, new Counter) {
		fmt.Printf("  [Subscriber 1] Counter changed from %d to %d", old.Value, new.Value)
	})

	subID2 := counter.Subscribe(func(old, new Counter) {
		if new.Value > 5 {
			fmt.Printf("  [Subscriber 2] High value alert: %d", new.Value)
		}
	})

	// Test value changes
	fmt.Println("  Setting counter values...")
	counter.SetValue(3)
	counter.SetValue(7)
	counter.Update(func(c Counter) Counter {
		c.Value += 10
		return c
	})

	// Test field-specific updates
	counter.SetName("UpdatedCounter")

	// Unsubscribe one listener
	counter.Unsubscribe(subID1)
	fmt.Println("  (Unsubscribed subscriber 1)")

	counter.SetValue(20) // Only subscriber 2 should react

	// Later unsubscribe subscriber 2 as well
	counter.Unsubscribe(subID2)
	fmt.Println("  (Unsubscribed subscriber 2)")

	counter.SetValue(30) // No subscribers should react

	// Demonstrate reactive mapping
	fmt.Println("  Creating mapped reactive...")
	stringReactive := MapCounter[string](counter, func(c Counter) string {
		return fmt.Sprintf("%s: %d", c.Name, c.Value)
	})

	stringReactive.Subscribe(func(old, new string) {
		fmt.Printf("  [String Reactive] %s", new)
	})

	counter.SetValue(25) // Should trigger both counter and string reactive

	// Demonstrate the difference between None and Wildcard
	fmt.Println("Demonstrating None vs Wildcard:")

	// Example with potentially missing data
	partialAddr := Address{
		Street: "", // Empty string (not missing, but empty)
		City:   "Seoul",
		Zip:    "12345",
	}

	partialAddr.Match().
		When(monad.N[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
			fmt.Println("  None pattern - this won't match empty string")
		}).
		When(monad.S(""), monad.S("Seoul"), monad.W[string](), func(a Address) {
			fmt.Println("  Empty street in Seoul (explicit empty string match)")
		}).
		When(monad.W[string](), monad.S("Seoul"), monad.W[string](), func(a Address) {
			fmt.Println("  Any Seoul address (wildcard catches everything)")
		}).
		Default(func(a Address) {
			fmt.Println("  No match")
		})
}
