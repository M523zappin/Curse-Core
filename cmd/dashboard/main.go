package main

import (
	"context"
	"flag"
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
	resume := flag.Bool("resume", false, "resume previous session state")
	flag.Parse()

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

	repoPath, _ := os.Getwd()

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
	gw.RegisterAdapter("codex", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewCodex(p, repoPath)
	})
	gw.RegisterAdapter("grep", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewGrep(p, repoPath)
	})
	gw.RegisterAdapter("echo", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewEcho(p)
	})
	gw.RegisterAdapter("eval", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewEval(p)
	})
	gw.RegisterAdapter("fortune", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewFortune(p)
	})
	gw.RegisterAdapter("system", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewSystem(p)
	})
	gw.RegisterAdapter("local-fallback", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewLocalFallback(p)
	})
	gw.RegisterAdapter("unsloth", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewUnsloth(p, curseDir)
	})
	gw.RegisterAdapter("subprocess", func(p gateway.ModelProfile) gateway.Adapter {
		return adapters.NewSubprocess(p)
	})

	gw.SetSyncOnInit(true)

	if err := gw.Init(context.Background()); err != nil {
		log.Fatalf("gateway init: %v", err)
	}

	if *resume {
		state, err := gw.RestoreSession()
		if err != nil {
			log.Printf("warning: session resume failed: %v", err)
		} else if state != nil {
			log.Printf("session restored: %s (model: %s, step: %d)",
				state.SessionID, state.ActiveModel, state.MachineStep)
		}
	}

	gw.StartEngine(context.Background())
	gw.StartScheduler(context.Background())

	model := dashboard.NewModel(gw)
	model.SetLogPaths(logPath, cpPath)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dashboard error: %v\n", err)
		os.Exit(1)
	}

	gw.Shutdown(context.Background())
	log.Println("session saved, shutdown complete")
}
