package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hatch/internal/config"
	"hatch/internal/zshrc"
)

// ─── Claude Code settings ────────────────────────────────────────────────────

// GetClaudeSettings returns the raw JSON content of ~/.claude/settings.json.
func (a *App) GetClaudeSettings() string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.claude/settings.json")
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveClaudeSettings validates and writes ~/.claude/settings.json.
func (a *App) SaveClaudeSettings(content string) error {
	if !json.Valid([]byte(content)) {
		return fmt.Errorf("invalid JSON")
	}
	home, _ := os.UserHomeDir()
	path := home + "/.claude/settings.json"
	if err := backupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

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
var claudeModelAliases = map[string]string{
	"opus":   "claude-opus-5",
	"sonnet": "claude-sonnet-5",
	"haiku":  "claude-haiku-4-5",
	"fable":  "claude-fable-5",
}

// resolveClaudeModel expands an alias to its concrete model ID; anything else
// (a pinned ID, a custom value, or "") is returned unchanged.
func resolveClaudeModel(model string) string {
	if concrete, ok := claudeModelAliases[model]; ok {
		return concrete
	}
	return model
}

// GetClaudeCodeConfig reads model and effortLevel from ~/.claude/settings.json.
//
// It deliberately does not shell out: `claude` has no subcommand that reports the
// running model, and an unrecognized argument is treated as a prompt — i.e. probing
// would start a real (slow, billed) session on every call.
func (a *App) GetClaudeCodeConfig() ClaudeCodeConfig {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.claude/settings.json")
	cfg := ClaudeCodeConfig{}
	if err == nil {
		var raw map[string]interface{}
		if json.Unmarshal(data, &raw) == nil {
			if m, ok := raw["model"].(string); ok {
				cfg.Model = m
			}
			if e, ok := raw["effortLevel"].(string); ok {
				cfg.EffortLevel = e
			}
		}
	}
	cfg.EffectiveModel = resolveClaudeModel(cfg.Model)
	return cfg
}

// SetClaudeCodeConfig merges model and effortLevel into ~/.claude/settings.json
// without touching other fields. Empty string removes the key.
func (a *App) SetClaudeCodeConfig(model, effortLevel string) error {
	home, _ := os.UserHomeDir()
	path := home + "/.claude/settings.json"

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
	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

// ─── Claude Code presets ──────────────────────────────────────────────────────

// GetClaudeCodePresets returns the Claude Code preset list, seeding defaults on first use.
func (a *App) GetClaudeCodePresets() []config.ClaudeCodePreset {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.cfg.ClaudeCodePresetsSeeded && len(a.cfg.ClaudeCodePresets) == 0 {
		a.cfg.ClaudeCodePresets = defaultClaudeCodePresets()
		a.cfg.ClaudeCodePresetsSeeded = true
		_ = a.cfg.Save()
	} else if migrated, changed := migrateLegacyClaudeCodePresets(a.cfg.ClaudeCodePresets); changed {
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
		preset.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
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

// defaultClaudeCodePresets seeds the shortcut list. The first three use aliases so
// they follow the latest model in each family; the top tier pins Fable explicitly.
// Haiku carries no effort level because it does not support the effort knob.
func defaultClaudeCodePresets() []config.ClaudeCodePreset {
	return []config.ClaudeCodePreset{
		{ID: "cc-fast", Name: "极速轻量", Model: "haiku", EffortLevel: ""},
		{ID: "cc-balanced", Name: "均衡编码", Model: "sonnet", EffortLevel: "medium"},
		{ID: "cc-deep", Name: "复杂分析", Model: "opus", EffortLevel: "high"},
		{ID: "cc-frontier", Name: "深度推理", Model: "claude-fable-5", EffortLevel: "xhigh"},
	}
}

// migrateLegacyClaudeCodePresets refreshes seeded defaults that the user never
// touched. A preset is only replaced when it still matches the old default
// exactly; edited, custom, and deleted presets are left alone.
func migrateLegacyClaudeCodePresets(presets []config.ClaudeCodePreset) ([]config.ClaudeCodePreset, bool) {
	legacy := map[string]config.ClaudeCodePreset{
		"cc-fast":     {ID: "cc-fast", Name: "快速响应", Model: "claude-haiku-4-5", EffortLevel: "low"},
		"cc-balanced": {ID: "cc-balanced", Name: "均衡编码", Model: "claude-sonnet-4-6", EffortLevel: "medium"},
		"cc-deep":     {ID: "cc-deep", Name: "复杂分析", Model: "claude-sonnet-4-6", EffortLevel: "high"},
		"cc-frontier": {ID: "cc-frontier", Name: "深度推理", Model: "claude-opus-4-8", EffortLevel: "xhigh"},
	}
	current := map[string]config.ClaudeCodePreset{}
	for _, preset := range defaultClaudeCodePresets() {
		current[preset.ID] = preset
	}

	out := make([]config.ClaudeCodePreset, len(presets))
	copy(out, presets)
	changed := false
	for i, preset := range out {
		old, ok := legacy[preset.ID]
		if !ok || preset != old {
			continue
		}
		out[i] = current[preset.ID]
		changed = true
	}
	return out, changed
}

// ─── Shared TOML helpers (used by app_claude.go and app_codex.go) ─────────────

func backupIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_, err := zshrc.Backup(path)
	return err
}

func readTOMLStringKey(content, key string) string {
	preamble, _ := splitTOMLPreamble(content)
	for _, line := range strings.Split(preamble, "\n") {
		if !isTOMLKeyLine(line, key) {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return ""
		}
		quoted := quotedTOMLStringPrefix(strings.TrimSpace(parts[1]))
		if quoted == "" {
			return ""
		}
		value, err := strconv.Unquote(quoted)
		if err != nil {
			return ""
		}
		return value
	}
	return ""
}

func readTOMLSectionBoolKey(content, section, key string) bool {
	inSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if isTOMLSectionLine(line, section) {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(trimmed, "[") {
			return false
		}
		if !inSection || !isTOMLKeyLine(line, key) {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return false
		}
		fields := strings.Fields(strings.TrimSpace(parts[1]))
		if len(fields) == 0 {
			return false
		}
		value, err := strconv.ParseBool(fields[0])
		if err != nil {
			return false
		}
		return value
	}
	return false
}

func setTOMLStringKey(content, key, value string) string {
	preamble, rest := splitTOMLPreamble(content)
	var kept []string
	for _, line := range strings.SplitAfter(preamble, "\n") {
		if line == "" || isTOMLKeyLine(line, key) {
			continue
		}
		kept = append(kept, line)
	}
	content = strings.TrimLeft(strings.Join(kept, "")+rest, "\n")
	if value == "" {
		return ensureTrailingNewline(content)
	}
	line := key + " = " + strconv.Quote(value) + "\n"
	if content == "" {
		return line
	}
	return line + "\n" + ensureTrailingNewline(content)
}

func setTOMLSectionBoolKey(content, section, key string, value bool) string {
	keyLine := key + " = " + strconv.FormatBool(value) + "\n"
	lines := strings.SplitAfter(content, "\n")
	var out []string
	inSection := false
	foundSection := false
	wroteKey := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			if inSection && !wroteKey {
				out = append(out, keyLine)
				wroteKey = true
			}
			inSection = false
			if isTOMLSectionLine(line, section) {
				inSection = true
				foundSection = true
			}
		}
		if inSection && isTOMLKeyLine(line, key) {
			if !wroteKey {
				out = append(out, keyLine)
				wroteKey = true
			}
			continue
		}
		out = append(out, line)
	}

	if inSection && !wroteKey {
		out = append(out, keyLine)
	}

	updated := strings.Join(out, "")
	if foundSection {
		return ensureTrailingNewline(updated)
	}

	updated = ensureTrailingNewline(updated)
	if strings.TrimSpace(updated) != "" {
		updated += "\n"
	}
	updated += "[" + section + "]\n" + keyLine
	return ensureTrailingNewline(updated)
}

func splitTOMLPreamble(content string) (string, string) {
	offset := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			return content[:offset], content[offset:]
		}
		offset += len(line)
	}
	return content, ""
}

func isTOMLKeyLine(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, key) {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, key))
	return strings.HasPrefix(rest, "=")
}

func isTOMLSectionLine(line, section string) bool {
	return strings.TrimSpace(line) == "["+section+"]"
}

func quotedTOMLStringPrefix(s string) string {
	if !strings.HasPrefix(s, `"`) {
		return ""
	}
	escaped := false
	for i := 1; i < len(s); i++ {
		switch {
		case escaped:
			escaped = false
		case s[i] == '\\':
			escaped = true
		case s[i] == '"':
			return s[:i+1]
		}
	}
	return ""
}

func ensureTrailingNewline(content string) string {
	if content == "" || strings.HasSuffix(content, "\n") {
		return content
	}
	return content + "\n"
}
