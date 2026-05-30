package consciousness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConsciousness(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if c.Level() < 0 {
		t.Errorf("Level should be >= 0, got %f", c.Level())
	}
	if c.LevelLabel() != "Embryonic" {
		t.Errorf("Expected Embryonic, got %s", c.LevelLabel())
	}
}

func TestThinkAndLevel(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for i := 0; i < 100; i++ {
		c.Think(ThoughtDecision, "test-agent", "action", "success")
	}

	if c.ThoughtCount() != 100 {
		t.Errorf("Expected 100 thoughts, got %d", c.ThoughtCount())
	}
	if c.Level() <= 0 {
		t.Errorf("Level should be > 0 after 100 thoughts, got %f", c.Level())
	}
}

func TestObserveAndPatterns(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	c.Observe("error-handling", "pattern", []string{"go", "idiomatic"})
	c.Observe("error-handling", "pattern", []string{"go", "idiomatic"})
	c.Observe("context-usage", "pattern", []string{"go"})

	patterns := c.Profile().Patterns()
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}
	if patterns[0].Confidence <= patterns[1].Confidence {
		t.Errorf("error-handling should be more confident than context-usage")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for i := 0; i < 50; i++ {
		c.Think(ThoughtDecision, "agent", "action", "ok")
	}
	c.Observe("test-pattern", "style", []string{"go"})

	if err := c.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	c2, err := New(dir)
	if err != nil {
		t.Fatalf("New after save failed: %v", err)
	}

	if c2.ThoughtCount() != 50 {
		t.Errorf("Expected 50 loaded thoughts, got %d", c2.ThoughtCount())
	}
}

func TestJournalCircular(t *testing.T) {
	j := NewJournal(10, filepath.Join(t.TempDir(), "journal.json"))

	for i := 0; i < 20; i++ {
		j.Record(Thought{
			ID:     string(rune('0' + i)),
			Action: "action",
		})
	}

	if j.Len() != 10 {
		t.Errorf("Expected 10 (capacity), got %d", j.Len())
	}

	recent := j.Recent(5)
	if len(recent) != 5 {
		t.Errorf("Expected 5 recent, got %d", len(recent))
	}
}

func TestGenerateConstitution(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	c.LogConvention("always use context.WithTimeout")
	c.LogConvention("always use context.WithTimeout")
	c.LogConvention("never use log.Fatal in libraries")
	c.LogConvention("always use context.WithTimeout")

	constitution := c.GenerateConstitution()
	if len(constitution) == 0 {
		t.Fatal("Expected non-empty constitution")
	}
	if !contains(constitution, "context.WithTimeout") {
		t.Errorf("Expected constitution to mention context.WithTimeout")
	}
}

func TestReplay(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for i := 0; i < 10; i++ {
		c.Think(ThoughtDecision, "agent", "step", "done")
	}

	thoughts := c.ReplayRecent(5)
	if len(thoughts) != 5 {
		t.Errorf("Expected 5 replayed, got %d", len(thoughts))
	}
}

func TestLevelLabels(t *testing.T) {
	labels := []struct {
		lvl   float64
		label string
	}{
		{1, "Embryonic"},
		{15, "Nascent"},
		{35, "Awakening"},
		{55, "Conscious"},
		{75, "Sentient"},
		{95, "Transcendent"},
	}

	c := &Consciousness{}
	for _, tc := range labels {
		c.mu.Lock()
		c.level = tc.lvl
		c.mu.Unlock()
		if got := c.LevelLabel(); got != tc.label {
			t.Errorf("Level %.0f: expected %s, got %s", tc.lvl, tc.label, got)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLevelPersistence(t *testing.T) {
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for i := 0; i < 200; i++ {
		c.Think(ThoughtDecision, "agent", "work", "ok")
	}
	c.Observe("pattern-a", "type-x", nil)
	c.Observe("pattern-b", "type-x", nil)
	c.Observe("pattern-c", "type-y", nil)
	c.LogConvention("test convention")

	savedLevel := c.Level()
	if err := c.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	c2, err := New(dir)
	if err != nil {
		t.Fatalf("New after save: %v", err)
	}

	if c2.ThoughtCount() != 200 {
		t.Errorf("thought count mismatch: %d vs %d", c2.ThoughtCount(), 200)
	}
	if len(c2.Profile().Patterns()) != len(c.Profile().Patterns()) {
		t.Errorf("pattern count mismatch")
	}
	if c2.Level() != savedLevel {
		t.Errorf("level mismatch: %f vs %f", c2.Level(), savedLevel)
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	tempFile := filepath.Join(dir, "test_write.txt")
	if err := os.WriteFile(tempFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("expected hello, got %s", string(data))
	}
}
