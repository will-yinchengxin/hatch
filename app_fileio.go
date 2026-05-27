package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ─── Script file I/O ─────────────────────────────────────────────────────────

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
	a.autoSource()
	return nil
}

// CreateScriptFile creates a new script file inside dirPath.
func (a *App) CreateScriptFile(dirPath, filename, content string) (string, error) {
	fullPath := filepath.Join(dirPath, filename)
	if _, err := os.Stat(fullPath); err == nil {
		return "", fmt.Errorf("文件已存在: %s", fullPath)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
		return "", err
	}
	a.autoSource()
	return fullPath, nil
}

// autoSource runs `source ~/.zshrc` in a background login shell.
func (a *App) autoSource() {
	go func() {
		_ = exec.Command("/bin/zsh", "-l", "-c", "source ~/.zshrc 2>/dev/null || true").Run()
	}()
}

// OpenFile opens a file with the system default application (macOS `open`).
func (a *App) OpenFile(path string) error {
	return exec.Command("open", path).Start()
}
