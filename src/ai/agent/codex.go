package agent

import (
	"fmt"
	"os"

	"hatch/core/fsutil"

	"github.com/pelletier/go-toml"
)

// CodexModelConfig holds the user-facing knobs written to ~/.codex/config.toml.
type CodexModelConfig struct {
	Model        string `json:"model"`
	EffortLevel  string `json:"effortLevel"`
	GoalsEnabled bool   `json:"goalsEnabled"`
}

// ReadCodexConfig returns the raw TOML of ~/.codex/config.toml ("" if absent).
func ReadCodexConfig() string {
	path, err := CodexConfigPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// ValidTOML reports a parse error for invalid TOML content.
func ValidTOML(content string) error {
	_, err := toml.Load(content)
	return err
}

// WriteCodexConfig validates and writes ~/.codex/config.toml.
func WriteCodexConfig(content string) error {
	if err := ValidTOML(content); err != nil {
		return fmt.Errorf("invalid TOML: %w", err)
	}
	path, err := CodexConfigPath()
	if err != nil {
		return err
	}
	if err := fsutil.BackupIfExists(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadCodexModelConfig reads Codex model, reasoning effort, and features.goals.
func ReadCodexModelConfig() CodexModelConfig {
	content := ReadCodexConfig()
	if content == "" {
		return CodexModelConfig{}
	}
	return CodexModelConfig{
		Model:        ReadTOMLStringKey(content, "model"),
		EffortLevel:  ReadTOMLStringKey(content, "model_reasoning_effort"),
		GoalsEnabled: ReadTOMLSectionBoolKey(content, "features", "goals"),
	}
}

// WriteCodexModelConfig merges model, reasoning effort, and features.goals into
// ~/.codex/config.toml without touching unrelated settings.
func WriteCodexModelConfig(model, effortLevel string, goalsEnabled bool) error {
	path, err := CodexConfigPath()
	if err != nil {
		return err
	}
	content := ReadCodexConfig()
	content = SetTOMLStringKey(content, "model_reasoning_effort", effortLevel)
	content = SetTOMLStringKey(content, "model", model)
	content = SetTOMLSectionBoolKey(content, "features", "goals", goalsEnabled)
	return fsutil.WriteWithBackup(path, []byte(content), 0o644)
}
