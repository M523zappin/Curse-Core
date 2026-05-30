package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)

	if n := s.Count(); n != 0 {
		t.Fatalf("expected 0 skills, got %d", n)
	}

	id := s.Add(Skill{
		Name:        "test-skill",
		Description: "a test skill for refactoring",
		Pattern:     "refactor",
		Steps:       []string{"analyze", "execute", "review"},
		Tags:        []string{"refactoring", "test"},
	})
	if id == "" {
		t.Fatal("expected non-empty skill ID")
	}

	if n := s.Count(); n != 1 {
		t.Fatalf("expected 1 skill, got %d", n)
	}

	sk := s.Get(id)
	if sk == nil {
		t.Fatal("expected to find skill by ID")
	}
	if sk.Name != "test-skill" {
		t.Fatalf("expected 'test-skill', got %s", sk.Name)
	}

	s.RecordResult(id, true)
	s.RecordResult(id, true)
	s.RecordResult(id, false)

	sk = s.Get(id)
	if sk.Successes != 2 {
		t.Fatalf("expected 2 successes, got %d", sk.Successes)
	}
	if sk.Failures != 1 {
		t.Fatalf("expected 1 failure, got %d", sk.Failures)
	}
}

func TestSearch(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-search")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	s.Add(Skill{Name: "refactor-go", Description: "refactor Go code", Pattern: "refactor", Tags: []string{"go"}})
	s.Add(Skill{Name: "test-rust", Description: "test Rust code", Pattern: "testing", Tags: []string{"rust"}})
	s.Add(Skill{Name: "deploy-docker", Description: "deploy docker containers", Pattern: "deploy", Tags: []string{"docker", "devops"}})

	results := s.Search("refactor go", 0)
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'refactor go', got %d", len(results))
	}
	if results[0].Name != "refactor-go" {
		t.Fatalf("expected 'refactor-go' as top result, got %s", results[0].Name)
	}

	results = s.Search("rust", 0)
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'rust', got %d", len(results))
	}

	results = s.Search("deploy docker devops", 0)
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'deploy docker devops', got %d", len(results))
	}

	results = s.Search("nonexistent", 0)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestGenerate(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-gen")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	sk := s.Generate("auto-fix", "automatically fix compilation errors", "compile error", []string{"detect error", "apply fix", "verify"}, []string{"auto-generated", "fix"})
	if sk == nil {
		t.Fatal("expected non-nil skill")
	}
	if s.Count() != 1 {
		t.Fatalf("expected 1 skill, got %d", s.Count())
	}
}

func TestPersistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-persist")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s1 := NewStore(dir)
	s1.Add(Skill{Name: "persist-test", Description: "test persistence", Pattern: "persist", Tags: []string{"test"}})
	if s1.Count() != 1 {
		t.Fatalf("expected 1 skill in s1, got %d", s1.Count())
	}

	s2 := NewStore(dir)
	if s2.Count() != 1 {
		t.Fatalf("expected 1 skill in s2 (loaded from disk), got %d", s2.Count())
	}
}

func TestBestScore(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-score")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	if score := s.BestScore(); score != 0 {
		t.Fatalf("expected 0 score for empty store, got %f", score)
	}

	id := s.Add(Skill{Name: "scored", Description: "scored skill", Pattern: "score"})
	s.RecordResult(id, true)
	s.RecordResult(id, true)
	s.RecordResult(id, false)

	if score := s.BestScore(); score != 0.67 {
		t.Fatalf("expected 0.67, got %f", score)
	}
}

func TestAll(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-all")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s := NewStore(dir)
	s.Add(Skill{Name: "z-skill", Description: "last"})
	s.Add(Skill{Name: "a-skill", Description: "first"})

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(all))
	}
}

func TestEmptyDir(t *testing.T) {
	s := NewStore("")
	if s.Count() != 0 {
		t.Fatalf("expected 0 skills with empty dir, got %d", s.Count())
	}
	id := s.Add(Skill{Name: "no-disk", Description: "should work without persistence"})
	if id == "" {
		t.Fatal("expected non-empty ID even without persistence")
	}
}

func TestLoadInvalidFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "skill-invalid")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "invalid.json"), []byte("{not json"), 0644)
	os.WriteFile(filepath.Join(dir, "not-a-skill.txt"), []byte("hello"), 0644)

	s := NewStore(dir)
	if s.Count() != 0 {
		t.Fatalf("expected 0 skills with invalid files, got %d", s.Count())
	}
}
