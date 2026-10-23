package app

import (
	"context"
	"strconv"
	"time"

	"hatch/ai/agent"
	"hatch/core/config"
)

// ─── AI · 助手（Claude Code / Codex） ────────────────────────────────────────

// GetClaudeCodeStatus reports claude CLI install/version, VPN tunnel state and
// cached rate-limit usage.
func (a *App) GetClaudeCodeStatus() agent.ClaudeCodeStatus { return agent.ClaudeStatus() }

// RefreshClaudeUsage sends a one-line Haiku prompt through the claude CLI to
// pull fresh 5h/7d rate limits into the statusline cache, then returns the
// updated status.
func (a *App) RefreshClaudeUsage() (agent.ClaudeCodeStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := agent.RefreshClaudeUsage(ctx); err != nil {
		return agent.ClaudeStatus(), err
	}
	return agent.ClaudeStatus(), nil
}

// GetCodexUsageStatus reports codex CLI install/version and the latest local
// rate-limit snapshot.
func (a *App) GetCodexUsageStatus() agent.CodexUsageStatus { return agent.CodexStatus() }

// GetClaudeSettings returns the raw JSON content of ~/.claude/settings.json.
func (a *App) GetClaudeSettings() string { return agent.ReadClaudeSettings() }

// SaveClaudeSettings validates and writes ~/.claude/settings.json.
func (a *App) SaveClaudeSettings(content string) error { return agent.WriteClaudeSettings(content) }

// GetClaudeCodeConfig reads model and effortLevel from ~/.claude/settings.json.
func (a *App) GetClaudeCodeConfig() agent.ClaudeCodeConfig { return agent.ReadClaudeCodeConfig() }

// SetClaudeCodeConfig merges model/effortLevel into ~/.claude/settings.json
// ("" removes the key).
func (a *App) SetClaudeCodeConfig(model, effortLevel string) error {
	return agent.WriteClaudeCodeConfig(model, effortLevel)
}

// GetCodexConfig returns the raw TOML content of ~/.codex/config.toml.
func (a *App) GetCodexConfig() string { return agent.ReadCodexConfig() }

// SaveCodexConfig validates and writes ~/.codex/config.toml.
func (a *App) SaveCodexConfig(content string) error { return agent.WriteCodexConfig(content) }

// GetCodexModelConfig reads Codex model, reasoning effort and features.goals.
func (a *App) GetCodexModelConfig() agent.CodexModelConfig { return agent.ReadCodexModelConfig() }

// SetCodexModelConfig merges model, reasoning effort and features.goals into
// ~/.codex/config.toml without touching unrelated settings.
func (a *App) SetCodexModelConfig(model, effortLevel string, goalsEnabled bool) error {
	return agent.WriteCodexModelConfig(model, effortLevel, goalsEnabled)
}

// ─── Presets (stored in Hatch config) ────────────────────────────────────────

// GetClaudeCodePresets returns the Claude Code preset list, seeding defaults on
// first use and refreshing untouched legacy defaults.
func (a *App) GetClaudeCodePresets() []config.ClaudeCodePreset {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.cfg.ClaudeCodePresetsSeeded && len(a.cfg.ClaudeCodePresets) == 0 {
		a.cfg.ClaudeCodePresets = agent.DefaultClaudeCodePresets()
		a.cfg.ClaudeCodePresetsSeeded = true
		_ = a.cfg.Save()
	} else if migrated, changed := agent.MigrateLegacyClaudeCodePresets(a.cfg.ClaudeCodePresets); changed {
		a.cfg.ClaudeCodePresets = migrated
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
		preset.ID = newPresetID()
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

// GetCodexPresets returns the Codex preset list, seeding defaults on first use
// and refreshing untouched legacy defaults.
func (a *App) GetCodexPresets() []config.CodexPreset {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.cfg.CodexPresetsSeeded && len(a.cfg.CodexPresets) == 0 {
		a.cfg.CodexPresets = agent.DefaultCodexPresets()
		a.cfg.CodexPresetsSeeded = true
		_ = a.cfg.Save()
	} else if migrated, changed := agent.MigrateLegacyCodexPresets(a.cfg.CodexPresets); changed {
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
		preset.ID = newPresetID()
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

func newPresetID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
