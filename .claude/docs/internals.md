# Hatch — Internal Packages Reference

## Scanner (`internal/scanner/`)

Three-tier detection per file, applied in order:

**Tier 1 — `@hatch.*` annotations** (first 50 lines, `#` `//` `--` comment styles)
```sh
# @hatch.name: My Tool
# @hatch.desc: Does something
# @hatch.type: service
# @hatch.commands: start, stop, restart
# @hatch.pidfile: /tmp/mytool.pid
# @hatch.tags: infra, deploy
# @hatch.schedule: @every 30m            (v1.2)
# @hatch.requires: kubectl, jq           (v1.2)
# @hatch.health: http://localhost:8080/healthz   (v1.2)
```

**Tier 2a — Service pattern detection** (first 200 lines) — requires both `start` AND `stop` as case arms / quoted comparisons.

**Tier 2b — `--help` probe** (2s timeout) — first non-empty output line becomes `Desc`.

**Tier 3 — Defaults** — `type=oneshot`, `name=filename`, `pidFile=/tmp/hatch/<name>.pid`.

**Override layer** — `config.OverrideMeta[path]` merged on top. Non-zero fields only win; use `DeleteScriptMeta` to fully reset.

---

## Executor (`internal/executor/`)

### Oneshot (`oneshot.go`)
- `RunOneshot` is a thin wrapper around `RunOneshotWithEnv(id, path, args, env, em, hook)`.
- Runs as `/bin/zsh -l -c 'source ~/.zshrc; exec <script> <args>'`
- Explicit `source ~/.zshrc` required — Wails GUI apps don't get interactive shells
- stdout/stderr streamed line-by-line via `Emitter.Log`
- First log line: `[hatch] command: <quoted invocation>` — env lines `[hatch] env: K=V` follow if any
- `Emitter.Done` fires with exit code
- `mergeEnv(map)` = `os.Environ()` + overrides (override wins); only applied when map non-empty
- `NotifyHook(exitCode, durationMs)` invoked after Done — used for macOS osascript banners
- Every run is archived to `runs/<name>/<id>.{log,json}` via `runRecorder` (see History below)

### Service (`service.go` — `ServiceManager`)
- `Start/Restart` are wrappers around `StartWithEnv/RestartWithEnv(args, env)` (v1.1).
- Start: `Setpgid: true` (own process group), PID to `pidFile`, optional env via `mergeEnv`
- Stop: SIGTERM to process group + PID file cleanup
- Status: `kill -0 <pid>` probe; stale PIDs auto-removed
- Restart: Stop (ignore error) → Start
- App startup: `restoreServices()` rebuilds and probes all stored services

### History (`history.go`, v1.1)
- `runRecorder` writes each line to `<runs>/<basename>/<id>.log` while the script runs (prefix `[out]`/`[err]`); finalize writes `<id>.json` with `RunRecord{Script,Args,StartedAt,EndedAt,ExitCode,DurationMs,StdoutTail,LogPath}`.
- `historyRoot()` (exported as `HistoryRoot()`) = `<cfgDir>/Hatch/runs`.
- `ListRuns(scriptPath)` returns newest-first records for one script.
- `PruneAllHistory(keepPerScript)` LRU per-script cap (default 50).
- `PruneHistoryOlderThan(retainDays)` time-based prune (default 14 d).
- Both pruners called from `App.ApplyRetention()` — startup + hourly ticker.

### Scheduler (`internal/scheduler/`, v1.2)
- Zero external deps; `ParseSchedule(spec)` returns a `nextFireFn(now) → time.Time`.
- Supports `@every Ns/Nm/Nh`, `@daily HH:MM`, `@hourly :MM`.
- Each `Schedule(id, spec, run)` spawns a goroutine that fires `run()` at each computed time; `panic` recovered.
- `Remove(id)` closes the goroutine's stop chan.
- `App.ApplyAllSchedules()` rebuilds full set on startup (after 500ms delay so store has data) and after every `SaveScriptMeta`.

### CLI Detection (`app.go` `findCLI`)
```go
func findCLI(name string) string {
    if path, err := exec.LookPath(name); err == nil { return path }
    out, _ := exec.Command("/bin/zsh", "-l", "-c", "command -v "+name).Output()
    return strings.TrimSpace(string(out))
}
```
Required for tools in `~/.local/bin` etc. which aren't in the GUI process PATH.

---

## Zshrc Manager (`internal/zshrc/manager.go`)

