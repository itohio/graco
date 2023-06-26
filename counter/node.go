package counter

import (
	"context"

	"github.com/itohio/graco/ticker"
)

type Node[T ticker.Countable] struct {
	name    string
	builder graco.EdgeBuilder[T]
	output  graco.TypedEdge[T]
	start   T
	step    T
}

func New[T ticker.Countable](name string, builder graco.EdgeBuilder[T], start, step T) *Node[T] {
	res := &Node[T]{
		name:    name,
		builder: builder,
		start:   start,
		step:    step,
	}
	return res
}

func (n *Node[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *Node[T]) Name() string { return n.name }

func (n *Node[T]) Connect() (graco.TypedEdge[T], error) {
	var err error
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *Node[T]) Start(ctx context.Context) error {
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
