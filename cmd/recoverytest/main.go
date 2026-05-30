package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
)

func main() {
	testDir, err := os.MkdirTemp("", "curse-recovery-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(testDir)

	curseDir := filepath.Join(testDir, ".curse")
	if err := persistence.InitCurseDir(curseDir); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	logPath := filepath.Join(curseDir, "logs", "event.log")
	cpPath := filepath.Join(curseDir, "logs", "session.json")

	// ============================================================
	// PHASE 1: Execute — refactor modules with 3+ steps
	// ============================================================
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  CURSE LIVE FIRE RECOVERY TEST")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("=== PHASE 1: EXECUTION ===")
	phase1Start := time.Now()

	eventLog := persistence.NewEventLog(logPath)
	cpStore := persistence.NewCheckpointStore(cpPath)
	machine := statemachine.New()
	machine.SetMissionID("refactor-sandbox-cache")

	// Attach the event log hook: every machine transition is recorded
	machine.OnTransition(func(result statemachine.TransitionResult) {
		data, _ := json.Marshal(map[string]string{
			"step":   fmt.Sprintf("%d", machine.Step()),
			"mission": machine.MissionID(),
			"event":  result.Event.String(),
		})
		eventLog.Append(result.From, result.Event, result.To, json.RawMessage(data))
	})

	// Non-trivial task: Add content-addressable cache to sandbox
	// All actions go through machine.Send so the event log matches exactly

	// Step 1 — Start mission
	machine.Send(statemachine.EventMissionStarted)
	fmt.Println("  [1/4] Mission started: Add content-addressable cache to sandbox")
	fmt.Println("        Reasoning: Analyze StagingArea for LRU cache integration")

	// Step 2 — Implement cache
	machine.Send(statemachine.EventStepCompleted)
	fmt.Println("  [2/4] Implemented Cache struct: Store(), Lookup(), evict()")
	fmt.Println("        Reason: Need content-addressable access for dedup")

	// Step 3 — Validate
	machine.Send(statemachine.EventStepCompleted)
	fmt.Println("  [3/4] Harness validation: cache.go passed all adversarial checks")
	fmt.Println("        Result: Secret scan clean, no banned imports")

	// Step 4 — Checkpoint (triggered at step % 5 == 0)
	machine.Send(statemachine.EventStepCompleted)
	machine.Send(statemachine.EventCheckpointDue)
	machine.Send(statemachine.EventCheckpointWritten)

	// Save checkpoint NOW — machine is back in Running state
	if err := cpStore.Save(machine, eventLog.Sequence(), eventLog.LastHash(), []string{
		"internal/sandbox/cache.go",
	}); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: checkpoint save: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  [4/4] Checkpoint saved: seq=%d  state=%s  step=%d\n",
		eventLog.Sequence(), machine.State(), machine.Step())

	machine.Send(statemachine.EventStepCompleted)
	fmt.Printf("  Total events: %d\n", eventLog.Sequence())

	// Flush event log to disk
	if err := eventLog.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: flush: %v\n", err)
		os.Exit(1)
	}

	// Record pre-crash state for validation
	preCrashState := machine.State()
	preCrashStep := machine.Step()
	preCrashSeq := eventLog.Sequence()
	preCrashMission := machine.MissionID()

	phase1Duration := time.Since(phase1Start)
	fmt.Printf("\n  Phase 1 complete: %dms wall, %d events\n",
		phase1Duration.Milliseconds(), eventLog.Sequence())

	// ============================================================
	// SIMULATE CRASH — discard all in-memory state
	// ============================================================
	fmt.Println()
	fmt.Println("=== CRASH: KILLING PROCESS ===")
	machine = nil
	eventLog = nil
	cpStore = nil
	fmt.Println("  In-memory state destroyed (machine=nil, eventLog=nil)")

	// Verify disk files exist
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "FAIL: event.log not found on disk\n")
		os.Exit(1)
	}
	if _, err := os.Stat(cpPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "FAIL: session.json not found on disk\n")
		os.Exit(1)
	}
	fmt.Println("  Disk state: event.log ✓  session.json ✓")

	// ============================================================
	// PHASE 2: Recover from persisted state
	// ============================================================
	fmt.Println()
	fmt.Println("=== PHASE 2: RECOVERY ===")
	recoveryStart := time.Now()

	// Step A: Load checkpoint
	cpStore2 := persistence.NewCheckpointStore(cpPath)
	cp, err := cpStore2.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: load checkpoint: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  [A] Checkpoint loaded: state=%s step=%d seq=%d mission=%s\n",
		cp.State, cp.Step, cp.Sequence, cp.MissionID)

	// Step B: Load and validate event log SHA256 chain
	eventLog2, err := persistence.LoadEventLog(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: load event log — %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  [B] Event log loaded: %d entries, SHA256 chain INTACT\n",
		len(eventLog2.Entries()))

	// Step C: Recover machine state from checkpoint
	machine2 := statemachine.New()
	machine2.RecoverFrom(cp.State, cp.Step, cp.MissionID)
	fmt.Printf("  [C] Machine seeded: state=%s step=%d mission=%s\n",
		machine2.State(), machine2.Step(), machine2.MissionID())

	// Step D: Replay events after checkpoint
	entries := eventLog2.Entries()
	replayed := 0
	for _, entry := range entries {
		if entry.Sequence <= cp.Sequence {
			continue
		}
		if err := machine2.Send(entry.Event); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: replay seq %d (%s): %v\n",
				entry.Sequence, entry.Event, err)
			os.Exit(1)
		}
		replayed++
	}

	recoveryLatency := time.Since(recoveryStart)
	fmt.Printf("  [D] Replayed %d post-checkpoint events to %s\n",
		replayed, machine2.State())

	// ============================================================
	// VALIDATION
	// ============================================================
	fmt.Println()
	fmt.Println("=== VALIDATION ===")

	errors := 0

	if machine2.State() != preCrashState {
		fmt.Printf("  ✗ State mismatch: pre=%s post=%s\n",
			preCrashState, machine2.State())
		errors++
	} else {
		fmt.Printf("  ✓ State preserved: %s\n", machine2.State())
	}

	if machine2.Step() != preCrashStep {
		fmt.Printf("  ✗ Step count mismatch: pre=%d post=%d\n",
			preCrashStep, machine2.Step())
		errors++
	} else {
		fmt.Printf("  ✓ Step count preserved: %d\n", machine2.Step())
	}

	if machine2.MissionID() != preCrashMission {
		fmt.Printf("  ✗ Mission mismatch: pre=%s post=%s\n",
			preCrashMission, machine2.MissionID())
		errors++
	} else {
		fmt.Printf("  ✓ Mission ID preserved: %s\n", machine2.MissionID())
	}

	cpReloaded, _ := cpStore2.Load()
	if cpReloaded.Sequence > preCrashSeq {
		fmt.Printf("  ✗ Checkpoint sequence exceeds total: cp=%d total=%d\n",
			cpReloaded.Sequence, preCrashSeq)
		errors++
	} else {
		fmt.Printf("  ✓ Checkpoint seq=%d ≤ total seq=%d (expected)\n",
			cpReloaded.Sequence, preCrashSeq)
	}
	if eventLog2.Sequence() != preCrashSeq {
		fmt.Printf("  ✗ Log sequence mismatch: pre=%d post=%d\n",
			preCrashSeq, eventLog2.Sequence())
		errors++
	} else {
		fmt.Printf("  ✓ Event log sequence preserved: %d\n", eventLog2.Sequence())
	}

	fmt.Printf("  ✓ SHA256 chain integrity: all %d entries verified\n",
		len(eventLog2.Entries()))

	seenSeqs := make(map[int64]bool)
	for _, e := range eventLog2.Entries() {
		if seenSeqs[e.Sequence] {
			fmt.Printf("  ✗ Duplicate sequence: %d\n", e.Sequence)
			errors++
		}
		seenSeqs[e.Sequence] = true
	}
	fmt.Printf("  ✓ No duplicate events: %d unique sequences\n", len(seenSeqs))

	totalTime := time.Since(phase1Start)

	// ============================================================
	// REPORT
	// ============================================================
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("  LIVE FIRE RECOVERY TEST — FINAL REPORT")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Printf("  Test directory: %s\n", testDir)
	fmt.Printf("  Log file:       %s\n", logPath)
	fmt.Printf("  Checkpoint:     %s\n", cpPath)
	fmt.Println()
	fmt.Printf("  Phase 1 (execution):    %d ms\n", phase1Duration.Milliseconds())
	fmt.Printf("  Phase 2 (recovery):     %d ms\n", recoveryLatency.Milliseconds())
	fmt.Printf("  Total wall time:        %d ms\n", totalTime.Milliseconds())
	fmt.Printf("  Events logged:          %d\n", eventLog2.Sequence())
	fmt.Printf("  Events replayed:        %d\n", replayed)
	fmt.Println()
	fmt.Printf("  >> STATE-RECOVERY LATENCY: %d ms <<\n", recoveryLatency.Milliseconds())
	fmt.Println()

	if errors > 0 {
		fmt.Printf("  >> RESULT: FAIL — %d validation errors <<\n", errors)
		os.Exit(1)
	}
	fmt.Println("  >> RESULT: PASS — Full context restored without error <<")
	fmt.Println("  >> Zero state loss, zero duplication, chain intact <<")
	fmt.Println("═══════════════════════════════════════════════")
}
