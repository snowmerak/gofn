package monad

import (
	"errors"
	"testing"
)

func TestEitherBasics(t *testing.T) {
	// Test Left creation
	left := Left[string, int]("error")
	if !left.IsLeft() {
		t.Error("Left should be Left")
	}
	if left.IsRight() {
		t.Error("Left should not be Right")
	}

	leftVal, rightVal, isRight := left.Unwrap()
	if isRight {
		t.Error("Left should not be Right")
	}
	if leftVal != "error" {
		t.Errorf("Expected 'error', got %s", leftVal)
	}
	if rightVal != 0 {
		t.Errorf("Expected zero value, got %d", rightVal)
	}

	// Test Right creation
	right := Right[string, int](42)
	if right.IsLeft() {
		t.Error("Right should not be Left")
	}
	if !right.IsRight() {
		t.Error("Right should be Right")
	}

	leftVal, rightVal, isRight = right.Unwrap()
	if !isRight {
		t.Error("Right should be Right")
	}
	if leftVal != "" {
		t.Errorf("Expected empty string, got %s", leftVal)
	}
	if rightVal != 42 {
		t.Errorf("Expected 42, got %d", rightVal)
	}
}

func TestEitherUnwrap(t *testing.T) {
	left := Left[string, int]("error")
	
	// Test UnwrapLeft
	leftVal := left.UnwrapLeft()
	if leftVal != "error" {
		t.Errorf("Expected 'error', got %s", leftVal)
	}

	// Test UnwrapLeftOr
	right := Right[string, int](42)
	leftVal = right.UnwrapLeftOr("default")
	if leftVal != "default" {
		t.Errorf("Expected 'default', got %s", leftVal)
	}

	// Test UnwrapRight
	rightVal := right.UnwrapRight()
	if rightVal != 42 {
		t.Errorf("Expected 42, got %d", rightVal)
	}

	// Test UnwrapRightOr
	rightVal = left.UnwrapRightOr(100)
	if rightVal != 100 {
		t.Errorf("Expected 100, got %d", rightVal)
	}
}

