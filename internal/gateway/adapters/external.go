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

type ExternalAdapter struct {
	profile gateway.ModelProfile
	client  *http.Client
}

func NewExternal(profile gateway.ModelProfile) *ExternalAdapter {
	return &ExternalAdapter{
		profile: profile,
		client:  &http.Client{Timeout: 300 * time.Second},
	}
}

func (a *ExternalAdapter) Name() string { return "external-api" }

func (a *ExternalAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *ExternalAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	model := req.Model
	if model == "" {
		model = a.profile.Model
	}
	msgs := make([]map[string]interface{}, 0)
	if req.System != "" {
		msgs = append(msgs, map[string]interface{}{"role": "system", "content": req.System})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]interface{}{"role": string(m.Role), "content": m.Content})
	}
	body := map[string]interface{}{
		"model":       model,
		"messages":    msgs,
		"max_tokens":  req.MaxTokens,
		"temperature": a.profile.Temperature,
	}
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
	}
	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.profile.Endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("external request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.profile.Auth.Header != "" && a.profile.Auth.EnvVar != "" {
		httpReq.Header.Set(a.profile.Auth.Header, "Bearer "+a.profile.Auth.EnvVar)
	}
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("external send: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *gateway.Usage `json:"usage,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("external decode: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("external: no choices returned")
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
