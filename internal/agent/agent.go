package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AgentRole string

const (
	RoleSecurity     AgentRole = "security-auditor"
	RoleRefactor     AgentRole = "refactoring"
	RoleInfra        AgentRole = "infrastructure"
	RoleReviewer     AgentRole = "code-reviewer"
	RoleTester       AgentRole = "testing"
	RoleArchitect    AgentRole = "architect"
	RoleDependency   AgentRole = "dependency-manager"
	RoleDocWriter    AgentRole = "documentation"
)

type AgentStatus string

const (
	StatusIdle       AgentStatus = "idle"
	StatusActive     AgentStatus = "active"
	StatusBlocked    AgentStatus = "blocked"
	StatusFailed     AgentStatus = "failed"
	StatusCompleted  AgentStatus = "completed"
)

type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityNormal TaskPriority = 1
	PriorityHigh   TaskPriority = 2
	PriorityCritical TaskPriority = 3
)

type Task struct {
	ID          string                 `json:"id"`
	Role        AgentRole              `json:"role"`
	Description string                 `json:"description"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    TaskPriority           `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	DependsOn   []string               `json:"depends_on"`
}

type TaskResult struct {
	TaskID    string            `json:"task_id"`
	Success   bool              `json:"success"`
	Output    string            `json:"output"`
	Artifacts []string          `json:"artifacts"`
	Error     string            `json:"error,omitempty"`
	Metrics   map[string]int64  `json:"metrics,omitempty"`
}

type Agent struct {
	Role        AgentRole    `json:"role"`
	ID          string       `json:"id"`
	Status      AgentStatus  `json:"status"`
	CurrentTask string       `json:"current_task,omitempty"`
	StartedAt   time.Time    `json:"started_at"`
	TaskCount   int          `json:"task_count"`
	mu          *sync.Mutex
	taskQueue   []Task
}

type Fleet struct {
	mu       sync.RWMutex
	agents   map[string]*Agent
	tasks    map[string]*Task
	results  map[string]*TaskResult
	dispatch func(agent *Agent, task Task) *TaskResult
	roles    map[AgentRole]int
}

func NewFleet() *Fleet {
	return &Fleet{
		agents:  make(map[string]*Agent),
		tasks:   make(map[string]*Task),
		results: make(map[string]*TaskResult),
		roles:   make(map[AgentRole]int),
	}
}

func (f *Fleet) RegisterRole(role AgentRole, count int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.roles[role] = count
	for i := 0; i < count; i++ {
		agent := &Agent{
			Role:      role,
			ID:        fmt.Sprintf("%s-%d", role, i+1),
			Status:    StatusIdle,
			StartedAt: time.Now(),
			mu:        &sync.Mutex{},
		}
		f.agents[agent.ID] = agent
	}
}

func (f *Fleet) SetDispatcher(d func(agent *Agent, task Task) *TaskResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dispatch = d
}

func (f *Fleet) Enqueue(task Task) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	if _, ok := f.roles[task.Role]; !ok && len(f.agents) > 0 {
		return fmt.Errorf("no agents registered for role %s", task.Role)
	}

	task.CreatedAt = time.Now()
	f.tasks[task.ID] = &task
	return nil
}

func (f *Fleet) AssignNext(ctx context.Context) *TaskResult {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, agent := range f.agents {
		if agent.Status != StatusIdle {
			continue
		}

		bestTask, bestPriority := (*Task)(nil), TaskPriority(-1)
		for _, task := range f.tasks {
			if task.Role != agent.Role {
				continue
			}
			if _, exists := f.results[task.ID]; exists {
				continue
			}
			if !f.dependenciesMet(task) {
				continue
			}
			if task.Priority > bestPriority {
				bestTask = task
				bestPriority = task.Priority
			}
		}

		if bestTask == nil {
			continue
		}

		agent.Status = StatusActive
		agent.CurrentTask = bestTask.ID
		agent.TaskCount++

		go func(a *Agent, t Task) {
			var recovered bool
			defer func() {
				if r := recover(); r != nil {
					recovered = true
					f.mu.Lock()
					a.Status = StatusIdle
					a.CurrentTask = ""
					f.results[t.ID] = &TaskResult{
						TaskID:  t.ID,
						Success: false,
						Error:   fmt.Sprintf("panic: %v", r),
					}
					f.mu.Unlock()
				}
			}()
			result := f.execute(a, t)
			if recovered {
				return
			}
			f.mu.Lock()
			a.Status = StatusIdle
			a.CurrentTask = ""
			f.results[t.ID] = result
			f.mu.Unlock()
		}(agent, *bestTask)

		return nil
	}

	return nil
}

func (f *Fleet) dependenciesMet(task *Task) bool {
	for _, depID := range task.DependsOn {
		result, exists := f.results[depID]
		if !exists {
			return false
		}
		_ = result
	}
	return true
}

func (f *Fleet) execute(agent *Agent, task Task) *TaskResult {
	if f.dispatch != nil {
		return f.dispatch(agent, task)
	}
	return &TaskResult{
		TaskID:  task.ID,
		Success: true,
		Output:  fmt.Sprintf("agent %s processed task %s", agent.ID, task.ID),
	}
}

func (f *Fleet) AgentStatus() []Agent {
	f.mu.RLock()
	defer f.mu.RUnlock()
	agents := make([]Agent, 0, len(f.agents))
	for _, a := range f.agents {
		a.mu.Lock()
		agents = append(agents, *a)
		a.mu.Unlock()
	}
	return agents
}

func (f *Fleet) PendingTasks() []Task {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var pending []Task
	for _, t := range f.tasks {
		if _, done := f.results[t.ID]; !done {
			pending = append(pending, *t)
		}
	}
	return pending
}

func (f *Fleet) CompletedTasks() []TaskResult {
	f.mu.RLock()
	defer f.mu.RUnlock()
	results := make([]TaskResult, 0, len(f.results))
	for _, r := range f.results {
		results = append(results, *r)
	}
	return results
}

func (f *Fleet) TaskCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.tasks)
}

func (f *Fleet) AgentCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.agents)
}

func (f *Fleet) RoleCount(role AgentRole) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.roles[role]
}
