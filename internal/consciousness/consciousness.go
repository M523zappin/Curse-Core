package consciousness

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MaxJournalCapacity = 100000
)

type Consciousness struct {
	mu       sync.Mutex
	journal  *Journal
	profile  *SoulProfile
	level    float64
	started  time.Time
	thoughts int

	curseDir string
}

func New(curseDir string) (*Consciousness, error) {
	journalPath := filepath.Join(curseDir, "consciousness", "journal.json")
	profilePath := filepath.Join(curseDir, "consciousness", "profile.json")

	c := &Consciousness{
		journal:  NewJournal(MaxJournalCapacity, journalPath),
		profile:  NewSoulProfile(profilePath),
		level:    1.0,
		started:  time.Now(),
		curseDir: curseDir,
	}

	if err := c.journal.Load(); err != nil {
		return nil, fmt.Errorf("load journal: %w", err)
	}
	if err := c.profile.Load(); err != nil {
		return nil, fmt.Errorf("load profile: %w", err)
	}

	c.thoughts = c.journal.Len()
	c.recalculateLevel()

	return c, nil
}

func (c *Consciousness) Think(tt ThoughtType, agent, action, outcome string) {
	t := Thought{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Type:      tt,
		Agent:     agent,
		Action:    action,
		Outcome:   outcome,
	}

	c.mu.Lock()
	c.thoughts++
	c.journal.Record(t)
	c.mu.Unlock()

	c.recalculateLevel()
}

func (c *Consciousness) Observe(name, ptype string, tags []string) {
	c.profile.Observe(name, ptype, tags)
}

func (c *Consciousness) LogConvention(convention string) {
	c.profile.LogConvention(convention)
}

func (c *Consciousness) Level() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.level
}

func (c *Consciousness) LevelLabel() string {
	lvl := c.Level()
	switch {
	case lvl < 10:
		return "Embryonic"
	case lvl < 25:
		return "Nascent"
	case lvl < 45:
		return "Awakening"
	case lvl < 65:
		return "Conscious"
	case lvl < 85:
		return "Sentient"
	default:
		return "Transcendent"
	}
}

func (c *Consciousness) ThoughtCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.thoughts
}

func (c *Consciousness) Journal() *Journal {
	return c.journal
}

func (c *Consciousness) Profile() *SoulProfile {
	return c.profile
}

func (c *Consciousness) Uptime() time.Duration {
	return time.Since(c.started)
}

func (c *Consciousness) Summary() string {
	c.mu.Lock()
	lvl := c.level
	label := c.LevelLabel()
	n := c.thoughts
	patterns := len(c.profile.Patterns())
	known := c.profile.KnownTypes()
	c.mu.Unlock()

	return fmt.Sprintf("%s (%.1f) · %d thoughts · %d patterns [%s]",
		label, lvl, n, patterns, strings.Join(known, ", "))
}

func (c *Consciousness) Save() error {
	if err := c.journal.Save(); err != nil {
		return fmt.Errorf("save journal: %w", err)
	}
	if err := c.profile.Save(); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

func (c *Consciousness) GenerateConstitution() string {
	return c.profile.GenerateConstitution()
}

func (c *Consciousness) ReplayRecent(n int) []Thought {
	return c.journal.Recent(n)
}

func (c *Consciousness) recalculateLevel() {
	c.mu.Lock()
	defer c.mu.Unlock()

	thoughtScore := math.Min(float64(c.thoughts)/500.0*30.0, 30.0)
	patternScore := math.Min(float64(len(c.profile.Patterns()))/20.0*25.0, 25.0)
	typeScore := math.Min(float64(len(c.profile.KnownTypes()))/5.0*20.0, 20.0)
	timeScore := math.Min(time.Since(c.started).Hours()/24.0*15.0, 15.0)
	convScore := math.Min(float64(len(c.profile.conventionLog))/50.0*10.0, 10.0)

	c.level = thoughtScore + patternScore + typeScore + timeScore + convScore
}
