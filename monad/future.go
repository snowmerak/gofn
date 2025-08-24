package monad

import (
	"context"
	"sync"
	"time"
)

// Future represents a computation that will complete in the future
// Uses sync.Cond for efficient waiting instead of channels
type Future[T any] struct {
	mu     *sync.Mutex
	cond   *sync.Cond
	done   bool
	result Result[T]
}

// NewFuture creates a new Future
func NewFuture[T any]() *Future[T] {
	mu := &sync.Mutex{}
	return &Future[T]{
		mu:   mu,
		cond: sync.NewCond(mu),
		done: false,
	}
}

// complete marks the Future as done with the given result
func (f *Future[T]) complete(result Result[T]) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	
	if f.done {
		return // already completed
	}
	
	f.result = result
	f.done = true
	f.cond.Broadcast() // wake up all waiting goroutines
}

// Complete manually completes the Future with a value
func (f *Future[T]) Complete(value T) {
	f.complete(Ok(value))
}

// CompleteWithError manually completes the Future with an error
func (f *Future[T]) CompleteWithError(err error) {
	f.complete(Err[T](err))
}

// IsDone returns true if the Future has completed
func (f *Future[T]) IsDone() bool {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	return f.done
}

// Poll returns the result if available, without blocking
func (f *Future[T]) Poll() (Result[T], bool) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	
	if f.done {
		return f.result, true
	}
	
	var zero Result[T]
	return zero, false
}

// Await waits for the Future to complete and returns the result
func (f *Future[T]) Await() Result[T] {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	
	for !f.done {
		f.cond.Wait()
	}
	
	return f.result
}

// AwaitWithContext waits for the Future to complete or context to be cancelled
func (f *Future[T]) AwaitWithContext(ctx context.Context) Result[T] {
	done := make(chan Result[T], 1)
	
	go func() {
		done <- f.Await()
	}()
	
	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		return Err[T](ctx.Err())
	}
}

// AwaitWithTimeout waits for the Future to complete or timeout
func (f *Future[T]) AwaitWithTimeout(timeout time.Duration) Result[T] {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.AwaitWithContext(ctx)
}





// Convenience functions for creating completed Futures

// CompletedFuture creates a Future that's already completed with a value
func CompletedFuture[T any](value T) *Future[T] {
	future := NewFuture[T]()
	future.Complete(value)
	return future
}

// FailedFuture creates a Future that's already completed with an error
func FailedFuture[T any](err error) *Future[T] {
	future := NewFuture[T]()
	future.CompleteWithError(err)
	return future
}

// RunAsync executes a function asynchronously and returns a Future
func RunAsync[T any](f func() Result[T]) *Future[T] {
	future := NewFuture[T]()
	
	go func() {
		result := f()
		future.complete(result)
	}()
	
	return future
}

// RunAsyncWithContext executes a function asynchronously with context
func RunAsyncWithContext[T any](ctx context.Context, f func(context.Context) Result[T]) *Future[T] {
	future := NewFuture[T]()
	
	go func() {
		result := f(ctx)
		future.complete(result)
	}()
	
	return future
}

// MapFuture transforms the result of a Future
func MapFuture[T, U any](future *Future[T], fn func(T) U) *Future[U] {
	newFuture := NewFuture[U]()
	
	go func() {
		result := future.Await()
		mappedResult := Map(result, fn)
		newFuture.complete(mappedResult)
	}()
	
	return newFuture
}

// AndThenFuture chains computations on a Future
func AndThenFuture[T, U any](future *Future[T], fn func(T) *Future[U]) *Future[U] {
	newFuture := NewFuture[U]()
	
	go func() {
		result := future.Await()
		if !result.IsOk() {
			val, err := result.Unwrap()
			_ = val // unused
			newFuture.CompleteWithError(err)
			return
		}
		
		val, _ := result.Unwrap()
		nextFuture := fn(val)
		nextResult := nextFuture.Await()
		newFuture.complete(nextResult)
	}()
	
	return newFuture
}

// Combine multiple Futures

// SequenceFutures waits for all Futures to complete and collects results
func SequenceFutures[T any](futures []*Future[T]) *Future[[]T] {
	resultFuture := NewFuture[[]T]()
	
	go func() {
		results := make([]T, len(futures))
		for i, future := range futures {
			result := future.Await()
			if !result.IsOk() {
				val, err := result.Unwrap()
				_ = val // unused
				resultFuture.CompleteWithError(err)
				return
			}
			val, _ := result.Unwrap()
			results[i] = val
		}
		resultFuture.Complete(results)
	}()
	
	return resultFuture
}

// RaceFutures returns the first Future to complete successfully
func RaceFutures[T any](futures []*Future[T]) *Future[T] {
	resultFuture := NewFuture[T]()
	
	if len(futures) == 0 {
		resultFuture.CompleteWithError(context.Canceled)
		return resultFuture
	}
	
	for _, future := range futures {
		go func(f *Future[T]) {
			result := f.Await()
			if result.IsOk() {
				val, _ := result.Unwrap()
				resultFuture.Complete(val)
			}
		}(future)
	}
	
	return resultFuture
}

// AllOrNone waits for all Futures and returns results only if all succeed
func AllOrNone[T any](futures []*Future[T]) *Future[[]T] {
	return SequenceFutures(futures)
}

// FirstCompleted returns the first Future to complete (success or failure)
func FirstCompleted[T any](futures []*Future[T]) *Future[T] {
	resultFuture := NewFuture[T]()
	
	if len(futures) == 0 {
		resultFuture.CompleteWithError(context.Canceled)
		return resultFuture
	}
	
	for _, future := range futures {
		go func(f *Future[T]) {
			result := f.Await()
			resultFuture.complete(result)
		}(future)
	}
	
	return resultFuture
}