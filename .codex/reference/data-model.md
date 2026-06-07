# Hatch — Core Data Model

Source: `internal/config/config.go`.

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
    Env          map[string]string // per-script env, merged over os.Environ
    NotifyOnDone bool              // macOS notification banner
    Schedule     string            // @every / @daily / @hourly
    Requires     []string          // required CLI commands
    HealthURL    string            // GET health target
    PreScripts   []string          // sequential pre-script chain
    LastRunAt, UpdatedAt int64
}
```

`Args nil` means show an args input. `Args []ArgDef{}` means hide the input. `mergeOverride` only replaces non-zero fields; use `DeleteScriptMeta` to fully reset auto-detection.

## Config

```go
type Config struct {
    Directories      []Directory
    Exports          []ExportVar
    OverrideMeta     map[string]ScriptMeta
    ZshrcAnnotations map[string]string
    AI               AIConfig
    AIProviders      []AIProvider
    ActiveProviderID string
    AutoLaunch       bool
    TraceLogDir      string
    SkillsDir        string
    DocLinks         []DocLink
    PromptTemplates  map[string]string
    Retention        DataRetention
    VPNAppDir        string
    VPNBackupDir     string
    MCPDir           string
    MCPServers       []MCPServer
    MCPSeeded        bool        // true after first MCP directory discovery
    Hooks            []HookDef
    HooksSeeded      bool        // true after the default hello-test hook is seeded once
}

type DataRetention struct { RunsRetainDays, RunsKeepPerName int }
type AIConfig       struct { Enabled bool; BaseURL, APIKey, Model string; TraceRetainDays int }
```

## MCP Types

`MCPDir` defaults to `~/MyHub/mcp`. `MCPServers` is Hatch's registry; `MCPSeeded` records that first directory discovery has run. Syncing renders entries into supported global config files, but Hatch does not keep MCP services running. Deleting a synced registry entry also removes its Hatch-managed target config entries.

```go
type MCPServer struct {
    ID, Name, Description string
    SourceType            string // local | git | npm | remote
    SourceURL, LocalPath, PackageName string
    Transport             string // stdio | http
    Command               string
    Args                  []string
    Cwd, URL              string
    Env, HTTPHeaders      map[string]string
    BearerTokenEnvVar     string
    StartupTimeoutSec, ToolTimeoutSec int
    Tags                  []string
    EnabledInClaude, EnabledInCodex bool
    LastTestResults       map[string]MCPTestResult
    UpdatedAt             int64
}

type MCPTestResult struct {
    Target, Status, Summary, Detail string
    ExitCode                         int
    CheckedAt                        int64
    Mode                            string // connection | config
}
```

## HookDef

One `command`-type hook authored once in Hatch and synced to supported targets. Persisted in `config.json` under `Config.Hooks`.

```go
type HookDef struct {
    ID              string   // stable id; also written as "hatchId" into target files
    Name            string
    Description     string
    Event           string   // PreToolUse | PostToolUse | SessionStart | Stop | ...
    Matcher         string   // tool/source filter; empty = match all
    Command         string   // shell command or executable (type is always "command")
    Args            []string // exec-form args; when set, no shell
    Timeout         int      // seconds; 0 = tool default
    StatusMessage   string
    EnabledInClaude bool
    EnabledInCodex  bool
    UpdatedAt       int64
}
```

Both targets share the schema `{"hooks": {"<Event>": [{"matcher", "hooks": [{"type":"command", ...}]}]}}`. Hatch handlers carry a `hatchId` field; sync removes/re-inserts only marked handlers and prunes empty matcher groups/events. Event support per target lives in `app_hooks.go` `hookEvents` (`Notification`/`SessionEnd` are only supported by the compatible target).

## VPN Types

`VPNStatus.ProxyGroups` and `TUNExcludedCIDRs` are live data from the Clash core API via Unix socket (`external-controller-unix`, default `/tmp/verge/verge-mihomo.sock`). `profiles.yaml selected[]` is only fallback data because it can be stale after in-session node switches.

```go
type VPNStatus struct {
    SystemProxyEnabled bool
    TUNEnabled         bool
    TUNInterface       string
    TUNExcludedCIDRs   []string
    CurrentGroup       string
    CurrentNode        string
    ProxyGroups        []VPNProxyGroup
    ExpireAt, Upload, Download, Total int64
}

type VPNProxyGroup struct {
    Name string
    Node string
}

type VPNEnhanceData struct {
    PrependRules    []string
    Hosts           map[string]string
    DNSFakeIPFilter []string
    HasData         bool
    SourceFiles     []string
}
```
