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
  <img src="assets/crona-portrait.jpg" alt="Author" width="100" style="border-radius: 50%;">
</p>

<p align="center">
  <b>Developed by <a href="https://github.com/M523zappin">M523zappin</a></b>
</p>

CURSE is a high-performance, autonomous terminal entity designed for deep software engineering. It bridges the gap between natural language intent and codebase execution through a unified, intelligence-driven interface. Operating fully offline with zero API keys, CURSE understands your codebase, dispatches specialized sub-agents, and maintains a persistent consciousness across sessions.

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


### NPM (Recommended for Node.js users)

```bash
npm install -g @m523zappin/curse
```

### Manual build

```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
```

After install:

```bash
curse
```

No configuration. No setup. The terminal UI boots in 12 seconds.

---

## Quick Start

### Natural Language First
CURSE is designed for direct interaction. There is no need to toggle modes—simply type your directive into the prompt to begin.

```text
>>> refactor this server to use context deadline instead of hardcoded timeouts
```

CURSE decomposes your request into discrete tasks, dispatches them to specialized sub-agents, collects results, and records the outcome in its consciousness journal.

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

### Auto-Detection Tiers

On first launch, CURSE automatically discovers available tools in priority order:

| Tier | Condition | Models Registered |
|---|---|---|
| 1 | Always available | codex, grep, eval, echo, fortune, system, local-fallback, mcp |
| 2 | Python interpreter found | subprocess helpers |
| 3 | Unsloth Python package installed | Preconfigured profiles for Llama 4, Qwen 3, DeepSeek Coder V3, Gemma 3, Phi 4, Mistral Large |
| 4 | Ollama server running (localhost:11434) | All pulled Ollama models |
| 5 | llama.cpp server running (localhost:8080) | All served models enumerated from `/v1/models` |
| 6 | LocalAI server running (localhost:8080) | All served models enumerated from `/v1/models` |

Detected profiles are written to `~/.config/curse/models.json`. The active model is persisted across restarts.

---

## Consciousness

CURSE maintains a persistent consciousness engine — a time-travel journal and soul profile that evolve across sessions. Every decision is recorded, and every outcome informs future behavior.

### Levels

| Score | Stage | Characteristics |
|---|---|---|
| 0–9 | Embryonic | Initial thoughts, learning fundamentals |
| 10–24 | Nascent | Pattern recognition begins |
| 25–44 | Awakening | Convention understanding develops |
| 45–64 | Conscious | Informed decision-making |
| 65–84 | Sentient | Anticipation of needs |
| 85–100 | Transcendent | Autonomous operation |

### Components

**Time-Travel Journal** — A circular buffer of up to 5,000 thoughts. Each thought carries a `prev_id` chain pointer that enables decision graph traversal and context reconstruction on restart. Persisted to `~/.curse/consciousness/journal.json`.

**Soul Profile** — Learns codebase patterns from mission outcomes. Each pattern receives a confidence weight (`1 - 1 / (observations + 1)`), and patterns are sorted by descending confidence. The profile tracks naming conventions, error handling styles, and architectural decisions. Persisted to `~/.curse/consciousness/profile.json`.

**Constitution Generation** — From observed conventions, CURSE auto-generates constitutional governance rules that grow more precise as the consciousness accumulates data.

### Level Formula

```
consciousness_level = thoughts(30%) + patterns(25%) + type_diversity(20%) + uptime(15%) + conventions(10%)
```

The current level is displayed in the dashboard VITAL SIGNS panel with color-coded indicators that shift as the entity advances through stages.

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

## Key Subsystems

### Sub-Agent Fleet
Eight specialized agent roles — Security, Refactoring (two), Infrastructure, Reviewer (two), Tester, Architect, Dependency Management, and Documentation. Tasks are dispatched by priority with dependency resolution and execute in parallel.

### State Machine
Eight states — Idle, Running, Paused, Checkpointing, Syncing, Error, Recovering, Shutdown — with a SHA256-chained event log providing a tamper-evident audit trail. Crash recovery completes in under 100 milliseconds.

### Self-Healing Loop
Over 20 recovery patterns including connection retry with exponential backoff, timeout doubling, port conflict resolution, and browser crash auto-restart. Recovery rate is tracked and displayed in the dashboard.

### Frozen-Snapshot Memory
The file `~/.curse/MEMORY.md` is read once at session start and embedded immutably into the system prompt, providing cross-session context without API overhead.

### Iteration Budget
A thread-safe 100-call budget per session. Completed tool calls refund iterations. A single grace call on exhaustion prevents runaway execution loops.

### Auto-Skill Generation
Every successful mission generates a reusable skill document with structured steps, tags, confidence scoring, and pattern matching. Skills are stored as JSON (for programmatic search) and markdown (for human readability) and are automatically matched to similar future tasks via weighted search.

### HITL Review
Destructive actions are staged in a sandbox and require human approval through three configurable scopes: Once, Session, and Permanent. The entire workflow is keyboard-driven.

### Knowledge Index
A persistent full-text search index with ADR journaling, tag filtering, and cross-session retention. Titles are weighted 3x, tags 2x, and body 1x in search scoring.

### LSP Integration
Automatically connects to gopls, typescript-language-server, pylsp, and rust-analyzer for diagnostics, completions, symbol lookup, and go-to-definition.

### Browser Automation
Playwright-driven browser control with vision buffer, UI classification, pre-click safety checks, and destructive action detection.

---

## Recovery
On restart, CURSE verifies SHA256 chain integrity, loads the last checkpoint, recovers the state machine, and replays the consciousness journal. Typical recovery time is under 100 milliseconds.

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