func TestEitherMatch(t *testing.T) {
	left := Left[string, int]("error")
	var result string

	left.Match(
		func(l string) { result = "left: " + l },
		func(r int) { result = "right: " + string(rune(r+48)) },
	)
	if result != "left: error" {
		t.Errorf("Expected 'left: error', got %s", result)
	}

	right := Right[string, int](42)
	right.Match(
		func(l string) { result = "left: " + l },
		func(r int) { result = "right: " + string(rune(r+48)) },
	)
	expected := "right: Z" // 42 + 48 = 90 ('Z')
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEitherMatchWithReturn(t *testing.T) {
	left := Left[string, int]("error")
	result := MatchWithReturn(left,
		func(l string) bool { return len(l) > 3 },
		func(r int) bool { return r > 40 },
	)
	if !result {
		t.Error("Expected true (len('error') > 3)")
	}

	right := Right[string, int](42)
	result = MatchWithReturn(right,
		func(l string) bool { return len(l) > 3 },
		func(r int) bool { return r > 40 },
	)
	if !result {
		t.Error("Expected true (42 > 40)")
	}

	small := Right[string, int](10)
	result = MatchWithReturn(small,
		func(l string) bool { return len(l) > 3 },
		func(r int) bool { return r > 40 },
	)
	if result {
		t.Error("Expected false (10 <= 40)")
	}
}

func TestMapLeft(t *testing.T) {
	left := Left[string, int]("error")
	mapped := MapLeft(left, func(s string) int { return len(s) })
	
	if !mapped.IsLeft() {
		t.Error("Mapped Left should be Left")
	}
	
	leftVal := mapped.UnwrapLeft()
	if leftVal != 5 { // len("error") = 5
		t.Errorf("Expected 5, got %d", leftVal)
	}

	right := Right[string, int](42)
	mapped = MapLeft(right, func(s string) int { return len(s) })
	
	if !mapped.IsRight() {
		t.Error("Mapped Right should stay Right")
	}
	
	rightVal := mapped.UnwrapRight()
	if rightVal != 42 {
		t.Errorf("Expected 42, got %d", rightVal)
	}
}

func TestMapRight(t *testing.T) {
	right := Right[string, int](42)
	mapped := MapRight(right, func(i int) string { return "number: " + string(rune(i+48)) })
	
	if !mapped.IsRight() {
		t.Error("Mapped Right should be Right")
	}
	
	rightVal := mapped.UnwrapRight()
	expected := "number: Z" // 42 + 48 = 90 ('Z')
	if rightVal != expected {
		t.Errorf("Expected %s, got %s", expected, rightVal)
	}

	left := Left[string, int]("error")
	mapped = MapRight(left, func(i int) string { return "never" })
	
	if !mapped.IsLeft() {
		t.Error("Mapped Left should stay Left")
	}
	
	leftVal := mapped.UnwrapLeft()
	if leftVal != "error" {
		t.Errorf("Expected 'error', got %s", leftVal)
	}
}

func TestBiMap(t *testing.T) {
	left := Left[string, int]("error")
	mapped := BiMap(left,
		func(s string) bool { return len(s) > 3 },
		func(i int) string { return "never" },
	)
	
	if !mapped.IsLeft() {
		t.Error("BiMapped Left should be Left")
	}
	
	leftVal := mapped.UnwrapLeft()
	if !leftVal {
		t.Error("Expected true (len('error') > 3)")
	}

	right := Right[string, int](42)
	mapped2 := BiMap(right,
		func(s string) bool { return false },
		func(i int) string { return "number: " + string(rune(i+48)) },
	)
	
	if !mapped2.IsRight() {
		t.Error("BiMapped Right should be Right")
	}
	
	rightVal := mapped2.UnwrapRight()
	expected := "number: Z" // 42 + 48 = 90 ('Z')
	if rightVal != expected {
		t.Errorf("Expected %s, got %s", expected, rightVal)
	}
}

func TestEitherSwap(t *testing.T) {
	left := Left[string, int]("error")
	swapped := left.Swap()
	
	if !swapped.IsRight() {
		t.Error("Swapped Left should be Right")
	}
	
	rightVal := swapped.UnwrapRight()
	if rightVal != "error" {
		t.Errorf("Expected 'error', got %s", rightVal)
	}

	right := Right[string, int](42)
	swapped2 := right.Swap()
	
	if !swapped2.IsLeft() {
		t.Error("Swapped Right should be Left")
	}
	
	leftVal := swapped2.UnwrapLeft()
	if leftVal != 42 {
		t.Errorf("Expected 42, got %d", leftVal)
	}
}

func TestEitherResultConversion(t *testing.T) {
	// Test ToResult
	rightEither := Right[error, int](42)
	result := ToResult(rightEither)
	
	if !result.IsOk() {
		t.Error("Right Either should convert to Ok Result")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	testErr := errors.New("test error")
	leftEither := Left[error, int](testErr)
	result = ToResult(leftEither)
	
	if result.IsOk() {
		t.Error("Left Either should convert to Err Result")
	}
	
	_, err = result.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}

	// Test FromResult
	okResult := Ok(42)
	either := FromResult(okResult)
	
	if !either.IsRight() {
		t.Error("Ok Result should convert to Right Either")
	}
	
	rightVal := either.UnwrapRight()
	if rightVal != 42 {
		t.Errorf("Expected 42, got %d", rightVal)
	}

	errResult := Err[int](testErr)
	either = FromResult(errResult)
	
	if !either.IsLeft() {
		t.Error("Err Result should convert to Left Either")
	}
	
	leftVal := either.UnwrapLeft()
	if leftVal.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", leftVal.Error())
	}
}

func TestEitherAliases(t *testing.T) {
	left := L[string, int]("error")
	if !left.IsLeft() {
		t.Error("L should create Left")
	}

	right := R[string, int](42)
	if !right.IsRight() {
		t.Error("R should create Right")
	}
}