package monad

import (
	"sync"
	"sync/atomic"
)

// Reactive wraps a value of type T and provides reactive capabilities
type Reactive[T any] struct {
	value       T
	subscribers map[int]func(old T, new T)
	nextID      int64
	mutex       sync.RWMutex
}

// NewReactive creates a new reactive wrapper around the given value
func NewReactive[T any](initial T) *Reactive[T] {
	return &Reactive[T]{
		value:       initial,
		subscribers: make(map[int]func(old T, new T)),
		nextID:      0,
	}
}

// Get returns the current value (thread-safe read)
func (r *Reactive[T]) Get() T {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.value
}

// Set updates the value and notifies all subscribers
func (r *Reactive[T]) Set(newValue T) {
	r.mutex.Lock()
	oldValue := r.value
	r.value = newValue
	
	// Copy subscribers to avoid holding lock during notifications
	subscribers := make(map[int]func(old T, new T))
	for id, callback := range r.subscribers {
		subscribers[id] = callback
	}
	r.mutex.Unlock()
	
	// Notify subscribers outside of lock to prevent deadlocks
	for _, callback := range subscribers {
		go callback(oldValue, newValue)
	}
}

// Update applies a function to the current value and sets the result
func (r *Reactive[T]) Update(fn func(T) T) {
	r.mutex.Lock()
	oldValue := r.value
	newValue := fn(r.value)
	r.value = newValue
	
	// Copy subscribers to avoid holding lock during notifications
	subscribers := make(map[int]func(old T, new T))
	for id, callback := range r.subscribers {
		subscribers[id] = callback
	}
	r.mutex.Unlock()
	
	// Notify subscribers outside of lock to prevent deadlocks
	for _, callback := range subscribers {
		go callback(oldValue, newValue)
	}
}

// Subscribe adds a callback that will be called when the value changes
// Returns a subscription ID that can be used to unsubscribe
func (r *Reactive[T]) Subscribe(callback func(old T, new T)) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	id := int(atomic.AddInt64(&r.nextID, 1))
	r.subscribers[id] = callback
	return id
}

// Unsubscribe removes a subscription by ID
func (r *Reactive[T]) Unsubscribe(id int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.subscribers, id)
}

// MapReactive creates a new reactive that transforms this reactive's value
func MapReactive[T any, U any](source *Reactive[T], transform func(T) U) *Reactive[U] {
	result := NewReactive(transform(source.Get()))
	
	source.Subscribe(func(old, new T) {
		result.Set(transform(new))
	})
	
	return result
}

// FilterReactive creates a new reactive that only updates when the predicate is true
func FilterReactive[T any](source *Reactive[T], predicate func(T) bool) *Reactive[T] {
	current := source.Get()
	result := NewReactive(current)
	
	source.Subscribe(func(old, new T) {
		if predicate(new) {
			result.Set(new)
		}
	})
	
	return result
}

// CombineReactives combines two reactives into one
func CombineReactives[T any, U any, V any](
	a *Reactive[T], 
	b *Reactive[U], 
	combiner func(T, U) V,
) *Reactive[V] {
	result := NewReactive(combiner(a.Get(), b.Get()))
	
	a.Subscribe(func(_, newA T) {
		result.Set(combiner(newA, b.Get()))
	})
	
	b.Subscribe(func(_, newB U) {
		result.Set(combiner(a.Get(), newB))
	})
	
	return result
}