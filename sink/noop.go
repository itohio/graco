package sink

import (
	"context"
)

type Node[T any] struct {
	name  string
	input graco.TypedEdge[T]
}

func New[T any](name string) *Node[T] {
	res := &Node[T]{
		name: name,
	}
	return res
}

func (n *Node[T]) Close() error { return nil }
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
		_, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}
	}
}
