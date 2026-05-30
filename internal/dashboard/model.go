package dashboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/M523zappin/Curse-Core/internal/computer"
	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
)

type animTick time.Time

type BootPhase int

const (
	BootPhaseScan   BootPhase = iota
	BootPhaseDetect
	BootPhaseAwaken
	BootPhaseActive
)

type Model struct {
	gateway      *gateway.Gateway
	width        int
	height       int
	ready        bool
	maxVisible   int
	traceItems   []TraceEntry
	lastSeqRead  int64
	paused       bool
	missionQueue *MissionQueueModel
	systemStatus *SystemStatusModel
	quitting     bool
	logPath      string
	cpPath       string
	animFrame    int
	reviewPanel  *ReviewPanelModel
	browserReady bool
	startTime    time.Time

	modelBrowserVisible bool
	modelBrowserIdx     int
	modelBrowserList    []string

	commandMode   bool
	commandBuffer string

	chatMode   bool
	chatBuffer string

	traceMu  sync.Mutex
	bootTick int
	detectedTools []string
	bootPhaseLogged BootPhase
}

type TraceEntry struct {
	Timestamp time.Time
	Message   string
	Level     string
}

var version = "v1.0.0"

func NewModel(gw *gateway.Gateway) *Model {
	tools := gateway.DetectLocalTools(context.Background())
	detected := make([]string, 0, len(tools))
	for _, t := range tools {
		label := t.Name
		if t.Version != "" {
			label += " " + t.Version
		}
		detected = append(detected, label)
	}

	m := &Model{
		gateway:      gw,
		traceItems:   make([]TraceEntry, 0),
		paused:       false,
		missionQueue: NewMissionQueueModel(gw.Queue()),
		systemStatus: NewSystemStatusModel(gw),
		maxVisible:   25,
		lastSeqRead:  0,
		startTime:    time.Now(),
		detectedTools: detected,
	}

	if gw.ReviewManager() != nil {
		m.reviewPanel = NewReviewPanelModel(gw.ReviewManager())
	}

	return m
}

func (m *Model) SetLogPaths(logPath, cpPath string) {
	m.logPath = logPath
	m.cpPath = cpPath
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return splashMsg{} },
		m.animTicker(),
	)
}

type splashMsg struct{}

func (m *Model) boot() BootPhase {
	t := m.bootTick
	switch {
	case t < 20:
		return BootPhaseScan
	case t < 40:
		return BootPhaseDetect
	case t < 60:
		return BootPhaseAwaken
	default:
		return BootPhaseActive
	}
}

