package engine

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/M523zappin/Curse-Core/internal/agent"
	"github.com/M523zappin/Curse-Core/internal/consciousness"
	"github.com/M523zappin/Curse-Core/internal/healing"
	"github.com/M523zappin/Curse-Core/internal/knowledge"
	"github.com/M523zappin/Curse-Core/internal/mission"
	"github.com/M523zappin/Curse-Core/internal/skill"
)

type Phase string

const (
	PhaseIdle      Phase = "idle"
	PhasePlanning  Phase = "planning"
	PhaseDispatching Phase = "dispatching"
	PhaseExecuting Phase = "executing"
	PhaseCollecting Phase = "collecting"
	PhaseLearning  Phase = "learning"
)

type Status struct {
	Phase      Phase
	MissionID  string
	TaskCount  int
	DoneCount  int
	SkillCount int
}

type Metrics struct {
	MissionsProcessed   int64
	TasksCompleted      int64
	SkillsGenerated     int64
	TotalDuration       time.Duration
	LastMissionDuration time.Duration
}

type Engine struct {
	queue     *mission.Queue
	fleet     *agent.Fleet
	skills    *skill.Store
	knowledge *knowledge.Index
	healer    *healing.HealingLoop
	mind      *consciousness.Consciousness

	ctx     context.Context
	cancel  context.CancelFunc
	onTrace func(string, string)
	ticker  *time.Ticker

	mu       sync.Mutex
	running  bool
	phase    Phase
	statusFn func(Status)

	metrics   Metrics
	metricsMu sync.Mutex
}

func New(
	queue *mission.Queue,
	fleet *agent.Fleet,
	skills *skill.Store,
	knowledge *knowledge.Index,
	healer *healing.HealingLoop,
	mind *consciousness.Consciousness,
) *Engine {
	return &Engine{
		queue:     queue,
		fleet:     fleet,
		skills:    skills,
		knowledge: knowledge,
		healer:    healer,
		mind:      mind,
		phase:     PhaseIdle,
	}
}

func (e *Engine) SetTraceHook(h func(string, string)) {
	e.onTrace = h
}

func (e *Engine) SetStatusHook(h func(Status)) {
	e.statusFn = h
}

func (e *Engine) trace(level, msg string) {
	if e.onTrace != nil {
		e.onTrace(level, msg)
	}
}

func (e *Engine) emitStatus() {
	e.mu.Lock()
	s := Status{
		Phase:      e.phase,
		SkillCount: e.skills.Count(),
	}
	results := e.fleet.CompletedTasks()
	s.DoneCount = len(results)
	s.MissionID = ""
	if m := e.queue.Peek(); m != nil {
		s.MissionID = m.ID
	}
	s.TaskCount = e.fleet.TaskCount()
	e.mu.Unlock()

	if e.statusFn != nil {
		e.statusFn(s)
	}
}

func (e *Engine) Metrics() Metrics {
	e.metricsMu.Lock()
	defer e.metricsMu.Unlock()
	return e.metrics
}

func (e *Engine) Run(ctx context.Context) {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.ctx, e.cancel = context.WithCancel(ctx)
	e.ticker = time.NewTicker(1 * time.Second)
	e.phase = PhaseIdle
	e.mu.Unlock()

	e.trace("system", "engine loop started")
	e.emitStatus()

	for {
		select {
		case <-e.ctx.Done():
			e.ticker.Stop()
			e.mu.Lock()
			e.running = false
			e.phase = PhaseIdle
			e.mu.Unlock()
			e.trace("system", "engine loop stopped")
			return

		case <-e.ticker.C:
			e.Tick()
		}
	}
}

func (e *Engine) Stop() {
	e.mu.Lock()
	cancel := e.cancel
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (e *Engine) Running() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *Engine) Phase() Phase {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.phase
}

