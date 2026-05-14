package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scanner"
	"hatch/internal/store"
	"hatch/internal/zshrc"
)

type App struct {
	ctx      context.Context
	cfg      *config.Config
	store    *store.Store
	services map[string]*executor.ServiceManager
	mu       sync.RWMutex
}

// ZshrcInfo is returned by CheckZshrc.
type ZshrcInfo struct {
	Shell      string   `json:"shell"`
	IsZsh      bool     `json:"isZsh"`
	ZshrcPath  string   `json:"zshrcPath"`
	HatchLines []string `json:"hatchLines"`
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
}

// restoreServices rebuilds ServiceManager instances for service-type scripts
// found in the store, and probes stale PID files on app start.
func (a *App) restoreServices() {
	for _, s := range a.store.All() {
		if s.Type == config.TypeService {
			sm := executor.NewServiceManager(s.Path, s.PidFile, executor.NewEmitter(a.ctx))
			a.mu.Lock()
			a.services[s.Path] = sm
			a.mu.Unlock()
			// Probe liveness — cleans up stale pidfiles automatically.
			sm.Status()
		}
	}
}

// ─── Config & Directories ────────────────────────────────────────────────────

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

// AddDirectory adds a directory to the config and optionally writes an export
// to ~/.zshrc.
func (a *App) AddDirectory(path, exportName string, isScriptDir bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, d := range a.cfg.Directories {
		if d.Path == path {
			return fmt.Errorf("directory already added: %s", path)
		}
	}
	a.cfg.Directories = append(a.cfg.Directories, config.Directory{
		Path:        path,
		IsScriptDir: isScriptDir,
		ExportName:  exportName,
	})

	if exportName != "" {
		found := false
		for i, e := range a.cfg.Exports {
			if e.Name == exportName {
				a.cfg.Exports[i].Value = path
				found = true
				break
			}
		}
		if !found {
			a.cfg.Exports = append(a.cfg.Exports, config.ExportVar{Name: exportName, Value: path})
		}
		if err := a.syncZshrc(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// RemoveDirectory removes a directory (and its export) from config.
func (a *App) RemoveDirectory(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var kept []config.Directory
	removedExport := ""
	for _, d := range a.cfg.Directories {
		if d.Path == path {
			removedExport = d.ExportName
		} else {
			kept = append(kept, d)
		}
	}
	a.cfg.Directories = kept

	if removedExport != "" {
		var exports []config.ExportVar
		for _, e := range a.cfg.Exports {
			if e.Name != removedExport {
				exports = append(exports, e)
			}
		}
		a.cfg.Exports = exports
		if err := a.syncZshrc(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// UpdateDirectoryExport sets (or clears) the export name for an existing directory
// and syncs ~/.zshrc immediately.
func (a *App) UpdateDirectoryExport(dirPath, exportName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	found := false
	for i, d := range a.cfg.Directories {
		if d.Path != dirPath {
			continue
		}
		found = true
		// Remove old export entry if name changed
		if d.ExportName != "" && d.ExportName != exportName {
			var kept []config.ExportVar
			for _, e := range a.cfg.Exports {
				if e.Name != d.ExportName {
					kept = append(kept, e)
				}
			}
			a.cfg.Exports = kept
		}
		a.cfg.Directories[i].ExportName = exportName
		break
	}
	if !found {
		return fmt.Errorf("directory not found: %s", dirPath)
	}

	if exportName != "" {
		updated := false
		for i, e := range a.cfg.Exports {
			if e.Name == exportName {
				a.cfg.Exports[i].Value = dirPath
				updated = true
				break
			}
		}
		if !updated {
			a.cfg.Exports = append(a.cfg.Exports, config.ExportVar{Name: exportName, Value: dirPath})
		}
		if err := a.syncZshrc(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// ─── Script scanning ─────────────────────────────────────────────────────────

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

	// Ensure a ServiceManager exists for every service-type script.
	a.mu.Lock()
	for _, s := range all {
		if s.Type == config.TypeService {
			if _, ok := a.services[s.Path]; !ok {
				a.services[s.Path] = executor.NewServiceManager(
					s.Path, s.PidFile, executor.NewEmitter(a.ctx),
				)
			}
		}
	}
	a.mu.Unlock()

	if a.store != nil {
		_ = a.store.Replace(all)
	}
	return all, nil
}

// GetScriptsByDir scans a single directory and returns its scripts.
func (a *App) GetScriptsByDir(dirPath string) ([]config.ScriptMeta, error) {
	a.mu.RLock()
	overrides := a.cfg.OverrideMeta
	a.mu.RUnlock()

	scripts, err := scanner.ScanDir(dirPath, overrides)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	for _, s := range scripts {
		if s.Type == config.TypeService {
			if _, ok := a.services[s.Path]; !ok {
				a.services[s.Path] = executor.NewServiceManager(
					s.Path, s.PidFile, executor.NewEmitter(a.ctx),
				)
			}
		}
	}
	a.mu.Unlock()

	if a.store != nil {
		for _, s := range scripts {
			_ = a.store.UpdateMeta(s)
		}
	}
	return scripts, nil
}

// ─── Script execution ────────────────────────────────────────────────────────

// RunScript launches a oneshot script in a goroutine.
// Progress and result arrive via executor:log / executor:done / executor:status events.
func (a *App) RunScript(id, path string, args []string) {
	em := executor.NewEmitter(a.ctx)
	go func() {
		_, _ = executor.RunOneshot(id, path, args, em)
	}()
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

	switch action {
	case "start":
		return sm.Start(args)
	case "stop":
		return sm.Stop()
	case "restart":
		return sm.Restart(args)
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

// ─── Zshrc & exports ─────────────────────────────────────────────────────────

// SaveExports persists the exports list to config and rewrites the zshrc block.
func (a *App) SaveExports(exports []config.ExportVar) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Exports = exports
	if err := a.syncZshrc(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// CheckZshrc returns current shell info and the Hatch managed block contents.
func (a *App) CheckZshrc() (ZshrcInfo, error) {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return ZshrcInfo{}, err
	}
	shell := zshrc.CurrentShell()
	lines, _ := zshrc.ReadBlock(p)
	return ZshrcInfo{
		Shell:      shell,
		IsZsh:      strings.HasSuffix(shell, "zsh"),
		ZshrcPath:  p,
		HatchLines: lines,
	}, nil
}

// syncZshrc rebuilds the managed block from cfg.Exports. Caller must hold mu.
func (a *App) syncZshrc() error {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return err
	}
	exports := make(map[string]string, len(a.cfg.Exports))
	for _, e := range a.cfg.Exports {
		if e.Name != "" {
			exports[e.Name] = e.Value
		}
	}
	if len(exports) == 0 {
		return zshrc.RemoveBlock(p)
	}
	return zshrc.WriteBlock(p, exports)
}
