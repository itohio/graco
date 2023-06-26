package graco

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

type Graph interface {
	AddNode(seq int, n ...Node) error
	AddEdge(seq int, e ...Edge) error
	Start(ctx context.Context) error
}

type withStartSequence[T any] struct {
	seq int
	val T
}
type withStartSequenceSlice[T any] []withStartSequence[T]

func (n withStartSequenceSlice[T]) Len() int           { return len(n) }
func (n withStartSequenceSlice[T]) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n withStartSequenceSlice[T]) Less(i, j int) bool { return n[i].seq > n[j].seq }
func (n withStartSequenceSlice[T]) Count(c func(T) bool) int {
	cnt := 0
	for _, n := range n {
		if c(n.val) {
			cnt++
		}
	}
	return cnt
}

type ConcurrentGraph struct {
	nodes withStartSequenceSlice[Node]
	edges withStartSequenceSlice[Edge]
}

func New() *ConcurrentGraph {
	return &ConcurrentGraph{}
}

func (g *ConcurrentGraph) AddNode(seq int, n ...Node) error {
	for _, n := range n {
		g.nodes = append(g.nodes, withStartSequence[Node]{
			seq: seq,
			val: n,
		})
	}
	return nil
}

func (g *ConcurrentGraph) AddEdge(seq int, e ...Edge) error {
	for _, e := range e {
		g.edges = append(g.edges, withStartSequence[Edge]{
			seq: seq,
			val: e,
		})
	}
	return nil
}

func (g *ConcurrentGraph) Start(pctx context.Context) error {
	sort.Sort(g.edges)
	sort.Sort(g.nodes)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancelCause(pctx)
	defer cancel(nil)

	errarr := make([]error, len(g.nodes)+g.edges.Count(func(e Edge) bool {
		_, ok := e.(EdgeStarter)
		return ok
	}))
	var errnum atomic.Uint32

	for _, e := range g.edges {
		e, ok := e.val.(EdgeStarter)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(e EdgeStarter) {
			err := e.Start(ctx)
			if errors.Is(err, context.Canceled) {
				err = nil
			}
			cancel(err)
			if err != nil {
				errarr[errnum.Add(1)-1] = fmt.Errorf("edge '%s' failed: %w", e.Name(), err)
			}
			wg.Done()
		}(e)
	}

	for _, n := range g.nodes {
		wg.Add(1)
		go func(n Node) {
			err := n.Start(ctx)
			if errors.Is(err, context.Canceled) {
				err = nil
			}
			cancel(err)
			if err != nil {
				errarr[errnum.Add(1)-1] = fmt.Errorf("node '%s' failed: %w", n.Name(), err)
			}
			wg.Done()
		}(n.val)
	}
	wg.Wait()
	return errors.Join(errarr...)
}
