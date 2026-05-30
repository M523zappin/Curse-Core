<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="curse-logo.svg">
    <img src="curse-logo.svg" alt="CURSE" width="600">
  </picture>
</p>

<p align="center">
  <b>Autonomous Terminal Entity</b><br>
  <sub>zero API keys · 12 adapters · TUI · fully offline</sub>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#install">Install</a> •
  <a href="#adapters">Adapters</a> •
  <a href="#architecture">Architecture</a>
</p>

**CURSE** — Cognitive Unified Runtime System Entity — is an autonomous terminal entity for software engineering. **No API keys. No cloud. 12 built-in adapters. Zero external dependencies.** A single <7 MB native binary that delivers a professional TUI dashboard, crash-recoverable state machine, built-in code analysis, auto-generated skills, and local LLM inference — all without ever reaching for a cloud service.

---

## Install

Install CURSE in one command. No API keys. No configuration. No `.env` file.

**Linux / macOS / WSL**
```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

**Windows (PowerShell 5.1+)**
```powershell
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

**Manual build** (requires Go 1.26+)
```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
./curse
```

After install, just run:
```
curse
```
The TUI dashboard boots autonomously with a 12-second animation sequence.

---

## Quick Start

```
$ curse

  ╔════════════════════════════════════════════════╗
  ║                   C U R S E                    ║
  ║         Scanning subsystems.....               ║
  ║         [████████████████░░░░░░░░]             ║
  ╚════════════════════════════════════════════════╝

  → scanning subsystems...
  ✓ python3 3.13.2
  ✓ ollama running at localhost:11434
  ◈ ESTABLISHING ENTITY CONSCIOUSNESS...

  ┌─ ENTITY CONSCIOUSNESS ──────────────────┐
  │  15:04:05 ▶ entity initialized          │
  │  15:04:06 ▶ model → codex               │
  │  15:04:07 ▶ 12 adapters ready           │
  │  15:04:08 ▶ awaiting directive          │
  └─────────────────────────────────────────┘
```

A 12-second boot animation scans subsystems, auto-detects tools, awakens the entity, then transitions to the live dashboard.

---

## Keybindings

| Key | Action |
|-----|--------|
| `/` | Command mode |
| `Ctrl+M` | Model browser |
| `Ctrl+P` | Pause / Resume |
| `Ctrl+B` | Start browser |
| `Ctrl+Y` | Sync constitution |
| `Ctrl+S` | Shutdown |
| `↑/↓` | Navigate |
| `Enter` | Select / approve |
| `Esc` | Close / reject |
| `o` / `s` / `p` | Approval scope |

### Commands

```
/model <name>        Switch model
/list                List all models
/stats               System telemetry (models, budget, memory, uptime)
/install-unsloth     Install Unsloth for local LLM inference
/help                Show help
/quit                Shutdown
```

---

## Adapters

12 model adapters. Zero API keys required.

| Adapter | Type | Deps | Description |
|---------|------|------|-------------|
| **codex** | AST | none | Go code analysis via `go/ast` |
| **grep** | Search | none | Full-text codebase search |
| **eval** | Math | none | Pure Go math evaluator |
| **echo** | Debug | none | Prompt structure echo |
| **fortune** | Fun | none | Programming quotes & facts |
| **system** | Info | none | Runtime telemetry |
| **unsloth** | LLM | Python | Local inference via Unsloth/Transformers |
| **ollama** | LLM | Ollama | Local Ollama API |
| **openai-compatible** | API | — | Any OpenAI endpoint |
| **subprocess** | Tool | — | Pipe prompts to executables |
| **local-fallback** | Guide | none | Startup guidance |
| **mcp** | Protocol | — | MCP protocol stub |

On first launch, CURSE auto-detects available tools:

```
Tier 1  builtin   codex · grep · eval · echo · fortune · system · fallback
                  → always available
Tier 2  python    python-helper · unsloth-fast · unsloth-powerful
                  → detected via python3 + pip check
Tier 3  ollama    ollama-<model> for each pulled model
                  → HTTP check localhost:11434
Tier 4  llama.cpp llama-server
                  → HTTP check localhost:8080
```

