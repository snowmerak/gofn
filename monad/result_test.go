package monad

import (
	"errors"
	"testing"
)

func TestResultBasics(t *testing.T) {
	// Test Ok creation
	ok := Ok(42)
	if !ok.IsOk() {
		t.Error("Ok should be Ok")
	}
	
	value, err := ok.Unwrap()
	if err != nil {
		t.Errorf("Ok should not have error, got %v", err)
	}
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Test Err creation
	testErr := errors.New("test error")
	errResult := Err[int](testErr)
	if errResult.IsOk() {
		t.Error("Err should not be Ok")
	}
	
	value, err = errResult.Unwrap()
	if err == nil {
		t.Error("Err should have error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
	if value != 0 {
		t.Errorf("Expected zero value, got %d", value)
	}
}

func TestMapResult(t *testing.T) {
	// Test mapping Ok
	ok := Ok(42)
	mapped := Map(ok, func(x int) string { return "number: " + string(rune(x+48)) })
	
	if !mapped.IsOk() {
		t.Error("Mapped Ok should be Ok")
	}
	
	value, err := mapped.Unwrap()
	if err != nil {
		t.Errorf("Mapped Ok should not have error, got %v", err)
	}
	expected := "number: Z" // 42 + 48 = 90 ('Z')
	if value != expected {
		t.Errorf("Expected %s, got %s", expected, value)
	}

	// Test mapping Err
	testErr := errors.New("test error")
	errResult := Err[int](testErr)
	mappedErr := Map(errResult, func(x int) string { return "never" })
	
	if mappedErr.IsOk() {
		t.Error("Mapped Err should not be Ok")
	}
	
	_, err = mappedErr.Unwrap()
	if err == nil {
		t.Error("Mapped Err should have error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestAndThenResult(t *testing.T) {
	// Test chaining Ok
	ok := Ok(42)
	result := AndThen(ok, func(x int) Result[string] {
		if x > 40 {
			return Ok("big")
		}
		return Err[string](errors.New("too small"))
	})
	
	if !result.IsOk() {
		t.Error("Result should be Ok")
	}
	
	value, err := result.Unwrap()
	if err != nil {
		t.Errorf("Result should not have error, got %v", err)
	}
	if value != "big" {
		t.Errorf("Expected 'big', got %s", value)
	}

	// Test chaining with failure
	small := Ok(10)
	result = AndThen(small, func(x int) Result[string] {
		if x > 40 {
			return Ok("big")
		}
		return Err[string](errors.New("too small"))
	})
	
	if result.IsOk() {
		t.Error("Result should not be Ok")
	}
	
	_, err = result.Unwrap()
	if err == nil {
		t.Error("Result should have error")
	}
	if err.Error() != "too small" {
		t.Errorf("Expected 'too small', got %s", err.Error())
	}

	// Test chaining Err
	errResult := Err[int](errors.New("original error"))
	result = AndThen(errResult, func(x int) Result[string] {
		return Ok("never")
	})
	
	if result.IsOk() {
		t.Error("Result should not be Ok")
	}
	
	_, err = result.Unwrap()
	if err == nil {
		t.Error("Result should have error")
	}
	if err.Error() != "original error" {
		t.Errorf("Expected 'original error', got %s", err.Error())
	}
}