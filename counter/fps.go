package counter

import (
	"context"
	"time"

	"github.com/itohio/graco/ticker"
)

type FPSNode[T ticker.CountableInt] struct {
	name     string
	builder  graco.EdgeBuilder[float32]
	output   graco.TypedEdge[float32]
	interval time.Duration
}

func NewFPS[T ticker.CountableInt](name string, builder graco.EdgeBuilder[float32], interval time.Duration) *FPSNode[T] {
	res := &FPSNode[T]{
		name:     name,
		builder:  builder,
		interval: interval,
	}
	return res
}

func (n *FPSNode[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *FPSNode[T]) Name() string { return n.name }

func (n *FPSNode[T]) Connect() (graco.TypedEdge[float32], error) {
	var err error
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *FPSNode[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	var (
		counter T
		ts      time.Time = time.Now()
		fps     float32
	)

	for {
		now := time.Now()
		delta := now.Sub(ts)
		if delta > n.interval {
			f := float32(counter) / float32(delta.Seconds())
			counter = 0
			fps = (fps + f) / 2
			ts = now
		}

		if err := n.output.Send(ctx, fps); err != nil {
			return err
		}
		counter++
	}
}
