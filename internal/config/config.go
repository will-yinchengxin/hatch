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
	RunsRetainDays  int `json:"runsRetainDays,omitempty"`  // 0 → default (14)
	RunsKeepPerName int `json:"runsKeepPerName,omitempty"` // 0 → default (50)
}

// MemoSettings controls where memo attachments are copied and when Hatch may
// clean up managed attachment files.
type MemoSettings struct {
	FilesDir    string `json:"filesDir,omitempty"`
	RetainDays  int    `json:"retainDays,omitempty"`  // 0 → default (90)
	AutoCleanup bool   `json:"autoCleanup,omitempty"` // false keeps records but skips file pruning
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

// ClaudeCodePreset is a named combination of model + effort level for Claude Code.
type ClaudeCodePreset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	EffortLevel string `json:"effortLevel"`
}

// CodexPreset is a named combination of model + reasoning effort for Codex.
type CodexPreset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Model       string `json:"model"`
	EffortLevel string `json:"effortLevel"`
}

// AgentResourceSelection describes whether one Hatch-managed resource should be
// enabled for Claude Code and/or Codex in an AI workspace profile.
type AgentResourceSelection struct {
	ID     string `json:"id"`
	Claude bool   `json:"claude"`
	Codex  bool   `json:"codex"`
}

// AIWorkspaceProfile is a manually applied, project-scoped combination of
// agent model settings and Hatch-managed tools. It intentionally does not
// store credentials or raw target configuration.
type AIWorkspaceProfile struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	RootPath     string                   `json:"rootPath"`
	ClaudeModel  string                   `json:"claudeModel,omitempty"`
	ClaudeEffort string                   `json:"claudeEffort,omitempty"`
	CodexModel   string                   `json:"codexModel,omitempty"`
	CodexEffort  string                   `json:"codexEffort,omitempty"`
	CodexGoals   bool                     `json:"codexGoals"`
	Skills       []AgentResourceSelection `json:"skills,omitempty"`
	MCPServers   []AgentResourceSelection `json:"mcpServers,omitempty"`
	Hooks        []AgentResourceSelection `json:"hooks,omitempty"`
	UpdatedAt    int64                    `json:"updatedAt"`
}

// MCPServer represents one MCP server managed by Hatch.
type MCPServer struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description,omitempty"`
	SourceType        string                   `json:"sourceType"` // local | git | npm | remote
	SourceURL         string                   `json:"sourceUrl,omitempty"`
	LocalPath         string                   `json:"localPath,omitempty"`
	PackageName       string                   `json:"packageName,omitempty"`
	Transport         string                   `json:"transport"` // stdio | http
	Command           string                   `json:"command,omitempty"`
	Args              []string                 `json:"args,omitempty"`
	Cwd               string                   `json:"cwd,omitempty"`
	URL               string                   `json:"url,omitempty"`
	Env               map[string]string        `json:"env,omitempty"`
	HTTPHeaders       map[string]string        `json:"httpHeaders,omitempty"`
	BearerTokenEnvVar string                   `json:"bearerTokenEnvVar,omitempty"`
	StartupTimeoutSec int                      `json:"startupTimeoutSec,omitempty"`
	ToolTimeoutSec    int                      `json:"toolTimeoutSec,omitempty"`
	Tags              []string                 `json:"tags,omitempty"`
	EnabledInClaude   bool                     `json:"enabledInClaude,omitempty"`
	EnabledInCodex    bool                     `json:"enabledInCodex,omitempty"`
	LastTestResults   map[string]MCPTestResult `json:"lastTestResults,omitempty"`
	UpdatedAt         int64                    `json:"updatedAt,omitempty"`
}

// MCPTestResult stores the last connection/config check for one MCP target.
type MCPTestResult struct {
	Target    string `json:"target"`
	Status    string `json:"status"` // success | failed | unknown
	Summary   string `json:"summary"`
	Detail    string `json:"detail,omitempty"`
	ExitCode  int    `json:"exitCode,omitempty"`
	CheckedAt int64  `json:"checkedAt,omitempty"`
	Mode      string `json:"mode,omitempty"` // connection | config
}

