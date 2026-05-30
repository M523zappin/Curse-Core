package engine

import "sync/atomic"

type IterationBudget struct {
	remaining int64
	maxIter   int64
}

func NewIterationBudget(max int) *IterationBudget {
	return &IterationBudget{
		remaining: int64(max),
		maxIter:   int64(max),
	}
}

func (b *IterationBudget) Consume() bool {
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
	atomic.AddInt64(&b.remaining, n)
}

func (b *IterationBudget) Remaining() int64 {
	return atomic.LoadInt64(&b.remaining)
}

func (b *IterationBudget) Reset() {
	atomic.StoreInt64(&b.remaining, b.maxIter)
}

func (b *IterationBudget) Exhausted() bool {
	return atomic.LoadInt64(&b.remaining) <= 0
}

var DefaultBudget = NewIterationBudget(100)
