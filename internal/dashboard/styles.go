package dashboard

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ── Cosmic Cyberpunk Identity ──────────────────────────────
//   A radical shift from "Embers" to "Void & Neon".
//   Deep obsidian, electric ultraviolet, and supernova cyan.
//   Geometric-abstract patterns. High contrast. Cybernetic elegance.

// ── Semantic Color Roles ────────────────────────────────────
var (
	ColorBg         = lipgloss.Color("#050508")
	ColorFg         = lipgloss.Color("#f0f0f5")
	ColorFgBright   = lipgloss.Color("#ffffff")
	ColorFgSubtle   = lipgloss.Color("#4a4a6a")
	ColorFgInactive = lipgloss.Color("#1a1a2a")
	ColorBorder     = lipgloss.Color("#1e1e30")
	ColorBorderDim  = lipgloss.Color("#0a0a12")

	// Signature — Ultraviolet / Cyber Neon
	ColorAccent     = lipgloss.Color("#bd00ff") // Electric Violet
	ColorAccentDim  = lipgloss.Color("#7a00aa")
	ColorAccentPulse = lipgloss.Color("#e066ff")

	// Cyber Patterns — Cyan / Magenta / Deep Space
	ColorCyan       = lipgloss.Color("#00f2ff") // Supernova Cyan
	ColorMagenta    = lipgloss.Color("#ff00ff") // Cyber Magenta
	ColorVoid       = lipgloss.Color("#020205")
	ColorNebula     = lipgloss.Color("#3a0088")

	// Semantic signals
	ColorSuccess    = lipgloss.Color("#00ffaa")
	ColorWarning    = lipgloss.Color("#ffaa00")
	ColorError      = lipgloss.Color("#ff0055")
	ColorProcessing = lipgloss.Color("#00f2ff")
	ColorPsychic    = lipgloss.Color("#ff00ff") // Alias for Magenta
	ColorSpiral     = lipgloss.Color("#00f2ff")  // Alias for Cyan
)

// ── Geometric Entity Mark (4 breathing frames) ──────────────
//   A geometric, minimalist mark representing a "Core".
//   Abstract hexagonal/diamond pulsing.

var entityFrames = [4][]string{
	{ // Frame 0 — stable
		`      ◢█████◣      `,
		`     ◢◤     ◥◣     `,
		`     █  ◈    █     `,
		`     ◥◣     ◢◤     `,
		`      ◥█████◤      `,
	},
	{ // Frame 1 — expanding
		`      ◢███████◣      `,
		`     ◢◤         ◥◣     `,
		`     █    ◈     █     `,
		`     ◥◣         ◢◤     `,
		`      ◥███████◤      `,
	},
	{ // Frame 2 — contracting
		`       ◢███◣       `,
		`      ◢◤   ◥◣      `,
		`      █ ◈  █      `,
		`      ◥◣   ◢◤      `,
		`       ◥███◤       `,
	},
	{ // Frame 3 — shimmering
		`      ◢░░░░░◣      `,
		`     ◢◤░░░░░◥◣     `,
		`     █  ◈   █     `,
		`     ◥◣░░░░░◢◤     `,
		`      ◥░░░░░◤      `,
	},
}

// EntityMark returns the geometric entity logo for a given animation frame.
func EntityMark(frame int) []string {
	return entityFrames[frame%4]
}

// ── Cyber Title Text ────────────────────────────────────────
//   A sharp, monospaced, futuristic "CURSE" title.

var CurseTitle = []string{
	` ▟▛▀▀▜▙  ▟▛▀▀▜▙  ▟▛▀▀▜▙  ▟▛▀▀▜▙  ▟▛▀▀▜▙`,
	` ▜▙▄▄▟▛  ▜▙▄▄▟▛  ▜▙▄▄▟▛  ▜▙▄▄▟▛  ▜▙▄▄▟▛`,
}

// ── Spinner ─────────────────────────────────────────────────
var spinnerFrames = [4]string{"▱", "▰", "▱", "▰"}

