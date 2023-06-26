package processor

import (
	"context"
)

var (
	_ ProcessCloser[int, float32] = processFuncWrapper[int, float32]{}
)

type processFuncWrapper[T, Res any] struct {
	f func(context.Context, T) (Res, error)
}

func (s processFuncWrapper[T, Res]) Close() error { return nil }
func (s processFuncWrapper[T, Res]) Process(ctx context.Context, val T) (Res, error) {
	return s.f(ctx, val)
}

func Func[T, Res any](f func(context.Context, T) (Res, error)) processFuncWrapper[T, Res] {
	return processFuncWrapper[T, Res]{
		f: f,
	}
}
