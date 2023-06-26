package graph

import (
	"context"
	"errors"
	"io"
)

type Edge interface {
	io.Closer
	// Name is the name of the edge
	Name() string
	// Nodes returns source and destination nodes
	Nodes() (src Node, dst Node)
	// Connect populates destination node reference. Must be called from destination node after connection.
	Connect(Node) error
}

type EdgeStarter interface {
	Edge
	// Start will do complex initialization of the edge
	Start(context.Context) error
}

type TypedEdge[T any] interface {
	Edge
	C() chan T
	Send(context.Context, T) error
	Recv(context.Context) (T, error)
}

type ChannelEdge[T any] struct {
	name     string
	src, dst Node
	ch       chan T
}

type EdgeBuilder[T any] func(name string, src Node) (TypedEdge[T], error)

func NewChannelBuilder[T any](name string, src Node, cap int, prime bool) EdgeBuilder[T] {
	if cap < 0 {
		cap = 0
	}
	return func(name string, src Node) (TypedEdge[T], error) {
		res := &ChannelEdge[T]{
			name: name,
			src:  src,
			ch:   make(chan T, cap),
		}
		if prime && cap > 0 {
			var zero T
			res.ch <- zero
		}
		return res, nil
	}
}

func (e *ChannelEdge[T]) Name() string        { return e.name }
func (e *ChannelEdge[T]) Nodes() (Node, Node) { return e.src, e.dst }
func (e *ChannelEdge[T]) Connect(dst Node) error {
	e.dst = dst
	return nil
}
func (e *ChannelEdge[T]) C() chan T { return e.ch }
func (e *ChannelEdge[T]) Close() error {
	close(e.ch)
	return nil
}

func (e *ChannelEdge[T]) Send(ctx context.Context, val T) error {
	if e.dst == nil {
		return errors.New("output disconnected")
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case e.ch <- val:
	}
	return nil
}

func (e *ChannelEdge[T]) Recv(ctx context.Context) (T, error) {
	var zero T
	if e.src == nil {
		return zero, errors.New("input disconnected")
	}

	select {
	case <-ctx.Done():
		return zero, context.Cause(ctx)
	case c, ok := <-e.ch:
		if !ok {
			return zero, io.EOF
		}
		return c, nil
	}
}

func IsEdgeValid(e Edge) error {
	if e == nil {
		return errors.New("nil")
	}
	src, dst := e.Nodes()
	if src == nil {
		return errors.New("src nil")
	}
	if dst == nil {
		return errors.New("dst nil")
	}
	return nil
}
