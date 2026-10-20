package app

import (
	"strings"

	"hatch/workbench/executor"
	"hatch/workbench/scheduler"
)

// ─── 工作台 · 定时任务 ───────────────────────────────────────────────────────

// ScheduleInfo describes a registered schedule job and its next fire time.
type ScheduleInfo struct {
	Path     string `json:"path"`
	Schedule string `json:"schedule"`
	NextRun  int64  `json:"nextRun"` // unix seconds; 0 if unknown
}

// ensureScheduler lazily creates the scheduler.
func (a *App) ensureScheduler() *scheduler.Scheduler {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sched == nil {
		a.sched = scheduler.New()
	}
	return a.sched
}

// ApplyAllSchedules registers/refreshes a job for every known script with a
// Schedule (override wins over scanned) and drops jobs no longer configured.
// Called at startup and after SaveScriptMeta.
func (a *App) ApplyAllSchedules() {
	sch := a.ensureScheduler()

	a.mu.RLock()
	overrideSchedules := make(map[string]string, len(a.cfg.OverrideMeta))
	for k, v := range a.cfg.OverrideMeta {
		overrideSchedules[k] = v.Schedule
	}
	a.mu.RUnlock()

	seen := map[string]bool{}
	if a.store != nil {
		for _, m := range a.store.All() {
			spec := m.Schedule
			if ov := overrideSchedules[m.Path]; ov != "" {
				spec = ov
			}
			if strings.TrimSpace(spec) == "" {
				continue
			}
			path := m.Path
			err := sch.Schedule(path, spec, func() {
				em := executor.NewEmitter(a.ctx)
				go func() {
					_, _ = executor.RunOneshotWithEnv(path, path, []string{}, a.scriptEnv(path), em, a.notifyHook(path))
				}()
			})
			if err == nil {
				seen[path] = true
			}
		}
	}

	for _, j := range sch.List() {
		if !seen[j.ID] {
			sch.Remove(j.ID)
		}
	}
}

// ListSchedules returns the registered schedule jobs and their next fire time.
func (a *App) ListSchedules() []ScheduleInfo {
	sch := a.ensureScheduler()
	jobs := sch.List()
	out := make([]ScheduleInfo, 0, len(jobs))
	for _, j := range jobs {
		info := ScheduleInfo{Path: j.ID, Schedule: j.Schedule}
		if nr, _ := sch.NextRun(j.ID); !nr.IsZero() {
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
