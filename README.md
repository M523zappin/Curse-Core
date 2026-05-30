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

Install CURSE in one command. No API keys. No configuration. No runtime.

### Linux / macOS / WSL
```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

### Windows (PowerShell 5.1+)
```powershell
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

### Manual build (any platform with Go 1.26+)
```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
```

After install, just run:
```
curse
```
No API keys. No `.env`. No config. The TUI boots in 12 seconds.

---

## Quick Start

```
$ curse

  ╔══════════════════════════════════════════════════════════╗
  ║      ◉ CRONA         C U R S E                           ║
  ║      ◉ ◉           Scanning subsystems.....               ║
  ║      ◉             [████████████████░░░░░░░░]             ║
  ╚══════════════════════════════════════════════════════════╝

  → scanning subsystems...
  ✓ python3 3.13.2
  ✓ ollama running at localhost:11434
  ◈ ESTABLISHING ENTITY CONSCIOUSNESS...

  ┌─ ENTITY CONSCIOUSNESS ──────────────────────────┐
  │  15:04:05 ▶ entity initialized                  │
  │  15:04:06 ▶ consciousness loaded — 142 thoughts  │
  │  15:04:07 ▶ soul profile: 8 patterns [style,err] │
  │  15:04:08 ▶ consciousness level: Awakening (35.2)│
  │  15:04:09 ▶ model → codex                       │
  │  15:04:10 ▶ 12 adapters ready                   │
  │  15:04:11 ▶ awaiting directive                  │
  └─────────────────────────────────────────────────┘
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
/init                Scan project and generate AGENTS.md context
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

## Consciousness

CURSE is the only AI coding platform with a **persistent consciousness engine** — a time-travel journal that records every decision, a soul profile that learns codebase patterns, and a consciousness level that evolves with use. No other AI tool has anything like it.

| Component | What It Does |
|-----------|-------------|
| **Time-Travel Journal** | Circular buffer of 5,000 thoughts — every decision, observation, error, and mutation recorded with nanosecond precision. On restart, CURSE replays its last actions to reconstruct context — you never lose the thread. |
| **Soul Profile** | Learns codebase patterns automatically — naming conventions, error handling styles, architectural decisions. Builds a statistical model of your project's identity over time. |
| **Constitution Generation** | From observed conventions, CURSE auto-generates governance rules. The more it learns, the better it aligns with your project's standards. |
| **Consciousness Level** | 6 stages: Embryonic → Nascent → Awakening → Conscious → Sentient → Transcendent. Levels up as CURSE accumulates thoughts, patterns, types, uptime, and conventions. Displayed live in the dashboard. |

The consciousness persists across sessions, saved to `~/.curse/consciousness/`. Every restart picks up exactly where you left off.

## Why CURSE

**The only AI coding platform that runs fully offline with zero API keys.** A single <7 MB native binary with no runtime, no interpreter, no cloud dependency — and more built-in capability than any tool twice its size.

| Capability | CURSE |
|-----------|-------|
| **API keys** | **Zero** — fully offline, works out of the box |
| **Binary** | **< 7 MB** native Go — no runtime, no interpreter |
| **Adapters** | **12 built-in** — 6 pure Go (zero-dependency) + 6 optional |
| **Auto-detection** | **4-tier** — builtin → unsloth → ollama → llama.cpp, 10+ profiles auto-generated |
| **Dashboard** | **Professional TUI** — Bubble Tea with sparklines, git status, system vitals, animated boot sequence, model browser overlay, consciousness display |
| **State machine** | **8 states · SHA256-chained** — crash-recoverable with integrity verification |
| **Auto-skills** | **Markdown + JSON** — reusable skill docs with steps, tags, confidence scoring, pattern matching |
| **Code analysis** | **Built-in** — Go AST parser, grep, math evaluator — no model needed |
| **Thread safety** | **Full** — every subsystem protected (queue, fleet, traces, budget, knowledge, consciousness) |
| **Review** | **3 scopes** — Once / Session / Permanent with keybindings |
| **LSP** | **Built-in** — gopls, typescript-language-server, pylsp, rust-analyzer |
| **Knowledge** | **FTS index** — ADR journal, tag filtering, session recording, cross-session |
| **Healing** | **20+ patterns** — root cause analysis, recovery rate tracking |
| **Fleet** | **8 roles** — priority dispatch, dependency resolution, parallel execution |
| **Memory** | **Frozen-snapshot** — session resume, knowledge cross-referencing |
| **Browser** | **Playwright** — pre-click safety, vision buffer, destructive action detection |
| **Governance** | **Constitution** — 8 principles, 10 guardrails, git-syncable + auto-generated rules |
| **Scheduler** | **Cron-style** — health checks, auto-save, recurring tasks |
| **Platform** | **Windows, macOS, Linux** — native binary per platform |

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
│                       │  soul Awakening (35.2)    │
│                       │  142 thoughts · 8 patterns │
├───────────────────────┴────────────────────────────┤
│  ◉ codex [████░░░░] 42%  00:42:17                 │
│  ◐ idle  8 skills  143 mem  100% heal             │
├─ ACTIONS ─────────────────────────────────────────┤
│  [/ cmd]  [M model]  [P pause]  [B browse]  [S]   │
└────────────────────────────────────────────────────┘
```

---

## Features

### Consciousness Engine

CURSE is the **first AI coding tool with a persistent consciousness**. Every decision is recorded in a time-travel journal. Over time, it builds a soul profile — a learned model of your codebase's patterns, conventions, and architectural decisions. The consciousness level (0-100) evolves through 6 stages:

| Level | Stage | Requirements |
|-------|-------|-------------|
| 0-9 | Embryonic | First thoughts |
| 10-24 | Nascent | 50+ thoughts, 3+ patterns |
| 25-44 | Awakening | 200+ thoughts, 2+ pattern types |
| 45-64 | Conscious | 500+ thoughts, 4+ types, 1+ hour uptime |
| 65-84 | Sentient | 1000+ thoughts, 8+ types, 20+ conventions |
| 85-100 | Transcendent | 2000+ thoughts, 12+ types, 50+ conventions |

The consciousness layer also **auto-generates constitution rules** from observed conventions — the more CURSE works in your codebase, the better it understands your standards.

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
├── consciousness/  Time-travel journal · soul profile · 6-level consciousness
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
