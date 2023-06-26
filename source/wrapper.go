package source

import "context"

var (
	_ SourceCloser[int] = sourceFuncWrapper[int]{}
)

type sourceFuncWrapper[T any] struct {
	f func(context.Context) (T, error)
}

func (s sourceFuncWrapper[T]) Close() error                          { return nil }
func (s sourceFuncWrapper[T]) Source(ctx context.Context) (T, error) { return s.f(ctx) }

func Func[T any](f func(context.Context) (T, error)) sourceFuncWrapper[T] {
	return sourceFuncWrapper[T]{
		f: f,
	}
}
