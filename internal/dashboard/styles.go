package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ── Semantic Color Palette (cool theme) ──────────────────────────

var (
	ColorBg         = lipgloss.Color("#1a1b26")
	ColorFg         = lipgloss.Color("#c0caf5")
	ColorFgBright   = lipgloss.Color("#e0e0e0")
	ColorFgSubtle   = lipgloss.Color("#565f89")
	ColorFgInactive = lipgloss.Color("#3b4261")
	ColorAccent     = lipgloss.Color("#00d4ff")
	ColorAccentDim  = lipgloss.Color("#2aa0b0")
	ColorAccentTeal = lipgloss.Color("#00ffab")
	ColorSuccess    = lipgloss.Color("#00ff88")
	ColorWarning    = lipgloss.Color("#ffc107")
	ColorError      = lipgloss.Color("#ff6b80")
	ColorProcessing = lipgloss.Color("#ffaa00")
	ColorBorder     = lipgloss.Color("#2f3b54")
	ColorBorderDim  = lipgloss.Color("#1e2030")
)

// Pulse colours cycled during thinking state (slowed for subtlety)
var pulseCycle = []lipgloss.Color{
	lipgloss.Color("#00d4ff"),
	lipgloss.Color("#00e5ff"),
	lipgloss.Color("#00f5d4"),
	lipgloss.Color("#00e5ff"),
	lipgloss.Color("#00d4ff"),
}

// ── Sparkline / Status Dots ──────────────────────────────────────
type DotStatus int

const (
	DotSecure    DotStatus = iota // green – SHA256 validated
	DotProcessing                 // amber – actively processing
	DotError                      // red – failure
	DotIdle                       // dim grey – no activity
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
	content := fmt.Sprintf("  %s  CURSE  %s  │  Model: %s  │  %s  ", dotStr, version, model, state)
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
	// Dim-on-inactivity: newer = brighter, older = dimmer
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

func FooterStyled(sessionID, model, cpInfo string, paused bool) string {
	pauseLabel := "Running"
	if paused {
		pauseLabel = "● Paused"
	}
	left := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(
		fmt.Sprintf("  Ctrl+P pause  Ctrl+M model  Ctrl+S quit  "),
	)
	right := lipgloss.NewStyle().Foreground(ColorFgInactive).Render(
		fmt.Sprintf("  %s  │  %s  │  %s  ", pauseLabel, sessionID, cpInfo),
	)
	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render("│")
	return left + sep + right
}

// ── Pulsing colour selector ──────────────────────────────────────

func PulseColor(frame int) lipgloss.Color {
	return pulseCycle[frame%len(pulseCycle)]
}
