package app

import (
	"fmt"
	"os"
	"path/filepath"

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
