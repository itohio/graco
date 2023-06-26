package stream

import (
	"context"
	"errors"
	"io"
)

type ReadCloser interface {
	io.Closer
	Read(context.Context) (Blob, error)
}

type ReaderNode struct {
	name    string
	builder graco.EdgeBuilder[Blob]
	output  graco.TypedEdge[Blob]
	reader  ReadCloser
}

func NewReader(name string, builder graco.EdgeBuilder[Blob], reader ReadCloser) *ReaderNode {
	res := &ReaderNode{
		name:    name,
		builder: builder,
	}
	return res
}

func (n *ReaderNode) Close() error {
	if n.output == nil {
		return nil
	}
	return errors.Join(n.reader.Close(), n.output.Close())
}
func (n *ReaderNode) Name() string { return n.name }

func (n *ReaderNode) Connect() (graco.TypedEdge[Blob], error) {
	var err error
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *ReaderNode) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	for {
		val, err := n.reader.Read(ctx)
		if err != nil {
			return err
		}

		if err := n.output.Send(ctx, val); err != nil {
			return err
		}
	}
}
