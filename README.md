# graco

## Overview
graco (Generic Concurrent Graph Computational framework) is a powerful standalone library designed to 
address the challenges faced during the implementation of various architecture solutions in different 
domains such as acoustic measurement processing, robotics brain processing, GPT complex processing, and more. 
This framework emerged from several iterations and learnings from previous attempts, which resulted in a 
comprehensive solution that leverages static typing to enhance reliability, performance, and code simplicity.

## Motivation
Prior to developing graco, numerous alternative approaches were explored, initially in Python and later in Go 1.18. 
However, these implementations lacked the desired level of generality and were prone to runtime failures caused by 
type mismatches. Although attempts were made to mitigate this issue by adding type checks, it became evident that a 
more robust solution was needed.

Based on these experiences, the decision was made to create graco as a straightforward concurrent graph computational 
framework that capitalizes on the benefits of static typing. By enforcing type matching at compile time, developers 
are alerted to any incompatible connections between computational units, resulting in a more reliable and maintainable 
system. Moreover, this approach not only simplifies the code and API but also enhances performance and composability.

## Key Features
- **Static Type Enforcement:** graco enforces type compatibility between computational units during compilation, preventing runtime errors caused by type mismatches and ensuring a more reliable system.
- **Simplified Code and API:** The framework's design focuses on simplicity, resulting in cleaner and more maintainable code. The streamlined API facilitates ease of use and reduces the learning curve for developers.
- **Enhanced Performance:** By leveraging static typing, graco achieves improved performance by eliminating unnecessary runtime checks and enabling efficient optimization by the compiler.
- **Composability:** The framework's architecture supports the seamless composition of computational units, allowing developers to construct complex systems by connecting compatible components.

## The Principle
graco follows a principle of concurrent graph computation, where nodes represent individual computational units that 
execute concurrently. These nodes communicate with each other using typed Golang channels, wrapped in `SourceEdge[T]` and `DestinationEdge[T, Tresp]`.

Each node in graco is implemented as e.g. a `Node[Tin, To]` structure, where `Tin` represents the input type and `To` represents 
the output type. This structure encapsulates the essential components and behaviors of a node.

In order for the node to be useful, it must contain a `Connect` method that accepts arguments of type `SourceEdge[T]` or `DestinationEdge[T, Tresp]`, and possibly returns one or more
`SourceEdge[T]` values plus an error.

### Node Connectivity
Nodes in graco expose a `Connect` method that allows connecting the input edge of the node to another typed edge. 
The `Connect` method takes an argument of type `SourceEdge[T]` and returns a `SourceEdge[To]` along with 
an error. This connection establishes the flow of data between nodes.

`SourceEdge[T]` is used to send data from a source node to a destination node, like a one-way street. However,
in some situations, it is desirable to have a request-reply behavior. This is where `DestinationEdge[T, Tresp]` comes in. It 
provides a method `Reply` that returns a feedback edge that is of type `Tresp`. This edge is used to send feedback data from destination node
to the source node asynchronously.

### Node Execution
The `Start` method initiates the execution of a node. It receives a context `ctx` as a parameter and runs indefinitely, 
processing input values and producing corresponding output values. 

### Primitives
graco provides a set of predefined primitives that can be used to construct complex computational systems:

- **Source**: Represents a node that accepts an interface to a data sourcer. It serves as the entry point of data into the graco graph.
- **Sink**: Represents a node that accepts an interface to a data sinker. It serves as the exit point of data from the graco graph.
- **Processor**: Represents a node that accepts an interface to process type A and produce type B. It encapsulates the processing logic for transforming input data.
- **Fan-in**: Allows specifying a synchronizer from multiple inputs, combining them into a single output.
- **Fan-in (two/three different types)**: Specialized fan-in nodes that handle combining inputs of two or three different types.
- **Fan-out**: Represents a node that routes data to multiple types, supporting clonable types for efficient distribution.
- **Rate limiters**: These primitives act as matchers between producers and consumers, ensuring that the consumer is rate-limited and can handle slow processing rates.
- **Tickers**: These primitives act as monotonic counters, time-stampers, or FPS counters.

By leveraging these primitives and custom nodes built with the `Node[Tin, To]` structure, users can construct sophisticated 
computational graphs to solve complex problems effectively.

## Installation
To use graco in your Go project, add the library as a dependency:

```shell
go get github.com/itohio/graco
```

## Example Usage

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/itohio/graco"
	"github.com/itohio/graco/fanin"
	"github.com/itohio/graco/processor"
	"github.com/itohio/graco/sink"
	"github.com/itohio/graco/ticker"
)

type pair struct {
	i int
	f float32
}

func main() {
	t1 := ticker.NewIndex[int]("int", 0, 1)
	t2 := ticker.NewIndex[float32]("float32", 5, .01)

	join := fanin.NewPair("pair", func(a int, b float32) (pair, error) {
		return pair{i: a, f: b}, nil
	})

	sum := processor.New[pair, float32](
		"sum",
		processor.Func[pair, float32](
			func(ctx context.Context, p pair) (float32, error) {
				return p.f + float32(p.i), nil
			},
		),
	)

	sink := sink.NewFunc[float32]("sum", sink.Func(func(ctx context.Context, val float32) error {
		log.Println(val)
		return nil
	}))

	e1, err := t1.Connect()
	if err != nil {
		panic(err)
	}
	e2, err := t2.Connect()
	if err != nil {
		panic(err)
	}
	ej, err := join.Connect(e1, e2)
	if err != nil {
		panic(err)
	}
	es, err := sum.Connect(ej)
	if err != nil {
		panic(err)
	}
	err = sink.Connect(es)
	if err != nil {
		panic(err)
	}

	g := graco.New()
	err = g.AddNode(0, t1, t2, join, sum, sink)
	if err != nil {
		panic(err)
	}
	err = g.AddEdge(0, e1, e2, ej, es)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

	go func() {
		err = g.Start(ctx)
		log.Println("Start finished with: ", err.Error())
		if err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
}
```

## Contributions
Contributions to graco are welcome! If you encounter any issues, have suggestions for improvements, or would like to contribute code, please open an issue or submit a pull request.

## License
graco is released under the [MIT License](https://opensource.org/licenses/MIT). See the [LICENSE](https://github.com/your-username/graco/blob/main/LICENSE) file for more details.

## Acknowledgements
The development of graco would not have been possible without the support and inspiration from the following individuals and projects:
- [Person/Project 1]
- [Person/Project 2]
- [Person/Project 3]

