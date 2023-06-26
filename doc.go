// Package graco provides a Generic Concurrent Graph Computational framework for building concurrent graph-based computations.
// It offers a set of interfaces, implementations, and utility functions to create and manage computational nodes and their connections within a graph.
//
// Key Interfaces:
// - Graph: Acts as a container for nodes and edges. Manages their lifetime and computational resources.
// - Node: Represents a computational unit.
// - Edge: Represents a connection between nodes in the computational graph.
// - EdgeStarter: Extends the Edge interface with a Start method for complex edge initialization.
// - TypedEdge[T]: Extends the Edge interface with methods for sending and receiving typed data over the edge.
// - EdgeBuilder[T]: A function signature for building typed edges.
//
// Key primitives:
// - Source
// - Sink
// - Processor
// - FanIn with several synchronizers
// - FanOut
// - Tickers
// - Rate Limiters
package graco
