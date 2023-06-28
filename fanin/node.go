package fanin

import (
	"context"
	"errors"
	"sync"

	"github.com/itohio/graco"
)

type Node[T any] struct {
	name           string
	synchroBuilder SynchronizerBuilder
	inputs         []graco.SourceEdge[T]
	output         graco.SourceEdge[[]T]
	synchro        Synchronizer
}

func New[T any](name string, synchro SynchronizerBuilder) *Node[T] {
	res := &Node[T]{
		name:           name,
		synchroBuilder: synchro,
	}
	return res
}

func (n *Node[T]) Close() error {
	if n.output == nil {
		return nil
	}
	return errors.Join(n.synchro.Close(), n.output.Close())
}
func (n *Node[T]) Name() string { return n.name }

func (n *Node[T]) Connect(in ...graco.SourceEdge[T]) (graco.SourceEdge[[]T], error) {
	n.inputs = in
	for _, i := range in {
		err := i.Connect(n)
		if err != nil {
			return nil, err
		}
	}
	var err error
	n.synchro, err = n.synchroBuilder(len(in))
	if err != nil {
		return nil, err
	}
	n.output, err = graco.NewSourceEdge[[]T]("o", n, 1, false)
	return n.output, err
}

func (n *Node[T]) Start(ctx context.Context) error {
	if err := graco.IsEdgeValid(n.output); err != nil {
		return err
	}
	for _, in := range n.inputs {
		if err := graco.IsEdgeValid(in); err != nil {
			return err
		}
	}
	if n.synchro == nil {
		return errors.New("synchro nil")
	}

	c := make(chan []any)
	var globalErr error
	go func() {
		for res := range c {
			arr := make([]T, len(res))
			for i, in := range res {
				val, ok := in.(T)
				if !ok {
					panic("type corruption")
				}
				arr[i] = val
			}

			if err := n.output.Send(ctx, arr); err != nil {
				globalErr = err
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for i, in := range n.inputs {
		wg.Add(1)
		go func(i int, in graco.SourceEdge[T]) {
			defer wg.Done()
			for {
				val, err := in.Recv(ctx)
				if err != nil {
					if globalErr == nil {
						globalErr = err
					}
					return
				}
				res := n.synchro.Add(i, val)
				if res != nil {
					c <- res
				}
			}
		}(i, in)
	}
	wg.Wait()
	close(c)
	return globalErr
}
