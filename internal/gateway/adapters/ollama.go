package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type OllamaModel struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}

type OllamaAdapter struct {
	profile gateway.ModelProfile
	client  *http.Client
}

func NewOllama(profile gateway.ModelProfile) *OllamaAdapter {
	return &OllamaAdapter{
		profile: profile,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *OllamaAdapter) Name() string { return "ollama" }

func (a *OllamaAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func DefaultOllamaEndpoint() string { return "http://localhost:11434" }

func (a *OllamaAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	endpoint := a.profile.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434/api/generate"
	}
	body := map[string]interface{}{
		"model":    a.profile.Model,
		"prompt":   buildPrompt(req),
		"stream":   false,
		"options": map[string]interface{}{
			"temperature": a.profile.Temperature,
			"num_predict": req.MaxTokens,
		},
	}
	if req.System != "" {
		body["system"] = req.System
	}
	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama send: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama decode: %w", err)
	}
	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: result.Response},
		Done:    result.Done,
	}, nil
}

func buildPrompt(req *gateway.Prompt) string {
	var b bytes.Buffer
	for _, m := range req.Messages {
		b.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}
	return b.String()
}

func DetectOllama(ctx context.Context) (string, bool) {
	base := DefaultOllamaEndpoint()
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/api/tags", nil)
	if err != nil {
		return "", false
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	return base, resp.StatusCode == http.StatusOK
}

func ListOllamaModels(ctx context.Context) ([]OllamaModel, error) {
	base := DefaultOllamaEndpoint()
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("ollama list request: %w", err)
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama list: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Models []OllamaModel `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama decode tags: %w", err)
	}
	return result.Models, nil
}
