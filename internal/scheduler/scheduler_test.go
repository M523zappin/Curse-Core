package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestAddTask(t *testing.T) {
	s := New()
	s.Add("test", 100*time.Millisecond, func(ctx context.Context) error {
		return nil
	})

	tasks := s.Tasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "test" {
		t.Fatalf("expected task name 'test', got %s", tasks[0].Name)
	}
}

func TestRunAndStop(t *testing.T) {
	s := New()
	s.Add("quick", 50*time.Millisecond, func(ctx context.Context) error {
		return nil
	})

	ctx := context.Background()
	go s.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	if !s.Running() {
		t.Fatal("expected scheduler to be running")
	}

	s.Stop()
	time.Sleep(50 * time.Millisecond)
}

func TestTaskExecutes(t *testing.T) {
	s := New()
	var called atomic.Int32

	s.Add("counter", 10*time.Millisecond, func(ctx context.Context) error {
		called.Add(1)
		return nil
	})

	ctx := context.Background()
	go s.Run(ctx)

	time.Sleep(100 * time.Millisecond)
	s.Stop()

	if n := called.Load(); n == 0 {
		t.Fatal("expected task to be called at least once")
	}
}

func TestMultipleTasks(t *testing.T) {
	s := New()
	var a, b atomic.Int32

	s.Add("task-a", 20*time.Millisecond, func(ctx context.Context) error {
		a.Add(1)
		return nil
	})
	s.Add("task-b", 30*time.Millisecond, func(ctx context.Context) error {
		b.Add(1)
		return nil
	})

	ctx := context.Background()
	go s.Run(ctx)

	time.Sleep(150 * time.Millisecond)
	s.Stop()

	if n := a.Load(); n == 0 {
		t.Fatal("expected task-a to be called")
	}
	if n := b.Load(); n == 0 {
		t.Fatal("expected task-b to be called")
	}
}

func TestTaskUpdate(t *testing.T) {
	s := New()
	s.Add("update", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	s.Add("update", 10*time.Millisecond, func(ctx context.Context) error {
		return nil
	})

	tasks := s.Tasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task after update, got %d", len(tasks))
	}
	if tasks[0].Interval != 10*time.Millisecond {
		t.Fatalf("expected interval 10ms, got %v", tasks[0].Interval)
	}
}

func TestNotRunningBeforeRun(t *testing.T) {
	s := New()
	if s.Running() {
		t.Fatal("expected scheduler to not be running before Run()")
	}
}

func TestDoubleRun(t *testing.T) {
	s := New()
	ctx := context.Background()
	go s.Run(ctx)
	go s.Run(ctx)

	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00:00"},
		{time.Second, "00:00:01"},
		{time.Minute, "00:01:00"},
		{time.Hour, "01:00:00"},
		{3661 * time.Second, "01:01:01"},
		{25 * time.Hour, "25:00:00"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.d)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
