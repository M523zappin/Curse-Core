package gateway

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/M523zappin/Curse-Core/internal/agent"
	"github.com/M523zappin/Curse-Core/internal/computer"
	"github.com/M523zappin/Curse-Core/internal/consciousness"
	"github.com/M523zappin/Curse-Core/internal/engine"
	"github.com/M523zappin/Curse-Core/internal/governance"
	"github.com/M523zappin/Curse-Core/internal/healing"
	"github.com/M523zappin/Curse-Core/internal/knowledge"
	"github.com/M523zappin/Curse-Core/internal/lsp"
	"github.com/M523zappin/Curse-Core/internal/mission"
	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/sandbox"
	"github.com/M523zappin/Curse-Core/internal/scheduler"
	"github.com/M523zappin/Curse-Core/internal/session"
	"github.com/M523zappin/Curse-Core/internal/skill"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
	"github.com/M523zappin/Curse-Core/internal/sync"
	"github.com/google/uuid"
)

type AdapterFactory func(ModelProfile) Adapter

type Gateway struct {
	machine         *statemachine.Machine
	eventLog        *persistence.EventLog
	checkpoint      *persistence.CheckpointStore
	staging         *sandbox.StagingArea
	queue           *mission.Queue
	reviewer        *governance.Reviewer
	registry        *ModelRegistry
	activeModel     string
	adapter         Adapter
	truncator       *ContextTruncator
	configDir       string
	curseDir        string
	repoPath        string
	syncer          *sync.Syncer
	adapterRegistry map[string]AdapterFactory
	syncOnInit      bool
	computer        *computer.ComputerController
	toolRegistry    *ToolRegistry
	visionEngine    *computer.VisionEngine
	safetyChecker   *computer.SafetyChecker
	reviewManager   *computer.ReviewManager
	fleet           *agent.Fleet
	healer          *healing.HealingLoop
	knowledge       *knowledge.Index
	lspClient       *lsp.Client
	lspConnected    bool

	engine      *engine.Engine
	skills      *skill.Store
	sched       *scheduler.Scheduler
	sessionSt   *session.Store
	sessionID   string
	startTime   time.Time
	resumed     bool

	memory         *MemoryStore
	budget         *engine.IterationBudget
	consciousness  *consciousness.Consciousness
	modelsPath     string
}

func New(curseDir, configDir string) *Gateway {
	gw := &Gateway{
		machine:         statemachine.New(),
		queue:           mission.NewQueue(),
		curseDir:        curseDir,
		configDir:       configDir,
		adapterRegistry: make(map[string]AdapterFactory),
		sessionID:       uuid.NewString(),
		startTime:       time.Now(),
	}
	if cwd, err := os.Getwd(); err == nil {
		gw.repoPath = cwd
	} else {
		gw.repoPath = filepath.Dir(configDir)
	}
	if _, err := os.Stat(filepath.Join(gw.repoPath, ".git")); err == nil {
		gw.syncer = sync.New(
			"https://github.com/M523zappin/Curse-Core.git",
			gw.repoPath,
		)
	}
	return gw
}

func (g *Gateway) RegisterAdapter(provider string, factory AdapterFactory) {
	g.adapterRegistry[provider] = factory
}

