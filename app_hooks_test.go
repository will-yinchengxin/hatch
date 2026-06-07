package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"hatch/internal/config"
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
		if matcherEquals(gm["matcher"], matcher) {
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
	if err := writeClaudeHookConfig(hook, true); err != nil {
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
	if err := writeClaudeHookConfig(hook, false); err != nil {
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
	if err := writeClaudeHookConfig(hook, true); err != nil {
		t.Fatal(err)
	}
	if hooksObj := readHooksObject(t, filepath.Join(home, ".claude", "settings.json")); hooksObj["Stop"] == nil {
		t.Fatal("expected Stop event to be present after enable")
	}
	if err := writeClaudeHookConfig(hook, false); err != nil {
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
	if err := writeCodexHooksFile(list); err != nil {
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
	if err := writeCodexHooksFile(list); err != nil {
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
	if ok, _ := claudeHookEntryExists(config.HookDef{Event: "PreToolUse", Matcher: "Bash", Command: "hand.sh"}); !ok {
		t.Fatal("expected hand-written entry to be detected as a conflict")
	}
	// hatchId-owned command → not a conflict.
	if ok, _ := claudeHookEntryExists(config.HookDef{Event: "PreToolUse", Matcher: "Bash", Command: "mine.sh"}); ok {
		t.Fatal("hatch-owned entry must not count as a hand-written conflict")
	}
	// Different command → no conflict.
	if ok, _ := claudeHookEntryExists(config.HookDef{Event: "PreToolUse", Matcher: "Bash", Command: "other.sh"}); ok {
		t.Fatal("non-matching command must not be a conflict")
	}
}

func TestListHooksSeedsHelloHookOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := &App{cfg: &config.Config{}}

	hooks, err := a.ListHooks()
	if err != nil {
		t.Fatal(err)
	}
	var hello *config.HookDef
	for i := range hooks {
		if hooks[i].ID == "hello-test" {
			hello = &hooks[i]
		}
	}
	if hello == nil {
		t.Fatal("expected seeded hello-test hook")
	}
	if hello.Event != "UserPromptSubmit" {
		t.Fatalf("seeded hook event = %s, want UserPromptSubmit", hello.Event)
	}
	if !hello.EnabledInClaude {
		t.Fatal("seeded hook should be enabled in Claude")
	}
	// It must be written into ~/.claude/settings.json.
	hooksObj := readHooksObject(t, filepath.Join(home, ".claude", "settings.json"))
	if len(handlersFor(hooksObj, "UserPromptSubmit", "")) != 1 {
		t.Fatalf("seeded hook not written to settings.json: %v", hooksObj)
	}
	// Seeding must not repeat.
	again, _ := a.ListHooks()
	n := 0
	for _, h := range again {
		if h.ID == "hello-test" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("hello hook seeded %d times, want 1", n)
	}
}

func TestSetHookEnabledGatesUnsupportedTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := &App{cfg: &config.Config{}}
	a.cfg.Hooks = []config.HookDef{{ID: "n", Event: "Notification", Command: "x"}}
	err := a.SetHookEnabled("n", hookTargetCodex, true)
	if err == nil {
		t.Fatal("expected error enabling a Codex-unsupported event")
	}
}

func TestSaveHookValidatesAndPreservesFlags(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	a := &App{cfg: &config.Config{}}

	if err := a.SaveHook(config.HookDef{Name: "x", Event: "PostToolUse"}); err == nil {
		t.Fatal("expected error for missing command")
	}
	if err := a.SaveHook(config.HookDef{Name: "x", Event: "BogusEvent", Command: "c"}); err == nil {
		t.Fatal("expected error for invalid event")
	}

	if err := a.SaveHook(config.HookDef{ID: "h1", Name: "First", Event: "Stop", Command: "a"}); err != nil {
		t.Fatal(err)
	}
	a.cfg.Hooks[0].EnabledInClaude = true
	if err := a.SaveHook(config.HookDef{ID: "h1", Name: "Edited", Event: "Stop", Command: "b"}); err != nil {
		t.Fatal(err)
	}
	if !a.cfg.Hooks[0].EnabledInClaude {
		t.Fatal("EnabledInClaude flag should be preserved across SaveHook")
	}
	if a.cfg.Hooks[0].Command != "b" || a.cfg.Hooks[0].Name != "Edited" {
		t.Fatalf("hook was not updated: %+v", a.cfg.Hooks[0])
	}
}
