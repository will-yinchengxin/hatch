package app

import (
	"hatch/core/config"
	"hatch/workbench/executor"
	"hatch/workbench/scanner"
)

// ─── 工作台 · 脚本扫描与元信息 ───────────────────────────────────────────────

// ScanAll rescans every configured script directory and returns the full list.
func (a *App) ScanAll() ([]config.ScriptMeta, error) {
	a.mu.RLock()
	dirs := make([]config.Directory, len(a.cfg.Directories))
	copy(dirs, a.cfg.Directories)
	overrides := a.cfg.OverrideMeta
	a.mu.RUnlock()

	var all []config.ScriptMeta
	for _, d := range dirs {
		if !d.IsScriptDir {
			continue
		}
		scripts, err := scanner.ScanDir(d.Path, overrides)
		if err != nil {
			continue
		}
		all = append(all, scripts...)
	}
	a.trackServices(all)

	if a.store != nil {
		_ = a.store.Replace(all)
	}
	return all, nil
}

// GetScriptsByDir scans a single directory and returns its scripts in the
// user's saved order.
func (a *App) GetScriptsByDir(dirPath string) ([]config.ScriptMeta, error) {
	a.mu.RLock()
	overrides := a.cfg.OverrideMeta
	order := a.cfg.ScriptOrders[dirPath]
	a.mu.RUnlock()

	scripts, err := scanner.ScanDir(dirPath, overrides)
	if err != nil {
		return nil, err
	}
	if len(order) > 0 {
		scripts = scanner.ApplyOrder(scripts, order)
	}
	a.trackServices(scripts)

	if a.store != nil {
		for _, s := range scripts {
			_ = a.store.UpdateMeta(s)
		}
	}
	return scripts, nil
}

// SaveScriptMeta stores user-edited metadata into config.OverrideMeta and
// re-applies schedules (Schedule may have changed).
func (a *App) SaveScriptMeta(meta config.ScriptMeta) error {
	a.mu.Lock()
	if a.cfg.OverrideMeta == nil {
		a.cfg.OverrideMeta = make(map[string]config.ScriptMeta)
	}
	a.cfg.OverrideMeta[meta.Path] = meta
	err := a.cfg.Save()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	go a.ApplyAllSchedules()
	return nil
}

// DeleteScriptMeta removes a user override so auto-detection resumes.
func (a *App) DeleteScriptMeta(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.cfg.OverrideMeta, path)
	return a.cfg.Save()
}

// SaveScriptOrder persists the display order for scripts in a directory.
func (a *App) SaveScriptOrder(dirPath string, paths []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cfg.ScriptOrders == nil {
		a.cfg.ScriptOrders = make(map[string][]string)
	}
	a.cfg.ScriptOrders[dirPath] = paths
	return a.cfg.Save()
}

// trackServices creates a ServiceManager for every newly seen service script.
func (a *App) trackServices(scripts []config.ScriptMeta) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, s := range scripts {
		if s.Type != config.TypeService {
			continue
		}
		if _, ok := a.services[s.Path]; !ok {
			a.services[s.Path] = executor.NewServiceManager(s.Path, s.PidFile, executor.NewEmitter(a.ctx))
		}
	}
}

// scriptMetas returns the user override (if any) and the last scanned meta
// (if any) for path. Per-field precedence is decided by each caller.
func (a *App) scriptMetas(path string) (override config.ScriptMeta, hasOverride bool, scanned config.ScriptMeta, hasScanned bool) {
	a.mu.RLock()
	override, hasOverride = a.cfg.OverrideMeta[path]
	a.mu.RUnlock()
	if a.store != nil {
		for _, s := range a.store.All() {
			if s.Path == path {
				return override, hasOverride, s, true
			}
		}
	}
	return override, hasOverride, config.ScriptMeta{}, false
}
