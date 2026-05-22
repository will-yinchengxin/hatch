package main

import (
	"fmt"
	"sort"

	"hatch/internal/config"
	"hatch/internal/executor"
	"hatch/internal/scanner"
)

// ─── Config & Directories ────────────────────────────────────────────────────

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

// AddDirectory adds a directory to the config and optionally writes an export to ~/.zshrc.
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

// applyScriptOrder reorders scripts according to a saved path list; new scripts append at the end.
func applyScriptOrder(scripts []config.ScriptMeta, order []string) []config.ScriptMeta {
	rank := make(map[string]int, len(order))
	for i, p := range order {
		rank[p] = i
	}
	result := make([]config.ScriptMeta, len(scripts))
	copy(result, scripts)
	sort.SliceStable(result, func(i, j int) bool {
		ri, iok := rank[result[i].Path]
		rj, jok := rank[result[j].Path]
		if iok && jok {
			return ri < rj
		}
		return iok
	})
	return result
}

// GetScriptsByDir scans a single directory and returns its scripts.
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
		scripts = applyScriptOrder(scripts, order)
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

// SaveScriptMeta stores user-edited metadata into config.OverrideMeta.
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
	// v1.2: re-apply schedules whenever meta is saved (Schedule may have changed)
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
