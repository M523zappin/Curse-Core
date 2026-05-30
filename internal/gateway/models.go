package gateway

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AuthConfig struct {
	Header string `json:"header,omitempty"`
	EnvVar string `json:"env_var,omitempty"`
}

type ModelProfile struct {
	Provider      string     `json:"provider"`
	Model         string     `json:"model"`
	ContextWindow int        `json:"context_window"`
	MaxTokens     int        `json:"max_tokens"`
	Temperature   float64    `json:"temperature"`
	Endpoint      string     `json:"endpoint"`
	Auth          AuthConfig `json:"auth"`
}

type ModelRegistry struct {
	Profiles         map[string]ModelProfile `json:"profiles"`
	Active           string                  `json:"active"`
	SelectionStrategy string                  `json:"selection_strategy"`
}

func LoadModels(path string) (*ModelRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read models config: %w", err)
	}
	var reg ModelRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse models config: %w", err)
	}
	reg.resolveEnvVars()
	return &reg, nil
}

func DefaultModelsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "curse", "models.json"), nil
}

func (reg *ModelRegistry) resolveEnvVars() {
	for name, profile := range reg.Profiles {
		profile.Endpoint = resolveEnv(profile.Endpoint)
		reg.Profiles[name] = profile
	}
}

func (reg *ModelRegistry) GetProfile(name string) (ModelProfile, bool) {
	p, ok := reg.Profiles[name]
	return p, ok
}

func (reg *ModelRegistry) ActiveProfile() (ModelProfile, bool) {
	return reg.GetProfile(reg.Active)
}

func (reg *ModelRegistry) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir models dir: %w", err)
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write models: %w", err)
	}
	return nil
}

func resolveEnv(s string) string {
	for strings.Contains(s, "${") {
		start := strings.Index(s, "${")
		rest := s[start+2:]
		end := strings.Index(rest, "}")
		if end == -1 {
			break
		}
		end += start + 2
		varName := s[start+2 : end]
		val := os.Getenv(varName)
		if val == "" {
			s = s[:start] + s[end+1:]
		} else {
			s = s[:start] + val + s[end+1:]
		}
	}
	return s
}