func (g *Gateway) Init(ctx context.Context) error {
	if err := persistence.InitCurseDir(g.curseDir); err != nil {
		return fmt.Errorf("init curse dir: %w", err)
	}
	logPath := filepath.Join(g.curseDir, "logs", "event.log")
	g.eventLog = persistence.NewEventLog(logPath)
	cpPath := filepath.Join(g.curseDir, "logs", "session.json")
	g.checkpoint = persistence.NewCheckpointStore(cpPath)
	g.staging = sandbox.New(
		filepath.Join(g.curseDir, "staging"),
		sandbox.ModeDraftFile,
	)

	g.computer = computer.New()
	g.visionEngine = computer.NewVisionEngine(filepath.Join(g.curseDir, "screenshots"))
	g.safetyChecker = computer.NewSafetyChecker(g.visionEngine)
	g.reviewManager = computer.NewReviewManager()
	g.toolRegistry = NewToolRegistry(g.computer)

	g.computer.SetReviewCallback(func(req computer.ReviewRequest) {
		if g.reviewManager != nil {
			g.reviewManager.SetCallback(func(r computer.ReviewRequest) {
				_ = r
			})
		}
	})

	// ── Consciousness Engine (time-travel journal + soul) ─
	g.InitConsciousness()

	// ── Memory System (frozen-snapshot) ───────────────────
	g.InitMemory()

	// ── Iteration Budget ──────────────────────────────────
	g.budget = engine.NewIterationBudget(0) // 0 = unlimited

	// ── Autonomous Architectural Backbone ─────────────────
	g.InitFleet()
	g.InitHealer()
	g.InitKnowledge()
	g.InitSkills()
	g.InitEngine()
	g.InitScheduler()
	g.InitSession()

	if lspPath := lsp.FindLSServer("go"); lspPath != "" {
		go g.InitLSP("go")
	}

	g.modelsPath = filepath.Join(g.configDir, "models.json")
	if _, err := os.Stat(g.modelsPath); err == nil {
		reg, err := LoadModels(g.modelsPath)
		if err != nil {
			return fmt.Errorf("load models: %w", err)
		}
		g.registry = reg
		if err := g.activateProfile(reg.Active); err != nil {
			return fmt.Errorf("activate profile %s: %w", reg.Active, err)
		}
	} else {
		reg := GenerateDefaultLocalRegistry(ctx)
		if reg != nil && len(reg.Profiles) > 0 {
			g.registry = reg
			if err := g.activateProfile(reg.Active); err != nil {
				g.registry = nil
			}
		}
	}
	constPath := filepath.Join(g.configDir, "..", "CONSTITUTION.md")
	if _, err := os.Stat(constPath); err == nil {
		constData, err := os.ReadFile(constPath)
		if err == nil {
			_ = constData
		}
	}
	g.machine.OnTransition(func(result statemachine.TransitionResult) {
		entry, _ := g.eventLog.Append(result.From, result.Event, result.To, nil)
		if g.visionBufferSnapshot() != "" {
			_ = entry
		}
	})

	if g.healer != nil {
		g.machine.OnTransition(func(result statemachine.TransitionResult) {
			if result.To == statemachine.StateError {
				g.healer.Handle("state-machine", fmt.Errorf("transitioned to Error from %s via %s",
					result.From, result.Event))
			}
		})
	}

	if g.syncOnInit && g.syncer != nil {
		if changed, err := g.SyncConstitution(); err == nil && changed {
			fmt.Println("✓ Constitution synced from remote")
		}
	}

	return nil
}

func (g *Gateway) InitConsciousness() {
	if g.consciousness != nil {
		return
	}
	c, err := consciousness.New(g.curseDir)
	if err != nil {
		return
	}
	g.consciousness = c
}

func (g *Gateway) Consciousness() *consciousness.Consciousness {
	return g.consciousness
}

func (g *Gateway) InitMemory() {
	if g.memory != nil {
		return
	}
	g.memory = NewMemoryStore(g.curseDir)
	if err := g.memory.Load(); err != nil {
		g.memory = nil
	}
}

func (g *Gateway) Memory() *MemoryStore {
	return g.memory
}

func (g *Gateway) Budget() *engine.IterationBudget {
	return g.budget
}

func (g *Gateway) InitSkills() {
	if g.skills != nil {
		return
	}
	skillsDir := filepath.Join(g.curseDir, "skills")
	g.skills = skill.NewStore(skillsDir)
	g.knowledge.Add(knowledge.KnowledgeEntry{
		Type:  knowledge.TypeDecision,
		Title: "Skill System initialized",
		Body:  fmt.Sprintf("Skill store at %s with %d existing skills", skillsDir, g.skills.Count()),
		Tags:  []string{"skill", "init"},
	})
}

func (g *Gateway) InitEngine() {
	if g.engine != nil {
		return
	}
	g.engine = engine.New(
		g.queue,
		g.fleet,
		g.skills,
		g.knowledge,
		g.healer,
		g.consciousness,
	)
	g.engine.SetTraceHook(func(level, msg string) {
		_ = level
		_ = msg
	})
	g.engine.SetStatusHook(func(s engine.Status) {
		_ = s
	})
}

