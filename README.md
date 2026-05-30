```
  ╔══════════════════════════════════════════════════════════════════╗
  ║                                                                  ║
  ║    ██████████  ██      ██  ██████████  ██████████  ██████████   ║
  ║    ██          ██      ██  ██      ██  ██          ██           ║
  ║    ██████████  ██      ██  ██████████  ██████████  ██████████   ║
  ║    ██          ██      ██  ██    ██            ██  ██           ║
  ║    ██████████  ██████████  ██      ██  ██████████  ██████████   ║
  ║                                                                  ║
  ║                     C U R S E                                   ║
  ║              Autonomous Terminal Entity                          ║
  ║                                                                  ║
  ║              zero API keys · 12 adapters · TUI                   ║
  ║                                                                  ║
  ╚══════════════════════════════════════════════════════════════════╝
```

CURSE is a persistent, autonomous terminal entity for software engineering.  
No API keys. No cloud. 12 built-in model adapters with zero dependencies.

---

## Install

```bash
# One line, any platform
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

```powershell
# Windows
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

```bash
# Manual
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
./curse
```

No `.env` file. No API keys. CURSE auto-detects everything on your system and starts.

---

## Quick Start

```
$ curse

  ╔═══════════════════════════════════════════════╗
  ║            C U R S E                          ║
  ║  Scanning subsystems.....                     ║
  ║  [████████████████░░░░░░░]                    ║
  ╚═══════════════════════════════════════════════╝

  → scanning subsystems...
  ✓ python3 3.13.2
  ✓ ollama running at localhost:11434
  ◈ ESTABLISHING ENTITY CONSCIOUSNESS...

  ┌─ ENTITY CONSCIOUSNESS ────────────────────────┐
  │  15:04:05 ▶ entity initialized                │
  │  15:04:06 ▶ model → codex                     │
  │  15:04:07 ▶ 12 adapters ready                 │
  │  15:04:08 ▶ awaiting directive                │
  └───────────────────────────────────────────────┘
```

The boot sequence runs a 12-second entity awakening animation, auto-detects available tools, then transitions to the live dashboard.

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
/stats               System telemetry
/install-unsloth     Install Unsloth (local LLM)
/help                Help
/quit                Shutdown
```

---

## Adapters

CURSE ships with 12 model adapters, all zero API key by design.

| Adapter | Type | Zero Dep | Description |
|---------|------|----------|-------------|
| **codex** | AST | ✓ | Go code analysis via `go/ast` |
| **grep** | Search | ✓ | Full-text codebase search |
| **eval** | Math | ✓ | Pure Go math evaluator |
| **echo** | Debug | ✓ | Prompt structure echo |
| **fortune** | Fun | ✓ | Programming quotes & facts |
| **system** | Info | ✓ | Runtime telemetry |
| **unsloth** | LLM | Python | Local inference via Unsloth/Transformers |
| **ollama** | LLM | Ollama | Local Ollama API |
| **openai-compatible** | API | — | Any OpenAI endpoint |
| **subprocess** | Tool | — | Pipe prompts to executables |
| **local-fallback** | Guide | ✓ | Startup guidance |
| **mcp** | Protocol | — | MCP protocol stub |

### Auto-Detection

On first launch, CURSE scans 4 tiers:

```
Tier 1  builtin    codex, grep, eval, echo, fortune, system, fallback
                  → always available, zero dependencies
Tier 2  python     python-helper, unsloth-fast, unsloth-powerful
                  → detected via python3 + pip check
Tier 3  ollama     ollama-<model> for each pulled model
                  → HTTP check localhost:11434
Tier 4  llama.cpp  llama-server
                  → HTTP check localhost:8080
```

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
│                       │  VITAL SIGNS               │
│                       │  CPU 4G  MEM 42M  GO 12   │
│                       │  ▃▄▆▇█▇▆▅▄▃▂▁             │
│                       │  git ⎇ main ●dirty +3     │
│                       │  model codex  via codex   │
├───────────────────────┴────────────────────────────┤
│  ◉ codex [████░░░░] 42%  00:42:17                 │
│  ◐ idle  8 skills  143 mem  100% heal             │
├─ ACTIONS ─────────────────────────────────────────┤
│  [/ cmd]│[Ctrl+M model]│[Ctrl+P]│[Ctrl+B]│[Ctrl+S]│
└───────────────────────────────────────────────────┘
```

