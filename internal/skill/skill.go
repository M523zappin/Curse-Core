package skill

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Skill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Pattern     string   `json:"pattern"`
	Steps       []string `json:"steps"`
	Tags        []string `json:"tags"`
	Successes   int      `json:"successes"`
	Failures    int      `json:"failures"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Store struct {
	skillsDir string
	skills    map[string]*Skill
	mu        sync.RWMutex
}

func NewStore(skillsDir string) *Store {
	s := &Store{
		skillsDir: skillsDir,
		skills:    make(map[string]*Skill),
	}
	s.Load()
	return s
}

func (s *Store) Add(skill Skill) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if skill.ID == "" {
		skill.ID = fmt.Sprintf("sk-%x", sha256.Sum256([]byte(skill.Name+skill.Description)))[:16]
	}
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}
	s.skills[skill.ID] = &skill
	if err := s.persist(&skill); err != nil {
		s.skills[skill.ID] = nil
		delete(s.skills, skill.ID)
		return ""
	}
	return skill.ID
}

func (s *Store) Search(query string, limit int) []*Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	terms := strings.Fields(query)

	type scored struct {
		skill *Skill
		score float64
	}
	var scoredSkills []scored

	for _, sk := range s.skills {
		score := 0.0
		body := strings.ToLower(sk.Description + " " + sk.Pattern + " " + strings.Join(sk.Steps, " "))
		name := strings.ToLower(sk.Name)

		for _, term := range terms {
			if strings.Contains(name, term) {
				score += 5.0
			}
			if strings.Contains(body, term) {
				score += 2.0
			}
			if strings.Contains(strings.ToLower(sk.Pattern), term) {
				score += 3.0
			}
			for _, tag := range sk.Tags {
				if strings.Contains(strings.ToLower(tag), term) {
					score += 4.0
				}
			}
		}

		if score > 0 {
			scoredSkills = append(scoredSkills, scored{skill: sk, score: score})
		}
	}

	sort.Slice(scoredSkills, func(i, j int) bool {
		return scoredSkills[i].score > scoredSkills[j].score
	})

	if limit > 0 && len(scoredSkills) > limit {
		scoredSkills = scoredSkills[:limit]
	}

	out := make([]*Skill, len(scoredSkills))
	for i, ss := range scoredSkills {
		out[i] = ss.skill
	}
	return out
}

func (s *Store) RecordResult(skillID string, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sk, ok := s.skills[skillID]
	if !ok {
		return
	}
	if success {
		sk.Successes++
	} else {
		sk.Failures++
	}
	sk.UpdatedAt = time.Now()
	s.persist(sk)
}

func (s *Store) Generate(name, description, pattern string, steps []string, tags []string) *Skill {
	skill := Skill{
		Name:        name,
		Description: description,
		Pattern:     pattern,
		Steps:       steps,
		Tags:        tags,
	}
	s.Add(skill)
	return &skill
}

func (s *Store) All() []*Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Skill, 0, len(s.skills))
	for _, sk := range s.skills {
		out = append(out, sk)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.skills)
}

func (s *Store) BestScore() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.skills) == 0 {
		return 0
	}
	total := 0
	successes := 0
	for _, sk := range s.skills {
		total += sk.Successes + sk.Failures
		successes += sk.Successes
	}
	if total == 0 {
		return 0
	}
	return math.Round(float64(successes)/float64(total)*100) / 100
}

func (s *Store) Get(id string) *Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.skills[id]
}

func (s *Store) SaveDoc(skillID string, doc string) error {
	if s.skillsDir == "" {
		return nil
	}
	dir := filepath.Join(s.skillsDir, "docs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.md", skillID))
	return os.WriteFile(path, []byte(doc), 0644)
}

func (s *Store) persist(skill *Skill) error {
	if s.skillsDir == "" {
		return nil
	}
	if err := os.MkdirAll(s.skillsDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(s.skillsDir, fmt.Sprintf("%s.json", skill.ID))
	data, err := json.MarshalIndent(skill, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Store) Load() {
	if s.skillsDir == "" {
		return
	}
	files, err := os.ReadDir(s.skillsDir)
	if err != nil {
		return
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.skillsDir, f.Name()))
		if err != nil {
			continue
		}
		var skill Skill
		if json.Unmarshal(data, &skill) != nil {
			continue
		}
		s.skills[skill.ID] = &skill
	}
}