func (m *Model) bootView(width int) string {
	if width < 60 {
		width = 60
	}
	f := m.animFrame
	phase := m.boot()

	accent := PulseColor(f)
	dimAccent := ColorAccentDim

	switch phase {
	case BootPhaseScan:
		scanPos := (m.bootTick * 2) % (width - 8)
		scanLine := strings.Repeat(" ", scanPos) + "█" + strings.Repeat(" ", (width-8)-scanPos-1)

		barFull := m.bootTick * (width - 12) / 20
		if barFull > width-12 {
			barFull = width - 12
		}
		loadingBar := "[" + strings.Repeat("█", barFull) + strings.Repeat("░", (width-12)-barFull) + "]"

		dots := strings.Repeat(".", (m.bootTick%8)+1)
		lines := []string{
			"",
			lipgloss.NewStyle().Foreground(dimAccent).Width(width).Align(lipgloss.Center).Render("╔" + strings.Repeat("═", width-4) + "╗"),
			lipgloss.NewStyle().Foreground(accent).Bold(true).Width(width).Align(lipgloss.Center).Render("   C U R S E"),
			lipgloss.NewStyle().Foreground(ColorFgSubtle).Width(width).Align(lipgloss.Center).Render("   Cognitive Unified Runtime System Entity"),
			lipgloss.NewStyle().Foreground(ColorFgInactive).Width(width).Align(lipgloss.Center).Render(""),
			lipgloss.NewStyle().Foreground(ColorFgSubtle).Width(width).Align(lipgloss.Center).Render("   Scanning subsystems" + dots),
			lipgloss.NewStyle().Foreground(ColorSpiral).Width(width).Align(lipgloss.Center).Render("   " + loadingBar),
			lipgloss.NewStyle().Foreground(ColorFgInactive).Width(width).Align(lipgloss.Center).Render(""),
			lipgloss.NewStyle().Foreground(ColorToxic).Width(width).Align(lipgloss.Center).Render("   " + scanLine),
			"",
			lipgloss.NewStyle().Foreground(dimAccent).Width(width).Align(lipgloss.Center).Render("╚" + strings.Repeat("═", width-4) + "╝"),
		}
		return strings.Join(lines, "\n")

	case BootPhaseDetect:
		barFull := (m.bootTick - 20) * (width - 12) / 20
		if barFull > width-12 {
			barFull = width - 12
		}
		loadingBar := "[" + strings.Repeat("█", barFull) + strings.Repeat("░", (width-12)-barFull) + "]"

		var detectLines []string
		detectLines = append(detectLines, "",
			lipgloss.NewStyle().Foreground(dimAccent).Width(width).Align(lipgloss.Center).Render("╔"+strings.Repeat("═", width-4)+"╗"),
			lipgloss.NewStyle().Foreground(accent).Bold(true).Width(width).Align(lipgloss.Center).Render("   SYSTEM DETECTION"),
			"")

		visibleCount := (m.bootTick - 20) / 2
		if visibleCount > len(m.detectedTools) {
			visibleCount = len(m.detectedTools)
		}
		for i, tool := range m.detectedTools {
			if i >= visibleCount {
				break
			}
			check := lipgloss.NewStyle().Foreground(ColorSuccess).Render("✓")
			toolStr := lipgloss.NewStyle().Foreground(ColorFg).Render(tool)
			detectLines = append(detectLines,
				lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(fmt.Sprintf("   %s %s", check, toolStr)))
		}
		if visibleCount < len(m.detectedTools) {
			detectLines = append(detectLines,
				lipgloss.NewStyle().Foreground(ColorFgInactive).Width(width).Align(lipgloss.Center).Render("   ... scanning ..."))
		}
		detectLines = append(detectLines, "",
			lipgloss.NewStyle().Foreground(ColorSpiral).Width(width).Align(lipgloss.Center).Render("   "+loadingBar),
			lipgloss.NewStyle().Foreground(dimAccent).Width(width).Align(lipgloss.Center).Render("╚"+strings.Repeat("═", width-4)+"╝"))
		return strings.Join(detectLines, "\n")

	case BootPhaseAwaken:
		eye := EntityMark(f)
		eyeBlock := strings.Join(eye, "\n")
		intensity := (m.bootTick - 40) * 5
		if intensity > 100 {
			intensity = 100
		}
		glow := ""
		if intensity > 50 {
			glow = lipgloss.NewStyle().Foreground(ColorWarning).Render(" ◈ ENTITY CONSCIOUSNESS ESTABLISHED ◈ ")
		} else {
			pct := lipgloss.NewStyle().Foreground(ColorSpiral).Render(fmt.Sprintf(" %d%%", intensity*2))
			glow = lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(" Awakening") + pct
		}
		lines := []string{
			"",
			lipgloss.NewStyle().Foreground(dimAccent).Width(width).Align(lipgloss.Center).Render("╔"+strings.Repeat("═", width-4)+"╗"),
			"",
			lipgloss.NewStyle().Foreground(accent).Width(width).Align(lipgloss.Center).Render(eyeBlock),
			"",
			lipgloss.NewStyle().Foreground(accent).Bold(true).Width(width).Align(lipgloss.Center).Render(strings.Join(CurseTitle, "\n")),
			"",
			lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(glow),
			"",
			lipgloss.NewStyle().Foreground(dimAccent).Width(width).Align(lipgloss.Center).Render("╚"+strings.Repeat("═", width-4)+"╝"),
		}
		return strings.Join(lines, "\n")

	default:
		return ""
	}
}

func (m *Model) showSplash() {
	if m.boot() == BootPhaseActive {
		return
	}
	phase := m.boot()
	if phase == m.bootPhaseLogged {
		return
	}
	m.bootPhaseLogged = phase

	switch phase {
	case BootPhaseScan:
		m.AddTrace("system", "⟐ scanning subsystems...")
	case BootPhaseDetect:
		for _, tool := range m.detectedTools {
			m.AddTrace("system", fmt.Sprintf("✓ %s", tool))
		}
		m.AddTrace("entity", "██████████████████████████████████████████████████████████")
		m.AddTrace("entity", "  ◈  ESTABLISHING ENTITY CONSCIOUSNESS...")
	case BootPhaseAwaken:
		m.AddTrace("entity", "  ██  cortex » 8-state machine  ·  SHA256-chain memory")
		m.AddTrace("entity", "  ██  agents » 8 specialized minds  ·  priority dispatch")
		m.AddTrace("entity", "  ██  senses » Playwright browser  ·  desktop  ·  vision")
		m.AddTrace("entity", "  ██  reflex » self-healing loop  ·  root-cause analysis")
		m.AddTrace("entity", "  ██  memory » persistent knowledge index  ·  ADR journal")
		m.AddTrace("entity", "  ██  language » LSP diagnostics  ·  gopls  ·  typescript")
		m.AddTrace("entity", "  ██  ethics » HITL review  ·  constitution guardrails")
		m.AddTrace("system", "  ◈  ENTITY ACTIVE  ·  awaiting directive")
		m.AddTrace("system", "  ◈  Ctrl+N talk  ·  Tab cycle model  ·  Ctrl+M browse  ·  Ctrl+P pause  ·  / commands")
	}
}

