package fanin

import (
	"context"
	"errors"

	"github.com/itohio/graco"
)

type PairMakerFunc[A, B, Res any] func(A, B) (Res, error)

type PairNode[A, B, Res any] struct {
	name   string
	a      graco.SourceEdge[A]
	b      graco.SourceEdge[B]
	output graco.SourceEdge[Res]
	make   PairMakerFunc[A, B, Res]
}

func NewPair[A, B, Res any](name string, make PairMakerFunc[A, B, Res]) *PairNode[A, B, Res] {
	res := &PairNode[A, B, Res]{
		name: name,
		make: make,
	}
	return res
}

func (n *PairNode[A, B, Res]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *PairNode[A, B, Res]) Name() string { return n.name }

func (n *PairNode[A, B, Res]) Connect(a graco.SourceEdge[A], b graco.SourceEdge[B]) (graco.SourceEdge[Res], error) {
	n.a = a
	n.b = b
	err := a.Connect(n)
	if err != nil {
		return nil, err
	}
	err = b.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = graco.NewSourceEdge[Res]("o", n, 1, false)
	return n.output, err
}

func (n *PairNode[A, B, Res]) Start(ctx context.Context) error {
	err := errors.Join(graco.IsEdgeValid(n.output),
		graco.IsEdgeValid(n.a),
		graco.IsEdgeValid(n.b),
	)
	if err != nil {
		return err
	}

	for {
		vala, err := n.a.Recv(ctx)
		if err != nil {
			return err
		}
		valb, err := n.b.Recv(ctx)
		if err != nil {
			return err
		}

		res, err := n.make(vala, valb)
		if err != nil {
			return err
		}

		if err := n.output.Send(ctx, res); err != nil {
			return err
		}
	}
}
