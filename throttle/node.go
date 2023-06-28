package throttle

import (
	"context"
	"io"
	"time"

	"github.com/itohio/graco"
)

type Node[T any] struct {
	name     string
	input    graco.SourceEdge[T]
	output   graco.SourceEdge[T]
	interval time.Duration
	drop     bool
}

func New[T any](name string, interval time.Duration, drop bool) *Node[T] {
	res := &Node[T]{
		name:     name,
		interval: interval,
		drop:     drop,
	}
	return res
}

func (n *Node[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *Node[T]) Name() string { return n.name }

func (n *Node[T]) Connect(in graco.SourceEdge[T]) (graco.SourceEdge[T], error) {
	n.input = in
	err := in.Connect(n)
	if err != nil {
		return nil, err
	}
	n.output, err = graco.NewSourceEdge[T]("o", n, 1, false)
	return n.output, err
}

func (n *Node[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.input); err != nil {
		return err
	}
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	var (
		err     error
		val     T
		gotVal  bool
		counter int
		ts      time.Time = time.Now()
	)
	for {
		if !gotVal {
			val, err = n.input.Recv(ctx)
			if err != nil {
				return err
			}
			gotVal = true
		}

		now := time.Now()
		delta := now.Sub(ts)
		if delta > n.interval {
			counter = 0
			ts = now
		}

		if counter >= 1 {
			if n.drop {
				if closer, ok := any(val).(io.Closer); ok {
					if err := closer.Close(); err != nil {
						return err
					}
				}
				gotVal = false
			}
			continue
		}

		if err := n.output.Send(ctx, val); err != nil {
			return err
		}
		counter++
	}
}