func (m *Model) animTicker() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return animTick(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}
		m.maxVisible = (msg.Height - 12) / 2
		if m.maxVisible < 8 {
			m.maxVisible = 8
		}

	case animTick:
		m.animFrame++
		m.pollEventLog()
		m.pollCheckpoint()
		initSystemSparklines()
		tickSparklines()
		if m.reviewPanel != nil {
			m.reviewPanel.Update(msg)
		}
		if m.boot() != BootPhaseActive {
			m.bootTick++
			m.showSplash()
		}
		return m, m.animTicker()

	case splashMsg:
		m.bootTick = 0

	case tea.KeyMsg:
		// ── Chat mode takes priority ──
		if m.chatMode {
			switch msg.String() {
			case "enter":
				m.executeChat()
				m.chatMode = false
				m.chatBuffer = ""
			case "esc":
				m.chatMode = false
				m.chatBuffer = ""
				m.AddTrace("system", "Natural language mode cancelled")
			case "backspace":
				if len(m.chatBuffer) > 0 {
					m.chatBuffer = m.chatBuffer[:len(m.chatBuffer)-1]
				}
			case "ctrl+c", "ctrl+s":
				m.quitting = true
				m.gateway.Machine().Send(statemachine.EventShutdownRequested)
				return m, tea.Quit
			default:
				if len(msg.String()) == 1 && msg.String()[0] >= 32 {
					m.chatBuffer += msg.String()
				}
			}
			return m, nil
		}

		// ── Command mode takes priority ──
		if m.commandMode {
			switch msg.String() {
			case "enter":
				m.executeCommand()
				m.commandMode = false
				m.commandBuffer = ""
			case "esc":
				m.commandMode = false
				m.commandBuffer = ""
			case "backspace":
				if len(m.commandBuffer) > 0 {
					m.commandBuffer = m.commandBuffer[:len(m.commandBuffer)-1]
				}
			case "ctrl+c", "ctrl+s":
				m.quitting = true
				m.gateway.Machine().Send(statemachine.EventShutdownRequested)
				return m, tea.Quit
			default:
				if len(msg.String()) == 1 && msg.String()[0] >= 32 {
					m.commandBuffer += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "ctrl+s":
			m.quitting = true
			m.gateway.Machine().Send(statemachine.EventShutdownRequested)
			return m, tea.Quit
		case "ctrl+n":
			m.chatMode = true
			m.chatBuffer = ""
			m.AddTrace("system", "╔══ Natural language mode — type a directive and press Enter ══╗")
			m.AddTrace("system", "║  Say anything: describe code, ask questions, request changes    ║")
			m.AddTrace("system", "║  Esc to cancel · Enter to send                                ║")
			m.AddTrace("system", "╚════════════════════════════════════════════════════════════════╝")
		case "ctrl+p":
			m.paused = !m.paused
			if m.paused {
				m.gateway.Machine().Send(statemachine.EventPauseRequested)
				m.AddTrace("system", "Paused via dashboard")
			} else {
				m.gateway.Machine().Send(statemachine.EventResumeRequested)
				m.AddTrace("system", "Resumed via dashboard")
			}
		case "ctrl+r":
			if m.paused {
				m.paused = false
				m.gateway.Machine().Send(statemachine.EventResumeRequested)
				m.AddTrace("system", "Resumed via dashboard")
			}
		case "ctrl+m":
			m.toggleModelBrowser()
		case "ctrl+y":
			m.AddTrace("system", "Syncing constitution from remote...")
			if changed, err := m.gateway.SyncConstitution(); err != nil {
				m.AddTrace("error", fmt.Sprintf("Sync failed: %v", err))
			} else if changed {
				m.AddTrace("system", "✓ Constitution updated from remote")
			} else {
				m.AddTrace("system", "→ Constitution already up to date")
			}
		case "ctrl+b":
			if !m.browserReady {
				m.browserReady = true
				go func() {
					if err := m.gateway.Computer().StartBrowser(); err != nil {
						m.AddTrace("error", fmt.Sprintf("Browser start failed: %v", err))
						m.browserReady = false
						return
					}
					m.AddTrace("system", "✓ Browser started (Playwright)")
				}()
			} else {
				m.AddTrace("system", "Browser already running")
			}
		case "/":
			m.commandMode = true
			m.commandBuffer = ""
		case "up":
			if m.modelBrowserVisible {
				if m.modelBrowserIdx > 0 {
					m.modelBrowserIdx--
				}
			} else if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SelectPrev()
			}
		case "down":
			if m.modelBrowserVisible {
				if m.modelBrowserIdx < len(m.modelBrowserList)-1 {
					m.modelBrowserIdx++
				}
			} else if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SelectNext()
			}
		case "enter":
			if m.modelBrowserVisible {
				if m.modelBrowserIdx < len(m.modelBrowserList) {
					selected := m.modelBrowserList[m.modelBrowserIdx]
					if selected != m.gateway.ActiveModel() {
						if err := m.gateway.SwitchModel(selected); err == nil {
							m.AddTrace("system", fmt.Sprintf("Switched model → %s", selected))
						}
					}
					m.modelBrowserVisible = false
				}
			} else if m.reviewPanel != nil && m.reviewPanel.Visible() {
				if err := m.reviewPanel.ApproveSelected(); err == nil {
					m.AddTrace("system", "✓ Review: action approved")
				}
			}
		case "o":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SetScope(computer.ScopeOnce)
				m.AddTrace("system", "⚙ Review scope: once")
			}
		case "s":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SetScope(computer.ScopeSession)
				m.AddTrace("system", "⚙ Review scope: session")
			}
		case "p":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SetScope(computer.ScopePermanent)
				m.AddTrace("system", "⚙ Review scope: permanent (trust)")
			}
		case "esc":
			if m.modelBrowserVisible {
				m.modelBrowserVisible = false
			} else if m.reviewPanel != nil && m.reviewPanel.Visible() {
				if err := m.reviewPanel.RejectSelected(); err == nil {
					m.AddTrace("system", "✗ Review: action rejected")
				}
			}
		case "q":
			if m.paused {
				m.quitting = true
				return m, tea.Quit
			}
		case "tab":
			m.cycleModel(1)
		case "shift+tab":
			m.cycleModel(-1)
		}
	}
	m.missionQueue.Update(msg)
	m.systemStatus.Update(msg)
	return m, nil
}

