package gateway

import (
	"os"
	"path/filepath"
	"sync"
)

type MemoryStore struct {
	mu       sync.RWMutex
	snapshot string
	filePath string
	loaded   bool
}

func NewMemoryStore(curseDir string) *MemoryStore {
	return &MemoryStore{
		filePath: filepath.Join(curseDir, "MEMORY.md"),
	}
}

func (ms *MemoryStore) Load() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	data, err := os.ReadFile(ms.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			ms.snapshot = ""
			ms.loaded = true
			return nil
		}
		return err
	}
	ms.snapshot = string(data)
	ms.loaded = true
	return nil
}

func (ms *MemoryStore) Save(content string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if err := os.WriteFile(ms.filePath, []byte(content), 0644); err != nil {
		return err
	}
	ms.snapshot = content
	return nil
}

func (ms *MemoryStore) Snapshot() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.snapshot
}

func (ms *MemoryStore) Loaded() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.loaded
}

func (ms *MemoryStore) Path() string {
	return ms.filePath
}
