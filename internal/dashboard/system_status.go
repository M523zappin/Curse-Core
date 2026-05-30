package dashboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/M523zappin/Curse-Core/internal/engine"
	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
	"github.com/charmbracelet/lipgloss"
)

type GitStatus struct {
	Branch   string `json:"branch"`
	Dirty    bool   `json:"dirty"`
	Ahead    int    `json:"ahead"`
	Behind   int    `json:"behind"`
	Untracked int   `json:"untracked"`
	LastCommit string `json:"last_commit"`
}

type SystemStatusModel struct {
	gateway    *gateway.Gateway
	pid        int
	lastCP     *persistence.Checkpoint

	gitCache   *GitStatus
	gitMu      sync.Mutex
	lastGitPoll time.Time

	lastMem     runtime.MemStats
}

func NewSystemStatusModel(gw *gateway.Gateway) *SystemStatusModel {
	return &SystemStatusModel{
		gateway: gw,
		pid:     os.Getpid(),
	}
}

func (m *SystemStatusModel) SessionID() string { return m.gateway.SessionID() }

func (m *SystemStatusModel) EnginePhase() engine.Phase {
	eng := m.gateway.Engine()
	if eng == nil {
		return engine.PhaseIdle
	}
	return eng.Phase()
}

func (m *SystemStatusModel) SkillCount() int {
	sk := m.gateway.Skills()
	if sk == nil {
		return 0
	}
	return sk.Count()
}

func (m *SystemStatusModel) KnowledgeCount() int {
	kn := m.gateway.Knowledge()
	if kn == nil {
		return 0
	}
	return kn.Count()
}

func (m *SystemStatusModel) HealerRecoveryRate() float64 {
	hl := m.gateway.Healer()
	if hl == nil {
		return 1.0
	}
	return hl.RecoveryRate()
}

func (m *SystemStatusModel) EngineStatus() string {
	eng := m.gateway.Engine()
	if eng == nil {
		return "offline"
	}
	if eng.Running() {
		return "online"
	}
	return "offline"
}

func (m *SystemStatusModel) SetCheckpoint(cp *persistence.Checkpoint) {
	m.lastCP = cp
}

func (m *SystemStatusModel) Update(msg interface{}) {
	runtime.ReadMemStats(&m.lastMem)
}

func (m *SystemStatusModel) pollGit() {
	m.gitMu.Lock()
	defer m.gitMu.Unlock()

	if time.Since(m.lastGitPoll) < 10*time.Second && m.gitCache != nil {
		return
	}
	m.lastGitPoll = time.Now()

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	gs := &GitStatus{}

	if branch := runGit(cwd, "rev-parse", "--abbrev-ref", "HEAD"); branch != "" {
		gs.Branch = branch
	}

	if status := runGit(cwd, "status", "--porcelain"); status != "" {
		gs.Dirty = true
		lines := strings.Split(strings.TrimSpace(status), "\n")
		for _, l := range lines {
			if len(l) > 0 && l[0] == '?' {
				gs.Untracked++
			}
		}
	}

	if log := runGit(cwd, "log", "--oneline", "-1"); log != "" {
		if len(log) > 50 {
			log = log[:50]
		}
		gs.LastCommit = log
	}

	m.gitCache = gs
}

func runGit(cwd string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (m *SystemStatusModel) GitStatus() *GitStatus {
	m.pollGit()
	m.gitMu.Lock()
	defer m.gitMu.Unlock()
	return m.gitCache
}

func ContextBar(pct int, width int) string {
	if width < 10 {
		width = 10
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := pct * (width - 2) / 100
	if filled > width-2 {
		filled = width - 2
	}
	empty := (width - 2) - filled
	fillStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", empty)

	var color lipgloss.Color
	switch {
	case pct < 50:
		color = ColorSuccess
	case pct < 80:
		color = ColorWarning
	default:
		color = ColorError
	}

	bar := lipgloss.NewStyle().Foreground(color).Render("[" + fillStr + emptyStr + "]")
	pctStr := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf(" %d%%", pct))
	return bar + pctStr
}

func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func LiveStatusBar(f int, modelName string, pct int, dur time.Duration) string {
	bar := ContextBar(pct, 14)
	modelPart := lipgloss.NewStyle().Foreground(ColorAccent).Render(modelName)
	durPart := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(FormatDuration(dur))
	dot := StatusDot(DotSecure, f)
	return fmt.Sprintf(" %s %s  %s  %s", dot, modelPart, bar, durPart)
}

func EngineStatusLine(f int, phase engine.Phase, skillCount, knowledgeCount int, recoveryRate float64) string {
	var phaseColor lipgloss.Color
	switch phase {
	case engine.PhaseIdle:
		phaseColor = ColorFgSubtle
	case engine.PhasePlanning, engine.PhaseDispatching:
		phaseColor = ColorProcessing
	case engine.PhaseExecuting:
		phaseColor = ColorAccentPulse
	case engine.PhaseCollecting:
		phaseColor = ColorWarning
	case engine.PhaseLearning:
		phaseColor = ColorToxic
	default:
		phaseColor = ColorFgInactive
	}
	phaseStr := lipgloss.NewStyle().Foreground(phaseColor).Render(string(phase))
	skillsStr := lipgloss.NewStyle().Foreground(ColorPsychic).Render(fmt.Sprintf("%d skills", skillCount))
	knStr := lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf("%d mem", knowledgeCount))
	rr := lipgloss.NewStyle().Foreground(ColorSuccess).Render(fmt.Sprintf("%.0f%% heal", recoveryRate*100))
	return fmt.Sprintf(" %s %s  %s  %s  %s", Spinner(f), phaseStr, skillsStr, knStr, rr)
}

