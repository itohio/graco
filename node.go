package graph

import (
	"context"
	"io"
)

type Node interface {
	io.Closer
	Name() string
	Start(context.Context) error
	// Connect(A TypedEdge[TA], B TypedEdge[TB]) (TypedEdge[ResA], TypedEdge[ResB], error)
}

type NodeBase struct {
}
