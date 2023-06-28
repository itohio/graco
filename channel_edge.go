package graco

import (
	"context"
	"errors"
	"io"
)

var (
	_ SourceEdge[int]               = (*ChannelSourceEdge[int])(nil)
	_ DestinationEdge[int, float32] = (*ChannelDestinationEdge[int, float32])(nil)
)

// ChannelSourceEdge is a basic implementation of the StreamingEdge[T] interface using channels.
type ChannelSourceEdge[T any] struct {
	name     string
	src, dst Node
	ch       chan T
}

// ChannelDestinationEdge is an extention.
type ChannelDestinationEdge[T, Tresp any] struct {
	*ChannelSourceEdge[T]
	reply *ChannelSourceEdge[Tresp]
}

// NewSourceEdge creates an ChannelSourceEdge[T] instance that satisfies SourceEdge[T] interface.
// prefix: Prefix string for naming the edge
// cap: Capacity of the underlying channel (buffer size)
// prime: Whether to prime the channel with a zero value
func NewSourceEdge[T any](name string, src Node, cap int, prime bool) (*ChannelSourceEdge[T], error) {
	if cap < 0 {
		cap = 0
	}
	res := &ChannelSourceEdge[T]{
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

// NewDestinationEdgeBuilder creates an EdgeBuilder for creating Edge instances that satisfy SourceEdge and DestinationEdge interfaces.
// cap: Capacity of the underlying channel (buffer size)
func NewDestinationEdge[T, Tresp any](prefix string, src Node, cap int) (*ChannelDestinationEdge[T, Tresp], error) {
	if cap < 0 {
		cap = 0
	}
	se, err := NewSourceEdge[T](prefix, src, cap, false)
	if err != nil {
		return nil, err
	}
	de, err := NewSourceEdge[Tresp](prefix, src, cap, false)
	if err != nil {
		return nil, err
	}
	res := &ChannelDestinationEdge[T, Tresp]{
		ChannelSourceEdge: se,
		reply:             de,
	}
	return res, nil
}

func (e *ChannelSourceEdge[T]) Name() string        { return e.name }
func (e *ChannelSourceEdge[T]) Nodes() (Node, Node) { return e.src, e.dst }
func (e *ChannelSourceEdge[T]) Connect(dst Node) error {
	e.dst = dst
	return nil
}
func (e *ChannelSourceEdge[T]) C() chan T { return e.ch }
func (e *ChannelSourceEdge[T]) Close() error {
	close(e.ch)
	return nil
}

func (e *ChannelSourceEdge[T]) Send(ctx context.Context, val T) error {
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

func (e *ChannelSourceEdge[T]) Recv(ctx context.Context) (T, error) {
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

func (e *ChannelDestinationEdge[T, Tr]) Reply() (SourceEdge[Tr], error) {
	return e.reply, nil
}
