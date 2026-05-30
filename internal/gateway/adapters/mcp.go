package adapters

import (
	"context"
	"fmt"

	"github.com/M523zappin/Curse-Core/internal/gateway"
)

type MCPAdapter struct {
	profile gateway.ModelProfile
}

func NewMCP(profile gateway.ModelProfile) *MCPAdapter {
	return &MCPAdapter{profile: profile}
}

func (a *MCPAdapter) Name() string { return "mcp" }

func (a *MCPAdapter) ModelInfo() gateway.ModelProfile { return a.profile }

func (a *MCPAdapter) Send(ctx context.Context, req *gateway.Prompt) (*gateway.Response, error) {
	return nil, fmt.Errorf("MCP adapter: not yet implemented")
}
