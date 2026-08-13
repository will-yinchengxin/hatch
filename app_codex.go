package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"hatch/internal/config"

	"github.com/pelletier/go-toml"
)

// ─── Codex settings ──────────────────────────────────────────────────────────

// GetCodexConfig returns the raw TOML content of ~/.codex/config.toml.
func (a *App) GetCodexConfig() string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.codex/config.toml")
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveCodexConfig validates and writes ~/.codex/config.toml.
func (a *App) SaveCodexConfig(content string) error {
	if _, err := toml.Load(content); err != nil {
		return fmt.Errorf("invalid TOML: %w", err)
	}
	home, _ := os.UserHomeDir()
	path := home + "/.codex/config.toml"
	if err := backupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// CodexModelConfig holds the user-facing knobs written to ~/.codex/config.toml.
type CodexModelConfig struct {
	Model        string `json:"model"`
	EffortLevel  string `json:"effortLevel"`
	GoalsEnabled bool   `json:"goalsEnabled"`
}

// GetCodexModelConfig reads Codex model, reasoning effort, and features.goals.
func (a *App) GetCodexModelConfig() CodexModelConfig {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.codex/config.toml")
	if err != nil {
		return CodexModelConfig{}
	}
	content := string(data)
	return CodexModelConfig{
		Model:        readTOMLStringKey(content, "model"),
		EffortLevel:  readTOMLStringKey(content, "model_reasoning_effort"),
		GoalsEnabled: readTOMLSectionBoolKey(content, "features", "goals"),
	}
}

// SetCodexModelConfig merges model, reasoning effort, and features.goals into
// ~/.codex/config.toml without touching unrelated settings.
func (a *App) SetCodexModelConfig(model, effortLevel string, goalsEnabled bool) error {
	home, _ := os.UserHomeDir()
	path := home + "/.codex/config.toml"

	content := ""
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	content = setTOMLStringKey(content, "model_reasoning_effort", effortLevel)
	content = setTOMLStringKey(content, "model", model)
	content = setTOMLSectionBoolKey(content, "features", "goals", goalsEnabled)

	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// ─── Codex presets ──────────────────────────────────────────────────────────

// GetCodexPresets returns the Codex preset list, seeding defaults on first use.
func (a *App) GetCodexPresets() []config.CodexPreset {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.cfg.CodexPresetsSeeded && len(a.cfg.CodexPresets) == 0 {
		a.cfg.CodexPresets = defaultCodexPresets()
		a.cfg.CodexPresetsSeeded = true
		_ = a.cfg.Save()
	} else if migrated, changed := migrateLegacyCodexPresets(a.cfg.CodexPresets); changed {
		a.cfg.CodexPresets = migrated
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
		{ID: "default-fast", Name: "极速轻量", Model: "gpt-5.6-luna", EffortLevel: "low"},
		{ID: "default-balanced", Name: "均衡编码", Model: "gpt-5.6-terra", EffortLevel: "medium"},
		{ID: "default-deep", Name: "复杂任务", Model: "gpt-5.6-sol", EffortLevel: "high"},
		{ID: "default-frontier", Name: "多代理攻坚", Model: "gpt-5.6-sol", EffortLevel: "ultra"},
	}
}

func migrateLegacyCodexPresets(presets []config.CodexPreset) ([]config.CodexPreset, bool) {
	legacy := map[string]config.CodexPreset{
		"default-fast":     {ID: "default-fast", Name: "快速日常", Model: "gpt-5.4-mini", EffortLevel: "low"},
		"default-balanced": {ID: "default-balanced", Name: "均衡编码", Model: "gpt-5.5", EffortLevel: "medium"},
		"default-deep":     {ID: "default-deep", Name: "复杂任务", Model: "gpt-5.5", EffortLevel: "high"},
		"default-frontier": {ID: "default-frontier", Name: "最强推理", Model: "gpt-5.5", EffortLevel: "xhigh"},
	}
	current := map[string]config.CodexPreset{}
	for _, preset := range defaultCodexPresets() {
		current[preset.ID] = preset
	}

	out := make([]config.CodexPreset, len(presets))
	copy(out, presets)
	changed := false
	for i, preset := range out {
		old, ok := legacy[preset.ID]
		if !ok || preset != old {
			continue
		}
		out[i] = current[preset.ID]
		changed = true
	}
	return out, changed
}
