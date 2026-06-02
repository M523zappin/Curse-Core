package dashboard

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IntentType represents the kind of natural language request detected.
type IntentType int

const (
	IntentUnknown IntentType = iota
	IntentRunTests
	IntentBuild
	IntentLint
	IntentSearch
	IntentGitStatus
	IntentGitDiff
	IntentGitLog
	IntentGitCommit
	IntentGitPush
	IntentListFiles
	IntentExplain
	IntentFix
	IntentCommitAll
)

// Intent holds the parsed natural language intent.
type Intent struct {
	Type    IntentType
	Command string // the shell command to execute
	Message string // what to tell the user
	IsExec  bool   // true = execute command, false = send to model
}

// parseIntent detects natural language developer requests and returns an Intent.
// Returns nil if the input should be passed to the model as-is.
func parseIntent(input string) *Intent {
	lower := strings.ToLower(strings.TrimSpace(input))

	// ── Test Commands ──
	if matchesAny(lower, []string{"run tests", "run the tests", "run test", "test this", "test it", "run all tests", "run unit tests", "execute tests", "run go tests"}) {
		cmd := "go test ./..."
		if runtime.GOOS == "windows" {
			cmd = "go test ./..."
		}
		return &Intent{Type: IntentRunTests, Command: cmd, Message: "Running tests...", IsExec: true}
	}
	if lower == "test" || lower == "run tests" {
		return &Intent{Type: IntentRunTests, Command: "go test ./...", Message: "Running tests...", IsExec: true}
	}

	// ── Build Commands ──
	if matchesAny(lower, []string{"build", "build this", "build the project", "compile", "compile this", "make build", "run build"}) {
		return &Intent{Type: IntentBuild, Command: "go build ./...", Message: "Building project...", IsExec: true}
	}

	// ── Lint Commands ──
	if matchesAny(lower, []string{"lint", "run lint", "run linter", "check for issues", "vet", "run vet", "go vet", "run golangci-lint", "check code quality"}) {
		return &Intent{Type: IntentLint, Command: "go vet ./...", Message: "Running linter...", IsExec: true}
	}
	if lower == "fmt" || lower == "format" || lower == "format code" || lower == "run fmt" || lower == "gofmt" {
		return &Intent{Type: IntentLint, Command: "gofmt -l .", Message: "Checking formatting...", IsExec: true}
	}

	// ── Git Commands ──
	if matchesAny(lower, []string{"git status", "status", "what's changed", "whats changed", "what changed", "repo status", "current status"}) {
		return &Intent{Type: IntentGitStatus, Command: "git status", Message: "Checking git status...", IsExec: true}
	}
	if matchesAny(lower, []string{"git diff", "diff", "show diff", "show changes", "what are the changes", "show me the changes", "what did i change"}) {
		return &Intent{Type: IntentGitDiff, Command: "git diff", Message: "Showing git diff...", IsExec: true}
	}
	if matchesAny(lower, []string{"git log", "log", "show log", "show commits", "recent commits", "commit history", "show git log", "what commits"}) {
		return &Intent{Type: IntentGitLog, Command: "git log --oneline -20", Message: "Showing recent commits...", IsExec: true}
	}
	if matchesAny(lower, []string{"git push", "push", "push to remote", "push changes", "push it"}) {
		return &Intent{Type: IntentGitPush, Command: "git push", Message: "Pushing to remote...", IsExec: true}
	}
	if matchesAny(lower, []string{"commit all", "commit everything", "git commit all", "commit all changes", "save everything", "commit -a"}) {
		return &Intent{Type: IntentCommitAll, Command: `git add -A && git commit -m "curse: auto-commit"`, Message: "Committing all changes...", IsExec: true}
	}
	if matchesAny(lower, []string{"git commit", "commit", "save changes", "make a commit"}) {
		return &Intent{Type: IntentGitCommit, Command: "git add -A && git commit -m \"curse: auto-commit\"", Message: "Committing changes...", IsExec: true}
	}
	if matchesAny(lower, []string{"git branch", "branches", "list branches", "show branches", "what branch"}) {
		return &Intent{Type: IntentGitStatus, Command: "git branch -a", Message: "Listing branches...", IsExec: true}
	}

	// ── File Operations ──
	if matchesAny(lower, []string{"list files", "show files", "what files", "what's in", "whats in", "show directory", "ls", "dir", "list directory", "show the project", "project structure", "show structure", "what's in this repo", "show me the project"}) {
		return &Intent{Type: IntentListFiles, Command: "", Message: "Listing project files...", IsExec: false}
	}
	if strings.HasPrefix(lower, "list files in ") || strings.HasPrefix(lower, "show files in ") || strings.HasPrefix(lower, "what's in ") || strings.HasPrefix(lower, "ls ") {
		dir := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(lower, "list files in "), "show files in "), "what's in "))
		dir = strings.TrimSpace(strings.TrimPrefix(dir, "ls "))
		if dir != "" {
			return &Intent{Type: IntentListFiles, Command: dir, Message: "Listing files...", IsExec: false}
		}
	}

	// ── Explain / Understand ──
	if matchesAny(lower, []string{"explain", "what does this do", "what is this", "help me understand", "describe", "walk me through"}) {
		return &Intent{Type: IntentExplain, Command: "", Message: "Analyzing...", IsExec: false}
	}
	if strings.HasPrefix(lower, "explain ") || strings.HasPrefix(lower, "what does ") {
		return &Intent{Type: IntentExplain, Command: "", Message: "Analyzing...", IsExec: false}
	}

	// ── Fix / Bug ──
	if matchesAny(lower, []string{"fix", "fix this", "fix the bug", "fix bug", "fix the issue", "fix error", "fix the error", "debug", "debug this", "find and fix"}) {
		return &Intent{Type: IntentFix, Command: "", Message: "Analyzing the issue...", IsExec: false}
	}
	if strings.HasPrefix(lower, "fix ") || strings.HasPrefix(lower, "debug ") {
		return &Intent{Type: IntentFix, Command: "", Message: "Analyzing...", IsExec: false}
	}

	// ── No match — send to model ──
	return nil
}

// matchesAny checks if input contains any of the given phrases.
// For short phrases (<=5 chars), requires whole-word match to avoid false positives.
func matchesAny(input string, phrases []string) bool {
	for _, phrase := range phrases {
		if input == phrase {
			return true
		}
		if len(phrase) <= 5 {
			if hasWord(input, phrase) {
				return true
			}
		} else if strings.Contains(input, phrase) {
			return true
		}
	}
	return false
}

// hasWord checks if the input contains a whole word match.
func hasWord(input, word string) bool {
	return input == word || strings.HasPrefix(input, word+" ") ||
		strings.HasSuffix(input, " "+word) ||
		strings.Contains(input, " "+word+" ")
}

// listProjectFiles returns a formatted listing of the project structure.
func listProjectFiles(repoPath string, subDir string) string {
	target := repoPath
	if subDir != "" && subDir != "." {
		target = filepath.Join(repoPath, subDir)
	}

	var files []string
	entries, err := os.ReadDir(target)
	if err != nil {
		return "Could not read directory: " + err.Error()
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" || name == "node_modules" || name == "vendor" || name == ".curse" {
			continue
		}
		if entry.IsDir() {
			files = append(files, name+"/")
		} else {
			files = append(files, name)
		}
	}

	if len(files) == 0 {
		return "(empty directory)"
	}
	return strings.Join(files, "  ")
}
