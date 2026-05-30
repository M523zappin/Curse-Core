# CURSE

### Cognitive Unified Runtime System Entity

**CURSE is not a wrapper. It is an orchestrator.**  
A persistent, autonomous terminal entity that manages your development lifecycle through a crash-recoverable state machine, a fleet of specialized sub-agents, browser-level computer control, and a self-healing feedback loop — all rendered through a professional-grade Bubble Tea TUI.

**Zero API keys required.** CURSE ships with 12 built-in model adapters including AST code analysis, codebase search, math evaluation, and optional Unsloth integration for local LLM inference — all offline, all free.

```
  ╔══════════════════════════════════════════════╗
  ║              C U R S E                       ║
  ║  Cognitive Unified Runtime System Entity     ║
  ║                                              ║
  ║  • 12 built-in model adapters, zero API keys ║
  ║  • State machine orchestration               ║
  ║  • Crash-recoverable event chain             ║
  ║  • Sub-agent fleet (8 domains)               ║
  ║  • Computer controller (browser + desktop)   ║
  ║  • Self-healing failure loop                 ║
  ║  • Persistent knowledge index                ║
  ║  • LSP-First diagnostics engine              ║
  ║  • HITL review mode                          ║
  ╚══════════════════════════════════════════════╝
```

---

## Quick Start

```bash
# Clone and build
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
./curse
```

No API keys. No configuration. CURSE auto-detects everything available on your system and starts immediately.

### One-Line Install

**Linux / macOS / WSL:**
```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

**Windows (PowerShell 5.1+):**
```powershell
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

**Pre-built binaries** are available in the `releases/` directory.

---

## Model Adapters (Zero API Keys)

| Adapter | How it works | Dependencies |
|---------|-------------|-------------|
| **codex** | Go AST analysis — lists functions, types, files, generates templates | None (pure Go) |
| **grep** | Full-text codebase search across .go/.ts/.py/.rs/.md | None (pure Go) |
| **eval** | Math evaluator — `sin(pi/4)*180/pi`, `sqrt(pow(3,2)+pow(4,2))` via Go AST | None (pure Go) |
| **echo** | Debug — echoes the full prompt structure back | None |
| **fortune** | Programming quotes, jokes, facts, riddles | None |
| **system** | Live system info — CPU, memory, Go runtime, goroutines | None |
| **unsloth** | Persistent Python subprocess running Unsloth/Transformers | Python + `pip install unsloth` |
| **ollama** | Local Ollama API — auto-detects running instance | Ollama installed + running |
| **openai-compatible** | Any OpenAI-compatible API endpoint | Configurable |
| **subprocess** | Pipes prompt to any executable (python3, llama-cli, node) | Executable in PATH |
| **local-fallback** | Helpful guide when no model is configured | None |
| **mcp** | MCP protocol stub (extensible) | Configurable |

### Auto-Detection

On first launch, CURSE scans your system in 4 tiers:

```
Tier 1 — Built-in (codex, grep, eval, echo, fortune, system, fallback)
  └── Always available, zero deps
Tier 2 — Python / Unsloth (python-helper, unsloth-fast, unsloth-powerful)
  └── Detects python3, checks for pip install unsloth
Tier 3 — Ollama (ollama-<model> for each pulled model)
  └── HTTP check at localhost:11434
Tier 4 — llama.cpp (llama-server)
  └── HTTP check at localhost:8080
```

### Switch Models

| Method | Action |
|--------|--------|
| `Ctrl+M` | Open model browser overlay |
| `/model <name>` | Switch by name (tab-complete with `/list`) |
| `/list` | Show all available models |
| `/install-unsloth` | Auto-install Unsloth via pip |
| Arrow keys + Enter | Navigate and select in model browser |

---

## Dashboard

### Keybindings

| Key | Action |
|-----|--------|
| `/` | Enter command mode |
| `Ctrl+M` | Open model browser overlay |
| `Ctrl+P` | Pause / Resume |
| `Ctrl+B` | Start browser (Playwright) |
| `Ctrl+Y` | Sync constitution from GitHub |
| `Ctrl+S` | Shutdown |
| `↑/↓` | Navigate model browser / review queue |
| `Enter` | Select model / approve review |
| `Esc` | Close overlay / reject review |

### Commands

