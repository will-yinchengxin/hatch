# Hatch — Project Reference

Hatch is a **macOS script management dashboard** built with Wails v2 (Go backend + WKWebView frontend).
Target: Apple Silicon, no App Store, no Apple Developer account.

## Tech Stack

| Layer | Technology |
|---|---|
| Desktop shell | Wails v2 — embeds Go binary + WKWebView |
| Backend | Go 1.21+ |
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
├── app.go                    # All Go methods bound to the frontend (the "API")
├── main.go                   # Wails entry point
├── internal/
│   ├── config/config.go      # Data types + JSON config load/save
│   ├── scanner/              # scanner.go (ScanDir) + parser.go (annotations, language)
│   ├── executor/             # oneshot.go / service.go / stream.go / cmd.go
│   ├── autostart/            # macOS LaunchAgent install/remove
│   ├── store/store.go        # JSON-backed script metadata cache
│   └── zshrc/                # manager.go (block r/w) + parser.go (structured parse)
└── frontend/
    ├── index.html            # All UI — Alpine.js declarative markup
    └── src/
        ├── main.js           # Alpine component: all state + methods
        └── style.css         # Full custom CSS
```

---

## Core Data Model (`internal/config/config.go`)

```go
type ScriptMeta struct {
    Path      string; Name string; Desc string
    Type      ScriptType      // "oneshot" | "service"
    Language  string          // "shell" | "python" | "ruby" | "node" | "binary"
    Commands  []string        // service sub-commands
    PidFile   string          // default: /tmp/hatch/<name>.pid
    Args      []ArgDef        // nil = show input; [] = hide input
    Presets   []CommandPreset // {Name, Args} — rendered as 🔖 buttons
    Tags      []string; LastRunAt int64; UpdatedAt int64
}

type Config struct {
    Directories      []Directory
    Exports          []ExportVar
    OverrideMeta     map[string]ScriptMeta
    ZshrcAnnotations map[string]string
    AI               AIConfig
    TraceLogDir      string    // default: ~/Library/Application Support/Hatch/logs/
    SkillsDir        string    // default: ~/.claude/skills/
    AutoLaunch       bool
}

type AIConfig struct {
    Enabled bool; BaseURL string; APIKey string; Model string
}
```

`Args nil` → show input box. `Args []ArgDef{}` → hide input box. These are different.
`mergeOverride` only replaces non-zero fields — use `DeleteScriptMeta` to fully reset auto-detection.

---

## App.go — Bound Methods

| Method | Purpose |
|---|---|
| `GetConfig()` | Full config |
| `ScanAll()` | Rescan all dirs → `[]ScriptMeta` |
| `GetScriptsByDir(path)` | Scan single dir |
| `AddDirectory / RemoveDirectory / UpdateDirectoryExport` | Dir management |
| `SaveScriptMeta / DeleteScriptMeta` | Override CRUD |
| `RunScript(id, path, args)` | Launch oneshot; events via EventsEmit |
| `ServiceAction(path, action, args)` | start / stop / restart |
| `GetAllServiceStatuses()` | `map[path]"running"\|"stopped"` |
| `SaveExports / CheckZshrc / ParseZshrc` | Shell management |
| `SaveZshrcAnnotation(key, note)` | Persist zshrc entry note |
| `SuggestScriptMeta / SuggestZshrcAnnotation` | AI fill (no auto-save) |
| `SuggestZshrcOrganization(mode)` | AI preview → `{summary, content}` |
| `ApplyZshrcOrganization(content)` | Backup + write zshrc |
| `GetAutoLaunchEnabled / SetAutoLaunchEnabled` | Login autostart |
| `GetClaudeCodeStatus()` | Claude CLI + VPN + usage from `~/.cache/will-statusline/last.json` |
| `GetCodexStatus()` | Codex CLI + VPN + usage from `~/.codex/sessions/` JSONL |
| `ListSkills() / SyncSkill / CreateSkill` | Skills management |
| `GetAppDataPaths() / SaveAppDataPaths` | Trace log dir + skills dir config |
| `GetTraceLog(date)` | Read trace JSONL for date |

**Not yet in wailsjs bindings** — call via `window['go']['main']['App']['Method']()`:
`SuggestZshrcOrganization`, `ApplyZshrcOrganization`, `GetAutoLaunchEnabled`, `SetAutoLaunchEnabled`,
`GetClaudeCodeStatus`, `GetCodexStatus`, `ListSkills`, `SyncSkill`, `CreateSkill`,
`GetAppDataPaths`, `SaveAppDataPaths`, `GetTraceLog`.

---

## Real-time Events (Go → JS via EventsEmit)

| Event | Payload | When |
|---|---|---|
| `executor:log` | `{id, line, stream}` | Every output line |
| `executor:done` | `{id, exitCode}` | Oneshot completed |
| `executor:status` | `{id, status}` | Service state change |

`id` = script's absolute `path`.

---

## Tab Navigation

```
tab = 'scripts' | 'settings' | 'ai'