func (e *Engine) Tick() {
	defer func() {
		if r := recover(); r != nil {
			e.trace("error", fmt.Sprintf("engine tick panic: %v\n%s", r, debug.Stack()))
			e.mu.Lock()
			e.phase = PhaseIdle
			e.mu.Unlock()
			e.emitStatus()
		}
	}()

	e.mu.Lock()
	e.phase = PhasePlanning
	e.mu.Unlock()

	m := e.queue.Dequeue()
	if m == nil {
		e.mu.Lock()
		e.phase = PhaseIdle
		e.mu.Unlock()
		e.trace("system", "no missions pending")
		e.emitStatus()
		return
	}

	missionStart := time.Now()
	e.trace("mission", fmt.Sprintf("processing mission: %s", m.Task))
	e.emitStatus()

	if e.mind != nil {
		e.mind.Think(consciousness.ThoughtDecision, "engine", "plan", m.Task)
	}

	relevantSkills := e.skills.Search(m.Task, 5)
	if len(relevantSkills) > 0 {
		e.trace("skill", fmt.Sprintf("found %d relevant skills for: %s", len(relevantSkills), m.Task))
		for _, sk := range relevantSkills {
			e.trace("skill", fmt.Sprintf("  ↳ %s (%s)", sk.Name, sk.Description))
		}
	}

	// Consult consciousness patterns for informed task planning
	if e.mind != nil {
		patterns := e.mind.Profile().TopPatterns(3)
		for _, p := range patterns {
			e.trace("consciousness", fmt.Sprintf("  pattern: %s (%s, confidence %.0f%%)", p.Name, p.Type, p.Confidence*100))
			e.mind.Observe("pattern-matched", "engine", []string{p.Type, p.Name})
		}
	}

	tasks := e.planTasks(m, relevantSkills)

	e.mu.Lock()
	e.phase = PhaseDispatching
	e.mu.Unlock()

	if e.mind != nil {
		e.mind.Think(consciousness.ThoughtDecision, "engine", "dispatch", fmt.Sprintf("%d tasks queued", len(tasks)))
	}
	for _, t := range tasks {
		if err := e.fleet.Enqueue(t); err != nil {
			e.trace("error", fmt.Sprintf("enqueue task: %v", err))
			if e.healer != nil {
				e.healer.Handle("engine", err)
			}
		}
	}

	e.mu.Lock()
	e.phase = PhaseExecuting
	e.mu.Unlock()

	if e.mind != nil {
		e.mind.Think(consciousness.ThoughtDecision, "engine", "execute", fmt.Sprintf("mission %s dispatched to fleet", m.ID))
	}

	e.dispatchFleet()
	results := e.collectResults(m.ID)

	e.mu.Lock()
	e.phase = PhaseCollecting
	e.mu.Unlock()

	if e.mind != nil {
		e.mind.Think(consciousness.ThoughtObservation, "engine", "collect", fmt.Sprintf("%d results from mission %s", len(results), m.ID))
	}

	e.processResults(m, results)

	e.mu.Lock()
	e.phase = PhaseLearning
	e.mu.Unlock()

	e.learnFromMission(m, results)

	missionDur := time.Since(missionStart)
	e.metricsMu.Lock()
	e.metrics.MissionsProcessed++
	e.metrics.TasksCompleted += int64(len(results))
	e.metrics.TotalDuration += missionDur
	e.metrics.LastMissionDuration = missionDur
	e.metricsMu.Unlock()

	e.trace("mission", fmt.Sprintf("mission complete in %s: %s", missionDur.Round(time.Millisecond), m.Task))

	if e.mind != nil {
		e.mind.Think(consciousness.ThoughtLearn, "engine", "complete",
			fmt.Sprintf("mission %s done in %s, %d results", m.ID, missionDur.Round(time.Millisecond), len(results)))
		_ = e.mind.Save()
	}

	e.mu.Lock()
	e.phase = PhaseIdle
	e.mu.Unlock()
	e.emitStatus()
}

func (e *Engine) dispatchFleet() {
	defer func() {
		if r := recover(); r != nil {
			e.trace("error", fmt.Sprintf("dispatch panic: %v", r))
		}
	}()

	ctx := e.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	for i := 0; i < 100; i++ {
		e.fleet.AssignNext(ctx)
		pending := e.fleet.PendingTasks()
		if len(pending) == 0 {
			return
		}
	}
	e.trace("system", "dispatch reached max iterations, some tasks may be pending")
}

func (e *Engine) collectResults(missionID string) (results []agent.TaskResult) {
	defer func() {
		if r := recover(); r != nil {
			e.trace("error", fmt.Sprintf("collect panic: %v", r))
			results = e.fleet.CompletedTasks()
		}
	}()

	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			e.trace("error", "collection timed out after 5 minutes")
			return e.fleet.CompletedTasks()
		case <-ticker.C:
			e.dispatchFleet()
			results := e.fleet.CompletedTasks()
			pending := e.fleet.PendingTasks()
			if len(pending) == 0 && len(results) > 0 {
				return results
			}
		}
	}
}