func Spinner(f int) string { return spinnerFrames[f%4] }

// ── Status Dots ─────────────────────────────────────────────
type DotStatus int

const (
	DotSecure     DotStatus = iota
	DotProcessing
	DotError
)

func StatusDot(s DotStatus, frame int) string {
	switch s {
	case DotSecure:
		p := []string{"◈", "◇", "◈", "◇", "◈", "◇", "◈", "◇"}
		return lipgloss.NewStyle().Foreground(ColorCyan).Render(p[frame%8])
	case DotProcessing:
		return lipgloss.NewStyle().Foreground(ColorAccent).Render(Spinner(frame))
	case DotError:
		if frame%2 == 0 {
			return lipgloss.NewStyle().Foreground(ColorError).Render("▰")
		}
		return lipgloss.NewStyle().Foreground(ColorError).Render("▱")
	default:
		return lipgloss.NewStyle().Foreground(ColorFgInactive).Render("·")
	}
}

// ── Glitch ──────────────────────────────────────────────────
func Glitch(s string, frame int) string {
	if rand.Intn(100) > 95 { // Reduced frequency for "professional" feel
		return s
	}
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	n := rand.Intn(len(runes))
	noise := []rune{'▙', '▟', '▚', '▞', '▗', '▖', '▘', '▙'}
	runes[n] = noise[rand.Intn(len(noise))]
	return string(runes)
}

// ── Pulse ───────────────────────────────────────────────────
var pulseCycle = []lipgloss.Color{
	ColorAccent,
	ColorCyan,
	ColorMagenta,
	ColorAccent,
}

func PulseColor(frame int) lipgloss.Color {
	return pulseCycle[frame%len(pulseCycle)]
}

// ── Title Bar ───────────────────────────────────────────────

func TitleBar(version, model, state string, dot DotStatus, frame int) string {
	dotStr := StatusDot(dot, frame)
	machineState := strings.ToUpper(state)

	var stateColor lipgloss.Color
	switch state {
	case "Running":
		stateColor = ColorCyan
	case "Error":
		stateColor = ColorError
	case "Paused":
		stateColor = ColorWarning
	case "Syncing":
		stateColor = ColorMagenta
	default:
		stateColor = ColorFgSubtle
	}

	accent := PulseColor(frame)
	barStyle := lipgloss.NewStyle().
		Background(accent).
		Foreground(ColorBg).
		Bold(true).
		Padding(0, 2)

	stateLabel := lipgloss.NewStyle().
		Foreground(stateColor).
		Background(accent).
		Render(machineState)

	content := fmt.Sprintf("  %s  ◈ CURSE // %s  │  %s  │  %s  ",
		dotStr, version, model, stateLabel)
	return barStyle.Render(content)
}

// ── Panel Header ────────────────────────────────────────────

func PanelHeader(title string, width int, accent lipgloss.Color, frame int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(accent).
		Bold(true)
	available := width - len(title) - 6
	if available < 2 {
		available = 2
	}
	dots := strings.Repeat("·", available-2)
	spinner := Spinner(frame)
	header := fmt.Sprintf(" %s  %s %s ",
		spinner,
		titleStyle.Render(title),
		lipgloss.NewStyle().Foreground(ColorFgInactive).Render(dots))
	return lipgloss.NewStyle().Foreground(accent).Render("▛" + header + "▜")
}

// ── Trace ───────────────────────────────────────────────────

