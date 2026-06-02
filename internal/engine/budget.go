package engine

import (
	"math"
	"sync/atomic"
)

type IterationBudget struct {
	remaining int64
	maxIter   int64
	unlimited bool
}

func NewIterationBudget(max int) *IterationBudget {
	if max <= 0 {
		return &IterationBudget{
			remaining: math.MaxInt64,
			maxIter:   math.MaxInt64,
			unlimited: true,
		}
	}
	return &IterationBudget{
		remaining: int64(max),
		maxIter:   int64(max),
	}
}

func (b *IterationBudget) Consume() bool {
	if b.unlimited {
		return true
	}
	for {
		current := atomic.LoadInt64(&b.remaining)
		if current <= 0 {
			return false
		}
		if atomic.CompareAndSwapInt64(&b.remaining, current, current-1) {
			return true
		}
	}
}

func (b *IterationBudget) Refund(n int64) {
	if b.unlimited {
		return
	}
	atomic.AddInt64(&b.remaining, n)
}

func (b *IterationBudget) Remaining() int64 {
	if b.unlimited {
		return math.MaxInt64
	}
	return atomic.LoadInt64(&b.remaining)
}

func (b *IterationBudget) Reset() {
	if b.unlimited {
		return
	}
	atomic.StoreInt64(&b.remaining, b.maxIter)
}

func (b *IterationBudget) Exhausted() bool {
	if b.unlimited {
		return false
	}
	return atomic.LoadInt64(&b.remaining) <= 0
}

var DefaultBudget = NewIterationBudget(0) // 0 = unlimited
