package processor

import (
	"context"
	"errors"
	"io"

	"github.com/itohio/graco"
)

var (
	ErrDrop = errors.New("drop")
	ErrStop = errors.New("stop")
)

type ProcessCloser[T, Res any] interface {
	io.Closer
	Process(context.Context, T) (Res, error)
}

type Node[Tin, To any] struct {
	name    string
	builder graco.EdgeBuilder[To]
	input   graco.TypedEdge[Tin]
	output  graco.TypedEdge[To]
	process ProcessCloser[Tin, To]
}

func New[Tin, To any](name string, builder graco.EdgeBuilder[To], processor ProcessCloser[Tin, To]) *Node[Tin, To] {
	res := &Node[Tin, To]{
		name:    name,
		builder: builder,
		process: processor,
	}
	return res
}

func (n *Node[T, To]) Close() error {
	err := n.process.Close()
	if n.output == nil {
		return err
	}
	return errors.Join(err, n.output.Close())
}
func (n *Node[T, To]) Name() string { return n.name }

func (n *Node[T, To]) Connect(in graco.TypedEdge[T]) (graco.TypedEdge[To], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *Node[T, To]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	if n.process == nil {
		return errors.New("processor nil")
	}

	for {
		val, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}

		res, err := n.process.Process(ctx, val)
		if errors.Is(err, ErrDrop) {
			continue
		}
		if err != nil {
			return err
		}

		if err := n.output.Send(ctx, res); err != nil {
			return err
		}
	}
}