func (g *Gateway) InitScheduler() {
	if g.sched != nil {
		return
	}
	g.sched = scheduler.New()

	g.sched.Add("health-check", 5*time.Minute, func(ctx context.Context) error {
		issues := 0
		if g.engine != nil && !g.engine.Running() {
			issues++
		}
		if g.knowledge == nil {
			issues++
		}
		if g.fleet == nil {
			issues++
		}
		if issues > 0 {
			return fmt.Errorf("%d subsystem(s) unhealthy", issues)
		}
		return nil
	})

	g.sched.Add("save-session", 30*time.Second, func(ctx context.Context) error {
		return g.SaveSession()
	})
}

func (g *Gateway) InitSession() {
	if g.sessionSt != nil {
		return
	}
	g.sessionSt = session.NewStore(g.curseDir)
}

func (g *Gateway) StartEngine(ctx context.Context) {
	if g.engine != nil && !g.engine.Running() {
		go g.engine.Run(ctx)
	}
}

func (g *Gateway) StartScheduler(ctx context.Context) {
	if g.sched != nil && !g.sched.Running() {
		go g.sched.Run(ctx)
	}
}

func (g *Gateway) SaveSession() error {
	if g.sessionSt == nil {
		return nil
	}
	state := session.State{
		SessionID:      g.sessionID,
		StartedAt:      g.startTime,
		ActiveModel:    g.activeModel,
		MachineState:   g.machine.State().String(),
		MachineStep:    g.machine.Step(),
		KnowledgeCount: g.knowledge.Count(),
		SkillCount:     g.skills.Count(),
		TaskCount:      g.fleet.TaskCount(),
	}
	return g.sessionSt.Save(state)
}

func (g *Gateway) RestoreSession() (*session.State, error) {
	if g.sessionSt == nil {
		return nil, fmt.Errorf("session store not initialized")
	}
	if !g.sessionSt.Exists() {
		return nil, nil
	}
	state, err := g.sessionSt.Load()
	if err != nil {
		return nil, err
	}
	g.resumed = true
	g.sessionID = state.SessionID
	g.startTime = state.StartedAt
	if state.ActiveModel != "" {
		_ = g.activateProfile(state.ActiveModel)
	}
	return state, nil
}

func (g *Gateway) Engine() *engine.Engine {
	return g.engine
}

func (g *Gateway) Skills() *skill.Store {
	return g.skills
}

func (g *Gateway) Scheduler() *scheduler.Scheduler {
	return g.sched
}

func (g *Gateway) SessionStore() *session.Store {
	return g.sessionSt
}

func (g *Gateway) SessionID() string {
	return g.sessionID
}

func (g *Gateway) StartTime() time.Time {
	return g.startTime
}

func (g *Gateway) Resumed() bool {
	return g.resumed
}

func (g *Gateway) Shutdown(ctx context.Context) error {
	if g.engine != nil {
		g.engine.Stop()
	}
	if g.sched != nil {
		g.sched.Stop()
	}
	if g.computer != nil {
		g.computer.StopBrowser()
	}
	if g.lspClient != nil {
		g.lspClient.Shutdown()
	}

	g.finishSession(ctx)
	g.SaveSession()

	return nil
}

func (g *Gateway) finishSession(ctx context.Context) {
	if g.knowledge == nil {
		return
	}
	duration := time.Since(g.startTime)
	taskCount := 0
	if g.fleet != nil {
		taskCount = g.fleet.TaskCount()
	}
	skillCount := 0
	if g.skills != nil {
		skillCount = g.skills.Count()
	}
	knCount := g.knowledge.Count()

	summary := fmt.Sprintf(
		"CURSE session %s completed.\nDuration: %s\nTasks: %d\nSkills: %d\nKnowledge entries: %d\nModel: %s",
		g.sessionID, knowledge.FormatDuration(duration), taskCount, skillCount, knCount, g.activeModel,
	)
	g.knowledge.RecordSession(g.sessionID, knowledge.SessionSummary{
		SessionID: g.sessionID,
		StartTime: g.startTime,
		Duration:  duration,
		TaskCount: taskCount,
		Summary:   summary,
	})
}

