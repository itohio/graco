package fanin

import (
	"container/ring"
	"errors"
	"io"
	"sync"
	"time"
)

var (
	_ Synchronizer = (*TimestampSynchronizer)(nil)
	_ Synchronizer = (*FullnessSynchronizer)(nil)
)

type WithTimestamp interface {
	Timestamp() time.Duration
}

type SynchronizerBuilder func(items int) (Synchronizer, error)

type Synchronizer interface {
	Add(idx int, val any) []any
	Close() error
}

type tsItem struct {
	ts  time.Duration
	val any
}

type base struct {
	sync.Mutex
	rings []*ring.Ring
	depth int
}

type TimestampSynchronizer struct {
	base
	matchIdx []int
	matchVal []any
	delta    time.Duration
}

// unlinkRing will unlink N items from the ring.
// unlinkRing will also close closable values with optionally ignoring the last one.
// N = (n > 0 && n < Len) ? n : Len
func unlinkRing(r *ring.Ring, n int, last bool) (*ring.Ring, error) {
	var err error
	N := r.Len()
	if n > 0 && n < N {
		N = n
	}
	for i := 0; i < N; i++ {
		if !last && i == N-1 {
			continue
		}
		switch val := r.Value.(type) {
		case tsItem:
			if closer, ok := val.val.(io.Closer); ok {
				err = errors.Join(err, closer.Close())
			}
		case io.Closer:
			err = errors.Join(err, val.Close())
		}
	}
	return r.Unlink(N), err
}

func NewTimestampSynchronizer(items, depth int, delta time.Duration) *TimestampSynchronizer {
	res := &TimestampSynchronizer{
		matchIdx: make([]int, items),
		matchVal: make([]any, items),
		delta:    delta,
	}
	res.init(items, depth)
	return res
}

func (s *base) init(items, depth int) {
	s.rings = make([]*ring.Ring, items)
	s.depth = depth

	for i := range s.rings {
		s.rings[i] = ring.New(0)
	}
}

func (s *base) add(idx int, val any) {
	s.rings[idx] = s.rings[idx].Link(&ring.Ring{Value: val})
	if s.rings[idx].Len() >= s.depth {
		s.rings[idx], _ = unlinkRing(s.rings[idx], 1, true)
	}
}

func (s *base) Close() error {
	var err error
	for _, r := range s.rings {
		_, errC := unlinkRing(r, -1, true)
		err = errors.Join(err, errC)
	}
	return err
}

func (s *TimestampSynchronizer) Add(idx int, val any) []any {
	defer func() {
		for i := range s.matchVal {
			s.matchVal[i] = nil
		}
	}()

	wts, ok := val.(WithTimestamp)
	if !ok {
		return nil
	}
	ts := wts.Timestamp()

	s.Lock()
	defer s.Unlock()

	s.matchVal[idx] = val
	s.add(idx, tsItem{ts: ts, val: val})

	s.matchIdx[idx] = s.rings[idx].Len() - 1
	matched := 1
	for i, r := range s.rings {
		if i == idx {
			continue
		}
		for j := 0; i < s.depth; i++ {
			tsi := r.Value.(tsItem)
			delta := ts - tsi.ts
			if delta < s.delta && delta > -s.delta {
				s.matchIdx[i] = j
				s.matchVal[i] = tsi.val
				matched++
				break
			}
		}
	}

	if matched != len(s.rings) {
		return nil
	}

	res := make([]any, len(s.rings))
	for i := range s.rings {
		res[i] = s.matchVal[i]
		s.rings[i], _ = unlinkRing(s.rings[i], s.matchIdx[i]+1, false)
	}
	return res
}

type FullnessSynchronizer struct {
	base
}

// NewFullnessSynchronizer creates a synchronizer that emits data on fullness
func NewFullnessSynchronizer(items, depth int) *FullnessSynchronizer {
	res := &FullnessSynchronizer{}
	res.init(items, depth)
	return res
}

func (s *FullnessSynchronizer) Add(idx int, val any) []any {
	s.Lock()
	defer s.Unlock()

	s.add(idx, val)

	matched := 0
	for _, r := range s.rings {
		if r.Len() > 0 {
			matched++
		}
	}

	if matched != len(s.rings) {
		return nil
	}

	res := make([]any, len(s.rings))
	for i, r := range s.rings {
		res[i] = r.Value
		s.rings[i], _ = unlinkRing(r, 1, false)
	}
	return res
}

type IntervalSynchronizer struct {
	base
	interval time.Duration
	ts       time.Time
}

// NewIntervalSynchronizer creates a synchronizer that emits data after elapsed interval unless it is empty
func NewIntervalSynchronizer(items, depth int, interval time.Duration) *IntervalSynchronizer {
	res := &IntervalSynchronizer{
		interval: interval,
		ts:       time.Now(),
	}
	res.init(items, depth)
	return res
}

func (s *IntervalSynchronizer) Add(idx int, val any) []any {
	now := time.Now()
	s.Lock()
	defer s.Unlock()

	s.add(idx, val)

	if now.Sub(s.ts) < s.interval {
		return nil
	}

	matched := 0
	for _, r := range s.rings {
		if r.Len() > 0 {
			matched++
		}
	}

	if matched == 0 {
		return nil
	}

	s.ts = now
	res := make([]any, len(s.rings))
	for i, r := range s.rings {
		res[i] = r.Value
		s.rings[i], _ = unlinkRing(r, 1, false)
	}
	return res
}
