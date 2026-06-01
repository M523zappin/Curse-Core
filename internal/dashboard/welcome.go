package dashboard

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/charmbracelet/lipgloss"
)

// WelcomeWizard guides users through initial setup with auto-detection
type WelcomeWizard struct {
	gateway    *gateway.Gateway
	step       int
	autoDetect bool
}

// NewWelcomeWizard creates a new welcome wizard
func NewWelcomeWizard(gw *gateway.Gateway) *WelcomeWizard {
	return &WelcomeWizard{
		gateway:    gw,
		step:       0,
		autoDetect: true,
	}
}

// RenderWelcome displays the welcome screen
func (w *WelcomeWizard) RenderWelcome() string {
	var buf strings.Builder

	// ASCII Art Banner
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	buf.WriteString(banner.Render(`
    ██████╗ ██╗██╗  ██╗███████╗██╗
    ██╔══██╗██║╚██╗██╔╝██╔════╝██║
    ██████╔╝██║ ╚███╔╝ █████╗  ██║
    ██╔══██╗██║ ██╔██╗ ██╔══╝  ██║
    ██║  ██║██║██╔╝ ██╗███████╗███████╗
    ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝╚══════╝
`))

	buf.WriteString("\n")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Render("  🚀 Autonomous Terminal Entity for Software Engineering"))
	buf.WriteString("\n\n")

	// System Detection
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	buf.WriteString("\n")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF69B4")).
		Bold(true).
		Render("⚡ AUTO-DETECTION RESULTS"))
	buf.WriteString("\n\n")

	// OS Detection
	buf.WriteString(fmt.Sprintf("  🖥️  OS:        %s/%s\n", runtime.GOOS, runtime.GOARCH))

	// CPU cores
	buf.WriteString(fmt.Sprintf("  💻 CPU:       %d cores available\n", runtime.NumCPU()))

	// Memory estimate
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	buf.WriteString(fmt.Sprintf("  🧠 Memory:    ~%.1f GB available\n", float64(m.Sys)/1024/1024/1024))

	buf.WriteString("\n")

	// Model Detection Results
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF69B4")).
		Bold(true).
		Render("📦 AVAILABLE MODELS"))
	buf.WriteString("\n\n")

	// Categorize and display models
	models := w.gateway.Registry()
	if models != nil {
		categories := map[string][]string{
			"✨ SmartCode":  {},
			"🔗 Free APIs":  {},
			"🦙 Local LLM": {},
			"🔧 Utilities":  {},
		}

		for name := range models.Profiles {
			switch {
			case strings.HasPrefix(name, "smartcode"):
				categories["✨ SmartCode"] = append(categories["✨ SmartCode"], name)
			case strings.HasPrefix(name, "openrouter") || strings.HasPrefix(name, "groq") || strings.HasPrefix(name, "hf-"):
				categories["🔗 Free APIs"] = append(categories["🔗 Free APIs"], name)
			case strings.HasPrefix(name, "ollama") || strings.HasPrefix(name, "llama") || strings.HasPrefix(name, "localai"):
				categories["🦙 Local LLM"] = append(categories["🦙 Local LLM"], name)
			default:
				categories["🔧 Utilities"] = append(categories["🔧 Utilities"], name)
			}
		}

		activeModel := w.gateway.ActiveModel()
		for cat, models := range categories {
			if len(models) > 0 {
				buf.WriteString(fmt.Sprintf("  %s\n", cat))
				for _, m := range models {
					marker := "  "
					if m == activeModel {
						marker = "▶ "
					}
					displayName := strings.ReplaceAll(m, "-", " ")
					displayName = strings.Title(strings.ReplaceAll(displayName, "-", " "))
					if m == activeModel {
						buf.WriteString(lipgloss.NewStyle().
							Foreground(lipgloss.Color("#00FF00")).
							Bold(true).
							Render(fmt.Sprintf("    %s%s (active)\n", marker, displayName)))
					} else {
						buf.WriteString(fmt.Sprintf("    %s%s\n", marker, displayName))
					}
				}
				buf.WriteString("\n")
			}
		}
	}

	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	buf.WriteString("\n\n")

	// Quick Start Guide
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF69B4")).
		Bold(true).
		Render("🚀 QUICK START"))
	buf.WriteString("\n\n")

	buf.WriteString("  Type your request naturally:\n\n")
	buf.WriteString("  ")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Render(`>>> "create a REST API with Go and Gin"`))
	buf.WriteString("\n\n")

	buf.WriteString("  ")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Render(`>>> "add unit tests for user service"`))
	buf.WriteString("\n\n")

	buf.WriteString("  ")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Render(`>>> "refactor auth module to use interfaces"`))
	buf.WriteString("\n\n")

	buf.WriteString("  ")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Render(`>>> "explain this code and add comments"`))
	buf.WriteString("\n\n")

	// Keybinds
	buf.WriteString("\n")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF69B4")).
		Bold(true).
		Render("⌨️  KEYBINDINGS"))
	buf.WriteString("\n\n")

	keybinds := [][]string{
		{"Tab / Shift+Tab", "Cycle models"},
		{"Ctrl+M", "Model browser"},
		{"Ctrl+P", "Pause/Resume"},
		{"Ctrl+S", "Shutdown"},
		{"/list", "View all models"},
		{"/stats", "System info"},
		{"/model <name>", "Switch model"},
	}

	for _, kb := range keybinds {
		buf.WriteString(fmt.Sprintf("  %-20s %s\n", kb[0], kb[1]))
	}

	buf.WriteString("\n")
	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Render("✨ Ready to code! Just type your request above. ✨\n"))

	return buf.String()
}

