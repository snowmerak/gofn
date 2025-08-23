package monad

// Generic Result type with basic combinators
type Result[T any] struct {
	val T
	err error
}

func Ok[T any](v T) Result[T]      { return Result[T]{val: v, err: nil} }
func Err[T any](e error) Result[T] { var z T; return Result[T]{val: z, err: e} }

func (r Result[T]) IsOk() bool         { return r.err == nil }
func (r Result[T]) Unwrap() (T, error) { return r.val, r.err }

func Map[T any, U any](r Result[T], f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.val))
}

func AndThen[T any, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return f(r.val)
}
