package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type LocalAIAdapter struct {
	profile gateway.ModelProfile
	client  *http.Client
}

func NewLocalAI(profile gateway.ModelProfile) *LocalAIAdapter {
	return &LocalAIAdapter{
		profile: profile,
		client: &http.Client{Timeout: 300 * time.Second},
	}
}

func (a *LocalAIAdapter) Name() string { return "localai" }

func (a *LocalAIAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *LocalAIAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
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
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
	}
	data, _ := json.Marshal(body)
	endpoint := a.profile.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:8080/v1/chat/completions"
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("localai request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("localai send: %w", err)
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
		return nil, fmt.Errorf("localai decode: %w (body: %s)", err, string(respBody[:min(len(respBody), 200)]))
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("localai: no choices returned")
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

func DefaultLocalAIEndpoint() string { return "http://localhost:8080" }

type LocalAIModel struct {
	ID string `json:"id"`
}

func DetectLocalAI(ctx context.Context) (string, bool) {
	base := DefaultLocalAIEndpoint()
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/v1/models", nil)
	if err != nil {
		return "", false
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	var result struct {
		Data []LocalAIModel `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return base, true
	}
	if len(result.Data) > 0 {
		return base, true
	}
	return base, true
}

func ListLocalAIModels(ctx context.Context) ([]string, error) {
	base := DefaultLocalAIEndpoint()
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
		Data []LocalAIModel `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	return models, nil
}
