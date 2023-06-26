package ticker

import (
	"context"
	"io"

	"github.com/itohio/graco"
)

type DropNode[T any] struct {
	name    string
	builder graco.EdgeBuilder[T]
	input   graco.TypedEdge[T]
	output  graco.TypedEdge[T]
}

func NewDrop[T any](name string, builder graco.EdgeBuilder[T]) *DropNode[T] {
	res := &DropNode[T]{
		name:    name,
		builder: builder,
	}
	return res
}

func (n *DropNode[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *DropNode[T]) Name() string { return n.name }

func (n *DropNode[T]) Connect(in graco.TypedEdge[T]) (graco.TypedEdge[T], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *DropNode[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	for {
		val, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case n.output.C() <- val:
		default:
			if closer, ok := any(val).(io.Closer); ok {
				if err := closer.Close(); err != nil {
					return err
				}
			}
		}
	}
}
