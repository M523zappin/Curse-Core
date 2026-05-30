package dashboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/M523zappin/Curse-Core/internal/mission"
)

type MissionQueueModel struct {
	queue *mission.Queue
}

func NewMissionQueueModel(q *mission.Queue) *MissionQueueModel {
	return &MissionQueueModel{queue: q}
}

func (m *MissionQueueModel) Update(msg tea.Msg) {
}

func (m *MissionQueueModel) View(width int, frame int) string {
	all := m.queue.All()

	todo := make([]string, 0)
	inProgress := make([]string, 0)
	done := make([]string, 0)

	for _, ms := range all {
		var statusStr string
		switch ms.Status {
		case mission.StatusTodo:
			statusStr = "TODO"
		case mission.StatusInProgress:
			statusStr = fmt.Sprintf("IN PROGRESS  step %d/%d", ms.Steps, ms.MaxSteps)
		case mission.StatusDone:
			statusStr = "DONE"
		}
		card := KanbanCardStyled(ms.ID[:8], ms.Task, statusStr, ms.Status == mission.StatusInProgress, width/3-1, frame)
		switch ms.Status {
		case mission.StatusTodo:
			todo = append(todo, card)
		case mission.StatusInProgress:
			inProgress = append(inProgress, card)
		case mission.StatusDone:
			done = append(done, card)
		}
	}

	if len(todo) == 0 {
		todo = append(todo, emptyCard("(empty)"))
	}
	if len(inProgress) == 0 {
		inProgress = append(inProgress, emptyCard("(none)"))
	}
	if len(done) == 0 {
		done = append(done, emptyCard("(none)"))
	}

	colW := width/3 - 2
	col1 := KanbanColumnStyled("TODO", ColorAccentDim, todo, colW)
	col2 := KanbanColumnStyled("IN PROGRESS", ColorProcessing, inProgress, colW)
	col3 := KanbanColumnStyled("DONE", ColorSuccess, done, colW)

	sep := lipgloss.NewStyle().Foreground(ColorBorder).Render("┊")
	cols := []string{col1, sep, col2, sep, col3}
	return strings.Join(cols, " ")
}

func emptyCard(msg string) string {
	return lipgloss.NewStyle().
		Foreground(ColorFgInactive).
		Italic(true).
		Padding(0, 1).
		Render(msg)
}
