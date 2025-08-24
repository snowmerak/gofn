package monad

// Option represents an optional value with pattern matching support
// Every Option is either Some (contains a value), None (explicitly empty), or Wildcard (matches anything)
type Option[T any] struct {
	value     *T
	isWildcard bool
}

// Some wraps a value in an Option
func Some[T any](value T) Option[T] {
	return Option[T]{value: &value, isWildcard: false}
}

// None returns an explicitly empty Option
func None[T any]() Option[T] {
	return Option[T]{value: nil, isWildcard: false}
}

// Wildcard returns a pattern that matches any value
func Wildcard[T any]() Option[T] {
	return Option[T]{value: nil, isWildcard: true}
}

// IsSome returns true if the option contains a value
func (o Option[T]) IsSome() bool {
	return o.value != nil && !o.isWildcard
}

// IsNone returns true if the option is explicitly empty (not wildcard)
func (o Option[T]) IsNone() bool {
	return o.value == nil && !o.isWildcard
}

// IsWildcard returns true if the option is a wildcard pattern
func (o Option[T]) IsWildcard() bool {
	return o.isWildcard
}

// Unwrap returns the contained value or panics if None or Wildcard
func (o Option[T]) Unwrap() T {
	if o.value == nil {
		if o.isWildcard {
			panic("called Unwrap on Wildcard value")
		}
		panic("called Unwrap on None value")
	}
	return *o.value
}

// UnwrapOr returns the contained value or a default
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.value == nil {
		return defaultValue
	}
	return *o.value
}

// Match checks if this Option pattern matches the given value
// - Some(x) matches only if the value equals x
// - None() never matches any actual value (used for explicit absence)
// - Wildcard() matches any value
func (o Option[T]) Match(value T) bool {
	if o.isWildcard {
		return true // Wildcard matches anything
	}
	if o.value == nil {
		return false // None doesn't match any actual value
	}
	return equals(*o.value, value)
}

// equals compares two values of the same type
func equals[T any](a, b T) bool {
	// This is a simplified comparison - in a real implementation,
	// you might want to use reflection or require Comparable interface
	return any(a) == any(b)
}

// Map applies a function to the contained value (if any)
func MapOption[T any, U any](o Option[T], f func(T) U) Option[U] {
	if o.isWildcard {
		return Wildcard[U]()
	}
	if o.value == nil {
		return None[U]()
	}
	return Some(f(*o.value))
}

// AndThen applies a function that returns an Option to the contained value
func AndThenOption[T any, U any](o Option[T], f func(T) Option[U]) Option[U] {
	if o.isWildcard {
		return Wildcard[U]()
	}
	if o.value == nil {
		return None[U]()
	}
	return f(*o.value)
}

// Helper functions for pattern matching
// S for Some - matches specific value
func S[T any](value T) Option[T] { return Some(value) }

// N for None - represents explicit absence (doesn't match actual values)
func N[T any]() Option[T] { return None[T]() }

// W for Wildcard - matches any value (pattern matching wildcard)
func W[T any]() Option[T] { return Wildcard[T]() }