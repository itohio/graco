package graco

import (
	"context"
	"errors"
	"io"
)

// Edge is an interface that provides method for edges
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

// SourceEdge is an interface that extends the Edge interface and provides methods for sending and receiving values of type T from source to destination.
type SourceEdge[T any] interface {
	Edge
	// C returns the underlying channel used for communication
	C() chan T
	// Send sends a value of type T over the edge. Used by source node.
	Send(context.Context, T) error
	// Recv receives a value of type T from the edge. Used by destination node.
	Recv(context.Context) (T, error)
}

// DestinationEdge is an interface that extends the SourceEdge[T] interface and provides methods for returning an edge used to reply by destination node.
type DestinationEdge[T, Tresp any] interface {
	SourceEdge[T]
	// Reply returns an edge from destination to source that can be used to send replies back to the source from destination node.
	Reply() (SourceEdge[Tresp], error)
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
