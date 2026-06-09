# Hatch — Project Reference

Hatch is a macOS script-management dashboard built with Wails v2: Go backend, WKWebView shell, Alpine.js frontend, Pico CSS base, and Vite bundling.

Target: Apple Silicon macOS. Distribution is outside the App Store and does not require an Apple Developer account.

## Source Of Truth

- `CLAUDE.md` is the Claude Code project reference. Keep it in sync when project behavior changes.
- `.claude/reference/data-model.md` records current config, script metadata, retention, and VPN data types.
- `.claude/reference/bound-methods.md` records Wails-bound Go methods and emitted events.
- `.claude/docs/frontend.md` records Alpine state, UI layout, CSS tokens, and frontend component conventions.
- `.claude/docs/internals.md` records scanner, executor, scheduler, zshrc, AI, data retention, backup, and pre-script behavior.
- `.claude/docs/gotchas.md` records non-obvious integration pitfalls.
- `.claude/skills/hatch-frontend/SKILL.md` must be used before editing `frontend/index.html`, `frontend/src/main.js`, or `frontend/src/style.css`.
- `.claude/skills/hatch-docs-update/SKILL.md` must be used when project docs need to be synchronized after feature, API, or architecture changes.

## Tech Stack

| Layer | Technology |
|---|---|
| Desktop shell | Wails v2 embedding a Go binary + WKWebView |
| Backend | Go 1.23+ |
| Frontend | Alpine.js v3 + Pico CSS v2 + Vite |
| Styling | Custom CSS variables, warm Anthropic-inspired light theme, no Tailwind |
| Config | JSON at `~/Library/Application Support/Hatch/config.json` |
| Runtime store | JSON at `~/Library/Application Support/Hatch/scripts.json` |

## Commands

```bash
$(go env GOPATH)/bin/wails dev
$(go env GOPATH)/bin/wails build -platform darwin/arm64
go test ./...
go test ./internal/... -v
make dev
make test
make build
make build-clean
make sign
make dmg
make release
make install
make open
```

Use `go test ./...` whenever a Wails-bound method, root package file, or generated binding surface may be affected. `go test ./internal/... -v` is enough for isolated internal package changes.

## Repository Shape

```text
hatch/
├── app.go                  # App startup, service restore, active AI config
├── app_scripts.go          # Directory CRUD, scanning, script metadata save/reset/order
├── app_executor.go         # RunScript, ServiceAction, RunWithPreScripts, run history, env/notify
├── app_scheduler.go        # Cron scheduling: ApplyAllSchedules, ListSchedules, RemoveSchedule
├── app_diagnostics.go      # CheckScriptHealth, CheckCommandAvailable, CheckPathHealth
├── app_backup.go           # ExportBackup, ImportBackup
├── app_zshrc.go            # zshrc parse, annotations, AI organization, ListAliasInventory
├── app_ai.go               # OpenAI-compatible providers, Suggest*, callAIOp, streaming, prompts
├── app_observability.go    # AI tool status, AI trace persistence, macOS notify
├── app_datadir.go          # Data inventory, cleanup, retention, Skill AI, paths, autolaunch, doc links
├── app_fileio.go           # ReadScriptFile, WriteScriptFile, CreateScriptFile, OpenFile
├── app_claude.go           # Claude Code settings, presets, TOML helpers
├── app_codex.go            # Codex settings, presets
├── app_vpn.go              # Clash Verge Rev config/profile/backup/IP/proxy helpers
├── app_skills.go           # SKILL.md master store and sync
├── app_hooks.go            # Command hook registry, Claude/Codex hook sync, AI draft
├── app_mcp.go              # MCP registry, AI recommendations, local/git/npm install, Claude/Codex sync, connection test
├── internal/               # config, scanner, executor, scheduler, autostart, store, zshrc
└── frontend/               # Alpine templates, state, CSS, assets, generated wailsjs
```

Never manually edit `frontend/wailsjs/**`; Wails regenerates it during `wails dev` or `wails build`.

## Runtime Data

