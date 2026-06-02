<p align="center">
  <img src="assets/curse-logo.svg" alt="CURSE" width="600">
</p>

<p align="center">
  <b>Unlimited autonomous terminal entity for software engineering</b><br>
  <sub>no limits · no ads · no friction · single binary · local-first · Windows / macOS / Linux</sub>
</p>

<p align="center">
  <a href="#install">Install</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#why-curse">Why CURSE</a> •
  <a href="#project-status">Project Status</a> •
  <a href="#quality-and-delivery">Quality & Delivery</a> •
  <a href="#interface">Interface</a> •
  <a href="#adapters">Adapters</a> •
  <a href="#architecture">Architecture</a>
</p>

<p align="center">
  <img src="assets/author-portrait.jpg" alt="Author" width="100" style="border-radius: 50%;">
</p>

<p align="center">
  <b>Developed by <a href="https://github.com/M523zappin">M523zappin</a></b>
</p>

<p align="center">
  <a href="https://github.com/M523zappin/Curse-Core/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/M523zappin/Curse-Core/ci.yml?branch=master&label=ci" alt="CI"></a>
  <a href="https://github.com/M523zappin/Curse-Core/releases"><img src="https://img.shields.io/github/v/release/M523zappin/Curse-Core" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

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
CURSE is designed for direct interaction. Type your directive and execute.

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

## Why CURSE

CURSE exists for one reason: **unlimited vibes**.

- No artificial limits — runs as long as needed, no iteration caps, no timeouts
- No ads, no tracking, no telemetry — pure execution, zero distractions
- Local-first with optional external model adapters — you own everything
- Crash-recoverable checkpointing and session continuity
- Single-binary deployment, zero mandatory API keys
- Multi-turn conversation with full context memory
- 14 model adapters, all local-first, most requiring zero configuration

---

## Project Status

CURSE is in active development. Core subsystems are in place and continuously validated in CI, while APIs and operator ergonomics will continue to evolve.

Recommended usage pattern:
- Start in non-critical repositories
- Pin known-good releases for reproducibility
- Track changes in `CHANGELOG.md`

---

## Quality and Delivery

Baseline local verification:

```bash
go mod download
go build ./cmd/...
go test ./...
```

CI and release model:
- CI gates: formatting, vet, tests, cross-platform build matrix
- Automated dependency updates (Go modules and GitHub Actions)
- Tagged releases with generated archives and checksums

On Unix-like systems with `make`, run:

```bash
make ci
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

CURSE includes 14 model adapters. None require API keys.

| Adapter | Type | Dependencies | Description |
|---|---|---|---|
| **codex** | AST | none | Go code analysis via `go/ast` |
| **grep** | Search | none | Full-text codebase search |
| **eval** | Math | none | Pure Go math evaluator |
| **echo** | Debug | none | Prompt reflection |
| **fortune** | Fun | none | Programming quotes |
| **system** | Info | none | Runtime telemetry |
| **local-fallback** | Guide | none | Startup guidance |
| **mcp** | Protocol | none | MCP protocol stub |
| **subprocess** | Tool | — | Pipe prompts to executables |
| **openai-compatible** | API | — | Any OpenAI-compatible endpoint |
| **unsloth** | LLM | Python + unsloth | Direct Python subprocess for local LLM inference |
| **ollama** | LLM | Ollama server | Local Ollama HTTP API |
| **llamacpp** | LLM | llama.cpp server | Native and OpenAI-compatible API |
| **localai** | LLM | LocalAI server | OpenAI-compatible with model listing |

---

## Architecture

```text
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
- File writes routed through staged sandbox workflows
- SHA256-chained event log for tamper-evident traceability
- Multi-scope approvals for destructive actions
- Constitutional governance with rule synchronization

---

## License

MIT
