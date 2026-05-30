package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type LlamaCppAdapter struct {
	profile gateway.ModelProfile
	client  *http.Client
}

func NewLlamaCpp(profile gateway.ModelProfile) *LlamaCppAdapter {
	return &LlamaCppAdapter{
		profile: profile,
		client:  &http.Client{Timeout: 300 * time.Second},
	}
}

func (a *LlamaCppAdapter) Name() string { return "llamacpp" }

func (a *LlamaCppAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *LlamaCppAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	if strings.Contains(a.profile.Endpoint, "/v1/chat/completions") {
		return a.sendOpenAI(ctx, req)
	}
	return a.sendNative(ctx, req)
}

func (a *LlamaCppAdapter) sendOpenAI(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	msgs := make([]map[string]interface{}, 0)
	if req.System != "" {
		msgs = append(msgs, map[string]interface{}{"role": "system", "content": req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]interface{}{"role": string(m.Role), "content": m.Content})
	}
	body := map[string]interface{}{
		"model":       a.profile.Model,
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": a.profile.Temperature,
		"stream":      false,
	}
	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.profile.Endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("llamacpp request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llamacpp send: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *gateway.Usage `json:"usage,omitempty"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("llamacpp decode: %w (body: %s)", err, string(respBody[:min(len(respBody), 200)]))
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("llamacpp: no choices")
	}
	msg := result.Choices[0].Message
	usage := gateway.Usage{}
	if result.Usage != nil {
		usage = *result.Usage
	}
	return &gateway.Response{
		Message: gateway.Message{Role: gateway.Role(msg.Role), Content: msg.Content},
		Usage:   usage,
		Done:    true,
	}, nil
}

func (a *LlamaCppAdapter) sendNative(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	prompt := buildPrompt(req)
	if req.System != "" {
		prompt = fmt.Sprintf("[INST] %s\n%s [/INST]", req.System, prompt)
	}
	body := map[string]interface{}{
		"prompt":            prompt,
		"n_predict":         req.MaxTokens,
		"temperature":       a.profile.Temperature,
		"stream":            false,
		"cache_prompt":      true,
	}
	if req.MaxTokens <= 0 {
		body["n_predict"] = 2048
	}
	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.profile.Endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("llamacpp native request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llamacpp native send: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	type nativeResponse struct {
		Content string `json:"content"`
		Tokens  int    `json:"tokens_predicted"`
	}
	var result nativeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("llamacpp native decode: %w (body: %s)", err, string(respBody[:min(len(respBody), 200)]))
	}
	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: result.Content},
		Usage:   gateway.Usage{CompletionTokens: result.Tokens},
		Done:    true,
	}, nil
}

func DefaultLlamaCppEndpoint() string { return "http://localhost:8080" }

func LlamaCppNativeEndpoint() string { return DefaultLlamaCppEndpoint() + "/completion" }

func LlamaCppOpenAIEndpoint() string { return DefaultLlamaCppEndpoint() + "/v1/chat/completions" }

type LlamaCppModel struct {
	ID     string `json:"id"`
	Object string `json:"object"`
}

func DetectLlamaCpp(ctx context.Context) (string, bool) {
	base := DefaultLlamaCppEndpoint()
	client := &http.Client{Timeout: 3 * time.Second}
	if req, err := http.NewRequestWithContext(ctx, "GET", base+"/health", nil); err == nil {
		if resp, err := client.Do(req); err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return base, true
			}
		}
	}
	return "", false
}

func ListLlamaCppModels(ctx context.Context) ([]string, error) {
	base := DefaultLlamaCppEndpoint()
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", base+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Data []LlamaCppModel `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []string{"default"}, nil
	}
	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	if len(models) == 0 {
		models = append(models, "default")
	}
	return models, nil
}
