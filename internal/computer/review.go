package computer

import (
	"fmt"
	"sync"
	"time"
)

const (
	ReviewTimeout = 5 * time.Minute
)

type ReviewManager struct {
	mu            sync.RWMutex
	pending       map[string]*ReviewRequest
	history       []ReviewRecord
	callback      func(ReviewRequest)
	autoApprove   bool
	maxBufferSize int
}

type ReviewRecord struct {
	Action    UIAction       `json:"action"`
	Decision  ReviewDecision `json:"decision"`
	Timestamp time.Time      `json:"timestamp"`
	Duration  time.Duration  `json:"duration_ms"`
}

func NewReviewManager() *ReviewManager {
	return &ReviewManager{
		pending:       make(map[string]*ReviewRequest),
		history:       make([]ReviewRecord, 0, 100),
		maxBufferSize: 100,
	}
}

func (rm *ReviewManager) RequestReview(action UIAction) (*ReviewRequest, <-chan ReviewDecision) {
	ch := make(chan ReviewDecision, 1)
	req := &ReviewRequest{
		Action:    action,
		Channel:   ch,
		CreatedAt: time.Now(),
	}

	rm.mu.Lock()
	rm.pending[action.ID] = req
	cb := rm.callback
	rm.mu.Unlock()

	if cb != nil {
		go cb(*req)
	}

	return req, ch
}

func (rm *ReviewManager) Resolve(actionID string, decision ReviewDecision) error {
	rm.mu.Lock()
	req, ok := rm.pending[actionID]
	if !ok {
		rm.mu.Unlock()
		return fmt.Errorf("no pending review for action %s", actionID)
	}

	req.Channel <- decision
	delete(rm.pending, actionID)

	record := ReviewRecord{
		Action:    req.Action,
		Decision:  decision,
		Timestamp: time.Now(),
		Duration:  time.Since(req.CreatedAt),
	}
	rm.history = append(rm.history, record)
	if len(rm.history) > rm.maxBufferSize {
		rm.history = rm.history[1:]
	}
	rm.mu.Unlock()

	return nil
}

func (rm *ReviewManager) SetCallback(cb func(ReviewRequest)) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.callback = cb
}

func (rm *ReviewManager) PendingReviews() []ReviewRequest {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	reviews := make([]ReviewRequest, 0, len(rm.pending))
	for _, req := range rm.pending {
		reviews = append(reviews, *req)
	}
	return reviews
}

func (rm *ReviewManager) PendingCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.pending)
}

func (rm *ReviewManager) ReviewHistory() []ReviewRecord {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	out := make([]ReviewRecord, len(rm.history))
	copy(out, rm.history)
	return out
}

func (rm *ReviewManager) SetAutoApprove(v bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.autoApprove = v
}

func (rm *ReviewManager) IsAutoApprove() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.autoApprove
}
