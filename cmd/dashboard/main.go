package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/M523zappin/Curse-Core/internal/dashboard"
	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/M523zappin/Curse-Core/internal/gateway/adapters"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("home dir: %v", err)
	}

	curseDir := filepath.Join(home, ".curse")
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("config dir: %v", err)
	}
	configDir = filepath.Join(configDir, "curse")

	logPath := filepath.Join(curseDir, "logs", "event.log")
	cpPath := filepath.Join(curseDir, "logs", "session.json")

	gw := gateway.New(curseDir, configDir)
	gw.RegisterAdapter("ollama", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewOllama(p)
	})
	gw.RegisterAdapter("openai-compatible", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewExternal(p)
	})
	gw.RegisterAdapter("mcp", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewMCP(p)
	})

	if err := gw.Init(context.Background()); err != nil {
		log.Fatalf("gateway init: %v", err)
	}

	model := dashboard.NewModel(gw)
	model.SetLogPaths(logPath, cpPath)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dashboard error: %v\n", err)
		os.Exit(1)
	}
}