var (
	lastEventLogModTime time.Time
	lastEventLogSize    int64
)

func (m *Model) pollEventLog() {
	if m.logPath == "" {
		return
	}
	info, err := os.Stat(m.logPath)
	if err != nil {
		return
	}
	if info.ModTime() == lastEventLogModTime && info.Size() == lastEventLogSize {
		return
	}
	lastEventLogModTime = info.ModTime()
	lastEventLogSize = info.Size()

	entries, err := persistence.LoadEventLog(m.logPath)
	if err != nil {
		return
	}
	for _, e := range entries.Entries() {
		if e.Sequence <= m.lastSeqRead {
			continue
		}
		label := fmt.Sprintf("%s", e.Event)
		var detail string
		if len(e.Data) > 0 {
			detail = string(e.Data)
			if len(detail) > 60 {
				detail = detail[:60] + "..."
			}
		}
		msg := fmt.Sprintf("[%s] %s", e.NewState.String(), label)
		if detail != "" {
			msg += " " + detail
		}
		m.traceItems = append(m.traceItems, TraceEntry{
			Timestamp: e.Timestamp,
			Message:   msg,
			Level:     "event",
		})
		m.lastSeqRead = e.Sequence
	}
}

func (m *Model) pollCheckpoint() {
	if m.cpPath == "" {
		return
	}
	cs := persistence.NewCheckpointStore(m.cpPath)
	if !cs.Exists() {
		return
	}
	cp, err := cs.Load()
	if err != nil {
		return
	}
	m.systemStatus.SetCheckpoint(cp)
}

