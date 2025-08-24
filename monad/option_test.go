package monad

import (
	"testing"
)

func TestOptionBasics(t *testing.T) {
	// Test Some creation
	some := Some(42)
	if !some.IsSome() {
		t.Error("Some should be Some")
	}
	if some.IsNone() {
		t.Error("Some should not be None")
	}
	if some.IsWildcard() {
		t.Error("Some should not be Wildcard")
	}

	// Test None creation
	none := None[int]()
	if none.IsSome() {
		t.Error("None should not be Some")
	}
	if !none.IsNone() {
		t.Error("None should be None")
	}
	if none.IsWildcard() {
		t.Error("None should not be Wildcard")
	}

	// Test Wildcard creation
	wildcard := Wildcard[int]()
	if wildcard.IsSome() {
		t.Error("Wildcard should not be Some")
	}
	if wildcard.IsNone() {
		t.Error("Wildcard should not be None")
	}
	if !wildcard.IsWildcard() {
		t.Error("Wildcard should be Wildcard")
	}
}

func TestOptionUnwrap(t *testing.T) {
	some := Some(42)
	value := some.Unwrap()
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Test UnwrapOr
	none := None[int]()
	value = none.UnwrapOr(100)
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}

	value = some.UnwrapOr(100)
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}
}

func TestOptionMatch(t *testing.T) {
	some := Some(42)
	
	// Test Some matching with correct value
	if !some.Match(42) {
		t.Error("Some(42) should match 42")
	}
	
	// Test Some not matching with different value
	if some.Match(100) {
		t.Error("Some(42) should not match 100")
	}

	none := None[int]()
	
	// Test None never matching any value
	if none.Match(42) {
		t.Error("None should not match any value")
	}
	if none.Match(0) {
		t.Error("None should not match any value")
	}

	wildcard := Wildcard[int]()
	
	// Test Wildcard matching any value
	if !wildcard.Match(42) {
		t.Error("Wildcard should match any value")
	}
	if !wildcard.Match(100) {
		t.Error("Wildcard should match any value")
	}
	if !wildcard.Match(-1) {
		t.Error("Wildcard should match any value")
	}
}

func TestMapOption(t *testing.T) {
	some := Some(42)
	mapped := MapOption(some, func(x int) string { return "value: " + string(rune(x+48)) })
	
	if !mapped.IsSome() {
		t.Error("Mapped Some should be Some")
	}
	
	result := mapped.Unwrap()
	expected := "value: Z" // 42 + 48 = 90 ('Z')
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	none := None[int]()
	mappedNone := MapOption(none, func(x int) string { return "never" })
	if !mappedNone.IsNone() {
		t.Error("Mapped None should be None")
	}
}

func TestAndThenOption(t *testing.T) {
	some := Some(42)
	result := AndThenOption(some, func(x int) Option[string] {
		if x > 40 {
			return Some("big")
		}
		return None[string]()
	})
	
	if !result.IsSome() {
		t.Error("Result should be Some")
	}
	if result.Unwrap() != "big" {
		t.Errorf("Expected 'big', got %s", result.Unwrap())
	}

	small := Some(10)
	result = AndThenOption(small, func(x int) Option[string] {
		if x > 40 {
			return Some("big")
		}
		return None[string]()
	})
	
	if !result.IsNone() {
		t.Error("Result should be None")
	}
}

func TestOptionAliases(t *testing.T) {
	s := S(42)
	if !s.IsSome() {
		t.Error("S should create Some")
	}

	n := N[int]()
	if !n.IsNone() {
		t.Error("N should create None")
	}

	w := W[int]()
	if !w.IsWildcard() {
		t.Error("W should create Wildcard")
	}
}