func (g *Gateway) activateProfile(name string) error {
	profile, ok := g.registry.GetProfile(name)
	if !ok {
		return fmt.Errorf("unknown profile: %s", name)
	}
	adapter := g.buildAdapter(profile)
	if adapter == nil {
		return fmt.Errorf("no adapter factory registered for provider %q", profile.Provider)
	}
	g.activeModel = name
	g.truncator = NewContextTruncator(profile)
	g.adapter = adapter
	return nil
}

func (g *Gateway) buildAdapter(profile ModelProfile) Adapter {
	factory, ok := g.adapterRegistry[profile.Provider]
	if !ok {
		return nil
	}
	return factory(profile)
}

func (g *Gateway) SwitchModel(name string) error {
	if g.registry == nil {
		return fmt.Errorf("no model registry loaded")
	}
	if err := g.activateProfile(name); err != nil {
		return err
	}
	g.registry.Active = name
	if g.modelsPath != "" {
		if err := g.registry.Save(g.modelsPath); err != nil {
			return fmt.Errorf("save registry: %w", err)
		}
	}
	return nil
}

func (g *Gateway) Adapter() Adapter {
	return g.adapter
}

func (g *Gateway) Truncator() *ContextTruncator {
	return g.truncator
}

func (g *Gateway) Machine() *statemachine.Machine {
	return g.machine
}

func (g *Gateway) EventLog() *persistence.EventLog {
	return g.eventLog
}

func (g *Gateway) Checkpoint() *persistence.CheckpointStore {
	return g.checkpoint
}

func (g *Gateway) Staging() *sandbox.StagingArea {
	return g.staging
}

func (g *Gateway) Queue() *mission.Queue {
	return g.queue
}

func (g *Gateway) ActiveModel() string {
	return g.activeModel
}

func (g *Gateway) Registry() *ModelRegistry {
	return g.registry
}

func (g *Gateway) SyncConstitution() (changed bool, err error) {
	if g.syncer == nil {
		return false, fmt.Errorf("no git repository found at %s", g.repoPath)
	}

	_ = g.machine.TriggerSync()

	constPath := filepath.Join(g.repoPath, "CONSTITUTION.md")
	changed, err = g.syncer.SyncConstitution(constPath)
	if err != nil {
		_ = g.machine.FailSync()
		return false, fmt.Errorf("sync constitution: %w", err)
	}

	_ = g.machine.CompleteSync()

	if changed {
		c, parseErr := governance.Parse(constPath)
		if parseErr == nil {
			g.reviewer = governance.NewReviewer(c)
		}
	}

	return changed, nil
}

func (g *Gateway) SyncStateChange(eventLogFile, sessionFile string) error {
	if g.syncer == nil {
		return nil
	}
	return sync.CommitAndPush(g.repoPath,
		fmt.Sprintf("curse: state sync — %s step=%d",
			g.machine.State(), g.machine.Step()))
}

func (g *Gateway) Syncer() *sync.Syncer {
	return g.syncer
}

func (g *Gateway) RepoPath() string {
	return g.repoPath
}

func (g *Gateway) InitFleet() {
	if g.fleet != nil {
		return
	}
	g.fleet = agent.NewFleet()
	g.fleet.RegisterRole(agent.RoleSecurity, 1)
	g.fleet.RegisterRole(agent.RoleRefactor, 2)
	g.fleet.RegisterRole(agent.RoleInfra, 1)
	g.fleet.RegisterRole(agent.RoleReviewer, 2)
	g.fleet.RegisterRole(agent.RoleTester, 1)
	g.fleet.RegisterRole(agent.RoleArchitect, 1)
	g.fleet.RegisterRole(agent.RoleDocWriter, 1)
	g.fleet.RegisterRole(agent.RoleDependency, 1)

	g.fleet.SetDispatcher(func(a *agent.Agent, t agent.Task) *agent.TaskResult {
		return g.dispatchAgentTask(a, t)
	})
}

