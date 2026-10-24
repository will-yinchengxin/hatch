package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"hatch/core/fsutil"
)

// ClaudeCodeConfig holds the two user-facing knobs written to ~/.claude/settings.json,
// plus EffectiveModel, which resolves an alias like "opus" to the concrete model it
// currently points at. Empty Model means Claude Code runs its own default.
type ClaudeCodeConfig struct {
	Model          string `json:"model"`          // explicit setting in settings.json, "" means default
	EffortLevel    string `json:"effortLevel"`    // explicit setting in settings.json
	EffectiveModel string `json:"effectiveModel"` // Model with aliases resolved, "" when unset
}

// claudeModelAliases maps the aliases accepted by `claude --model` to the concrete
// model each one currently resolves to. Keep in sync with claudeModels in main.js.
// (Anthropic API targets as of 2026-10; opusplan plans with Opus, best is
// Fable when the account has it.)
var claudeModelAliases = map[string]string{
	"opus":     "claude-opus-5-5",
	"sonnet":   "claude-sonnet-5-5",
	"haiku":    "claude-haiku-5-5",
	"fable":    "claude-fable-5-1",
	"best":     "claude-fable-5-1",
	"opusplan": "claude-opus-5-5",
}

// ResolveClaudeModel expands an alias to its concrete model ID; anything else
// (a pinned ID, a custom value, or "") is returned unchanged.
func ResolveClaudeModel(model string) string {
	if concrete, ok := claudeModelAliases[model]; ok {
		return concrete
	}
	return model
}

// ReadClaudeSettings returns the raw JSON of ~/.claude/settings.json ("" if absent).
func ReadClaudeSettings() string {
	path, err := ClaudeSettingsPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// WriteClaudeSettings validates and writes ~/.claude/settings.json.
func WriteClaudeSettings(content string) error {
	if !json.Valid([]byte(content)) {
		return fmt.Errorf("invalid JSON")
	}
	path, err := ClaudeSettingsPath()
	if err != nil {
		return err
	}
	if err := fsutil.BackupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadClaudeCodeConfig reads model and effortLevel from ~/.claude/settings.json.
//
// It deliberately does not shell out: `claude` has no subcommand that reports the
// running model, and an unrecognized argument is treated as a prompt — i.e. probing
// would start a real (slow, billed) session on every call.
func ReadClaudeCodeConfig() ClaudeCodeConfig {
	cfg := ClaudeCodeConfig{}
	var raw map[string]interface{}
	if json.Unmarshal([]byte(ReadClaudeSettings()), &raw) == nil {
		if m, ok := raw["model"].(string); ok {
			cfg.Model = m
		}
		if e, ok := raw["effortLevel"].(string); ok {
			cfg.EffortLevel = e
		}
	}
	cfg.EffectiveModel = ResolveClaudeModel(cfg.Model)
	return cfg
}

// WriteClaudeCodeConfig merges model and effortLevel into ~/.claude/settings.json
// without touching other fields. Empty string removes the key.
func WriteClaudeCodeConfig(model, effortLevel string) error {
	path, err := ClaudeSettingsPath()
	if err != nil {
		return err
	}
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &raw) //nolint:errcheck
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}
	if model != "" {
		raw["model"] = model
	} else {
		delete(raw, "model")
	}
	if effortLevel != "" {
		raw["effortLevel"] = effortLevel
	} else {
		delete(raw, "effortLevel")
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteWithBackup(path, out, 0644)
}
