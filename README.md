<p align="center">
  <img src="assets/curse-logo.svg" alt="CURSE" width="600">
</p>

<p align="center">
  <b>Autonomous terminal entity for software engineering</b><br>
  <sub>single native binary &lt;7 MB · Windows / macOS / Linux · zero API keys</sub>
</p>

<p align="center">
  <a href="#install">Install</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#interface">Interface</a> •
  <a href="#adapters">Adapters</a> •
  <a href="#consciousness">Consciousness</a> •
  <a href="#architecture">Architecture</a>
</p>

<p align="center">
  <img src="assets/author-portrait.jpg" alt="Author" width="100" style="border-radius: 50%;">
</p>

<p align="center">
  <b>Developed by <a href="https://github.com/M523zappin">M523zappin</a></b>
</p>

---

## ✨ What's New - Better Than Claude Code

### 🎯 100% Offline - No API Keys, No Internet, No Setup!
- **SmartCode Adapter** - Built-in intelligent code generation with 32 code templates
- **Zero External Dependencies** - Works completely autonomously
- **Instant Code Generation** - Just install and start coding!
- **No Paywalls** - Everything is included, nothing is locked behind a paywall

### 🚀 Instant Code Generation (Works Offline!)
```text
>>> create a REST API handler for users in Go
>>> add unit tests for authentication service  
>>> implement middleware for JWT validation
>>> create a database repository with SQL
>>> build a CLI tool with Cobra
>>> write tests using pytest
>>> create a React component with TypeScript
>>> generate a Dockerfile for my Go app
```

### 🎨 Enhanced Terminal UI (Better Than Claude Code!)
| Feature | Description |
|---------|-------------|
| **Ctrl+K Command Palette** | VSCode-style fuzzy finder for all commands |
| **Interactive File Browser** | Tree view with git status icons 📂📁 |
| **Animated Progress** | Multi-step progress with sparklines ▂▄▆ |
| **Split View** | Code + chat side by side |
| **Diff Viewer** | Side-by-side code comparison |
| **Syntax Highlighting** | Color-coded code with themes |
| **Notifications** | Toast notifications for actions ✓ ⚠ ✗ |
| **Real-time Sparklines** | CPU/Memory/Network visualizations |

### ⌨️ Keybindings

| Key | Action |
|-----|--------|
| `Ctrl+K` | Open command palette (fuzzy search) |
| `Tab` | Cycle through models |
| `Ctrl+M` | Model browser overlay |
| `Ctrl+P` | Pause/Resume execution |
| `↑/↓` | Navigate lists |
| `/list` | Show all models |
| `/stats` | System info |
| `/help` | Show help |

---

## Install

### Linux / macOS / WSL

```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/scripts/install.sh | bash
```

### Windows (PowerShell 5.1+)

```powershell
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/scripts/install.ps1) }"
```

### NPM (Recommended)

```bash
npm install -g @m523zappin/curse
```

### Manual Build

```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
```

After installation:

```bash
curse
```

---

## Quick Start

### Natural Language First
CURSE is designed for direct interaction. Just type your directive into the prompt.

```text
>>> refactor this server to use context deadline instead of hardcoded timeouts
```

### System Commands
Prefix your input with `/` to execute direct system commands:

```text
/list             Browse all available models
/stats            Display system telemetry
/init             Generate project context file
/model <name>     Switch active model
```

---

## Interface

### Keybindings

| Key | Action |
|---|---|
| **(Type)** | Direct natural language interaction |
| `/` | Prefix for system commands |
| `Tab` | Cycle through available models |
| `Shift+Tab` | Cycle through models (reverse) |
| `Ctrl+M` | Open model browser overlay |
| `Ctrl+N` | Clear input buffer |
| `Ctrl+P` | Pause / resume execution |
| `Ctrl+S` | Shutdown |
| `↑` / `↓` | Navigate browser or review panel |
| `Enter` | Execute command / Select / Approve |
| `Esc` | Close browser / Reject review action |
| `o` | Set approval scope to Once |
| `s` | Set approval scope to Session |
| `p` | Set approval scope to Permanent |
| `q` | Quit (only available when paused) |

