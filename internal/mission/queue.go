package mission

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Status int

const (
	StatusTodo       Status = iota
	StatusInProgress
	StatusDone
)

type Mission struct {
	ID        string    `json:"id"`
	Task      string    `json:"task"`
	Tags      []string  `json:"tags"`
	ModelHint string    `json:"model_hint,omitempty"`
	Status    Status    `json:"status"`
	Steps     int       `json:"steps"`
	MaxSteps  int       `json:"max_steps"`
	CreatedAt time.Time `json:"created_at"`
}

func New(task string, tags []string, modelHint string, maxSteps int) *Mission {
	return &Mission{
		ID:        uuid.NewString(),
		Task:      task,
		Tags:      tags,
		ModelHint: modelHint,
		Status:    StatusTodo,
		Steps:     0,
		MaxSteps:  maxSteps,
		CreatedAt: time.Now().UTC(),
	}
}

type Queue struct {
	mu       sync.Mutex
	missions []*Mission
}

func NewQueue() *Queue {
	return &Queue{missions: make([]*Mission, 0)}
}

func (q *Queue) Enqueue(m *Mission) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.missions = append(q.missions, m)
}

func (q *Queue) Dequeue() *Mission {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.missions) == 0 {
		return nil
	}
	m := q.missions[0]
	q.missions = q.missions[1:]
	return m
}

func (q *Queue) Peek() *Mission {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.missions) == 0 {
		return nil
	}
	return q.missions[0]
}

func (q *Queue) SetStatus(id string, status Status) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, m := range q.missions {
		if m.ID == id {
			m.Status = status
			return
		}
	}
}

func (q *Queue) All() []*Mission {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*Mission, len(q.missions))
	copy(out, q.missions)
	return out
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.missions)
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.missions = make([]*Mission, 0)
}