// HookDef represents one command hook managed by Hatch. A single definition is
// authored once and synced to Claude Code (~/.claude/settings.json) and/or Codex
// (~/.codex/hooks.json), where Hatch-owned handlers carry a "hatchId" marker.
type HookDef struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Event           string   `json:"event"`             // PreToolUse | PostToolUse | SessionStart | ...
	Matcher         string   `json:"matcher,omitempty"` // tool/source filter; empty = match all
	Command         string   `json:"command"`           // shell command or executable (type is always "command")
	Args            []string `json:"args,omitempty"`    // exec-form arguments; when set, no shell
	Timeout         int      `json:"timeout,omitempty"` // seconds; 0 = tool default
	StatusMessage   string   `json:"statusMessage,omitempty"`
	EnabledInClaude bool     `json:"enabledInClaude,omitempty"`
	EnabledInCodex  bool     `json:"enabledInCodex,omitempty"`
	UpdatedAt       int64    `json:"updatedAt,omitempty"`
}

type Config struct {
	Directories             []Directory           `json:"directories"`
	Exports                 []ExportVar           `json:"exports"`
	OverrideMeta            map[string]ScriptMeta `json:"overrideMeta"`
	ZshrcAnnotations        map[string]string     `json:"zshrcAnnotations,omitempty"`
	AI                      AIConfig              `json:"ai,omitempty"`
	AutoLaunch              bool                  `json:"autoLaunch,omitempty"`
	TraceLogDir             string                `json:"traceLogDir,omitempty"` // AI call log dir, default: <configDir>/Hatch/traces
	SkillsDir               string                `json:"skillsDir,omitempty"`   // skills master dir, default: <configDir>/Hatch/skills
	DocLinks                []DocLink             `json:"docLinks,omitempty"`
	AIProviders             []AIProvider          `json:"aiProviders,omitempty"`
	ActiveProviderID        string                `json:"activeProviderId,omitempty"`
	PromptTemplates         map[string]string     `json:"promptTemplates,omitempty"`         // v1.2: op -> override system prompt
	Retention               DataRetention         `json:"retention,omitempty"`               // v1.3+: data cleanup policy
	Memo                    MemoSettings          `json:"memo,omitempty"`                    // v1.4+: memo attachment storage policy
	VPNAppDir               string                `json:"vpnAppDir,omitempty"`               // Clash Verge Rev data dir (default: auto-detect)
	VPNBackupDir            string                `json:"vpnBackupDir,omitempty"`            // VPN backup dir (default: Hatch/vpn-backups)
	ClaudeCodePresets       []ClaudeCodePreset    `json:"claudeCodePresets,omitempty"`       // saved model+effort presets
	ClaudeCodePresetsSeeded bool                  `json:"claudeCodePresetsSeeded,omitempty"` // true after first default preset seed
	CodexPresets            []CodexPreset         `json:"codexPresets,omitempty"`            // saved model+effort presets
	CodexPresetsSeeded      bool                  `json:"codexPresetsSeeded,omitempty"`      // true after first default preset seed
	ScriptOrders            map[string][]string   `json:"scriptOrders,omitempty"`            // dirPath -> ordered script paths
	MCPDir                  string                `json:"mcpDir,omitempty"`                  // MCP master dir, default: ~/MyHub/mcp if present
	MCPServers              []MCPServer           `json:"mcpServers,omitempty"`              // MCP registry managed by Hatch
	MCPSeeded               bool                  `json:"mcpSeeded,omitempty"`               // true after first MCP directory discovery
	Hooks                   []HookDef             `json:"hooks,omitempty"`                   // command hook registry managed by Hatch
	HooksSeeded             bool                  `json:"hooksSeeded,omitempty"`             // true after the default hook is seeded once
	AIWorkspaceProfiles     []AIWorkspaceProfile  `json:"aiWorkspaceProfiles,omitempty"`     // manually applied project AI profiles
	LastAppliedWorkspaceID  string                `json:"lastAppliedWorkspaceId,omitempty"`  // last successfully applied profile
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
