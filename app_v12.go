package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scheduler"
)

// ─── v1.2: Scheduler ─────────────────────────────────────────────────────────

// ensureScheduler lazily creates the scheduler.
func (a *App) ensureScheduler() *scheduler.Scheduler {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sched == nil {
		a.sched = scheduler.New()
	}
	return a.sched
}

// ApplyAllSchedules walks all known scripts and registers/refreshes scheduler entries
// based on each ScriptMeta.Schedule (and the Override map). Call at startup and after
// SaveScriptMeta.
func (a *App) ApplyAllSchedules() {
	sch := a.ensureScheduler()

	// Build effective meta map: override > store > scanned
	a.mu.RLock()
	overrides := make(map[string]config.ScriptMeta)
	for k, v := range a.cfg.OverrideMeta {
		overrides[k] = v
	}
	a.mu.RUnlock()

	seen := map[string]bool{}
	var metas []config.ScriptMeta
	if a.store != nil {
		metas = a.store.All()
	}
	for _, m := range metas {
		if ov, ok := overrides[m.Path]; ok {
			if ov.Schedule != "" {
				m.Schedule = ov.Schedule
			}
		}
		if strings.TrimSpace(m.Schedule) == "" {
			continue
		}
		mp := m
		path := m.Path
		err := sch.Schedule(path, m.Schedule, func() {
			em := executor.NewEmitter(a.ctx)
			args := []string{}
			go func() {
				_, _ = executor.RunOneshotWithEnv(path, path, args, a.scriptEnv(path), em, a.notifyHook(path))
			}()
			_ = mp // ensure capture
		})
		if err == nil {
			seen[path] = true
		}
	}

	// Remove jobs that are no longer in the set
	for _, j := range sch.List() {
		if !seen[j.ID] {
			sch.Remove(j.ID)
		}
	}
}

// ListSchedules returns the registered schedule jobs and their next fire time.
type ScheduleInfo struct {
	Path     string `json:"path"`
	Schedule string `json:"schedule"`
	NextRun  int64  `json:"nextRun"` // unix seconds; 0 if unknown
}

func (a *App) ListSchedules() []ScheduleInfo {
	sch := a.ensureScheduler()
	jobs := sch.List()
	out := make([]ScheduleInfo, 0, len(jobs))
	for _, j := range jobs {
		nr, _ := sch.NextRun(j.ID)
		info := ScheduleInfo{Path: j.ID, Schedule: j.Schedule}
		if !nr.IsZero() {
			info.NextRun = nr.Unix()
		}
		out = append(out, info)
	}
	return out
}

// RemoveSchedule unregisters a script's scheduler entry (does not change config).
func (a *App) RemoveSchedule(path string) {
	if a.sched != nil {
		a.sched.Remove(path)
	}
}

// ─── v1.2: Health & dependency check ─────────────────────────────────────────

type HealthCheckResult struct {
	Script         string   `json:"script"`
	MissingDeps    []string `json:"missingDeps"`
	HealthURL      string   `json:"healthUrl"`
	HealthOK       bool     `json:"healthOk"`
	HealthStatus   int      `json:"healthStatus"` // HTTP status code; 0 if no URL or unreachable
	HealthLatency  int64    `json:"healthLatencyMs"`
	HealthError    string   `json:"healthError"`
	HasAnnotations bool     `json:"hasAnnotations"`
}

// CheckScriptHealth verifies @hatch.requires deps and @hatch.health URL for one script.
// Backed by 2s HTTP timeout to avoid blocking the UI.
func (a *App) CheckScriptHealth(path string) HealthCheckResult {
	r := HealthCheckResult{Script: path}

	// Find meta — try override first, then store
	var meta config.ScriptMeta
	a.mu.RLock()
	if ov, ok := a.cfg.OverrideMeta[path]; ok {
		meta = ov
	}
	a.mu.RUnlock()
	if meta.Path == "" && a.store != nil {
		for _, m := range a.store.All() {
			if m.Path == path {
				meta = m
				break
			}
		}
	}

	r.HasAnnotations = len(meta.Requires) > 0 || meta.HealthURL != ""

	for _, dep := range meta.Requires {
		if findCLI(dep) == "" {
			r.MissingDeps = append(r.MissingDeps, dep)
		}
	}

	r.HealthURL = meta.HealthURL
	if meta.HealthURL != "" {
		client := &http.Client{Timeout: 2 * time.Second}
		start := time.Now()
		resp, err := client.Get(meta.HealthURL)
		r.HealthLatency = time.Since(start).Milliseconds()
		if err != nil {
			r.HealthError = err.Error()
		} else {
			defer resp.Body.Close()
			r.HealthStatus = resp.StatusCode
			r.HealthOK = resp.StatusCode >= 200 && resp.StatusCode < 400
		}
	}

	return r
}

// CheckSystemRequirements checks if a command exists on PATH, returning a quick yes/no.
func (a *App) CheckCommandAvailable(name string) bool {
	return findCLI(name) != ""
}

// ─── v1.2: AI streaming (minimal) ────────────────────────────────────────────

// SuggestScriptContentStream streams script content suggestions back via events.
// Emits "ai:stream" events with {id, chunk, done} payload.
// The current implementation calls the non-streaming endpoint and emits the
// result as one chunk — a placeholder that keeps the streaming API stable
// while leaving room to swap in true SSE later.
func (a *App) SuggestScriptContentStream(id, path, currentContent, instruction string) {
	go func() {
		text, err := a.SuggestScriptContent(path, currentContent, instruction)
		if err != nil {
			a.emitStream(id, "", true, err.Error())
			return
		}
		a.emitStream(id, text, true, "")
	}()
}

// emitStream sends one chunk to the frontend.
func (a *App) emitStream(id, chunk string, done bool, errMsg string) {
	if a.ctx == nil {
		return
	}
	// Use a typed inline struct via the trace channel.
	payload := map[string]any{
		"id":    id,
		"chunk": chunk,
		"done":  done,
		"error": errMsg,
	}
	wailsEmitStream(a.ctx, payload)
}

// (wailsEmitStream lives in app.go to avoid import cycles in tests.)

// ─── v1.2: Prompt templates ──────────────────────────────────────────────────

// GetPromptTemplates returns the current prompt templates map.
func (a *App) GetPromptTemplates() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]string, len(a.cfg.PromptTemplates))
	for k, v := range a.cfg.PromptTemplates {
		out[k] = v
	}
	return out
}

// SavePromptTemplates replaces the prompt templates map.
func (a *App) SavePromptTemplates(tpl map[string]string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.PromptTemplates = tpl
	return a.cfg.Save()
}

// effectivePrompt returns the user override (if any) for an op, else the built-in default.
func (a *App) effectivePrompt(op, defaultPrompt string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.PromptTemplates != nil {
		if v, ok := a.cfg.PromptTemplates[op]; ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	return defaultPrompt
}

// _ = unused to silence linter on RemoveSchedule path
var _ = fmt.Sprintf
