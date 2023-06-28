package ticker

import (
	"context"
	"time"

	"github.com/itohio/graco/source"
)

type Countable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~complex64 | ~complex128
}

type CountableInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func NewIndex[T Countable](name string, start, step T) *source.Node[T] {
	return source.New[T](
		name,
		source.Func[T](
			func(ctx context.Context) (T, error) {
				var s T
				s, start = start, start+step
				return s, nil
			},
		),
	)
}

func NewTimestamp(name string) *source.Node[int64] {
	return source.New[int64](
		name,
		source.Func[int64](
			func(ctx context.Context) (int64, error) {
				return time.Now().Unix(), nil
			},
		),
	)
}
