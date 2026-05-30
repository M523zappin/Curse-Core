package knowledge

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type EntryType string

const (
	TypeADR          EntryType = "adr"
	TypeArchitecture EntryType = "architecture"
	TypeDebugSession EntryType = "debug"
	TypeDecision     EntryType = "decision"
	TypePattern      EntryType = "pattern"
)

type KnowledgeEntry struct {
	ID          string                 `json:"id"`
	Type        EntryType              `json:"type"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Tags        []string               `json:"tags"`
	CodeRefs    []string               `json:"code_refs"`
	Related     []string               `json:"related"`
	Timestamp   time.Time              `json:"timestamp"`
	Checksum    string                 `json:"checksum"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type SessionSummary struct {
	SessionID string    `json:"session_id"`
	StartTime time.Time `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	TaskCount int       `json:"task_count"`
	Summary   string    `json:"summary"`
}

type SearchResult struct {
	Entry    KnowledgeEntry `json:"entry"`
	Score    float64        `json:"score"`
	Matches  []string       `json:"matches"`
}

type Index struct {
	mu       sync.RWMutex
	entries  []KnowledgeEntry
	byTag    map[string][]int
	byType   map[EntryType][]int
	indexDir string
}

func NewIndex(indexDir string) *Index {
	idx := &Index{
		entries:  make([]KnowledgeEntry, 0),
		byTag:    make(map[string][]int),
		byType:   make(map[EntryType][]int),
		indexDir: indexDir,
	}
	idx.Load()
	return idx
}

func (idx *Index) Add(entry KnowledgeEntry) string {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("k-%x", sha256.Sum256([]byte(entry.Title+entry.Body)))[:16]
	}
	entry.Timestamp = time.Now()
	data, _ := json.Marshal(entry)
	entry.Checksum = fmt.Sprintf("%x", sha256.Sum256(data))

	pos := len(idx.entries)
	idx.entries = append(idx.entries, entry)

	for _, tag := range entry.Tags {
		idx.byTag[tag] = append(idx.byTag[tag], pos)
	}
	idx.byType[entry.Type] = append(idx.byType[entry.Type], pos)

	idx.persist(entry)
	return entry.ID
}

func (idx *Index) Search(query string, limit int) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	query = strings.ToLower(query)
	terms := strings.Fields(query)

	var results []SearchResult
	for _, entry := range idx.entries {
		score := 0.0
		var matches []string

		body := strings.ToLower(entry.Body)
		title := strings.ToLower(entry.Title)

		for _, term := range terms {
			if strings.Contains(title, term) {
				score += 3.0
				matches = append(matches, fmt.Sprintf("title:%s", term))
			}
			if strings.Contains(body, term) {
				score += 1.0
				matches = append(matches, fmt.Sprintf("body:%s", term))
			}
			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), term) {
					score += 2.0
					matches = append(matches, fmt.Sprintf("tag:%s", term))
				}
			}
		}

		if score > 0 {
			results = append(results, SearchResult{
				Entry:   entry,
				Score:   score,
				Matches: matches,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

func (idx *Index) ByType(typ EntryType) []KnowledgeEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	positions, ok := idx.byType[typ]
	if !ok {
		return nil
	}
	entries := make([]KnowledgeEntry, len(positions))
	for i, pos := range positions {
		entries[i] = idx.entries[pos]
	}
	return entries
}

func (idx *Index) ByTag(tag string) []KnowledgeEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	positions, ok := idx.byTag[tag]
	if !ok {
		return nil
	}
	entries := make([]KnowledgeEntry, len(positions))
	for i, pos := range positions {
		entries[i] = idx.entries[pos]
	}
	return entries
}

func (idx *Index) RecordDebug(errorMsg, solution string, refs []string) string {
	return idx.Add(KnowledgeEntry{
		Type:     TypeDebugSession,
		Title:    fmt.Sprintf("Debug: %s", truncateStr(errorMsg, 60)),
		Body:     fmt.Sprintf("Error: %s\n\nSolution: %s", errorMsg, solution),
		Tags:     extractTags(errorMsg + " " + solution),
		CodeRefs: refs,
	})
}

func (idx *Index) RecordADR(title, body string, tags []string) string {
	return idx.Add(KnowledgeEntry{
		Type:  TypeADR,
		Title: title,
		Body:  body,
		Tags:  tags,
	})
}

func (idx *Index) All() []KnowledgeEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]KnowledgeEntry, len(idx.entries))
	copy(out, idx.entries)
	return out
}

func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.entries)
}

func (idx *Index) persist(entry KnowledgeEntry) {
	if idx.indexDir == "" {
		return
	}
	os.MkdirAll(idx.indexDir, 0755)
	path := filepath.Join(idx.indexDir, fmt.Sprintf("%s.json", entry.ID))
	data, _ := json.MarshalIndent(entry, "", "  ")
	os.WriteFile(path, data, 0644)
}

func (idx *Index) Load() {
	if idx.indexDir == "" {
		return
	}
	files, err := os.ReadDir(idx.indexDir)
	if err != nil {
		return
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(idx.indexDir, f.Name()))
		if err != nil {
			continue
		}
		var entry KnowledgeEntry
		if json.Unmarshal(data, &entry) != nil {
			continue
		}
		pos := len(idx.entries)
		idx.entries = append(idx.entries, entry)
		for _, tag := range entry.Tags {
			idx.byTag[tag] = append(idx.byTag[tag], pos)
		}
		idx.byType[entry.Type] = append(idx.byType[entry.Type], pos)
	}
}

func extractTags(s string) []string {
	tagSet := make(map[string]bool)
	lower := strings.ToLower(s)
	knownTags := []string{
		"api", "database", "network", "security", "auth",
		"frontend", "backend", "deployment", "testing",
		"compilation", "runtime", "concurrency", "memory",
		"configuration", "migration", "refactoring",
	}
	var tags []string
	for _, tag := range knownTags {
		if strings.Contains(lower, tag) {
			tagSet[tag] = true
		}
	}
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func (idx *Index) RecordSession(sessionID string, summary SessionSummary) string {
	return idx.Add(KnowledgeEntry{
		Type:  TypeADR,
		Title: fmt.Sprintf("Session: %s", sessionID),
		Body: fmt.Sprintf("Session %s ran for %s, completed %d tasks.\n\nSummary: %s",
			sessionID, FormatDuration(summary.Duration), summary.TaskCount, summary.Summary),
		Tags: []string{"session", "summary"},
	})
}

func (idx *Index) QueryContext(limit int) []KnowledgeEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	n := len(idx.entries)
	if n == 0 {
		return nil
	}

	start := n - limit
	if start < 0 {
		start = 0
	}

	out := make([]KnowledgeEntry, 0, n-start)
	for i := n - 1; i >= start; i-- {
		out = append(out, idx.entries[i])
	}
	return out
}

func (idx *Index) RecentByType(typ EntryType, limit int) []KnowledgeEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	positions, ok := idx.byType[typ]
	if !ok {
		return nil
	}

	start := len(positions) - limit
	if start < 0 {
		start = 0
	}

	out := make([]KnowledgeEntry, 0, len(positions)-start)
	for i := len(positions) - 1; i >= start; i-- {
		out = append(out, idx.entries[positions[i]])
	}
	return out
}

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
