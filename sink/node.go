package sink

import (
	"context"
	"io"

	"github.com/itohio/graco"
)

type SinkCloser[T any] interface {
	io.Closer
	Sink(context.Context, T) error
}

type Node[T any] struct {
	name  string
	input graco.TypedEdge[T]
	f     SinkCloser[T]
}

func New[T any](name string) *Node[T] {
	res := &Node[T]{
		name: name,
	}
	return res
}

func NewFunc[T any](name string, f SinkCloser[T]) *Node[T] {
	res := &Node[T]{
		name: name,
		f:    f,
	}
	return res
}

func (n *Node[T]) Close() error { return n.f.Close() }
func (n *Node[T]) Name() string { return n.name }

func (n *Node[T]) Connect(in graco.TypedEdge[T]) error {
	n.input = in
	return in.Connect(n)
}

func (n *Node[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}

	for {
		val, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}

		if n.f != nil {
			if err := n.f.Sink(ctx, val); err != nil {
				return err
			}
		}
	}
}
