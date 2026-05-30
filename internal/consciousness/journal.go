package consciousness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ThoughtType string

const (
	ThoughtDecision    ThoughtType = "decision"
	ThoughtObservation ThoughtType = "observation"
	ThoughtError       ThoughtType = "error"
	ThoughtLearn       ThoughtType = "learn"
	ThoughtMutation    ThoughtType = "mutation"
)

type Thought struct {
	ID        string      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Type      ThoughtType `json:"type"`
	Agent     string      `json:"agent"`
	Action    string      `json:"action"`
	Outcome   string      `json:"outcome"`
	Duration  int64       `json:"duration_ms"`
	File      string      `json:"file,omitempty"`
	PrevID    string      `json:"prev_id,omitempty"`
}

type Journal struct {
	mu       sync.Mutex
	thoughts []Thought
	capacity int
	cursor   int
	count    int
	path     string
}

func NewJournal(capacity int, path string) *Journal {
	return &Journal{
		thoughts: make([]Thought, capacity),
		capacity: capacity,
		cursor:   0,
		count:    0,
		path:     path,
	}
}

func (j *Journal) Record(t Thought) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.count > 0 {
		prev := j.thoughts[(j.cursor-1+j.capacity)%j.capacity]
		if prev.ID != "" {
			t.PrevID = prev.ID
		}
	}

	j.thoughts[j.cursor] = t
	j.cursor = (j.cursor + 1) % j.capacity
	if j.count < j.capacity {
		j.count++
	}
}

func (j *Journal) Recent(n int) []Thought {
	j.mu.Lock()
	defer j.mu.Unlock()

	if n > j.count {
		n = j.count
	}
	result := make([]Thought, n)
	idx := (j.cursor - n + j.capacity) % j.capacity
	for i := 0; i < n; i++ {
		result[i] = j.thoughts[(idx+i)%j.capacity]
	}
	return result
}

func (j *Journal) Replay() []Thought {
	return j.Recent(j.count)
}

func (j *Journal) Save() error {
	j.mu.Lock()
	thoughts := make([]Thought, j.count)
	idx := (j.cursor - j.count + j.capacity) % j.capacity
	for i := 0; i < j.count; i++ {
		thoughts[i] = j.thoughts[(idx+i)%j.capacity]
	}
	j.mu.Unlock()

	dir := filepath.Dir(j.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("journal mkdir: %w", err)
	}
	data, err := json.Marshal(thoughts)
	if err != nil {
		return fmt.Errorf("journal marshal: %w", err)
	}
	if err := os.WriteFile(j.path, data, 0644); err != nil {
		return fmt.Errorf("journal write: %w", err)
	}
	return nil
}

func (j *Journal) Load() error {
	data, err := os.ReadFile(j.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("journal read: %w", err)
	}
	var thoughts []Thought
	if err := json.Unmarshal(data, &thoughts); err != nil {
		return fmt.Errorf("journal unmarshal: %w", err)
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	for i, t := range thoughts {
		j.thoughts[i] = t
	}
	j.count = len(thoughts)
	j.cursor = j.count % j.capacity
	return nil
}

func (j *Journal) Len() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.count
}
