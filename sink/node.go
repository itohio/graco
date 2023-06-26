package sink

import (
	"context"
	"io"
)

type CloserNode struct {
	name  string
	input graco.TypedEdge[io.Closer]
}

func NewCloser(name string) *CloserNode {
	res := &CloserNode{
		name: name,
	}
	return res
}

func (n *CloserNode) Close() error { return nil }
func (n *CloserNode) Name() string { return n.name }

func (n *CloserNode) Connect(in graco.TypedEdge[io.Closer]) error {
	n.input = in
	return in.Connect(n)
}

func (n *CloserNode) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}

	for {
		val, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}

		if err := val.Close(); err != nil {
			return err
		}
	}
}
