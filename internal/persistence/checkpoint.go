package persistence

import (
	"encoding/json"
	"time"

	"github.com/M523zappin/Curse-Core/internal/statemachine"
)

type Checkpoint struct {
	State       statemachine.State `json:"state"`
	Step        int                `json:"step"`
	MissionID   string             `json:"mission_id"`
	Sequence    int64              `json:"sequence"`
	LastHash    string             `json:"last_hash"`
	ActiveModel string             `json:"active_model,omitempty"`
	StagedFiles []string           `json:"staged_files,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
}

type CheckpointStore struct {
	filePath string
}

func NewCheckpointStore(filePath string) *CheckpointStore {
	return &CheckpointStore{filePath: filePath}
}

func (cs *CheckpointStore) Save(m *statemachine.Machine, seq int64, lastHash string, staged []string) error {
	cp := Checkpoint{
		State:       m.State(),
		Step:        m.Step(),
		MissionID:   m.MissionID(),
		Sequence:    seq,
		LastHash:    lastHash,
		StagedFiles: staged,
		Timestamp:   time.Now().UTC(),
	}
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(cs.filePath, data)
}

func (cs *CheckpointStore) Load() (*Checkpoint, error) {
	data, err := readFile(cs.filePath)
	if err != nil {
		return nil, err
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, err
	}
	return &cp, nil
}

func (cs *CheckpointStore) Exists() bool {
	return fileExists(cs.filePath)
}

func (cs *CheckpointStore) Path() string {
	return cs.filePath
}
