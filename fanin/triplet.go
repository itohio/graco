package fanin

import (
	"context"
	"errors"

	"github.com/itohio/graco"
)

type TripletMakerFunc[A, B, C, Res any] func(A, B, C) (Res, error)

type TripletNode[A, B, C, Res any] struct {
	name   string
	a      graco.SourceEdge[A]
	b      graco.SourceEdge[B]
	c      graco.SourceEdge[C]
	output graco.SourceEdge[Res]
	make   TripletMakerFunc[A, B, C, Res]
}

func NewTriplet[A, B, C, Res any](name string, make TripletMakerFunc[A, B, C, Res]) *TripletNode[A, B, C, Res] {
	res := &TripletNode[A, B, C, Res]{
		name: name,
		make: make,
	}
	return res
}

func (n *TripletNode[A, B, C, Res]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *TripletNode[A, B, C, Res]) Name() string { return n.name }

func (n *TripletNode[A, B, C, Res]) Connect(a graco.SourceEdge[A], b graco.SourceEdge[B]) (graco.SourceEdge[Res], error) {
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

func (n *TripletNode[A, B, C, Res]) Start(ctx context.Context) error {
	err := errors.Join(graco.IsEdgeValid(n.output),
		graco.IsEdgeValid(n.a),
		graco.IsEdgeValid(n.b),
		graco.IsEdgeValid(n.c),
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
		valc, err := n.c.Recv(ctx)
		if err != nil {
			return err
		}

		res, err := n.make(vala, valb, valc)
		if err != nil {
			return err
		}

		if err := n.output.Send(ctx, res); err != nil {
			return err
		}
	}
}
