package adapters

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type LocalFallbackAdapter struct {
	profile gateway.ModelProfile
}

func NewLocalFallback(profile gateway.ModelProfile) *LocalFallbackAdapter {
	return &LocalFallbackAdapter{profile: profile}
}

func (a *LocalFallbackAdapter) Name() string { return "local-fallback" }

func (a *LocalFallbackAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *LocalFallbackAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	msg := fmt.Sprintf(`[CURSE Fallback Adapter — %s/%s]

No upstream AI model is currently configured. CURSE is running in fallback mode.

To connect an AI model:

  1. Install Ollama (https://ollama.ai) and pull a model:
     $ ollama pull codellama:7b

  2. Or configure an API provider in ~/.config/curse/models.json

  3. Press Ctrl+M in the dashboard to switch models.

  %s
  Received %d message(s) with system prompt: %s`,
		runtime.GOOS, runtime.GOARCH,
		time.Now().Format("15:04:05 UTC"),
		len(req.Messages),
		truncateStr(req.System, 60))

	return &gateway.Response{
		Message: gateway.Message{
			Role:    gateway.RoleAssistant,
			Content: msg,
		},
		Done: true,
	}, nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
