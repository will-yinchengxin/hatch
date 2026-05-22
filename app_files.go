package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// ── Claude Code model + effort config ─────────────────────────────────────────

// ClaudeCodeConfig holds the two user-facing knobs written to ~/.claude/settings.json,
// plus EffectiveModel which reflects what Claude Code is actually running (via `claude model`).
type ClaudeCodeConfig struct {
	Model         string `json:"model"`         // explicit setting in settings.json, "" means default
	EffortLevel   string `json:"effortLevel"`   // explicit setting in settings.json
	EffectiveModel string `json:"effectiveModel"` // actual model in use (from `claude model`)
}

// GetClaudeCodeConfig reads model and effortLevel from ~/.claude/settings.json,
// and queries `claude model` to get the actual effective model.
func (a *App) GetClaudeCodeConfig() ClaudeCodeConfig {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.claude/settings.json")
	cfg := ClaudeCodeConfig{}
	if err == nil {
		var raw map[string]interface{}
		if json.Unmarshal(data, &raw) == nil {
			if m, ok := raw["model"].(string); ok {
				cfg.Model = m
			}
			if e, ok := raw["effortLevel"].(string); ok {
				cfg.EffortLevel = e
			}
		}
	}

	// Run `claude model` to get the effective model (works even when no model is set in settings).
	if claudePath := findCLI("claude"); claudePath != "" {
		if out, err2 := exec.Command(claudePath, "model").Output(); err2 == nil {
			// Output: "You're currently running on **Claude Sonnet 4.6** (`claude-sonnet-4-6`)."
			if idx := strings.Index(string(out), "`"); idx >= 0 {
				rest := string(out)[idx+1:]
				if end := strings.Index(rest, "`"); end > 0 {
					cfg.EffectiveModel = rest[:end]
				}
			}
		}
	}

	return cfg
}

// SetClaudeCodeConfig merges model and effortLevel into ~/.claude/settings.json
// without touching other fields. Empty string removes the key.
func (a *App) SetClaudeCodeConfig(model, effortLevel string) error {
	home, _ := os.UserHomeDir()
	path := home + "/.claude/settings.json"

	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &raw) //nolint:errcheck
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}

	if model != "" {
		raw["model"] = model
	} else {
		delete(raw, "model")
	}
	if effortLevel != "" {
		raw["effortLevel"] = effortLevel
	} else {
		delete(raw, "effortLevel")
	}

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := backupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

// ── Claude Code presets ────────────────────────────────────────────────────────

func (a *App) GetClaudeCodePresets() []config.ClaudeCodePreset {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.cfg.ClaudeCodePresetsSeeded && len(a.cfg.ClaudeCodePresets) == 0 {
		a.cfg.ClaudeCodePresets = defaultClaudeCodePresets()
		a.cfg.ClaudeCodePresetsSeeded = true
		_ = a.cfg.Save()
	}
	out := make([]config.ClaudeCodePreset, len(a.cfg.ClaudeCodePresets))
	copy(out, a.cfg.ClaudeCodePresets)
	return out
}

// SaveClaudeCodePreset creates or updates a preset (matched by ID).
func (a *App) SaveClaudeCodePreset(preset config.ClaudeCodePreset) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.ClaudeCodePresetsSeeded = true
	if preset.ID == "" {
		preset.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	for i, p := range a.cfg.ClaudeCodePresets {
		if p.ID == preset.ID {
			a.cfg.ClaudeCodePresets[i] = preset
			return a.cfg.Save()
		}
	}
	a.cfg.ClaudeCodePresets = append(a.cfg.ClaudeCodePresets, preset)
	return a.cfg.Save()
}

// DeleteClaudeCodePreset removes a preset by ID.
func (a *App) DeleteClaudeCodePreset(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.ClaudeCodePresetsSeeded = true
	keep := a.cfg.ClaudeCodePresets[:0]
	for _, p := range a.cfg.ClaudeCodePresets {
		if p.ID != id {
			keep = append(keep, p)
		}
	}
	a.cfg.ClaudeCodePresets = keep
	return a.cfg.Save()
}