scripts + !selectedDir  →  Directory card grid + workspace-settings accordion
scripts + selectedDir   →  Script list for that dir
settings               →  (removed — merged into scripts home as accordion)
ai + aiSubTab          →  概览 | 连接 | Skills | 工具
editTarget != null     →  Edit drawer overlay
```

**Settings is now an accordion on the scripts home page** (not a separate tab). Sections: ⚙ 环境与路径配置 · 📁 目录管理 · 🔑 环境变量 · 📄 Shell 配置解析.

---

## AI Module

Top-level `✦ AI` tab with left sidebar sub-navigation:

| Sub-tab | Content |
|---|---|
| 📖 概览 | Capability summary, quick usage tips |
| 🔌 连接 | OpenAI-compatible model config + trace log viewer |
| 📚 Skills | List/sync/create Skills for Claude Code & Codex |
| ⚡ 工具 | Claude Code status card + Codex status card |

Hero stats: AI状态 · 模型 · Skills count card (purple, shows claude/codex breakdown).

**Claude Code status** (`GetClaudeCodeStatus`): reads `~/.cache/will-statusline/last.json` for rate limit pcts + plan type. Falls back to `zsh -l -c "command -v claude"` for CLI detection (GUI apps don't inherit user PATH).

**Codex status** (`GetCodexStatus`): reads most-recent JSONL in `~/.codex/sessions/`.

**Skills** (`ListSkills`): scans `SkillsDir` (default `~/.claude/skills/`) for `SKILL.md` files, returns name/description/provider info. `SyncSkill` copies a skill into the Codex skills dir. `CreateSkill` writes a new SKILL.md from AI suggestions.

**Trace log**: every AI call appends a JSONL record to `TraceLogDir/<YYYY-MM-DD>.jsonl`.

---

## Runtime Data Paths

| Purpose | Path |
|---|---|
| Config | `~/Library/Application Support/Hatch/config.json` |
| Script cache | `~/Library/Application Support/Hatch/scripts.json` |
| Trace logs | `~/Library/Application Support/Hatch/logs/<date>.jsonl` (configurable) |
| PID files | `/tmp/hatch/<name>.pid` |
| Zshrc managed block | `~/.zshrc` (Hatch block only) |
| Login autostart | `~/Library/LaunchAgents/com.hatch.app.plist` |
| Claude usage cache | `~/.cache/will-statusline/last.json` |
| Codex sessions | `~/.codex/sessions/` |

---

## Detailed Reference

- [Frontend architecture, Alpine state, UI design tokens, CSS rules](.claude/docs/frontend.md)
- [Internals: scanner, executor, zshrc manager](.claude/docs/internals.md)
- [Known gotchas & Pico CSS integration notes](.claude/docs/gotchas.md)
- [Frontend skill (component patterns, CSS classes)](.claude/skills/hatch-frontend/SKILL.md)