Managed block format:
```sh
# >>> Hatch managed block >>>
export MY_TOOLS="/path"
export PATH=$MY_TOOLS:$PATH
# <<< Hatch managed block <<<
```

- `WriteBlock` — atomically replaces block, leaves everything else untouched
- `RemoveBlock` — called when exports list is empty
- Keys sorted alphabetically for stable diffs
- `$` in export values escaped before regex substitution
- PATH line auto-appended with all exported dir vars
- `Backup(path)` — copies to `<path>.hatch-backup-<YYYYMMDD-HHMMSS>`
- `WriteContent(path, content)` — replaces entire file, ensures trailing newline

---

## AI Calls (`app.go`)

- HTTP timeout: 45s
- `BaseURL` + `/chat/completions` (appended only if missing)
- Script analysis: truncates content at 12,000 bytes
- Zshrc reorganization: truncates at 24,000 bytes, returns `{summary, content}`
- All AI results populate UI for user confirmation — never auto-saved

### Trace Log
Every AI call appends JSONL to `TraceLogDir/<YYYY-MM-DD>.jsonl` (default: `~/Library/Application Support/Hatch/traces/`). Pruned by `AI.TraceRetainDays` (default 3) on startup + each `recordTrace`. Read via `GetAITraces` (newest first across retain window) or `GetAITraceStorageInfo` for size/count.

### Prompt Templates (v1.2)
- `Config.PromptTemplates map[string]string` — key = op name (e.g. `SuggestScriptMeta`).
- `callAIOp(op, ...)` calls `effectivePrompt(op, default)` so user overrides automatically take effect for every Suggest.
- `GetDefaultPrompts()` returns the canonical built-in prompts; the UI shows them inline so users know what they're overriding.

### Streaming (v1.2)
- `SuggestScriptContentStream(id, ...)` emits `ai:stream` events with `{id, chunk, done, error}`.
- Current implementation calls the blocking endpoint and emits the result as one chunk — protocol fixed, true SSE deferred.

### Skill AI (v1.3+)
- `SuggestSkillContent(name, current, instruction)` — frontmatter-aware system prompt. Forbids markdown fences in output. New op key auto-shows up in the Prompt sub-tab.

---

## Config (`internal/config/config.go`)

`os.UserConfigDir()` → `~/Library/Application Support` on macOS.
Both `config.json` and `scripts.json` in same dir.
Store falls back to `$TMPDIR/hatch/scripts.json` if `UserConfigDir()` fails.
Corrupt JSON silently resets to empty (no crash).

---

## Data-Dir Dashboard + Retention (`app_datadir.go`, v1.3+)

`GetAppDataInventory()` returns one `DataDirEntry` per managed location (config / store / traces / runs / skills / pids / backups) with size, file count, oldest+newest mtime, and a `cleanable` flag.

`CleanupDataDir(key)`:
- `traces` → `ClearAITraces()` (deletes all `.jsonl`)
- `runs` → walks `runs/*` and removes all `.log` + `.json`
- `pids` → only removes PID files whose process is dead (`os.FindProcess` + `signal(0)`)

`Retention`:
- `Config.AI.TraceRetainDays` (default 3)
- `Config.Retention.RunsRetainDays` (default 14)
- `Config.Retention.RunsKeepPerName` (default 50)
- `ApplyRetention()` runs on startup and every hour via a ticker bound to the App ctx (exits when window closes).

---

## Backup / Restore (`app_v13.go`, v1.3)

- `ExportBackup(targetDir)` packs `config.json` + `scripts.json` + entire skills dir into `hatch-backup-<YYYYMMDD-HHMMSS>.zip` (default ~/Desktop) using stdlib `archive/zip`. No external deps.
- `ImportBackup(zipPath)` overlays files in-place: `config.json` and `scripts.json` go back to `<cfgDir>/Hatch/`; `skills/*` go back to current `SkillsDir`. Memory `a.cfg` is reloaded; ServiceManagers + scheduler rebuild on next access. App restart still recommended for full reset.

---

## Pre-Script Chain (`app_v13.go`, v1.3)

- `ScriptMeta.PreScripts []string` — paths run sequentially before the main script.
- `RunScript` transparently routes through `RunWithPreScripts` when set.
- Each pre runs as its own `RunOneshotWithEnv` (with the pre's own env if configured) but **does not emit its own events** — instead writes `[hatch] pre: <path>` lines to the main script's log stream.
- Any pre exiting != 0 → main aborts with `done(exitCode=99) + status=error`.
