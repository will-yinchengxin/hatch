package app

import (
	"fmt"
	"path/filepath"

	"hatch/core/config"
	"hatch/core/fsutil"
	"hatch/workbench/zshrc"
)

// ─── 工作台 · 目录与环境变量 ─────────────────────────────────────────────────

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
		a.upsertExportLocked(exportName, path)
		if err := a.syncZshrcLocked(); err != nil {
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
	// Drop per-script settings of files that lived directly in this directory.
	for p := range a.cfg.OverrideMeta {
		if filepath.Dir(p) == filepath.Clean(path) {
			delete(a.cfg.OverrideMeta, p)
		}
	}
	delete(a.cfg.ScriptOrders, path)

	if removedExport != "" {
		a.removeExportLocked(removedExport)
		if err := a.syncZshrcLocked(); err != nil {
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
			a.removeExportLocked(d.ExportName)
		}
		a.cfg.Directories[i].ExportName = exportName
		break
	}
	if !found {
		return fmt.Errorf("directory not found: %s", dirPath)
	}

	if exportName != "" {
		a.upsertExportLocked(exportName, dirPath)
		if err := a.syncZshrcLocked(); err != nil {
			return err
		}
	}
	return a.cfg.Save()
}

// SaveExports persists the exports list to config and rewrites the zshrc block.
func (a *App) SaveExports(exports []config.ExportVar) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.Exports = exports
	if err := a.syncZshrcLocked(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// upsertExportLocked records a directory export; directories always join PATH.
func (a *App) upsertExportLocked(name, value string) {
	for i, e := range a.cfg.Exports {
		if e.Name == name {
			a.cfg.Exports[i].Value = value
			a.cfg.Exports[i].Kind = "path"
			return
		}
	}
	a.cfg.Exports = append(a.cfg.Exports, config.ExportVar{Name: name, Value: value, Kind: "path"})
}

func (a *App) removeExportLocked(name string) {
	var kept []config.ExportVar
	for _, e := range a.cfg.Exports {
		if e.Name != name {
			kept = append(kept, e)
		}
	}
	a.cfg.Exports = kept
}

// syncZshrcLocked rebuilds the managed ~/.zshrc block from cfg.Exports.
// Caller must hold a.mu.
func (a *App) syncZshrcLocked() error {
	p, err := zshrc.ZshrcPath()
	if err != nil {
		return err
	}
	vars := make([]zshrc.Var, 0, len(a.cfg.Exports))
	seen := make(map[string]bool, len(a.cfg.Exports))
	for _, e := range a.cfg.Exports {
		if e.Name == "" || seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		// "~/x" would be single-quoted in the block, so expand it here.
		vars = append(vars, zshrc.Var{Name: e.Name, Value: fsutil.ExpandHome(e.Value), InPath: e.InPath()})
	}
	if len(vars) == 0 {
		return zshrc.RemoveBlock(p)
	}
	return zshrc.WriteVars(p, vars)
}