---

## Why CURSE

The only AI coding platform that runs **fully offline with zero API keys** — surpassing both Claude Code and Hermes Agent.

| Feature | CURSE | Claude Code | Hermes Agent |
|---------|-------|-------------|--------------|
| **API keys required** | **None** — fully offline | Anthropic API key required | API key or local LLM |
| **Binary size** | **< 7 MB** (single native binary) | ~200 MB (npm + deps) | Python + pip environment |
| **Runtime** | **None** — native Go binary | Node.js required | Python 3.10+ required |
| **Built-in adapters** | **12** (6 pure Go, zero-dependency) | 1 (Claude API only) | LLM backends only |
| **Auto code analysis** | **Built-in** (go/ast, grep, math) | ❌ Requires model | ❌ Requires model |
| **State machine** | **8 states · SHA256-chained** — crash-recoverable | ❌ No crash recovery | ❌ No crash recovery |
| **Thread safety** | **Full** — every subsystem mutex-protected | Not applicable (Node.js) | Not applicable (Python) |
| **Auto-skill generation** | **Built-in** — markdown + JSON with confidence scoring | Via plugins | Via learning loop |
| **TUI dashboard** | **Professional** — sparklines, git status, system vitals, animated boot, model browser | Basic CLI | TUI only |
| **Self-healing** | **20+ patterns** — exponential backoff, port recovery, browser restart | ❌ Not available | ❌ Not available |
| **Sub-agent fleet** | **8 specialized roles** — priority dispatch, parallel execution | ✓ Similar | ❌ Not available |
| **Frozen memory** | **Session-persistent** — MEMORY.md embedded immutably | CLAUDE.md only | FTS5 memory |
| **LSP integration** | **Built-in** — gopls, ts-server, pylsp, rust-analyzer | ❌ Not built-in | ❌ Not built-in |
| **Browser automation** | **Playwright** — vision buffer, pre-click safety, destructive detection | Computer Use (beta) | ❌ Not built-in |
| **Review scopes** | **3 scopes** — Once / Session / Permanent | Approve only | Approve only |
| **Git governance** | **Constitution** — 8 principles, 10 guardrails, git-syncable | CLAUDE.md only | ❌ Not available |
| **Multi-platform** | **Windows, macOS, Linux** — one binary each | macOS, Linux only | Linux, macOS, WSL2 |
| **Auto-detection** | **4-tier** — builtin → unsloth → ollama → llama.cpp | ❌ Not available | Model detection only |
| **Unsloth LLM** | **Built-in adapter** — persistent Python subprocess | ❌ Not available | ❌ Not available |
| **License** | MIT | Proprietary | MIT |

---

## Dashboard

```
┌─ TITLE ───────────────────────────────────────────┐
│  ◉ CURSE v1.0.0  │  codex  │  RUNNING             │
├──────────────────────┬────────────────────────────┤
│  ◐ DIRECTIVES        │  ◑ CONSCIOUSNESS           │
│  ┌────┬────┬──────┐  │  [15:04] entity init      │
│  │todo│prg │ done │  │  [15:04] model → codex    │
│  │    │    │      │  │  [15:04] sync complete    │
│  └────┴────┴──────┘  │                            │
│                       │  CPU 4G  MEM 42M  GO 12   │
│                       │  ▃▄▆▇█▇▆▅▄▃▂▁             │
│                       │  git ⎇ main ●dirty +3     │
│                       │  model codex  via codex   │
├───────────────────────┴────────────────────────────┤
│  ◉ codex [████░░░░] 42%  00:42:17                 │
│  ◐ idle  8 skills  143 mem  100% heal             │
├─ ACTIONS ─────────────────────────────────────────┤
│  [/ cmd]  [M model]  [P pause]  [B browse]  [S]   │
└────────────────────────────────────────────────────┘
```

---

## Features

### Auto-Skill Generation

When CURSE completes a successful mission, it automatically generates a reusable skill document. Each skill includes:

