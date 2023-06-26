package ticker

import (
	"context"
	"time"

	"github.com/itohio/graco"
	"github.com/itohio/graco/source"
)

func NewFps(name string, builder graco.EdgeBuilder[float32], interval time.Duration, smooth float32) *source.Node[float32] {
	var (
		counter int
		ts      time.Time = time.Now()
		fps     float32
	)

	return source.New[float32](
		name,
		builder,
		source.Func[float32](
			func(ctx context.Context) (float32, error) {
				now := time.Now()
				delta := now.Sub(ts)
				if delta > interval {
					momentFPS := float64(counter) / delta.Seconds()
					counter = 0
					fps = fps*(1-smooth) + smooth*float32(momentFPS)
				}

				counter++
				return fps, nil
			},
		),
	)
}
