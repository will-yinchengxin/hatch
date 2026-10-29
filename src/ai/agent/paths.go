// Package agent reads and writes the local configuration of the two coding
// agents Hatch manages — Claude Code and Codex — and reports their CLI/usage
// status. It is the single place that knows where those tools keep their files.
package agent

import (
	"os"
	"path/filepath"
)

const (
	TargetClaude = "claude"
	TargetCodex  = "codex"
)

func homePath(elem ...string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{home}, elem...)...), nil
}

// ClaudeSettingsPath is ~/.claude/settings.json (model, effort, hooks).
func ClaudeSettingsPath() (string, error) { return homePath(".claude", "settings.json") }

// ClaudeJSONPath is ~/.claude.json (global MCP servers).
func ClaudeJSONPath() (string, error) { return homePath(".claude.json") }

// ClaudeSkillsDir is ~/.claude/skills.
func ClaudeSkillsDir() (string, error) { return homePath(".claude", "skills") }

// CodexConfigPath is ~/.codex/config.toml (model, features, MCP servers).
func CodexConfigPath() (string, error) { return homePath(".codex", "config.toml") }

// CodexHooksPath is ~/.codex/hooks.json.
func CodexHooksPath() (string, error) { return homePath(".codex", "hooks.json") }

// CodexSkillsDir is ~/.codex/skills.
func CodexSkillsDir() (string, error) { return homePath(".codex", "skills") }

// CodexSessionsDir is ~/.codex/sessions (rollout JSONL with rate limits).
func CodexSessionsDir() (string, error) { return homePath(".codex", "sessions") }

// ClaudeUsageCachePath is the statusline hook's rate-limit cache.
func ClaudeUsageCachePath() (string, error) {
	return homePath(".cache", "will-statusline", "last.json")
}
