package monad

import (
	"errors"
	"testing"
)

func TestPipelineBasics(t *testing.T) {
	// Test successful pipeline
	pipeline := NewPipeline(Ok(42))
	val, err := pipeline.Unwrap()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test OkP
	okPipeline := OkP(100)
	val, err = okPipeline.Unwrap()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}

	// Test ErrP
	testErr := errors.New("test error")
	errPipeline := ErrP[int](testErr)
	_, err = errPipeline.Unwrap()
	
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestMapP(t *testing.T) {
	// Test mapping successful pipeline
	pipeline := NewPipeline(Ok(42))
	mapped := MapP(pipeline, func(x int) string { return "value: " + string(rune(x+48)) })
	val, err := mapped.Unwrap()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "value: Z" // 42 + 48 = 90 ('Z')
	if val != expected {
		t.Errorf("Expected %s, got %s", expected, val)
	}

	// Test mapping failed pipeline
	testErr := errors.New("test error")
	errPipeline := ErrP[int](testErr)
	mapped2 := MapP(errPipeline, func(x int) string { return "never" })
	_, err = mapped2.Unwrap()
	
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestAndThenP(t *testing.T) {
	// Test chaining successful pipeline
	pipeline := NewPipeline(Ok(42))
	chained := AndThenP(pipeline, func(x int) Result[string] {
		if x > 40 {
			return Ok("big")
		}
		return Err[string](errors.New("too small"))
	})
	val, err := chained.Unwrap()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != "big" {
		t.Errorf("Expected 'big', got %s", val)
	}

	// Test chaining with failure
	small := NewPipeline(Ok(10))
	chained = AndThenP(small, func(x int) Result[string] {
		if x > 40 {
			return Ok("big")
		}
		return Err[string](errors.New("too small"))
	})
	_, err = chained.Unwrap()
	
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "too small" {
		t.Errorf("Expected 'too small', got %s", err.Error())
	}

	// Test chaining error pipeline
	testErr := errors.New("original error")
	errPipeline := ErrP[int](testErr)
	chained2 := AndThenP(errPipeline, func(x int) Result[string] {
		return Ok("never")
	})
	_, err = chained2.Unwrap()
	
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "original error" {
		t.Errorf("Expected 'original error', got %s", err.Error())
	}
}

func TestThenP(t *testing.T) {
	// Test successful then
	pipeline := NewPipeline(Ok(42))
	then := ThenP(pipeline, func(x int) error {
		if x <= 40 {
			return errors.New("too small")
		}
		return nil
	})
	val, err := then.Unwrap()
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test then with failure
	small := NewPipeline(Ok(10))
	then = ThenP(small, func(x int) error {
		if x <= 40 {
			return errors.New("too small")
		}
		return nil
	})
	_, err = then.Unwrap()
	
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "too small" {
		t.Errorf("Expected 'too small', got %s", err.Error())
	}

	// Test then with error pipeline
	testErr := errors.New("original error")
	errPipeline := ErrP[int](testErr)
	then2 := ThenP(errPipeline, func(x int) error {
		return nil
	})
	_, err = then2.Unwrap()
	
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "original error" {
		t.Errorf("Expected 'original error', got %s", err.Error())
	}
}

func TestPipelineChaining(t *testing.T) {
	// Test complex pipeline chaining
	pipeline := NewPipeline(Ok(10))
	
	result := ThenP(
		AndThenP(
			MapP(pipeline, func(x int) int { return x * 2 }),
			func(x int) Result[int] { return Ok(x + 5) },
		),
		func(x int) error {
			// Just validate the result
			if x != 25 {
				return errors.New("unexpected value")
			}
			return nil
		},
	)
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// 10 * 2 + 5 = 25
	expected := 25
	if val != expected {
		t.Errorf("Expected %d, got %d", expected, val)
	}
}