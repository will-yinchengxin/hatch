package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"hatch/core/config"
)

// readHooksObject parses the "hooks" object from a settings/hooks JSON file.
func readHooksObject(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	hooksObj, _ := raw["hooks"].(map[string]interface{})
	return hooksObj
}

// handlersFor returns the handler list for an event+matcher in a hooks object.
func handlersFor(hooksObj map[string]interface{}, event, matcher string) []interface{} {
	groups, _ := hooksObj[event].([]interface{})
	for _, g := range groups {
		gm, _ := g.(map[string]interface{})
		if gm == nil {
			continue
		}
		if MatcherEquals(gm["matcher"], matcher) {
			hs, _ := gm["hooks"].([]interface{})
			return hs
		}
	}
	return nil
}

func TestWriteClaudeHookConfigEnableAndDisable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(claudeDir, "settings.json")
	// Pre-existing settings with a hand-written hook that must be preserved.
	initial := `{
  "model": "claude-opus-4-8",
  "hooks": {
    "PostToolUse": [
      {"matcher": "Edit|Write", "hooks": [{"type": "command", "command": "user-script.sh"}]}
    ]
  }
}`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	hook := config.HookDef{ID: "fmt", Event: "PostToolUse", Matcher: "Edit|Write", Command: "npx prettier --write .", Timeout: 60}
	if err := WriteClaude(hook, true); err != nil {
		t.Fatal(err)
	}

	hooksObj := readHooksObject(t, path)
	handlers := handlersFor(hooksObj, "PostToolUse", "Edit|Write")
	if len(handlers) != 2 {
		t.Fatalf("expected 2 handlers (user + hatch), got %d", len(handlers))
	}
	// The model key must survive.
	data, _ := os.ReadFile(path)
	var raw map[string]interface{}
	_ = json.Unmarshal(data, &raw)
	if raw["model"] != "claude-opus-4-8" {
		t.Fatalf("model key was lost: %v", raw["model"])
	}

	// Disable: hatch handler removed, user handler kept.
	if err := WriteClaude(hook, false); err != nil {
		t.Fatal(err)
	}
	hooksObj = readHooksObject(t, path)
	handlers = handlersFor(hooksObj, "PostToolUse", "Edit|Write")
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler after disable, got %d", len(handlers))
	}
	hm, _ := handlers[0].(map[string]interface{})
	if hm["command"] != "user-script.sh" {
		t.Fatalf("user handler was not preserved: %v", hm)
	}
}

func TestWriteClaudeHookConfigPrunesEmptyEvent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	hook := config.HookDef{ID: "notify", Event: "Stop", Command: "osascript -e 'beep'"}
	if err := WriteClaude(hook, true); err != nil {
		t.Fatal(err)
	}
	if hooksObj := readHooksObject(t, filepath.Join(home, ".claude", "settings.json")); hooksObj["Stop"] == nil {
		t.Fatal("expected Stop event to be present after enable")
	}
	if err := WriteClaude(hook, false); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	var raw map[string]interface{}
	_ = json.Unmarshal(data, &raw)
	if _, ok := raw["hooks"]; ok {
		t.Fatalf("expected hooks key to be pruned when empty, got: %s", data)
	}
}

func TestWriteCodexHooksFileMergesAndPreserves(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(codexDir, "hooks.json")
	initial := `{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "hand-written.sh"}]}
    ]
  }
}`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	list := []config.HookDef{
		{ID: "fmt", Event: "PostToolUse", Matcher: "Edit|Write", Command: "prettier", EnabledInCodex: true},
		{ID: "off", Event: "Stop", Command: "noop", EnabledInCodex: false},
		{ID: "notif", Event: "Notification", Command: "x", EnabledInCodex: true}, // unsupported by codex → skipped
	}
	if err := WriteCodex(list); err != nil {
		t.Fatal(err)
	}

	hooksObj := readHooksObject(t, path)
	// Hand-written entry preserved.
	if h := handlersFor(hooksObj, "PreToolUse", "Bash"); len(h) != 1 {
		t.Fatalf("hand-written PreToolUse handler not preserved: %v", hooksObj)
	}
	// Enabled+supported hook written.
	if h := handlersFor(hooksObj, "PostToolUse", "Edit|Write"); len(h) != 1 {
		t.Fatalf("expected fmt hook written, got: %v", hooksObj)
	}
	// Disabled hook not written.
	if hooksObj["Stop"] != nil {
		t.Fatal("disabled hook should not be written")
	}
	// Codex-unsupported event skipped.
	if hooksObj["Notification"] != nil {
		t.Fatal("codex-unsupported event should be skipped")
	}

	// Re-run with fmt disabled: hatch entry removed, hand-written kept.
	list[0].EnabledInCodex = false
	if err := WriteCodex(list); err != nil {
		t.Fatal(err)
	}
	hooksObj = readHooksObject(t, path)
	if hooksObj["PostToolUse"] != nil {
		t.Fatal("fmt hook should be removed after disable")
	}
	if h := handlersFor(hooksObj, "PreToolUse", "Bash"); len(h) != 1 {
		t.Fatal("hand-written handler should still be preserved after re-sync")
	}
}

func TestHookEntryExistsInFileDetectsHandwrittenOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(claudeDir, "settings.json")
	content := `{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [
        {"type": "command", "command": "hand.sh"},
        {"type": "command", "command": "mine.sh", "hatchId": "mine"}
      ]}
    ]
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Hand-written command → conflict.
	if ok, _ := ClaudeEntryExists(config.HookDef{Event: "PreToolUse", Matcher: "Bash", Command: "hand.sh"}); !ok {
		t.Fatal("expected hand-written entry to be detected as a conflict")
	}
	// hatchId-owned command → not a conflict.
	if ok, _ := ClaudeEntryExists(config.HookDef{Event: "PreToolUse", Matcher: "Bash", Command: "mine.sh"}); ok {
		t.Fatal("hatch-owned entry must not count as a hand-written conflict")
	}
	// Different command → no conflict.
	if ok, _ := ClaudeEntryExists(config.HookDef{Event: "PreToolUse", Matcher: "Bash", Command: "other.sh"}); ok {
		t.Fatal("non-matching command must not be a conflict")
	}
}