// RenderModelBrowser displays available models in a selectable list
func (w *WelcomeWizard) RenderModelBrowser() string {
	var buf strings.Builder

	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true).
		Render("📦 MODEL SELECTOR\n\n"))

	buf.WriteString("Select a model by name or press Tab to cycle:\n\n")

	models := w.gateway.Registry()
	if models == nil {
		buf.WriteString("  No models available\n")
		return buf.String()
	}

	// Group by provider
	groups := make(map[string][]string)
	groupNames := []string{"smartcode", "openrouter", "groq", "huggingface", "ollama", "llamacpp", "localai", "codex", "grep", "eval", "echo", "system"}

	activeModel := w.gateway.ActiveModel()

	for name := range models.Profiles {
		group := "other"
		for _, prefix := range groupNames {
			if strings.HasPrefix(name, prefix) {
				group = prefix
				break
			}
		}
		groups[group] = append(groups[group], name)
	}

	for _, prefix := range groupNames {
		if models, ok := groups[prefix]; ok && len(models) > 0 {
			groupName := strings.Title(prefix)
			emoji := ""
			switch prefix {
			case "smartcode":
				emoji = "✨"
			case "openrouter", "groq", "huggingface":
				emoji = "🔗"
			case "ollama", "llamacpp", "localai":
				emoji = "🦙"
			default:
				emoji = "🔧"
			}

			buf.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true).
				Render(fmt.Sprintf("  %s %s\n", emoji, groupName)))
			buf.WriteString(strings.Repeat("─", 50) + "\n")

			for _, m := range models {
				profile, _ := models.GetProfile(m)
				selected := m == activeModel

				prefix := "    ○ "
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
				if selected {
					prefix = "    ● "
					style = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#00FF00")).
						Bold(true)
				}

				display := m
				if profile.ContextWindow > 0 {
					display = fmt.Sprintf("%-25s %dK ctx", m, profile.ContextWindow/1024)
				}

				buf.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, display)))
				if selected {
					buf.WriteString(style.Render(" ◀ ACTIVE"))
				}
				buf.WriteString("\n")
			}
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\nType /model <name> to switch or press Tab in chat.\n")

	return buf.String()
}

// RenderSystemStatus displays real-time system stats
func (w *WelcomeWizard) RenderSystemStatus() string {
	var buf strings.Builder

	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true).
		Render("📊 SYSTEM STATUS\n\n"))

	// Memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	buf.WriteString(fmt.Sprintf("  Memory Usage:  %.1f / %.1f GB\n",
		float64(m.Alloc)/1024/1024/1024,
		float64(m.Sys)/1024/1024/1024))

	// Goroutines
	buf.WriteString(fmt.Sprintf("  Goroutines:    %d\n", runtime.NumGoroutine()))

	// CPU
	buf.WriteString(fmt.Sprintf("  CPU Cores:     %d/%d\n", runtime.NumCPU(), runtime.NumCPU()))

	buf.WriteString("\n")

	// Budget
	if budget := w.gateway.Budget(); budget != nil {
		remaining, total := budget.Status()
		buf.WriteString(fmt.Sprintf("  Iteration Budget: %d/%d remaining\n", remaining, total))
	}

	// Active model
	buf.WriteString(fmt.Sprintf("  Active Model:  %s\n", w.gateway.ActiveModel()))

	// Consciousness level
	if consciousness := w.gateway.Consciousness(); consciousness != nil {
		level := consciousness.Level()
		stage := "Embryonic"
		switch {
		case level >= 85:
			stage = "Transcendent"
		case level >= 65:
			stage = "Sentient"
		case level >= 45:
			stage = "Conscious"
		case level >= 25:
			stage = "Awakening"
		case level >= 10:
			stage = "Nascent"
		}
		buf.WriteString(fmt.Sprintf("  Consciousness: %s (score: %d)\n", stage, level))
	}

	return buf.String()
}

