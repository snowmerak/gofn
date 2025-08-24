package monad

import (
	"context"
)

// Task represents a computation that can be executed asynchronously
// It's a function that takes a context and returns a Result
type Task[T any] func(context.Context) Result[T]

// NewTask creates a new Task from a function
func NewTask[T any](f func(context.Context) Result[T]) Task[T] {
	return Task[T](f)
}

// NewTaskFromFunc creates a Task from a regular function that might panic or return error
func NewTaskFromFunc[T any](f func(context.Context) (T, error)) Task[T] {
	return func(ctx context.Context) Result[T] {
		val, err := f(ctx)
		if err != nil {
			return Err[T](err)
		}
		return Ok(val)
	}
}

// NewTaskFromValue creates a Task that immediately returns a value
func NewTaskFromValue[T any](value T) Task[T] {
	return func(ctx context.Context) Result[T] {
		return Ok(value)
	}
}

// NewTaskFromError creates a Task that immediately returns an error
func NewTaskFromError[T any](err error) Task[T] {
	return func(ctx context.Context) Result[T] {
		return Err[T](err)
	}
}

// Run executes the Task and returns a Future
func (t Task[T]) Run(ctx context.Context) *Future[T] {
	future := NewFuture[T]()

	go func() {
		result := t(ctx)
		future.complete(result)
	}()

	return future
}

// MapTask transforms the result of a Task
func MapTask[T, U any](task Task[T], f func(T) U) Task[U] {
	return func(ctx context.Context) Result[U] {
		result := task(ctx)
		return Map(result, f)
	}
}

// AndThenTask chains computations
func AndThenTask[T, U any](task Task[T], f func(T) Task[U]) Task[U] {
	return func(ctx context.Context) Result[U] {
		result := task(ctx)
		if !result.IsOk() {
			val, err := result.Unwrap()
			_ = val // unused
			return Err[U](err)
		}
		val, _ := result.Unwrap()
		return f(val)(ctx)
	}
}

// SequenceTasks executes Tasks sequentially and collects results
func SequenceTasks[T any](tasks []Task[T]) Task[[]T] {
	return func(ctx context.Context) Result[[]T] {
		results := make([]T, 0, len(tasks))
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return Err[[]T](ctx.Err())
			default:
			}

			result := task(ctx)
			if !result.IsOk() {
				val, err := result.Unwrap()
				_ = val // unused
				return Err[[]T](err)
			}
			val, _ := result.Unwrap()
			results = append(results, val)
		}
		return Ok(results)
	}
}

// ParallelTasks executes Tasks in parallel and collects results
func ParallelTasks[T any](tasks []Task[T]) Task[[]T] {
	return func(ctx context.Context) Result[[]T] {
		futures := make([]*Future[T], len(tasks))

		// Start all tasks
		for i, task := range tasks {
			futures[i] = task.Run(ctx)
		}

		// Collect results
		results := make([]T, len(tasks))
		for i, future := range futures {
			result := future.AwaitWithContext(ctx)
			if !result.IsOk() {
				val, err := result.Unwrap()
				_ = val // unused
				return Err[[]T](err)
			}
			val, _ := result.Unwrap()
			results[i] = val
		}

		return Ok(results)
	}
}

// RaceTasks executes Tasks in parallel and returns the first successful result
func RaceTasks[T any](tasks []Task[T]) Task[T] {
	return func(ctx context.Context) Result[T] {
		if len(tasks) == 0 {
			return Err[T](context.Canceled)
		}

		futures := make([]*Future[T], len(tasks))
		done := make(chan Result[T], len(tasks))

		// Start all tasks
		for i, task := range tasks {
			futures[i] = task.Run(ctx)
			go func(future *Future[T]) {
				result := future.AwaitWithContext(ctx)
				if result.IsOk() {
					done <- result
				}
			}(futures[i])
		}

		// Wait for first success or context cancellation
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			return Err[T](ctx.Err())
		}
	}
}