func (e *Engine) planTasks(m *mission.Mission, relevantSkills []*skill.Skill) []agent.Task {
	var tasks []agent.Task

	contextData := map[string]interface{}{
		"mission_id":   m.ID,
		"mission_task": m.Task,
		"tags":         m.Tags,
	}

	if e.mind != nil {
		patterns := e.mind.Profile().TopPatterns(5)
		if len(patterns) > 0 {
			var patternNames []string
			for _, p := range patterns {
				patternNames = append(patternNames, p.Name)
			}
			contextData["consciousness_patterns"] = patternNames
		}
	}

	if e.knowledge != nil {
		recent := e.knowledge.QueryContext(5)
		contextData["context_entries"] = len(recent)
		if len(recent) > 0 {
			var refs []string
			for _, k := range recent {
				refs = append(refs, k.Title)
			}
			contextData["recent_knowledge"] = refs
		}
	}

	if len(relevantSkills) > 0 {
		var skillNames []string
		for _, sk := range relevantSkills {
			skillNames = append(skillNames, sk.Name)
		}
		contextData["relevant_skills"] = skillNames
	}

	tasks = append(tasks, agent.Task{
		ID:          fmt.Sprintf("%s-analysis", m.ID),
		Role:        agent.RoleArchitect,
		Description: fmt.Sprintf("Analyze mission: %s", m.Task),
		Payload:     contextData,
		Priority:    agent.PriorityHigh,
	})

	execPayload := map[string]interface{}{
		"mission_id":   m.ID,
		"mission_task": m.Task,
		"model_hint":   m.ModelHint,
	}

	if e.knowledge != nil {
		execPayload["recent_knowledge"] = contextData["recent_knowledge"]
	}

	tasks = append(tasks, agent.Task{
		ID:          fmt.Sprintf("%s-exec", m.ID),
		Role:        agent.RoleRefactor,
		Description: fmt.Sprintf("Execute: %s", m.Task),
		Payload:     execPayload,
		Priority:    agent.PriorityNormal,
		DependsOn:   []string{fmt.Sprintf("%s-analysis", m.ID)},
	})

	tasks = append(tasks, agent.Task{
		ID:          fmt.Sprintf("%s-review", m.ID),
		Role:        agent.RoleReviewer,
		Description: fmt.Sprintf("Review results for: %s", m.Task),
		Payload: map[string]interface{}{
			"mission_id": m.ID,
		},
		Priority: agent.PriorityNormal,
		DependsOn: []string{fmt.Sprintf("%s-exec", m.ID)},
	})

	return tasks
}

func (e *Engine) processResults(m *mission.Mission, results []agent.TaskResult) {
	pendingLearnTasks := 0
	for _, t := range e.fleet.PendingTasks() {
		if strings.Contains(t.ID, "-learn") {
			pendingLearnTasks++
		}
	}

	for _, r := range results {
		level := "task"
		if !r.Success {
			level = "error"
			if e.healer != nil {
				e.healer.Handle("engine:task", fmt.Errorf("task %s failed: %s", r.TaskID, r.Error))
			}
		}
		e.trace(level, fmt.Sprintf("task %s: %s", r.TaskID, r.Output))

		if pendingLearnTasks >= 5 {
			continue
		}

		e.fleet.Enqueue(agent.Task{
			ID:          fmt.Sprintf("%s-learn", r.TaskID),
			Role:        agent.RoleDocWriter,
			Description: fmt.Sprintf("Record knowledge from task %s", r.TaskID),
			Payload: map[string]interface{}{
				"task_id": r.TaskID,
				"output":  r.Output,
				"success": r.Success,
			},
			Priority: agent.PriorityLow,
		})
		pendingLearnTasks++
	}
}

