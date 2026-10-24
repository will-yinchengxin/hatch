package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hatch/core/config"
)

func TestLoadLatestCodexRateLimitSnapshot(t *testing.T) {
	dir := t.TempDir()
	sessionDir := filepath.Join(dir, "2026", "05", "15")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `{"timestamp":"2026-05-15T08:30:00.000Z","payload":{"type":"token_count","rate_limits":{"plan_type":"plus","primary":{"used_percent":12,"window_minutes":300,"resets_at":1778849000},"secondary":{"used_percent":9,"window_minutes":10080,"resets_at":1779358000}}}}
{"timestamp":"2026-05-15T08:40:00.000Z","payload":{"type":"token_count","rate_limits":{"plan_type":"pro","rate_limit_reached_type":"primary","primary":{"used_percent":36,"window_minutes":300,"resets_at":1778849521},"secondary":{"used_percent":22,"window_minutes":10080,"resets_at":1779358482}}}}
`
	if err := os.WriteFile(filepath.Join(sessionDir, "rollout-test.jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	snapshot, ok := LoadLatestCodexRateLimitSnapshot(dir)
	if !ok {
		t.Fatal("expected a Codex rate-limit snapshot")
	}
	if snapshot.Timestamp != "2026-05-15T08:40:00.000Z" {
		t.Fatalf("expected latest timestamp, got %q", snapshot.Timestamp)
	}
	if snapshot.PlanType != "pro" || snapshot.LimitReached != "primary" {
		t.Fatalf("unexpected plan/limit: %#v", snapshot)
	}
	if !snapshot.Primary.Available || snapshot.Primary.UsedPercent != 36 || snapshot.Primary.WindowMinutes != 300 {
		t.Fatalf("unexpected primary window: %#v", snapshot.Primary)
	}
	if !snapshot.Secondary.Available || snapshot.Secondary.UsedPercent != 22 || snapshot.Secondary.WindowMinutes != 10080 {
		t.Fatalf("unexpected secondary window: %#v", snapshot.Secondary)
	}
}

func TestParseCodexRateLimitLineIgnoresNonTokenCount(t *testing.T) {
	_, ok := ParseCodexRateLimitLine(`{"timestamp":"2026-05-15T08:40:00.000Z","payload":{"type":"other"}}`)
	if ok {
		t.Fatal("expected non-token_count payload to be ignored")
	}
}

func TestParseCodexRateLimitLineClassifiesWeeklyOnlyPrimary(t *testing.T) {
	snapshot, ok := ParseCodexRateLimitLine(`{"timestamp":"2026-08-13T09:09:44.778Z","payload":{"type":"token_count","rate_limits":{"plan_type":"plus","primary":{"used_percent":4,"window_minutes":10080,"resets_at":1787215190},"secondary":null}}}`)
	if !ok {
		t.Fatal("expected weekly-only rate-limit snapshot")
	}
	if snapshot.Primary.Available {
		t.Fatalf("5h window should be unavailable: %#v", snapshot.Primary)
	}
	if !snapshot.Secondary.Available || snapshot.Secondary.WindowMinutes != 10080 || snapshot.Secondary.UsedPercent != 4 {
		t.Fatalf("weekly window was not classified as 7d: %#v", snapshot.Secondary)
	}
}

func TestClassifyCodexRateLimitWindowsUsesDurationNotPosition(t *testing.T) {
	weekly := CodexRateLimitWindow{Available: true, UsedPercent: 11, WindowMinutes: 10080}
	fiveHour := CodexRateLimitWindow{Available: true, UsedPercent: 22, WindowMinutes: 300}
	gotFiveHour, gotSevenDay := ClassifyCodexRateLimitWindows(weekly, fiveHour)
	if gotFiveHour != fiveHour || gotSevenDay != weekly {
		t.Fatalf("windows classified by position: 5h=%#v 7d=%#v", gotFiveHour, gotSevenDay)
	}
}

func TestCodexModelConfigTOMLHelpersOnlyTouchTopLevelKeys(t *testing.T) {
	content := `# user config
model = "gpt-5.4"
model_reasoning_effort = "high"

[features]
goals = false

[profiles.deep]
model = "do-not-touch"
model_reasoning_effort = "do-not-touch"
goals = false
`

	if got := ReadTOMLStringKey(content, "model"); got != "gpt-5.4" {
		t.Fatalf("model = %q", got)
	}
	if got := ReadTOMLStringKey(content, "model_reasoning_effort"); got != "high" {
		t.Fatalf("effort = %q", got)
	}
	if got := ReadTOMLSectionBoolKey(content, "features", "goals"); got {
		t.Fatalf("goals = %v", got)
	}

	updated := SetTOMLStringKey(content, "model_reasoning_effort", "xhigh")
	updated = SetTOMLStringKey(updated, "model", "gpt-5.5")
	updated = SetTOMLSectionBoolKey(updated, "features", "goals", true)

	if got := ReadTOMLStringKey(updated, "model"); got != "gpt-5.5" {
		t.Fatalf("updated model = %q", got)
	}
	if got := ReadTOMLStringKey(updated, "model_reasoning_effort"); got != "xhigh" {
		t.Fatalf("updated effort = %q", got)
	}
	if got := ReadTOMLSectionBoolKey(updated, "features", "goals"); !got {
		t.Fatalf("updated goals = %v", got)
	}
	if !strings.Contains(updated, `[profiles.deep]`) || !strings.Contains(updated, `model = "do-not-touch"`) {
		t.Fatalf("nested section was modified:\n%s", updated)
	}
	if !strings.Contains(updated, `goals = false`) {
		t.Fatalf("nested goals setting was modified:\n%s", updated)
	}
}

func TestSetTOMLSectionBoolKeyCreatesFeaturesSection(t *testing.T) {
	updated := SetTOMLSectionBoolKey(`model = "gpt-5.5"`+"\n", "features", "goals", true)

	if !strings.Contains(updated, "[features]\ngoals = true\n") {
		t.Fatalf("features.goals section not created:\n%s", updated)
	}
	if got := ReadTOMLSectionBoolKey(updated, "features", "goals"); !got {
		t.Fatalf("goals = %v", got)
	}
}

func TestDefaultCodexPresetsUseCurrentModelFamily(t *testing.T) {
	presets := DefaultCodexPresets()
	want := map[string]string{
		"default-fast":     "gpt-6-luna/low",
		"default-balanced": "gpt-6-sol/medium",
		"default-deep":     "gpt-6.1-sol/high",
		"default-frontier": "gpt-6-astra/xhigh",
	}
	for _, preset := range presets {
		if got := preset.Model + "/" + preset.EffortLevel; got != want[preset.ID] {
			t.Fatalf("preset %q = %q, want %q", preset.ID, got, want[preset.ID])
		}
	}
}

func TestMigrateLegacyCodexPresetsOnlyUpdatesUntouchedDefaults(t *testing.T) {
	legacy := []config.CodexPreset{
		{ID: "default-fast", Name: "快速日常", Model: "gpt-5.4-mini", EffortLevel: "low"},
		{ID: "default-balanced", Name: "我的均衡", Model: "gpt-5.5", EffortLevel: "medium"},
		{ID: "custom", Name: "自定义", Model: "private-model", EffortLevel: "high"},
	}

	got, changed := MigrateLegacyCodexPresets(legacy)
	if !changed {
		t.Fatal("expected untouched legacy preset to migrate")
	}
	if got[0].Model != "gpt-6-luna" || got[0].EffortLevel != "low" || got[0].Name != "极速轻量" {
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

func TestMigrateLegacyCodexPresetsNoopForCurrentValues(t *testing.T) {
	presets := DefaultCodexPresets()
	got, changed := MigrateLegacyCodexPresets(presets)
	if changed {
		t.Fatal("current presets should not be reported as migrated")
	}
	for i := range presets {
		if got[i] != presets[i] {
			t.Fatalf("preset %d changed: %#v", i, got[i])
		}
	}
}

func TestMigratePresetsCoversEveryPastGeneration(t *testing.T) {
	gen56 := []config.CodexPreset{{ID: "default-deep", Name: "复杂任务", Model: "gpt-5.6-sol", EffortLevel: "high"}}
	got, changed := MigrateLegacyCodexPresets(gen56)
	if !changed || got[0].Model != "gpt-6.1-sol" {
		t.Fatalf("5.6 default not migrated: %#v", got)
	}
	gen50 := []config.ClaudeCodePreset{{ID: "cc-fast", Name: "极速轻量", Model: "haiku", EffortLevel: ""}}
	gotC, changed := MigrateLegacyClaudeCodePresets(gen50)
	if !changed || gotC[0].EffortLevel != "low" {
		t.Fatalf("Claude 5.0 default not migrated: %#v", gotC)
	}
}
