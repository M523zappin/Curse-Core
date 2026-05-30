package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type EchoAdapter struct {
	profile gateway.ModelProfile
}

func NewEcho(profile gateway.ModelProfile) *EchoAdapter {
	return &EchoAdapter{profile: profile}
}

func (a *EchoAdapter) Name() string { return "echo" }
func (a *EchoAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *EchoAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("🔊 Echo Adapter — %s\n\n", time.Now().Format(time.RFC1123)))
	if req.System != "" {
		b.WriteString(fmt.Sprintf("System: %s\n\n", req.System))
	}
	for i, m := range req.Messages {
		b.WriteString(fmt.Sprintf("[%d] %s: %s\n", i, m.Role, m.Content))
	}
	if len(req.Tools) > 0 {
		b.WriteString(fmt.Sprintf("\nTools: %d available\n", len(req.Tools)))
	}
	b.WriteString(fmt.Sprintf("\nMaxTokens: %d\n", req.MaxTokens))
	if req.Model != "" {
		b.WriteString(fmt.Sprintf("Model: %s\n", req.Model))
	}
	return &gateway.Response{
		Message: gateway.Message{Role: gateway.RoleAssistant, Content: b.String()},
		Done:    true,
	}, nil
}
