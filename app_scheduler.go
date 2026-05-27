package main

import (
	"strings"

	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scheduler"
)

// ─── Scheduler ───────────────────────────────────────────────────────────────

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

// ScheduleInfo describes a registered schedule job and its next fire time.
type ScheduleInfo struct {
	Path     string `json:"path"`
	Schedule string `json:"schedule"`
	NextRun  int64  `json:"nextRun"` // unix seconds; 0 if unknown
}

// ListSchedules returns the registered schedule jobs and their next fire time.
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
