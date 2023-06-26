package processor

import (
	"context"
	"errors"
)

type PassthroughNode[Tin, To any] struct {
	name       string
	builderIn  graco.EdgeBuilder[Tin]
	builderRes graco.EdgeBuilder[To]
	input      graco.TypedEdge[Tin]
	output     graco.TypedEdge[Tin]
	res        graco.TypedEdge[To]
	process    ProcessorFunc[Tin, To]
}

func NewPassthrough[Tin, To any](name string, builderIn graco.EdgeBuilder[Tin], builderRes graco.EdgeBuilder[To], processor ProcessorFunc[Tin, To]) *PassthroughNode[Tin, To] {
	res := &PassthroughNode[Tin, To]{
		name:       name,
		builderIn:  builderIn,
		builderRes: builderRes,
		process:    processor,
	}
	return res
}

func (n *PassthroughNode[T, To]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *PassthroughNode[T, To]) Name() string { return n.name }

func (n *PassthroughNode[T, To]) Connect(in graco.TypedEdge[T]) (graco.TypedEdge[T], graco.TypedEdge[To], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, nil, err
	}
	n.output, err = n.builderIn("o", n)
	if err != nil {
		return nil, nil, err
	}
	n.res, err = n.builderRes("o", n)
	if err != nil {
		return nil, nil, err
	}
	return n.output, n.res, nil
}

func (n *PassthroughNode[T, To]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}
	if err := graco.IsEdgeValid(n.res); err != nil {
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

		res, err := n.process(ctx, val)
		if err != nil {
			return err
		}

		if err := n.output.Send(ctx, val); err != nil {
			return err
		}
		if err := n.res.Send(ctx, res); err != nil {
			return err
		}
	}
}
