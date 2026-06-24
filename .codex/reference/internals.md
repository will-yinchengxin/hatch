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

## Hooks

- `app_hooks.go` manages a registry of `command`-type hooks (`Config.Hooks`). One `HookDef` syncs to two targets sharing the same hooks JSON schema: Claude Code reads `~/.claude/settings.json` `hooks`; Codex reads `~/.codex/hooks.json` (Hatch uses the dedicated file rather than inline `config.toml` `[hooks]` to avoid touching the MCP block).
- Both writers operate on a generic hooks object (`map[event] -> []group`, each group `{matcher, hooks:[handler]}`) via `removeHatchHandler` / `insertHatchHandler`. Hatch handlers carry `hatchId`; removal targets one id (`writeClaudeHookConfig`) or all Hatch handlers (`writeCodexHooksFile` regenerates the whole Hatch portion). Hand-written handlers (no `hatchId`) are never removed; empty groups/events are pruned.
- `SetHookEnabled` gates Codex-unsupported events (`hookEventSupportedBy`) and, on first-time enable only, refuses to overwrite a matching hand-written hook (`claudeHookEntryExists` / `codexHookEntryExists` match event+matcher+command on non-`hatchId` handlers). Re-syncing an already-enabled hook overwrites its own marked handler. For Codex, the desired full list is written to file before the registry flag is committed, so a failed write leaves state untouched.
- `SaveHook` is pure (registry only, preserves sync flags); the frontend re-syncs already-enabled targets after save. `SuggestHookFromDescription` (draft from text) and `SuggestHookEdit` (refine an existing `HookDef` per an instruction, preserving its id) both go through `callAIOp` and return drafts for UI confirmation — they never persist. Both writers back up the target file via `backupIfExists` before writing.
- `ListHooks` seeds a default `hello-test` hook once (guarded by `Config.HooksSeeded`): a `UserPromptSubmit` hook that injects a confirmation into the context when the prompt contains "hello". Seeding enables it for Claude (writes `~/.claude/settings.json`) unless a hand-written conflict exists, so the capability is demonstrable out of the box.

## AI Workspace Profiles

- Profiles are manually previewed and applied; they never auto-switch from the current directory. They combine the project root, agent model settings, and target switches for Hatch-managed Skills, MCP, and Hooks.
- Applying snapshots Claude/Codex config files and master-known Skill targets before reusing normal sync methods. A failed apply restores the snapshot; the newest 20 private snapshots are retained.
- A divergent target Skill is never overwritten or removed: it is warned when deselected and blocks a selected sync. `SaveCodexConfig` validates TOML before writing.

## MCP

Hatch stores MCP registry data in `Config.MCPServers` and defaults the source directory to `~/MyHub/mcp`. First load seeds records from that directory when the registry is empty, then sets `Config.MCPSeeded` so discovery does not repeatedly re-import removed entries. The UI intentionally separates acquisition/import (local directory, GitHub clone, npm/npx registration) from target sync (Claude Code/Codex switches).

`SuggestMCPRecommendations(query)` returns ranked MCP candidates with version/source notes, install steps, required env, fit/risk notes, and a ready-to-edit `MCPServer`. It uses the active AI provider when configured and falls back to an offline curated catalog. Elasticsearch recommendations prefer Elastic Agent Builder's MCP endpoint when the user has Elastic 9.2+ or Serverless, otherwise prefer Elastic's official Docker stdio server for Claude Code/Codex compatibility, include streamable-HTTP for shared deployments, and keep third-party npx/local-build options as explicit alternatives. Recommendations only fill the frontend editor; they are not persisted until the user saves.

Sync targets:

- Claude Code: writes `~/.claude.json` top-level `mcpServers.<id>`.
- Codex: writes `~/.codex/config.toml` `[mcp_servers.<id>]` tables.

Both writes call `backupIfExists` first. Hatch does not keep MCP services running. `TestMCPConnection` shells out to `claude mcp get <id>` for Claude Code status and `codex mcp get <id>` for Codex config validation; Codex CLI output may not include live connection status, so those results can be `unknown` with `Mode=config`.

First-time sync refuses to overwrite an existing same-name target MCP entry. After Hatch has marked that target enabled, later MCP edits may resync that same entry. Deleting a Hatch MCP record also removes its previously synced Claude Code/Codex target entries before removing the registry record. `UninstallMCPFromDisk` additionally deletes git/local source dirs; `UpdateMCPFromGit` runs `git pull`, re-infers command/args, and resyncs enabled targets. The UI displays MCP command, args, headers, and env values in plaintext by default because the module is meant for local inspection and troubleshooting.

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

The VPN module targets Clash Verge Rev. It can read/edit configured YAML files, parse profiles, export/import backups, read live proxy groups over the Clash Unix socket, inspect enhance data from `merge` and `rules` profiles, and produce an ordinary-user health report.

`GetVPNStatus()` should prefer live Clash API data. `profiles.yaml` selected nodes are fallback only.

`GetVPNHealthReport()` aggregates directory availability, active profile, proxy/TUN mode, Clash core API reachability, current node, subscription expiry, and backup presence into `ok/warn/error` checks with Chinese suggestions. If a user-configured `VPNAppDir` is invalid, the health report surfaces that as an error rather than silently relying on auto-detection.

`ValidateVPNSettings()` checks that the Clash data directory exists and looks like a Clash Verge Rev directory, and verifies the backup directory can be created and written before saving settings.

`SaveVPNConfigFile` / `SaveVPNProfileContent` run `validateYAML` before overwriting so a malformed edit cannot break the Clash core; `.js` script profiles skip validation. `ImportVPNBackup` rejects entries that resolve outside the Clash dir (separator-boundary check). `DeleteVPNBackup` only removes `clash-verge-*.zip` files located directly inside the managed backup dir.

`PreviewVPNBackup()` inspects a zip without writing files, reports key config files, warns about missing `profiles.yaml`/`profiles/`, and flags suspicious paths. `ImportVPNBackup()` creates a `clash-verge-pre-restore-<ts>.zip` backup first when current config exists, then restores safe zip entries.
