package source

import (
	"context"
	"errors"
	"io"

	"github.com/itohio/graco"
)

type SourceCloser[T any] interface {
	io.Closer
	Source(context.Context) (T, error)
}

type Node[T any] struct {
	name   string
	output graco.SourceEdge[T]
	f      SourceCloser[T]
}

func New[T any](name string, f SourceCloser[T]) *Node[T] {
	res := &Node[T]{
		name: name,
		f:    f,
	}
	return res
}

func (n *Node[T]) Close() error {
	err := n.f.Close()
	if n.output == nil {
		return err
	}
	return errors.Join(err, n.output.Close())
}
func (n *Node[T]) Name() string { return n.name }

func (n *Node[T]) Connect() (graco.SourceEdge[T], error) {
	var err error
	n.output, err = graco.NewSourceEdge[T]("o", n, 1, false)
	return n.output, err
}

func (n *Node[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	for {
		val, err := n.f.Source(ctx)
		if err != nil {
			return err
		}

		if err := n.output.Send(ctx, val); err != nil {
			return err
		}
	}
}