func (e *Engine) learnFromMission(m *mission.Mission, results []agent.TaskResult) {
	if e.mind != nil {
		e.mind.Observe("mission-"+m.Task, "mission", append(m.Tags, "completed"))
		for _, r := range results {
			if r.Success {
				e.mind.Observe("success", "outcome", []string{r.TaskID})
				e.mind.LogConvention(fmt.Sprintf("task %s succeeded via %s", r.TaskID, r.Output[:min(len(r.Output), 80)]))
			} else {
				e.mind.Observe("failure", "outcome", []string{r.TaskID, r.Error})
				e.mind.LogConvention(fmt.Sprintf("task %s failed: %s", r.TaskID, r.Error))
			}
		}
	}

	if e.knowledge == nil {
		return
	}

	body := fmt.Sprintf("Mission: %s\nTags: %v\nResults: %d tasks\n", m.Task, m.Tags, len(results))
	successes := 0
	for _, r := range results {
		if r.Success {
			successes++
		}
	}
	body += fmt.Sprintf("Success rate: %d/%d\n", successes, len(results))

	e.knowledge.Add(knowledge.KnowledgeEntry{
		Type:  knowledge.TypeDecision,
		Title: fmt.Sprintf("Mission: %s", truncateStr(m.Task, 80)),
		Body:  body,
		Tags:  append(m.Tags, "mission", "completed"),
	})

	if successes == len(results) && len(results) > 0 {
		skillName := e.deriveSkillName(m.Task)
		steps := e.deriveSteps(m, results)
		tags := append(m.Tags, "auto-generated", "skill")

		skill := e.skills.Generate(skillName, m.Task, m.Task, steps, tags)

		if e.skills != nil {
			doc := e.generateSkillDoc(skill, m, results)
			e.skills.SaveDoc(skill.ID, doc)
		}

		e.metricsMu.Lock()
		e.metrics.SkillsGenerated++
		e.metricsMu.Unlock()

		e.trace("skill", fmt.Sprintf("generated skill: %s (%s) — %d steps, usable for similar tasks", skill.Name, skill.ID, len(steps)))
	}
}

func (e *Engine) deriveSkillName(task string) string {
	words := strings.Fields(task)
	var clean []string
	for _, w := range words {
		w = strings.Trim(w, ".,!?;:\"'()[]{}")
		if len(w) > 2 || w == "go" || w == "do" || w == "be" || w == "to" || w == "in" || w == "on" || w == "at" || w == "by" || w == "of" || w == "or" || w == "as" {
			clean = append(clean, w)
		}
	}
	if len(clean) > 5 {
		clean = clean[:5]
	}
	if len(clean) == 0 {
		return "Auto-task"
	}
	prefix := "Skill-"
	for i, w := range clean {
		if len(w) > 1 {
			clean[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return prefix + strings.Join(clean, "-")
}

func (e *Engine) deriveSteps(m *mission.Mission, results []agent.TaskResult) []string {
	steps := []string{
		fmt.Sprintf("Analyze: %s", m.Task),
		"Decompose into sub-tasks by role",
		"Dispatch to specialized fleet agents",
		"Collect and verify results",
		"Record knowledge from outcomes",
	}
	if len(results) > 0 {
		successMsg := "Verify all tasks pass review"
		for _, r := range results {
			if !r.Success && r.Error != "" {
				successMsg = fmt.Sprintf("Handle errors: %s", truncateStr(r.Error, 60))
				break
			}
		}
		steps = append(steps, successMsg)
	}
	return steps
}

func (e *Engine) generateSkillDoc(skill *skill.Skill, m *mission.Mission, results []agent.TaskResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", skill.Name))
	b.WriteString(fmt.Sprintf("> **Auto-generated skill** from mission: _%s_\n\n", m.Task))
	b.WriteString("## Description\n\n")
	b.WriteString(fmt.Sprintf("This skill encapsulates the procedure for: %s\n\n", m.Task))
	b.WriteString("## Tags\n\n")
	for _, tag := range skill.Tags {
		b.WriteString(fmt.Sprintf("- `%s`\n", tag))
	}
	b.WriteString("\n## Steps\n\n")
	for i, step := range skill.Steps {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}
	b.WriteString("\n## Pattern\n\n")
	b.WriteString("```\n" + skill.Pattern + "\n```\n\n")
	b.WriteString("## Usage\n\n")
	b.WriteString("Apply this skill when the task matches the pattern above. ")
	b.WriteString("The fleet dispatcher will automatically match this skill to similar tasks.\n\n")
	b.WriteString("## Performance\n\n")

	successes := 0
	for _, r := range results {
		if r.Success {
			successes++
		}
	}
	total := len(results)
	if total > 0 {
		pct := float64(successes) / float64(total) * 100
		b.WriteString(fmt.Sprintf("- Success rate: **%.0f%%** (%d/%d)\n", pct, successes, total))
	}
	b.WriteString(fmt.Sprintf("- Total executions: **1** (freshly created)\n"))
	b.WriteString(fmt.Sprintf("- Confidence: **learning** (needs more executions)\n\n"))
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("*Created by CURSE Auto-Skill System at %s*\n", time.Now().UTC().Format(time.RFC3339)))
	return b.String()
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