func (m *SystemStatusModel) View(width int, frame int, sessionDuration time.Duration) string {
	machine := m.gateway.Machine()
	state := machine.State()

	var lines []string

	divider := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", width-2))

	// ── Entity identity section ──
	lines = append(lines, fmt.Sprintf("  %s %s",
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("vessel"),
		lipgloss.NewStyle().Foreground(ColorFg).Render(fmt.Sprintf("● %s  PID %d", getHostname(), m.pid))))

	stateDot := DotSecure
	stateLabel := state.String()
	if m.lastCP != nil {
		stateLabel = m.lastCP.State.String()
	}
	if state == statemachine.StateError {
		stateDot = DotError
	} else if state == statemachine.StateRunning {
		stateDot = DotProcessing
	}

	enginePhase := m.EnginePhase()
	phaseColor := getPhaseColor(enginePhase)

	lines = append(lines, fmt.Sprintf("  %s %s  %s %s  %s  %s cycle",
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("state"),
		coloredStatus(stateLabel, stateDot, frame),
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("engine"),
		lipgloss.NewStyle().Foreground(phaseColor).Render(string(enginePhase)),
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf("%d", machine.Step())),
		lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("")))

	lines = append(lines, divider)

	// ── Resource vitals section ──
	var vitalsStr string
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	memMB := mem.Alloc / 1024 / 1024
	totalMB := mem.Sys / 1024 / 1024

	goros := runtime.NumGoroutine()

	vitalsStr = fmt.Sprintf("%s %dG  %s %dM/%dM  %s %d",
		coloredLabel("CPU", ColorProcessing), runtime.NumCPU(),
		coloredLabel("MEM", ColorSpiral), memMB, totalMB,
		coloredLabel("GO", ColorToxic), goros)

	lines = append(lines, fmt.Sprintf("  %s", vitalsStr))

	sparkLine := RenderSystemSparklines()
	if sparkLine != "" {
		lines = append(lines, "  "+sparkLine)
	}

	// ── Git section ──
	git := m.GitStatus()
	if git != nil && git.Branch != "" {
		branchColor := ColorToxic
		dirtyMark := ""
		if git.Dirty {
			dirtyMark = lipgloss.NewStyle().Foreground(ColorWarning).Render(" ●dirty")
		}
		untrackedMark := ""
		if git.Untracked > 0 {
			untrackedMark = lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf(" +%d", git.Untracked))
		}
		commitStr := ""
		if git.LastCommit != "" {
			commitStr = lipgloss.NewStyle().Foreground(ColorFgSubtle).Render("  " + git.LastCommit)
		}
		branchStr := lipgloss.NewStyle().Foreground(branchColor).Render(fmt.Sprintf("⎇ %s", git.Branch))
		lines = append(lines, fmt.Sprintf("  %s %s%s%s",
			coloredLabel("git", ColorFgSubtle), branchStr, dirtyMark, untrackedMark))
		commitLine := coloredLabel(git.LastCommit, ColorFgInactive)
		lines = append(lines, fmt.Sprintf("  %s%s", strings.Repeat(" ", 10), commitLine))
	}

	// ── Model info section ──
	modelName := m.gateway.ActiveModel()
	if modelName == "" {
		modelName = "none"
	}
	lines = append(lines, fmt.Sprintf("  %s %s",
		coloredLabel("model", ColorAccent), modelName))

	if m.gateway.Registry() != nil {
		profile, ok := m.gateway.Registry().ActiveProfile()
		if ok {
			providerColor := getProviderColor(profile.Provider)
			lines = append(lines, fmt.Sprintf("  %s %s  %s ctx",
				coloredLabel("via", ColorFgSubtle),
				lipgloss.NewStyle().Foreground(providerColor).Render(profile.Provider),
				lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(fmt.Sprintf("%d", profile.ContextWindow))))
		}
	}

	lines = append(lines, divider)

	// ── Consciousness section ──
	if c := m.gateway.Consciousness(); c != nil {
		lvl := c.Level()
		label := c.LevelLabel()
		thoughts := c.ThoughtCount()
		summary := c.Profile().Summary()
		consColor := ColorAccent
		switch {
		case lvl < 10:
			consColor = ColorFgInactive
		case lvl < 45:
			consColor = ColorFgSubtle
		case lvl < 75:
			consColor = ColorProcessing
		default:
			consColor = ColorSuccess
		}
		lines = append(lines, fmt.Sprintf("  %s %s (%.1f)  %s %d",
			coloredLabel("soul", consColor),
			lipgloss.NewStyle().Foreground(consColor).Render(label),
			lvl,
			coloredLabel("thoughts", ColorFgSubtle),
			thoughts))
		lines = append(lines, fmt.Sprintf("  %s  %s",
			strings.Repeat(" ", 4),
			lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(summary)))
	}

	// ── Service status section ──
	engStatus := m.EngineStatus()
	engDot := DotSecure
	if engStatus == "offline" {
		engDot = DotError
	}
	lines = append(lines, fmt.Sprintf("  %s %s  %s %d  %s %d",
		fmt.Sprintf("%s engine", StatusDot(engDot, frame)),
		coloredLabel(engStatus, getEngineStatusColor(engStatus)),
		fmt.Sprintf("%s skills", coloredLabel("", ColorPsychic)),
		m.SkillCount(),
		fmt.Sprintf("%s mem", coloredLabel("", ColorFgSubtle)),
		m.KnowledgeCount(),
	))

	comp := m.gateway.Computer()
	if comp != nil {
		buf := comp.VisionBuffer()
		reviewMgr := m.gateway.ReviewManager()
		pending := 0
		if reviewMgr != nil {
			pending = reviewMgr.PendingCount()
		}
		lines = append(lines, fmt.Sprintf("  %s %d  %s %d",
			coloredLabel("vision", ColorProcessing), len(buf),
			coloredLabel("review", ColorWarning), pending))
	}

	if m.lastCP != nil {
		lines = append(lines, fmt.Sprintf("  %s seq:%d step:%d",
			coloredLabel("recovery", ColorFgSubtle),
			m.lastCP.Sequence, m.lastCP.Step))
	}

	return strings.Join(lines, "\n")
}