func TraceItemStyled(ts time.Time, msg string, age time.Duration, width int, frame int) string {
	var timeColor, arrowColor, msgColor lipgloss.Color
	switch {
	case age < 3*time.Second:
		timeColor = ColorFgSubtle
		arrowColor = ColorCyan
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
	arrow := " ❯"
	if strings.HasPrefix(msg, "═══") {
		arrow = " ─"
		msgColor = ColorMagenta
	}
	msg = Glitch(msg, frame)
	prefix := lipgloss.NewStyle().Foreground(timeColor).Render(t) +
		lipgloss.NewStyle().Foreground(arrowColor).Render(arrow) +
		lipgloss.NewStyle().Foreground(msgColor).Render(" "+msg)
	if len(prefix) > width-2 {
		prefix = prefix[:width-5] + "..."
	}
	return prefix
}

// ── Kanban ──────────────────────────────────────────────────
func KanbanCardStyled(id, task, status string, active bool, width int, frame int) string {
	borderClr := ColorBorder
	if active {
		borderClr = PulseColor(frame)
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderClr).
		Padding(0, 1).
		Width(width - 2)
	label := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(id)
	taskLine := lipgloss.NewStyle().Foreground(ColorFg).Render(task)

	var statusColor lipgloss.Color
	switch status {
	case "completed":
		statusColor = ColorSuccess
	case "in_progress", "active":
		statusColor = ColorCyan
	case "blocked", "failed":
		statusColor = ColorError
	default:
		statusColor = ColorAccentDim
	}
	statusLine := lipgloss.NewStyle().Foreground(statusColor).Render(status)
	return style.Render(label + "\n" + taskLine + "\n" + statusLine)
}

func KanbanColumnStyled(title string, color lipgloss.Color, cards []string, width int) string {
	titleLine := lipgloss.NewStyle().Foreground(color).Bold(true).Padding(0, 1).Render(title)
	items := strings.Join(cards, "\n")
	return titleLine + "\n" + items
}

// ── System Status ───────────────────────────────────────────

func StatusLineStyled(label, value string, dot DotStatus, frame int) string {
	return fmt.Sprintf(" %s  %-14s%s",
		StatusDot(dot, frame),
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(label+":"),
		lipgloss.NewStyle().Foreground(ColorFg).Render(value),
	)
}

// ── Footer ──────────────────────────────────────────────────
func FooterStyled(sessionID, model, cpInfo string, paused bool, frame int, extra ...string) string {
	pauseLabel := "▶ ACTIVE"
	pauseColor := ColorCyan
	if paused {
		pauseLabel = "⏸ PAUSED"
		pauseColor = ColorWarning
	}
	pauseStyle := lipgloss.NewStyle().Foreground(pauseColor).Render(pauseLabel)

	spinner := Spinner(frame)
	left := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(
		fmt.Sprintf("  %s  / cmd  Ctrl+M model  Ctrl+P pause  Ctrl+B browse  Ctrl+Y sync  Ctrl+S quit  ", spinner),
	)
	right := fmt.Sprintf("  %s  │  %s  │  %s  ",
		pauseStyle,
		lipgloss.NewStyle().Foreground(ColorFgInactive).Render(sessionID),
		lipgloss.NewStyle().Foreground(ColorFgInactive).Render(cpInfo),
	)
	if len(extra) > 0 && extra[0] != "" {
		warnStyle := lipgloss.NewStyle().Foreground(ColorWarning).Render(extra[0])
		right += "  " + warnStyle
	}
	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render("│")
	return left + sep + right
}

// ── Box Helpers ─────────────────────────────────────────────

func boxTop(width int, color lipgloss.Color) string {
	if width < 4 { width = 4 }
	inner := strings.Repeat("─", width-2)
	return lipgloss.NewStyle().Foreground(color).Render("▛" + inner + "▜")
}

func boxBottom(width int, color lipgloss.Color) string {
	if width < 4 { width = 4 }
	inner := strings.Repeat("─", width-2)
	return lipgloss.NewStyle().Foreground(color).Render("▙" + inner + "▟")
}

func boxLine(width int, color lipgloss.Color) string {
	if width < 4 { width = 4 }
	inner := strings.Repeat("─", width-2)
	return lipgloss.NewStyle().Foreground(color).Render("▙" + inner + "▟")
}

func boxContent(line string, width int, color lipgloss.Color) string {
	content := lipgloss.NewStyle().Width(width - 2).Render(line)
	return lipgloss.NewStyle().Foreground(color).Render("▙") + content + lipgloss.NewStyle().Foreground(color).Render("▟")
}
