package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	snapshot, ok := loadLatestCodexRateLimitSnapshot(dir)
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
	_, ok := parseCodexRateLimitLine(`{"timestamp":"2026-05-15T08:40:00.000Z","payload":{"type":"other"}}`)
	if ok {
		t.Fatal("expected non-token_count payload to be ignored")
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

	if got := readTOMLStringKey(content, "model"); got != "gpt-5.4" {
		t.Fatalf("model = %q", got)
	}
	if got := readTOMLStringKey(content, "model_reasoning_effort"); got != "high" {
		t.Fatalf("effort = %q", got)
	}
	if got := readTOMLSectionBoolKey(content, "features", "goals"); got {
		t.Fatalf("goals = %v", got)
	}

	updated := setTOMLStringKey(content, "model_reasoning_effort", "xhigh")
	updated = setTOMLStringKey(updated, "model", "gpt-5.5")
	updated = setTOMLSectionBoolKey(updated, "features", "goals", true)

	if got := readTOMLStringKey(updated, "model"); got != "gpt-5.5" {
		t.Fatalf("updated model = %q", got)
	}
	if got := readTOMLStringKey(updated, "model_reasoning_effort"); got != "xhigh" {
		t.Fatalf("updated effort = %q", got)
	}
	if got := readTOMLSectionBoolKey(updated, "features", "goals"); !got {
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
	updated := setTOMLSectionBoolKey(`model = "gpt-5.5"`+"\n", "features", "goals", true)

	if !strings.Contains(updated, "[features]\ngoals = true\n") {
		t.Fatalf("features.goals section not created:\n%s", updated)
	}
	if got := readTOMLSectionBoolKey(updated, "features", "goals"); !got {
		t.Fatalf("goals = %v", got)
	}
}
