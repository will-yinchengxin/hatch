# Hatch — Internals Reference

## Scanner

Detection order per file:

1. `@hatch.*` annotations in first 50 lines, supporting `#`, `//`, and `--` comments.
2. Service pattern detection in first 200 lines. Requires both `start` and `stop`.
3. `--help` probe with 2s timeout; first non-empty line becomes `Desc`.
4. Defaults: `type=oneshot`, name from filename, `pidFile=/tmp/hatch/<name>.pid`.
5. `OverrideMeta[path]` merged on top with non-zero fields only.

Supported annotations include `name`, `desc`, `type`, `commands`, `pidfile`, `tags`, `schedule`, `requires`, and `health`.

## Executor

- Oneshot runs through `RunOneshotWithEnv(id, path, args, env, em, hook)`.
- Command shell is `/bin/zsh -l -c 'source ~/.zshrc; exec <script> <args>'`.
- Output streams line-by-line through `Emitter.Log`; first line is `[hatch] command: ...`.
- Env overrides merge over `os.Environ`.
- `NotifyHook` sends macOS notifications after completion when enabled.
- Run history writes `<runs>/<basename>/<id>.log` while running and `<id>.json` on finalize.

Services:

- `StartWithEnv` / `RestartWithEnv` support per-script env.
- `Setpgid: true`; PID written to `pidFile`.
- `Stop` sends SIGTERM to the service process group and removes PID files.
- `Status` uses `kill -0` and removes stale PID files.
- App startup restores service managers and probes all stored service scripts.

## Scheduler

`internal/scheduler` has no external deps. `ParseSchedule` supports:

- `@every Ns`, `@every Nm`, `@every Nh`
- `@daily HH:MM`
- `@hourly :MM`

`ApplyAllSchedules()` rebuilds schedules on startup and after every `SaveScriptMeta`. The startup delay is intentional so the store has current script metadata.

## zshrc

Managed block:

```sh
# >>> Hatch managed block >>>
export MY_TOOLS="/path"
export PATH=$MY_TOOLS:$PATH
# <<< Hatch managed block <<<
```

Rules:

- `WriteBlock` atomically replaces only the Hatch block.
- `RemoveBlock` runs when exports are empty.
- Export keys are sorted.
- `$` in values is escaped before regex replacement.
- `Backup(path)` writes `<path>.hatch-backup-<YYYYMMDD-HHMMSS>`.
- `WriteContent` replaces the entire file and ensures trailing newline.

Parser entry keys:

| Type | Key |
|---|---|
| `function` | function name |
| `alias` | alias name |
| `export` | variable name |
| `path` | `PATH` |
| `source` | `source:N` |
| `section` | `comment:N` |
| `other` | `other:N` |

Positional keys change when the file is edited.

## AI

- Chat Completions compatible API, timeout 45s.
- Base URL can be provider base or full `/chat/completions`; suffix is appended only when missing.
- Script analysis truncates content to 12,000 bytes.
- zshrc organization truncates content to 24,000 bytes and returns `{summary, content}`.
- `PromptTemplates` override effective system prompts by op key.
- `GetDefaultPrompts()` must stay in sync with hard-coded defaults in suggestion methods.
- Trace JSONL is written under `TraceLogDir/<YYYY-MM-DD>.jsonl` and pruned by retention.
- Streaming protocol emits `ai:stream`; current implementation may still send one chunk.
- Skill AI returns full `SKILL.md` content without markdown fences.
- Skill import fingerprints the whole skill directory by relative path + file content. Same-name differences return conflicts for user choice; Hatch UI saves still resync master content to enabled Claude/Codex targets.
- AI Tools writes Codex shortcuts into `~/.codex/config.toml`: top-level `model`, top-level `model_reasoning_effort`, and `[features] goals = true/false`. The merge helpers only touch those keys and preserve nested profiles.

## Data Retention And Backup

`GetAppDataInventory()` reports config, store, traces, runs, skills, pids, and backups with size, file count, oldest/newest mtime, and cleanable flag.

`CleanupDataDir`:

- `traces`: deletes all trace JSONL.
- `runs`: deletes run `.log` and `.json` files.
- `pids`: removes only stale PID files.

Retention fields:

- `AI.TraceRetainDays`, default 3.
- `Retention.RunsRetainDays`, default 14.
- `Retention.RunsKeepPerName`, default 50.

`ExportBackup(targetDir)` zips `config.json`, `scripts.json`, and skills. `ImportBackup(zipPath)` overlays those files back into the current data locations and reloads memory config; app restart remains recommended.

## Pre-Script Chain

`ScriptMeta.PreScripts` are run sequentially before the main script. `RunScript` routes to `RunWithPreScripts` automatically when set.

Each pre-script uses its own env if configured, but does not emit its own events. The main log receives `[hatch] pre: <path>` lines. Any pre-script non-zero exit aborts the main run with exit code 99 and status error.

## VPN

The VPN module targets Clash Verge Rev. It can read/edit configured YAML files, parse profiles, export/import backups, read live proxy groups over the Clash Unix socket, and inspect enhance data from `merge` and `rules` profiles.

`GetVPNStatus()` should prefer live Clash API data. `profiles.yaml` selected nodes are fallback only.

`SaveVPNConfigFile` / `SaveVPNProfileContent` run `validateYAML` before overwriting so a malformed edit cannot break the Clash core; `.js` script profiles skip validation. `ImportVPNBackup` rejects entries that resolve outside the Clash dir (separator-boundary check). `DeleteVPNBackup` only removes `clash-verge-*.zip` files located directly inside the managed backup dir.
