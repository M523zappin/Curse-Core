package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type GrepAdapter struct {
	profile  gateway.ModelProfile
	repoPath string
}

func NewGrep(profile gateway.ModelProfile, repoPath string) *GrepAdapter {
	return &GrepAdapter{profile: profile, repoPath: repoPath}
}

func (a *GrepAdapter) Name() string { return "grep" }
func (a *GrepAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *GrepAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	q := ""
	for _, m := range req.Messages {
		if m.Role == gateway.RoleUser {
			q = m.Content
			break
		}
	}
	if q == "" {
		q = req.System
	}

	pattern := extractSearchPattern(q)
	if pattern == "" {
		return &gateway.Response{
			Message: gateway.Message{Role: gateway.RoleAssistant, Content: "Send a search query to grep the codebase."},
			Done:    true,
		}, nil
	}

	var results []string
	filepath.Walk(a.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".go", ".ts", ".js", ".py", ".rs", ".md", ".json", ".yaml", ".yml", ".toml", ".html", ".css":
		default:
			return nil
		}
		if strings.Contains(path, "vendor") || strings.Contains(path, ".git") || strings.Contains(path, "node_modules") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
				rel, _ := filepath.Rel(a.repoPath, path)
				results = append(results, fmt.Sprintf("%s:%d: %s", rel, i+1, strings.TrimSpace(line)))
				if len(results) >= 30 {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if len(results) == 0 {
		return &gateway.Response{
			Message: gateway.Message{Role: gateway.RoleAssistant, Content: fmt.Sprintf("No matches for %q in codebase.", pattern)},
			Done:    true,
		}, nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🔍 grep %q — %d result(s):\n\n", pattern, len(results)))
	for _, r := range results {
		b.WriteString("  " + r + "\n")
	}
	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: b.String()},
		Done:    true,
	}, nil
}

func extractSearchPattern(q string) string {
	clean := strings.TrimSpace(q)
	clean = strings.TrimPrefix(clean, "grep ")
	clean = strings.TrimPrefix(clean, "search ")
	clean = strings.TrimPrefix(clean, "find ")
	clean = strings.Trim(clean, `"'`)
	return clean
}