```
/model <name>        Switch to a model
/list                List all available models
/stats               Show system telemetry
/install-unsloth     Install Unsloth via pip
/help                Show all commands
/quit                Shutdown
```

### Visual Layout

```
┌─ TITLE BAR ──────────────────────────────────────────┐
│  ◉ CURSE v1.0.0 │ codex │ RUNNING                    │
├───────────────────────┬──────────────────────────────┤
│  ENTITY DIRECTIVES    │  ENTITY CONSCIOUSNESS        │
│  ┌─────┬──────┬──────┐│  [15:04:05] system initialized│
│  │TODO │IN PRG│ DONE ││  [15:04:06] model → codex    │
│  │     │      │      ││  [15:04:07] sync complete    │
│  └─────┴──────┴──────┘│                              │
│                       │  VITAL SIGNS                  │
│                       │  vessel ● host  PID 1234      │
│                       │  state ◉ RUNNING  engine idle │
│                       │  CPU 4G  MEM 42M/128M  GO 12  │
│                       │  CPU:▃▄▆▇█▇▆▅▄▃▂▁ 4           │
│                       │  git ⎇ main ●dirty +3         │
│                       │  model codex  via codex       │
├───────────────────────┴──────────────────────────────┤
│  ◉ codex [████░░░░] 42%  00:42:17                    │
│  ◐ planning  8 skills  143 mem  100% heal            │
├─ QUICK ACTIONS ──────────────────────────────────────┤
│  [/ cmd]│[Ctrl+M model]│[Ctrl+P pause]│[Ctrl+B browse]│
├──────────────────────────────────────────────────────┤
│  ● RUNNING │ session-id │ seq:42                      │
└──────────────────────────────────────────────────────┘
```

---

## Architecture

```
cmd/
├── curse-init/        # Bootstrap CLI — clones, scaffolds, configures
├── dashboard/         # TUI entry point — launches the entity
├── gateway/           # Programmatic Gateway API
└── recoverytest/      # Crash-recovery Live Fire test

internal/
├── statemachine/      # 8 states, 15 events, SHA256-chained transitions
├── persistence/       # Append-only event.log, checkpoint save/load
├── governance/        # CONSTITUTION.md parser, 10-guardrail Reviewer
├── sandbox/           # UUID-indexed staging with Approve/Reject workflow
├── gateway/           # Model-agnostic Adapter pipeline + Tool Registry
├── gateway/adapters/  # 12 providers: codex, grep, eval, echo, fortune,
│                      #   system, unsloth, ollama, openai-compatible,
│                      #   subprocess, local-fallback, mcp
├── computer/          # Playwright browser, desktop OS, vision, safety
├── agent/             # Sub-agent Fleet — 8 specialized roles
├── healing/           # Fail-safe loop — root cause analysis + auto-fix
├── knowledge/         # Live index — ADRs, debug sessions, full-text search
├── lsp/               # LSP client — gopls, ts-server, pylsp integration
├── mission/           # Kanban queue with priority + dependency ordering
├── dashboard/         # Bubble Tea TUI — sparklines, git status, quickbar
├── sync/              # Git-based constitution syncer
└── ...                # Supporting modules
```

### State Machine

| State | Description |
|-------|-------------|
| `Idle` | Initial, awaiting mission |
| `Running` | Actively executing |
| `Paused` | User or system pause |
| `Checkpointing` | Writing SHA256 checkpoint |
| `Syncing` | Pulling latest constitution from GitHub |
| `Error` | Unrecoverable error |
| `Recovering` | Replaying event log on restart |
| `Shutdown` | Graceful termination |

Every transition is logged to a SHA256-chained event log for tamper-evident crash recovery.

### Sub-Agent Fleet

| Role | Count | Domain |
|------|-------|--------|
| Security Auditor | 1 | Vulnerability scanning, secret detection |
| Refactoring | 2 | Code restructuring, tech debt reduction |
| Infrastructure | 1 | Deployment, CI/CD, container orchestration |
| Code Reviewer | 2 | PR review, style enforcement |
| Tester | 1 | Test generation, coverage analysis |
| Architect | 1 | Design decisions, ADR management |
| Dependency Manager | 1 | Update analysis, vulnerability patching |
| Documentation | 1 | README, API docs, changelogs |

Tasks are assigned by priority with dependency resolution.

