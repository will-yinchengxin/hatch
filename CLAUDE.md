# Hatch — Project Reference

Hatch is a **macOS script management dashboard** built with Wails v2 (Go backend + WKWebView frontend).
Target: Apple Silicon, no App Store, no Apple Developer account.

## Tech Stack

| Layer | Technology |
|---|---|
| Desktop shell | Wails v2 — embeds Go binary + WKWebView |
| Backend | Go 1.23+ |
| Frontend | Alpine.js + Pico.css v2 + Vite |
| Styling | Custom CSS variables, Anthropic-inspired warm light theme |
| Config | JSON at `~/Library/Application Support/Hatch/config.json` |

Build: `$(go env GOPATH)/bin/wails build -platform darwin/arm64` · Dev: `$(go env GOPATH)/bin/wails dev`
Tests: `go test ./...` (after `app.go` changes) or `go test ./internal/... -v`
Make targets: `dev / test / build / build-clean / sign / dmg / release / install / open`

---

## Repository Layout

```
hatch/
├── app.go                  # App struct + startup; method set is split below.
├── app_scripts.go          # Directory CRUD + scanning + SaveScriptMeta
├── app_executor.go         # RunScript / ServiceAction / Run history / scriptEnv
├── app_zshrc.go            # zshrc parse / annotate / AI organize
├── app_ai.go               # AI providers / Suggest* / callAIOp / trace emit
├── app_observability.go    # Claude/Codex status / trace persistence / macOS notify
├── app_files.go            # Script files / AppDataPaths / AutoLaunch / DocLinks
├── app_v12.go              # v1.2: scheduler / health-check / prompt templates / stream
├── app_v13.go              # v1.3: PATH health / backup-restore / pre-script chain
├── app_datadir.go          # Data-dir inventory + cleanup + retention + Skill AI
├── skills.go               # SKILL.md store + claude/codex sync
├── main.go                 # Wails entry
├── internal/
│   ├── config/             # Config + ScriptMeta + Save/Load
│   ├── scanner/            # ScanDir + @hatch.* annotation parser
│   ├── executor/           # oneshot.go / service.go / stream.go / cmd.go / history.go
│   ├── scheduler/          # tiny @every / @daily / @hourly scheduler (no deps)
│   ├── autostart/          # macOS LaunchAgent
│   ├── store/              # scripts.json cache
│   └── zshrc/              # managed block + parser
└── frontend/
    ├── index.html          # Alpine.js declarative markup (all UI)
    └── src/
        ├── main.js         # Alpine component: state + methods
        └── style.css       # Full custom CSS (sections per feature)
```

---

## Core Data Model (`internal/config/config.go`)

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
}

type DataRetention struct { RunsRetainDays, RunsKeepPerName int }
type AIConfig       struct { Enabled bool; BaseURL, APIKey, Model string; TraceRetainDays int }
```

`Args nil` → show input; `Args []ArgDef{}` → hide. `mergeOverride` only replaces non-zero fields — use `DeleteScriptMeta` to fully reset auto-detection.

---

## App.go — Bound Methods

| Method | Purpose |
|---|---|
| `GetConfig / ScanAll / GetScriptsByDir` | Config + scan |
| `AddDirectory / RemoveDirectory / UpdateDirectoryExport` | Directory CRUD |
| `SaveScriptMeta / DeleteScriptMeta` | Override CRUD (re-applies schedules) |
| `RunScript(id, path, args)` | Oneshot launch (pre-scripts auto-chained if set) |
| `ServiceAction(path, action, args)` | start / stop / restart with env |
| `GetAllServiceStatuses()` | path → "running" \| "stopped" |
| `ListRuns / GetRunLog / DeleteRun` | v1.1 run-history archive |
| `SaveExports / CheckZshrc / ParseZshrc / SaveZshrcAnnotation` | Shell config |
| `SuggestZshrcOrganization / ApplyZshrcOrganization` | AI zshrc rewrite preview + write |
| `GetAutoLaunchEnabled / SetAutoLaunchEnabled` | Login autostart |
| `GetClaudeCodeStatus / GetCodexUsageStatus` | Status cards |
| `GetSkills / GetSkillContent / SaveSkillContent / DeleteSkill / GetSkillsStats` | Skill CRUD |
| `SetSkillSync(name, target, enabled) / ImportSkillsFrom / CreateSkillTemplate` | Skill sync + import |
| `SuggestSkillContent / SuggestSkillFromDescription` | v1.3 Skill AI optimize |
| `GetAppDataPaths / SaveAppDataPaths` | Trace + Skills dir config |
| `GetAITraces / ClearAITraces / GetAITraceStorageInfo` | AI trace inspect |
| `SaveAIConfig / GetAIProviders / SaveAIProviders / TestAIConfig` | Provider mgmt |
| `SuggestScriptMeta / SuggestZshrcAnnotation / SuggestScriptContent` | AI Suggest |
| `SuggestScriptContentStream(id, ...)` | v1.2 streaming (emits `ai:stream`) |
| `GetPromptTemplates / SavePromptTemplates / GetDefaultPrompts` | v1.2 prompt overrides |
| `ApplyAllSchedules / ListSchedules / RemoveSchedule` | v1.2 scheduler control |
| `CheckScriptHealth(path) / CheckCommandAvailable(name)` | v1.2 dep + health probe |
| `CheckPathHealth / ListAliasInventory` | v1.3 PATH + alias inspection |
| `ExportBackup(dir) / ImportBackup(zipPath)` | v1.3 zip backup |
| `RunWithPreScripts(id, path, args, pre)` | v1.3 explicit chain runner |
| `GetAppDataInventory / CleanupDataDir(key) / GetRetentionPolicy / SaveRetentionPolicy` | v1.3 data dashboard |
| `ListRecentRunsSorted(limit)` | flat newest-first run list |
| `GetClaudeSettings / SaveClaudeSettings / GetCodexConfig / SaveCodexConfig` | Edit ~/.claude / ~/.codex configs |
| `GetDocLinks / SaveDocLinks / OpenFile / ReadScriptFile / WriteScriptFile / CreateScriptFile` | File I/O |

Methods added after the last `wails dev`/`wails build` must be called via `window['go']['main']['App']['Method']()` until bindings regenerate.

---

## Real-time Events (Go → JS via EventsEmit)

| Event | Payload | When |
|---|---|---|
| `executor:log` | `{id, line, stream}` | Every stdout/stderr line |
| `executor:done` | `{id, exitCode}` | Oneshot completed |
| `executor:status` | `{id, status}` | running / stopped / error |
| `ai:trace` | `AICallTrace` | Each AI call recorded |
| `ai:stream` | `{id, chunk, done, error}` | Streaming Suggest output |

`id` = script's absolute `path` for executor events.

---

## Tab Navigation

```
tab = 'scripts' | 'ai'

