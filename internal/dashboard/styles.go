package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ── CURSE Persona: High-Contrast Minimalist Palette ──────────────
//   The terminal IS the entity. Every pixel communicates state.
//   Nothing is decorative — everything is signal.

var (
	// Core: deep void bg, sharp white fg — maximum contrast
	ColorBg         = lipgloss.Color("#000000")
	ColorFg         = lipgloss.Color("#e0e0e0")
	ColorFgBright   = lipgloss.Color("#ffffff")
	ColorFgSubtle   = lipgloss.Color("#808080")
	ColorFgInactive = lipgloss.Color("#404040")

	// Accent: neon cyan — the operating pulse
	ColorAccent     = lipgloss.Color("#00d4ff")
	ColorAccentDim  = lipgloss.Color("#0088aa")
	ColorAccentTeal = lipgloss.Color("#00ffab")

	// Semantic: green / amber / red — nothing else
	ColorSuccess    = lipgloss.Color("#00ff88")
	ColorWarning    = lipgloss.Color("#ffbb00")
	ColorError      = lipgloss.Color("#ff3344")
	ColorProcessing = lipgloss.Color("#ff8800")

	// Borders: near-invisible until active
	ColorBorder    = lipgloss.Color("#222222")
	ColorBorderDim = lipgloss.Color("#111111")
)

// Pulse: a single cyan beat, no rainbow
var pulseCycle = []lipgloss.Color{
	lipgloss.Color("#00d4ff"),
	lipgloss.Color("#00bbdd"),
	lipgloss.Color("#00d4ff"),
}

// ── Status Dots ──────────────────────────────────────────────────
//   Minimal. Three states only. No idle dot clutter.

type DotStatus int

const (
	DotSecure     DotStatus = iota // green — integrity confirmed
	DotProcessing                  // amber — actively working
	DotError                       // red — fault detected
)

func StatusDot(s DotStatus, pulse bool) string {
	switch s {
	case DotSecure:
		return lipgloss.NewStyle().Foreground(ColorSuccess).Render("●")
	case DotProcessing:
		if pulse {
			return lipgloss.NewStyle().Foreground(ColorProcessing).Render("◉")
		}
		return lipgloss.NewStyle().Foreground(ColorProcessing).Render("●")
	case DotError:
		return lipgloss.NewStyle().Foreground(ColorError).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(ColorFgInactive).Render("○")
	}
}

// ── Splash: CURSE entry identity ─────────────────────────────────

func SplashScreen(width int) string {
	if width < 50 {
		width = 50
	}
	logo := []string{
		"   ╔══════════════════════════════════════════════╗",
		"   ║              C U R S E                       ║",
		"   ║  Cognitive Unified Runtime System Entity     ║",
		"   ║                                              ║",
		"   ║  • State machine orchestration               ║",
		"   ║  • Crash-recoverable event chain             ║",
		"   ║  • Sub-agent fleet (8 domains)               ║",
		"   ║  • Computer controller (browser + desktop)   ║",
		"   ║  • Self-healing failure loop                 ║",
		"   ║  • Persistent knowledge index                ║",
		"   ║  • LSP-First diagnostics engine              ║",
		"   ║  • HITL review mode                          ║",
		"   ╚══════════════════════════════════════════════╝",
	}
	style := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Width(width).
		Align(lipgloss.Center)
	return style.Render(strings.Join(logo, "\n"))
}

// ── Box-Drawing Helpers ──────────────────────────────────────────

func boxTop(width int, color lipgloss.Color) string {
	if width < 4 {
		width = 4
	}
	inner := strings.Repeat("─", width-2)
	return lipgloss.NewStyle().Foreground(color).Render("╭" + inner + "╮")
}

func boxBottom(width int, color lipgloss.Color) string {
	if width < 4 {
		width = 4
	}
	inner := strings.Repeat("─", width-2)
	return lipgloss.NewStyle().Foreground(color).Render("╰" + inner + "╯")
}

func boxLine(width int, color lipgloss.Color) string {
	if width < 4 {
		width = 4
	}
	inner := strings.Repeat("─", width-2)
	return lipgloss.NewStyle().Foreground(color).Render("├" + inner + "┤")
}

