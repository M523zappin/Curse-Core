package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/M523zappin/Curse-Core/internal/sync"
)

const defaultRepo = "https://github.com/M523zappin/Curse-Core.git"

func main() {
	repoURL := flag.String("repo", defaultRepo, "GitHub repository URL")
	dest := flag.String("dest", "", "Destination directory (default: ./curse)")
	flag.Parse()

	destDir := *dest
	if destDir == "" {
		cwd, _ := os.Getwd()
		destDir = filepath.Join(cwd, "curse")
	}

	fmt.Println("╭──────────────────────────────────────╮")
	fmt.Println("│  Curse-Core — Autonomous Mission     │")
	fmt.Println("│  Control Platform Initialization     │")
	fmt.Println("╰──────────────────────────────────────╯")
	fmt.Println()

	if _, err := os.Stat(destDir); err == nil {
		fmt.Printf("✓ Already initialized at %s\n", destDir)
	} else {
		fmt.Printf("Cloning %s → %s ...\n", *repoURL, destDir)
		if err := sync.Clone(*repoURL, destDir); err != nil {
			log.Fatalf("Clone failed: %v", err)
		}
		fmt.Println("✓ Repository cloned")
	}

	envPath := filepath.Join(destDir, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		examplePath := filepath.Join(destDir, ".env.example")
		if data, err := os.ReadFile(examplePath); err == nil {
			if err := os.WriteFile(envPath, data, 0600); err != nil {
				log.Fatalf("Create .env: %v", err)
			}
			fmt.Println("✓ .env created from .env.example (edit with your keys)")
		}
	} else {
		fmt.Println("✓ .env already exists")
	}

	modelsPath := scaffoldModels(destDir)
	if modelsPath != "" {
		fmt.Println("✓ models.json scaffolded at " + modelsPath)
	} else {
		fmt.Println("✓ models.json already exists (skipped)")
	}

	runtimeDir := filepath.Join(destDir, ".curse")
	if err := os.MkdirAll(runtimeDir, 0755); err == nil {
		os.MkdirAll(filepath.Join(runtimeDir, "logs"), 0755)
		os.MkdirAll(filepath.Join(runtimeDir, "staging"), 0755)
		fmt.Println("✓ Runtime directories created (.curse/logs, .curse/staging)")
	}

	binaryDir := filepath.Join(destDir, "bin")
	if err := os.MkdirAll(binaryDir, 0755); err == nil {
		fmt.Println("✓ Binary directory created (bin/)")
	}

	fmt.Println()
	fmt.Println("╭──────────────────────────────────────╮")
	fmt.Println("│  Initialization Complete!            │")
	fmt.Println("│                                      │")
	fmt.Println("│  Next steps:                         │")
	fmt.Println("│  1. Run 'curse' to launch dashboard  │")
	fmt.Println("│  2. Press Ctrl+M to switch models    │")
	fmt.Println("│  3. No API keys needed (local mode)  │")
	fmt.Println("╰──────────────────────────────────────╯")
}

func scaffoldModels(destDir string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(destDir, "config")
	}
	modelsDir := filepath.Join(configDir, "curse")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return ""
	}
	modelsPath := filepath.Join(modelsDir, "models.json")
	if _, err := os.Stat(modelsPath); err == nil {
		return ""
	}

	ctx := context.Background()
	reg := gateway.GenerateDefaultLocalRegistry(ctx)
	if reg == nil || len(reg.Profiles) == 0 {
		reg = &gateway.ModelRegistry{
			Active: "codex",
			Profiles: map[string]gateway.ModelProfile{
				"codex": {
					Provider:      "codex",
					Model:         "codex",
					Endpoint:      "builtin://ast-analysis",
					ContextWindow: 32768,
					MaxTokens:     4096,
					Temperature:   0.0,
				},
				"fallback": {
					Provider:      "local-fallback",
					Model:         "fallback",
					Endpoint:      "builtin://local",
					ContextWindow: 4096,
					MaxTokens:    1024,
					Temperature:   0.5,
				},
			},
		}
	}
	data, _ := json.MarshalIndent(reg, "", "  ")
	os.WriteFile(modelsPath, data, 0644)
	return modelsPath
}