### Computer Controller

- **Browser**: Playwright-based (chromium/firefox/webkit) via `npx playwright run`
- **Desktop**: Application launch, file operations, terminal commands
- **Vision**: Screenshot capture, element HTML extraction, UI element classification
- **Safety Check**: Pre-click screenshots with destructive action detection (delete/purchase/submit)
- **HITL Review**: Destructive actions pause in TUI for user confirmation (Enter/ Esc)

### Self-Healing Loop

Errors are caught, classified (info/warning/critical), analyzed for root cause (20+ pattern matchers), and automatically remediated. Built-in handlers include:
- Connection refused → exponential backoff retry
- Timeout → 2x timeout retry
- Port conflict → kill + reassign
- Browser crash → automatic restart

### Knowledge Index

Every session writes to `.curse/knowledge/` as JSON entries. The index supports:
- Full-text search (title 3x, tag 2x, body 1x weighting)
- Tag and type filtering
- ADR recording, debug session capture, architectural decisions
- Persistent across restarts (loaded from disk)

### LSP Integration

CURSE auto-detects and connects to language servers:
- Go → `gopls`
- TypeScript/JavaScript → `typescript-language-server`
- Python → `pylsp`
- Rust → `rust-analyzer`

Provides: diagnostics, completions, document symbols, go-to-definition, hover information.

---

## Configuration

Models are auto-detected on first launch. To customize, edit `~/.config/curse/models.json` (Linux/macOS) or `%APPDATA%/curse/models.json` (Windows):

```json
{
  "active": "codex",
  "selection_strategy": "manual",
  "profiles": {
    "codex": {
      "provider": "codex",
      "model": "codex",
      "endpoint": "builtin://ast-analysis",
      "context_window": 32768,
      "max_tokens": 4096,
      "temperature": 0.0
    },
    "unsloth-fast": {
      "provider": "unsloth",
      "model": "unsloth/Llama-3.2-1B-Instruct",
      "endpoint": "python://unsloth",
      "context_window": 8192,
      "max_tokens": 2048,
      "temperature": 0.3
    }
  }
}
```

Environment variables (optional):
- `OLLAMA_ENDPOINT` — Local Ollama server (default: `http://localhost:11434`)
- `OPENAI_API_KEY` — OpenAI-compatible API key
- `MCP_ENDPOINT` — MCP server endpoint

---

## Unsloth Integration

CURSE can use [Unsloth](https://github.com/unslothai/unsloth) for local LLM inference:

```bash
# From within CURSE:
/install-unsloth

# Or manually:
pip install unsloth transformers torch accelerate
```

Once installed, CURSE auto-detects Unsloth and adds profiles for:
- `unsloth/Llama-3.2-1B-Instruct` (fast, CPU-friendly)
- `unsloth/Llama-3.2-3B-Instruct`
- `unsloth/Mistral-7B-Instruct-v0.3`
- `unsloth/Qwen2.5-1.5B-Instruct`
- `unsloth/gemma-2-2b-it`
- `unsloth/Phi-3.5-mini-instruct`

The adapter keeps the model loaded in a persistent Python subprocess for zero-latency inference across requests.

---

## System Vitals

The dashboard shows live system telemetry updated every 200ms:

- **CPU**: logical core count
- **MEM**: Go process memory (allocated / total system)
- **GO**: goroutine count
- **Sparklines**: 12-sample rolling window for CPU, memory, goroutines
- **Git status**: branch name, dirty state, untracked files
- **Engine phase**: idle / planning / dispatching / executing / collecting / learning

---

## Security

- **CONSTITUTION.md** — 8 principles, 10 guardrails enforced by the Reviewer sub-agent
- **Draft Before Write** — All file writes staged through sandbox for approval
- **No Secrets** — Credentials via `.env` only (gitignored)
- **SHA256 Chain** — Tamper-evident event log with chain integrity validation
- **HITL Review** — Destructive actions require human confirmation

---

## Recovery

CURSE is designed to survive crashes. On restart:
1. Event log is loaded and SHA256 chain integrity is verified
2. Last checkpoint is loaded (session state, step counter, mission ID)
3. State machine is recovered to its previous state
4. Processing resumes from the last checkpoint

Live Fire tests validate recovery with 47-61ms typical latency.

---

## License

MIT