func (g *Gateway) InitHealer() {
	if g.healer != nil {
		return
	}
	g.healer = healing.NewHealingLoop()
	g.healer.RegisterHandler("browser", func(inc healing.Incident) (string, bool, error) {
		if g.computer != nil {
			g.computer.StopBrowser()
			if err := g.computer.StartBrowser(); err == nil {
				return "browser restarted", true, nil
			}
		}
		return "browser restart failed", false, nil
	})
}

func (g *Gateway) InitKnowledge() {
	if g.knowledge != nil {
		return
	}
	indexDir := filepath.Join(g.curseDir, "knowledge")
	g.knowledge = knowledge.NewIndex(indexDir)
	g.knowledge.RecordADR("CURSE Platform Architecture",
		"CURSE is a persistent autonomous TUI orchestration platform with crash-recoverable state machine, "+
			"draft-staging sandbox, model-agnostic gateway adapters, and a professional-grade Bubble Tea dashboard.\n\n"+
			"Key architectural decisions:\n"+
			"- SHA256-chained event.log for tamper-evident crash recovery\n"+
			"- State machine with 8 states including Syncing\n"+
			"- Gateway uses AdapterFactory registry pattern\n"+
			"- Computer Controller with Playwright-based browser automation\n"+
			"- Human-in-the-Loop review mode for destructive actions",
		[]string{"architecture", "go", "tui", "state-machine"})
	_ = g.knowledge
}

func (g *Gateway) InitLSP(language string) {
	if g.lspConnected {
		return
	}
	serverPath := lsp.FindLSServer(language)
	if serverPath == "" {
		return
	}
	g.lspClient = lsp.NewClient(serverPath, g.repoPath)
	if err := g.lspClient.Connect(context.Background()); err != nil {
		g.lspConnected = false
		return
	}
	g.lspConnected = true
	g.openProjectFiles()
}

func (g *Gateway) openProjectFiles() {
	if !g.lspConnected || g.lspClient == nil {
		return
	}
	filepath.Walk(g.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".go", ".ts", ".js", ".py", ".rs", ".json", ".md":
			data, readErr := os.ReadFile(path)
			if readErr == nil {
				g.lspClient.OpenDocument(path, string(data))
			}
		}
		return nil
	})
}

func (g *Gateway) dispatchAgentTask(a *agent.Agent, t agent.Task) *agent.TaskResult {
	result := &agent.TaskResult{
		TaskID:  t.ID,
		Success: true,
		Output:  fmt.Sprintf("agent %s (%s) completed %s", a.ID, a.Role, t.Description),
	}

	if g.healer != nil {
		g.healer.Handle("agent:"+string(a.Role), fmt.Errorf("task executed: %s", t.ID))
	}

	if g.knowledge != nil {
		g.knowledge.Add(knowledge.KnowledgeEntry{
			Type:  knowledge.TypeDecision,
			Title: fmt.Sprintf("Agent %s executed task %s", a.ID, t.ID),
			Body:  t.Description,
			Tags:  []string{string(a.Role), "task"},
		})
	}

	return result
}

func (g *Gateway) SetSyncOnInit(v bool) {
	g.syncOnInit = v
}

func (g *Gateway) Fleet() *agent.Fleet {
	return g.fleet
}

func (g *Gateway) Healer() *healing.HealingLoop {
	return g.healer
}

func (g *Gateway) Knowledge() *knowledge.Index {
	return g.knowledge
}

func (g *Gateway) LSP() *lsp.Client {
	return g.lspClient
}

func (g *Gateway) LSPConnected() bool {
	return g.lspConnected
}

func (g *Gateway) Computer() *computer.ComputerController {
	return g.computer
}

func (g *Gateway) ToolRegistry() *ToolRegistry {
	return g.toolRegistry
}

func (g *Gateway) VisionEngine() *computer.VisionEngine {
	return g.visionEngine
}

func (g *Gateway) SafetyChecker() *computer.SafetyChecker {
	return g.safetyChecker
}

func (g *Gateway) ReviewManager() *computer.ReviewManager {
	return g.reviewManager
}

func (g *Gateway) visionBufferSnapshot() string {
	if g.computer == nil {
		return ""
	}
	buf := g.computer.VisionBuffer()
	if len(buf) == 0 {
		return ""
	}
	last := buf[len(buf)-1]
	return last.Screenshot
}