func (m *Model) executeCommand() {
	cmd := strings.TrimSpace(m.commandBuffer)

	switch {
	case cmd == "" || cmd == "/":
		return

	case cmd == "/help" || cmd == "/h":
		m.AddTrace("system", "═ KEYS: Ctrl+N talk  ·  Ctrl+M model browser  ·  Tab cycle model  ·  Ctrl+P pause  ·  Ctrl+R resume")
		m.AddTrace("system", "═ KEYS: Ctrl+B browser  ·  Ctrl+Y sync  ·  Ctrl+S quit  ·  ↑↓ navigate  ·  Enter select  ·  Esc reject  ·  o/s/p scope  ·  q quit")
		m.AddTrace("system", "═ CMDS: /model <name>  ·  /list  ·  /stats  ·  /init  ·  /install-unsloth  ·  /help  ·  /quit")

	case cmd == "/install-unsloth" || cmd == "/iu":
		m.AddTrace("system", "═ Installing unsloth... this may take a few minutes")
		go func() {
			if err := installUnsloth(); err != nil {
				m.AddTrace("error", fmt.Sprintf("═ Install failed: %v", err))
			} else {
				m.AddTrace("system", "✓ Unsloth installed! Run /models to see available models, then /model <name> to switch")
			}
		}()

	case cmd == "/quit" || cmd == "/q" || cmd == "/exit":
		m.quitting = true
		m.gateway.Machine().Send(statemachine.EventShutdownRequested)

	case cmd == "/list" || cmd == "/ls":
		if m.gateway.Registry() != nil {
			reg := m.gateway.Registry()
			names := make([]string, 0, len(reg.Profiles))
			for name := range reg.Profiles {
				names = append(names, name)
			}
			sort.Strings(names)
			active := m.gateway.ActiveModel()
			var b strings.Builder
			b.WriteString(fmt.Sprintf("═ Models (%d):", len(names)))
			for _, name := range names {
				p, ok := reg.GetProfile(name)
				mark := " "
				if name == active {
					mark = "●"
				}
				prov := "?"
				if ok {
					prov = p.Provider
				}
				b.WriteString(fmt.Sprintf(" %s%s[%s]", mark, name, prov))
			}
			m.AddTrace("system", b.String())
		} else {
			m.AddTrace("system", "═ No model registry loaded")
		}

	case cmd == "/init":
		m.AddTrace("system", "═ Scanning project for AGENTS.md generation...")
		go func() {
			initPath := m.gateway.RepoPath()
			if initPath == "" || initPath == "." {
				initPath, _ = os.Getwd()
			}
			agentsFile := filepath.Join(initPath, "AGENTS.md")
			if _, err := os.Stat(agentsFile); err == nil {
				m.AddTrace("system", "  AGENTS.md already exists at "+agentsFile)
				return
			}
			// Scan project structure
			var lines []string
			lines = append(lines, "# CURSE Project Context")
			lines = append(lines, "")
			lines = append(lines, "Auto-generated by CURSE /init. Edit to guide the entity.")
			lines = append(lines, "")
			lines = append(lines, "## Project")
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("- Root: %s", initPath))
			// Check for common files
			for _, f := range []string{"go.mod", "package.json", "Cargo.toml", "pyproject.toml", "Gemfile", "build.gradle"} {
				if _, err := os.Stat(filepath.Join(initPath, f)); err == nil {
					lines = append(lines, fmt.Sprintf("- Detected: %s", f))
				}
			}
			lines = append(lines, "")
			lines = append(lines, "## Commands")
			lines = append(lines, "")
			for _, c := range [][2]string{
				{"build", "go build ./..."},
				{"test", "go test ./..."},
				{"lint", "go vet ./..."},
			} {
				if _, err := exec.LookPath(strings.Fields(c[1])[0]); err == nil {
					lines = append(lines, fmt.Sprintf("- %s: `%s`", c[0], c[1]))
				}
			}
			lines = append(lines, "")
			lines = append(lines, "## Conventions")
			lines = append(lines, "")
			lines = append(lines, "- Follow existing code style in the codebase")
			lines = append(lines, "- Write tests for new functionality")
			lines = append(lines, "- Keep functions focused and small")
			content := strings.Join(lines, "\n") + "\n"
			if err := os.WriteFile(agentsFile, []byte(content), 0644); err != nil {
				m.AddTrace("error", fmt.Sprintf("  Write failed: %v", err))
				return
			}
			m.AddTrace("system", fmt.Sprintf("✓ AGENTS.md created with %d lines", len(lines)))
		}()

	case cmd == "/stats" || cmd == "/st":
		reg := m.gateway.Registry()
		modelCount := 0
		if reg != nil {
			modelCount = len(reg.Profiles)
		}
		budgetRem := 0
		if m.gateway.Budget() != nil {
			budgetRem = int(m.gateway.Budget().Remaining())
		}
		memSnapshot := ""
		if m.gateway.Memory() != nil && m.gateway.Memory().Loaded() {
			memSnapshot = "●"
		} else {
			memSnapshot = "○"
		}
		m.AddTrace("system", fmt.Sprintf("═ Models: %d · Active: %s · State: %s · Step: %d · Budget: %d · Mem: %s · Uptime: %s",
			modelCount,
			m.gateway.ActiveModel(),
			m.gateway.Machine().State().String(),
			m.gateway.Machine().Step(),
			budgetRem,
			memSnapshot,
			time.Since(m.startTime).Round(time.Second).String()))

	case strings.HasPrefix(cmd, "/model "):
		name := strings.TrimSpace(cmd[7:])
		if name == "" {
			m.AddTrace("error", "═ Usage: /model <name> — use /list to see available models")
			return
		}
		if m.gateway.Registry() == nil {
			m.AddTrace("error", "═ No model registry loaded")
			return
		}
		if _, ok := m.gateway.Registry().GetProfile(name); !ok {
			m.AddTrace("error", fmt.Sprintf("═ Unknown model %q — use /list to see available models", name))
			return
		}
		if err := m.gateway.SwitchModel(name); err != nil {
			m.AddTrace("error", fmt.Sprintf("═ Switch failed: %v", err))
		} else {
			m.AddTrace("system", fmt.Sprintf("✓ Switched model → %s", name))
		}

	default:
		// Non-commands are treated as natural language
		m.chatBuffer = cmd
		m.executeChat()
		m.chatBuffer = ""
	}
}

