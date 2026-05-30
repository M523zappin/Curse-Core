package consciousness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Pattern struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Observations int      `json:"observations"`
	Confidence   float64  `json:"confidence"`
	Tags         []string `json:"tags"`
}

type SoulProfile struct {
	mu            sync.Mutex
	patterns      map[string]*Pattern
	conventionLog []string
	path          string
}

func NewSoulProfile(path string) *SoulProfile {
	return &SoulProfile{
		patterns:      make(map[string]*Pattern),
		conventionLog: make([]string, 0, 100),
		path:          path,
	}
}

func (sp *SoulProfile) Observe(name, ptype string, tags []string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	key := name + "|" + ptype
	p, ok := sp.patterns[key]
	if !ok {
		p = &Pattern{
			Name: name,
			Type: ptype,
			Tags: tags,
		}
		sp.patterns[key] = p
	}
	p.Observations++
	p.Confidence = 1.0 - (1.0 / float64(p.Observations+1))
	if len(tags) > 0 {
		existing := make(map[string]bool)
		for _, t := range p.Tags {
			existing[t] = true
		}
		for _, t := range tags {
			if !existing[t] {
				p.Tags = append(p.Tags, t)
			}
		}
	}
}

func (sp *SoulProfile) LogConvention(convention string) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.conventionLog = append(sp.conventionLog, convention)
	if len(sp.conventionLog) > 1000 {
		sp.conventionLog = sp.conventionLog[len(sp.conventionLog)-500:]
	}
}

func (sp *SoulProfile) Patterns() []Pattern {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	result := make([]Pattern, 0, len(sp.patterns))
	for _, p := range sp.patterns {
		result = append(result, *p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})
	return result
}

func (sp *SoulProfile) TopPatterns(n int) []Pattern {
	all := sp.Patterns()
	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

func (sp *SoulProfile) KnownTypes() []string {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	types := make(map[string]bool)
	for _, p := range sp.patterns {
		types[p.Type] = true
	}
	result := make([]string, 0, len(types))
	for t := range types {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}

func (sp *SoulProfile) GenerateConstitution() string {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if len(sp.conventionLog) == 0 {
		return ""
	}

	freq := make(map[string]int)
	for _, c := range sp.conventionLog {
		freq[c]++
	}

	type counted struct {
		text string
		n    int
	}
	sorted := make([]counted, 0, len(freq))
	for text, n := range freq {
		sorted = append(sorted, counted{text, n})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].n > sorted[j].n
	})

	var b strings.Builder
	b.WriteString("# Auto-Generated Constitution Rules\n")
	b.WriteString("# Derived from agent experience\n\n")
	for i, c := range sorted {
		if i >= 20 {
			break
		}
		if c.n < 2 {
			continue
		}
		fmt.Fprintf(&b, "- %s (observed %dx)\n", c.text, c.n)
	}
	return b.String()
}

func (sp *SoulProfile) Summary() string {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return fmt.Sprintf("%d patterns (%d types), %d conventions logged",
		len(sp.patterns), len(sp.KnownTypes()), len(sp.conventionLog))
}

func (sp *SoulProfile) Save() error {
	sp.mu.Lock()
	data := struct {
		Patterns      map[string]*Pattern `json:"patterns"`
		Conventions   []string            `json:"conventions"`
	}{
		Patterns:    sp.patterns,
		Conventions: sp.conventionLog,
	}
	sp.mu.Unlock()

	dir := filepath.Dir(sp.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("profile mkdir: %w", err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("profile marshal: %w", err)
	}
	return os.WriteFile(sp.path, raw, 0644)
}

func (sp *SoulProfile) Load() error {
	data, err := os.ReadFile(sp.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("profile read: %w", err)
	}
	var loaded struct {
		Patterns    map[string]*Pattern `json:"patterns"`
		Conventions []string            `json:"conventions"`
	}
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("profile unmarshal: %w", err)
	}
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if loaded.Patterns != nil {
		sp.patterns = loaded.Patterns
	}
	if loaded.Conventions != nil {
		sp.conventionLog = loaded.Conventions
	}
	return nil
}
