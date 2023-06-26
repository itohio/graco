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
	builder1 := graco.NewChannelBuilder[int]("int", 1, false)
	builder2 := graco.NewChannelBuilder[float32]("float32", 1, false)
	builder3 := graco.NewChannelBuilder[pair]("pair", 1, false)

	t1 := ticker.NewIndex("int", builder1, 0, 1)
	t2 := ticker.NewIndex("float32", builder2, 5, .01)

	join := fanin.NewPair("pair", builder3, func(a int, b float32) (pair, error) {
		return pair{i: a, f: b}, nil
	})

	sum := processor.New[pair, float32](
		"sum",
		builder2,
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
