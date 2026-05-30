package gateway

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/M523zappin/Curse-Core/internal/governance"
	"github.com/M523zappin/Curse-Core/internal/mission"
	"github.com/M523zappin/Curse-Core/internal/persistence"
	"github.com/M523zappin/Curse-Core/internal/sandbox"
	"github.com/M523zappin/Curse-Core/internal/statemachine"
	"github.com/M523zappin/Curse-Core/internal/sync"
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
}

func New(curseDir, configDir string) *Gateway {
	gw := &Gateway{
		machine:         statemachine.New(),
		queue:           mission.NewQueue(),
		curseDir:        curseDir,
		configDir:       configDir,
		adapterRegistry: make(map[string]AdapterFactory),
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
	modelsPath := filepath.Join(g.configDir, "models.json")
	if _, err := os.Stat(modelsPath); err == nil {
		reg, err := LoadModels(modelsPath)
		if err != nil {
			return fmt.Errorf("load models: %w", err)
		}
		g.registry = reg
		if err := g.activateProfile(reg.Active); err != nil {
			return fmt.Errorf("activate profile %s: %w", reg.Active, err)
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
		g.eventLog.Append(result.From, result.Event, result.To, nil)
	})

	if g.syncOnInit && g.syncer != nil {
		if changed, err := g.SyncConstitution(); err == nil && changed {
			fmt.Println("✓ Constitution synced from remote")
		}
	}

	return nil
}

func (g *Gateway) activateProfile(name string) error {
	profile, ok := g.registry.GetProfile(name)
	if !ok {
		return fmt.Errorf("unknown profile: %s", name)
	}
	g.activeModel = name
	g.truncator = NewContextTruncator(profile)
	g.adapter = g.buildAdapter(profile)
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
	return g.activateProfile(name)
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

func (g *Gateway) SetSyncOnInit(v bool) {
	g.syncOnInit = v
}
