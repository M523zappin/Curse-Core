package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type LocalToolInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type"`
}

func DetectLocalTools(ctx context.Context) []LocalToolInfo {
	tools := make([]LocalToolInfo, 0)

	checks := []struct {
		name     string
		checkCmd string
		version  string
		toolType string
	}{
		{"python3", "python3 --version", "--version", "runtime"},
		{"python", "python --version", "--version", "runtime"},
		{"ollama", "ollama --version", "--version", "llm-server"},
		{"llama-cli", "llama-cli --version", "--version", "llm-cli"},
		{"llama-server", "llama-server --version", "--version", "llm-server"},
		{"node", "node --version", "--version", "runtime"},
	}

	for _, c := range checks {
		path := findExec(c.checkCmd)
		if path == "" {
			parts := strings.Fields(c.checkCmd)
			path = findExec(parts[0])
		}
		if path != "" {
			info := LocalToolInfo{
				Name: c.name,
				Path: path,
				Type: c.toolType,
			}
			if c.version != "" {
				parts := strings.Fields(c.checkCmd)
				cmd := exec.CommandContext(ctx, parts[0], c.version)
				if out, err := cmd.Output(); err == nil {
					info.Version = strings.TrimSpace(string(out))
				}
			}
			tools = append(tools, info)
		}
	}

	if isOllamaRunning(ctx) {
		tools = append(tools, LocalToolInfo{
			Name: "ollama-server",
			Path: "http://localhost:11434",
			Type: "llm-server",
		})
	}

	if isLlamaServerRunning(ctx) {
		tools = append(tools, LocalToolInfo{
			Name: "llama-server",
			Path: "http://localhost:8080",
			Type: "llm-server",
		})
	}

	return tools
}

