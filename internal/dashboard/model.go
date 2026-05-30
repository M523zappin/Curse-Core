package dashboard

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
)

type pulseTick time.Time

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
	pulseFrame   int
	reviewPanel  *ReviewPanelModel
	browserReady bool
}

type TraceEntry struct {
	Timestamp time.Time
	Message   string
	Level     string
}

var version = "v1.0.0"

func NewModel(gw *gateway.Gateway) *Model {
	m := &Model{
		gateway:      gw,
		traceItems:   make([]TraceEntry, 0),
		paused:       false,
		missionQueue: NewMissionQueueModel(gw.Queue()),
		systemStatus: NewSystemStatusModel(gw),
		maxVisible:   25,
		lastSeqRead:  0,
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
		m.pollLog(),
		m.pulseTicker(),
	)
}

type splashMsg struct{}

func (m *Model) showSplash(ready bool) {
	if ready {
		return
	}
	m.AddTrace("entity", "═══ CURSE Cognitive Unified Runtime System Entity ═══")
	m.AddTrace("entity", "State machine: 8 states · 15 events · SHA256 chain")
	m.AddTrace("entity", "Agent fleet:  8 specialized roles (security, refactor, infra, ...)")
	m.AddTrace("entity", "Computer:     Playwright browser + desktop OS control")
	m.AddTrace("entity", "Healing:      Auto root-cause analysis + recovery loops")
	m.AddTrace("entity", "Knowledge:    Live ADR/debug index → .curse/knowledge/")
	m.AddTrace("entity", "LSP:          Auto-detect gopls / typescript-language-server")
	m.AddTrace("entity", "Review:       HITL confirmation for destructive actions")
	m.AddTrace("system", "Awaiting mission — Ctrl+B to start browser, Ctrl+P to pause")
}

func (m *Model) pollLog() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return pulseTick(t)
	})
}

func (m *Model) pulseTicker() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
		return pulseTick(t)
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

	case pulseTick:
		m.pulseFrame++
		m.pollEventLog()
		m.pollCheckpoint()
		if m.reviewPanel != nil {
			m.reviewPanel.Update(msg)
		}
		return m, tea.Batch(m.pollLog(), m.pulseTicker())

	case splashMsg:
		m.showSplash(true)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+s":
			m.quitting = true
			m.gateway.Machine().Send(statemachine.EventShutdownRequested)
			return m, tea.Quit
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
			m.cycleModel()
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
				go func() {
					if err := m.gateway.Computer().StartBrowser(); err != nil {
						m.AddTrace("error", fmt.Sprintf("Browser start failed: %v", err))
						return
					}
					m.AddTrace("system", "✓ Browser started (Playwright)")
				}()
				m.browserReady = true
			} else {
				m.AddTrace("system", "Browser already running")
			}

		case "up":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SelectPrev()
			}
		case "down":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				m.reviewPanel.SelectNext()
			}
		case "enter":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				if err := m.reviewPanel.ApproveSelected(); err == nil {
					m.AddTrace("system", "✓ Review: action approved")
				}
			}
		case "esc":
			if m.reviewPanel != nil && m.reviewPanel.Visible() {
				if err := m.reviewPanel.RejectSelected(); err == nil {
					m.AddTrace("system", "✗ Review: action rejected")
				}
			}
		case "q":
			if m.paused {
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	m.missionQueue.Update(msg)
	m.systemStatus.Update(msg)
	return m, nil
}

func (m *Model) pollEventLog() {
	if m.logPath == "" {
		return
	}
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

func (m *Model) cycleModel() {
	if m.gateway.Registry() != nil {
		reg := m.gateway.Registry()
		profiles := make([]string, 0, len(reg.Profiles))
		for name := range reg.Profiles {
			profiles = append(profiles, name)
		}
		if len(profiles) == 0 {
			return
		}
		for i, name := range profiles {
			if name == m.gateway.ActiveModel() {
				next := (i + 1) % len(profiles)
				m.gateway.SwitchModel(profiles[next])
				m.AddTrace("system", fmt.Sprintf("Switched model → %s", profiles[next]))
				break
			}
		}
	}
}

func (m *Model) View() string {
	if !m.ready || len(m.traceItems) < 2 {
		m.showSplash(m.ready)
		if !m.ready {
			return SplashScreen(m.width)
		}
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

	// Title bar
	title := TitleBar(version, modelName, state, stateDot)

	// Layout dimensions
	leftWidth := m.width * 42 / 100
	if leftWidth < 28 {
		leftWidth = 28
	}
	rightWidth := m.width - leftWidth - 8
	if rightWidth < 30 {
		rightWidth = 30
	}

	// Determine active panel pulse colour
	borderAccent := PulseColor(m.pulseFrame)
	if m.paused || state == "Error" {
		borderAccent = ColorBorder
	}

	// Panels
	leftPane := renderPanel("MISSION QUEUE", m.missionQueue.View(leftWidth), leftWidth, borderAccent, state == "Running" && !m.paused)
	tracePane := renderPanel("REASONING TRACE / EVENT STREAM", m.renderTrace(rightWidth), rightWidth, borderAccent, state == "Running" && !m.paused)
	systemPane := renderPanel("SYSTEM STATUS", m.systemStatus.View(rightWidth), rightWidth, ColorBorder, false)

	// Assemble right column
	rightCol := lipgloss.JoinVertical(lipgloss.Top, tracePane, "\n", systemPane)

	// Body
	body := lipgloss.JoinHorizontal(lipgloss.Top, "  ", leftPane, "  ", rightCol, "  ")

	// Review Panel (HITL overlay)
	var reviewView string
	if m.reviewPanel != nil && m.reviewPanel.Visible() {
		reviewView = m.reviewPanel.View(rightWidth)
	}

	// Footer
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
	footer := FooterStyled(sessionID, modelName, fmt.Sprintf("seq:%d", m.lastSeqRead), m.paused, footerExtra)

	content := lipgloss.JoinVertical(lipgloss.Top, title, "\n", body, "\n", footer)
	if reviewView != "" {
		content = lipgloss.JoinVertical(lipgloss.Top, content, "\n", reviewView)
	}
	return content + "\n"
}

func (m *Model) renderTrace(width int) string {
	now := time.Now()
	items := make([]string, 0, len(m.traceItems))
	start := 0
	if len(m.traceItems) > m.maxVisible {
		start = len(m.traceItems) - m.maxVisible
	}
	for _, t := range m.traceItems[start:] {
		age := now.Sub(t.Timestamp)
		items = append(items, TraceItemStyled(t.Timestamp, t.Message, age, width-2))
	}
	content := strings.Join(items, "\n")
	if content == "" {
		content = lipgloss.NewStyle().Foreground(ColorFgInactive).Render("  polling event.log for activity...")
	}
	return content
}

func (m *Model) AddTrace(level, msg string) {
	m.traceItems = append(m.traceItems, TraceEntry{
		Timestamp: time.Now().UTC(),
		Message:   msg,
		Level:     level,
	})
}

func (m *Model) Quitting() bool {
	return m.quitting
}

// ── Panel Renderer ───────────────────────────────────────────────

func renderPanel(header, content string, width int, accent lipgloss.Color, active bool) string {
	panelStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderForeground(accent).
		Padding(0, 1)
	head := PanelHeader(header, width, accent)
	return panelStyle.Render(head + "\n" + content)
}
