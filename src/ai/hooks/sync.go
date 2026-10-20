package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"hatch/ai/agent"
	"hatch/core/config"
	"hatch/core/fsutil"
)

// WriteClaude adds (enabled) or removes this hook's Hatch-owned handler in
// ~/.claude/settings.json, leaving every other key and hand-written hook intact.
func WriteClaude(hook config.HookDef, enabled bool) error {
	path, err := agent.ClaudeSettingsPath()
	if err != nil {
		return err
	}
	return rewriteHooksFile(path, "~/.claude/settings.json", func(hooksObj map[string]interface{}) {
		removeHatchHandler(hooksObj, hook.ID)
		if enabled {
			insertHatchHandler(hooksObj, hook)
		}
	})
}

// WriteCodex regenerates the Hatch-managed portion of ~/.codex/hooks.json from
// the given registry, preserving any hand-written (non-hatchId) handlers.
func WriteCodex(list []config.HookDef) error {
	path, err := agent.CodexHooksPath()
	if err != nil {
		return err
	}
	return rewriteHooksFile(path, "~/.codex/hooks.json", func(hooksObj map[string]interface{}) {
		removeHatchHandler(hooksObj, "") // strip all Hatch-managed handlers
		for _, h := range list {
			if h.EnabledInCodex && SupportedBy(h.Event, TargetCodex) {
				insertHatchHandler(hooksObj, h)
			}
		}
	})
}

// ClaudeEntryExists reports a hand-written Claude handler matching hook.
func ClaudeEntryExists(hook config.HookDef) (bool, error) {
	path, err := agent.ClaudeSettingsPath()
	if err != nil {
		return false, err
	}
	return EntryExistsInFile(path, hook)
}

// CodexEntryExists reports a hand-written Codex handler matching hook.
func CodexEntryExists(hook config.HookDef) (bool, error) {
	path, err := agent.CodexHooksPath()
	if err != nil {
		return false, err
	}
	return EntryExistsInFile(path, hook)
}

// EntryExistsInFile reports whether the file already has a hand-written
// (non-hatchId) handler under the same event+matcher with the same command.
func EntryExistsInFile(path string, hook config.HookDef) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return false, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("invalid %s: %w", path, err)
	}
	hooksObj, _ := raw["hooks"].(map[string]interface{})
	groups, _ := hooksObj[hook.Event].([]interface{})
	for _, g := range groups {
		gm, ok := g.(map[string]interface{})
		if !ok || !MatcherEquals(gm["matcher"], hook.Matcher) {
			continue
		}
		handlers, _ := gm["hooks"].([]interface{})
		for _, h := range handlers {
			hm, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if hid, has := hm["hatchId"].(string); has && hid != "" {
				continue // Hatch-owned, not a hand-written conflict
			}
			if cmd, _ := hm["command"].(string); cmd == hook.Command {
				return true, nil
			}
		}
	}
	return false, nil
}

// MatcherEquals compares a raw JSON matcher value with a hook matcher.
func MatcherEquals(v interface{}, matcher string) bool {
	s, _ := v.(string)
	return s == matcher
}

// rewriteHooksFile loads a JSON file, lets mutate edit its "hooks" object, and
// writes it back (with a backup), pruning the key when it ends up empty.
func rewriteHooksFile(path, label string, mutate func(map[string]interface{})) error {
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("invalid %s: %w", label, err)
		}
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}
	hooksObj, _ := raw["hooks"].(map[string]interface{})
	if hooksObj == nil {
		hooksObj = map[string]interface{}{}
	}
	mutate(hooksObj)
	if len(hooksObj) == 0 {
		delete(raw, "hooks")
	} else {
		raw["hooks"] = hooksObj
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteWithBackup(path, append(out, '\n'), 0o644)
}

// renderHandler builds the JSON handler object Hatch writes into both targets.
func renderHandler(hook config.HookDef) map[string]interface{} {
	h := map[string]interface{}{
		"type":    "command",
		"command": hook.Command,
		"hatchId": hook.ID,
	}
	if len(hook.Args) > 0 {
		args := make([]interface{}, len(hook.Args))
		for i, a := range hook.Args {
			args[i] = a
		}
		h["args"] = args
	}
	if hook.Timeout > 0 {
		h["timeout"] = hook.Timeout
	}
	if hook.StatusMessage != "" {
		h["statusMessage"] = hook.StatusMessage
	}
	return h
}

// removeHatchHandler drops Hatch-owned handlers from a hooks object. When id is
// empty, all hatchId-marked handlers are removed; otherwise only the matching id.
// Hand-written handlers (no hatchId) are never touched. Empty groups/events prune.
func removeHatchHandler(hooksObj map[string]interface{}, id string) {
	for event, v := range hooksObj {
		groups, ok := v.([]interface{})
		if !ok {
			continue
		}
		var keptGroups []interface{}
		for _, g := range groups {
			gm, ok := g.(map[string]interface{})
			if !ok {
				keptGroups = append(keptGroups, g)
				continue
			}
			handlers, _ := gm["hooks"].([]interface{})
			var keptHandlers []interface{}
			for _, h := range handlers {
				if hm, ok := h.(map[string]interface{}); ok {
					if hid, has := hm["hatchId"].(string); has && hid != "" && (id == "" || hid == id) {
						continue
					}
				}
				keptHandlers = append(keptHandlers, h)
			}
			if len(keptHandlers) == 0 {
				continue
			}
			gm["hooks"] = keptHandlers
			keptGroups = append(keptGroups, gm)
		}
		if len(keptGroups) == 0 {
			delete(hooksObj, event)
		} else {
			hooksObj[event] = keptGroups
		}
	}
}

// insertHatchHandler appends the hook's rendered handler into the right event,
// reusing an existing matcher group when one matches, otherwise creating one.
func insertHatchHandler(hooksObj map[string]interface{}, hook config.HookDef) {
	handler := renderHandler(hook)
	groups, _ := hooksObj[hook.Event].([]interface{})
	for _, g := range groups {
		gm, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		if MatcherEquals(gm["matcher"], hook.Matcher) {
			handlers, _ := gm["hooks"].([]interface{})
			gm["hooks"] = append(handlers, handler)
			hooksObj[hook.Event] = groups
			return
		}
	}
	newGroup := map[string]interface{}{"hooks": []interface{}{handler}}
	if hook.Matcher != "" {
		newGroup["matcher"] = hook.Matcher
	}
	hooksObj[hook.Event] = append(groups, newGroup)
}
