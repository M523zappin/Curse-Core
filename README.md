```
  ╔══════════════════════════════════════════════════════════════════╗
  ║                                                                  ║
  ║      ██████████  ██████████  ██████████  ██████████  ██████████  ║
  ║      ██          ██          ██              ██      ██      ██  ║
  ║      ██████████  ██████████  ██████████      ██      ██████████  ║
  ║              ██          ██          ██      ██      ██      ██  ║
  ║      ██████████  ██████████  ██████████      ██      ██      ██  ║
  ║                                                                  ║
  ║      ██     ██                                                    ║
  ║      ██     ██   ╔═══════════════════════════════════════════╗    ║
  ║      █████████   ║  COGNITIVE UNIFIED RUNTIME SYSTEM ENTITY ║    ║
  ║      ██     ██   ║     Zero API Keys · 12 Adapters · TUI   ║    ║
  ║      ██     ██   ╚═══════════════════════════════════════════╝    ║
  ║                                                                  ║
  ║      ◈  cortex · agents · senses · reflex · memory · language ◈  ║
  ║                                                                  ║
  ╚══════════════════════════════════════════════════════════════════╝
```

# CURSE

**Cognitive Unified Runtime System Entity** — a persistent, autonomous terminal entity for software engineering.

No API keys. No cloud dependency. 12 built-in model adapters with auto-detection. Professional-grade Bubble Tea TUI with sparklines, git status, system vitals, and a self-healing feedback loop.

---

## Quick Start

```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
./curse
```

CURSE auto-detects everything available on your system and starts immediately. No `.env` file needed. No API keys to configure.

### One-Line Install

```bash
# Linux / macOS / WSL
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

```powershell
# Windows (PowerShell 5.1+)
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

**Pre-built binaries** available in `releases/`.

---

## Visual Identity

```
┌─ TITLE BAR ──────────────────────────────────────────────┐
│  ◉ CURSE v1.0.0  │  codex  │  RUNNING                    │
├──────────────────────────┬───────────────────────────────┤
│  ◐ ENTITY DIRECTIVES     │  ◑ ENTITY CONSCIOUSNESS       │
│  ┌─────┬──────┬────────┐ │  [15:04:05] entity initialized│
│  │todo │in prg│  done  │ │  [15:04:06] model → codex    │
│  │     │      │        │ │  [15:04:07] sync complete    │
│  └─────┴──────┴────────┘ │                               │
│                          │  VITAL SIGNS                   │
│                          │  vessel ● myhost  PID 1234    │
│                          │  state ◉ RUNNING  engine idle │
│                          │  CPU 4G  MEM 42M/128M  GO 12  │
│                          │  CPU:▃▄▆▇█▇▆▅▄▃▂▁ 4           │
│                          │  git ⎇ main ●dirty +3         │
│                          │  model codex  via codex       │
├──────────────────────────┴───────────────────────────────┤
│  ◉ codex [████░░░░] 42%  00:42:17                        │
│  ◐ idle  8 skills  143 mem  100% heal                    │
├─ QUICK ACTIONS ──────────────────────────────────────────┤
│  [/ cmd]│[Ctrl+M model]│[Ctrl+P pause]│[Ctrl+B browse]│  │
├──────────────────────────────────────────────────────────┤
│  ● RUNNING │ session-id │ seq:42                         │
└──────────────────────────────────────────────────────────┘
```

---

## Keybindings

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
| `o` / `s` / `p` | Set approval scope (once/session/permanent) |

## Commands

```
/model <name>        Switch to a model
/list                List all available models
/stats               Show system telemetry (models, budget, memory, uptime)
/install-unsloth     Install Unsloth via pip for local LLM inference
/help                Show all commands
/quit                Shutdown
```

---

## 12 Built-in Adapters (Zero API Keys)

| Adapter | Type | What it does | Deps |
|---------|------|-------------|------|
| **codex** | AST | Go code analysis — lists functions, types, files | None |
| **grep** | Search | Full-text codebase search across .go/.ts/.py/.rs/.md | None |
| **eval** | Math | Pure Go math evaluator via `go/parser` (sin, cos, sqrt, pow, log) | None |
| **echo** | Debug | Echoes the full prompt structure back | None |
| **fortune** | Fun | Programming quotes, jokes, facts, riddles, motivation | None |
| **system** | Info | Live system runtime — CPU, memory, goroutines, Go version | None |
| **unsloth** | LLM | Persistent Python subprocess running Unsloth/Transformers | Python + `pip install unsloth` |
| **ollama** | LLM | Local Ollama API with auto-detect | Ollama installed + running |
| **openai-compatible** | API | Any OpenAI-compatible endpoint | Configurable endpoint |
| **subprocess** | Tool | Pipes prompts to any executable | The executable |
| **local-fallback** | Guide | Helpful startup guide when no model is configured | None |
| **mcp** | Protocol | MCP protocol stub for extensibility | Configurable |

### Auto-Detection (4 Tiers)

```
Tier 1 ── Built-in (codex, grep, eval, echo, fortune, system, fallback)
            Always available, zero deps, no configuration needed

Tier 2 ── Python / Unsloth (python-helper, unsloth-fast, unsloth-powerful)
            Detects python3, checks for pip install unsloth

Tier 3 ── Ollama (ollama-<model> for each pulled model)
            HTTP check at localhost:11434

Tier 4 ── llama.cpp (llama-server)
            HTTP check at localhost:8080
```

