package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"hatch/ai/leaderboard"
	"hatch/ai/llm"
	"hatch/core/config"
	"hatch/memo"
	"hatch/workbench/executor"
	"hatch/workbench/scheduler"
	"hatch/workbench/store"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the single Wails-bound object (window.go.main.App.*). The root
// package is only the binding layer: each file maps to one UI module and keeps
// methods thin — locking, persisting config and emitting events — while the
// real work lives in internal/<module>/...
//
//	app.go / system_*.go   lifecycle, module toggles, data dir, backup
//	workbench_*.go         🗂 工作台  → internal/workbench/*
//	ai_*.go                ✦ AI      → internal/ai/*
//	vpn.go                 🔒 VPN    → internal/vpn
//	memo*.go               ✓ 备忘录   → internal/memo
type App struct {
	ctx      context.Context
	cfg      *config.Config
	store    *store.Store
	services map[string]*executor.ServiceManager
	mu       sync.RWMutex
	mcpOpMu  sync.Mutex // serializes MCP config writes (CLI remove+add is not atomic)

	leaderboards *leaderboard.Service // lazily created by modelLeaderboards()

	tracesMu sync.Mutex
	traces   []llm.Trace
	traceSeq int64

	sched *scheduler.Scheduler

	memoStore  *memo.Store
	memoTimers map[string]*time.Timer
	memoMu     sync.Mutex
}

// maxTraces caps the in-memory AI trace ring used when the trace dir is unavailable.
const maxTraces = 50

func NewApp() *App {
	return &App{
		services:   make(map[string]*executor.ServiceManager),
		memoTimers: make(map[string]*time.Timer),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	a.cfg = cfg

	st, err := store.New()
	if err != nil {
		st, _ = store.NewWithPath("/tmp/hatch-scripts.json")
	}
	a.store = st

	if ms, err := memo.New(memo.DefaultStorePath()); err == nil {
		a.memoStore = ms
	}

	a.restoreServices()
	a.migrateAIProviders()
	a.ScheduleMemoReminders()

	// Retention (AI traces, run history, memo files) on startup, then hourly.
	go a.ApplyRetention()
	go func() {
		t := time.NewTicker(1 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				a.ApplyRetention()
			}
		}
	}()

	go func() {
		// run after a short delay so ScanAll has populated the store on first launch
		time.Sleep(500 * time.Millisecond)
		a.ApplyAllSchedules()
	}()
}

// restoreServices rebuilds ServiceManager instances for service-type scripts.
func (a *App) restoreServices() {
	for _, s := range a.store.All() {
		if s.Type == config.TypeService {
			sm := a.newServiceManager(s)
			a.mu.Lock()
			a.services[s.Path] = sm
			a.mu.Unlock()
			sm.Status()
		}
	}
}

// ─── AI plumbing shared by every module ──────────────────────────────────────

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
// Caller must hold a.mu.
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

// readyAI returns the active provider config, or an error explaining what is
// missing. Suggest* methods call it before doing any work.
func (a *App) readyAI() (config.AIConfig, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
	a.mu.RUnlock()
	return ai, llm.Validate(ai)
}

// callAIOp runs one AI op: it applies the user's prompt override for op (if
// any), calls the model, and records the trace (memory + JSONL + ai:trace event).
func (a *App) callAIOp(op string, ai config.AIConfig, systemPrompt, userPrompt string) (string, error) {
	if err := llm.Validate(ai); err != nil {
		return "", err
	}
	content, trace, err := llm.Chat(a.ctx, op, ai, a.effectivePrompt(op, systemPrompt), userPrompt)
	if trace.At != "" { // the request was actually sent
		a.tracesMu.Lock()
		a.traceSeq++
		trace.ID = a.traceSeq
		a.tracesMu.Unlock()
		a.recordTrace(trace)
	}
	return content, err
}

// effectivePrompt returns the user override (if any) for an op, else the built-in default.
func (a *App) effectivePrompt(op, defaultPrompt string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if v, ok := a.cfg.PromptTemplates[op]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	return defaultPrompt
}

func (a *App) emit(event string, payload any) {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, event, payload)
	}
}
