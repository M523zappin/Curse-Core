package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type QuickAction struct {
	Key       string
	Label     string
	Desc      string
	Color     lipgloss.Color
}

var defaultQuickActions = []QuickAction{
	{Key: "Ctrl+N", Label: "talk", Desc: "natural language", Color: ColorPsychic},
	{Key: "Tab", Label: "model", Desc: "cycle models", Color: ColorSpiral},
	{Key: "/", Label: "cmd", Desc: "commands", Color: ColorAccent},
	{Key: "Ctrl+M", Label: "browse", Desc: "model browser", Color: ColorProcessing},
	{Key: "Ctrl+P", Label: "pause", Desc: "pause/resume", Color: ColorWarning},
	{Key: "Ctrl+Y", Label: "sync", Desc: "sync constitution", Color: ColorToxic},
	{Key: "Ctrl+S", Label: "quit", Desc: "shutdown", Color: ColorError},
}

func renderQuickBar(width int, frame int) string {
	if width < 40 {
		width = 40
	}

	segs := make([]string, 0, len(defaultQuickActions))
	for _, a := range defaultQuickActions {
		keyStyle := lipgloss.NewStyle().
			Foreground(ColorBg).
			Background(a.Color).
			Bold(true).
			Padding(0, 1)
		keyStr := keyStyle.Render(a.Key)
		labelStyle := lipgloss.NewStyle().Foreground(ColorFgSubtle)
		segs = append(segs, keyStr+labelStyle.Render(" "+a.Label))
	}

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render("│")
	bar := strings.Join(segs, fmt.Sprintf(" %s ", sep))

	return lipgloss.NewStyle().
		Width(width - 2).
		Render(bar)
}
