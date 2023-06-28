package throttle

import (
	"context"
	"time"

	"github.com/itohio/graco"
)

type SleeperNode[T any] struct {
	name     string
	input    graco.SourceEdge[T]
	output   graco.SourceEdge[T]
	interval time.Duration
}

func NewSleeper[T any](name string, limit int, interval time.Duration) *SleeperNode[T] {
	res := &SleeperNode[T]{
		name:     name,
		interval: interval,
	}
	return res
}

func (n *SleeperNode[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *SleeperNode[T]) Name() string { return n.name }

func (n *SleeperNode[T]) Connect(in graco.SourceEdge[T]) (graco.SourceEdge[T], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = graco.NewSourceEdge[T]("o", n, 1, false)
	return n.output, err
}

func (n *SleeperNode[T]) Start(ctx context.Context) error {
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

		if err := n.output.Send(ctx, val); err != nil {
			return err
		}

		time.Sleep(n.interval)
	}
}
