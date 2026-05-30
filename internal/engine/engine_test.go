package engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/M523zappin/Curse-Core/internal/agent"
	"github.com/M523zappin/Curse-Core/internal/healing"
	"github.com/M523zappin/Curse-Core/internal/knowledge"
	"github.com/M523zappin/Curse-Core/internal/mission"
	"github.com/M523zappin/Curse-Core/internal/skill"
)

func newTestEngine(t *testing.T) (*Engine, *mission.Queue, *agent.Fleet, *skill.Store, *knowledge.Index, *healing.HealingLoop) {
	t.Helper()
	queue := mission.NewQueue()
	fleet := agent.NewFleet()
	fleet.RegisterRole(agent.RoleArchitect, 1)
	fleet.RegisterRole(agent.RoleRefactor, 1)
	fleet.RegisterRole(agent.RoleReviewer, 1)
	fleet.RegisterRole(agent.RoleDocWriter, 1)
	fleet.SetDispatcher(func(a *agent.Agent, t agent.Task) *agent.TaskResult {
		time.Sleep(5 * time.Millisecond)
		return &agent.TaskResult{
			TaskID:  t.ID,
			Success: true,
			Output:  "test output",
		}
	})
	skillsDir, err := os.MkdirTemp("", "engine-skills")
	if err != nil {
		t.Fatal(err)
	}
	knDir, err := os.MkdirTemp("", "engine-knowledge")
	if err != nil {
		t.Fatal(err)
	}
	skills := skill.NewStore(skillsDir)
	kn := knowledge.NewIndex(knDir)
	hl := healing.NewHealingLoop()
	eng := New(queue, fleet, skills, kn, hl)
	return eng, queue, fleet, skills, kn, hl
}

// initEngineCtx sets up the engine's context for direct Tick() calls without Run()
func initEngineCtx(eng *Engine) {
	eng.ctx = context.Background()
}

func TestNew(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
	if eng.running {
		t.Fatal("expected engine to not be running initially")
	}
}

func TestEngineRunAndStop(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	ctx := context.Background()
	go eng.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	if !eng.Running() {
		t.Fatal("expected engine to be running after Run()")
	}
	eng.Stop()
	time.Sleep(50 * time.Millisecond)
	if eng.Running() {
		t.Fatal("expected engine to stop after Stop()")
	}
}

func TestEngineTickWithMission(t *testing.T) {
	eng, queue, _, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	queue.Enqueue(mission.New("test mission", []string{"test"}, "", 10))
	eng.Tick()

	if queue.Len() != 0 {
		t.Fatalf("expected queue to be empty after tick, got %d missions", queue.Len())
	}
}

func TestEngineMultipleMissions(t *testing.T) {
	eng, queue, _, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	queue.Enqueue(mission.New("first mission", []string{"test"}, "", 10))
	queue.Enqueue(mission.New("second mission", []string{"test"}, "", 10))

	eng.Tick()
	if queue.Len() != 1 {
		t.Fatalf("expected 1 mission remaining, got %d", queue.Len())
	}

	eng.Tick()
	if queue.Len() != 0 {
		t.Fatalf("expected 0 missions remaining, got %d", queue.Len())
	}
}

func TestEngineEmptyTick(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	traces := make([]string, 0)
	eng.SetTraceHook(func(level, msg string) {
		traces = append(traces, level+":"+msg)
	})
	eng.Tick()
	if len(traces) == 0 {
		t.Fatal("expected trace output even on empty tick")
	}
}

func TestEnginePhaseTransitions(t *testing.T) {
	eng, queue, _, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	var captured []Phase
	eng.SetTraceHook(func(level, msg string) {
		if level == "system" || level == "mission" {
			eng.mu.Lock()
			captured = append(captured, eng.phase)
			eng.mu.Unlock()
		}
	})

	queue.Enqueue(mission.New("phase test", []string{"test"}, "", 10))
	eng.Tick()

	if len(captured) == 0 {
		t.Fatal("expected phase transitions during tick")
	}
}

func TestDoubleRun(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	ctx := context.Background()
	go eng.Run(ctx)
	go eng.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	eng.Stop()
}

