package monad

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTaskBasics(t *testing.T) {
	// Test NewTask
	task := NewTask(func(ctx context.Context) Result[int] {
		return Ok(42)
	})
	
	result := task(context.Background())
	if !result.IsOk() {
		t.Error("Task should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test NewTaskFromFunc
	taskFromFunc := NewTaskFromFunc(func(ctx context.Context) (string, error) {
		return "hello", nil
	})
	
	result2 := taskFromFunc(context.Background())
	if !result2.IsOk() {
		t.Error("TaskFromFunc should return Ok")
	}
	
	val2, err := result2.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val2 != "hello" {
		t.Errorf("Expected 'hello', got %s", val2)
	}

	// Test NewTaskFromFunc with error
	taskWithError := NewTaskFromFunc(func(ctx context.Context) (string, error) {
		return "", errors.New("test error")
	})
	
	result3 := taskWithError(context.Background())
	if result3.IsOk() {
		t.Error("Task with error should return Err")
	}
	
	_, err = result3.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestTaskFromValue(t *testing.T) {
	task := NewTaskFromValue(100)
	result := task(context.Background())
	
	if !result.IsOk() {
		t.Error("TaskFromValue should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 100 {
		t.Errorf("Expected 100, got %d", val)
	}
}

func TestTaskFromError(t *testing.T) {
	testErr := errors.New("test error")
	task := NewTaskFromError[int](testErr)
	result := task(context.Background())
	
	if result.IsOk() {
		t.Error("TaskFromError should return Err")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestTaskRun(t *testing.T) {
	task := NewTaskFromValue(42)
	future := task.Run(context.Background())
	
	if !future.IsDone() {
		// Wait a bit for the task to complete
		time.Sleep(10 * time.Millisecond)
	}
	
	result := future.Await()
	if !result.IsOk() {
		t.Error("Task run should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestTaskWithContext(t *testing.T) {
	// Test context cancellation
	task := NewTask(func(ctx context.Context) Result[int] {
		select {
		case <-time.After(100 * time.Millisecond):
			return Ok(42)
		case <-ctx.Done():
			return Err[int](ctx.Err())
		}
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	result := task(ctx)
	if result.IsOk() {
		t.Error("Task should be cancelled")
	}
	
	_, err := result.Unwrap()
	if err == nil {
		t.Error("Expected context error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestMapTask(t *testing.T) {
	task := NewTaskFromValue(42)
	mapped := MapTask(task, func(x int) string {
		return "value: " + string(rune(x+48))
	})
	
	result := mapped(context.Background())
	if !result.IsOk() {
		t.Error("Mapped task should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "value: Z" // 42 + 48 = 90 ('Z')
	if val != expected {
		t.Errorf("Expected %s, got %s", expected, val)
	}

	// Test mapping error task
	testErr := errors.New("test error")
	errTask := NewTaskFromError[int](testErr)
	mapped2 := MapTask(errTask, func(x int) string { return "never" })
	
	result2 := mapped2(context.Background())
	if result2.IsOk() {
		t.Error("Mapped error task should remain error")
	}
	
	_, err = result2.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %s", err.Error())
	}
}

func TestAndThenTask(t *testing.T) {
	task := NewTaskFromValue(42)
	chained := AndThenTask(task, func(x int) Task[string] {
		if x > 40 {
			return NewTaskFromValue("big")
		}
		return NewTaskFromError[string](errors.New("too small"))
	})
	
	result := chained(context.Background())
	if !result.IsOk() {
		t.Error("Chained task should return Ok")
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != "big" {
		t.Errorf("Expected 'big', got %s", val)
	}

	// Test chaining with failure
	small := NewTaskFromValue(10)
	chained2 := AndThenTask(small, func(x int) Task[string] {
		if x > 40 {
			return NewTaskFromValue("big")
		}
		return NewTaskFromError[string](errors.New("too small"))
	})
	
	result2 := chained2(context.Background())
	if result2.IsOk() {
		t.Error("Chained task should return error")
	}
	
	_, err = result2.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "too small" {
		t.Errorf("Expected 'too small', got %s", err.Error())
	}
}

func TestSequenceTasks(t *testing.T) {
	tasks := []Task[int]{
		NewTaskFromValue(10),
		NewTaskFromValue(20),
		NewTaskFromValue(30),
	}
	
	sequenced := SequenceTasks(tasks)
	result := sequenced(context.Background())
	
	if !result.IsOk() {
		t.Error("Sequenced tasks should return Ok")
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
	tasksWithError := []Task[int]{
		NewTaskFromValue(10),
		NewTaskFromError[int](errors.New("middle error")),
		NewTaskFromValue(30),
	}
	
	sequenced2 := SequenceTasks(tasksWithError)
	result2 := sequenced2(context.Background())
	
	if result2.IsOk() {
		t.Error("Sequenced tasks with error should return Err")
	}
	
	_, err = result2.Unwrap()
	if err == nil {
		t.Error("Expected error")
	}
	if err.Error() != "middle error" {
		t.Errorf("Expected 'middle error', got %s", err.Error())
	}
}

func TestParallelTasks(t *testing.T) {
	tasks := []Task[int]{
		NewTask(func(ctx context.Context) Result[int] {
			time.Sleep(10 * time.Millisecond)
			return Ok(10)
		}),
		NewTask(func(ctx context.Context) Result[int] {
			time.Sleep(20 * time.Millisecond)
			return Ok(20)
		}),
		NewTask(func(ctx context.Context) Result[int] {
			time.Sleep(5 * time.Millisecond)
			return Ok(30)
		}),
	}
	
	parallel := ParallelTasks(tasks)
	start := time.Now()
	result := parallel(context.Background())
	duration := time.Since(start)
	
	if !result.IsOk() {
		t.Error("Parallel tasks should return Ok")
	}
	
	// Should complete in roughly the time of the longest task (20ms), not the sum (35ms)
	if duration > 50*time.Millisecond {
		t.Errorf("Parallel execution took too long: %v", duration)
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
}

func TestRaceTasks(t *testing.T) {
	tasks := []Task[int]{
		NewTask(func(ctx context.Context) Result[int] {
			time.Sleep(50 * time.Millisecond)
			return Ok(10)
		}),
		NewTask(func(ctx context.Context) Result[int] {
			time.Sleep(10 * time.Millisecond)
			return Ok(20)
		}),
		NewTask(func(ctx context.Context) Result[int] {
			time.Sleep(100 * time.Millisecond)
			return Ok(30)
		}),
	}
	
	race := RaceTasks(tasks)
	start := time.Now()
	result := race(context.Background())
	duration := time.Since(start)
	
	if !result.IsOk() {
		t.Error("Race tasks should return Ok")
	}
	
	// Should complete quickly (around 10ms for the fastest task)
	if duration > 30*time.Millisecond {
		t.Errorf("Race execution took too long: %v", duration)
	}
	
	val, err := result.Unwrap()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should get the result from the fastest task
	if val != 20 {
		t.Errorf("Expected 20 (fastest task), got %d", val)
	}

	// Test race with empty slice
	emptyRace := RaceTasks([]Task[int]{})
	result2 := emptyRace(context.Background())
	
	if result2.IsOk() {
		t.Error("Empty race should return error")
	}
}