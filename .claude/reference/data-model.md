# Hatch — Core Data Model

Source: `internal/config/config.go`

## ScriptMeta

```go
type ScriptMeta struct {
    Path, Name, Desc, Language string
    Type         ScriptType        // "oneshot" | "service"
    Commands     []string          // service sub-commands
    PidFile      string            // default /tmp/hatch/<name>.pid
    Args         []ArgDef          // nil = show input; [] = hide input
    Presets      []CommandPreset   // named arg invocations
    Tags         []string
    Env          map[string]string // v1.1 per-script env (merged over os.Environ)
    NotifyOnDone bool              // v1.1 macOS osascript banner
    Schedule     string            // v1.2 @every / @daily / @hourly
    Requires     []string          // v1.2 required CLI commands
    HealthURL    string            // v1.2 GET /healthz target
    PreScripts   []string          // v1.3 sequential pre-script chain
    LastRunAt, UpdatedAt int64
}
```

`Args nil` → show input box; `Args []ArgDef{}` → hide entirely.  
`mergeOverride` only replaces non-zero fields — use `DeleteScriptMeta` to fully reset auto-detection.

## Config

```go
type Config struct {
    Directories      []Directory
    Exports          []ExportVar
    OverrideMeta     map[string]ScriptMeta
    ZshrcAnnotations map[string]string
    AI               AIConfig         // legacy single-provider (kept in sync)
    AIProviders      []AIProvider     // multi-provider list
    ActiveProviderID string
    AutoLaunch       bool
    TraceLogDir      string           // default <cfgDir>/Hatch/traces
    SkillsDir        string           // default <cfgDir>/Hatch/skills
    DocLinks         []DocLink
    PromptTemplates  map[string]string // v1.2 op → override system prompt
    Retention        DataRetention     // v1.3 cleanup policy
    VPNAppDir        string            // Clash Verge Rev data dir (default: auto-detect)
    VPNBackupDir     string            // VPN backup dir (default: Hatch/vpn-backups)
}

type DataRetention struct { RunsRetainDays, RunsKeepPerName int }
type AIConfig       struct { Enabled bool; BaseURL, APIKey, Model string; TraceRetainDays int }
```