| Purpose | Path |
|---|---|
| Config | `~/Library/Application Support/Hatch/config.json` |
| Script cache | `~/Library/Application Support/Hatch/scripts.json` |
| AI traces | `~/Library/Application Support/Hatch/traces/<date>.jsonl` by default |
| Run history | `~/Library/Application Support/Hatch/runs/<script>/<id>.{log,json}` |
| Master skills | `~/Library/Application Support/Hatch/skills/` by default |
| PID files | `/tmp/hatch/<script-name>.pid` |
| zshrc managed block | `~/.zshrc` Hatch block only |
| zshrc backups | `~/.zshrc.hatch-backup-<YYYYMMDD-HHMMSS>` |
| App autostart | `~/Library/LaunchAgents/com.hatch.app.plist` |
| Claude usage cache | `~/.cache/will-statusline/last.json` |
| Codex sessions | `~/.codex/sessions/` |
| Codex hooks (Hatch-managed) | `~/.codex/hooks.json` |
| Claude hooks (Hatch-managed) | `~/.claude/settings.json` `hooks` key |
| Backup export | `~/Desktop/hatch-backup-<ts>.zip` by default |
| VPN config | Clash Verge Rev app dir, auto-detected unless configured |
| VPN backups | `~/Library/Application Support/Hatch/vpn-backups/` by default |

Retention is applied on startup and hourly: AI trace days, run-history days, and run-history per-script cap.

## Backend Rules

- `ScriptMeta.Args == nil` means show the args input. `[]ArgDef{}` means explicitly hide it.
- `mergeOverride` only applies non-zero override fields; use `DeleteScriptMeta` to fully reset auto-detection.
- Service detection requires both `start` and `stop` in the first 200 lines.
- Oneshot scripts run through `/bin/zsh -l -c` and explicitly `source ~/.zshrc` before `exec`, because Finder/Dock launches do not get an interactive shell.
- Service processes use `Setpgid: true`; stop sends SIGTERM to the service process group.
- `findCLI()` should be used for user-installed tools that may only exist in login shell PATH.
- AI suggestions populate UI for user confirmation; they must not silently persist metadata, notes, prompts, skills, or zshrc changes.
- AI API keys are stored as local plaintext config; never print them in logs or expose them in screenshots.
- MCP env values are stored as local plaintext config; sync writes global target config but Hatch does not keep MCP services running.
- Hooks are `command`-type only (shared by Claude Code and Codex). One `HookDef` is authored once and synced to Claude (`~/.claude/settings.json` `hooks`) and/or Codex (`~/.codex/hooks.json`). Hatch-owned handlers carry a `hatchId` marker so sync updates/removes only its own entries and never touches hand-written hooks; first-time enable refuses to overwrite a matching hand-written hook. Codex-unsupported events (e.g. `Notification`, `SessionEnd`) are gated off for the Codex target.

## Frontend Rules

- Use the `hatch-frontend` skill before frontend edits.
- `frontend/src/main.js` imports Pico first and `./style.css` second. `frontend/index.html` must not link `/src/style.css`.
- Preserve the warm off-white, charcoal text, muted clay accent visual direction. Do not switch back to cold blue or dark neon styling unless explicitly requested.
- New feature markup should not use inline `style="..."`; add named CSS classes in `frontend/src/style.css`.
- Before using newly added Go methods through generated imports, run Wails regeneration. Until then call `window['go']['main']['App']['MethodName'](...)`.
- `setupEvents()` must clear prior Wails listeners with `EventsOff` before re-registering to avoid HMR duplicate handlers.
- In nested Alpine `x-for`, avoid "set reactive key then read it inside a method" flows; use an inline click handler for preset runs.

## Documentation Discipline

When code changes affect stable behavior, update the smallest relevant docs:

- Bound methods or event payloads: `.claude/reference/bound-methods.md` and `CLAUDE.md`.
- Config structs or script metadata: `.claude/reference/data-model.md` and `CLAUDE.md`.
- Frontend state, layout, or CSS classes: `.claude/docs/frontend.md` and `.claude/skills/hatch-frontend/SKILL.md`.
- Scanner, executor, scheduler, AI, data retention, backup, or VPN internals: `.claude/docs/internals.md`.
- Non-obvious bugs or integration pitfalls: `.claude/docs/gotchas.md`.

Do not create hooks, custom agents, subagents, or background automation without explicit user approval.
