<p align="center">
  <img src="assets/curse-logo.svg" alt="CURSE" width="600">
</p>

<p align="center">
  <b>Autonomous terminal entity for software engineering</b><br>
  <sub>single native binary • Windows / macOS / Linux • zero API keys</sub>
</p>

---

## Install

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/M523zappin/Curse-Core/main/scripts/install.ps1 | iex
```

### Linux / macOS
```bash
curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/main/scripts/install.sh | bash
```

---

## Quick Start

```bash
curse
```

Then just type what you want:

```
>>> create a REST API handler for users in Go
>>> add unit tests for authentication
>>> implement JWT middleware
>>> write a Dockerfile
```

No API keys needed. No cloud setup. Works 100% offline.

---

## Features

- **32 Code Templates** - Go, Python, TypeScript, DevOps (works offline)
- **Smart Auto-Detection** - Picks the right tools automatically
- **Terminal UI** - Beautiful TUI with syntax highlighting
- **Git Integration** - Tracks all changes with SHA256 chain
- **Review System** - Approve/reject file changes before they happen

---

## Keybindings

| Key | Action |
|-----|--------|
| Tab | Cycle models |
| Ctrl+M | Model browser |
| Ctrl+K | Command palette |
| Ctrl+P | Pause/Resume |
| Up/Down | Navigate |
| Enter | Execute |
| Esc | Close/Cancel |

---

## License

MIT
