# Hatch — Wails Bound Methods And Events

Methods are exposed from `app.go` and sibling `app_*.go` files to the Wails JS frontend. Methods added after the last `wails dev` or `wails build` are not in generated imports yet; call them as `window['go']['main']['App']['Method'](...)` until bindings regenerate.

## Bound Methods

| Method | Purpose |
|---|---|
| `GetConfig / ScanAll / GetScriptsByDir` | Config + scanning |
| `AddDirectory / RemoveDirectory / UpdateDirectoryExport` | Directory CRUD |
| `SaveScriptMeta / DeleteScriptMeta / SaveScriptOrder` | Override CRUD and ordering |
| `RunScript(id, path, args)` | Oneshot launch; pre-scripts auto-chain when configured |
| `ServiceAction(path, action, args)` | start / stop / restart with env |
| `GetAllServiceStatuses()` | path to running/stopped |
| `ListRuns / GetRunLog / DeleteRun / ListRecentRunsSorted` | Run-history archive |
| `SaveExports / CheckZshrc / ParseZshrc / SaveZshrcAnnotation` | Shell config |
| `SuggestZshrcOrganization / ApplyZshrcOrganization` | AI zshrc preview + backup/write |
| `GetAutoLaunchEnabled / SetAutoLaunchEnabled` | Login autostart |
| `GetClaudeCodeStatus / GetCodexUsageStatus` | Status cards |
| `GetSkills / GetSkillContent / SaveSkillContent / DeleteSkill / DeleteSkillEverywhere / GetSkillsStats` | Skill CRUD |
| `SetSkillSync / ImportSkillsFrom / ResolveSkillImportConflict / CreateSkillTemplate` | Skill sync + import conflict handling |
| `SuggestSkillContent / SuggestSkillFromDescription` | Skill AI optimize/generation |
| `GetAppDataPaths / SaveAppDataPaths` | Trace + skills dir config |
| `GetAITraces / ClearAITraces / GetAITraceStorageInfo` | AI trace inspect |
| `SaveAIConfig / GetAIProviders / SaveAIProviders / TestAIConfig` | Provider management |
| `SuggestScriptMeta / SuggestZshrcAnnotation / SuggestScriptContent` | AI suggestions |
| `SuggestScriptContentStream` | Emits `ai:stream`; protocol fixed even if currently one chunk |
| `GetPromptTemplates / SavePromptTemplates / GetDefaultPrompts` | Prompt overrides |
| `ApplyAllSchedules / ListSchedules / RemoveSchedule` | Scheduler control |
| `CheckScriptHealth / CheckCommandAvailable / CheckPathHealth / ListAliasInventory` | Dependency, health, PATH, alias inspection |
| `ExportBackup / ImportBackup` | Config/store/skills zip backup |
| `RunWithPreScripts` | Explicit chain runner |
| `GetAppDataInventory / CleanupDataDir / GetRetentionPolicy / SaveRetentionPolicy` | Data dashboard |
| `GetClaudeSettings / SaveClaudeSettings / GetClaudeCodeConfig / SetClaudeCodeConfig` | Claude config editors |
| `GetCodexConfig / SaveCodexConfig / GetCodexModelConfig / SetCodexModelConfig` | Codex config editors, including model, reasoning effort, and `[features].goals` |
| `GetClaudeCodePresets / SaveClaudeCodePreset / DeleteClaudeCodePreset` | Claude preset management |
| `GetCodexPresets / SaveCodexPreset / DeleteCodexPreset` | Codex preset management |
| `GetDocLinks / SaveDocLinks / OpenFile / ReadScriptFile / WriteScriptFile / CreateScriptFile` | File and doc-link operations |
| `UninstallMCPFromDisk(id)` | Remove from Hatch registry + Claude/Codex configs + delete local dir (git/local only) |
| `UpdateMCPFromGit(id)` | `git pull` in local dir, re-infer command/args, resync to enabled targets |
| `GetVPNSettings / SaveVPNSettings / GetVPNAppDir / GetVPNProfiles / GetVPNStatus / GetVPNCurrentIP` | VPN settings and live status |
| `GetVPNEnhanceData` | Parse merge/rules profiles into prepend rules, hosts, DNS fake-IP filters |
| `GetVPNConfigFile / SaveVPNConfigFile / GetVPNProfileContent / SaveVPNProfileContent` | VPN config/profile edit |
| `ExportVPNBackup / ImportVPNBackup / ListVPNBackups` | VPN backup/restore |
| `OpenVPNDir / OpenVPNBackupDir` | Finder open helpers |

## Events

| Event | Payload | When |
|---|---|---|
| `executor:log` | `{id, line, stream}` | Every stdout/stderr line |
| `executor:done` | `{id, exitCode}` | Oneshot completed |
| `executor:status` | `{id, status}` | service running/stopped/error |
| `ai:trace` | `AICallTrace` | Each AI call recorded |
| `ai:stream` | `{id, chunk, done, error}` | Streaming suggest output |

Executor event `id` is the script absolute path.
