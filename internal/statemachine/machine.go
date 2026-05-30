package statemachine

import "fmt"

type TransitionHook func(TransitionResult)

type Machine struct {
	currentState State
	stepCounter  int
	missionID    string
	errorStack   []string
	hooks        []TransitionHook
}

func New() *Machine {
	return &Machine{
		currentState: StateIdle,
		stepCounter:  0,
	}
}

func (m *Machine) Send(event Event) error {
	from := m.currentState
	to, ok := ValidTransition(from, event)
	if !ok {
		return fmt.Errorf("illegal transition: %s → %s (from state %s)", from, event, from)
	}
	m.currentState = to
	result := TransitionResult{From: from, To: to, Event: event}

	if event == EventStepCompleted {
		m.stepCounter++
	}
	if event == EventCheckpointWritten {
		m.stepCounter = 0
	}

	for _, hook := range m.hooks {
		hook(result)
	}
	return nil
}

func (m *Machine) RecoverFrom(state State, step int, missionID string) {
	m.currentState = state
	m.stepCounter = step
	m.missionID = missionID
}

func (m *Machine) State() State {
	return m.currentState
}

func (m *Machine) Step() int {
	return m.stepCounter
}

func (m *Machine) MissionID() string {
	return m.missionID
}

func (m *Machine) SetMissionID(id string) {
	m.missionID = id
}

func (m *Machine) OnTransition(hook TransitionHook) {
	m.hooks = append(m.hooks, hook)
}

func (m *Machine) CheckpointDue() bool {
	return m.stepCounter > 0 && m.stepCounter%5 == 0
}

func (m *Machine) TriggerSync() error {
	return m.Send(EventSyncTriggered)
}

func (m *Machine) CompleteSync() error {
	return m.Send(EventSyncCompleted)
}

func (m *Machine) FailSync() error {
	return m.Send(EventSyncFailed)
}
