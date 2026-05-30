package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/M523zappin/Curse-Core/internal/computer"
	"github.com/charmbracelet/lipgloss"
)

type ReviewMode int

const (
	ReviewHidden ReviewMode = iota
	ReviewPending
	ReviewApproved
	ReviewRejected
)

type ReviewPanelModel struct {
	mode       ReviewMode
	pending    []computer.ReviewRequest
	selected   int
	reviewer   *computer.ReviewManager
	visible    bool
	lastUpdate time.Time
}

func NewReviewPanelModel(rm *computer.ReviewManager) *ReviewPanelModel {
	return &ReviewPanelModel{
		mode:       ReviewHidden,
		reviewer:   rm,
		visible:    false,
		lastUpdate: time.Now(),
	}
}

func (rp *ReviewPanelModel) Update(msg interface{}) {
	rp.pending = rp.reviewer.PendingReviews()
	if len(rp.pending) > 0 {
		rp.visible = true
		rp.mode = ReviewPending
	} else {
		rp.visible = false
		rp.mode = ReviewHidden
	}
	rp.lastUpdate = time.Now()
}

func (rp *ReviewPanelModel) View(width int, frame int) string {
	if !rp.visible || len(rp.pending) == 0 {
		return ""
	}

	if width < 30 {
		width = 30
	}

	var sections []string

	headerColor := ColorWarning
	if rp.mode == ReviewPending {
		headerColor = ColorError
	}

	spinner := Spinner(frame)
	header := lipgloss.NewStyle().
		Foreground(headerColor).
		Bold(true).
		Render(fmt.Sprintf("  %s  HUMAN OVERSIGHT — %d action(s) await review", spinner, len(rp.pending)))
	sections = append(sections, header)
	sections = append(sections, lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("═", width-2)))

	for i, req := range rp.pending {
		action := req.Action
		selected := i == rp.selected

		borderClr := ColorBorder
		if selected {
			borderClr = PulseColor(frame)
		}

		card := strings.Builder{}
		card.WriteString(fmt.Sprintf(" ◉ Action: %s\n", action.Type))
		card.WriteString(fmt.Sprintf("   Target: %s\n", truncateStr(action.Target, 40)))
		if action.Value != "" {
			card.WriteString(fmt.Sprintf("   Value:  %s\n", truncateStr(action.Value, 40)))
		}

		safetyLabel := "SAFE"
		safetyColor := ColorSuccess
		switch action.SafetyLevel {
		case computer.SafetyWarning:
			safetyLabel = "WARNING"
			safetyColor = ColorWarning
		case computer.SafetyDestructive:
			safetyLabel = "DESTRUCTIVE"
			safetyColor = ColorError
		}
		dot := StatusDot(DotSecure, frame)
		if action.SafetyLevel == computer.SafetyDestructive {
			dot = StatusDot(DotError, frame)
		}
		safetyStr := lipgloss.NewStyle().Foreground(safetyColor).Bold(true).Render(safetyLabel)
		card.WriteString(fmt.Sprintf("   Safety: %s %s\n", dot, safetyStr))

		card.WriteString(fmt.Sprintf("   Time:   %s\n", action.Timestamp.Format("15:04:05")))

		if action.ElementHTML != "" {
			htmlSnippet := truncateStr(action.ElementHTML, 60)
			card.WriteString(fmt.Sprintf("   Element: %s\n", htmlSnippet))
		}

		if action.Screenshot != "" {
			card.WriteString(fmt.Sprintf("   Screenshot: captured (%d bytes)\n", len(action.Screenshot)))
		}

		card.WriteString("")
		selDot := Spinner(frame)
		if selected {
			selDot = StatusDot(DotProcessing, frame)
		}
		scopeHint := "  o=once  s=session  p=permanent"
		card.WriteString(lipgloss.NewStyle().Foreground(ColorWarning).Render(fmt.Sprintf("  %s ↑↓  Enter: approve  Esc: reject", selDot)))
		card.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorFgSubtle).Render(scopeHint))

		cardStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderClr).
			Padding(0, 1).
			Width(width - 4)
		sections = append(sections, cardStyle.Render(card.String()))

		if i < len(rp.pending)-1 {
			sections = append(sections, "")
		}
	}

	selDot := Spinner(frame + 2)
	footer := lipgloss.NewStyle().
		Foreground(ColorFgSubtle).
		Render(fmt.Sprintf("  %s  ↑↓ navigate · Enter approve · Esc reject   │  %d pending", selDot, len(rp.pending)))
	sections = append(sections, footer)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorError).
		Padding(0, 1).
		Width(width)

	return panelStyle.Render(strings.Join(sections, "\n"))
}

func (rp *ReviewPanelModel) SelectNext() {
	if len(rp.pending) > 0 {
		rp.selected = (rp.selected + 1) % len(rp.pending)
	}
}

func (rp *ReviewPanelModel) SelectPrev() {
	if len(rp.pending) > 0 {
		rp.selected--
		if rp.selected < 0 {
			rp.selected = len(rp.pending) - 1
		}
	}
}

func (rp *ReviewPanelModel) ApproveSelected() error {
	if rp.selected < len(rp.pending) {
		req := rp.pending[rp.selected]
		return rp.reviewer.Resolve(req.Action.ID, computer.ReviewDecision{Approved: true, Reason: "user confirmed"})
	}
	return nil
}

func (rp *ReviewPanelModel) RejectSelected() error {
	if rp.selected < len(rp.pending) {
		req := rp.pending[rp.selected]
		return rp.reviewer.Resolve(req.Action.ID, computer.ReviewDecision{Approved: false, Reason: "user rejected"})
	}
	return nil
}

func (rp *ReviewPanelModel) PendingCount() int {
	return len(rp.pending)
}

func (rp *ReviewPanelModel) SetScope(s computer.ApprovalScope) {}

func (rp *ReviewPanelModel) Visible() bool {
	return rp.visible
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
