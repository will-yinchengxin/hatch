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
```

**Tier 2a — Service pattern detection** (first 200 lines) — requires both `start` AND `stop` as case arms / quoted comparisons.

**Tier 2b — `--help` probe** (2s timeout) — first non-empty output line becomes `Desc`.

**Tier 3 — Defaults** — `type=oneshot`, `name=filename`, `pidFile=/tmp/hatch/<name>.pid`.

**Override layer** — `config.OverrideMeta[path]` merged on top. Non-zero fields only win; use `DeleteScriptMeta` to fully reset.

---

## Executor (`internal/executor/`)

### Oneshot (`oneshot.go`)
- Runs as `/bin/zsh -l -c 'source ~/.zshrc; exec <script> <args>'`
- Explicit `source ~/.zshrc` required — Wails GUI apps don't get interactive shells
- stdout/stderr streamed line-by-line via `Emitter.Log`
- First log line: `[hatch] command: <quoted invocation>`
- `Emitter.Done` fires with exit code

### Service (`service.go` — `ServiceManager`)
- Start: `Setpgid: true` (own process group), PID to `pidFile`
- Stop: SIGTERM to process group + PID file cleanup
- Status: `kill -0 <pid>` probe; stale PIDs auto-removed
- Restart: Stop (ignore error) → Start
- App startup: `restoreServices()` rebuilds and probes all stored services

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
Every AI call appends JSONL to `TraceLogDir/<YYYY-MM-DD>.jsonl` (default: `~/Library/Application Support/Hatch/logs/`). Readable via `GetTraceLog(date)`.

---

## Config (`internal/config/config.go`)

`os.UserConfigDir()` → `~/Library/Application Support` on macOS.
Both `config.json` and `scripts.json` in same dir.
Store falls back to `$TMPDIR/hatch/scripts.json` if `UserConfigDir()` fails.
Corrupt JSON silently resets to empty (no crash).
