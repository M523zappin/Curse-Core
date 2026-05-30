package sandbox

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStageAndApprove(t *testing.T) {
	tmpDir := t.TempDir()
	sa := New(filepath.Join(tmpDir, ".curse", "staging"), ModeDraftFile)
	sf, err := sa.Stage(filepath.Join(tmpDir, "out", "test.txt"), []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if sf.Status != StatusPending {
		t.Fatalf("expected Pending, got %v", sf.Status)
	}
	if sa.Count() != 1 {
		t.Fatalf("expected 1 staged file, got %d", sa.Count())
	}
	if err := sa.Approve(sf.ID); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(tmpDir, "out", "test.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected 'hello', got '%s'", string(data))
	}
	if sa.Count() != 0 {
		t.Fatalf("expected 0 staged files after approve, got %d", sa.Count())
	}
}

func TestStageAndReject(t *testing.T) {
	tmpDir := t.TempDir()
	sa := New(filepath.Join(tmpDir, ".curse", "staging"), ModeDraftFile)
	sf, err := sa.Stage(filepath.Join(tmpDir, "out", "secret.txt"), []byte("sensitive"))
	if err != nil {
		t.Fatal(err)
	}
	if err := sa.Reject(sf.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "out", "secret.txt")); !os.IsNotExist(err) {
		t.Fatal("expected target file to not exist after reject")
	}
}

func TestStageMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	sa := New(filepath.Join(tmpDir, ".curse", "staging"), ModeDraftFile)
	sa.Stage(filepath.Join(tmpDir, "a.txt"), []byte("aaa"))
	sa.Stage(filepath.Join(tmpDir, "b.txt"), []byte("bbb"))
	if sa.Count() != 2 {
		t.Fatalf("expected 2 staged, got %d", sa.Count())
	}
}
