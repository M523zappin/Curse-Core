package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type State struct {
	SessionID      string    `json:"session_id"`
	StartedAt      time.Time `json:"started_at"`
	ActiveModel    string    `json:"active_model"`
	MachineState   string    `json:"machine_state"`
	MachineStep    int       `json:"machine_step"`
	MissionID      string    `json:"mission_id"`
	KnowledgeCount int       `json:"knowledge_count"`
	SkillCount     int       `json:"skill_count"`
	TaskCount      int       `json:"task_count"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(sessionDir string) *Store {
	os.MkdirAll(sessionDir, 0755)
	return &Store{
		path: filepath.Join(sessionDir, "session.json"),
	}
}

func (s *Store) Save(state State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}

func (s *Store) Load() (*State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &state, nil
}

func (s *Store) Exists() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := os.Stat(s.path)
	return err == nil
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.Remove(s.path)
}

func (s *Store) Path() string {
	return s.path
}

func NewState(sessionID, activeModel, machineState string, machineStep int) State {
	return State{
		SessionID:    sessionID,
		StartedAt:    time.Now(),
		ActiveModel:  activeModel,
		MachineState: machineState,
		MachineStep:  machineStep,
	}
}
