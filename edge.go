package graco

import (
	"context"
	"errors"
	"io"
	"strings"
)

type Edge interface {
	io.Closer
	// Name is the name of the edge
	Name() string
	// Nodes returns source and destination nodes
	Nodes() (src Node, dst Node)
	// Connect populates destination node reference. Must be called from destination node during connection.
	Connect(Node) error
}

// EdgeStarter is an interface that extends the Edge interface and provides a method for complex initialization of the edge.
type EdgeStarter interface {
	Edge
	// Start will do complex initialization of the edge
	Start(context.Context) error
}

// TypedEdge is an interface that extends the Edge interface and provides methods for sending and receiving values of type T.
type TypedEdge[T any] interface {
	Edge
	// C returns the underlying channel used for communication
	C() chan T
	// Send sends a value of type T over the edge
	Send(context.Context, T) error
	// Recv receives a value of type T from the edge
	Recv(context.Context) (T, error)
}

// EdgeBuilder is a function type that creates a TypedEdge[T] given a name and source Node.
type EdgeBuilder[T any] func(name string, src Node) (TypedEdge[T], error)

// ChannelEdge is a basic implementation of the TypedEdge[T] interface using channels.
type ChannelEdge[T any] struct {
	name     string
	src, dst Node
	ch       chan T
}

// IsEdgeValid checks if an Edge is valid by ensuring it is not nil and has both source and destination nodes.
// Returns an error if the Edge is invalid.
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

// NewChannelBuilder creates an EdgeBuilder for creating ChannelEdge instances.
// prefix: Prefix string for naming the edge
// cap: Capacity of the underlying channel (buffer size)
// prime: Whether to prime the channel with a zero value
func NewChannelBuilder[T any](prefix string, cap int, prime bool) EdgeBuilder[T] {
	if cap < 0 {
		cap = 0
	}
	return func(name string, src Node) (TypedEdge[T], error) {
		res := &ChannelEdge[T]{
			name: strings.Join([]string{prefix, name}, "."),
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
