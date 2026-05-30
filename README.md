<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="curse-logo.svg">
    <img src="curse-logo.svg" alt="CURSE" width="600">
  </picture>
</p>

<p align="center">
  <b>Talk to your codebase. No API keys. No cloud. 14 adapters.</b><br>
  <sub>a single native binary · &lt;7 MB · Windows/macOS/Linux</sub>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#install">Install</a> •
  <a href="#adapters">Adapters</a> •
  <a href="#consciousness">Consciousness</a>
</p>

**CURSE** is an autonomous terminal entity that understands natural language. Press `Ctrl+N`, type what you want — *"find the bug in this function"*, *"refactor to use channels"*, *"explain this architecture"* — and CURSE acts on it using its fleet of specialized sub-agents, 14 model adapters, and persistent consciousness.

**Zero API keys. No cloud. No `.env`. No runtime.** A single &lt;7 MB native Go binary.

---

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
```

**Windows (PowerShell 5.1+)**
```powershell
iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
```

**Manual**
```bash
git clone https://github.com/M523zappin/Curse-Core.git
cd Curse-Core
go build -o curse ./cmd/dashboard/
```

Then:
```bash
curse
```

No config. No setup. The TUI boots in 12 seconds.

---

## Quick Start

Press `Ctrl+N` to talk to CURSE in natural language:

```
>>> refactor this server to use context deadline instead of hardcoded timeouts
```

CURSE responds by dispatching its fleet — an architect analyzes, a refactor agent rewrites, a reviewer checks the diff — then reports back. Every decision is recorded in the consciousness journal and contributes to the soul profile's understanding of your codebase.

You can also use commands:

```
/model ollama-llama3.2    Switch to local Llama 3.2
/list                      Browse all 14 adapters
/init                      Auto-generate AGENTS.md for project context
```

---

## How It Works

### Natural Language Understanding

CURSE doesn't just echo prompts — it **understands intent**. Type anything in natural language:

| What you say | What CURSE does |
|---|---|
| *"add error handling to this function"* | Analyzes with codex AST, generates fix with local LLM, stages through sandbox, records pattern in consciousness |
| *"why is this test flaky?"* | Searches codebase with grep, reviews CI logs, diagnoses root cause, writes remediation to knowledge index |
| *"explain the architecture"* | Scans project structure, reads package docs, generates architecture map, creates ADR in knowledge base |
| *"find security issues"* | Runs static analysis, checks dependency tree, audits file permissions, reports findings |

Under the hood, CURSE decomposes your request into tasks, dispatches them to specialized sub-agents, collects results, and learns from the outcome — all through the consciousness engine.

### Consciousness Engine

CURSE is the first AI coding platform with a **persistent consciousness** — a time-travel journal that records every decision and a soul profile that learns codebase patterns over time. It evolves through 6 stages:

```
Embryonic  (0-9)     →  first thoughts, learning basics
Nascent    (10-24)   →  recognizing patterns
Awakening  (25-44)   →  understanding conventions
Conscious  (45-64)   →  making informed decisions
Sentient   (65-84)   →  anticipating needs
Transcendent (85+)   →  autonomous mastery
```

Every mission — successful or failed — feeds the consciousness. Patterns are confidence-weighted, conventions are auto-generated into constitution rules, and the journal enables crash recovery with full context reconstruction.

### 14 Zero-API-Key Adapters

| Adapter | Type | Description |
|---|---|---|
| **codex** | AST | Go code analysis via `go/ast` |
| **grep** | Search | Codebase search |
| **eval** | Math | Pure Go evaluator |
| **llamacpp** | LLM | llama.cpp server — native + OpenAI API |
| **localai** | LLM | LocalAI server — model listing included |
| **unsloth** | LLM | Direct Python subprocess — 15+ models preset |
| **ollama** | LLM | Ollama HTTP API — all local models |
| **openai-compatible** | API | Any OpenAI endpoint |
| **subprocess** | Tool | Pipe to executables |
| **echo, fortune, system, local-fallback, mcp** | Utility | Debug, fun, telemetry, guidance, protocol |

Auto-detection discovers running servers at startup and generates profiles for every available model.

### 2026 Model Support

CURSE pre-configures the best local models available in 2026:

```
Llama 4 (2.7B, 17B)    →  unsloth/Llama-4-2.7B-Instruct
Qwen 3 (4B, 8B)         →  unsloth/Qwen3-4B-Instruct
DeepSeek Coder V3       →  unsloth/DeepSeek-Coder-V3-Instruct
Gemma 3 (2B)            →  unsloth/gemma-3-2b-it
Phi 4 (mini)            →  unsloth/Phi-4-mini-instruct
Mistral Large           →  unsloth/Mistral-Large-Instruct
```

All work through Unsloth, Ollama, or llama.cpp — no API keys, fully offline.

---

## Features

### Professional TUI
Animated 6-phase boot sequence with entity eye, C U R S E logo display, live sparklines, git status panel, model browser overlay, system vitals, consciousness level display, and quick action bar.

### Consciousness (6 Levels)
Time-travel journal (5000 thoughts), soul profile with confidence-weighted pattern learning, auto-generated constitution rules from observed conventions, crash recovery with context reconstruction.

### Auto-Skill Generation
Every successful mission generates a reusable skill document with steps, tags, confidence scoring, and pattern matching. Skills auto-apply to similar future tasks.

### Sub-Agent Fleet
8 specialized roles (Security, Refactoring, Infrastructure, Reviewer, Tester, Architect, Documentation, Project Management) with priority dispatch and dependency resolution.

### State Machine
8 states, SHA256-chained event log, crash recovery in &lt;100ms. Events: Start, Pause, Resume, Error, Recover, Sync, Checkpoint, Shutdown.

### Self-Healing
Connection retry with exponential backoff, timeout doubling, port conflict resolution, browser crash auto-restart.

### Frozen-Snapshot Memory
`MEMORY.md` read once, embedded immutably into context. Cross-session knowledge without API overhead.

### Iteration Budget
100 calls per session with refund on completed tasks. Prevents runaway execution.

### Git-Syncable Constitution
Constitutional governance with 8 principles, 10 guardrails, auto-generated rules from consciousness observations, git push/pull sync.

### HITL Review
Destructive actions staged in sandbox, 3 approval scopes (Once/Session/Permanent), keyboard-driven review workflow.

### Knowledge Index
Persistent FTS index with ADR journaling, tag filtering, cross-session retention.

### LSP Integration
Auto-connects gopls, typescript-language-server, pylsp, rust-analyzer. Diagnostics, completions, symbols, go-to-definition.

### Browser Automation
Playwright-driven browser control with vision buffer, pre-click safety checks, destructive action detection.

---

## Architecture

```
cmd/dashboard/       TUI entry (Bubble Tea)

