package app

import (
	"fmt"
	"os"
	"path/filepath"

	"hatch/core/fsutil"
	"hatch/core/macos"
	"hatch/workbench/diagnostics"
)

// ─── 工作台 · 脚本文件与诊断 ─────────────────────────────────────────────────

// ReadScriptFile returns the raw text content of a script file.
func (a *App) ReadScriptFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteScriptFile writes new content to an existing script file.
func (a *App) WriteScriptFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	macos.SourceZshrcAsync()
	return nil
}

// CreateScriptFile creates a new executable script file inside dirPath.
func (a *App) CreateScriptFile(dirPath, filename, content string) (string, error) {
	if filename == "" || filename != filepath.Base(filename) || filename[0] == '.' {
		return "", fmt.Errorf("文件名不合法: %q", filename)
	}
	fullPath := filepath.Join(dirPath, filename)
	if _, err := os.Stat(fullPath); err == nil {
		return "", fmt.Errorf("文件已存在: %s", fullPath)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
		return "", err
	}
	macos.SourceZshrcAsync()
	return fullPath, nil
}

// TrashScriptFile moves a script inside a configured script directory to the
// macOS Trash and drops its Hatch metadata. Returns the path in the Trash.
func (a *App) TrashScriptFile(path string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	inside := false
	for _, d := range a.cfg.Directories {
		if d.IsScriptDir && fsutil.IsPathInside(path, d.Path) {
			inside = true
			break
		}
	}
	if !inside {
		return "", fmt.Errorf("只能删除脚本目录中的文件: %s", path)
	}
	dst, err := macos.MoveToTrash(path)
	if err != nil {
		return "", err
	}
	delete(a.cfg.OverrideMeta, path)
	for dir, order := range a.cfg.ScriptOrders {
		kept := order[:0]
		for _, p := range order {
			if p != path {
				kept = append(kept, p)
			}
		}
		a.cfg.ScriptOrders[dir] = kept
	}
	return dst, a.cfg.Save()
}

// RevealFile selects path in Finder.
func (a *App) RevealFile(path string) error {
	return macos.Reveal(path)
}

// OpenFile opens a file with the system default application (macOS `open`).
func (a *App) OpenFile(path string) error {
	return macos.Open(path)
}

// CheckScriptHealth verifies @hatch.requires deps and @hatch.health URL for one script.
func (a *App) CheckScriptHealth(path string) diagnostics.HealthCheckResult {
	meta, hasOv, scanned, hasScanned := a.scriptMetas(path)
	if (!hasOv || meta.Path == "") && hasScanned {
		meta = scanned
	}
	return diagnostics.CheckScript(path, meta)
}

// CheckCommandAvailable checks if a command exists on the login-shell PATH.
func (a *App) CheckCommandAvailable(name string) bool {
	return macos.FindCLI(name) != ""
}

// CheckPathHealth inspects $PATH from a login zsh and reports missing/duplicate dirs.
func (a *App) CheckPathHealth() diagnostics.PathHealth {
	return diagnostics.CheckPath()
}
