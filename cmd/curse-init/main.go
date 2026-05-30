package main

import (
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

	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(destDir, "config")
	}
	modelsDir := filepath.Join(configDir, "curse")
	if err := os.MkdirAll(modelsDir, 0755); err == nil {
		modelsPath := filepath.Join(modelsDir, "models.json")
		if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
			reg := gateway.ModelRegistry{
				Active: "fast-edit",
				Profiles: map[string]gateway.ModelProfile{
					"fast-edit": {
						Provider:      "ollama",
						Model:         "codellama:7b",
						Endpoint:      "${OLLAMA_ENDPOINT}",
						ContextWindow:  8192,
					},
					"deep-reasoning": {
						Provider:      "openai-compatible",
						Model:         "gpt-4o",
						Endpoint:      "https://api.openai.com/v1",
						Auth:          gateway.AuthConfig{Header: "Authorization", EnvVar: "OPENAI_API_KEY"},
						ContextWindow:  128000,
					},
					"mcp-editor": {
						Provider:      "mcp",
						Model:         "editor",
						Endpoint:      "${MCP_ENDPOINT}",
						ContextWindow:  32768,
					},
				},
			}
			data, _ := json.MarshalIndent(reg, "", "  ")
			os.WriteFile(modelsPath, data, 0644)
			fmt.Println("✓ models.json scaffolded at " + modelsPath)
		}
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
	fmt.Println("│  1. Edit .env with your API keys     │")
	fmt.Println("│  2. Run 'curse' to launch dashboard  │")
	fmt.Println("│  3. Or build from source: go build   │")
	fmt.Println("╰──────────────────────────────────────╯")
}