---

## Features

### Frozen-Snapshot Memory

`~/.curse/MEMORY.md` is read once at session start and embedded immutably into the prompt. Cross-session context without API overhead. Changes apply on next session — preserving prompt cache efficiency.

### Iteration Budget

Thread-safe counter (default 100 calls per session). Completed tool calls refund iterations. One grace call on exhaustion prevents runaway loops.

### Approval Scopes

| Key | Scope | Effect |
|-----|-------|--------|
| `o` | Once | Approve this action only |
| `s` | Session | Approve all similar this session |
| `p` | Permanent | Trust this action type forever |

### Self-Healing Loop

| Pattern | Handler |
|---------|---------|
| Connection refused | Exponential backoff |
| Timeout | 2× timeout retry |
| Port conflict | Kill + reassign |
| Browser crash | Auto restart |

### Sub-Agent Fleet

| Role | Count | Domain |
|------|-------|--------|
| Security | 1 | Vulnerability scanning |
| Refactoring | 2 | Code restructuring |
| Infrastructure | 1 | CI/CD, containers |
| Reviewer | 2 | PR review |
| Tester | 1 | Tests, coverage |
| Architect | 1 | Design decisions |
| Dependencies | 1 | Updates, patches |
| Documentation | 1 | Docs, changelogs |

### State Machine

| State | Description |
|-------|-------------|
| `Idle` | Awaiting mission |
| `Running` | Active execution |
| `Paused` | Suspended |
| `Checkpointing` | Writing SHA256 checkpoint |
| `Syncing` | Pulling constitution |
| `Error` | Unrecoverable |
| `Recovering` | Replaying event log |
| `Shutdown` | Graceful exit |

SHA256-chained event log. Recovery in 47-61ms.

### Unsloth Integration

```bash
# Inside CURSE:
/install-unsloth

# Or manually:
pip install unsloth transformers torch accelerate
```

Model loaded once, kept alive in persistent Python subprocess. Zero-latency inference across calls.

### Computer Controller

- **Browser**: Playwright (chromium/firefox/webkit)
- **Vision**: Screenshot, HTML extraction, UI classification
- **Safety**: Destructive action detection + HITL review

### Knowledge Index

Persistent JSON store with full-text search (title 3×, tag 2×, body 1×), ADR recording, tag filtering, cross-session persistence.

### LSP Integration

Auto-connects: `gopls`, `typescript-language-server`, `pylsp`, `rust-analyzer`.  
Diagnostics, completions, symbols, go-to-definition, hover.

---

## Architecture

```
cmd/
├── curse-init/     Bootstrap CLI
├── dashboard/      TUI entry point
└── gateway/        Headless API

internal/
├── statemachine/   8 states · 15 events · SHA256 chain
├── persistence/    Event log · checkpoint
├── governance/     Constitution · 10 guardrails
├── sandbox/        Draft-stage · approve/reject
├── gateway/        Adapter pipeline · tool registry
├── gateway/adapters/  12 providers
├── computer/       Browser · vision · safety
├── agent/          Fleet · 8 roles
├── healing/        20+ recovery patterns
├── knowledge/      FTS index · ADR journal
├── lsp/            gopls · ts-server · pylsp
├── mission/        Priority queue
├── dashboard/      Sparklines · git · quickbar
├── engine/         Autonomous loop · budget
├── scheduler/      Cron tasks
├── session/        Cross-session state
├── skill/          Progressive disclosure
└── sync/           Git constitution sync
```

---

## Security

- **CONSTITUTION.md**: 8 principles, 10 guardrails
- **Draft Before Write**: All mutations staged for review
- **Zero API Keys**: No secrets, no `.env`, no cloud
- **SHA256 Chain**: Tamper-evident event log
- **HITL Review**: Destructive actions require confirmation
- **Approval Scopes**: Once / Session / Permanent

---

## Recovery

1. SHA256 chain integrity verified
2. Last checkpoint loaded (state, step, mission)
3. State machine recovered
4. Processing resumes from checkpoint

Live Fire tests: 47-61ms typical.

---

## License

MIT