func (m *Model) executeChat() {
	input := strings.TrimSpace(m.chatBuffer)
	if input == "" {
		return
	}
	m.AddTrace("user", fmt.Sprintf(">>> %s", input))
	m.AddTrace("system", "  processing...")

	go func() {
		adapter := m.gateway.Adapter()
		if adapter == nil {
			m.AddTrace("error", "  No active model — use /list to see available models, /model <name> to switch")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		prompt := &gateway.Prompt{
			Messages: []gateway.Message{
				{Role: gateway.RoleUser, Content: input},
			},
			MaxTokens: 4096,
		}

		resp, err := adapter.Send(ctx, prompt)
		if err != nil {
			m.AddTrace("error", fmt.Sprintf("  Response failed: %v", err))
			return
		}
		text := resp.Message.Content
		if len(text) > 400 {
			m.AddTrace("model", fmt.Sprintf("  %s", text[:400]))
			m.AddTrace("system", fmt.Sprintf("  ... (response truncated, full length: %d chars)", len(text)))
		} else {
			m.AddTrace("model", fmt.Sprintf("  %s", text))
		}
	}()
}

func (m *Model) toggleModelBrowser() {
	if m.modelBrowserVisible {
		m.modelBrowserVisible = false
		return
	}
	if m.gateway.Registry() != nil {
		reg := m.gateway.Registry()
		m.modelBrowserList = make([]string, 0, len(reg.Profiles))
		for name := range reg.Profiles {
			m.modelBrowserList = append(m.modelBrowserList, name)
		}
		active := m.gateway.ActiveModel()
		m.modelBrowserIdx = 0
		for i, name := range m.modelBrowserList {
			if name == active {
				m.modelBrowserIdx = i
				break
			}
		}
		m.modelBrowserVisible = true
	}
}

func (m *Model) cycleModel(direction int) {
	reg := m.gateway.Registry()
	if reg == nil || len(reg.Profiles) == 0 {
		return
	}
	names := make([]string, 0, len(reg.Profiles))
	for name := range reg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	active := m.gateway.ActiveModel()
	idx := -1
	for i, name := range names {
		if name == active {
			idx = i
			break
		}
	}
	if idx == -1 {
		idx = 0
	} else {
		idx = (idx + direction + len(names)) % len(names)
	}
	selected := names[idx]
	if selected == active {
		return
	}
	if err := m.gateway.SwitchModel(selected); err != nil {
		m.AddTrace("error", fmt.Sprintf("Switch failed: %v", err))
	} else {
		m.AddTrace("system", fmt.Sprintf("Switched model → %s", selected))
	}
}

func (m *Model) renderModelBrowser(overlayWidth int) string {
	if !m.modelBrowserVisible || len(m.modelBrowserList) == 0 {
		return ""
	}

	active := m.gateway.ActiveModel()
	reg := m.gateway.Registry()

	var lines []string
	headerColor := PulseColor(m.animFrame)

	title := lipgloss.NewStyle().
		Foreground(headerColor).
		Bold(true).
		Render(fmt.Sprintf("  ◈  SELECT MODEL  ◈  (↑↓ enter  esc  tab to cycle)"))
	lines = append(lines, "", title)
	lines = append(lines, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", overlayWidth-4)))

	for i, name := range m.modelBrowserList {
		profile, ok := reg.GetProfile(name)
		if !ok {
			continue
		}

		isActive := name == active
		cursor := "  "
		nameColor := ColorFg
		if isActive {
			cursor = "→ "
			nameColor = ColorAccent
		}
		if i == m.modelBrowserIdx && !isActive {
			cursor = "▸ "
			nameColor = PulseColor(m.animFrame)
		}

		var providerColor lipgloss.Color
		switch profile.Provider {
		case "codex":
			providerColor = ColorToxic
		case "grep":
			providerColor = ColorSuccess
		case "echo":
			providerColor = ColorFgSubtle
		case "eval":
			providerColor = ColorPsychic
		case "fortune":
			providerColor = ColorWarning
		case "system":
			providerColor = ColorSpiral
		case "ollama":
			providerColor = ColorSpiral
		case "openai-compatible":
			providerColor = ColorPsychic
		case "subprocess":
			providerColor = ColorWarning
		case "local-fallback":
			providerColor = ColorError
		default:
			providerColor = ColorFgSubtle
		}

		activeMark := ""
		if isActive {
			activeMark = lipgloss.NewStyle().Foreground(ColorSuccess).Render(" ● ACTIVE")
		}

		providerStr := lipgloss.NewStyle().Foreground(providerColor).Render(profile.Provider)
		modelStr := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(profile.Model)
		nameStr := lipgloss.NewStyle().Foreground(nameColor).Render(cursor + name)

		line := fmt.Sprintf("  %s  [%s]  %s%s", nameStr, providerStr, modelStr, activeMark)
		if isActive {
			line = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2a")).Render(line)
		} else if i == m.modelBrowserIdx {
			line = lipgloss.NewStyle().Background(lipgloss.Color("#0a0a18")).Render(line)
		}
		lines = append(lines, line)
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", overlayWidth-4)))

	hint := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("  ↑↓ navigate · enter switch · esc close  ")
	lines = append(lines, "", hint)

	panel := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(headerColor).
		Padding(0, 1).
		Width(overlayWidth - 2).
		Render(strings.Join(lines, "\n"))

	return panel
}

func (m *Model) View() string {
	f := m.animFrame

	if !m.ready {
		return ""
	}
	if m.boot() != BootPhaseActive {
		return m.bootView(m.width)
	}

	state := m.gateway.Machine().State().String()
	stateDot := DotSecure
	if state == "Error" {
		stateDot = DotError
	} else if state == "Running" {
		stateDot = DotProcessing
	}
	modelName := m.gateway.ActiveModel()
	if modelName == "" {
		modelName = "none"
	}

	title := TitleBar(version, modelName, state, stateDot, f)

	leftWidth := m.width * 42 / 100
	if leftWidth < 28 {
		leftWidth = 28
	}
	rightWidth := m.width - leftWidth - 8
	if rightWidth < 30 {
		rightWidth = 30
	}

	borderAccent := PulseColor(f)
	if m.paused || state == "Error" {
		borderAccent = ColorBorder
	}

	leftPane := renderPanel("ENTITY DIRECTIVES", m.missionQueue.View(leftWidth, f), leftWidth, borderAccent, state == "Running" && !m.paused, f)
	tracePane := renderPanel("ENTITY CONSCIOUSNESS", m.renderTrace(rightWidth, f), rightWidth, borderAccent, state == "Running" && !m.paused, f)
	sessionDuration := time.Since(m.startTime)
	systemPane := renderPanel("VITAL SIGNS", m.systemStatus.View(rightWidth, f, sessionDuration), rightWidth, ColorBorder, false, f)

	rightCol := lipgloss.JoinVertical(lipgloss.Top, tracePane, "\n", systemPane)
	body := lipgloss.JoinHorizontal(lipgloss.Top, "  ", leftPane, "  ", rightCol, "  ")

	var reviewView string
	if m.reviewPanel != nil && m.reviewPanel.Visible() {
		reviewView = m.reviewPanel.View(rightWidth, f)
	}

	sessionID := m.gateway.Machine().MissionID()
	if sessionID == "" {
		sessionID = "---"
	}
	reviewPending := 0
	if m.reviewPanel != nil {
		reviewPending = m.reviewPanel.PendingCount()
	}
	footerExtra := ""
	if reviewPending > 0 {
		footerExtra = fmt.Sprintf("  ⚠ %d review(s) pending", reviewPending)
	}
	footer := FooterStyled(sessionID, modelName, fmt.Sprintf("seq:%d", m.lastSeqRead), m.paused, f, footerExtra)

	// ── Live Status Bar ──
	profile, ok := m.gateway.Registry().ActiveProfile()
	ctxPct := 0
	if ok && profile.ContextWindow > 0 {
		steps := m.gateway.Machine().Step()
		ctxPct = steps * 3
		if ctxPct > 95 {
			ctxPct = 95
		}
	}
	statusModel := m.systemStatus
	liveBar := LiveStatusBar(f, modelName, ctxPct, time.Since(m.startTime))
	engPhase := statusModel.EnginePhase()
	engLine := EngineStatusLine(f, engPhase, statusModel.SkillCount(), statusModel.KnowledgeCount(), statusModel.HealerRecoveryRate())

	statusBar := lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("─", m.width-4))
	statusBarContent := lipgloss.JoinVertical(lipgloss.Top,
		statusBar,
		liveBar,
		engLine)

	quickBar := renderQuickBar(m.width-4, f)
	content := lipgloss.JoinVertical(lipgloss.Top, title, "\n", body, "\n", statusBarContent, "\n", quickBar, "\n", footer)
	if reviewView != "" {
		content = lipgloss.JoinVertical(lipgloss.Top, content, "\n", reviewView)
	}
	if m.chatMode {
		chatBar := m.renderChatBar(m.width - 4)
		content = lipgloss.JoinVertical(lipgloss.Top, content, "\n", chatBar)
	} else if m.commandMode {
		cmdBar := m.renderCommandBar(m.width - 4)
		content = lipgloss.JoinVertical(lipgloss.Top, content, "\n", cmdBar)
	}
	if m.modelBrowserVisible {
		browserOverlay := m.renderModelBrowser(m.width - 4)
		if browserOverlay != "" {
			overlayBox := lipgloss.NewStyle().
				Width(m.width).
				Align(lipgloss.Center).
				Render(browserOverlay)
			content = lipgloss.JoinVertical(lipgloss.Top, content, "\n\n", overlayBox)
		}
	}
	return content + "\n"
}

func (m *Model) renderCommandBar(width int) string {
	prompt := "/ "
	display := prompt + m.commandBuffer
	cursor := " "
	if time.Now().UnixMilli()/500%2 == 0 {
		cursor = "▌"
	}
	display += cursor

	bar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PulseColor(m.animFrame)).
		Padding(0, 1).
		Width(width - 2).
		Render(lipgloss.NewStyle().Foreground(ColorAccent).Render(display))

	return bar
}

func (m *Model) renderChatBar(width int) string {
	prefix := ">>> "
	display := prefix + m.chatBuffer
	cursor := " "
	if time.Now().UnixMilli()/500%2 == 0 {
		cursor = "▌"
	}
	display += cursor

	bar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorAccent).
		Padding(0, 1).
		Width(width - 2).
		Render(lipgloss.NewStyle().Foreground(ColorPsychic).Render(display))

	return bar
}

