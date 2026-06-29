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
| `ListAIWorkspaceProfiles / SaveAIWorkspaceProfile / DeleteAIWorkspaceProfile` | Project-scoped AI profile CRUD |
| `CaptureCurrentAIWorkspaceProfile / PreviewAIWorkspaceProfileApply / ApplyAIWorkspaceProfile` | Capture current managed state; preflight/de-sensitized diff; confirmed apply with snapshot |
| `ListAIWorkspaceSnapshots / RollbackAIWorkspaceSnapshot / GetAIEnvironmentHealth` | Rollback inventory and AI CLI/config/profile-drift diagnostics |
| `GetClaudeCodePresets / SaveClaudeCodePreset / DeleteClaudeCodePreset` | Claude preset management |
| `GetCodexPresets / SaveCodexPreset / DeleteCodexPreset` | Codex preset management |
| `GetDocLinks / SaveDocLinks / OpenFile / ReadScriptFile / WriteScriptFile / CreateScriptFile` | File and doc-link operations |
| `GetMemoDashboard / ListMemos / SaveMemo / DeleteMemo / MarkMemoDone` | Memo/todo dashboard and CRUD; reminders are rescheduled after saves |
| `ImportMemoAttachment / DeleteMemoAttachment / OpenMemoAttachment / GetMemoAttachmentPreview` | Memo link/file/image attachment import, managed-file cleanup, opening, and image preview |
| `ListVocabEntries / SaveVocabEntry / DeleteVocabEntry / ReviewVocabEntry / LookupVocabWord / SpeakVocabWord / SuggestVocabEntry` | Vocab CRUD, review scheduling, online DictionaryAPI lookup, macOS `say` pronunciation fallback, and AI draft completion |
| `GetMemoSettings / SaveMemoSettings / CleanupMemoFiles` | Memo attachment storage directory and retention policy |
| `GetMCPSettings / SaveMCPSettings(mcpDir)` | Read/write MCP master dir (default `~/MyHub/mcp` if present); `SaveMCPSettings` creates the dir |
| `ListMCPServers / SaveMCPServer / DeleteMCPServer / GetMCPTemplates` | MCP registry CRUD + built-in template gallery; first `ListMCPServers` seeds the registry from disk once (`MCPSeeded`) |
| `SuggestMCPRecommendations(query)` | Return AI/offline MCP recommendation candidates and install notes; UI draft only |
| `ImportMCPFromLocalDir / InstallMCPFromGit / SaveNpmMCPServer` | Add MCP from local dir / git clone / npm package |
| `SetMCPEnabled(id, target, enabled)` | Sync MCP to `~/.claude.json` or `~/.codex/config.toml`; first-time enable refuses to overwrite a same-named hand-written entry (resync of an already-enabled MCP overwrites its own block) |
| `TestMCPConnection(id, target)` | Claude: `claude mcp get` real connection check; Codex: config readability only |
| `UninstallMCPFromDisk(id)` | Remove from Hatch registry + Claude/Codex configs + delete local dir (git/local only) |
| `UpdateMCPFromGit(id)` | `git pull` in local dir, re-infer command/args, resync to enabled targets |
| `ListHooks / SaveHook / DeleteHook / GetHooksStats` | Command-hook registry CRUD; `SaveHook` preserves sync flags; `DeleteHook` also removes from enabled targets |
| `SetHookEnabled(id, target, enabled)` | Sync one hook to `~/.claude/settings.json` (`hooks`) or `~/.codex/hooks.json`; first-time enable refuses to overwrite a matching hand-written hook; rejects Codex-unsupported events |
| `GetHookTemplates / GetHookEventSupport` | Built-in hook template gallery; event catalog with per-target (claude/codex) support flags |
| `SuggestHookFromDescription(desc)` | AI drafts a command `HookDef` from natural language; UI-only, never persisted |
| `SuggestHookEdit(current, instruction)` | AI refines an existing `HookDef` per an instruction (preserves id); UI-only, never persisted |
| `GetVPNSettings / SaveVPNSettings / ValidateVPNSettings / GetVPNAppDir / GetVPNProfiles / GetVPNStatus / GetVPNCurrentIP / GetVPNHealthReport` | VPN settings, path validation, live status, and ordinary-user health report |
| `GetVPNEnhanceData` | Parse merge/rules profiles into prepend rules, hosts, DNS fake-IP filters |
| `GetVPNConfigFile / SaveVPNConfigFile / GetVPNProfileContent / SaveVPNProfileContent` | VPN config/profile edit; saves validate YAML before overwrite (`.js` script profiles skip validation) and write a `.bak` |
| `ExportVPNBackup / PreviewVPNBackup / ImportVPNBackup / ListVPNBackups / DeleteVPNBackup` | VPN backup/preview/restore/delete; import is zip-slip guarded and creates a pre-restore backup when current config exists; delete only accepts `clash-verge-*.zip` inside the managed backup dir |
| `OpenVPNDir / OpenVPNBackupDir` | Finder open helpers |

## Events

| Event | Payload | When |
|---|---|---|
| `executor:log` | `{id, line, stream}` | Every stdout/stderr line |
| `executor:done` | `{id, exitCode}` | Oneshot completed |
| `executor:status` | `{id, status}` | service running/stopped/error |
| `ai:trace` | `AICallTrace` | Each AI call recorded |
| `ai:stream` | `{id, chunk, done, error}` | Streaming suggest output |
| `memo:reminder` | `MemoItem` | Memo reminder fires while Hatch is running |

Executor event `id` is the script absolute path.
