package monad

// Either represents a value of one of two possible types.
// By convention, Left is used for "failure" cases and Right for "success" cases,
// but both are considered valid values (unlike Result where error is exceptional).
type Either[L, R any] struct {
	left    L
	right   R
	isRight bool
}

// Left creates an Either with a Left value
func Left[L, R any](left L) Either[L, R] {
	var zeroRight R
	return Either[L, R]{left: left, right: zeroRight, isRight: false}
}

// Right creates an Either with a Right value
func Right[L, R any](right R) Either[L, R] {
	var zeroLeft L
	return Either[L, R]{left: zeroLeft, right: right, isRight: true}
}

// IsLeft returns true if this Either contains a Left value
func (e Either[L, R]) IsLeft() bool {
	return !e.isRight
}

// IsRight returns true if this Either contains a Right value
func (e Either[L, R]) IsRight() bool {
	return e.isRight
}

// Unwrap returns the Left value, Right value, and a boolean indicating if it's Right
func (e Either[L, R]) Unwrap() (L, R, bool) {
	return e.left, e.right, e.isRight
}

// UnwrapLeft returns the Left value if present, panics otherwise
func (e Either[L, R]) UnwrapLeft() L {
	if e.isRight {
		panic("called UnwrapLeft on Right value")
	}
	return e.left
}

// UnwrapRight returns the Right value if present, panics otherwise
func (e Either[L, R]) UnwrapRight() R {
	if !e.isRight {
		panic("called UnwrapRight on Left value")
	}
	return e.right
}

// UnwrapLeftOr returns the Left value if present, otherwise returns the default
func (e Either[L, R]) UnwrapLeftOr(defaultLeft L) L {
	if e.isRight {
		return defaultLeft
	}
	return e.left
}

// UnwrapRightOr returns the Right value if present, otherwise returns the default
func (e Either[L, R]) UnwrapRightOr(defaultRight R) R {
	if !e.isRight {
		return defaultRight
	}
	return e.right
}

// Match performs pattern matching on Either
func (e Either[L, R]) Match(onLeft func(L), onRight func(R)) {
	if e.isRight {
		onRight(e.right)
	} else {
		onLeft(e.left)
	}
}

// MatchWithReturn performs pattern matching and returns a value
func MatchWithReturn[L, R, T any](e Either[L, R], onLeft func(L) T, onRight func(R) T) T {
	if e.isRight {
		return onRight(e.right)
	}
	return onLeft(e.left)
}

// MapLeft applies a function to the Left value if present
func MapLeft[L, R, U any](e Either[L, R], f func(L) U) Either[U, R] {
	if e.isRight {
		return Right[U, R](e.right)
	}
	return Left[U, R](f(e.left))
}

// MapRight applies a function to the Right value if present
func MapRight[L, R, U any](e Either[L, R], f func(R) U) Either[L, U] {
	if !e.isRight {
		return Left[L, U](e.left)
	}
	return Right[L, U](f(e.right))
}

// BiMap applies functions to both Left and Right values
func BiMap[L, R, U, V any](e Either[L, R], leftF func(L) U, rightF func(R) V) Either[U, V] {
	if e.isRight {
		return Right[U, V](rightF(e.right))
	}
	return Left[U, V](leftF(e.left))
}

// AndThenLeft chains computations on Left values
func AndThenLeft[L, R, U any](e Either[L, R], f func(L) Either[U, R]) Either[U, R] {
	if e.isRight {
		return Right[U, R](e.right)
	}
	return f(e.left)
}

// AndThenRight chains computations on Right values (equivalent to flatMap)
func AndThenRight[L, R, U any](e Either[L, R], f func(R) Either[L, U]) Either[L, U] {
	if !e.isRight {
		return Left[L, U](e.left)
	}
	return f(e.right)
}

// Swap swaps Left and Right values
func (e Either[L, R]) Swap() Either[R, L] {
	if e.isRight {
		return Left[R, L](e.right)
	}
	return Right[R, L](e.left)
}

// ToResult converts Either to Result, treating Left as error
// This requires Left to be of type error
func ToResult[R any](e Either[error, R]) Result[R] {
	if e.isRight {
		return Ok(e.right)
	}
	return Err[R](e.left)
}

// FromResult converts Result to Either
func FromResult[R any](r Result[R]) Either[error, R] {
	val, err := r.Unwrap()
	if err != nil {
		return Left[error, R](err)
	}
	return Right[error, R](val)
}

// Convenient aliases
func L[L, R any](left L) Either[L, R]   { return Left[L, R](left) }
func R[L, R any](right R) Either[L, R]  { return Right[L, R](right) }