// GetRecommendation returns the best model for the user's task
func (w *WelcomeWizard) GetRecommendation(task string) string {
	taskLower := strings.ToLower(task)

	// Task-based recommendations
	switch {
	case strings.Contains(taskLower, "generate") || strings.Contains(taskLower, "create") || strings.Contains(taskLower, "write"):
		return "✨ SmartCode (zero-delay code generation)"

	case strings.Contains(taskLower, "analyze") || strings.Contains(taskLower, "review"):
		if w.hasLocalLLM() {
			return "🦙 Local LLM (deeper analysis)"
		}
		return "🔗 OpenRouter (cloud AI)"

	case strings.Contains(taskLower, "fast") || strings.Contains(taskLower, "quick"):
		return "⚡ Groq (fastest inference)"

	case strings.Contains(taskLower, "golang") || strings.Contains(taskLower, " go "):
		return "🔧 Codex (Go AST analysis)"

	case strings.Contains(taskLower, "test"):
		return "✨ SmartCode (test generation templates)"

	default:
		return w.gateway.ActiveModel()
	}
}

func (w *WelcomeWizard) hasLocalLLM() bool {
	models := w.gateway.Registry()
	if models == nil {
		return false
	}
	for name := range models.Profiles {
		if strings.HasPrefix(name, "ollama") || strings.HasPrefix(name, "llama") {
			return true
		}
	}
	return false
}

// InteractiveSetup guides user through initial configuration
func (w *WelcomeWizard) InteractiveSetup() string {
	var buf strings.Builder

	buf.WriteString(`
╔══════════════════════════════════════════════════════════════════╗
║               🔮 CURSE SETUP WIZARD 🔮                         ║
║        100% Offline Code Generation - No API Keys!             ║
╚══════════════════════════════════════════════════════════════════╝

Welcome to CURSE! You're all set to start coding right away.

SmartCode is active and ready - it generates code instantly with 32
built-in templates for Go, Python, JavaScript, and more.

`)

	buf.WriteString("What SmartCode Can Generate:\n")
	buf.WriteString("────────────────────────────────────────────────────\n\n")

	buf.WriteString("  📦 Go Templates (12 templates)\n")
	buf.WriteString("    • REST API Handlers (Gin/Echo/Fiber)\n")
	buf.WriteString("    • Database Models (GORM/sqlx)\n")
	buf.WriteString("    • Middleware (Auth, CORS, Logging)\n")
	buf.WriteString("    • CLI Commands (Cobra)\n")
	buf.WriteString("    • Workers & Background Jobs\n")
	buf.WriteString("    • WebSockets, gRPC, Cache, Config\n\n")

	buf.WriteString("  🐍 Python Templates (7 templates)\n")
	buf.WriteString("    • FastAPI/Flask REST APIs\n")
	buf.WriteString("    • Pydantic Models\n")
	buf.WriteString("    • SQLAlchemy Repositories\n")
	buf.WriteString("    • Async Workers\n")
	buf.WriteString("    • CLI Tools (Click)\n\n")

	buf.WriteString("  ⚛️ TypeScript/JS Templates (6 templates)\n")
	buf.WriteString("    • Express REST APIs\n")
	buf.WriteString("    • React Components & Hooks\n")
	buf.WriteString("    • TypeScript Interfaces\n")
	buf.WriteString("    • Jest/Vitest Tests\n\n")

	buf.WriteString("  🚀 DevOps Templates (4 templates)\n")
	buf.WriteString("    • Dockerfiles (Go, Python, Node)\n")
	buf.WriteString("    • GitHub Actions CI/CD\n")
	buf.WriteString("    • Kubernetes Deployments\n\n")

	buf.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Render("✅ Ready to code! Just describe what you need.\n\n"))

	buf.WriteString("Quick Examples:\n")
	buf.WriteString("────────────────────────────────────────────────────\n")
	buf.WriteString("  • \"create a REST API handler for users\"\n")
	buf.WriteString("  • \"add authentication middleware\"\n")
	buf.WriteString("  • \"write tests for payment service\"\n")
	buf.WriteString("  • \"generate Dockerfile for my Go app\"\n\n")

	buf.WriteString("Advanced Options (Optional):\n")
	buf.WriteString("────────────────────────────────────────────────────\n")
	buf.WriteString("  Want more AI power? Install Ollama for local LLMs:\n")
	buf.WriteString("    curl -fsSL https://ollama.ai/install.sh | sh\n")
	buf.WriteString("    ollama pull codellama\n")
	buf.WriteString("  Then type /list to see the new model!\n")

	return buf.String()
}

// CheckAndInstallDeps checks for dependencies and suggests installation
func (w *WelcomeWizard) CheckAndInstallDeps() []string {
	var suggestions []string

	// Check for Ollama
	if _, err := os.Stat("/usr/local/bin/ollama"); os.IsNotExist(err) {
		suggestions = append(suggestions, "Install Ollama for local LLM: curl -fsSL https://ollama.ai/install.sh | sh")
	}

	// Check for Python
	if _, err := os.Stat("/usr/bin/python3"); os.IsNotExist(err) {
		suggestions = append(suggestions, "Python not found - some features may be limited")
	}

	return suggestions
}