func defaultClaudeCodePresets() []config.ClaudeCodePreset {
	return []config.ClaudeCodePreset{
		{ID: "cc-fast", Name: "快速响应", Model: "claude-haiku-4-5", EffortLevel: "low"},
		{ID: "cc-balanced", Name: "均衡编码", Model: "claude-sonnet-4-6", EffortLevel: "medium"},
		{ID: "cc-deep", Name: "复杂分析", Model: "claude-sonnet-4-6", EffortLevel: "high"},
		{ID: "cc-frontier", Name: "深度推理", Model: "claude-opus-4-7", EffortLevel: "xhigh"},
	}
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

// ── Codex model + reasoning effort config ───────────────────────────────────

// CodexModelConfig holds the two user-facing knobs written to ~/.codex/config.toml.
type CodexModelConfig struct {
	Model       string `json:"model"`
	EffortLevel string `json:"effortLevel"`
}

// GetCodexModelConfig reads model and model_reasoning_effort from ~/.codex/config.toml.
func (a *App) GetCodexModelConfig() CodexModelConfig {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.codex/config.toml")
	if err != nil {
		return CodexModelConfig{}
	}
	content := string(data)
	return CodexModelConfig{
		Model:       readTOMLStringKey(content, "model"),
		EffortLevel: readTOMLStringKey(content, "model_reasoning_effort"),
	}
}

// SetCodexModelConfig merges model and model_reasoning_effort into
// ~/.codex/config.toml without touching unrelated settings.
func (a *App) SetCodexModelConfig(model, effortLevel string) error {
	home, _ := os.UserHomeDir()
	path := home + "/.codex/config.toml"

	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	content = setTOMLStringKey(content, "model_reasoning_effort", effortLevel)
	content = setTOMLStringKey(content, "model", model)

	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func readTOMLStringKey(content, key string) string {
	preamble, _ := splitTOMLPreamble(content)
	for _, line := range strings.Split(preamble, "\n") {
		if !isTOMLKeyLine(line, key) {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return ""
		}
		quoted := quotedTOMLStringPrefix(strings.TrimSpace(parts[1]))
		if quoted == "" {
			return ""
		}
		value, err := strconv.Unquote(quoted)
		if err != nil {
			return ""
		}
		return value
	}
	return ""
}

func setTOMLStringKey(content, key, value string) string {
	preamble, rest := splitTOMLPreamble(content)
	var kept []string
	for _, line := range strings.SplitAfter(preamble, "\n") {
		if line == "" || isTOMLKeyLine(line, key) {
			continue
		}
		kept = append(kept, line)
	}
	content = strings.TrimLeft(strings.Join(kept, "")+rest, "\n")
	if value == "" {
		return ensureTrailingNewline(content)
	}
	line := key + " = " + strconv.Quote(value) + "\n"
	if content == "" {
		return line
	}
	return line + "\n" + ensureTrailingNewline(content)
}

func splitTOMLPreamble(content string) (string, string) {
	offset := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			return content[:offset], content[offset:]
		}
		offset += len(line)
	}
	return content, ""
}

func isTOMLKeyLine(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, key) {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, key))
	return strings.HasPrefix(rest, "=")
}

func quotedTOMLStringPrefix(s string) string {
	if !strings.HasPrefix(s, `"`) {
		return ""
	}
	escaped := false
	for i := 1; i < len(s); i++ {
		switch {
		case escaped:
			escaped = false
		case s[i] == '\\':
			escaped = true
		case s[i] == '"':
			return s[:i+1]
		}
	}
	return ""
}

func ensureTrailingNewline(content string) string {
	if content == "" || strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}

// ── Codex presets ───────────────────────────────────────────────────────────

func (a *App) GetCodexPresets() []config.CodexPreset {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.cfg.CodexPresetsSeeded && len(a.cfg.CodexPresets) == 0 {
		a.cfg.CodexPresets = defaultCodexPresets()
		a.cfg.CodexPresetsSeeded = true
		_ = a.cfg.Save()
	}
	out := make([]config.CodexPreset, len(a.cfg.CodexPresets))
	copy(out, a.cfg.CodexPresets)
	return out
}

// SaveCodexPreset creates or updates a preset (matched by ID).
func (a *App) SaveCodexPreset(preset config.CodexPreset) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.CodexPresetsSeeded = true
	if preset.ID == "" {
		preset.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	for i, p := range a.cfg.CodexPresets {
		if p.ID == preset.ID {
			a.cfg.CodexPresets[i] = preset
			return a.cfg.Save()
		}
	}
	a.cfg.CodexPresets = append(a.cfg.CodexPresets, preset)
	return a.cfg.Save()
}

// DeleteCodexPreset removes a preset by ID.
func (a *App) DeleteCodexPreset(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg.CodexPresetsSeeded = true
	keep := a.cfg.CodexPresets[:0]
	for _, p := range a.cfg.CodexPresets {
		if p.ID != id {
			keep = append(keep, p)
		}
	}
	a.cfg.CodexPresets = keep
	return a.cfg.Save()
}

func defaultCodexPresets() []config.CodexPreset {
	return []config.CodexPreset{
		{ID: "default-fast", Name: "快速日常", Model: "gpt-5.4-mini", EffortLevel: "low"},
		{ID: "default-balanced", Name: "均衡编码", Model: "gpt-5.5", EffortLevel: "medium"},
		{ID: "default-deep", Name: "复杂任务", Model: "gpt-5.5", EffortLevel: "high"},
		{ID: "default-frontier", Name: "最强推理", Model: "gpt-5.5", EffortLevel: "xhigh"},
	}
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
