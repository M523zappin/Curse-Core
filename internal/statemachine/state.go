package statemachine

import "fmt"

type State int

const (
	StateIdle         State = iota // 0 — initial, no mission active
	StateRunning                   // 1 — actively executing
	StatePaused                    // 2 — user or system pause
	StateCheckpointing             // 3 — writing checkpoint (every 5 steps)
	StateError                     // 4 — unrecoverable error
	StateRecovering                // 5 — replaying event log on restart
	StateShutdown                  // 6 — graceful termination
)

var stateNames = map[State]string{
	StateIdle:         "Idle",
	StateRunning:      "Running",
	StatePaused:       "Paused",
	StateCheckpointing: "Checkpointing",
	StateError:        "Error",
	StateRecovering:   "Recovering",
	StateShutdown:     "Shutdown",
}

func (s State) String() string {
	if name, ok := stateNames[s]; ok {
		return name
	}
	return fmt.Sprintf("State(%d)", s)
}

type Event int

const (
	EventMissionStarted       Event = iota // Idle → Running
	EventStepCompleted                     // Running → Running (inc step counter)
	EventCheckpointDue                     // Running → Checkpointing (step % 5 == 0)
	EventCheckpointWritten                 // Checkpointing → Running
	EventPauseRequested                    // Running → Paused
	EventResumeRequested                   // Paused → Running
	EventErrorOccurred                     // Running → Running (non-fatal, logged)
	EventFatalError                        // any  → Error
	EventRecoveryInitiated                 // Error → Recovering
	EventRecoveryCompleted                 // Recovering → Running
	EventRecoveryFailed                    // Recovering → Error
	EventShutdownRequested                 // any  → Shutdown
)

var eventNames = map[Event]string{
	EventMissionStarted:       "MissionStarted",
	EventStepCompleted:        "StepCompleted",
	EventCheckpointDue:        "CheckpointDue",
	EventCheckpointWritten:    "CheckpointWritten",
	EventPauseRequested:       "PauseRequested",
	EventResumeRequested:      "ResumeRequested",
	EventErrorOccurred:        "ErrorOccurred",
	EventFatalError:           "FatalError",
	EventRecoveryInitiated:    "RecoveryInitiated",
	EventRecoveryCompleted:    "RecoveryCompleted",
	EventRecoveryFailed:       "RecoveryFailed",
	EventShutdownRequested:    "ShutdownRequested",
}

func (e Event) String() string {
	if name, ok := eventNames[e]; ok {
		return name
	}
	return fmt.Sprintf("Event(%d)", e)
}

type TransitionResult struct {
	From  State
	To    State
	Event Event
	Err   error
}

var transitionTable [7][12]State

func init() {
	stable := &transitionTable
	for i := range stable {
		for j := range stable[i] {
			stable[i][j] = State(255) // illegal sentinel
		}
	}
	stable[StateIdle][EventMissionStarted] = StateRunning
	stable[StateRunning][EventStepCompleted] = StateRunning
	stable[StateRunning][EventCheckpointDue] = StateCheckpointing
	stable[StateCheckpointing][EventCheckpointWritten] = StateRunning
	stable[StateRunning][EventPauseRequested] = StatePaused
	stable[StatePaused][EventResumeRequested] = StateRunning
	stable[StateRunning][EventErrorOccurred] = StateRunning
	stable[StateIdle][EventFatalError] = StateError
	stable[StateRunning][EventFatalError] = StateError
	stable[StateCheckpointing][EventFatalError] = StateError
	stable[StatePaused][EventFatalError] = StateError
	stable[StateRecovering][EventFatalError] = StateError
	stable[StateError][EventRecoveryInitiated] = StateRecovering
	stable[StateRecovering][EventRecoveryCompleted] = StateRunning
	stable[StateRecovering][EventRecoveryFailed] = StateError
	stable[StateIdle][EventShutdownRequested] = StateShutdown
	stable[StateRunning][EventShutdownRequested] = StateShutdown
	stable[StatePaused][EventShutdownRequested] = StateShutdown
	stable[StateCheckpointing][EventShutdownRequested] = StateShutdown
	stable[StateError][EventShutdownRequested] = StateShutdown
	stable[StateRecovering][EventShutdownRequested] = StateShutdown
}

func ValidTransition(from State, event Event) (State, bool) {
	if int(from) >= len(transitionTable) || int(event) >= len(transitionTable[from]) {
		return State(255), false
	}
	to := transitionTable[from][event]
	if to == State(255) {
		return State(255), false
	}
	return to, true
}