func findExec(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func isOllamaRunning(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/tags", nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func isLlamaServerRunning(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/health", nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func AutoDetectLocalModels(ctx context.Context) (map[string]ModelProfile, string, []LocalToolInfo) {
	profiles := make(map[string]ModelProfile)
	tools := DetectLocalTools(ctx)
	var firstKey string

	repoPath, _ := os.Getwd()

	// ═══════════════════════════════════════════════════
	// Tier 1 — Built-in zero-dependency adapters
	// ═══════════════════════════════════════════════════

	profiles["codex"] = ModelProfile{
		Provider:      "codex",
		Model:         "codex",
		Endpoint:      "builtin://ast-analysis",
		ContextWindow: 32768,
		MaxTokens:     4096,
		Temperature:   0.0,
	}
	firstKey = "codex"

	profiles["grep"] = ModelProfile{
		Provider:      "grep",
		Model:         "grep",
		Endpoint:      "builtin://code-search",
		ContextWindow: 8192,
		MaxTokens:     4096,
		Temperature:   0.0,
	}

	profiles["echo"] = ModelProfile{
		Provider:      "echo",
		Model:         "echo",
		Endpoint:      "builtin://debug",
		ContextWindow: 4096,
		MaxTokens:    2048,
		Temperature:   0.5,
	}

	profiles["eval"] = ModelProfile{
		Provider:      "eval",
		Model:         "eval",
		Endpoint:      "builtin://math",
		ContextWindow: 1024,
		MaxTokens:    512,
		Temperature:   0.0,
	}

	profiles["fortune"] = ModelProfile{
		Provider:      "fortune",
		Model:         "fortune",
		Endpoint:      "builtin://fun",
		ContextWindow: 2048,
		MaxTokens:    1024,
		Temperature:   0.8,
	}

	profiles["system"] = ModelProfile{
		Provider:      "system",
		Model:         "system",
		Endpoint:      "builtin://system-info",
		ContextWindow: 2048,
		MaxTokens:    1024,
		Temperature:   0.0,
	}

	profiles["fallback"] = ModelProfile{
		Provider:      "local-fallback",
		Model:         "fallback",
		Endpoint:      "builtin://local",
		ContextWindow: 4096,
		MaxTokens:    1024,
		Temperature:   0.5,
	}

	// ═══════════════════════════════════════════════════
	// Tier 2 — Python subprocess (if available)
	// ═══════════════════════════════════════════════════

	pythonPath := findExec("python3")
	if pythonPath == "" {
		pythonPath = findExec("python")
	}
	if pythonPath != "" {
		profiles["python-helper"] = ModelProfile{
			Provider:      "subprocess",
			Model:         pythonPath + ` -c "import sys; print(sys.stdin.read())"`,
			Endpoint:      "subprocess://stdin",
			ContextWindow: 8192,
			MaxTokens:     2048,
			Temperature:   0.3,
		}

		profiles["python-repl"] = ModelProfile{
			Provider:      "subprocess",
			Model:         pythonPath + ` -c "import sys; exec(sys.stdin.read())"`,
			Endpoint:      "subprocess://stdin",
			ContextWindow: 4096,
			MaxTokens:     1024,
			Temperature:   0.0,
		}
	}

	// ═══════════════════════════════════════════════════
	// Tier 3 — Unsloth (local LLM via Python)
	// ═══════════════════════════════════════════════════

	if unslothAvail, unslothVer := isUnslothAvailable(ctx); unslothAvail {
		profiles["unsloth-fast"] = ModelProfile{
			Provider:      "unsloth",
			Model:         "unsloth/Llama-3.2-1B-Instruct",
			Endpoint:      "python://unsloth",
			ContextWindow: 8192,
			MaxTokens:     2048,
			Temperature:   0.3,
		}
		profiles["unsloth-powerful"] = ModelProfile{
			Provider:      "unsloth",
			Model:         "unsloth/Mistral-7B-Instruct-v0.3",
			Endpoint:      "python://unsloth",
			ContextWindow: 32768,
			MaxTokens:     4096,
			Temperature:   0.2,
		}
		if firstKey != "" && strings.HasPrefix(firstKey, "codex") {
			firstKey = "unsloth-fast"
		}

		for _, m := range listUnslothModelNames() {
			name := "us-" + shortModelName(m)
			if _, exists := profiles[name]; !exists {
				profiles[name] = ModelProfile{
					Provider:      "unsloth",
					Model:         m,
					Endpoint:      "python://unsloth",
					ContextWindow: 8192,
					MaxTokens:     2048,
					Temperature:   0.3,
				}
			}
		}

		_ = unslothVer
	}

	// ═══════════════════════════════════════════════════
	// Tier 4 — Ollama (local LLM server)
	// ═══════════════════════════════════════════════════

	if isOllamaRunning(ctx) {
		ollamaModels, err := listOllamaModels(ctx)
		if err == nil && len(ollamaModels) > 0 {
			for _, m := range ollamaModels {
				name := "ollama-" + m
				profiles[name] = ModelProfile{
					Provider:      "ollama",
					Model:         m,
					Endpoint:      "http://localhost:11434/api/generate",
					ContextWindow: 8192,
					MaxTokens:     4096,
					Temperature:   0.3,
				}
				if firstKey != "" && strings.HasPrefix(firstKey, "codex") {
					firstKey = name
				}
			}
		}
	}

	// ═══════════════════════════════════════════════════
	// Tier 4 — llama.cpp server (local)
	// ═══════════════════════════════════════════════════

	if isLlamaServerRunning(ctx) {
		profiles["llama-server"] = ModelProfile{
			Provider:      "openai-compatible",
			Model:         "local",
			Endpoint:      "http://localhost:8080/v1/chat/completions",
			ContextWindow: 4096,
			MaxTokens:     2048,
			Temperature:   0.3,
		}
	}

	// ═══════════════════════════════════════════════════
	// Detect codebase language & add specialized profiles
	// ═══════════════════════════════════════════════════

	if hasGoFiles(repoPath) {
		profiles["codex-go"] = ModelProfile{
			Provider:      "codex",
			Model:         "codex",
			Endpoint:      "builtin://ast-analysis",
			ContextWindow: 65536,
			MaxTokens:     8192,
			Temperature:   0.1,
		}
	}

	if hasPythonFiles(repoPath) {
		profiles["grep-python"] = ModelProfile{
			Provider:      "grep",
			Model:         "grep",
			Endpoint:      "builtin://code-search",
			ContextWindow: 16384,
			MaxTokens:     4096,
			Temperature:   0.0,
		}
	}

	return profiles, firstKey, tools
}

func hasGoFiles(dir string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return filepath.SkipAll
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			if !strings.Contains(path, "vendor") && !strings.Contains(path, ".git") {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func hasPythonFiles(dir string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return filepath.SkipAll
		}
		if !info.IsDir() && strings.HasSuffix(path, ".py") {
			if !strings.Contains(path, ".git") {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})
	return found
}

func isUnslothAvailable(ctx context.Context) (bool, string) {
	python := findExec("python3")
	if python == "" {
		python = findExec("python")
	}
	if python == "" {
		return false, ""
	}
	cmd := exec.CommandContext(ctx, python, "-c", "import unsloth; print(unsloth.__version__)")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(out.String())
}

func shortModelName(full string) string {
	parts := strings.Split(full, "/")
	last := parts[len(parts)-1]
	last = strings.NewReplacer("-Instruct", "", "-instruct", "", ".", "-").Replace(last)
	return strings.ToLower(last)
}

func listUnslothModelNames() []string {
	return []string{
		"unsloth/Llama-3.2-1B-Instruct",
		"unsloth/Llama-3.2-3B-Instruct",
		"unsloth/Mistral-7B-Instruct-v0.3",
		"unsloth/Qwen2.5-1.5B-Instruct",
		"unsloth/Qwen2.5-7B-Instruct",
		"unsloth/gemma-2-2b-it",
		"unsloth/Phi-3.5-mini-instruct",
	}
}

func listOllamaModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/tags", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		names = append(names, m.Name)
	}
	return names, nil
}

func GenerateDefaultLocalRegistry(ctx context.Context) *ModelRegistry {
	profiles, firstKey, _ := AutoDetectLocalModels(ctx)
	reg := &ModelRegistry{
		Profiles:         profiles,
		Active:           firstKey,
		SelectionStrategy: "manual",
	}
	return reg
}
