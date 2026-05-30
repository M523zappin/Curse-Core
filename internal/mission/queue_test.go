package mission

import "testing"

func TestEnqueueDequeue(t *testing.T) {
	q := NewQueue()
	m := New("test task", []string{"edit"}, "", 10)
	q.Enqueue(m)
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
	got := q.Dequeue()
	if got.ID != m.ID {
		t.Fatalf("wrong mission dequeued")
	}
	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got %d", q.Len())
	}
}

func TestPeek(t *testing.T) {
	q := NewQueue()
	q.Enqueue(New("first", nil, "", 5))
	q.Enqueue(New("second", nil, "", 5))
	m := q.Peek()
	if m.Task != "first" {
		t.Fatalf("expected 'first', got '%s'", m.Task)
	}
	if q.Len() != 2 {
		t.Fatalf("peek should not dequeue")
	}
}

func TestSetStatus(t *testing.T) {
	q := NewQueue()
	m := New("task", nil, "", 5)
	q.Enqueue(m)
	q.SetStatus(m.ID, StatusInProgress)
	if q.All()[0].Status != StatusInProgress {
		t.Fatalf("expected InProgress")
	}
}
