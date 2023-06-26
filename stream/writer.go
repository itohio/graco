package stream

import (
	"context"
	"errors"
	"io"
)

type WriterNode struct {
	name    string
	builder graco.EdgeBuilder[Blob]
	input   graco.TypedEdge[Blob]
	output  graco.TypedEdge[Blob]
	writer  io.WriteCloser
}

func NewWriter(name string, builder graco.EdgeBuilder[Blob], writer io.WriteCloser) *WriterNode {
	res := &WriterNode{
		name:    name,
		builder: builder,
		writer:  writer,
	}
	return res
}

func (n *WriterNode) Close() error {
	if n.output == nil {
		return nil
	}
	return errors.Join(n.writer.Close(), n.output.Close())
}
func (n *WriterNode) Name() string { return n.name }

func (n *WriterNode) Connect(in graco.TypedEdge[Blob]) (graco.TypedEdge[Blob], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *WriterNode) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	for {
		val, err := n.input.Recv(ctx)
		if err != nil {
			return err
		}

		data := val.Data()
		N := len(data)
		nacc := 0
		for {
			n, err := n.writer.Write(data)
			if err != nil {
				return err
			}
			nacc += n
			if nacc >= N {
				break
			}
		}

		if err := n.output.Send(ctx, val); err != nil {
			return err
		}
	}
}