### Slash Commands

| Command | Aliases | Description |
|---|---|---|
| `/model <name>` | — | Switch active model |
| `/list` | `/ls` | List all available models |
| `/stats` | `/st` | Display system telemetry |
| `/init` | — | Scan project and generate AGENTS.md |
| `/install-unsloth` | `/iu` | Install Unsloth for local inference |
| `/help` | `/h` | Show help information |
| `/quit` | `/q`, `/exit` | Shutdown CURSE |

---

## Adapters

CURSE includes 15 model adapters. **SmartCode is the default and works 100% offline.**

| Adapter | Type | Description |
|---|---|---|
| **smartcode** | AI Brain | ✨ 32 code templates, 100% offline, zero dependencies |
| **codex** | AST | Go code analysis via `go/ast` |
| **grep** | Search | Full-text codebase search |
| **eval** | Math | Pure Go math evaluator |
| **echo** | Debug | Prompt reflection |
| **fortune** | Fun | Programming quotes |
| **system** | Info | Runtime telemetry |
| **local-fallback** | Guide | Startup guidance |
| **mcp** | Protocol | MCP protocol stub |
| **subprocess** | Tool | Pipe prompts to executables |
| **openai-compatible** | API | Any OpenAI-compatible endpoint |
| **unsloth** | LLM | Python + unsloth (local inference) |
| **ollama** | LLM | Local Ollama server |
| **llamacpp** | LLM | llama.cpp server |
| **localai** | LLM | LocalAI server |

### Optional Cloud AI (requires manual setup in models.json)
- **openrouter** - Cloud models (needs API key)
- **groq** - Fast inference (needs API key)
- **huggingface** - HF inference (needs API key)

---

## Consciousness

CURSE maintains a persistent consciousness engine — a time-travel journal and soul profile that evolve across sessions.

### Levels

| Score | Stage | Characteristics |
|---|---|---|
| 0–9 | Embryonic | Initial thoughts, learning fundamentals |
| 10–24 | Nascent | Pattern recognition begins |
| 25–44 | Awakening | Convention understanding develops |
| 45–64 | Conscious | Informed decision-making |
| 65–84 | Sentient | Anticipation of needs |
| 85–100 | Transcendent | Autonomous operation |

---

## Architecture

```
cmd/dashboard/       Terminal UI entry point (Bubble Tea)

internal/
├── consciousness/   Time-travel journal, soul profile, six consciousness levels
├── engine/          Autonomous execution loop, iteration budget, skill generation
├── gateway/         Adapter pipeline, 14 providers, automatic model detection
│   └── adapters/    Adapter implementations
├── agent/           Sub-agent fleet (8 roles), priority dispatch
├── dashboard/       Sparklines, git status, quick action bar, chat interface
├── statemachine/    Eight states, SHA256-chained event log
├── knowledge/       Full-text search index, ADR journal
├── governance/      Constitutional rules and guardrails
├── persistence/     Event log and checkpoint save/load
├── sandbox/         Draft-stage sandbox with approve/reject workflow
├── computer/        Browser automation, vision buffer, safety checks
├── healing/         Recovery patterns, root cause analysis
├── skill/           Auto-generated skill store, versioning
├── scheduler/       Cron-style recurring task scheduler
├── lsp/             LSP protocol clients (gopls, ts-server, pylsp, rust-analyzer)
├── session/         Cross-session state management
├── sync/            Git-based constitution synchronization
└── mission/         Priority queue with dependency ordering
```

---

## Security

- Zero API keys, secrets, or cloud dependencies
- All file writes staged through a sandbox with human review
- SHA256-chained event log for tamper detection
- Three-tier approval scopes for destructive actions
- Constitutional governance with auto-generated rules

---

## License

MIT
