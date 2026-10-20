package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"hatch/ai/hooks"
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
		if hooks.MatcherEquals(gm["matcher"], matcher) {
			hs, _ := gm["hooks"].([]interface{})
			return hs
		}
	}
	return nil
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
	err := a.SetHookEnabled("n", hooks.TargetCodex, true)
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
