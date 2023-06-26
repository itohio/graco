package fanout

import (
	"context"
	"errors"

	"github.com/itohio/graco"
)

type Cloner[T any] interface {
	Clone() (T, error)
}

type Node[T any] struct {
	name    string
	builder graco.EdgeBuilder[T]
	input   graco.TypedEdge[T]
	outputs []graco.TypedEdge[T]
}

func New[T any](name string, builder graco.EdgeBuilder[T], N int) *Node[T] {
	res := &Node[T]{
		name:    name,
		builder: builder,
		outputs: make([]graco.TypedEdge[T], N),
	}
	return res
}

func (n *Node[T]) Close() error {
	if n.outputs == nil {
		return nil
	}
	es := make([]error, len(n.outputs))
	for i, o := range n.outputs {
		es[i] = o.Close()
	}
	return errors.Join(es...)
}
func (n *Node[T]) Name() string { return n.name }

func (n *Node[T]) Connect(in graco.TypedEdge[T]) ([]graco.TypedEdge[T], error) {
	n.input = in
	err := in.Connect(n)
	for i := range n.outputs {
		n.outputs[i], err = n.builder("o", n)
	}
	return n.outputs, err
}

func (n *Node[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}
	for _, out := range n.outputs {
		if err := graco.IsEdgeValid(out); err != nil {
			return err
		}
	}

	for {
		val, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}

		cloner, ok := any(val).(Cloner[T])

		for _, o := range n.outputs {
			v := val
			if ok {
				v, err = cloner.Clone()
				if err != nil {
					return err
				}
			}
			if err := o.Send(ctx, v); err != nil {
				return err
			}
		}
	}
}
