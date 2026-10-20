package agent

import "hatch/core/config"

// DefaultClaudeCodePresets seeds the shortcut list. The first three use aliases so
// they follow the latest model in each family; the top tier pins Fable explicitly.
// Haiku carries no effort level because it does not support the effort knob.
func DefaultClaudeCodePresets() []config.ClaudeCodePreset {
	return []config.ClaudeCodePreset{
		{ID: "cc-fast", Name: "极速轻量", Model: "haiku", EffortLevel: ""},
		{ID: "cc-balanced", Name: "均衡编码", Model: "sonnet", EffortLevel: "medium"},
		{ID: "cc-deep", Name: "复杂分析", Model: "opus", EffortLevel: "high"},
		{ID: "cc-frontier", Name: "深度推理", Model: "claude-fable-5", EffortLevel: "xhigh"},
	}
}

// MigrateLegacyClaudeCodePresets refreshes seeded defaults that the user never
// touched. A preset is only replaced when it still matches the old default
// exactly; edited, custom, and deleted presets are left alone.
func MigrateLegacyClaudeCodePresets(presets []config.ClaudeCodePreset) ([]config.ClaudeCodePreset, bool) {
	legacy := map[string]config.ClaudeCodePreset{
		"cc-fast":     {ID: "cc-fast", Name: "快速响应", Model: "claude-haiku-4-5", EffortLevel: "low"},
		"cc-balanced": {ID: "cc-balanced", Name: "均衡编码", Model: "claude-sonnet-4-6", EffortLevel: "medium"},
		"cc-deep":     {ID: "cc-deep", Name: "复杂分析", Model: "claude-sonnet-4-6", EffortLevel: "high"},
		"cc-frontier": {ID: "cc-frontier", Name: "深度推理", Model: "claude-opus-4-8", EffortLevel: "xhigh"},
	}
	return migratePresets(presets, legacy, DefaultClaudeCodePresets())
}

// DefaultCodexPresets seeds the Codex shortcut list.
func DefaultCodexPresets() []config.CodexPreset {
	return []config.CodexPreset{
		{ID: "default-fast", Name: "极速轻量", Model: "gpt-5.6-luna", EffortLevel: "low"},
		{ID: "default-balanced", Name: "均衡编码", Model: "gpt-5.6-terra", EffortLevel: "medium"},
		{ID: "default-deep", Name: "复杂任务", Model: "gpt-5.6-sol", EffortLevel: "high"},
		{ID: "default-frontier", Name: "多代理攻坚", Model: "gpt-5.6-sol", EffortLevel: "ultra"},
	}
}

// MigrateLegacyCodexPresets is the Codex counterpart of MigrateLegacyClaudeCodePresets.
func MigrateLegacyCodexPresets(presets []config.CodexPreset) ([]config.CodexPreset, bool) {
	legacy := map[string]config.CodexPreset{
		"default-fast":     {ID: "default-fast", Name: "快速日常", Model: "gpt-5.4-mini", EffortLevel: "low"},
		"default-balanced": {ID: "default-balanced", Name: "均衡编码", Model: "gpt-5.5", EffortLevel: "medium"},
		"default-deep":     {ID: "default-deep", Name: "复杂任务", Model: "gpt-5.5", EffortLevel: "high"},
		"default-frontier": {ID: "default-frontier", Name: "最强推理", Model: "gpt-5.5", EffortLevel: "xhigh"},
	}
	return migratePresets(presets, legacy, DefaultCodexPresets())
}

// migratePresets replaces every preset that still equals its legacy default
// with the current default of the same ID.
func migratePresets[P interface {
	config.ClaudeCodePreset | config.CodexPreset
}](presets []P, legacy map[string]P, defaults []P) ([]P, bool) {
	current := make(map[string]P, len(defaults))
	for _, p := range defaults {
		current[idOf(p)] = p
	}
	out := make([]P, len(presets))
	copy(out, presets)
	changed := false
	for i, p := range out {
		old, ok := legacy[idOf(p)]
		if !ok || p != old {
			continue
		}
		out[i] = current[idOf(p)]
		changed = true
	}
	return out, changed
}

func idOf[P config.ClaudeCodePreset | config.CodexPreset](p P) string {
	switch v := any(p).(type) {
	case config.ClaudeCodePreset:
		return v.ID
	case config.CodexPreset:
		return v.ID
	}
	return ""
}