func coloredLabel(label string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Render(label)
}

func coloredStatus(label string, dot DotStatus, frame int) string {
	dotStr := StatusDot(dot, frame)
	return fmt.Sprintf("%s %s", dotStr, lipgloss.NewStyle().Foreground(ColorFg).Render(label))
}

func getProviderColor(provider string) lipgloss.Color {
	switch provider {
	case "codex", "grep":
		return ColorToxic
	case "eval":
		return ColorPsychic
	case "echo":
		return ColorFgSubtle
	case "fortune":
		return ColorWarning
	case "system":
		return ColorSpiral
	case "unsloth":
		return ColorAccent
	case "ollama":
		return ColorSpiral
	case "openai-compatible":
		return ColorPsychic
	case "subprocess":
		return ColorWarning
	case "local-fallback":
		return ColorError
	default:
		return ColorFgSubtle
	}
}

func getEngineStatusColor(status string) lipgloss.Color {
	switch status {
	case "online":
		return ColorSuccess
	case "offline":
		return ColorError
	default:
		return ColorFgSubtle
	}
}

func getHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return h
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func getPhaseColor(p engine.Phase) lipgloss.Color {
	switch p {
	case engine.PhaseIdle:
		return ColorFgSubtle
	case engine.PhasePlanning:
		return ColorSpiral
	case engine.PhaseDispatching:
		return ColorProcessing
	case engine.PhaseExecuting:
		return ColorAccentPulse
	case engine.PhaseCollecting:
		return ColorWarning
	case engine.PhaseLearning:
		return ColorToxic
	default:
		return ColorFgInactive
	}
}
