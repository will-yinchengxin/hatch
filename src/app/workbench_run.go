package app

import (
	"fmt"
	"strings"

	"hatch/core/macos"
	"hatch/workbench/executor"
)

// ─── 工作台 · 运行脚本与服务 ─────────────────────────────────────────────────

// RunScript launches a oneshot script in a goroutine. Progress and result
// arrive via executor:log / executor:done / executor:status events. When the
// script has PreScripts configured, they run sequentially first.
func (a *App) RunScript(id, path string, args []string) {
	if pre := a.scriptPreScripts(path); len(pre) > 0 {
		a.RunWithPreScripts(id, path, args, pre)
		return
	}
	em := executor.NewEmitter(a.ctx)
	go func() {
		_, _ = executor.RunOneshotWithEnv(id, path, args, a.scriptEnv(path), em, a.notifyHook(path))
	}()
}

// RunWithPreScripts executes pre-scripts in order, then the target. Each
// pre-script must exit 0; a "[hatch] pre: " log line marks every step.
func (a *App) RunWithPreScripts(id, path string, args []string, preScripts []string) {
	go func() {
		em := executor.NewEmitter(a.ctx)
		for _, pre := range preScripts {
			pre = strings.TrimSpace(pre)
			if pre == "" {
				continue
			}
			em.Log(id, "[hatch] pre: "+pre, "stdout")
			res, err := executor.RunOneshotWithEnv(pre, pre, nil, a.scriptEnv(pre), nil, nil)
			if err != nil || res == nil || res.ExitCode != 0 {
				code := -1
				if res != nil {
					code = res.ExitCode
				}
				em.Log(id, fmt.Sprintf("[hatch] pre failed: %s (exit %d)", pre, code), "stderr")
				em.Done(id, 99)
				em.Status(id, "error")
				return
			}
		}
		_, _ = executor.RunOneshotWithEnv(id, path, args, a.scriptEnv(path), em, a.notifyHook(path))
	}()
}

// ServiceAction executes start / stop / restart on a service script.
func (a *App) ServiceAction(path, action string, args []string) error {
	_, _, scanned, _ := a.scriptMetas(path)
	a.mu.Lock()
	sm, ok := a.services[path]
	if !ok {
		sm = executor.NewServiceManager(path, scanned.PidFile, executor.NewEmitter(a.ctx))
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

// ─── Run history ─────────────────────────────────────────────────────────────

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

// ─── Per-script run settings (override > scanned) ────────────────────────────

// scriptPreScripts returns the configured pre-script chain for a path.
func (a *App) scriptPreScripts(path string) []string {
	ov, hasOv, scanned, _ := a.scriptMetas(path)
	if hasOv && len(ov.PreScripts) > 0 {
		return append([]string(nil), ov.PreScripts...)
	}
	if len(scanned.PreScripts) > 0 {
		return append([]string(nil), scanned.PreScripts...)
	}
	return nil
}

// scriptEnv returns the per-script env map.
func (a *App) scriptEnv(path string) map[string]string {
	ov, hasOv, scanned, _ := a.scriptMetas(path)
	src := scanned.Env
	if hasOv && len(ov.Env) > 0 {
		src = ov.Env
	}
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// notifyHook returns a callback that sends a macOS notification after a
// oneshot finishes, or nil when the script meta has NotifyOnDone=false.
func (a *App) notifyHook(path string) func(exitCode int, durationMs int64) {
	ov, hasOv, scanned, hasScanned := a.scriptMetas(path)
	notify, name := false, ""
	if hasOv {
		notify, name = ov.NotifyOnDone, ov.Name
	}
	if !notify && hasScanned {
		notify = scanned.NotifyOnDone
		if name == "" {
			name = scanned.Name
		}
	}
	if !notify {
		return nil
	}
	if name == "" {
		name = path
	}
	return func(exitCode int, durationMs int64) {
		msg := fmt.Sprintf("✓ 完成 (%dms)", durationMs)
		if exitCode != 0 {
			msg = fmt.Sprintf("✗ 退出码 %d (%dms)", exitCode, durationMs)
		}
		macos.Notify("Hatch: "+name, msg)
	}
}
