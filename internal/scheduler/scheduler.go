package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Task struct {
	Name     string
	Interval time.Duration
	Handler  func(context.Context) error
	LastRun  time.Time
}

type Scheduler struct {
	mu     sync.Mutex
	tasks  []Task
	ticker *time.Ticker
	cancel context.CancelFunc
	running bool
}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Add(name string, interval time.Duration, handler func(context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.tasks {
		if t.Name == name {
			s.tasks[i].Interval = interval
			s.tasks[i].Handler = handler
			return
		}
	}

	s.tasks = append(s.tasks, Task{
		Name:     name,
		Interval: interval,
		Handler:  handler,
	})
}

func (s *Scheduler) Run(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	ctx, s.cancel = context.WithCancel(ctx)
	s.ticker = time.NewTicker(1 * time.Second)
	s.mu.Unlock()

	log.Println("[scheduler] started")
	s.runReady(ctx)

	for {
		select {
		case <-ctx.Done():
			s.ticker.Stop()
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			log.Println("[scheduler] stopped")
			return

		case <-s.ticker.C:
			s.runReady(ctx)
		}
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Scheduler) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) Tasks() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Task, len(s.tasks))
	copy(out, s.tasks)
	return out
}

func (s *Scheduler) runReady(ctx context.Context) {
	s.mu.Lock()
	tasks := make([]Task, len(s.tasks))
	copy(tasks, s.tasks)
	s.mu.Unlock()

	now := time.Now()
	for _, t := range tasks {
		if now.Sub(t.LastRun) >= t.Interval {
			s.mu.Lock()
			for i := range s.tasks {
				if s.tasks[i].Name == t.Name {
					s.tasks[i].LastRun = now
				}
			}
			s.mu.Unlock()

			go func(task Task) {
				start := time.Now()
				err := task.Handler(ctx)
				duration := time.Since(start)
				if err != nil {
					log.Printf("[scheduler] task %s failed after %v: %v", task.Name, duration, err)
				} else {
					log.Printf("[scheduler] task %s completed in %v", task.Name, duration)
				}
			}(t)
		}
	}
}

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
