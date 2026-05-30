package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Syncer struct {
	RepoURL    string
	RepoPath   string
	RemoteName string
	Branch     string
}

func New(repoURL, repoPath string) *Syncer {
	return &Syncer{
		RepoURL:    repoURL,
		RepoPath:   repoPath,
		RemoteName: "origin",
		Branch:     "master",
	}
}

func (s *Syncer) FetchConstitution() ([]byte, error) {
	return s.RemoteFile("CONSTITUTION.md")
}

func (s *Syncer) RemoteFile(path string) ([]byte, error) {
	cmd := exec.Command("git", "fetch", s.RemoteName, s.Branch)
	cmd.Dir = s.RepoPath
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}

	showCmd := exec.Command("git", "show", fmt.Sprintf("%s/%s:%s", s.RemoteName, s.Branch, path))
	showCmd.Dir = s.RepoPath
	out, err := showCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s: %w", path, err)
	}
	return out, nil
}

func HashContent(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (s *Syncer) SyncConstitution(localPath string) (changed bool, err error) {
	remote, err := s.FetchConstitution()
	if err != nil {
		return false, fmt.Errorf("fetch remote constitution: %w", err)
	}

	local, err := os.ReadFile(localPath)
	if err != nil {
		local = []byte{}
	}

	if HashContent(remote) == HashContent(local) {
		return false, nil
	}

	if err := os.WriteFile(localPath, remote, 0644); err != nil {
		return false, fmt.Errorf("write constitution: %w", err)
	}

	if err := os.WriteFile(localPath+".sha256", []byte(HashContent(remote)), 0644); err != nil {
		return false, fmt.Errorf("write constitution hash: %w", err)
	}

	return true, nil
}

func (s *Syncer) IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = s.RepoPath
	return cmd.Run() == nil
}

func (s *Syncer) HasRemote() bool {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = s.RepoPath
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), s.RemoteName)
}

func Clone(repoURL, destDir string) error {
	parent := filepath.Dir(destDir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("create parent: %w", err)
	}
	cmd := exec.Command("git", "clone", repoURL, destDir)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

func Pull(repoPath string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

func CommitAndPush(repoPath, message string) error {
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = repoPath
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoPath
	statusOut, _ := statusCmd.Output()
	if len(strings.TrimSpace(string(statusOut))) == 0 {
		return nil
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = repoPath
	if err := commitCmd.Run(); err != nil {
		return nil
	}

	pushCmd := exec.Command("git", "push")
	pushCmd.Dir = repoPath
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}
