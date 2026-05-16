package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"hatch/internal/autostart"
	"hatch/internal/config"
	"hatch/internal/zshrc"
)

// AppDataPaths holds all configurable directory paths.
type AppDataPaths struct {
	TraceLogDir      string `json:"traceLogDir"`
	TraceLogDirIsSet bool   `json:"traceLogDirIsSet"`
	SkillsDir        string `json:"skillsDir"`
	SkillsDirIsSet   bool   `json:"skillsDirIsSet"`
	DefaultTraceDir  string `json:"defaultTraceDir"`
	DefaultSkillsDir string `json:"defaultSkillsDir"`
}

// GetAppDataPaths returns current path configuration with effective values.
func (a *App) GetAppDataPaths() AppDataPaths {
	a.mu.RLock()
	traceCfg := a.cfg.TraceLogDir
	skillsCfg := a.cfg.SkillsDir
	a.mu.RUnlock()

	defTrace := defaultTraceDir()
	defSkills := defaultSkillsDir()

	return AppDataPaths{
		TraceLogDir:      orDefault(traceCfg, defTrace),
		TraceLogDirIsSet: traceCfg != "",
		SkillsDir:        orDefault(skillsCfg, defSkills),
		SkillsDirIsSet:   skillsCfg != "",
		DefaultTraceDir:  defTrace,
		DefaultSkillsDir: defSkills,
	}
}

func orDefault(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

// SaveAppDataPaths saves custom directory overrides (empty string = reset to default).
func (a *App) SaveAppDataPaths(traceLogDir, skillsDir string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.TraceLogDir = strings.TrimSpace(traceLogDir)
	a.cfg.SkillsDir = strings.TrimSpace(skillsDir)
	return a.cfg.Save()
}

func (a *App) GetAutoLaunchEnabled() bool {
	return autostart.Enabled()
}

func (a *App) SetAutoLaunchEnabled(enabled bool) error {
	if err := autostart.SetEnabled(enabled); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.AutoLaunch = enabled
	err := a.cfg.Save()
	a.mu.Unlock()
	return err
}

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

// ─── Config file management ───────────────────────────────────────────────────

func (a *App) GetClaudeSettings() string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.claude/settings.json")
	if err != nil {
		return ""
	}
	return string(data)
}

func (a *App) SaveClaudeSettings(content string) error {
	if !json.Valid([]byte(content)) {
		return fmt.Errorf("invalid JSON")
	}
	home, _ := os.UserHomeDir()
	path := home + "/.claude/settings.json"
	if err := backupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func (a *App) GetCodexConfig() string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.codex/config.toml")
	if err != nil {
		return ""
	}
	return string(data)
}

func (a *App) SaveCodexConfig(content string) error {
	home, _ := os.UserHomeDir()
	path := home + "/.codex/config.toml"
	if err := backupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func backupIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_, err := zshrc.Backup(path)
	return err
}

// ── Doc links ──────────────────────────────────────────────────────────────

func (a *App) GetDocLinks() []config.DocLink {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]config.DocLink, len(a.cfg.DocLinks))
	copy(out, a.cfg.DocLinks)
	return out
}

func (a *App) SaveDocLinks(links []config.DocLink) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.DocLinks = links
	return a.cfg.Save()
}

// OpenFile opens a file with the system default application (macOS `open`).
func (a *App) OpenFile(path string) error {
	return exec.Command("open", path).Start()
}