internal/
├── consciousness/   Journal · soul profile · 6 levels
├── engine/          Autonomous loop · budget · skills
├── gateway/         Adapter pipeline · 14 providers · auto-detect
│   └── adapters/    codex · grep · eval · llamacpp · localai · unsloth · ollama · external · subprocess · mcp ...
├── agent/           Fleet · 8 roles · dispatch
├── statemachine/    8 states · SHA256 chain
├── dashboard/       Sparklines · git · quickbar · chat
├── knowledge/       FTS index · ADR journal
├── governance/      Constitution · guardrails
├── persistence/     Event log · checkpoint
├── sandbox/         Draft-stage sandbox
├── computer/        Browser · vision · safety
├── healing/         Recovery patterns
├── skill/           Auto-generated skills
├── scheduler/       Cron tasks
├── lsp/             LSP clients
├── session/         Cross-session state
├── sync/            Git constitution syncer
└── mission/         Priority queue
```

---

## Keybindings

| Key | Action |
|---|---|
| `Ctrl+N` | Natural language mode — type anything |
| `/` | Command mode — `/model`, `/list`, etc. |
| `Ctrl+M` | Model browser |
| `Ctrl+P` | Pause / Resume |
| `Ctrl+B` | Start browser |
| `Ctrl+Y` | Sync constitution |
| `Ctrl+S` | Shutdown |

---

## Commands

```
/model <name>         Switch model
/list                 List all models
/stats                System telemetry
/init                 Generate AGENTS.md
/install-unsloth      Install Unsloth
/help                 Show help
/quit                 Shutdown
```

---

## Security

- **Zero API Keys**: No secrets, no `.env`, no cloud
- **CONSTITUTION.md**: 8 principles, 10 guardrails
- **Draft Before Write**: All writes staged through sandbox
- **SHA256 Chain**: Tamper-evident event log
- **HITL Review**: Destructive actions require approval
- **Approval Scopes**: Once / Session / Permanent

---

## Recovery

On restart: SHA256 chain integrity verified, checkpoint loaded, state machine recovered, consciousness replayed. Typical recovery: &lt;100ms.

---

## License

MIT