---

## Features

### Frozen-Snapshot Memory

CURSE reads `~/.curse/MEMORY.md` at session start and embeds it immutably into the system prompt. This cross-session memory survives restarts and provides persistent context without API overhead. Changes take effect on the next session — preserving prompt cache efficiency.

### Iteration Budget

A thread-safe counter limits tool calls per session (default: 100). Completed tool calls refund iterations back to the budget. On exhaustion, one grace call is allowed for a summary response, preventing runaway loops.

### Approval Scopes (HITL Review)

Destructive actions (file deletes, terminal commands, browser purchases) pause in the TUI for human confirmation with three scope levels:

| Scope | Behavior |
|-------|----------|
| `o` — Once | Approve this action only |
| `s` — Session | Approve all similar actions this session |
| `p` — Permanent | Trust this action type permanently |

### Self-Healing Loop

Errors are caught, classified (info/warning/critical), analyzed for root cause (20+ pattern matchers), and automatically remediated.

| Pattern | Handler |
|---------|---------|
| Connection refused | Exponential backoff retry |
| Timeout | 2× timeout retry |
| Port conflict | Kill + reassign |
| Browser crash | Automatic restart |

### Sub-Agent Fleet (8 Specialized Roles)

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

### Unsloth Integration

Keep a HuggingFace model loaded in a persistent Python subprocess for zero-latency local inference:

```bash
# From within CURSE:
/install-unsloth

# Or manually:
pip install unsloth transformers torch accelerate
```

Models auto-detected:
- `unsloth/Llama-3.2-1B-Instruct` (fast, CPU-friendly)
- `unsloth/Mistral-7B-Instruct-v0.3`
- `unsloth/Qwen2.5-7B-Instruct`
- And more...

### Computer Controller

- **Browser**: Playwright-based (chromium/firefox/webkit)
- **Vision**: Screenshot capture, HTML extraction, UI classification
- **Safety**: Destructive action detection with HITL review
- **Desktop**: File operations, application launch, terminal commands

### State Machine (8 States)

| State | Description |
|-------|-------------|
| `Idle` | Initial, awaiting mission |
| `Running` | Actively executing |
| `Paused` | User or system pause |
| `Checkpointing` | Writing SHA256 checkpoint |
| `Syncing` | Pulling latest constitution |
| `Error` | Unrecoverable error |
| `Recovering` | Replaying event log on restart |
| `Shutdown` | Graceful termination |

SHA256-chained event log ensures tamper-evident crash recovery. Recovery in 47-61ms typical.

### Knowledge Index

Persistent JSON knowledge store with:
- Full-text search (title 3×, tag 2×, body 1× weighting)
- ADR recording, debug session capture
- Tag and type filtering
- Cross-session persistence

### LSP Integration

Auto-detects and connects to language servers:
- Go → `gopls`
- TypeScript → `typescript-language-server`
- Python → `pylsp`
- Rust → `rust-analyzer`

Provides: diagnostics, completions, document symbols, go-to-definition, hover.

---

## Configuration

Models are auto-detected on first launch. To customize:

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
    }
  }
}
```

Location: `~/.config/curse/models.json` (Linux/macOS) or `%APPDATA%/curse/models.json` (Windows).

---

## Architecture

```
cmd/
├── curse-init/        # Bootstrap CLI — clones, scaffolds
├── dashboard/         # TUI entry point (Bubble Tea)
└── gateway/           # Headless API entry point

internal/
├── statemachine/      # 8 states, 15 events, SHA256 chain
├── persistence/       # Event log, checkpoint save/load
├── governance/        # CONSTITUTION.md parser, 10 guardrails
├── sandbox/           # Draft-staging with approve/reject
├── gateway/           # Adapter pipeline + tool registry
├── gateway/adapters/  # 12 providers (see above)
├── computer/          # Browser, desktop, vision, safety
├── agent/             # Sub-agent fleet (8 roles)
├── healing/           # Self-healing loop (20+ patterns)
├── knowledge/         # Persistent index, full-text search
├── lsp/               # gopls, ts-server, pylsp client
├── mission/           # Kanban queue with priority
├── dashboard/         # TUI: sparklines, git, quickbar
├── engine/            # Autonomous loop + iteration budget
├── scheduler/         # Cron-style recurring tasks
├── session/           # Cross-session state
├── skill/             # Progressive-disclosure skills
└── sync/              # Git-based constitution syncer
```

---

## Security

- **CONSTITUTION.md**: 8 principles, 10 guardrails
- **Draft Before Write**: All writes staged through sandbox
- **No Secrets**: Zero API key design; no `.env` needed
- **SHA256 Chain**: Tamper-evident event log
- **HITL Review**: Destructive actions require human confirmation
- **Approval Scopes**: Once / Session / Permanent

---

## Recovery

On restart:
1. SHA256 chain integrity verified
2. Last checkpoint loaded (state, step, mission ID)
3. State machine recovered to previous state
4. Processing resumes from checkpoint

Live Fire tests: 47-61ms typical recovery latency.

---

## License

MIT