scripts + !selectedDir  →  Directory card grid + workspace-settings accordion
scripts + selectedDir   →  Script list for that dir
ai + aiSubTab           →  概览 | 连接 | Skills | 工具 | Prompt | 数据
editTarget   != null    →  Edit drawer (env / notify / schedule / requires / health / preScripts)
historyTarget != null   →  Run-history modal
quickAIOpen             →  Header popover: model switch + recent traces
```

Workspace settings live on the scripts home page accordion: ⚙ 环境与路径配置 · 📁 目录管理 · 🔑 环境变量 · 📄 Shell 配置解析.

---

## AI Module

| Sub-tab | Content |
|---|---|
| 📖 概览 | Capability summary |
| 🔌 连接 | Providers + trace log viewer |
| 📚 Skills | List/edit/sync/create + ✦ AI optimize bar in editor |
| ⚡ 工具 | Claude Code + Codex status cards |
| 📝 Prompt | Per-op system-prompt override editor (shows defaults inline) |
| 📦 数据 | Data-dir inventory + retention policy + per-dir 📂/🗑 |

**Header quick-AI popover** (next to ⏻ 自启): badge shows active provider `✦ name · model`; click for model switcher + last 8 traces.

**Trace log**: every AI call appends a JSONL record to `TraceLogDir/<YYYY-MM-DD>.jsonl`. Pruned by `AI.TraceRetainDays`.

---

## Runtime Data Paths

| Purpose | Path |
|---|---|
| Config | `~/Library/Application Support/Hatch/config.json` |
| Script cache | `~/Library/Application Support/Hatch/scripts.json` |
| AI trace logs | `~/Library/Application Support/Hatch/traces/<date>.jsonl` (configurable) |
| Run history | `~/Library/Application Support/Hatch/runs/<script>/<id>.{log,json}` |
| Skills master | `~/Library/Application Support/Hatch/skills/` (default; configurable) |
| PID files | `/tmp/hatch/<name>.pid` |
| Zshrc managed block | `~/.zshrc` (Hatch block only) |
| Login autostart | `~/Library/LaunchAgents/com.hatch.app.plist` |
| Claude usage cache | `~/.cache/will-statusline/last.json` |
| Codex sessions | `~/.codex/sessions/` |
| Backup exports | `~/Desktop/hatch-backup-<ts>.zip` (default) |

Retention runs on startup + hourly ticker (`ApplyRetention`): trace days, runs days, runs-per-script cap.

---

## Detailed Reference

- [Frontend architecture, Alpine state, UI design tokens, CSS rules](.claude/docs/frontend.md)
- [Internals: scanner, executor, zshrc, scheduler, history](.claude/docs/internals.md)
- [Known gotchas & Pico CSS integration notes](.claude/docs/gotchas.md)
- [Frontend skill (component patterns, CSS classes)](.claude/skills/hatch-frontend/SKILL.md)
- [Upgrade log (v1.1 → v1.3)](UPGRADE_LOG.md)
