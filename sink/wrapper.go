package sink

import (
	"context"
	"io"
)

var (
	_ SinkCloser[int] = sinkFuncWrapper[int]{}
)

type sinkFuncWrapper[T any] struct {
	f func(context.Context, T) error
}

func (s sinkFuncWrapper[T]) Close() error                          { return nil }
func (s sinkFuncWrapper[T]) Sink(ctx context.Context, val T) error { return s.f(ctx, val) }

func Func[T any](f func(context.Context, T) error) sinkFuncWrapper[T] {
	return sinkFuncWrapper[T]{
		f: f,
	}
}

func Closer[T io.Closer](name string) sinkFuncWrapper[T] {
	res := Func[T](func(ctx context.Context, val T) error {
		return val.Close()
	})
	return res
}
