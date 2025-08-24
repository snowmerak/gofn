package monad

// Option represents an optional value: every Option is either Some and contains a value, or None
type Option[T any] struct {
	value *T
}

// Some wraps a value in an Option
func Some[T any](value T) Option[T] {
	return Option[T]{value: &value}
}

// None returns an empty Option
func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

// IsSome returns true if the option contains a value
func (o Option[T]) IsSome() bool {
	return o.value != nil
}

// IsNone returns true if the option is empty
func (o Option[T]) IsNone() bool {
	return o.value == nil
}

// Unwrap returns the contained value or panics if None
func (o Option[T]) Unwrap() T {
	if o.value == nil {
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

// Map applies a function to the contained value (if any)
func MapOption[T any, U any](o Option[T], f func(T) U) Option[U] {
	if o.value == nil {
		return None[U]()
	}
	return Some(f(*o.value))
}

// AndThen applies a function that returns an Option to the contained value
func AndThenOption[T any, U any](o Option[T], f func(T) Option[U]) Option[U] {
	if o.value == nil {
		return None[U]()
	}
	return f(*o.value)
}

// Helper functions - S for Some, N for None
func S[T any](value T) Option[T] { return Some(value) }
func N[T any]() Option[T]        { return None[T]() }