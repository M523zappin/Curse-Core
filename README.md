# CURSE

### Cognitive Unified Runtime System Entity

**CURSE is not a wrapper. It is an orchestrator.**  
A persistent, autonomous terminal entity that manages your development lifecycle through a crash-recoverable state machine, a fleet of specialized sub-agents, browser-level computer control, and a self-healing feedback loop — all rendered through a professional-grade Bubble Tea TUI.

```
  ╔══════════════════════════════════════════════╗
  ║              C U R S E                       ║
  ║  Cognitive Unified Runtime System Entity     ║
  ║                                              ║
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

## Install

### Linux / macOS / WSL

```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

The installer handles: dependencies (git, Go), repository clone, binary build, `~/.local/bin` PATH registration, `.env` scaffolding, and GitHub CLI authentication.

### Windows (PowerShell 5.1+)

```powershell
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

The installer handles: Git for Windows, Go, PATH registration, `.env` scaffolding, and `gh` auth.

### Manual

```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
./curse
```

### Pre-built Binaries

Pre-compiled binaries are available in the `releases/` directory for immediate use without a Go toolchain.

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
├── gateway/adapters/  # Ollama, OpenAI-compatible, MCP providers
├── computer/          # Playwright browser, desktop OS, vision, safety
├── agent/             # Sub-agent Fleet — 8 specialized roles
├── healing/           # Fail-safe loop — root cause analysis + auto-fix
├── knowledge/         # Live index — ADRs, debug sessions, full-text search
├── lsp/               # LSP client — gopls, ts-server, pylsp integration
├── mission/           # Kanban queue with priority + dependency ordering
├── dashboard/         # Bubble Tea TUI — mission queue, trace, status, review
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

## Keybindings

| Key | Action |
|-----|--------|
| `Ctrl+P` | Pause / Resume |
| `Ctrl+B` | Start browser (Playwright) |
| `Ctrl+Y` | Sync constitution from GitHub |
| `Ctrl+M` | Cycle active model |
| `Ctrl+S` | Shutdown |
| `↑/↓` | Navigate review queue |
| `Enter` | Approve review action |
| `Esc` | Reject review action |

---

## Configuration

Edit `~/.local/share/curse/models.json` (Linux) or `%APPDATA%/curse/models.json` (Windows):

```json
{
  "active": "fast-edit",
  "profiles": {
    "fast-edit": {
      "provider": "ollama",
      "model": "codellama:7b",
      "endpoint": "${OLLAMA_ENDPOINT}",
      "context_window": 8192
    }
  }
}
```

Environment variables in `.env`:
- `OLLAMA_ENDPOINT` — Local Ollama server
- `OPENAI_API_KEY` — OpenAI-compatible API key
- `MCP_ENDPOINT` — MCP server WebSocket endpoint

---

## Security

- **CONSTITUTION.md** — 8 principles, 10 guardrails enforced by the Reviewer sub-agent
- **Draft Before Write** — All file writes staged through sandbox for approval
- **No Secrets** — Credentials via `.env` only (gitignored)
- **SHA256 Chain** — Tamper-evident event log with chain integrity validation
- **HITL Review** — Destructive actions (file delete, financial transactions, terminal commands) require human confirmation

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