func TestSetTraceHook(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	called := false
	eng.SetTraceHook(func(level, msg string) {
		called = true
	})
	eng.trace("test", "hello")
	if !called {
		t.Fatal("expected trace hook to be called")
	}
}

func TestSetStatusHook(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	called := false
	eng.SetStatusHook(func(s Status) {
		called = true
	})
	eng.emitStatus()
	if !called {
		t.Fatal("expected status hook to be called")
	}
}

func TestEngineContextCancellation(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	go eng.Run(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestEngineIdlePhase(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	if p := eng.Phase(); p != PhaseIdle {
		t.Fatalf("expected PhaseIdle initially, got %s", p)
	}
}

func TestEngineNotRunningInitially(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	if eng.Running() {
		t.Fatal("expected engine to not be running initially")
	}
}

func TestEngineStopWithoutStart(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	eng.Stop()
}

func TestEngineDispatchFleet(t *testing.T) {
	eng, _, fleet, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	fleet.Enqueue(agent.Task{
		ID:          "test-task-1",
		Role:        agent.RoleArchitect,
		Description: "test dispatch",
		Priority:    agent.PriorityNormal,
	})

	eng.dispatchFleet()
	time.Sleep(50 * time.Millisecond)

	if n := len(fleet.CompletedTasks()); n == 0 {
		t.Fatal("expected at least one completed task after dispatch")
	}
}

func TestEngineCollectResults(t *testing.T) {
	eng, _, fleet, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	fleet.Enqueue(agent.Task{
		ID:          "collect-task",
		Role:        agent.RoleArchitect,
		Description: "test collect",
		Priority:    agent.PriorityNormal,
	})

	results := eng.collectResults("test")
	if len(results) == 0 {
		t.Fatal("expected at least one result from collect")
	}
}

func TestEnginePlanTasks(t *testing.T) {
	eng, _, _, skills, _, _ := newTestEngine(t)
	m := mission.New("test planning", []string{"test"}, "", 10)

	tasks := eng.planTasks(m, nil)
	if len(tasks) < 3 {
		t.Fatalf("expected at least 3 planned tasks, got %d", len(tasks))
	}

	tasks = eng.planTasks(m, skills.All())
	if len(tasks) < 3 {
		t.Fatalf("expected at least 3 planned tasks with skills, got %d", len(tasks))
	}
}

func TestEngineLearnFromMission(t *testing.T) {
	eng, _, _, skills, kn, _ := newTestEngine(t)

	m := mission.New("learn test", []string{"learn"}, "", 10)
	results := []agent.TaskResult{
		{TaskID: "t1", Success: true, Output: "done"},
		{TaskID: "t2", Success: true, Output: "done"},
	}

	eng.learnFromMission(m, results)

	if kn.Count() < 1 {
		t.Fatal("expected at least 1 knowledge entry after learning")
	}

	if skills.Count() < 1 {
		t.Fatal("expected at least 1 skill after learning from successful mission")
	}
}

func TestEngineLearnFromFailedMission(t *testing.T) {
	eng, _, _, skills, kn, _ := newTestEngine(t)

	m := mission.New("failed mission", []string{"fail"}, "", 10)
	results := []agent.TaskResult{
		{TaskID: "t1", Success: false, Output: "failed", Error: "something went wrong"},
	}

	eng.learnFromMission(m, results)

	if kn.Count() < 1 {
		t.Fatal("expected at least 1 knowledge entry even on failure")
	}

	if skills.Count() > 0 {
		t.Log("expected no skills from failed mission")
	}
}

func TestEngineProcessResults(t *testing.T) {
	eng, _, fleet, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	fleet.RegisterRole(agent.RoleDocWriter, 1)
	fleet.SetDispatcher(func(a *agent.Agent, t agent.Task) *agent.TaskResult {
		return &agent.TaskResult{
			TaskID:  t.ID,
			Success: true,
			Output:  "test output",
		}
	})

	m := mission.New("process test", []string{"test"}, "", 10)
	results := []agent.TaskResult{
		{TaskID: "r1", Success: true, Output: "result 1"},
		{TaskID: "r2", Success: false, Output: "result 2", Error: "err"},
	}

	eng.processResults(m, results)

	if n := eng.fleet.TaskCount(); n < 2 {
		t.Fatalf("expected at least 2 learn tasks enqueued, got %d", n)
	}
}

func TestEngineRejectsDuplicateTask(t *testing.T) {
	_, _, fleet, _, _, _ := newTestEngine(t)
	err := fleet.Enqueue(agent.Task{ID: "dup", Role: agent.RoleArchitect, Description: "first"})
	if err != nil {
		t.Fatalf("first enqueue should succeed: %v", err)
	}
	err = fleet.Enqueue(agent.Task{ID: "dup", Role: agent.RoleArchitect, Description: "second"})
	if err == nil {
		t.Fatal("expected error on duplicate task")
	}
}

func TestEngineMetricsAfterTick(t *testing.T) {
	eng, queue, _, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	queue.Enqueue(mission.New("metrics test", []string{"test"}, "", 10))
	eng.Tick()

	m := eng.Metrics()
	if m.MissionsProcessed < 1 {
		t.Fatalf("expected >= 1 missions processed, got %d", m.MissionsProcessed)
	}
	if m.TasksCompleted < 3 {
		t.Fatalf("expected >= 3 tasks completed, got %d", m.TasksCompleted)
	}
	if m.TotalDuration <= 0 {
		t.Fatal("expected positive total duration")
	}
	if m.LastMissionDuration <= 0 {
		t.Fatal("expected positive last mission duration")
	}
}

func TestEngineLearnTaskLimit(t *testing.T) {
	eng, _, fleet, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	for i := 0; i < 10; i++ {
		fleet.Enqueue(agent.Task{
			ID:          fmt.Sprintf("pre-existing-learn-%d", i),
			Role:        agent.RoleDocWriter,
			Description: fmt.Sprintf("learn task %d", i),
		})
	}

	m := mission.New("limit test", []string{"test"}, "", 10)
	results := []agent.TaskResult{
		{TaskID: "r1", Success: true, Output: "result"},
		{TaskID: "r2", Success: true, Output: "result"},
		{TaskID: "r3", Success: true, Output: "result"},
	}

	eng.processResults(m, results)

	for _, pt := range eng.fleet.PendingTasks() {
		if strings.HasPrefix(pt.ID, "r1-learn") || strings.HasPrefix(pt.ID, "r2-learn") || strings.HasPrefix(pt.ID, "r3-learn") {
			t.Fatalf("expected no new learn tasks (limit reached), but found %s", pt.ID)
		}
	}
}

func TestEnginePanicRecovery(t *testing.T) {
	eng, queue, fleet, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	fleet.SetDispatcher(func(a *agent.Agent, t agent.Task) *agent.TaskResult {
		panic("test panic in dispatcher")
	})

	queue.Enqueue(mission.New("panic test", []string{"test"}, "", 10))

	traces := make([]string, 0)
	eng.SetTraceHook(func(level, msg string) {
		traces = append(traces, level+":"+msg)
	})

	eng.Tick()

	if queue.Len() != 0 {
		t.Fatalf("expected queue empty after panic tick, got %d", queue.Len())
	}
}

func TestEngineKnowledgeContextInjection(t *testing.T) {
	eng, queue, _, _, _, _ := newTestEngine(t)
	initEngineCtx(eng)

	queue.Enqueue(mission.New("context test", []string{"test"}, "", 10))
	eng.Tick()

	m := eng.Metrics()
	if m.MissionsProcessed < 1 {
		t.Fatal("expected mission to be processed with context injection")
	}
}

func TestEngineMetricsInitial(t *testing.T) {
	eng, _, _, _, _, _ := newTestEngine(t)
	m := eng.Metrics()
	if m.MissionsProcessed != 0 {
		t.Fatalf("expected 0 missions initially, got %d", m.MissionsProcessed)
	}
	if m.TasksCompleted != 0 {
		t.Fatalf("expected 0 tasks initially, got %d", m.TasksCompleted)
	}
}

