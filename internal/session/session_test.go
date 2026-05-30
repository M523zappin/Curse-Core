package session

import (
	"os"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-save")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)

	state := State{
		SessionID:      "test-session",
		StartedAt:      time.Now(),
		ActiveModel:    "gpt4",
		MachineState:   "Running",
		MachineStep:    42,
		KnowledgeCount: 10,
		SkillCount:     5,
		TaskCount:      3,
	}

	if err := s.Save(state); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.SessionID != "test-session" {
		t.Fatalf("expected session 'test-session', got %s", loaded.SessionID)
	}
	if loaded.ActiveModel != "gpt4" {
		t.Fatalf("expected model 'gpt4', got %s", loaded.ActiveModel)
	}
	if loaded.MachineStep != 42 {
		t.Fatalf("expected step 42, got %d", loaded.MachineStep)
	}
	if loaded.KnowledgeCount != 10 {
		t.Fatalf("expected knowledge count 10, got %d", loaded.KnowledgeCount)
	}
	if loaded.SkillCount != 5 {
		t.Fatalf("expected skill count 5, got %d", loaded.SkillCount)
	}
}

func TestExists(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-exists")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	if s.Exists() {
		t.Fatal("expected store to not exist before save")
	}

	s.Save(State{SessionID: "test"})
	if !s.Exists() {
		t.Fatal("expected store to exist after save")
	}
}

func TestClear(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-clear")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	s.Save(State{SessionID: "test"})

	if err := s.Clear(); err != nil {
		t.Fatalf("clear: %v", err)
	}

	if s.Exists() {
		t.Fatal("expected store to not exist after clear")
	}
}

func TestLoadNonExistent(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-noexist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	_, err = s.Load()
	if err == nil {
		t.Fatal("expected error loading non-existent session")
	}
}

func TestNewState(t *testing.T) {
	state := NewState("sess-1", "model-x", "Running", 7)
	if state.SessionID != "sess-1" {
		t.Fatalf("expected 'sess-1', got %s", state.SessionID)
	}
	if state.ActiveModel != "model-x" {
		t.Fatalf("expected 'model-x', got %s", state.ActiveModel)
	}
	if state.MachineState != "Running" {
		t.Fatalf("expected 'Running', got %s", state.MachineState)
	}
	if state.MachineStep != 7 {
		t.Fatalf("expected step 7, got %d", state.MachineStep)
	}
	if state.StartedAt.IsZero() {
		t.Fatal("expected non-zero StartedAt")
	}
}

func TestPath(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-path")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	if s.Path() == "" {
		t.Fatal("expected non-empty path")
	}
}
