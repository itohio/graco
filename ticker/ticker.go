package ticker

import (
	"context"
	"time"

	"github.com/itohio/graco"
)

type Node struct {
	name     string
	builder  graco.EdgeBuilder[int64]
	output   graco.TypedEdge[int64]
	interval time.Duration
}

func New(name string, builder graco.EdgeBuilder[int64], interval time.Duration) *Node {
	res := &Node{
		name:     name,
		builder:  builder,
		interval: interval,
	}
	return res
}

func (n *Node) Close() error {
	if n.output == nil {
		return nil
	}
	return n.output.Close()
}
func (n *Node) Name() string { return n.name }

func (n *Node) Connect() (graco.TypedEdge[int64], error) {
	var err error
	n.output, err = n.builder("o", n)
	return n.output, err
}

func (n *Node) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}

	ticker := time.NewTicker(n.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-ticker.C:
		}
		if err := n.output.Send(ctx, time.Now().Unix()); err != nil {
			return err
		}
	}
}
