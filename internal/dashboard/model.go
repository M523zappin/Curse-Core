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
}

type TraceEntry struct {
	Timestamp time.Time
	Message   string
	Level     string
}

var version = "v0.1"

func NewModel(gw *gateway.Gateway) *Model {
	return &Model{
		gateway:      gw,
		traceItems:   make([]TraceEntry, 0),
		paused:       false,
		missionQueue: NewMissionQueueModel(gw.Queue()),
		systemStatus: NewSystemStatusModel(gw),
		maxVisible:   25,
		lastSeqRead:  0,
	}
}

func (m *Model) SetLogPaths(logPath, cpPath string) {
	m.logPath = logPath
	m.cpPath = cpPath
}

func (m *Model) Init() tea.Cmd {
	m.AddTrace("system", "Gateway initialized, awaiting mission")
	return tea.Batch(
		m.pollLog(),
		m.pulseTicker(),
	)
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
		return m, tea.Batch(m.pollLog(), m.pulseTicker())

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
	if !m.ready {
		return "Initializing Curse Gateway..."
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

	// Footer
	sessionID := m.gateway.Machine().MissionID()
	if sessionID == "" {
		sessionID = "---"
	}
	footer := FooterStyled(sessionID, modelName, fmt.Sprintf("seq:%d", m.lastSeqRead), m.paused)

	return lipgloss.JoinVertical(lipgloss.Top, title, "\n", body, "\n", footer) + "\n"
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
