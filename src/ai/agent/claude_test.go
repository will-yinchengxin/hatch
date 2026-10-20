package agent

import (
	"testing"

	"hatch/core/config"
)

func TestDefaultClaudeCodePresetsUseCurrentModelFamily(t *testing.T) {
	presets := DefaultClaudeCodePresets()
	want := map[string]string{
		"cc-fast":     "haiku/",
		"cc-balanced": "sonnet/medium",
		"cc-deep":     "opus/high",
		"cc-frontier": "claude-fable-5/xhigh",
	}
	for _, preset := range presets {
		if got := preset.Model + "/" + preset.EffortLevel; got != want[preset.ID] {
			t.Fatalf("preset %q = %q, want %q", preset.ID, got, want[preset.ID])
		}
	}
}

func TestMigrateLegacyClaudeCodePresetsOnlyUpdatesUntouchedDefaults(t *testing.T) {
	legacy := []config.ClaudeCodePreset{
		{ID: "cc-fast", Name: "快速响应", Model: "claude-haiku-4-5", EffortLevel: "low"},
		{ID: "cc-balanced", Name: "我的均衡", Model: "claude-sonnet-4-6", EffortLevel: "medium"},
		{ID: "custom", Name: "自定义", Model: "private-model", EffortLevel: "high"},
	}

	got, changed := MigrateLegacyClaudeCodePresets(legacy)
	if !changed {
		t.Fatal("expected untouched legacy preset to migrate")
	}
	if got[0].Model != "haiku" || got[0].EffortLevel != "" || got[0].Name != "极速轻量" {
		t.Fatalf("legacy default was not migrated: %#v", got[0])
	}
	if got[1] != legacy[1] {
		t.Fatalf("edited default should be preserved: %#v", got[1])
	}
	if got[2] != legacy[2] {
		t.Fatalf("custom preset should be preserved: %#v", got[2])
	}
	if len(got) != len(legacy) {
		t.Fatalf("migration must not restore deleted defaults: got %d, want %d", len(got), len(legacy))
	}
}

func TestMigrateLegacyClaudeCodePresetsNoopForCurrentValues(t *testing.T) {
	presets := DefaultClaudeCodePresets()
	got, changed := MigrateLegacyClaudeCodePresets(presets)
	if changed {
		t.Fatal("current presets should not be reported as migrated")
	}
	for i := range presets {
		if got[i] != presets[i] {
			t.Fatalf("preset %d changed: %#v", i, got[i])
		}
	}
}

func TestResolveClaudeModelExpandsAliasesOnly(t *testing.T) {
	cases := map[string]string{
		"opus":              "claude-opus-5",
		"sonnet":            "claude-sonnet-5",
		"haiku":             "claude-haiku-4-5",
		"fable":             "claude-fable-5",
		"claude-opus-4-8":   "claude-opus-4-8",
		"some-custom-model": "some-custom-model",
		"":                  "",
	}
	for in, want := range cases {
		if got := ResolveClaudeModel(in); got != want {
			t.Fatalf("ResolveClaudeModel(%q) = %q, want %q", in, got, want)
		}
	}
}