func installUnsloth() error {
	python := findPython()
	if python == "" {
		return fmt.Errorf("python3 not found — install Python first: https://python.org")
	}
	cmd := exec.Command(python, "-m", "pip", "install", "unsloth", "transformers", "torch", "accelerate")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func findPython() string {
	for _, name := range []string{"python3", "python"} {
		path, err := exec.LookPath(name)
		if err == nil {
			return path
		}
	}
	return ""
}

func (m *Model) renderTrace(width int, frame int) string {
	now := time.Now()
	m.traceMu.Lock()
	total := len(m.traceItems)
	start := 0
	if total > m.maxVisible {
		start = total - m.maxVisible
	}
	snapshot := make([]TraceEntry, total-start)
	copy(snapshot, m.traceItems[start:])
	m.traceMu.Unlock()

	items := make([]string, 0, len(snapshot))
	for _, t := range snapshot {
		age := now.Sub(t.Timestamp)
		items = append(items, TraceItemStyled(t.Timestamp, t.Message, age, width-2, frame))
	}
	content := strings.Join(items, "\n")
	if content == "" {
		content = lipgloss.NewStyle().Foreground(ColorFgInactive).Render("  polling event.log for activity...")
	}
	return content
}

func (m *Model) AddTrace(level, msg string) {
	m.traceMu.Lock()
	m.traceItems = append(m.traceItems, TraceEntry{
		Timestamp: time.Now().UTC(),
		Message:   msg,
		Level:     level,
	})
	m.traceMu.Unlock()
}

func (m *Model) Quitting() bool {
	return m.quitting
}

// ── Panel Renderer ──────────────────────────────────────────

func renderPanel(header, content string, width int, accent lipgloss.Color, active bool, frame int) string {
	panelStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderForeground(accent).
		Padding(0, 1)
	head := PanelHeader(header, width, accent, frame)
	return panelStyle.Render(head + "\n" + content)
}
