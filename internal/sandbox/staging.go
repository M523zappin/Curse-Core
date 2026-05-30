package sandbox

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Mode int

const (
	ModeDraftFile Mode = iota
	ModeGitWorktree
)

type Status int

const (
	StatusPending  Status = iota
	StatusApproved
	StatusRejected
)

type StagedFile struct {
	ID         string    `json:"id"`
	SourcePath string    `json:"source_path"`
	TargetPath string    `json:"target_path"`
	Status     Status    `json:"status"`
	Checksum   string    `json:"checksum"`
	Content    []byte    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

type StagingArea struct {
	mu       sync.Mutex
	root     string
	mode     Mode
	index    map[string]*StagedFile
}

func New(root string, mode Mode) *StagingArea {
	os.MkdirAll(root, 0755)
	return &StagingArea{
		root:  root,
		mode:  mode,
		index: make(map[string]*StagedFile),
	}
}

func (s *StagingArea) Stage(targetPath string, content []byte) (*StagedFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	sourcePath := filepath.Join(s.root, id)
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		return nil, fmt.Errorf("write staged file: %w", err)
	}
	hash := sha256.Sum256(content)
	sf := &StagedFile{
		ID:         id,
		SourcePath: sourcePath,
		TargetPath: targetPath,
		Status:     StatusPending,
		Checksum:   fmt.Sprintf("%x", hash),
		Content:    content,
		CreatedAt:  time.Now().UTC(),
	}
	s.index[id] = sf
	return sf, nil
}

func (s *StagingArea) Approve(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[id]
	if !ok {
		return fmt.Errorf("staged file %s not found", id)
	}
	dir := filepath.Dir(sf.TargetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(sf.TargetPath, sf.Content, 0644); err != nil {
		return err
	}
	os.Remove(sf.SourcePath)
	sf.Status = StatusApproved
	delete(s.index, id)
	return nil
}

func (s *StagingArea) Reject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf, ok := s.index[id]
	if !ok {
		return fmt.Errorf("staged file %s not found", id)
	}
	os.Remove(sf.SourcePath)
	sf.Status = StatusRejected
	delete(s.index, id)
	return nil
}

func (s *StagingArea) List() []StagedFile {
	s.mu.Lock()
	defer s.mu.Unlock()

	files := make([]StagedFile, 0, len(s.index))
	for _, sf := range s.index {
		files = append(files, *sf)
	}
	return files
}

func (s *StagingArea) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.index)
}

func (s *StagingArea) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta := filepath.Join(s.root, ".index.json")
	data, err := json.MarshalIndent(s.index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(meta, data, 0644)
}

func (s *StagingArea) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta := filepath.Join(s.root, ".index.json")
	data, err := os.ReadFile(meta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var idx map[string]*StagedFile
	if err := json.Unmarshal(data, &idx); err != nil {
		return err
	}
	for id, sf := range idx {
		if content, err := os.ReadFile(sf.SourcePath); err == nil {
			sf.Content = content
		}
		s.index[id] = sf
	}
	return nil
}
