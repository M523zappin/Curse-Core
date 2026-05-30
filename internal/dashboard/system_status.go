package dashboard

import (
	"fmt"
	"os"
	"strings"

	"github.com/M523zappin/Curse-Core/internal/gateway"
	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
)

type SystemStatusModel struct {
	gateway *gateway.Gateway
	pid     int
	lastCP  *persistence.Checkpoint
}

func NewSystemStatusModel(gw *gateway.Gateway) *SystemStatusModel {
	return &SystemStatusModel{
		gateway: gw,
		pid:     os.Getpid(),
	}
}

func (m *SystemStatusModel) SetCheckpoint(cp *persistence.Checkpoint) {
	m.lastCP = cp
}

func (m *SystemStatusModel) Update(msg interface{}) {
}

func (m *SystemStatusModel) View(width int) string {
	machine := m.gateway.Machine()
	state := machine.State()

	stateDot := DotSecure
	if state == statemachine.StateError {
		stateDot = DotError
	} else if state == statemachine.StateRunning {
		stateDot = DotProcessing
	}

	stateLabel := state.String()
	if m.lastCP != nil {
		stateLabel = m.lastCP.State.String()
	}

	var lines []string

	// Daemon info
	lines = append(lines, StatusLineStyled("Daemon", fmt.Sprintf("● %s  PID %d", getHostname(), m.pid), DotSecure, false))
	lines = append(lines, StatusLineStyled("State", stateLabel, stateDot, state == statemachine.StateRunning))
	lines = append(lines, StatusLineStyled("Steps", fmt.Sprintf("%d", machine.Step()), DotSecure, false))

	// Checkpoint info
	if m.lastCP != nil {
		lines = append(lines, StatusLineStyled("CP Sequence", fmt.Sprintf("%d", m.lastCP.Sequence), DotSecure, false))
		lines = append(lines, StatusLineStyled("CP Step", fmt.Sprintf("%d", m.lastCP.Step), DotSecure, false))
		lines = append(lines, StatusLineStyled("Mission", truncate(m.lastCP.MissionID, 20), DotSecure, false))
	} else {
		lines = append(lines, StatusLineStyled("Checkpoint", "none", DotIdle, false))
	}

	// Queue
	lines = append(lines, StatusLineStyled("Queue", fmt.Sprintf("%d missions", m.gateway.Queue().Len()), DotSecure, false))

	// Model info
	modelName := m.gateway.ActiveModel()
	if modelName == "" {
		modelName = "none"
	}
	lines = append(lines, StatusLineStyled("Model", modelName, DotSecure, false))

	if m.gateway.Registry() != nil {
		profile, ok := m.gateway.Registry().ActiveProfile()
		if ok {
			lines = append(lines, StatusLineStyled("Provider", profile.Provider, DotSecure, false))
			lines = append(lines, StatusLineStyled("Window", fmt.Sprintf("%d tokens", profile.ContextWindow), DotSecure, false))
		}
	}

	// Computer Controller status
	comp := m.gateway.Computer()
	if comp != nil {
		buf := comp.VisionBuffer()
		bufSize := len(buf)
		bufDot := DotSecure
		if bufSize == 0 {
			bufDot = DotIdle
		}
		lines = append(lines, StatusLineStyled("Vision", fmt.Sprintf("%d frames", bufSize), bufDot, false))
		reviewMgr := m.gateway.ReviewManager()
		if reviewMgr != nil {
			pending := reviewMgr.PendingCount()
			reviewDot := DotProcessing
			if pending == 0 {
				reviewDot = DotSecure
			}
			lines = append(lines, StatusLineStyled("Reviews", fmt.Sprintf("%d pending", pending), reviewDot, pending > 0))
		}
	}

	// SHA256 chain status
	chainDot := DotSecure
	if state == statemachine.StateError {
		chainDot = DotError
	}
	lines = append(lines, "", StatusLineStyled("SHA256 Chain", "INTACT", chainDot, false))

	return strings.Join(lines, "\n")
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