func boxContent(line string, width int, color lipgloss.Color) string {
	content := lipgloss.NewStyle().Width(width - 2).Render(line)
	return lipgloss.NewStyle().Foreground(color).Render("│") + content + lipgloss.NewStyle().Foreground(color).Render("│")
}

// ── Title Bar ────────────────────────────────────────────────────

func TitleBar(version, model, state string, dot DotStatus) string {
	dotStr := StatusDot(dot, state == "Running")
	barStyle := lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBg).
		Bold(true).
		Padding(0, 2)
	content := fmt.Sprintf("  %s  CURSE  %s  │  %s  │  %s  ", dotStr, version, model, state)
	return barStyle.Render(content)
}

// ── Panel Header ─────────────────────────────────────────────────

func PanelHeader(title string, width int, accent lipgloss.Color) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)
	available := width - len(title) - 2
	if available < 2 {
		available = 2
	}
	dots := strings.Repeat("·", available-2)
	header := fmt.Sprintf(" %s %s ", titleStyle.Render(title), lipgloss.NewStyle().Foreground(ColorFgInactive).Render(dots))
	return lipgloss.NewStyle().Foreground(accent).Render("╭" + header + "╮")
}

// ── Trace Item ───────────────────────────────────────────────────

func TraceItemStyled(ts time.Time, msg string, age time.Duration, width int) string {
	var timeColor, arrowColor, msgColor lipgloss.Color
	switch {
	case age < 3*time.Second:
		timeColor = ColorFgSubtle
		arrowColor = ColorAccent
		msgColor = ColorFgBright
	case age < 15*time.Second:
		timeColor = ColorFgSubtle
		arrowColor = ColorAccentDim
		msgColor = ColorFg
	default:
		timeColor = ColorFgInactive
		arrowColor = ColorFgInactive
		msgColor = ColorFgSubtle
	}

	t := ts.Format("15:04:05")
	prefix := lipgloss.NewStyle().Foreground(timeColor).Render(t) +
		lipgloss.NewStyle().Foreground(arrowColor).Render(" ▶") +
		lipgloss.NewStyle().Foreground(msgColor).Render(" "+msg)
	if len(prefix) > width-2 {
		prefix = prefix[:width-5] + "..."
	}
	return prefix
}

// ── Kanban Card ──────────────────────────────────────────────────

func KanbanCardStyled(id, task, status string, active bool, width int) string {
	borderClr := ColorBorder
	if active {
		borderClr = ColorAccent
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderClr).
		Padding(0, 1).
		Width(width - 2)
	label := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(id)
	taskLine := lipgloss.NewStyle().Foreground(ColorFg).Render(task)
	statusLine := lipgloss.NewStyle().Foreground(ColorAccentDim).Render(status)
	return style.Render(label + "\n" + taskLine + "\n" + statusLine)
}

func KanbanColumnStyled(title string, color lipgloss.Color, cards []string, width int) string {
	titleLine := lipgloss.NewStyle().Foreground(color).Bold(true).Padding(0, 1).Render(title)
	items := strings.Join(cards, "\n")
	return titleLine + "\n" + items
}

// ── System Status Line ───────────────────────────────────────────

func StatusLineStyled(label, value string, dot DotStatus, pulse bool) string {
	return fmt.Sprintf(" %s  %-14s%s",
		StatusDot(dot, pulse),
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(label+":"),
		lipgloss.NewStyle().Foreground(ColorFg).Render(value),
	)
}

// ── Footer ───────────────────────────────────────────────────────

func FooterStyled(sessionID, model, cpInfo string, paused bool, extra ...string) string {
	pauseLabel := "RUNNING"
	if paused {
		pauseLabel = "● PAUSED"
	}
	left := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(
		fmt.Sprintf("  Ctrl+P pause  Ctrl+B browser  Ctrl+Y sync  Ctrl+S quit  "),
	)
	right := lipgloss.NewStyle().Foreground(ColorFgInactive).Render(
		fmt.Sprintf("  %s  │  %s  │  %s  ", pauseLabel, sessionID, cpInfo),
	)
	if len(extra) > 0 && extra[0] != "" {
		right += lipgloss.NewStyle().Foreground(ColorWarning).Render(extra[0])
	}
	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render("│")
	return left + sep + right
}

// ── Pulse ────────────────────────────────────────────────────────

func PulseColor(frame int) lipgloss.Color {
	return pulseCycle[frame%len(pulseCycle)]
}
