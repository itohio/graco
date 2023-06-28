package throttle

import (
	"context"
	"io"

	"github.com/itohio/graco"
)

type DropNode[T any] struct {
	name   string
	input  graco.SourceEdge[T]
	output graco.SourceEdge[T]
}

func NewDrop[T any](name string) *DropNode[T] {
	res := &DropNode[T]{
		name: name,
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

func (n *DropNode[T]) Connect(in graco.SourceEdge[T]) (graco.SourceEdge[T], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = graco.NewSourceEdge[T]("o", n, 1, false)
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
