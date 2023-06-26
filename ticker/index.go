package ticker

import (
	"context"

	"github.com/itohio/graco"
)

type Countable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~complex64 | ~complex128
}

type CountableInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type IndexNode[T Countable] struct {
	name    string
	builder graco.EdgeBuilder[T]
	output  graco.TypedEdge[T]
	start   T
	step    T
}

func NewIndex[T Countable](name string, builder graco.EdgeBuilder[T], start, step T) *IndexNode[T] {
	res := &IndexNode[T]{
		name:    name,
		builder: builder,
		start:   start,
		step:    step,
	}
	return res
}

func (n *IndexNode[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *IndexNode[T]) Name() string { return n.name }

func (n *IndexNode[T]) Connect() (graco.TypedEdge[T], error) {
	var err error
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *IndexNode[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	for {
		if err := n.output.Send(ctx, n.start); err != nil {
			return err
		}
		n.start += n.step
	}
}
