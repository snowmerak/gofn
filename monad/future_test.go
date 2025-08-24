package monad

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFutureBasics(t *testing.T) {
	future := NewFuture[int]()
	
	// Should not be done initially
	if future.IsDone() {
		t.Error("New future should not be done")
	}
	
	// Poll should return false
	_, available := future.Poll()
	if available {
		t.Error("New future should not have result available")
	}
	
	// Complete the future
	future.Complete(42)
	
	// Should be done now
	if !future.IsDone() {
		t.Error("Completed future should be done")
	}
	
	// Poll should return the result
	result, available := future.Poll()
	if !available {
		t.Error("Completed future should have result available")
	}
	
	if !result.IsOk() {
		t.Error("Completed future should be Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestFutureCompleteWithError(t *testing.T) {
	future := NewFuture[int]()
	testErr := errors.New("test error")
	
	future.CompleteWithError(testErr)
	
	if !future.IsDone() {
		t.Error("Future completed with error should be done")
	}
	
	result := future.Await()
	if result.IsOk() {
		t.Error("Future completed with error should not be Ok")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestFutureAwait(t *testing.T) {
	future := NewFuture[int]()
	
	// Complete the future after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		future.Complete(100)
	}()
	
	// Await should block until completion
	start := time.Now()
	result := future.Await()
	duration := time.Since(start)
	
	if duration < 5*time.Millisecond {
		t.Error("Await should have waited for completion")
	}
	
	if !result.IsOk() {
		t.Error("Awaited future should be Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}

func TestFutureAwaitWithContext(t *testing.T) {
	future := NewFuture[int]()
	
	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	
	// Complete after timeout
	go func() {
		time.Sleep(50 * time.Millisecond)
		future.Complete(42)
	}()
	
	result := future.AwaitWithContext(ctx)
	if result.IsOk() {
		t.Error("Future should be cancelled by context")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected context error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestFutureAwaitWithTimeout(t *testing.T) {
	future := NewFuture[int]()
	
	// Complete after timeout
	go func() {
		time.Sleep(50 * time.Millisecond)
		future.Complete(42)
	}()
	
	result := future.AwaitWithTimeout(20 * time.Millisecond)
	if result.IsOk() {
		t.Error("Future should timeout")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected timeout error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestCompletedFuture(t *testing.T) {
	future := CompletedFuture(42)
	
	if !future.IsDone() {
		t.Error("CompletedFuture should be done immediately")
	}
	
	result := future.Await()
	if !result.IsOk() {
		t.Error("CompletedFuture should be Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestFailedFuture(t *testing.T) {
	testErr := errors.New("test error")
	future := FailedFuture[int](testErr)
	
	if !future.IsDone() {
		t.Error("FailedFuture should be done immediately")
	}
	
	result := future.Await()
	if result.IsOk() {
		t.Error("FailedFuture should not be Ok")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestRunAsync(t *testing.T) {
	future := RunAsync(func() Result[int] {
		time.Sleep(10 * time.Millisecond)
		return Ok(42)
	})
	
	// Should not be done immediately
	if future.IsDone() {
		// Give it a moment
		time.Sleep(1 * time.Millisecond)
	}
	
	result := future.Await()
	if !result.IsOk() {
		t.Error("RunAsync should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestRunAsyncWithContext(t *testing.T) {
	ctx := context.Background()
	future := RunAsyncWithContext(ctx, func(ctx context.Context) Result[string] {
		return Ok("hello")
	})
	
	result := future.Await()
	if !result.IsOk() {
		t.Error("RunAsyncWithContext should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != "hello" {
		t.Errorf("Expected 'hello', got %s", val)
	}
}

func TestMapFuture(t *testing.T) {
	future := CompletedFuture(42)
	mapped := MapFuture(future, func(x int) string {
		return "value: " + string(rune(x+48))
	})
	
	result := mapped.Await()
	if !result.IsOk() {
		t.Error("Mapped future should be Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "value: Z" // 42 + 48 = 90 ('Z')
	if val != expected {
		t.Errorf("Expected %s, got %s", expected, val)
	}

	// Test mapping failed future
	testErr := errors.New("test error")
	failedFuture := FailedFuture[int](testErr)
	mapped2 := MapFuture(failedFuture, func(x int) string { return "never" })
	
	result2 := mapped2.Await()
	if result2.IsOk() {
		t.Error("Mapped failed future should remain failed")
	}
	
	_, err = result2.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestAndThenFuture(t *testing.T) {
	future := CompletedFuture(42)
	chained := AndThenFuture(future, func(x int) *Future[string] {
		if x > 40 {
			return CompletedFuture("big")
		}
		return FailedFuture[string](errors.New("too small"))
	})
	
	result := chained.Await()
	if !result.IsOk() {
		t.Error("Chained future should be Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != "big" {
		t.Errorf("Expected 'big', got %s", val)
	}
}

func TestSequenceFutures(t *testing.T) {
	futures := []*Future[int]{
		CompletedFuture(10),
		CompletedFuture(20),
		CompletedFuture(30),
	}
	
	sequenced := SequenceFutures(futures)
	result := sequenced.Await()
	
	if !result.IsOk() {
		t.Error("Sequenced futures should be Ok")
	}
	
	vals, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	expected := []int{10, 20, 30}
	if len(vals) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(vals))
	}
	
	for i, exp := range expected {
		if i >= len(vals) || vals[i] != exp {
			t.Errorf("Expected %d at index %d, got %v", exp, i, vals)
		}
	}

	// Test with one failure
	testErr := errors.New("middle error")
	futuresWithError := []*Future[int]{
		CompletedFuture(10),
		FailedFuture[int](testErr),
		CompletedFuture(30),
	}
	
	sequenced2 := SequenceFutures(futuresWithError)
	result2 := sequenced2.Await()
	
	if result2.IsOk() {
		t.Error("Sequenced futures with error should be Err")
	}
	
	_, err = result2.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "middle error" {
		t.Errorf("Expected 'middle error', got %s", err.Error())
	}
}

func TestRaceFutures(t *testing.T) {
	futures := []*Future[int]{
		RunAsync(func() Result[int] {
			time.Sleep(50 * time.Millisecond)
			return Ok(10)
		}),
		RunAsync(func() Result[int] {
			time.Sleep(10 * time.Millisecond)
			return Ok(20)
		}),
		RunAsync(func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(30)
		}),
	}
	
	race := RaceFutures(futures)
	start := time.Now()
	result := race.Await()
	duration := time.Since(start)
	
	if !result.IsOk() {
		t.Error("Race futures should return Ok")
	}
	
	// Should complete quickly (around 10ms for the fastest)
	if duration > 30*time.Millisecond {
		t.Errorf("Race execution took too long: %v", duration)
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should get the result from the fastest future
	if val != 20 {
		t.Errorf("Expected 20 (fastest future), got %d", val)
	}
}

func TestFirstCompleted(t *testing.T) {
	futures := []*Future[int]{
		RunAsync(func() Result[int] {
			time.Sleep(50 * time.Millisecond)
			return Ok(10)
		}),
		RunAsync(func() Result[int] {
			time.Sleep(10 * time.Millisecond)
			return Err[int](errors.New("fast error"))
		}),
		RunAsync(func() Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(30)
		}),
	}
	
	first := FirstCompleted(futures)
	start := time.Now()
	result := first.Await()
	duration := time.Since(start)
	
	// Should complete quickly (around 10ms for the fastest, even if it's an error)
	if duration > 30*time.Millisecond {
		t.Errorf("FirstCompleted took too long: %v", duration)
	}
	
	// Should get the first completed result (which is an error in this case)
	if result.IsOk() {
		t.Error("Expected error from fastest future")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "fast error" {
		t.Errorf("Expected 'fast error', got %s", err.Error())
	}
}