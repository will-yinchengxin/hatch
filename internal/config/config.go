package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ScriptType string

const (
	TypeOneshot ScriptType = "oneshot"
	TypeService ScriptType = "service"
)

type ArgDef struct {
	Name     string `json:"name"`
	Flag     string `json:"flag"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
}

// CommandPreset is a saved named invocation of a script with fixed arguments.
type CommandPreset struct {
	Name string `json:"name"` // display name, e.g. "2026年4月日报"
	Args string `json:"args"` // space-separated args string
}

type ScriptMeta struct {
	Path         string            `json:"path"`
	Name         string            `json:"name"`
	Desc         string            `json:"desc"`
	Type         ScriptType        `json:"type"`
	Language     string            `json:"language"`
	Commands     []string          `json:"commands"`
	PidFile      string            `json:"pidFile"`
	Args         []ArgDef          `json:"args"`
	Presets      []CommandPreset   `json:"presets,omitempty"`
	Tags         []string          `json:"tags"`
	Env          map[string]string `json:"env,omitempty"`          // v1.1: per-script env overrides
	NotifyOnDone bool              `json:"notifyOnDone,omitempty"` // v1.1: send macOS notification when oneshot completes
	Schedule     string            `json:"schedule,omitempty"`     // v1.2: scheduler spec (@every 30s / @daily HH:MM / @hourly :MM)
	Requires     []string          `json:"requires,omitempty"`     // v1.2: required CLI commands
	HealthURL    string            `json:"healthUrl,omitempty"`    // v1.2: health-check URL for service scripts
	PreScripts   []string          `json:"preScripts,omitempty"`   // v1.3: scripts to run (in order) before this one; failure aborts
	LastRunAt    int64             `json:"lastRunAt"`
	UpdatedAt    int64             `json:"updatedAt"`
}

type Directory struct {
	Path        string `json:"path"`
	IsScriptDir bool   `json:"isScriptDir"`
	ExportName  string `json:"exportName"`
}

type ExportVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AIConfig struct {
	Enabled         bool   `json:"enabled"`
	BaseURL         string `json:"baseUrl"`
	APIKey          string `json:"apiKey"`
	Model           string `json:"model"`
	TraceRetainDays int    `json:"traceRetainDays,omitempty"` // 0 → default (3)
}

// DataRetention controls automatic cleanup of generated data.
type DataRetention struct {
	RunsRetainDays   int `json:"runsRetainDays,omitempty"`   // 0 → default (14)
	RunsKeepPerName  int `json:"runsKeepPerName,omitempty"`  // 0 → default (50)
}

// AIProvider represents one named AI backend connection.
type AIProvider struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
}

// DocLink is a quick-access reference to a local document file.
type DocLink struct {
	Path  string `json:"path"`
	Label string `json:"label"`
}

type Config struct {
	Directories      []Directory           `json:"directories"`
	Exports          []ExportVar           `json:"exports"`
	OverrideMeta     map[string]ScriptMeta `json:"overrideMeta"`
	ZshrcAnnotations map[string]string     `json:"zshrcAnnotations,omitempty"`
	AI               AIConfig              `json:"ai,omitempty"`
	AutoLaunch       bool                  `json:"autoLaunch,omitempty"`
	TraceLogDir      string                `json:"traceLogDir,omitempty"`  // AI call log dir, default: <configDir>/Hatch/traces
	SkillsDir        string                `json:"skillsDir,omitempty"`    // skills master dir, default: <configDir>/Hatch/skills
	DocLinks         []DocLink             `json:"docLinks,omitempty"`
	AIProviders      []AIProvider          `json:"aiProviders,omitempty"`
	ActiveProviderID string                `json:"activeProviderId,omitempty"`
	PromptTemplates  map[string]string     `json:"promptTemplates,omitempty"` // v1.2: op -> override system prompt
	Retention        DataRetention         `json:"retention,omitempty"`       // v1.3+: data cleanup policy
}

func defaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Hatch", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{OverrideMeta: map[string]ScriptMeta{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.OverrideMeta == nil {
		cfg.OverrideMeta = map[string]ScriptMeta{}
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := defaultPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
