package monad

// Pipeline is a small wrapper around Result[T] that provides chaining helpers.
type Pipeline[T any] struct {
	res Result[T]
}

func NewPipeline[T any](r Result[T]) Pipeline[T] { return Pipeline[T]{res: r} }
func OkP[T any](v T) Pipeline[T]                 { return NewPipeline(Ok(v)) }
func ErrP[T any](e error) Pipeline[T]            { return NewPipeline(Err[T](e)) }

// MapP applies f to the inner value when Ok, producing Pipeline[U].
func MapP[T any, U any](p Pipeline[T], f func(T) U) Pipeline[U] {
	if !p.res.IsOk() {
		return NewPipeline(Err[U](p.res.err))
	}
	v, _ := p.res.Unwrap()
	return NewPipeline(Ok(f(v)))
}

// AndThenP applies f which returns a Result[U] when current is Ok.
func AndThenP[T any, U any](p Pipeline[T], f func(T) Result[U]) Pipeline[U] {
	if !p.res.IsOk() {
		return NewPipeline(Err[U](p.res.err))
	}
	v, _ := p.res.Unwrap()
	return NewPipeline(f(v))
}

// ThenP runs a side-effecting function that may return an error; preserves original value on success.
func ThenP[T any](p Pipeline[T], f func(T) error) Pipeline[T] {
	if !p.res.IsOk() {
		return p
	}
	v, _ := p.res.Unwrap()
	if err := f(v); err != nil {
		return NewPipeline(Err[T](err))
	}
	return p
}

func (p Pipeline[T]) Unwrap() (T, error) { return p.res.Unwrap() }
