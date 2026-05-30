package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	ctx := context.Background()
	if err := gw.Init(ctx); err != nil {
		log.Fatalf("gateway init: %v", err)
	}
	fmt.Printf("Curse Gateway running\n")
	fmt.Printf("  State:    %s\n", gw.Machine().State().String())
	fmt.Printf("  Model:    %s\n", gw.ActiveModel())
	fmt.Printf("  Curse:    %s\n", curseDir)
	fmt.Printf("  Config:   %s\n", configDir)
	select {}
}
