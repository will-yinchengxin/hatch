package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scheduler"
	"hatch/internal/store"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// wailsEmitStream emits an AI stream chunk event to the frontend.
func wailsEmitStream(ctx context.Context, payload map[string]any) {
	wailsRuntime.EventsEmit(ctx, "ai:stream", payload)
}

// App is the Wails-bound root object. Methods are split across multiple
// files by responsibility (see app_scripts.go, app_executor.go, app_ai.go,
// app_zshrc.go, app_observability.go, app_files.go, skills.go) — Wails only
// looks at the method set, so the split is purely organizational.
type App struct {
	ctx      context.Context
	cfg      *config.Config
	store    *store.Store
	services map[string]*executor.ServiceManager
	mu       sync.RWMutex

	tracesMu sync.Mutex
	traces   []AICallTrace
	traceSeq int64

	sched *scheduler.Scheduler // v1.2
}

func NewApp() *App {
	return &App{
		services: make(map[string]*executor.ServiceManager),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{OverrideMeta: map[string]config.ScriptMeta{}}
	}
	a.cfg = cfg

	st, err := store.New()
	if err != nil {
		st, _ = store.NewWithPath("/tmp/hatch-scripts.json")
	}
	a.store = st

	a.restoreServices()
	a.migrateAIProviders()

	go func() {
		if dir, err := a.traceDir(); err == nil {
			pruneTraceFiles(dir, a.traceRetainDays())
		}
	}()

	// v1.1: prune stale run-history records on startup
	go executor.PruneAllHistory(50)

	// v1.2: apply persisted schedules
	go func() {
		// run after a short delay so ScanAll has populated the store on first launch
		time.Sleep(500 * time.Millisecond)
		a.ApplyAllSchedules()
	}()
}

// migrateAIProviders creates a provider entry from the legacy AI config if none exist yet.
func (a *App) migrateAIProviders() {
	if len(a.cfg.AIProviders) == 0 && a.cfg.AI.BaseURL != "" {
		id := fmt.Sprintf("p-%d", time.Now().UnixNano())
		a.cfg.AIProviders = []config.AIProvider{{
			ID:      id,
			Name:    "默认",
			BaseURL: a.cfg.AI.BaseURL,
			APIKey:  a.cfg.AI.APIKey,
			Model:   a.cfg.AI.Model,
		}}
		a.cfg.ActiveProviderID = id
		_ = a.cfg.Save()
	}
}

// activeAIConfig returns the AIConfig for the currently active provider.
func (a *App) activeAIConfig() config.AIConfig {
	if a.cfg.ActiveProviderID != "" {
		for _, p := range a.cfg.AIProviders {
			if p.ID == a.cfg.ActiveProviderID {
				return config.AIConfig{
					Enabled:         a.cfg.AI.Enabled,
					BaseURL:         p.BaseURL,
					APIKey:          p.APIKey,
					Model:           p.Model,
					TraceRetainDays: a.cfg.AI.TraceRetainDays,
				}
			}
		}
	}
	return a.cfg.AI
}

// restoreServices rebuilds ServiceManager instances for service-type scripts.
func (a *App) restoreServices() {
	for _, s := range a.store.All() {
		if s.Type == config.TypeService {
			sm := executor.NewServiceManager(s.Path, s.PidFile, executor.NewEmitter(a.ctx))
			a.mu.Lock()
			a.services[s.Path] = sm
			a.mu.Unlock()
			sm.Status()
		}
	}
}
