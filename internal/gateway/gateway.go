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
	adapterRegistry map[string]AdapterFactory
}

func New(curseDir, configDir string) *Gateway {
	return &Gateway{
		machine:         statemachine.New(),
		queue:           mission.NewQueue(),
		curseDir:        curseDir,
		configDir:       configDir,
		adapterRegistry: make(map[string]AdapterFactory),
	}
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