- **Structured markdown docs** with description, tags, steps, and usage instructions
- **Confidence tracking** via success/failure counters per skill
- **Pattern matching** — skills auto-apply to similar future tasks via weighted search
- **Versioned skill store** with JSON + markdown dual persistence
- **Team-shareable** — skills are plain files in `~/.curse/skills/`

Skills are stored as both JSON (for programmatic search) and markdown (for human readability), enabling both machine dispatch and developer review.

### Frozen-Snapshot Memory

`~/.curse/MEMORY.md` read once at session start, embedded immutably into the system prompt. Cross-session context without API overhead. Changes apply on next session — preserves prompt cache.

### Iteration Budget

Thread-safe counter: 100 calls per session. Completed tool calls refund iterations. One grace call on exhaustion prevents runaway loops.

### Approval Scopes

| Key | Scope | Effect |
|-----|-------|--------|
| `o` | Once | Approve this action only |
| `s` | Session | Approve all similar this session |
| `p` | Permanent | Trust this action type forever |

### Self-Healing Loop

| Pattern | Response |
|---------|----------|
| Connection refused | Exponential backoff |
| Timeout | 2× timeout retry |
| Port conflict | Kill + reassign |
| Browser crash | Auto restart |

### Sub-Agent Fleet

8 specialized roles: Security, Refactoring (2), Infrastructure, Reviewer (2), Tester, Architect, Dependency Manager, Documentation. Tasks dispatched by priority with dependency resolution.

### State Machine

8 states: `Idle`, `Running`, `Paused`, `Checkpointing`, `Syncing`, `Error`, `Recovering`, `Shutdown`. SHA256-chained event log. Recovery in 47-61ms.

### Unsloth Integration

```bash
/install-unsloth              # inside CURSE
pip install unsloth           # or manually
```

Model loaded once, kept alive in persistent Python subprocess. Zero-latency inference across calls.

### Computer Controller

Playwright browser automation (chromium/firefox/webkit), screenshot vision, UI classification, destructive action detection with HITL review.

### Knowledge Index

Persistent JSON store with full-text search (title 3×, tag 2×, body 1×), ADR recording, tag filtering, cross-session persistence.

### LSP Integration

Auto-connects: `gopls`, `typescript-language-server`, `pylsp`, `rust-analyzer`. Diagnostics, completions, symbols, go-to-definition, hover.

---

## Architecture

```
cmd/
├── curse-init/     Bootstrap CLI
├── dashboard/      TUI entry point (Bubble Tea)
└── gateway/        Headless API

internal/
├── statemachine/   8 states · 15 events · SHA256 chain
├── persistence/    Event log · checkpoint save/load
├── governance/     Constitution · 10 guardrails
├── sandbox/        Draft-stage with approve/reject
├── gateway/        Adapter pipeline + tool registry
├── gateway/adapters/  12 providers
├── computer/       Browser · desktop · vision · safety
├── agent/          Fleet · 8 specialized roles
├── healing/        20+ recovery patterns
├── knowledge/      FTS index · ADR journal
├── lsp/            gopls · ts-server · pylsp client
├── mission/        Priority queue with dependency ordering
├── dashboard/      Sparklines · git status · quickbar
├── engine/         Autonomous loop · iteration budget
├── scheduler/      Cron-style recurring tasks
├── session/        Cross-session state
├── skill/          Progressive disclosure skills
└── sync/           Git-based constitution syncer
```

---

## Security

- **CONSTITUTION.md**: 8 principles, 10 guardrails
- **Draft Before Write**: All writes staged through sandbox
- **Zero API Keys**: No secrets, no `.env`, no cloud
- **SHA256 Chain**: Tamper-evident event log
- **HITL Review**: Destructive actions require human confirmation
- **Approval Scopes**: Once / Session / Permanent

---

## Recovery

On restart: SHA256 chain integrity verified, checkpoint loaded, state machine recovered, processing resumes. Live Fire tests: 47-61ms typical.

---

## License

MIT
