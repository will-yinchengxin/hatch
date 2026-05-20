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

## Commands

```
Build:  $(go env GOPATH)/bin/wails build -platform darwin/arm64
Dev:    $(go env GOPATH)/bin/wails dev
Tests:  go test ./...   (or: go test ./internal/... -v after app.go changes)
Make:   dev / test / build / build-clean / sign / dmg / release / install / open
```

## Repository Layout

```
hatch/
├── app.go                  # App struct + startup
├── app_scripts.go          # Directory CRUD + scanning + SaveScriptMeta
├── app_executor.go         # RunScript / ServiceAction / run history / scriptEnv
├── app_zshrc.go            # zshrc parse / annotate / AI organize
├── app_ai.go               # AI providers / Suggest* / callAIOp / trace emit
├── app_observability.go    # Claude/Codex status / trace persistence / macOS notify
├── app_files.go            # Script files / AppDataPaths / AutoLaunch / DocLinks
├── app_v12.go              # v1.2: scheduler / health-check / prompt templates / stream
├── app_v13.go              # v1.3: PATH health / backup-restore / pre-script chain
├── app_datadir.go          # Data-dir inventory + cleanup + retention + Skill AI
├── app_vpn.go              # VPN module: Clash Verge Rev config/profile/backup + IP probe
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
| VPN config | `~/Library/Application Support/io.github.clash-verge-rev.clash-verge-rev/` (configurable) |
| VPN backups | `~/Library/Application Support/Hatch/vpn-backups/` (configurable) |

Retention runs on startup + hourly ticker (`ApplyRetention`): trace days, runs days, runs-per-script cap.

## Reference

| Topic | File |
|---|---|
| Core data model (`ScriptMeta` / `Config` structs) | [.claude/reference/data-model.md](.claude/reference/data-model.md) |
| All bound Go methods + real-time events | [.claude/reference/bound-methods.md](.claude/reference/bound-methods.md) |
| Frontend: Alpine state, UI tokens, CSS classes | [.claude/docs/frontend.md](.claude/docs/frontend.md) |
| Internals: scanner, executor, zshrc, scheduler | [.claude/docs/internals.md](.claude/docs/internals.md) |
| Known gotchas & Pico CSS notes | [.claude/docs/gotchas.md](.claude/docs/gotchas.md) |
| VPN module: Clash Verge Rev details | [implementation-notes.md](implementation-notes.md) |
| Frontend skill (component patterns) | [.claude/skills/hatch-frontend/SKILL.md](.claude/skills/hatch-frontend/SKILL.md) |
| Upgrade log (v1.1 → v1.3) | [UPGRADE_LOG.md](UPGRADE_LOG.md) |
