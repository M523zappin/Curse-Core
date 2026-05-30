package statemachine

import (
	"testing"
)

func TestNewMachineStartsIdle(t *testing.T) {
	m := New()
	if m.State() != StateIdle {
		t.Fatalf("expected Idle, got %s", m.State())
	}
}

func TestIdleToRunning(t *testing.T) {
	m := New()
	if err := m.Send(EventMissionStarted); err != nil {
		t.Fatal(err)
	}
	if m.State() != StateRunning {
		t.Fatalf("expected Running, got %s", m.State())
	}
}

func TestRunningStepCounts(t *testing.T) {
	m := New()
	m.Send(EventMissionStarted)
	for i := 0; i < 3; i++ {
		m.Send(EventStepCompleted)
	}
	if m.Step() != 3 {
		t.Fatalf("expected step 3, got %d", m.Step())
	}
}

func TestCheckpointDue(t *testing.T) {
	m := New()
	m.Send(EventMissionStarted)
	for i := 0; i < 5; i++ {
		m.Send(EventStepCompleted)
	}
	if !m.CheckpointDue() {
		t.Fatal("expected CheckpointDue after 5 steps")
	}
}

func TestCheckpointFlow(t *testing.T) {
	m := New()
	m.Send(EventMissionStarted)
	for i := 0; i < 5; i++ {
		m.Send(EventStepCompleted)
	}
	if m.State() != StateRunning {
		t.Fatalf("expected Running, got %s", m.State())
	}
}

func TestPauseResume(t *testing.T) {
	m := New()
	m.Send(EventMissionStarted)
	if err := m.Send(EventPauseRequested); err != nil {
		t.Fatal(err)
	}
	if m.State() != StatePaused {
		t.Fatalf("expected Paused, got %s", m.State())
	}
	if err := m.Send(EventResumeRequested); err != nil {
		t.Fatal(err)
	}
	if m.State() != StateRunning {
		t.Fatalf("expected Running, got %s", m.State())
	}
}

func TestRecoveryFlow(t *testing.T) {
	m := New()
	m.Send(EventMissionStarted)
	m.Send(EventFatalError)
	if m.State() != StateError {
		t.Fatalf("expected Error, got %s", m.State())
	}
	m.Send(EventRecoveryInitiated)
	if m.State() != StateRecovering {
		t.Fatalf("expected Recovering, got %s", m.State())
	}
	m.Send(EventRecoveryCompleted)
	if m.State() != StateRunning {
		t.Fatalf("expected Running, got %s", m.State())
	}
}

func TestRecoverFromCheckpoint(t *testing.T) {
	m := New()
	m.RecoverFrom(StateRunning, 7, "mission-42")
	if m.State() != StateRunning {
		t.Fatalf("expected Running, got %s", m.State())
	}
	if m.Step() != 7 {
		t.Fatalf("expected step 7, got %d", m.Step())
	}
	if m.MissionID() != "mission-42" {
		t.Fatalf("expected mission-42, got %s", m.MissionID())
	}
}

func TestIllegalTransition(t *testing.T) {
	m := New()
	err := m.Send(EventPauseRequested)
	if err == nil {
		t.Fatal("expected error for illegal transition Idle → PauseRequested")
	}
}

func TestShutdownFromAnyState(t *testing.T) {
	states := []State{StateIdle, StateRunning, StatePaused, StateCheckpointing, StateError, StateRecovering}
	for _, s := range states {
		m := New()
		switch s {
		case StateRunning:
			m.Send(EventMissionStarted)
		case StatePaused:
			m.Send(EventMissionStarted)
			m.Send(EventPauseRequested)
		case StateCheckpointing:
			m.Send(EventMissionStarted)
			for i := 0; i < 5; i++ {
				m.Send(EventStepCompleted)
			}
		case StateError:
			m.Send(EventMissionStarted)
			m.Send(EventFatalError)
		case StateRecovering:
			m.Send(EventMissionStarted)
			m.Send(EventFatalError)
			m.Send(EventRecoveryInitiated)
		}
		if err := m.Send(EventShutdownRequested); err != nil {
			t.Fatalf("shutdown from %s should be valid: %s", s, err)
		}
		if m.State() != StateShutdown {
			t.Fatalf("expected Shutdown from %s, got %s", s, m.State())
		}
	}
}

func TestHookCalled(t *testing.T) {
	m := New()
	called := false
	m.OnTransition(func(r TransitionResult) {
		called = true
		if r.From != StateIdle || r.To != StateRunning || r.Event != EventMissionStarted {
			t.Fatalf("unexpected hook args: %+v", r)
		}
	})
	m.Send(EventMissionStarted)
	if !called {
		t.Fatal("hook was not called")
	}
}
