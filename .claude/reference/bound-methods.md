# Hatch — Go Bound Methods & Events

Methods exposed to the Wails JS frontend via `app.go` and sibling files.

> Methods added after the last `wails dev`/`wails build` must be called as
> `window['go']['main']['App']['Method']()` until bindings regenerate.

## Bound Methods

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
| `GetVPNAppDir / GetVPNProfiles / GetVPNStatus / GetVPNCurrentIP` | VPN read — `GetVPNStatus` queries Clash Unix socket for live proxy+TUN state |
| `GetVPNEnhanceData()` | Parse all `merge`+`rules` profiles → prepend rules, hosts, DNS fake-IP list |
| `GetVPNConfigFile / SaveVPNConfigFile` | VPN config edit (verge/config/dns/profiles yaml) |
| `GetVPNProfileContent / SaveVPNProfileContent` | VPN profile file edit |
| `ExportVPNBackup / ImportVPNBackup / ListVPNBackups` | VPN zip backup / restore |
| `GetVPNSettings / SaveVPNSettings` | VPN configurable dirs (appDir + backupDir) |
| `OpenVPNDir / OpenVPNBackupDir` | Finder open VPN data / backup dirs |

## Real-time Events (Go → JS via EventsEmit)

| Event | Payload | When |
|---|---|---|
| `executor:log` | `{id, line, stream}` | Every stdout/stderr line |
| `executor:done` | `{id, exitCode}` | Oneshot completed |
| `executor:status` | `{id, status}` | running / stopped / error |
| `ai:trace` | `AICallTrace` | Each AI call recorded |
| `ai:stream` | `{id, chunk, done, error}` | Streaming Suggest output |

`id` = script's absolute `path` for executor events.
