package main

import (
	"fmt"

	"hatch/internal/executor"
)

// ─── Script execution ────────────────────────────────────────────────────────

// RunScript launches a oneshot script in a goroutine.
// Progress and result arrive via executor:log / executor:done / executor:status events.
// v1.3: if the script has PreScripts configured, they run sequentially first.
func (a *App) RunScript(id, path string, args []string) {
	pre := a.scriptPreScripts(path)
	if len(pre) > 0 {
		a.RunWithPreScripts(id, path, args, pre)
		return
	}
	em := executor.NewEmitter(a.ctx)
	go func() {
		_, _ = executor.RunOneshotWithEnv(id, path, args, a.scriptEnv(path), em, a.notifyHook(path))
	}()
}

// scriptPreScripts returns the configured pre-script chain for a path.
func (a *App) scriptPreScripts(path string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.OverrideMeta != nil {
		if m, ok := a.cfg.OverrideMeta[path]; ok && len(m.PreScripts) > 0 {
			out := make([]string, len(m.PreScripts))
			copy(out, m.PreScripts)
			return out
		}
	}
	if a.store != nil {
		for _, s := range a.store.All() {
			if s.Path == path && len(s.PreScripts) > 0 {
				out := make([]string, len(s.PreScripts))
				copy(out, s.PreScripts)
				return out
			}
		}
	}
	return nil
}

// ServiceAction executes start / stop / restart on a service script.
func (a *App) ServiceAction(path, action string, args []string) error {
	a.mu.Lock()
	sm, ok := a.services[path]
	if !ok {
		pidFile := ""
		if a.store != nil {
			for _, s := range a.store.All() {
				if s.Path == path {
					pidFile = s.PidFile
					break
				}
			}
		}
		sm = executor.NewServiceManager(path, pidFile, executor.NewEmitter(a.ctx))
		a.services[path] = sm
	}
	a.mu.Unlock()

	env := a.scriptEnv(path)

	switch action {
	case "start":
		return sm.StartWithEnv(args, env)
	case "stop":
		return sm.Stop()
	case "restart":
		return sm.RestartWithEnv(args, env)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// ─── v1.1: Run history ───────────────────────────────────────────────────────

// ListRuns returns archived run records for a script (newest first).
func (a *App) ListRuns(path string) ([]executor.RunRecord, error) {
	return executor.ListRuns(path)
}

// GetRunLog returns the full log file contents for one archived run.
func (a *App) GetRunLog(path, runID string) (string, error) {
	return executor.GetRunLog(path, runID)
}

// DeleteRun removes one archived run.
func (a *App) DeleteRun(path, runID string) error {
	return executor.DeleteRun(path, runID)
}

// GetAllServiceStatuses returns path → "running"|"stopped" for all tracked services.
func (a *App) GetAllServiceStatuses() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]string, len(a.services))
	for path, sm := range a.services {
		status, _ := sm.Status()
		out[path] = string(status)
	}
	return out
}

// scriptEnv returns the per-script env map (from Override or scanned meta).
func (a *App) scriptEnv(path string) map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.cfg.OverrideMeta != nil {
		if m, ok := a.cfg.OverrideMeta[path]; ok && len(m.Env) > 0 {
			out := make(map[string]string, len(m.Env))
			for k, v := range m.Env {
				out[k] = v
			}
			return out
		}
	}
	if a.store != nil {
		for _, s := range a.store.All() {
			if s.Path == path && len(s.Env) > 0 {
				out := make(map[string]string, len(s.Env))
				for k, v := range s.Env {
					out[k] = v
				}
				return out
			}
		}
	}
	return nil
}

// notifyHook returns a callback that fires after a oneshot finishes — used for
// macOS native notifications when the script meta has NotifyOnDone=true.
func (a *App) notifyHook(path string) func(exitCode int, durationMs int64) {
	a.mu.RLock()
	notify := false
	name := ""
	if a.cfg.OverrideMeta != nil {
		if m, ok := a.cfg.OverrideMeta[path]; ok {
			notify = m.NotifyOnDone
			name = m.Name
		}
	}
	if !notify && a.store != nil {
		for _, s := range a.store.All() {
			if s.Path == path {
				notify = s.NotifyOnDone
				if name == "" {
					name = s.Name
				}
				break
			}
		}
	}
	a.mu.RUnlock()
	if !notify {
		return nil
	}
	if name == "" {
		name = path
	}
	return func(exitCode int, durationMs int64) {
		title := "Hatch: " + name
		var msg string
		if exitCode == 0 {
			msg = fmt.Sprintf("✓ 完成 (%dms)", durationMs)
		} else {
			msg = fmt.Sprintf("✗ 退出码 %d (%dms)", exitCode, durationMs)
		}
		sendMacNotification(title, msg)
	}
}
