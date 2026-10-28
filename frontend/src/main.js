import '@picocss/pico/css/pico.min.css';
import './style.css';
import Alpine from 'alpinejs';
import { leaderboardState } from './leaderboards';

import {
  GetConfig,
  ScanAll,
  AddDirectory,
  RemoveDirectory,
  UpdateDirectoryExport,
  RunScript,
  ServiceAction,
  GetAllServiceStatuses,
  SaveExports,
  CheckZshrc,
  SaveScriptMeta,
  DeleteScriptMeta,
} from '../wailsjs/go/app/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

// New methods added to Go but not yet in wailsjs/App.js (regenerated on next build).
const SaveAIConfig        = (ai)      => window['go']['app']['App']['SaveAIConfig'](ai);
const TestAIConfig        = (ai)      => window['go']['app']['App']['TestAIConfig'](ai);
const SuggestScriptMeta   = (meta)    => window['go']['app']['App']['SuggestScriptMeta'](meta);
const GetAutoLaunchEnabled = () => window['go']['app']['App']['GetAutoLaunchEnabled']();
const SetAutoLaunchEnabled = (enabled) => window['go']['app']['App']['SetAutoLaunchEnabled'](enabled);
const GetClaudeCodeStatus   = ()                            => window['go']['app']['App']['GetClaudeCodeStatus']();
const GetCodexUsageStatus   = ()                            => window['go']['app']['App']['GetCodexUsageStatus']();
const ReadScriptFile        = (path)                        => window['go']['app']['App']['ReadScriptFile'](path);
const WriteScriptFile       = (path, content)               => window['go']['app']['App']['WriteScriptFile'](path, content);
const CreateScriptFile      = (dirPath, filename, content)  => window['go']['app']['App']['CreateScriptFile'](dirPath, filename, content);
const SuggestScriptContent  = (path, cur, inst)             => window['go']['app']['App']['SuggestScriptContent'](path, cur, inst);
const GetAITraces           = ()                            => window['go']['app']['App']['GetAITraces']();
const ClearAITraces         = ()                            => window['go']['app']['App']['ClearAITraces']();
const GetAITraceStorageInfo = ()                            => window['go']['app']['App']['GetAITraceStorageInfo']();
const GetSkills             = ()                            => window['go']['app']['App']['GetSkills']();
const GetSkillsStats        = ()                            => window['go']['app']['App']['GetSkillsStats']();
const GetSkillContent       = (name)                        => window['go']['app']['App']['GetSkillContent'](name);
const SaveSkillContent      = (name, content)               => window['go']['app']['App']['SaveSkillContent'](name, content);
const DeleteSkill           = (name)                        => window['go']['app']['App']['DeleteSkill'](name);
const DeleteSkillEverywhere = (name)                        => window['go']['app']['App']['DeleteSkillEverywhere'](name);
const SetSkillSync          = (name, target, enabled)       => window['go']['app']['App']['SetSkillSync'](name, target, enabled);
const ImportSkillsFrom      = (source)                      => window['go']['app']['App']['ImportSkillsFrom'](source);
const ResolveSkillImportConflict = (name, source, useSource) => window['go']['app']['App']['ResolveSkillImportConflict'](name, source, useSource);
const CreateSkillTemplate   = (name)                        => window['go']['app']['App']['CreateSkillTemplate'](name);
const GetAppDataPaths       = ()                            => window['go']['app']['App']['GetAppDataPaths']();
const SaveAppDataPaths      = (traceDir, skillsDir)         => window['go']['app']['App']['SaveAppDataPaths'](traceDir, skillsDir);
const GetAIProviders        = ()                            => window['go']['app']['App']['GetAIProviders']();
const SaveAIProviders       = (providers, activeID)         => window['go']['app']['App']['SaveAIProviders'](providers, activeID);
const GetDocLinks           = ()                            => window['go']['app']['App']['GetDocLinks']();
const SaveDocLinks          = (links)                       => window['go']['app']['App']['SaveDocLinks'](links);
const OpenFile              = (path)                        => window['go']['app']['App']['OpenFile'](path);
const GetVPNAppDir          = ()                            => window['go']['app']['App']['GetVPNAppDir']();
const GetVPNProfiles        = ()                            => window['go']['app']['App']['GetVPNProfiles']();
const GetVPNConfigFile      = (name)                        => window['go']['app']['App']['GetVPNConfigFile'](name);
const SaveVPNConfigFile     = (name, content)               => window['go']['app']['App']['SaveVPNConfigFile'](name, content);
const GetVPNProfileContent  = (uid)                         => window['go']['app']['App']['GetVPNProfileContent'](uid);
const SaveVPNProfileContent = (uid, content)                => window['go']['app']['App']['SaveVPNProfileContent'](uid, content);
const ExportVPNBackup       = ()                            => window['go']['app']['App']['ExportVPNBackup']();
const ImportVPNBackup       = (zipPath)                     => window['go']['app']['App']['ImportVPNBackup'](zipPath);
const ListVPNBackups        = ()                            => window['go']['app']['App']['ListVPNBackups']();
const DeleteVPNBackup       = (zipPath)                     => window['go']['app']['App']['DeleteVPNBackup'](zipPath);
const OpenVPNDir            = ()                            => window['go']['app']['App']['OpenVPNDir']();
const OpenVPNBackupDir      = ()                            => window['go']['app']['App']['OpenVPNBackupDir']();
const GetVPNSettings        = ()                            => window['go']['app']['App']['GetVPNSettings']();
const SaveVPNSettings       = (appDir, backupDir)           => window['go']['app']['App']['SaveVPNSettings'](appDir, backupDir);
const GetVPNStatus          = ()                            => window['go']['app']['App']['GetVPNStatus']();
const GetVPNCurrentIP       = ()                            => window['go']['app']['App']['GetVPNCurrentIP']();
const GetVPNHealthReport    = ()                            => window['go']['app']['App']['GetVPNHealthReport']();
const ValidateVPNSettings   = (appDir, backupDir)           => window['go']['app']['App']['ValidateVPNSettings'](appDir, backupDir);
const PreviewVPNBackup      = (zipPath)                     => window['go']['app']['App']['PreviewVPNBackup'](zipPath);
const GetClaudeSettings     = ()                            => window['go']['app']['App']['GetClaudeSettings']();
const SaveClaudeSettings    = (content)                     => window['go']['app']['App']['SaveClaudeSettings'](content);
const GetCodexConfig        = ()                            => window['go']['app']['App']['GetCodexConfig']();
const SaveCodexConfig       = (content)                     => window['go']['app']['App']['SaveCodexConfig'](content);
const GetMCPInventory       = (includeCLI)                  => window['go']['app']['App']['GetMCPInventory'](includeCLI);
const GetMCPCatalog         = ()                            => window['go']['app']['App']['GetMCPCatalog']();
const ParseMCPInput         = (text)                        => window['go']['app']['App']['ParseMCPInput'](text);
const SearchMCPRegistry     = (query)                       => window['go']['app']['App']['SearchMCPRegistry'](query);
const LookupMCPRegistry     = (name)                        => window['go']['app']['App']['LookupMCPRegistry'](name);
const PreviewMCPApply       = (req)                         => window['go']['app']['App']['PreviewMCPApply'](req);
const ApplyMCP              = (req)                         => window['go']['app']['App']['ApplyMCP'](req);
const SetMCPEnabled         = (name, key, enabled)          => window['go']['app']['App']['SetMCPEnabled'](name, key, enabled);
const RemoveMCPEntry        = (name, key)                   => window['go']['app']['App']['RemoveMCPEntry'](name, key);
const CopyMCPEntry          = (name, from, to)              => window['go']['app']['App']['CopyMCPEntry'](name, from, to);
const ProbeMCP              = (name, key)                   => window['go']['app']['App']['ProbeMCP'](name, key);
const SetMCPDisabledTools   = (name, key, tools)            => window['go']['app']['App']['SetMCPDisabledTools'](name, key, tools);
const DiscardLegacyMCP      = (name)                        => window['go']['app']['App']['DiscardLegacyMCP'](name);
const ListHooks             = ()                            => window['go']['app']['App']['ListHooks']();
const SaveHook              = (hook)                        => window['go']['app']['App']['SaveHook'](hook);
const DeleteHook            = (id)                          => window['go']['app']['App']['DeleteHook'](id);
const SetHookEnabled        = (id, target, enabled)         => window['go']['app']['App']['SetHookEnabled'](id, target, enabled);
const GetHookTemplates      = ()                            => window['go']['app']['App']['GetHookTemplates']();
const GetHooksStats         = ()                            => window['go']['app']['App']['GetHooksStats']();
const GetHookEventSupport   = ()                            => window['go']['app']['App']['GetHookEventSupport']();
const SuggestHookFromDescription = (desc)                   => window['go']['app']['App']['SuggestHookFromDescription'](desc);
const SuggestHookEdit       = (hook, instruction)          => window['go']['app']['App']['SuggestHookEdit'](hook, instruction);
const ListAIWorkspaceProfiles = ()                         => window['go']['app']['App']['ListAIWorkspaceProfiles']();
const SaveAIWorkspaceProfile = (profile)                   => window['go']['app']['App']['SaveAIWorkspaceProfile'](profile);
const DeleteAIWorkspaceProfile = (id)                      => window['go']['app']['App']['DeleteAIWorkspaceProfile'](id);
const PreviewAIWorkspaceProfileApply = (id)                => window['go']['app']['App']['PreviewAIWorkspaceProfileApply'](id);
const ApplyAIWorkspaceProfile = (id)                       => window['go']['app']['App']['ApplyAIWorkspaceProfile'](id);
const ListAIWorkspaceSnapshots = ()                        => window['go']['app']['App']['ListAIWorkspaceSnapshots']();
const RollbackAIWorkspaceSnapshot = (id)                   => window['go']['app']['App']['RollbackAIWorkspaceSnapshot'](id);
const GetAIEnvironmentHealth = ()                          => window['go']['app']['App']['GetAIEnvironmentHealth']();
const GetMemoDashboard     = ()                            => window['go']['app']['App']['GetMemoDashboard']();
const ListMemos            = ()                            => window['go']['app']['App']['ListMemos']();
const SaveMemo             = (item)                        => window['go']['app']['App']['SaveMemo'](item);
const DeleteMemo           = (id)                          => window['go']['app']['App']['DeleteMemo'](id);
const MarkMemoDone         = (id, done)                    => window['go']['app']['App']['MarkMemoDone'](id, done);
const ImportMemoAttachment = (path, label, kind)           => window['go']['app']['App']['ImportMemoAttachment'](path, label, kind);
const DeleteMemoAttachment = (memoID, attachmentID)        => window['go']['app']['App']['DeleteMemoAttachment'](memoID, attachmentID);
const OpenMemoAttachment   = (att)                         => window['go']['app']['App']['OpenMemoAttachment'](att);
const GetMemoAttachmentPreview = (att)                     => window['go']['app']['App']['GetMemoAttachmentPreview'](att);
const ListVocabEntries     = ()                            => window['go']['app']['App']['ListVocabEntries']();
const SaveVocabEntry       = (entry)                       => window['go']['app']['App']['SaveVocabEntry'](entry);
const DeleteVocabEntry     = (id)                          => window['go']['app']['App']['DeleteVocabEntry'](id);
const ReviewVocabEntry     = (id, result)                  => window['go']['app']['App']['ReviewVocabEntry'](id, result);
const LookupVocabWord      = (word)                        => window['go']['app']['App']['LookupVocabWord'](word);
const SuggestVocabEntry    = (entry)                       => window['go']['app']['App']['SuggestVocabEntry'](entry);
const SpeakVocabWord       = (word)                        => window['go']['app']['App']['SpeakVocabWord'](word);
const GetMemoSettings      = ()                            => window['go']['app']['App']['GetMemoSettings']();
const SaveMemoSettings     = (dir, days, auto)             => window['go']['app']['App']['SaveMemoSettings'](dir, days, auto);
const CleanupMemoFiles     = ()                            => window['go']['app']['App']['CleanupMemoFiles']();
const GetModuleSettings    = ()                            => window['go']['app']['App']['GetModuleSettings']();

window.hatchApp = function () {
  return {
    ...leaderboardState(),
    // ── State ──────────────────────────────────────────────────────────
    tab: 'ai',
    // Top-level module visibility, from config.json "modules" (GetModuleSettings).
    modules: { workbench: true, ai: true, vpn: true, memo: true, host: true },
    hostSnapshot: null,
    hostHardware: null,
    hostPrevious: null,
    hostRates: [],
    hostShowIdleNet: false,   // idle (0 B/s) interfaces are folded behind a toggle
    hostDiskResult: null,
    hostNetworkResult: null,
    hostDiskBusy: false,
    hostNetworkBusy: false,
    hostStorageBusy: false,
    hostStorage: null,
    hostPorts: null,
    hostPortsBusy: false,
    hostLoading: false,
    hostError: '',
    hostTimer: null,

    allScripts: [],

    // Workbench (工作台): sidebar sub-page, master list filter, selected script + detail pane
    wbSubTab: 'scripts',   // scripts | env | dirs | runs
    wbQuery: '',
    wbFilter: 'all',       // all | fav | service | scheduled
    wbSelected: null,      // script path
    wbPane: 'run',         // run | code | config | history
    wbArgValues: {},       // path → { argIndex: value } for declared @hatch.arg fields
    wbCode: { path: '', content: '', original: '', loading: false, saving: false, blocked: '', size: 0 },
    wbHealth: { path: '', result: null, loading: false },   // @hatch.requires / health URL check for the selected script   // blocked: '' | 'binary' | 'large'
    wbAI: { instruction: '', busy: false, result: null, diff: [], stats: { add: 0, del: 0 } },
    wbFormSnapshot: '',
    wbRecentRuns: [],
    wbRunsFailedOnly: false,
    // Global env editor (rows mirror cfg.exports plus UI-only fields)
    envRows: [],
    envDirty: false,
    envQuery: '',
    envReveal: {},
    envImportOpen: false,
    envImportText: '',
    // New script dialog + ⌘K palette
    newScript: { open: false, dir: '', filename: '', mode: 'template', template: 'zsh', instruction: '', content: '', busy: false, saving: false },
    palette: { open: false, query: '', index: 0 },

    loading: false,
    error: '',
    errorTimer: null,
    success: '',
    successTimer: null,

    statuses: {},
    runningSet: {},
    scriptArgs: {},
    scriptLogs: {},
    scriptExitCodes: {},   // path → exit code (number)

    cfg: { directories: [], exports: [] },
    aiForm: { enabled: false, baseUrl: '', apiKey: '', model: '' },
    aiSaving: false,
    aiTesting: false,
    aiTestResult: '',

    // Multi-provider state
    aiProviders: [],
    activeProviderID: '',

    // v1.3+: header quick AI panel
    quickAIOpen: false,
    quickAIActiveID: '',
    quickAITraceID: null,
    providerForm: null,   // null = closed; object = editing/creating
    providerTesting: false,
    providerTestResult: '',

    aiBusyScript: null,
    zshrcInfo: null,
    newDir: { path: '' },
    savingExports: false,

    inlineExportDir: null,
    inlineExportName: '',

    // Script config form (workbench 配置 pane)
    editForm: {},
    editSaving: false,

    // v1.3+: data dir inventory & retention
    dataInventory: [],
    backupBusy: false,
    backupResult: null,      // ExportBackup result { path, size, numFiles }
    backupImportPath: '',
    retention: { traceRetainDays: 3, runsRetainDays: 14, runsKeepPerName: 50 },
    retentionSaving: false,

    // v1.2: prompt templates
    promptTemplates: { SuggestScriptMeta: '', SuggestZshrcAnnotation: '', SuggestScriptContent: '', SuggestZshrcOrganization: '', TestAIConfig: '', SuggestHookFromDescription: '', SuggestHookEdit: '', SuggestVPNRuleChange: '', SuggestVPNEnhanceEdit: '' },
    promptDefaults: {},
    promptDefaultsOpen: {},
    promptSaving: false,
    promptOps: [
      { key: 'SuggestScriptMeta',           label: '脚本元信息分析',   group: '脚本' },
      { key: 'SuggestScriptContent',        label: '脚本内容生成',     group: '脚本' },
      { key: 'SuggestZshrcAnnotation',      label: 'zshrc 条目备注',   group: 'Zshrc' },
      { key: 'SuggestZshrcOrganization',    label: 'zshrc 整理重排',   group: 'Zshrc' },
      { key: 'SuggestHookFromDescription',  label: 'Hook 起草',        group: 'Hooks' },
      { key: 'SuggestHookEdit',             label: 'Hook 优化',        group: 'Hooks' },
      { key: 'SuggestVPNRuleChange',        label: '一句话改分流规则', group: 'VPN' },
      { key: 'SuggestVPNEnhanceEdit',       label: '订阅增强文件修改', group: 'VPN' },
      { key: 'TestAIConfig',                label: '配置连通性测试',   group: '系统' },
    ],

    // v1.1: run history
    historyTarget: null,
    historyRecords: [],
    historyLoading: false,
    historyOpenId: null,
    historyLogText: '',

    autoLaunchEnabled: false,
    autoLaunchSaving: false,

    // Claude Code status
    claudeStatus: null,
    claudeStatusLoading: false,

    // Codex usage status
    codexStatus: null,
    codexStatusLoading: false,

    // Claude Code model + effort config.
    // Single source of truth for the model catalog: `kind` splits the dropdown into
    // aliases (follow the latest model in a family) and pinned version IDs; `efforts`
    // is the effort levels that model actually accepts (Haiku supports none).
    // Claude Code model catalog (Anthropic API, checked 2026-10 against code.claude.com/docs/en/model-config).
    claudeModels: [
      { value: 'opus',              label: 'Opus（最新 · 5.5）',          kind: 'alias',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'sonnet',            label: 'Sonnet（最新 · 5.5）',        kind: 'alias',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'fable',             label: 'Fable（最新 · 5.1）',         kind: 'alias',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'haiku',             label: 'Haiku（最新 · 5.5）',         kind: 'alias',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'best',              label: '最强可用（有 Fable 用 Fable）', kind: 'alias',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'opusplan',          label: '规划用 Opus，执行用 Sonnet',  kind: 'alias',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-fable-5-1',  label: 'Fable 5.1',      kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-opus-5-5',   label: 'Opus 5.5',       kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-sonnet-5-5', label: 'Sonnet 5.5',     kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-haiku-5-5',  label: 'Haiku 5.5',      kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-fable-5',    label: 'Fable 5',        kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-opus-5',     label: 'Opus 5',         kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-sonnet-5',   label: 'Sonnet 5',       kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-opus-4-8',   label: 'Opus 4.8',       kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-opus-4-7',   label: 'Opus 4.7',       kind: 'pinned', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'claude-opus-4-6',   label: 'Opus 4.6',       kind: 'pinned', efforts: ['low', 'medium', 'high', 'max'] },
      { value: 'claude-sonnet-4-6', label: 'Sonnet 4.6',     kind: 'pinned', efforts: ['low', 'medium', 'high', 'max'] },
      { value: 'claude-haiku-4-5',  label: 'Haiku 4.5',      kind: 'pinned', efforts: [] },
    ],
    claudeCodeConfig: { model: '', effortLevel: '' },
    claudeCodeConfigSaving: false,
    claudeCodePresets: [],
    claudeCodePresetForm: { id: '', name: '', model: '', effortLevel: '' },
    claudeCodePresetOpen: false,  // inline add/edit form visible

    // Codex model + reasoning effort + goal-mode config
    // Codex model catalog (checked 2026-10 against learn.chatgpt.com/docs/models). kind: current | legacy.
    // Efforts per the docs: 6.1 Sol low→ultra (max/ultra depend on settings); 6 Luna up to max, no ultra;
    // 6 Astra presets low/medium/xhigh (xhigh omitted on some paid plans); 6 Sol levels beyond medium are
    // not documented, so it is capped at xhigh like Astra. 5.6 entries keep their pre-6 lists.
    codexModels: [
      { value: 'gpt-6.1-sol',   label: 'GPT-6.1 Sol · 复杂编码 / 代理任务', kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh', 'max', 'ultra'] },
      { value: 'gpt-6-astra',   label: 'GPT-6 Astra · 最难的端到端任务',    kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh'] },
      { value: 'gpt-6-sol',     label: 'GPT-6 Sol',                         kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh'] },
      { value: 'gpt-6-luna',    label: 'GPT-6 Luna · 轻快、可重复的任务',   kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'gpt-5.6-sol',   label: 'GPT-5.6 Sol',                       kind: 'legacy',  efforts: ['low', 'medium', 'high', 'xhigh', 'max', 'ultra'] },
      { value: 'gpt-5.6-terra', label: 'GPT-5.6 Terra',                     kind: 'legacy',  efforts: ['low', 'medium', 'high', 'xhigh', 'max', 'ultra'] },
      { value: 'gpt-5.6-luna',  label: 'GPT-5.6 Luna',                      kind: 'legacy',  efforts: ['low', 'medium', 'high', 'xhigh', 'max'] },
      { value: 'gpt-5.5',       label: 'GPT-5.5 · 仅 API Key 登录可用',     kind: 'legacy',  efforts: ['low', 'medium', 'high', 'xhigh'] },
    ],
    codexModelConfig: { model: '', effortLevel: '', goalsEnabled: false },
    codexModelConfigSaving: false,
    codexPresets: [],
    codexPresetForm: { id: '', name: '', model: '', effortLevel: '' },
    codexPresetOpen: false,

    // Config file editors
    claudeSettingsOpen: false,
    claudeSettingsContent: '',
    claudeSettingsSaving: false,
    codexConfigOpen: false,
    codexConfigContent: '',
    codexConfigSaving: false,

    // AI trace log
    aiTraces: [],
    aiTraceExpanded: {},
    aiTraceStorage: null,
    aiRetainDaysSaving: false,
    aiTraceCollapsed: true,

    // Skills management
    skills: [],
    skillsStats: { total: 0, claude: 0, codex: 0, both: 0 },
    skillsLoading: false,
    skillSearch: '',
    skillFilter: 'all',   // 'all' | 'claude' | 'codex'
    skillEditTarget: null,   // skill name being edited, or null
    skillEditContent: '',
    skillEditSaving: false,
    skillEditIsNew: false,
    skillNewName: '',
    skillSyncBusy: {},       // { "name:target": true }
    skillImporting: null,    // null | "claude" | "codex"
    skillImportResult: null,
    skillConflictBusy: {},   // { "source:name": true }

    // MCP (live view of Claude Code / Codex config)
    mcpInventory: null,      // GetMCPInventory result: { rows, projects, sources, legacy, cliIncluded }
    mcpCatalog: [],
    mcpLoading: false,
    mcpCLILoading: false,
    mcpProbingAll: false,
    mcpSearch: '',
    mcpFilter: 'all',        // all | claude | codex | drift | off | readonly
    mcpExpanded: {},         // { name: true }
    mcpBusy: {},             // { "name|locationKey|op": true }
    mcpAdd: null,            // add / edit / copy drawer state
    confirmDialog: { open: false, title: '', message: '', confirmText: '', danger: true },
    confirmResolve: null,

    // Hooks management
    hooks: [],
    hooksStats: { total: 0, claude: 0, codex: 0 },
    hooksLoading: false,
    hooksSearch: '',
    hooksFilter: 'all',          // 'all' | 'claude' | 'codex'
    hookEventSupport: [],        // [{event,label,claude,codex,matcherHint}]
    hookTemplates: [],
    hookTemplateGalleryOpen: false,
    hookEditOpen: false,
    hookEditIsNew: false,
    hookForm: null,              // editing draft
    hookSaving: false,
    hookSyncBusy: {},            // { "id:target": true }
    hookSuggestDesc: '',
    hookSuggestBusy: false,
    hookAIInstruction: '',       // AI optimize bar inside the editor
    hookAIBusy: false,

    // AI workspace profiles
    workspaceProfiles: [],
    workspaceSnapshots: [],
    workspaceHealth: null,
    workspaceLoading: false,
    workspacePreview: null,
    workspacePreviewing: false,
    workspaceApplying: false,
    workspaceFormOpen: false,
    workspaceForm: null,

    // AI sub-navigation
    aiSubTab: 'overview',    // 'overview' | 'connect' | 'tools' | 'workspace' | 'skills' | 'mcp' | 'hooks' | 'docs' | 'prompts' | 'data'

    // VPN module
    vpnSubTab: 'overview',   // 'overview' | 'subs' | 'rules' | 'backup' | 'config'
    vpnProfiles: [],
    vpnCurrentProfile: null,
    vpnAppDir: '',
    vpnStatus: null,
    vpnCurrentIP: null,
    vpnIPLoading: false,
    vpnHealthReport: null,
    vpnHealthLoading: false,
    vpnSettings: null,
    vpnSettingsForm: { appDir: '', backupDir: '' },
    vpnPathValidation: null,
    vpnSettingsSaving: false,
    vpnConfigFile: 'verge.yaml',
    vpnConfigContent: '',
    vpnConfigSaving: false,
    vpnConfigError: '',
    vpnBackups: [],
    vpnBackupDir: '',
    vpnBackupPreview: null,
    vpnExternalZipPath: '',
    vpnBackupBusy: false,
    vpnRestoreBusy: false,
    vpnMsg: '',
    vpnStatusRefreshing: false,
    vpnDelayTesting: false,
    // 订阅 (subscriptions overview)
    vpnSubs: null,           // { subscriptions, global, files }
    vpnSubsLoading: false,
    vpnSubsError: '',
    // 自定义规则 (visual editor over the current subscription's rules/merge files)
    vpnRules: null,          // RulesState from GetVPNRules
    vpnRulesDraft: null,     // editable copy of vpnRules.draft
    vpnRulesSnapshot: '',    // JSON of the last loaded draft, for dirty tracking
    vpnRulesLoading: false,
    vpnRulesSaving: false,
    vpnRulesError: '',
    vpnRulesIssues: [],      // [{section, index, message}] from SaveVPNRules
    vpnRulesSection: 'rules', // rules | hosts | fakeip | dns | tun
    vpnRuleTypeHelpOpen: false,
    vpnAppendOpen: false,
    vpnAI: { text: '', busy: false, error: '', preview: null },  // 一句话改规则 (SuggestVPNRuleChange)
    vpnEnhAI: null,          // { row, sub, info, text, busy, error, preview, applying } — AI edit of one enhancement file
    vpnBulk: { open: false, text: '', target: 'DIRECT', busy: false },
    vpnRuleTest: { input: '', busy: false, result: null, error: '' },

    // Memo module: todos (today / upcoming / done) + word book + flashcard review
    memoSubTab: 'today',      // today | upcoming | done | words | review | storage
    memoDashboard: { today: 0, overdue: 0, reminders: 0, reviewDue: 0, totalMemos: 0, totalVocab: 0 },
    memos: [],
    memoSearch: '',
    memoQuick: '',            // quick-add text, parsed by parseQuickTodo
    memoDraft: null,          // editable copy of the selected todo; fields auto-save
    memoDraftSaving: false,
    memoAttachmentForm: { path: '' },
    memoSettings: { filesDir: '', retainDays: 90, autoCleanup: true },
    memoSettingsSaving: false,
    memoPreviewCache: {},
    vocabEntries: [],
    vocabSearch: '',
    vocabFilter: 'all',       // all | due | new | mastered
    vocabLimit: 40,
    vocabQuick: '',
    vocabDraft: null,         // selected or new word (explicit save)
    vocabDraftSnapshot: '',
    vocabLookupBusy: false,
    vocabAIBusy: false,
    vocabSaving: false,
    review: { queue: [], index: 0, revealed: false, tally: { hard: 0, known: 0, mastered: 0 } },

    // App data paths config
    dataPaths: {},           // AppDataPaths from Go
    dataPathsForm: { traceLogDir: '', skillsDir: '' },
    dataPathsSaving: false,

    // Doc links (AI > Docs quick-open documents)
    docLinks: [],
    docLinkForm: { path: '', label: '' },
    docLinksSaving: false,
    editingDocLinkIdx: -1,
    editingDocLinkForm: { path: '', label: '' },

    // Global search / replace overlay
    searchOpen: false,
    searchReplaceMode: false,
    searchQuery: '',
    searchReplaceText: '',
    searchMatches: [],
    searchCurrentIdx: -1,
    searchActiveEl: null,

    // ── Computed ───────────────────────────────────────────────────────
    get scriptDirs() {
      return (this.cfg.directories || []).filter(d => d.isScriptDir);
    },

    get dirScriptCount() {
      const counts = {};
      for (const s of this.allScripts) {
        const dir = this.cfg.directories?.find(d =>
          s.path.startsWith(d.path + '/') || s.path.startsWith(d.path + '\\')
        );
        if (dir) counts[dir.path] = (counts[dir.path] || 0) + 1;
      }
      return counts;
    },

    // ── Workbench (工作台) ─────────────────────────────────────────────
    get wbFilterOptions() {
      const all = this.allScripts;
      return [
        { id: 'all', label: '全部', count: 0 },
        { id: 'fav', label: '★ 收藏', count: all.filter(s => s.favorite).length },
        { id: 'service', label: '服务', count: all.filter(s => s.type === 'service').length },
        { id: 'scheduled', label: '定时', count: all.filter(s => s.schedule).length },
      ].filter(f => f.id === 'all' || f.count > 0);
    },

    // Scripts matching the search + filter, ranked by fuzzy score when searching.
    get wbVisibleScripts() {
      const q = this.wbQuery.trim();
      let list = this.allScripts.filter(s =>
        this.wbFilter === 'fav' ? s.favorite :
        this.wbFilter === 'service' ? s.type === 'service' :
        this.wbFilter === 'scheduled' ? !!s.schedule : true);
      if (!q) return list;
      return list
        .map(s => ({ s, score: this.fuzzyScore(q, this.wbSearchText(s)) }))
        .filter(x => x.score > 0)
        .sort((a, b) => b.score - a.score)
        .map(x => x.s);
    },

    // List groups: 收藏 first, then one group per script directory.
    // While searching, a single ranked "结果" group keeps the best match on top.
    get wbGroups() {
      const list = this.wbVisibleScripts;
      if (this.wbQuery.trim()) return list.length ? [{ key: 'q', label: '搜索结果', items: list }] : [];
      const byName = (a, b) => (a.name || '').localeCompare(b.name || '', 'zh');
      const groups = [];
      const favs = this.wbFilter === 'fav' ? [] : list.filter(s => s.favorite).sort(byName);
      if (favs.length) groups.push({ key: 'fav', label: '★ 收藏', items: favs });
      for (const d of this.scriptDirs) {
        const items = list.filter(s => this.wbDirOf(s.path) === d.path).sort(byName);
        if (items.length) groups.push({ key: d.path, label: d.exportName ? '$' + d.exportName : this.baseName(d.path), items });
      }
      return groups;
    },

    get wbCurrent() {
      return this.wbSelected ? this.allScripts.find(s => s.path === this.wbSelected) || null : null;
    },

    get wbRunningCount() {
      const svc = this.allScripts.filter(s => s.type === 'service' && this.statuses[s.path] === 'running').length;
      return svc + Object.keys(this.runningSet).length;
    },

    get wbCodeDirty() {
      return !!this.wbCode.path && this.wbCode.content !== this.wbCode.original;
    },

    get wbFormDirty() {
      return !!this.wbFormSnapshot && JSON.stringify(this.editForm) !== this.wbFormSnapshot;
    },

    get wbScriptEnvList() {
      return this.allScripts.filter(s => Object.keys(s.env || {}).length > 0);
    },

    get envHasIssues() {
      return this.envRows.some(e => this.envIssue(e));
    },

    get envImportPreview() {
      return this.parseDotenv(this.envImportText);
    },

    // ⌘K items: scripts, their presets, and a few navigation actions.
    get paletteItems() {
      const q = this.palette.query.trim();
      const items = [];
      for (const s of this.allScripts) {
        items.push({ key: 's:' + s.path, kind: 'script', script: s, icon: s.favorite ? '★' : this.langIcon(s.language),
          title: s.name, sub: s.desc || this.wbDisplayPath(s), hint: s.type === 'service' ? '服务' : '运行',
          text: this.wbSearchText(s), boost: s.favorite ? 2 : 0 });
        for (const p of this.normalizedPresets(s)) {
          items.push({ key: 'p:' + s.path + ':' + p.name, kind: 'preset', script: s, preset: p, icon: '🔖',
            title: s.name + ' › ' + p.name, sub: p.args, hint: '运行预设', text: s.name + ' ' + p.name + ' ' + p.args, boost: 0 });
        }
      }
      items.push({ key: 'a:new', kind: 'action', action: 'new', icon: '＋', title: '新建脚本', sub: '从模板或 AI 草稿创建', hint: '', text: '新建脚本 new script create', boost: -1 });
      items.push({ key: 'a:env', kind: 'action', action: 'env', icon: '🔑', title: '环境变量', sub: '编辑 ~/.zshrc 中的 Hatch 变量', hint: '', text: '环境变量 env export path', boost: -1 });
      if (!q) {
        const recent = new Map(this.wbRecentRuns.map((r, i) => [r.script, i]));
        return items.filter(it => it.kind !== 'preset')
          .sort((a, b) => (b.boost - a.boost) || ((recent.get(a.script?.path) ?? 999) - (recent.get(b.script?.path) ?? 999)))
          .slice(0, 50);
      }
      return items
        .map(it => { const m = this.fuzzyScore(q, it.text); return { it, score: m && m + (it.boost > 0 ? 1 : 0) }; })  // boost ranks matches, never adds non-matches
        .filter(x => x.score > 0)
        .sort((a, b) => b.score - a.score)
        .slice(0, 50)
        .map(x => x.it);
    },

    scriptTemplates: [
      { id: 'zsh', label: 'zsh 脚本', hint: '日常自动化，带参数与错误处理', ext: '.sh', body:
`#!/bin/zsh
# @hatch.name: {{name}}
# @hatch.desc: 一句话说明这个脚本做什么
# @hatch.arg: 目标 | | | required
set -euo pipefail

target="\${1:?用法: {{file}} <目标>}"
echo "处理: $target"
` },
      { id: 'python', label: 'Python 脚本', hint: 'argparse 参数，适合数据处理', ext: '.py', body:
`#!/usr/bin/env python3
# @hatch.name: {{name}}
# @hatch.desc: 一句话说明这个脚本做什么
# @hatch.arg: 输入文件 | | | required
# @hatch.arg: 输出目录 | --out | . | optional
import argparse


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("input")
    parser.add_argument("--out", default=".")
    args = parser.parse_args()
    print(f"input={args.input} out={args.out}")


if __name__ == "__main__":
    main()
` },
      { id: 'node', label: 'Node 脚本', hint: '调用 HTTP API、处理 JSON', ext: '.js', body:
`#!/usr/bin/env node
// @hatch.name: {{name}}
// @hatch.desc: 一句话说明这个脚本做什么
// @hatch.arg: URL | | https://example.com | optional
const url = process.argv[2] || 'https://example.com';

const res = await fetch(url);
console.log(res.status, (await res.text()).slice(0, 200));
` },
      { id: 'service', label: '常驻服务', hint: 'start / stop / status，Hatch 管理进程', ext: '.sh', body:
`#!/bin/zsh
# @hatch.name: {{name}}
# @hatch.desc: 启停一个后台进程
# @hatch.type: service
# @hatch.commands: start, stop, restart, status
case "\${1:-}" in
  start)   exec python3 -m http.server 8000 ;;
  stop)    echo "Hatch 会向进程组发送 SIGTERM" ;;
  restart) "$0" stop; exec "$0" start ;;
  status)  echo "running" ;;
  *)       echo "用法: $0 {start|stop|restart|status}"; exit 2 ;;
esac
` },
    ],

    // ── Memo: todos ────────────────────────────────────────────────────
    get memoOverdue() {
      const today = this.todayDate();
      return this.memos.filter(m => m.status !== 'done' && m.date && m.date < today)
        .sort((a, b) => a.date.localeCompare(b.date) || (a.reminderAt || 0) - (b.reminderAt || 0));
    },

    get memoTodayOpen() {
      const today = this.todayDate();
      return this.memos.filter(m => m.status !== 'done' && (m.date === today || !m.date)).sort(this.memoOrder);
    },

    get memoTodayCount() {
      return this.memoOverdue.length + this.memoTodayOpen.length;
    },

    get memoUpcoming() {
      const today = this.todayDate();
      return this.memos.filter(m => m.status !== 'done' && m.date > today)
        .sort((a, b) => a.date.localeCompare(b.date) || this.memoOrder(a, b));
    },

    get memoQuickParsed() {
      return this.parseQuickTodo(this.memoQuick);
    },

    // Groups for the current todo view; a search query spans every todo.
    get memoGroups() {
      const q = this.memoSearch.trim().toLowerCase();
      if (q) {
        const hits = this.memos.filter(m => [m.title, m.body, ...(m.tags || [])].some(v => (v || '').toLowerCase().includes(q)))
          .sort((a, b) => (a.status === 'done') - (b.status === 'done') || b.date.localeCompare(a.date));
        return hits.length ? [{ key: 'q', label: '搜索结果', items: hits, showDate: true }] : [];
      }
      const today = this.todayDate();
      if (this.memoSubTab === 'today') {
        const groups = [];
        if (this.memoOverdue.length) groups.push({ key: 'overdue', label: '已过期', tone: 'warn', items: this.memoOverdue, showDate: true });
        if (this.memoTodayOpen.length) groups.push({ key: 'today', label: '今天', items: this.memoTodayOpen });
        const doneToday = this.memos.filter(m => m.status === 'done' && m.date === today);
        if (doneToday.length) groups.push({ key: 'done', label: '今天已完成', tone: 'muted', items: doneToday });
        return groups;
      }
      const byDate = new Map();
      const list = this.memoSubTab === 'upcoming'
        ? this.memoUpcoming
        : this.memos.filter(m => m.status === 'done').sort((a, b) => b.date.localeCompare(a.date) || (b.updatedAt || 0) - (a.updatedAt || 0)).slice(0, 200);
      for (const m of list) {
        if (!byDate.has(m.date)) byDate.set(m.date, { key: m.date, label: this.memoDateLabel(m.date, true), items: [] });
        byDate.get(m.date).items.push(m);
      }
      return [...byDate.values()];
    },

    // ── Memo: words ────────────────────────────────────────────────────
    get vocabFilterOptions() {
      const all = this.vocabEntries;
      return [
        { id: 'all', label: '全部', count: all.length },
        { id: 'due', label: '待复习', count: this.reviewDueEntries.length },
        { id: 'new', label: '新词', count: all.filter(v => !v.lastReviewedAt).length },
        { id: 'mastered', label: '已掌握', count: all.filter(v => (v.reviewLevel || 0) >= 6).length },
      ];
    },

    get filteredVocabEntries() {
      const q = this.vocabSearch.trim().toLowerCase();
      const now = Math.floor(Date.now() / 1000);
      return this.vocabEntries.filter(v => {
        if (this.vocabFilter === 'due' && v.nextReviewAt && v.nextReviewAt > now) return false;
        if (this.vocabFilter === 'new' && v.lastReviewedAt) return false;
        if (this.vocabFilter === 'mastered' && (v.reviewLevel || 0) < 6) return false;
        if (!q) return true;
        return [v.word, v.chinese, v.example, ...(v.tags || [])].some(x => (x || '').toLowerCase().includes(q));
      }).sort((a, b) => {
        if (q) {
          const aw = (a.word || '').toLowerCase().startsWith(q), bw = (b.word || '').toLowerCase().startsWith(q);
          if (aw !== bw) return aw ? -1 : 1;
        }
        return (b.createdAt || 0) - (a.createdAt || 0);
      });
    },

    get visibleVocab() {
      return this.filteredVocabEntries.slice(0, this.vocabLimit);
    },

    get vocabDraftDirty() {
      return !!this.vocabDraft && JSON.stringify(this.vocabDraft) !== this.vocabDraftSnapshot;
    },

    get vocabDraftDuplicate() {
      const d = this.vocabDraft;
      if (!d || d.id) return null;
      const w = (d.word || '').trim().toLowerCase();
      return w ? this.vocabEntries.find(v => (v.word || '').toLowerCase() === w) || null : null;
    },

    get reviewCard() {
      const r = this.review;
      if (!r.queue.length || r.index >= r.queue.length) return null;
      return this.vocabEntries.find(v => v.id === r.queue[r.index]) || null;
    },

    get reviewDueEntries() {
      const now = Math.floor(Date.now() / 1000);
      return this.vocabEntries.filter(v => !v.nextReviewAt || v.nextReviewAt <= now);
    },

    get totalScriptCount() {
      return this.allScripts.length;
    },

    get serviceScriptCount() {
      return this.allScripts.filter(s => s.type === 'service').length;
    },

    get oneshotScriptCount() {
      return this.allScripts.filter(s => s.type === 'oneshot').length;
    },

    get exportCount() {
      return (this.cfg.exports || []).filter(e => e.name).length;
    },

    get directoryCount() {
      return (this.cfg.directories || []).length;
    },

    get aiReady() {
      if (!this.aiForm.enabled) return false;
      if (this.aiProviders.length > 0 && this.activeProviderID) {
        const p = this.aiProviders.find(p => p.id === this.activeProviderID);
        return !!(p && p.baseUrl && p.apiKey && p.model);
      }
      return !!(this.aiForm.baseUrl && this.aiForm.apiKey && this.aiForm.model);
    },

    get activeProvider() {
      return this.aiProviders.find(p => p.id === this.activeProviderID) || null;
    },

    get quickAIActiveLabel() {
      const p = this.activeProvider;
      if (!p) return '';
      const label = p.name || '未命名';
      return p.model ? `✦ ${label} · ${p.model}` : `✦ ${label}`;
    },

    async toggleQuickAI() {
      this.quickAIOpen = !this.quickAIOpen;
      if (this.quickAIOpen) {
        this.quickAIActiveID = this.activeProviderID;
        // refresh providers + traces lazily
        try { if (!this.aiProviders.length) await this.loadAIProviders(); } catch (_) {}
        try { await this.loadAITraces(); } catch (_) {}
      }
    },

    async switchQuickAIProvider() {
      if (!this.quickAIActiveID || this.quickAIActiveID === this.activeProviderID) return;
      try {
        await SaveAIProviders(this.aiProviders, this.quickAIActiveID);
        this.activeProviderID = this.quickAIActiveID;
        const p = this.activeProvider;
        // Keep legacy aiForm in sync so the connect page reflects switch
        if (p && this.aiForm) {
          this.aiForm.baseUrl = p.baseUrl;
          this.aiForm.apiKey  = p.apiKey;
          this.aiForm.model   = p.model;
        }
        this.showSuccess(`已切换到 ${p ? (p.name || p.model) : this.quickAIActiveID}`);
      } catch (e) {
        this.showError(`切换失败: ${e}`);
        this.quickAIActiveID = this.activeProviderID;
      }
    },

    // ── Lifecycle ──────────────────────────────────────────────────────
    async init() {
      this.applyInitialRoute();
      await this.loadModules({ init: true });
      await this.loadConfig();
      await this.loadRecentRuns();
      await this.scanAll();
      // Selection can change via list clicks, ⌘K, ↑/↓ or a restored state.
      this.$watch('wbSelected', () => this.wbCheckHealth(this.wbCurrent));
      // VPN code reports through vpnMsg; show it as a toast like every other
      // module instead of a banner that pushes the page down.
      this.$watch('vpnMsg', (v) => {
        if (!v) return;
        if (/失败|未保存|请输入|错误/.test(v)) this.showError(v); else this.showSuccess(v);
        this.vpnMsg = '';
      });
      this.wbCheckHealth(this.wbCurrent);
      if (this.tab === 'scripts' && this.wbSubTab === 'env') this.envLoad();
      this.loadDataPaths();
      this.loadDocLinks();
      this.loadAIProviders();
      if (this.tab === 'ai') { this.loadAIHeroStats(); this.aiGo(this.aiSubTab); }
      if (this.tab === 'memo') this.loadMemoModule();
      if (this.tab === 'vpn') { this.loadVPNOverview(); if (this.vpnSubTab !== 'overview') this.vpnGo(this.vpnSubTab); }
      if (this.tab === 'host') this.openHost();
      this.setupEvents();
      this.setupModuleReload();
      this.setupSearchShortcuts();
      this.setupWorkbenchKeys();
      setInterval(() => this.refreshStatuses(), 3000);
      if (this.modules.vpn) {
        try { this.vpnStatus = await GetVPNStatus(); } catch (_) {}
      }
    },

    // ── Modules (config.json "modules") ───────────────────────────────────
    // tab ids map to module keys; 'scripts' is the 工作台 (workbench) module.
    moduleOn(tab) {
      const key = tab === 'scripts' ? 'workbench' : tab;
      return this.modules[key] !== false;
    },

    openTab(tab, sub) {
      if (!this.moduleOn(tab)) return;
      if (this.hostTimer) { clearTimeout(this.hostTimer); this.hostTimer = null; }
      this.tab = tab;
      if (tab === 'ai') { this.loadAITraces(); this.loadAIHeroStats(); this.aiGo(sub || this.aiSubTab); }
      else if (tab === 'vpn') this.loadVPNOverview();
      else if (tab === 'memo') this.loadMemoModule();
      else if (tab === 'host') this.openHost();
    },

    // Switches the AI sub-page and (re)loads its data, so every entry point —
    // sidebar, overview cards, top tab, header refresh — shows fresh numbers.
    aiGo(sub) {
      this.aiSubTab = sub;
      const loaders = {
        leaderboards: () => this.loadLeaderboards(),
        overview: () => this.loadOverview(),
        connect: () => this.loadAIProviders(),
        tools: () => this.loadAIUsageStatuses(),
        workspace: () => this.loadWorkspace(),
        skills: () => this.loadSkills(),
        mcp: () => this.loadMCP(),
        hooks: () => this.loadHooks(),
        docs: () => this.loadDocLinks(),
        prompts: () => this.loadPromptTemplates(),
        data: () => this.loadDataInventory(),
      };
      loaders[sub]?.();
    },

    // Hero stat cards (Skills / MCP / Hooks) are visible on every AI sub-page.
    async loadAIHeroStats() {
      try { this.skillsStats = await GetSkillsStats(); } catch (_) {}
      try { this.hooksStats = await GetHooksStats(); } catch (_) {}
      try { if (!this.mcpInventory) await this.loadMCP(); } catch (_) {}
    },

    // Re-reads module visibility from config.json; if the current tab was
    // hidden, falls back to the first visible module. During init only the
    // tab is switched — init() itself loads the landing page.
    async loadModules({ init = false } = {}) {
      try {
        const m = await GetModuleSettings();
        if (m) {
          this.modules = {
            workbench: m.workbench !== false,
            ai: m.ai !== false,
            vpn: m.vpn !== false,
            memo: m.memo !== false,
            host: m.host !== false,
          };
        }
      } catch (_) {}
      if (this.moduleOn(this.tab)) return;
      const next = ['ai', 'scripts', 'vpn', 'memo', 'host'].find(t => this.moduleOn(t)) || 'scripts';
      if (init) {
        this.tab = next;
      } else {
        this.openTab(next);
      }
    },

    // Apply config.json edits when the window regains focus. The handler is
    // kept on window so HMR re-inits replace it instead of stacking listeners.
    setupModuleReload() {
      if (window.__hatchModuleReload) window.removeEventListener('focus', window.__hatchModuleReload);
      window.__hatchModuleReload = () => this.loadModules();
      window.addEventListener('focus', window.__hatchModuleReload);
    },

    applyInitialRoute() {
      const params = new URLSearchParams(window.location.search || '');
      const ai = params.get('aiSubTab') || params.get('ai');
      const memo = params.get('memoSubTab') || params.get('memo');
      const vpn = params.get('vpnSubTab') || params.get('vpn');
      const allowedAI = ['leaderboards', 'overview', 'connect', 'tools', 'workspace', 'skills', 'mcp', 'hooks', 'docs', 'prompts', 'data'];
      const allowedMemo = ['today', 'upcoming', 'done', 'words', 'review', 'storage'];
      if (allowedAI.includes(ai)) {
        this.tab = 'ai';
        this.aiSubTab = ai;
      }
      if (['overview', 'subs', 'rules', 'backup', 'config'].includes(vpn)) {
        this.tab = 'vpn';
        this.vpnSubTab = vpn;
      }
      if (allowedMemo.includes(memo)) {
        this.tab = 'memo';
        this.memoSubTab = memo;
      }
      const wb = params.get('wbSubTab') || params.get('wb');
      if (['scripts', 'env', 'dirs', 'runs'].includes(wb)) {
        this.tab = 'scripts';
        this.wbSubTab = wb;
      }
      if (this.aiSubTab === 'mcp' && params.get('mcpAdd') === '1') {
        this.openMCPAdd(params.get('mcpAddTab') || 'paste');
      }
    },

    async loadConfig() {
      try {
        this.cfg = await GetConfig() || { directories: [], exports: [] };
        this.aiForm = {
          enabled: !!this.cfg.ai?.enabled,
          baseUrl: this.cfg.ai?.baseUrl || 'https://api.openai.com/v1',
          apiKey: this.cfg.ai?.apiKey || '',
          model: this.cfg.ai?.model || '',
          traceRetainDays: this.cfg.ai?.traceRetainDays || 3,
        };
      } catch (e) { this.showError(`加载配置失败: ${e}`);   }
      try { this.zshrcInfo = await CheckZshrc(); } catch (_) {}
      try { this.autoLaunchEnabled = await GetAutoLaunchEnabled(); } catch (_) {
        this.autoLaunchEnabled = !!this.cfg.autoLaunch;
      }
    },

    async scanAll() {
      this.loading = true;
      try {
        this.allScripts = await ScanAll() || [];
        this.wbEnsureSelection();
        await this.refreshStatuses();
      } catch (e) { this.showError(`扫描失败: ${e}`);   }
      finally { this.loading = false; }
    },

    // ── Workbench: navigation + selection ─────────────────────────────
    wbGo(sub) {
      this.wbSubTab = sub;
      if (sub === 'env') { if (!this.envDirty) this.envLoad(); }
      else if (sub === 'runs') this.loadRecentRuns();
    },

    wbDirOf(path) {
      const d = this.scriptDirs.find(d => path.startsWith(d.path + '/'));
      return d ? d.path : '';
    },

    wbSearchText(s) {
      return [s.name, this.baseName(s.path), s.desc, (s.tags || []).join(' '), s.language].filter(Boolean).join(' ');
    },

    // Subsequence match; contiguous runs and word starts score higher. 0 = no match.
    fuzzyScore(query, text) {
      const q = query.toLowerCase().replace(/\s+/g, '');
      const t = String(text || '').toLowerCase();
      if (!q) return 1;
      const direct = t.indexOf(q);
      if (direct >= 0) return 100 - Math.min(direct, 50) + q.length;
      let score = 0, ti = 0, run = 0;
      for (const ch of q) {
        const found = t.indexOf(ch, ti);
        if (found < 0) return 0;
        run = found === ti ? run + 1 : 1;
        score += run + (found === 0 || /[\s\-_./]/.test(t[found - 1]) ? 2 : 0);
        ti = found + 1;
      }
      return score;
    },

    async wbSelect(script, pane) {
      if (!script) return;
      if (this.wbSelected !== script.path) {
        if ((this.wbCodeDirty || this.wbFormDirty) &&
            !(await this.askConfirm('当前脚本有未保存的修改，切换后会丢失。', { title: '放弃修改？', confirmText: '放弃修改' }))) return;
        this.wbSelected = script.path;
        this.wbCode = { path: '', content: '', original: '', loading: false, saving: false };
        this.wbAIDiscard();
        this.wbAI.instruction = '';
        this.wbFormSnapshot = '';
        this.$nextTick(() => document.querySelector('.wb-item.active')?.scrollIntoView({ block: 'nearest' }));
      }
      this.wbOpenPane(pane || this.wbPane);
    },

    // Only scripts that declare dependencies or a health URL are checked.
    async wbCheckHealth(s, force = false) {
      if (!s || !((s.requires || []).length || s.healthUrl)) { this.wbHealth = { path: s?.path || '', result: null, loading: false }; return; }
      if (!force && this.wbHealth.path === s.path && this.wbHealth.result) return;
      this.wbHealth = { path: s.path, result: null, loading: true };
      try {
        const result = await window['go']['app']['App']['CheckScriptHealth'](s.path);
        if (this.wbHealth.path === s.path) this.wbHealth = { path: s.path, result, loading: false };
      } catch (_) {
        if (this.wbHealth.path === s.path) this.wbHealth.loading = false;
      }
    },

    wbMoveSelection(delta) {
      const flat = this.wbGroups.flatMap(g => g.items);
      if (!flat.length) return;
      const i = flat.findIndex(s => s.path === this.wbSelected);
      const next = flat[Math.max(0, Math.min(flat.length - 1, i + delta))] || flat[0];
      this.wbSelect(next);
      this.$nextTick(() => document.querySelector('.wb-item.active')?.scrollIntoView({ block: 'nearest' }));
    },

    wbOpenPane(pane) {
      this.wbPane = pane;
      const s = this.wbCurrent;
      if (!s) return;
      if (pane === 'code' && this.wbCode.path !== s.path) this.wbLoadCode(s);
      if (pane === 'config' && !this.wbFormSnapshot) this.wbLoadForm(s);
      if (pane === 'history') this.openHistory(s);
    },

    wbEnsureSelection() {
      if (this.wbCurrent || !this.allScripts.length) return;
      const recent = this.wbRecentRuns.find(r => this.allScripts.some(s => s.path === r.script));
      const first = this.wbGroups[0]?.items[0];
      this.wbSelected = recent ? recent.script : first?.path || null;
    },

    wbStatusClass(s) {
      if (this.runningSet[s.path] || this.statuses[s.path] === 'running') return 'is-running';
      const code = this.scriptExitCodes[s.path] ?? this.wbLastRun(s.path)?.exitCode;
      if (code === undefined) return '';
      return code === 0 ? 'is-ok' : 'is-err';
    },

    wbLastRun(path) {
      return this.wbRecentRuns.find(r => r.script === path);
    },

    wbScriptName(path) {
      return this.allScripts.find(s => s.path === path)?.name || this.baseName(path);
    },

    wbOpenRun(r) {
      const s = this.allScripts.find(x => x.path === r.script);
      if (!s) { this.showError('脚本已不在扫描目录中'); return; }
      this.wbGo('scripts');
      this.wbSelect(s, 'history');
    },

    async loadRecentRuns() {
      try { this.wbRecentRuns = await window['go']['app']['App']['ListRecentRunsSorted'](80) || []; } catch (_) {}
    },

    wbDisplayPath(s) {
      const dir = this.scriptDirs.find(d => s.path.startsWith(d.path + '/'));
      return dir?.exportName ? '$' + dir.exportName + s.path.slice(dir.path.length) : s.path.replace(/^\/Users\/[^/]+/, '~');
    },

    // Command the user can paste into a terminal. Uses the directory's PATH
    // variable when it has one so the line keeps working if the folder moves.
    wbShellCommand(s, withArgs = false) {
      const dir = this.scriptDirs.find(d => s.path.startsWith(d.path + '/'));
      const cmd = dir?.exportName ? `"$${dir.exportName}/${this.baseName(s.path)}"` : this.shellArg(s.path);
      if (!withArgs && !this.wbHasArgs(s)) return cmd;
      const args = this.wbArgs(s, { lenient: true }).args.map(a => this.shellArg(a));
      return [cmd, ...args].join(' ');
    },

    wbHasArgs(s) {
      return (s.args || []).length > 0 || !!(this.scriptArgs[s.path] || '').trim();
    },

    shellArg(v) {
      return /^[A-Za-z0-9_\-+=.,/:@%]+$/.test(v) ? v : this.shellSingleQuote(v);
    },

    revealFile(path) {
      window['go']['app']['App']['RevealFile'](path).catch(e => this.showError(`打开 Finder 失败: ${e}`));
    },

    async toggleFavorite(s) {
      const next = !s.favorite;
      try {
        await window['go']['app']['App']['SetScriptFavorite'](s.path, next);
        s.favorite = next;
        this.allScripts = [...this.allScripts];
      } catch (e) { this.showError(`收藏失败: ${e}`); }
    },

    async trashScript(s) {
      if (!(await this.askConfirm(`把「${this.baseName(s.path)}」移到废纸篓？\n可以在 Finder 的废纸篓中恢复，Hatch 中的配置会一并清除。`, { title: '删除脚本', confirmText: '移到废纸篓' }))) return;
      try {
        await window['go']['app']['App']['TrashScriptFile'](s.path);
        this.wbSelected = null;
        this.wbFormSnapshot = '';
        this.wbCode = { path: '', content: '', original: '', loading: false, saving: false };
        await this.scanAll();
        this.showSuccess('已移到废纸篓');
      } catch (e) { this.showError(`删除失败: ${e}`); }
    },

    // ── Workbench: run ────────────────────────────────────────────────
    wbArgValue(path, i) {
      const v = this.wbArgValues[path]?.[i];
      return v === undefined ? '' : v;
    },

    wbSetArgValue(path, i, value) {
      this.wbArgValues = { ...this.wbArgValues, [path]: { ...(this.wbArgValues[path] || {}), [i]: value } };
    },

    // Build argv from declared fields (flag + value, or positional) or the free-text box.
    // An empty field is omitted so the script applies its own default (the
    // declared default is only a placeholder; argv is not shell-expanded).
    // Positional args stop at the first empty one to keep later ones in place.
    wbArgs(s, { lenient = false } = {}) {
      const defs = s.args || [];
      if (!defs.length) return { args: this.splitArgs(this.scriptArgs[s.path] || ''), missing: [] };
      const args = [], missing = [];
      let positionalGap = false;
      defs.forEach((d, i) => {
        const v = this.wbArgValue(s.path, i).trim() ? this.wbArgValue(s.path, i) : '';
        if (!v) {
          if (d.required) { missing.push(d.name); if (lenient) args.push(`<${d.name}>`); }
          if (!d.flag) positionalGap = true;
          return;
        }
        if (d.flag) args.push(d.flag, v);
        else if (!positionalGap) args.push(v);
      });
      return { args, missing };
    },

    async wbRun(s, argsOverride = null) {
      if (!s || this.runningSet[s.path]) return;
      let args = argsOverride;
      if (!args) {
        const built = this.wbArgs(s);
        if (built.missing.length) { this.showError(`请填写必填参数：${built.missing.join('、')}`); return; }
        args = built.args;
      }
      const id = s.path;
      this.runningSet = { ...this.runningSet, [id]: true };
      this.scriptLogs[id] = [];
      const codes = { ...this.scriptExitCodes };
      delete codes[id];
      this.scriptExitCodes = codes;
      try {
        await RunScript(id, s.path, args);
      } catch (e) {
        this.showError(`运行失败: ${e}`);
        const next = { ...this.runningSet };
        delete next[id];
        this.runningSet = next;
      }
    },

    wbRunPreset(s, p) {
      const action = this.presetAction(p);
      if (s.type === 'service' && action) {
        this.serviceAction(s, action, this.splitArgs(p.args).slice(1));
        return;
      }
      this.wbPane = 'run';
      this.wbRun(s, this.splitArgs(p.args));
    },

    wbServiceCommands(s) {
      const cmds = (s.commands || []).length ? s.commands : ['start', 'stop', 'restart'];
      return cmds;
    },

    wbRelTime(t) {
      const ms = typeof t === 'number' ? t : Date.parse(t);
      if (!ms) return '';
      const diff = (Date.now() - ms) / 1000;
      if (diff < 60) return '刚刚';
      if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前';
      if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前';
      if (diff < 86400 * 7) return Math.floor(diff / 86400) + ' 天前';
      return new Date(ms).toLocaleDateString();
    },

    wbDuration(ms) {
      if (ms == null) return '';
      if (ms < 1000) return ms + 'ms';
      if (ms < 60000) return (ms / 1000).toFixed(1) + 's';
      return Math.floor(ms / 60000) + 'm' + Math.round((ms % 60000) / 1000) + 's';
    },

    // ── Workbench: code editor + AI diff ──────────────────────────────
    async wbLoadCode(s) {
      this.wbCode = { path: s.path, content: '', original: '', loading: true, saving: false, blocked: '', size: 0 };
      try {
        const f = await ReadScriptFile(s.path);
        const blocked = f.binary ? 'binary' : f.truncated ? 'large' : '';
        const text = f.content || '';
        if (this.wbCode.path === s.path) this.wbCode = { ...this.wbCode, content: text, original: text, loading: false, blocked, size: f.size };
      } catch (e) {
        this.wbCode.loading = false;
        this.showError(`读取文件失败: ${e}`);
      }
    },

    async wbSaveCode() {
      if (!this.wbCodeDirty || this.wbCode.saving) return;
      this.wbCode.saving = true;
      try {
        await WriteScriptFile(this.wbCode.path, this.wbCode.content);
        this.wbCode.original = this.wbCode.content;
        await this.scanAll();
        this.showSuccess('脚本已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.wbCode.saving = false;
      }
    },

    async wbRevertCode() {
      if (!(await this.askConfirm('撤销编辑器中所有未保存的修改？', { title: '还原', confirmText: '还原' }))) return;
      this.wbCode.content = this.wbCode.original;
    },

    wbInsertTab(e) {
      const el = e.target;
      const { selectionStart: a, selectionEnd: b, value } = el;
      el.value = value.slice(0, a) + '  ' + value.slice(b);
      el.selectionStart = el.selectionEnd = a + 2;
      el.dispatchEvent(new Event('input'));
    },

    async wbAIGenerate() {
      const s = this.wbCurrent;
      if (!s || !this.wbAI.instruction.trim()) return;
      if (!this.aiReady) { this.showError('请先在 ✦ AI → 连接 中配置并启用 AI'); return; }
      if (this.wbCode.path !== s.path) await this.wbLoadCode(s);
      if (this.wbCode.blocked) { this.showError(this.wbCode.blocked === 'binary' ? '二进制文件无法用 AI 修改' : '文件过大，无法用 AI 修改'); return; }
      this.wbAI.busy = true;
      try {
        const result = await SuggestScriptContent(s.path, this.wbCode.content, this.wbAI.instruction);
        const diff = this.lineDiff(this.wbCode.content, result);
        this.wbAI = {
          ...this.wbAI, result,
          diff: this.compactDiff(diff),
          stats: { add: diff.filter(d => d.type === 'add').length, del: diff.filter(d => d.type === 'del').length },
        };
      } catch (e) {
        this.showError(`AI 生成失败: ${e}`);
      } finally {
        this.wbAI.busy = false;
      }
    },

    wbAIAccept() {
      if (this.wbAI.result === null) return;
      this.wbCode.content = this.wbAI.result;
      this.wbAI = { ...this.wbAI, result: null, diff: [], instruction: '' };
      this.showSuccess('已放入编辑器，确认后点「保存」写入文件');
    },

    wbAIDiscard() {
      this.wbAI = { ...this.wbAI, result: null, diff: [], stats: { add: 0, del: 0 } };
    },

    // Line-level LCS diff → [{type: same|add|del, text}]
    lineDiff(a, b) {
      const x = String(a || '').split('\n'), y = String(b || '').split('\n');
      const n = x.length, m = y.length;
      if (n * m > 4_000_000) return [...x.map(t => ({ type: 'del', text: t })), ...y.map(t => ({ type: 'add', text: t }))];
      const dp = Array.from({ length: n + 1 }, () => new Uint16Array(m + 1));
      for (let i = n - 1; i >= 0; i--)
        for (let j = m - 1; j >= 0; j--)
          dp[i][j] = x[i] === y[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1]);
      const out = [];
      let i = 0, j = 0;
      while (i < n && j < m) {
        if (x[i] === y[j]) { out.push({ type: 'same', text: x[i] }); i++; j++; }
        else if (dp[i + 1][j] >= dp[i][j + 1]) out.push({ type: 'del', text: x[i++] });
        else out.push({ type: 'add', text: y[j++] });
      }
      while (i < n) out.push({ type: 'del', text: x[i++] });
      while (j < m) out.push({ type: 'add', text: y[j++] });
      return out;
    },

    // Keep 2 lines of context around changes; fold the rest into "… N 行未改动".
    compactDiff(diff, ctx = 2) {
      const keep = diff.map(() => false);
      diff.forEach((d, i) => {
        if (d.type === 'same') return;
        for (let k = Math.max(0, i - ctx); k <= Math.min(diff.length - 1, i + ctx); k++) keep[k] = true;
      });
      const out = [];
      let skipped = 0;
      diff.forEach((d, i) => {
        if (keep[i]) {
          if (skipped) out.push({ type: 'gap', text: `… ${skipped} 行未改动` });
          skipped = 0;
          out.push(d);
        } else skipped++;
      });
      if (skipped) out.push({ type: 'gap', text: `… ${skipped} 行未改动` });
      return out;
    },

    // ── Workbench: config form ────────────────────────────────────────
    wbLoadForm(s) {
      const defs = (s.args || []).map(a => ({ name: a.name || '', flag: a.flag || '', default: a.default || '', required: !!a.required }));
      this.editForm = {
        name: s.name || '',
        desc: s.desc || '',
        type: s.type || 'oneshot',
        commandsRaw: (s.commands || []).join(', '),
        pidFile: s.pidFile || '',
        tagsRaw: (s.tags || []).join(', '),
        argMode: defs.length ? 'fields' : (Array.isArray(s.args) ? 'none' : 'free'),
        argDefs: defs,
        presets: this.normalizedPresets(s),
        envPairs: Object.entries(s.env || {}).map(([key, value]) => ({ key, value })),
        notifyOnDone: !!s.notifyOnDone,
        schedule: s.schedule || '',
        requiresRaw: (s.requires || []).join(', '),
        healthUrl: s.healthUrl || '',
        preScriptsRaw: (s.preScripts || []).join('\n'),
      };
      this.wbFormSnapshot = JSON.stringify(this.editForm);
    },

    async wbSaveConfig() {
      const s = this.wbCurrent;
      if (!s) return;
      const f = this.editForm;
      const list = raw => (raw || '').split(',').map(x => x.trim()).filter(Boolean);
      const defs = f.argDefs.filter(a => a.name.trim()).map(a => ({ ...a, name: a.name.trim(), flag: a.flag.trim() }));
      this.editSaving = true;
      try {
        await SaveScriptMeta({
          ...s,
          name: f.name.trim(),
          desc: f.desc.trim(),
          type: f.type,
          commands: f.type === 'service' ? list(f.commandsRaw) : [],
          pidFile: f.type === 'service' ? f.pidFile : '',
          tags: list(f.tagsRaw),
          // nil = free-text box, [] = no args, [...] = generated fields
          args: f.type !== 'oneshot' ? null : f.argMode === 'none' ? [] : f.argMode === 'fields' && defs.length ? defs : null,
          presets: this.normalizePresetList(f.presets),
          env: f.envPairs.reduce((acc, p) => { const k = (p.key || '').trim(); if (k) acc[k] = p.value || ''; return acc; }, {}),
          notifyOnDone: !!f.notifyOnDone,
          schedule: f.schedule.trim(),
          requires: list(f.requiresRaw),
          healthUrl: f.healthUrl.trim(),
          preScripts: f.preScriptsRaw.split('\n').map(x => x.trim()).filter(Boolean),
        });
        await this.scanAll();
        const fresh = this.allScripts.find(x => x.path === s.path);
        if (fresh) { this.wbLoadForm(fresh); this.wbCheckHealth(fresh, true); }
        this.showSuccess('配置已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.editSaving = false;
      }
    },

    async wbResetConfig() {
      const s = this.wbCurrent;
      if (!s) return;
      if (!(await this.askConfirm(`清除「${s.name}」在 Hatch 中的手动配置（参数、预设、变量、定时等），恢复为脚本头部注释与自动检测的结果？`, { title: '恢复自动检测', confirmText: '恢复' }))) return;
      try {
        await DeleteScriptMeta(s.path);
        await this.scanAll();
        const fresh = this.allScripts.find(x => x.path === s.path);
        if (fresh) this.wbLoadForm(fresh);
        this.showSuccess('已恢复自动检测');
      } catch (e) { this.showError(`重置失败: ${e}`); }
    },

    async wbAIFillConfig() {
      const s = this.wbCurrent;
      if (!s) return;
      if (!this.aiReady) { this.showError('请先在 ✦ AI → 连接 中配置并启用 AI'); return; }
      this.aiBusyScript = s.path;
      try {
        const sug = await SuggestScriptMeta(s);
        if (!this.wbFormSnapshot) this.wbLoadForm(s);
        const f = this.editForm;
        f.name = sug.name || f.name;
        f.desc = sug.desc || f.desc;
        f.type = sug.type || f.type;
        if ((sug.commands || []).length) f.commandsRaw = sug.commands.join(', ');
        if (sug.pidFile) f.pidFile = sug.pidFile;
        if ((sug.tags || []).length) f.tagsRaw = sug.tags.join(', ');
        if ((sug.args || []).length) {
          f.argMode = 'fields';
          f.argDefs = sug.args.map(a => ({ name: a.name || '', flag: a.flag || '', default: a.default || '', required: !!a.required }));
        } else if (f.type === 'oneshot' && !f.argDefs.length) {
          f.argMode = sug.needsArgs ? 'free' : 'none';
        }
        const presets = this.normalizePresetList(sug.presets || []);
        if (presets.length) f.presets = presets;
        this.showSuccess('AI 已填写，检查后点「保存配置」');
      } catch (e) {
        this.showError(`AI 分析失败: ${e}`);
      } finally {
        this.aiBusyScript = null;
      }
    },

    // ── Workbench: global env editor ──────────────────────────────────
    envLoad() {
      const dirVars = new Map(this.scriptDirs.filter(d => d.exportName).map(d => [d.exportName, d.path]));
      this.envRows = (this.cfg.exports || []).map((e, i) => {
        const inPath = e.kind ? e.kind === 'path' : /^(\/|~\/|\$)/.test((e.value || '').trim());
        return { _id: 'e' + i + Date.now(), _dir: dirVars.get(e.name) || '', name: e.name || '', value: e.value || '',
          kind: inPath ? 'path' : 'var', secret: !!e.secret, note: e.note || '' };
      });
      this.envDirty = false;
      this.envReveal = {};
    },

    envAdd(row = {}) {
      this.envRows = [...this.envRows, { _id: 'n' + Date.now() + Math.random(), _dir: '', name: '', value: '', kind: 'var', secret: false, note: '', ...row }];
      this.envDirty = true;
      if (!row.name) this.$nextTick(() => {
        const inputs = document.querySelectorAll('.wb-env-name input');
        inputs[inputs.length - 1]?.focus();
      });
    },

    async envRemove(i) {
      const e = this.envRows[i];
      if (e.name && !(await this.askConfirm(`删除变量 ${e.name}？保存后会从 ~/.zshrc 中移除。`, { title: '删除变量', confirmText: '删除' }))) return;
      this.envRows = this.envRows.filter((_, idx) => idx !== i);
      this.envDirty = true;
    },

    envMatches(e) {
      const q = this.envQuery.trim().toLowerCase();
      return !q || (e.name + ' ' + e.note + ' ' + (e.secret ? '' : e.value)).toLowerCase().includes(q);
    },

    envIssue(e) {
      if (!e.name.trim()) return e.value ? '缺少名称' : '';
      if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(e.name.trim())) return '只能用字母、数字、下划线，且不以数字开头';
      if (['PATH', 'HOME', 'SHELL', 'USER'].includes(e.name.trim())) return '系统变量，不建议在这里覆盖';
      if (this.envRows.filter(x => x.name.trim() === e.name.trim()).length > 1) return '名称重复';
      return '';
    },

    async envSave() {
      if (this.envHasIssues) return;
      this.savingExports = true;
      try {
        const exports = this.envRows.filter(e => e.name.trim()).map(e => ({
          name: e.name.trim(), value: e.value, kind: e.kind, secret: !!e.secret, note: e.note.trim(),
        }));
        await SaveExports(exports);
        await this.loadConfig();
        this.envLoad();
        this.showSuccess('已写入 ~/.zshrc，新开的终端生效');
      } catch (e) { this.showError(`保存失败: ${e}`); }
      finally { this.savingExports = false; }
    },

    // KEY=value lines; tolerates `export`, quotes and comments.
    parseDotenv(text) {
      const out = [];
      for (const raw of String(text || '').split('\n')) {
        const line = raw.trim();
        if (!line || line.startsWith('#')) continue;
        const m = line.match(/^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$/);
        if (!m) continue;
        let v = m[2].trim();
        if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) v = v.slice(1, -1);
        else v = v.replace(/\s+#.*$/, '');
        out.push({ name: m[1], value: v });
      }
      return out;
    },

    envImport() {
      const rows = this.envImportPreview;
      for (const r of rows) {
        const existing = this.envRows.find(e => e.name === r.name && !e._dir);
        const secret = /KEY|TOKEN|SECRET|PASSWORD|PASS/i.test(r.name);
        // Only executable dirs join $PATH; a FOO_DIR=/some/dir is an ordinary variable.
        const binDir = /^(\/|~\/|\$)/.test(r.value) && (/\/s?bin\/?$/.test(r.value) || /BIN$/i.test(r.name));
        if (existing) existing.value = r.value;
        else this.envAdd({ name: r.name, value: r.value, secret, kind: binDir ? 'path' : 'var' });
      }
      this.envDirty = true;
      this.envImportOpen = false;
      this.envImportText = '';
      this.showSuccess(`已导入 ${rows.length} 个变量，确认后保存`);
    },

    envPreview() {
      const rows = this.envRows.filter(e => e.name.trim()).sort((a, b) => a.name.localeCompare(b.name));
      const lines = ['# >>> Hatch managed block >>>'];
      const home = (this.zshrcInfo?.zshrcPath || '').replace(/\/\.zshrc$/, '');
      const expand = v => (home && /^~(\/|$)/.test(v) ? home + v.slice(1) : v);
      for (const e of rows) lines.push(`export ${e.name.trim()}=${this.shellValue(expand(e.value))}`);
      const paths = rows.filter(e => e.kind === 'path').map(e => '$' + e.name.trim());
      if (paths.length) lines.push(`export PATH=${paths.join(':')}:$PATH`);
      lines.push('# <<< Hatch managed block <<<');
      return lines.join('\n');
    },

    // ── Workbench: new script dialog ──────────────────────────────────
    openNewScriptDialog() {
      const dir = this.wbCurrent ? this.wbDirOf(this.wbCurrent.path) : this.scriptDirs[0]?.path || '';
      if (!dir) { this.wbGo('dirs'); this.showError('先添加一个脚本目录'); return; }
      this.newScript = { open: true, dir, filename: '', mode: 'template', template: '', instruction: '', content: '', rendered: '', busy: false, saving: false };
      this.applyScriptTemplate(this.scriptTemplates[0]);
    },

    scriptTemplateExt() {
      return this.scriptTemplates.find(t => t.id === this.newScript.template)?.ext || '.sh';
    },

    renderScriptTemplate(t, file) {
      return t.body.replaceAll('{{name}}', file.replace(/\.[^.]+$/, '')).replaceAll('{{file}}', file);
    },

    closeNewScriptDialog() {
      this.newScript = { ...this.newScript, open: false };
    },

    // Switching template swaps a known extension on the typed name and
    // re-renders the body (unless the user already edited it).
    applyScriptTemplate(t) {
      const ns = this.newScript;
      const name = ns.filename.trim();
      if (name && this.scriptTemplates.some(x => name.endsWith(x.ext))) ns.filename = name.replace(/\.[^.]+$/, '') + t.ext;
      ns.template = t.id;
      ns.content = ns.rendered = this.renderScriptTemplate(t, this.newScriptFileName());
    },

    // Final file name: typed name, plus the template extension when it has none.
    newScriptFileName() {
      const name = this.newScript.filename.trim() || 'my-script';
      return /\.[A-Za-z0-9]+$/.test(name) ? name : name + this.scriptTemplateExt();
    },

    async newScriptAIGenerateDraft() {
      const ns = this.newScript;
      if (!this.aiReady || !ns.instruction.trim()) return;
      ns.busy = true;
      try {
        ns.content = await SuggestScriptContent(this.newScriptFileName(), '', ns.instruction);
        ns.rendered = '';
        const py = /^#!.*python/.test(ns.content), js = /^#!.*node/.test(ns.content);
        ns.template = py ? 'python' : js ? 'node' : 'zsh';
      } catch (e) {
        this.showError(`AI 生成失败: ${e}`);
      } finally {
        ns.busy = false;
      }
    },

    async createNewScript() {
      const ns = this.newScript;
      if (!ns.filename.trim() || !ns.dir) return;
      const name = this.newScriptFileName();
      const t = this.scriptTemplates.find(x => x.id === ns.template);
      if (t && ns.mode === 'template' && ns.content === ns.rendered) ns.content = this.renderScriptTemplate(t, name);
      if (name.includes('/') || name.startsWith('.')) { this.showError('文件名不能包含 / 或以 . 开头'); return; }
      ns.saving = true;
      try {
        const path = await CreateScriptFile(ns.dir, name, ns.content);
        await this.scanAll();
        this.closeNewScriptDialog();
        this.wbGo('scripts');
        this.wbQuery = '';
        const s = this.allScripts.find(x => x.path === path);
        if (s) await this.wbSelect(s, 'run');
        this.showSuccess(`已创建 ${name}`);
      } catch (e) {
        this.showError(`创建失败: ${e}`);
      } finally {
        ns.saving = false;
      }
    },

    // ── Workbench: ⌘K command palette ─────────────────────────────────
    openPalette() {
      if (!this.moduleOn('scripts')) return;
      this.palette = { open: true, query: '', index: 0 };
      if (!this.wbRecentRuns.length) this.loadRecentRuns();
    },

    closePalette() {
      this.palette = { ...this.palette, open: false };
    },

    paletteMove(delta) {
      const n = this.paletteItems.length;
      if (!n) return;
      this.palette.index = (this.palette.index + delta + n) % n;
      this.$nextTick(() => document.querySelector('.wb-palette-item.active')?.scrollIntoView({ block: 'nearest' }));
    },

    // Enter runs (when no required input is missing); ⌘Enter / missing args opens the detail page.
    async paletteExec(it, openOnly = false) {
      if (!it) return;
      this.closePalette();
      if (this.tab !== 'scripts') this.openTab('scripts');
      if (it.kind === 'action') {
        if (it.action === 'new') this.openNewScriptDialog();
        else this.wbGo(it.action);
        return;
      }
      this.wbGo('scripts');
      this.wbQuery = '';
      await this.wbSelect(it.script, 'run');
      if (openOnly) return;
      if (it.kind === 'preset') { this.wbRunPreset(it.script, it.preset); return; }
      if (it.script.type === 'service') return;
      const { missing } = this.wbArgs(it.script);
      if (missing.length) {
        this.$nextTick(() => document.querySelector('.wb-arg input')?.focus());
        return;
      }
      this.wbRun(it.script);
    },

    setupWorkbenchKeys() {
      if (window.__hatchWbKeys) window.removeEventListener('keydown', window.__hatchWbKeys);
      window.__hatchWbKeys = (e) => {
        if (e.metaKey && e.key.toLowerCase() === 'k') {
          e.preventDefault();
          this.palette.open ? this.closePalette() : this.openPalette();
        } else if (e.metaKey && e.key === 'Enter' && this.tab === 'scripts' && this.wbSubTab === 'scripts' && this.wbPane === 'run' && this.wbCurrent && !this.palette.open) {
          e.preventDefault();
          this.wbRun(this.wbCurrent);
        }
      };
      window.addEventListener('keydown', window.__hatchWbKeys);
    },

    // ── Events ─────────────────────────────────────────────────────────
    setupEvents() {
      EventsOff('executor:log', 'executor:done', 'executor:status', 'ai:trace', 'memo:reminder');
      EventsOn('ai:trace', (trace) => {
        this.aiTraces = [trace, ...this.aiTraces].slice(0, 50);
      });

      EventsOn('memo:reminder', (item) => {
        this.showSuccess(`备忘提醒：${item.title || '待办事项'}`);
        this.loadMemoModule();
      });

      EventsOn('executor:log', (ev) => {
        const { id, line, stream } = ev;
        if (!this.scriptLogs[id]) this.scriptLogs[id] = [];
        this.scriptLogs[id].push({ line, stream });
        if (this.scriptLogs[id].length > 500) {
          this.scriptLogs[id] = this.scriptLogs[id].slice(-500);
        }
        this.scrollLog(id);
      });

      EventsOn('executor:done', (ev) => {
        delete this.runningSet[ev.id];
        this.runningSet = { ...this.runningSet };
        this.scriptExitCodes = { ...this.scriptExitCodes, [ev.id]: ev.exitCode };
        this.loadRecentRuns();
        if (this.historyTarget?.path === ev.id && this.wbPane === 'history') this.openHistory(this.historyTarget);
      });

      EventsOn('executor:status', (ev) => {
        this.statuses = { ...this.statuses, [ev.id]: ev.status };
        if (ev.status !== 'running') {
          delete this.runningSet[ev.id];
          this.runningSet = { ...this.runningSet };
        }
      });
    },

    async refreshStatuses() {
      try { this.statuses = await GetAllServiceStatuses() || {}; } catch (_) {}
    },

    scrollLog(id) {
      this.$nextTick(() => {
        const el = document.getElementById('log-' + this.safeId(id));
        if (el) el.scrollTop = el.scrollHeight;
      });
    },

    async serviceAction(script, action, args = []) {
      const id = script.path;
      if (!this.scriptLogs[id]) this.scriptLogs[id] = [];
      this.scriptLogs[id].push({ line: `[hatch] ${action}…`, stream: 'stdout' });
      try {
        await ServiceAction(id, action, args);
        this.scriptLogs[id].push({ line: `[hatch] ${action} OK`, stream: 'stdout' });
        await this.refreshStatuses();
      } catch (e) {
        this.scriptLogs[id].push({ line: `[hatch] ${action} error: ${e}`, stream: 'stderr' });
        this.showError(`${action} 失败: ${e}`);
      }
      this.scrollLog(id);
    },

    logLines(path) { return this.scriptLogs[path] || []; },

    presetAction(preset) {
      const action = this.splitArgs(preset?.args || '')[0] || '';
      return ['start', 'stop', 'restart'].includes(action) ? action : '';
    },

    presetButtonLabel(preset) {
      const action = this.presetAction(preset);
      const prefix = action === 'start' ? '▶ ' : action === 'stop' ? '■ ' : action === 'restart' ? '↺ ' : '🔖 ';
      return prefix + (preset?.name || preset?.args || '预设');
    },

    presetClass(preset) {
      const action = this.presetAction(preset);
      return {
        'preset-action-btn': !!action,
        'preset-start': action === 'start',
        'preset-stop': action === 'stop',
        'preset-restart': action === 'restart',
      };
    },

    isPresetDisabled(script, preset) {
      const action = this.presetAction(preset);
      if (action) return false;
      return !!this.runningSet[script.path];
    },

    clearLog(path) {
      this.scriptLogs[path] = [];
      delete this.scriptExitCodes[path];
      this.scriptExitCodes = { ...this.scriptExitCodes };
    },

    // ── Settings ───────────────────────────────────────────────────────
    async addDirectory() {
      if (!this.newDir.path) return;
      try {
        await AddDirectory(this.newDir.path, '', true);
        await this.loadConfig();
        await this.scanAll();
        this.newDir = { path: '' };
        this.showSuccess('目录已添加');
      } catch (e) { this.showError(`添加目录失败: ${e}`);   }
    },

    async removeDirectory(path) {
      if (!(await this.askConfirm(`移除目录「${path}」？\n脚本文件不会被删除，但该目录的导出与 Hatch 中的配置会被移除。`, { title: '确认移除', confirmText: '移除' }))) return;
      try {
        await RemoveDirectory(path);
        await this.loadConfig();
        await this.scanAll();
        this.showSuccess('目录已移除');
      } catch (e) { this.showError(`移除目录失败: ${e}`);   }
    },

    startInlineExport(dir) {
      const seg = dir.path.split('/').filter(Boolean).pop() || '';
      this.inlineExportName = dir.exportName || seg.replace(/[^a-zA-Z0-9_]/g, '_').toUpperCase();
      this.inlineExportDir = dir.path;
      this.$nextTick(() => {
        const el = document.querySelector('.inline-export-input');
        if (el) { el.focus(); el.select(); }
      });
    },

    cancelInlineExport() {
      this.inlineExportDir = null;
      this.inlineExportName = '';
    },

    async confirmInlineExport() {
      if (!this.inlineExportDir || !this.inlineExportName) return;
      try {
        await UpdateDirectoryExport(this.inlineExportDir, this.inlineExportName);
        await this.loadConfig();
        const name = this.inlineExportName;
        this.inlineExportDir = null;
        this.inlineExportName = '';
        this.showSuccess(`$${name} 已写入 ~/.zshrc`);
      } catch (e) { this.showError(`设置导出失败: ${e}`);   }
    },

    async clearExport(dir) {
      if (!(await this.askConfirm(`取消「${dir.path}」的路径导出？\n会从 ~/.zshrc 的 Hatch 区块中移除该导出。`, { title: '取消路径导出', confirmText: '取消导出' }))) return;
      try {
        await UpdateDirectoryExport(dir.path, '');
        await this.loadConfig();
        this.showSuccess('路径导出已取消');
      } catch (e) { this.showError(`取消导出失败: ${e}`);   }
    },

    shellValue(value) {
      const text = String(value || '');
      if (!text) return '""';
      return /^[A-Za-z0-9/._\-+:]+$/.test(text) ? text : this.shellSingleQuote(text);
    },

    shellSingleQuote(value) {
      return "'" + String(value).replaceAll("'", "'\\''") + "'";
    },

    async toggleAutoLaunch() {
      const next = !!this.autoLaunchEnabled;
      this.autoLaunchSaving = true;
      try {
        await SetAutoLaunchEnabled(next);
        this.autoLaunchEnabled = await GetAutoLaunchEnabled();
        this.cfg.autoLaunch = this.autoLaunchEnabled;
        this.showSuccess(this.autoLaunchEnabled ? '开机自启已开启' : '开机自启已关闭');
      } catch (e) {
        this.autoLaunchEnabled = !next;
        this.showError(`设置开机自启失败: ${e}`);
      } finally {
        this.autoLaunchSaving = false;
      }
    },

    async saveAIConfig() {
      this.aiSaving = true;
      this.aiTestResult = '';
      try {
        await SaveAIConfig({
          enabled: !!this.aiForm.enabled,
          baseUrl: (this.aiForm.baseUrl || '').trim(),
          apiKey: (this.aiForm.apiKey || '').trim(),
          model: (this.aiForm.model || '').trim(),
          traceRetainDays: this.aiForm.traceRetainDays || 3,
        });
        await this.loadConfig();
        this.showSuccess('AI 配置已保存');
      } catch (e) {
        this.showError(`保存 AI 配置失败: ${e}`);  
      } finally {
        this.aiSaving = false;
      }
    },

    async testAIConfig() {
      this.aiTesting = true;
      this.aiTestResult = '';
      try {
        const result = await TestAIConfig({
          enabled: true,
          baseUrl: (this.aiForm.baseUrl || '').trim(),
          apiKey: (this.aiForm.apiKey || '').trim(),
          model: (this.aiForm.model || '').trim(),
        });
        this.aiTestResult = `连接成功：${result}`;
        this.showSuccess(`✓ 连接成功：${result}`);
      } catch (e) {
        this.showError(`AI 连接失败: ${e}`);  
      } finally {
        this.aiTesting = false;
      }
    },

    // ── Script metadata editing ────────────────────────────────────────
    // v1.3+: data dir inventory + retention
    async loadDataInventory() {
      try {
        const getInv = window?.go?.app?.App?.GetAppDataInventory;
        const getRet = window?.go?.app?.App?.GetRetentionPolicy;
        if (typeof getInv === 'function') {
          this.dataInventory = (await getInv()) || [];
        }
        if (typeof getRet === 'function') {
          const r = await getRet();
          this.retention = {
            traceRetainDays: r.traceRetainDays || 3,
            runsRetainDays:  r.runsRetainDays || 14,
            runsKeepPerName: r.runsKeepPerName || 50,
          };
        }
      } catch (e) { this.showError(`加载数据目录失败: ${e}`); }
    },
    async saveRetention() {
      this.retentionSaving = true;
      try {
        const fn = window?.go?.app?.App?.SaveRetentionPolicy;
        if (typeof fn !== 'function') throw new Error('binding 未生成');
        await fn(this.retention.traceRetainDays|0, this.retention.runsRetainDays|0, this.retention.runsKeepPerName|0);
        this.showSuccess('保留策略已保存并立即应用。');
        await this.loadDataInventory();
      } catch (e) { this.showError(`保存失败: ${e}`); }
      finally { this.retentionSaving = false; }
    },
    async cleanupDataDir(d) {
      if (!d.cleanable) return;
      if (!(await this.askConfirm(`确认清空「${d.label}」(${d.fileCount} 个文件, ${this.formatBytes(d.totalBytes)})？此操作不可撤销。`, { title: '确认清空', confirmText: '清空' }))) return;
      try {
        const fn = window?.go?.app?.App?.CleanupDataDir;
        if (typeof fn !== 'function') throw new Error('binding 未生成');
        const msg = await fn(d.key);
        this.showSuccess(msg || '已清理');
        await this.loadDataInventory();
      } catch (e) { this.showError(`清理失败: ${e}`); }
    },
    async openDataDir(path) {
      try {
        const fn = window?.go?.app?.App?.OpenFile;
        if (typeof fn === 'function') await fn(path);
      } catch (_) {}
    },
    formatBytes(n) {
      if (!n) return '0 B';
      const u = ['B','KB','MB','GB'];
      let i = 0;
      while (n >= 1024 && i < u.length-1) { n /= 1024; i++; }
      return n.toFixed(i === 0 ? 0 : 1) + ' ' + u[i];
    },
    formatTime(unix) {
      if (!unix) return '';
      return new Date(unix * 1000).toLocaleString();
    },

    // v1.2: Prompt templates (uses dynamic bindings).
    async loadPromptTemplates() {
      try {
        const getOverrides = window?.go?.app?.App?.GetPromptTemplates;
        const getDefaults = window?.go?.app?.App?.GetDefaultPrompts;
        if (typeof getDefaults === 'function') {
          this.promptDefaults = (await getDefaults()) || {};
        }
        if (typeof getOverrides !== 'function') return;
        const current = (await getOverrides()) || {};
        const next = { ...this.promptTemplates };
        for (const op of this.promptOps) {
          next[op.key] = current[op.key] || '';
        }
        this.promptTemplates = next;
      } catch (e) {
        this.showError(`加载 Prompt 模板失败: ${e}`);
      }
    },
    togglePromptDefault(key) {
      this.promptDefaultsOpen = { ...this.promptDefaultsOpen, [key]: !this.promptDefaultsOpen[key] };
    },
    useDefaultAsTemplate(key) {
      this.promptTemplates = { ...this.promptTemplates, [key]: this.promptDefaults[key] || '' };
    },
    async savePromptTemplates() {
      this.promptSaving = true;
      try {
        const fn = window?.go?.app?.App?.SavePromptTemplates;
        if (typeof fn !== 'function') throw new Error('binding 未生成，请运行 wails dev 或 wails build 重新生成');
        const payload = {};
        for (const op of this.promptOps) {
          const v = (this.promptTemplates[op.key] || '').trim();
          if (v) payload[op.key] = v;
        }
        await fn(payload);
        this.showSuccess('Prompt 模板已保存（立即对后续 AI 调用生效）');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.promptSaving = false;
      }
    },
    async resetPromptTemplate() {
      if (!(await this.askConfirm('确认清空所有自定义 Prompt 模板，恢复内置默认？', { title: '恢复默认 Prompt', confirmText: '清空并恢复' }))) return;
      const blank = {};
      for (const op of this.promptOps) blank[op.key] = '';
      this.promptTemplates = blank;
      await this.savePromptTemplates();
    },

    // v1.1: Run history (uses dynamic bindings — not in wailsjs until rebuild).
    async openHistory(script) {
      this.historyTarget = script;
      this.historyRecords = [];
      this.historyOpenId = null;
      this.historyLogText = '';
      this.historyLoading = true;
      try {
        const fn = window?.go?.app?.App?.ListRuns;
        if (typeof fn === 'function') {
          const runs = await fn(script.path);
          this.historyRecords = runs || [];
        }
      } catch (e) {
        this.showError(`加载历史失败: ${e}`);
      } finally {
        this.historyLoading = false;
      }
    },
    closeHistory() {
      this.historyTarget = null;
      this.historyRecords = [];
      this.historyOpenId = null;
      this.historyLogText = '';
    },
    async viewHistoryLog(r) {
      if (this.historyOpenId === r.id) {
        this.historyOpenId = null;
        return;
      }
      this.historyOpenId = r.id;
      this.historyLogText = '加载日志中…';
      try {
        const fn = window?.go?.app?.App?.GetRunLog;
        if (typeof fn === 'function') {
          this.historyLogText = await fn(this.historyTarget.path, r.id);
        }
      } catch (e) {
        this.historyLogText = `读取失败: ${e}`;
      }
    },
    async deleteHistory(r) {
      if (!(await this.askConfirm('删除这条运行记录及其日志？此操作不可撤销。'))) return;
      try {
        const fn = window?.go?.app?.App?.DeleteRun;
        if (typeof fn === 'function') {
          await fn(this.historyTarget.path, r.id);
        }
        this.historyRecords = this.historyRecords.filter(x => x.id !== r.id);
      } catch (e) {
        this.showError(`删除失败: ${e}`);
      }
    },
    rerunFromHistory(r) {
      const script = this.allScripts.find(s => s.path === this.historyTarget?.path);
      if (!script) return;
      this.wbPane = 'run';
      this.wbRun(script, r.args || []);
    },

    addEnvPair() {
      if (!this.editForm.envPairs) this.editForm.envPairs = [];
      this.editForm.envPairs = [...this.editForm.envPairs, { key: '', value: '' }];
    },
    removeEnvPair(i) {
      this.editForm.envPairs = this.editForm.envPairs.filter((_, idx) => idx !== i);
    },

    removePreset(i) {
      this.editForm.presets = this.editForm.presets.filter((_, idx) => idx !== i);
    },

    // ── Skills management ─────────────────────────────────────────────
    async loadSkills() {
      this.skillsLoading = true;
      try {
        [this.skills, this.skillsStats] = await Promise.all([GetSkills(), GetSkillsStats()]);
        this.skills = this.skills || [];
      } catch (e) { this.showError(`加载 Skills 失败: ${e}`); }
      finally { this.skillsLoading = false; }
    },

    async openSkillEditor(name) {
      try {
        this.skillEditContent = await GetSkillContent(name);
        this.skillEditTarget = name;
        this.skillEditIsNew = false;
      } catch (e) { this.showError(`读取 Skill 失败: ${e}`); }
    },

    async openNewSkillEditor() {
      const name = (this.skillNewName || '').trim();
      if (!name) { this.showError('请先输入 Skill 名称'); return; }
      if (!/^[\w-]+$/.test(name)) { this.showError('名称只能含字母、数字、连字符'); return; }
      this.skillEditContent = await CreateSkillTemplate(name);
      this.skillEditTarget = name;
      this.skillEditIsNew = true;
      this.skillNewName = '';
    },

    closeSkillEditor() {
      this.skillEditTarget = null;
      this.skillEditContent = '';
      this.skillEditIsNew = false;
    },

    // v1.3+: Skill AI optimization
    skillAIInstruction: '',
    skillAIBusy: false,
    async optimizeSkillWithAI() {
      if (!this.skillEditTarget) return;
      const instruction = (this.skillAIInstruction || '').trim();
      if (!instruction) {
        this.showError('请先填写优化指令，例如"补全 frontmatter / 加触发词 / 重写工作流程"');
        return;
      }
      this.skillAIBusy = true;
      try {
        const fn = window?.go?.app?.App?.SuggestSkillContent;
        if (typeof fn !== 'function') throw new Error('binding 未生成，请重新 wails build');
        const out = await fn(this.skillEditTarget, this.skillEditContent || '', instruction);
        if (out) {
          this.skillEditContent = out;
          this.showSuccess('已用 AI 生成新内容，确认后点保存。');
        }
      } catch (e) {
        this.showError(`AI 优化失败: ${e}`);
      } finally {
        this.skillAIBusy = false;
      }
    },

    async saveSkillContent() {
      if (!this.skillEditTarget) return;
      this.skillEditSaving = true;
      try {
        const name = this.skillEditTarget;
        // capture current sync state before saving
        const skill = this.skills.find(s => s.name === name);
        await SaveSkillContent(name, this.skillEditContent);
        // re-sync to any targets that were already enabled
        const resynced = [];
        if (skill?.inClaude) {
          await SetSkillSync(name, 'claude', true);
          resynced.push('Claude');
        }
        if (skill?.inCodex) {
          await SetSkillSync(name, 'codex', true);
          resynced.push('Codex');
        }
        await this.loadSkills();
        this.closeSkillEditor();
        const syncNote = resynced.length ? `，已同步到 ${resynced.join(' / ')}` : '';
        this.showSuccess(`Skill「${name}」已保存${syncNote}`);
      } catch (e) { this.showError(`保存 Skill 失败: ${e}`); }
      finally { this.skillEditSaving = false; }
    },

    async deleteSkill(name) {
      if (!(await this.askConfirm(`仅从 Hatch 删除 Skill「${name}」？\nClaude / Codex 中已同步的副本会保留。`))) return;
      try {
        const skill = this.skills.find(s => s.name === name);
        await DeleteSkill(name);
        if (this.skillEditTarget === name) this.closeSkillEditor();
        await this.loadSkills();
        const kept = [];
        if (skill?.inClaude) kept.push('Claude');
        if (skill?.inCodex) kept.push('Codex');
        const keepNote = kept.length ? `，${kept.join(' / ')} 保留不变` : '';
        this.showSuccess(`Skill「${name}」已从 Hatch 删除${keepNote}`);
      } catch (e) { this.showError(`删除 Skill 失败: ${e}`); }
    },

    async deleteSkillEverywhere(name) {
      if (!(await this.askConfirm(`从 Hatch、Claude、Codex 全部删除 Skill「${name}」？此操作不可撤销。`))) return;
      try {
        await DeleteSkillEverywhere(name);
        if (this.skillEditTarget === name) this.closeSkillEditor();
        await this.loadSkills();
        this.showSuccess(`Skill「${name}」已从 Hatch / Claude / Codex 删除`);
      } catch (e) { this.showError(`删除 Skill 失败: ${e}`); }
    },

    async toggleSkillSync(name, target, enabled) {
      const key = name + ':' + target;
      this.skillSyncBusy = { ...this.skillSyncBusy, [key]: true };
      try {
        await SetSkillSync(name, target, enabled);
        // Update local state immediately without full reload
        const idx = this.skills.findIndex(s => s.name === name);
        if (idx >= 0) {
          const s = { ...this.skills[idx] };
          if (target === 'claude') s.inClaude = enabled;
          else s.inCodex = enabled;
          this.skills = [...this.skills.slice(0, idx), s, ...this.skills.slice(idx + 1)];
        }
        await GetSkillsStats().then(st => this.skillsStats = st).catch(() => {});
        const label = target === 'claude' ? 'Claude' : 'Codex';
        this.showSuccess(enabled ? `「${name}」已同步到 ${label}` : `「${name}」已从 ${label} 移除`);
      } catch (e) { this.showError(`同步失败: ${e}`); }
      finally {
        const b = { ...this.skillSyncBusy };
        delete b[key];
        this.skillSyncBusy = b;
      }
    },

    async importSkills(source) {
      this.skillImporting = source;
      this.skillImportResult = null;
      try {
        const r = await ImportSkillsFrom(source);
        this.skillImportResult = r;
        await this.loadSkills();
        const n = (r?.imported || []).length;
        const c = (r?.conflicts || []).length;
        const label = source === 'claude' ? 'Claude' : 'Codex';
        if (n > 0 || c > 0) {
          const parts = [];
          if (n > 0) parts.push(`导入 ${n} 个`);
          if (c > 0) parts.push(`发现冲突 ${c} 个`);
          this.showSuccess(`已从 ${label} ${parts.join('，')}`);
        } else {
          this.showSuccess('无变化');
        }
      } catch (e) { this.showError(`导入失败: ${e}`); }
      finally { this.skillImporting = null; }
    },

    skillSourceLabel(source) {
      return source === 'claude' ? 'Claude' : 'Codex';
    },

    async resolveSkillConflict(conflict, useSource) {
      if (!conflict?.name || !conflict?.source) return;
      const key = conflict.source + ':' + conflict.name;
      this.skillConflictBusy = { ...this.skillConflictBusy, [key]: true };
      try {
        await ResolveSkillImportConflict(conflict.name, conflict.source, useSource);
        const rest = (this.skillImportResult?.conflicts || []).filter(c => !(c.name === conflict.name && c.source === conflict.source));
        this.skillImportResult = { ...(this.skillImportResult || {}), conflicts: rest };
        await this.loadSkills();
        const label = this.skillSourceLabel(conflict.source);
        this.showSuccess(useSource ? `Skill「${conflict.name}」已使用 ${label} 版本` : `Skill「${conflict.name}」已保留 Hatch 版本`);
      } catch (e) {
        this.showError(`处理冲突失败: ${e}`);
      } finally {
        const b = { ...this.skillConflictBusy };
        delete b[key];
        this.skillConflictBusy = b;
      }
    },

    // Wails' macOS WebView does not implement the JS confirm panel, so
    // window.confirm() silently returns false. Every destructive action awaits
    // this in-app dialog instead.
    askConfirm(message, { title = '确认删除', confirmText = '删除', danger = true } = {}) {
      if (this.confirmResolve) this.confirmResolve(false);
      this.confirmDialog = { open: true, title, message, confirmText, danger };
      return new Promise(resolve => { this.confirmResolve = resolve; });
    },

    settleConfirm(ok) {
      const resolve = this.confirmResolve;
      this.confirmResolve = null;
      this.confirmDialog = { ...this.confirmDialog, open: false };
      if (resolve) resolve(ok);
    },

    // ── MCP ───────────────────────────────────────────────────────────
    // Claude Code / Codex config files are the source of truth; the page shows
    // GetMCPInventory rows (one per server name, one entry per location) and
    // every write goes through preview → apply.
    async loadMCP(includeCLI = false) {
      if (!includeCLI && this.mcpLoading) return;   // hero stats + page loader may race
      if (includeCLI) this.mcpCLILoading = true; else this.mcpLoading = true;
      try {
        const [inv, catalog] = await Promise.all([
          GetMCPInventory(includeCLI),
          this.mcpCatalog.length ? Promise.resolve(this.mcpCatalog) : GetMCPCatalog(),
        ]);
        this.mcpInventory = inv || { rows: [], projects: [], sources: [] };
        this.mcpCatalog = catalog || [];
      } catch (e) {
        this.showError(`读取 MCP 配置失败: ${e}`);
      } finally {
        this.mcpLoading = false;
        this.mcpCLILoading = false;
      }
    },

    mcpBusyKey(name, key, op) { return `${name}|${key}|${op}`; },
    mcpIsBusy(name, key, op) { return !!this.mcpBusy[this.mcpBusyKey(name, key, op)]; },
    async mcpRun(name, key, op, fn) {
      const k = this.mcpBusyKey(name, key, op);
      if (this.mcpBusy[k]) return;
      this.mcpBusy = { ...this.mcpBusy, [k]: true };
      try { return await fn(); }
      finally { const b = { ...this.mcpBusy }; delete b[k]; this.mcpBusy = b; }
    },

    toggleMCPRow(name) {
      this.mcpExpanded = { ...this.mcpExpanded, [name]: !this.mcpExpanded[name] };
    },

    mcpSummary(spec) {
      if (!spec) return '';
      if (spec.transport === 'stdio') return [spec.command, ...(spec.args || [])].join(' ');
      return spec.url || '';
    },

    mcpEntryState(e) {
      if (e.stale) return { cls: 'mcp-chip-stale', text: '目录已不存在' };
      if (e.readOnly) return { cls: 'mcp-chip-ro', text: '只读' };
      if (e.parked) return { cls: 'mcp-chip-off', text: '已停用' };
      return e.enabled ? { cls: 'mcp-chip-on', text: '启用' } : { cls: 'mcp-chip-off', text: '已停用' };
    },

    mcpProbeClass(p) {
      if (!p) return 'mcp-probe-none';
      return { ok: 'mcp-probe-ok', failed: 'mcp-probe-fail', timeout: 'mcp-probe-fail', auth_required: 'mcp-probe-auth' }[p.status] || 'mcp-probe-none';
    },

    mcpProbeLabel(p) {
      if (!p) return '未测试';
      return { ok: `✓ ${(p.tools || []).length} 个工具`, failed: '✕ 失败', timeout: '✕ 超时', auth_required: '需授权', unsupported: '不支持探测' }[p.status] || p.status;
    },

    async setMCPEntryEnabled(row, e, enabled) {
      await this.mcpRun(row.name, e.locationKey, 'toggle', async () => {
        try {
          await SetMCPEnabled(row.name, e.locationKey, enabled);
          this.showSuccess(`「${row.name}」已在 ${e.label} ${enabled ? '启用' : '停用（配置已保留，可随时恢复）'}`);
        } catch (err) {
          this.showError(`切换失败: ${err}`);
        }
        await this.loadMCP();
      });
    },

    async probeMCPEntry(row, e, quiet = false) {
      return this.mcpRun(row.name, e.locationKey, 'probe', async () => {
        try {
          const res = await ProbeMCP(row.name, e.locationKey);
          if (!quiet) {
            if (res.status === 'ok') this.showSuccess(`「${row.name}」${res.summary}`);
            else this.showError(`「${row.name}」@${e.label}：${res.summary}`);
          }
          this.patchMCPProbe(row.name, e.locationKey, res);
          return res;
        } catch (err) {
          if (!quiet) this.showError(`测试失败: ${err}`);
        }
      });
    },

    patchMCPProbe(name, key, res) {
      const inv = this.mcpInventory;
      if (!inv) return;
      const rows = inv.rows.map(r => r.name !== name ? r : {
        ...r, entries: r.entries.map(x => x.locationKey === key ? { ...x, probe: res } : x),
      });
      this.mcpInventory = { ...inv, rows };
    },

    async probeAllMCP() {
      const targets = [];
      for (const r of this.filteredMCPRows) {
        for (const e of r.entries) if (!e.readOnly && !e.stale && e.enabled) targets.push([r, e]);
      }
      if (!targets.length) { this.showError('没有可测试的已启用 MCP'); return; }
      this.mcpProbingAll = true;
      let ok = 0;
      try {
        for (const [r, e] of targets) {
          const res = await this.probeMCPEntry(r, e, true);
          if (res && res.status === 'ok') ok++;
        }
      } finally { this.mcpProbingAll = false; }
      const bad = targets.length - ok;
      bad ? this.showError(`已测试 ${targets.length} 项，${bad} 项未通过`) : this.showSuccess(`已测试 ${targets.length} 项，全部连接成功`);
    },

    async removeMCPEntry(row, e) {
      const extra = e.location.scope === 'project' ? '\n这会修改项目内共享的 .mcp.json。' : '';
      if (!(await this.askConfirm(`从 ${e.label} 删除 MCP「${row.name}」？\n文件：${e.file || ''}${extra}`))) return;
      await this.mcpRun(row.name, e.locationKey, 'remove', async () => {
        try {
          await RemoveMCPEntry(row.name, e.locationKey);
          this.showSuccess(`已从 ${e.label} 删除「${row.name}」`);
        } catch (err) {
          this.showError(`删除失败: ${err}`);
        }
        await this.loadMCP();
      });
    },

    // Drift fix: copy one location's version over every other writable one.
    async syncMCPFrom(row, e) {
      const targets = row.entries.filter(x => !x.readOnly && !x.stale && x.locationKey !== e.locationKey);
      if (!targets.length) return;
      const labels = targets.map(x => x.label).join('、');
      if (!(await this.askConfirm(`以 ${e.label} 的配置为准，覆盖 ${labels} 中的「${row.name}」？\n各位置独有的扩展字段（如 Codex 工具审批）会保留。`, { title: '同步配置', confirmText: '同步', danger: false }))) return;
      await this.mcpRun(row.name, e.locationKey, 'sync', async () => {
        try {
          await CopyMCPEntry(row.name, e.locationKey, targets.map(x => x.locationKey));
          this.showSuccess(`已用 ${e.label} 的配置同步「${row.name}」`);
        } catch (err) {
          this.showError(`同步失败: ${err}`);
        }
        await this.loadMCP();
      });
    },

    mcpDisabledTools(e) {
      return (e?.spec?.extra?.disabled_tools || []).map(String);
    },

    async toggleMCPTool(row, e, tool, on) {
      const set = new Set(this.mcpDisabledTools(e));
      on ? set.delete(tool) : set.add(tool);
      await this.mcpRun(row.name, e.locationKey, 'tools', async () => {
        try {
          await SetMCPDisabledTools(row.name, e.locationKey, [...set]);
          this.showSuccess(`Codex 已${on ? '启用' : '禁用'}工具 ${tool}`);
        } catch (err) {
          this.showError(`保存失败: ${err}`);
        }
        await this.loadMCP();
      });
    },

    async discardLegacyMCP(d) {
      if (!(await this.askConfirm(`丢弃旧版 Hatch 注册表中的「${d.name}」？它从未写入 Claude / Codex，丢弃后不可恢复。`, { title: '丢弃旧条目', confirmText: '丢弃' }))) return;
      try {
        await DiscardLegacyMCP(d.name);
        await this.loadMCP();
      } catch (err) { this.showError(`丢弃失败: ${err}`); }
    },

    // ── MCP add / edit drawer ─────────────────────────────────────────
    blankMCPForm() {
      return { name: '', title: '', description: '', transport: 'stdio', command: '', argsRaw: '', cwd: '', url: '', envRows: [], headerRows: [], bearerTokenEnvVar: '', startupTimeoutSec: 0, toolTimeoutSec: 0, extra: null, sourceClient: '', origin: 'manual', notes: [], homepage: '', variant: '' };
    },

    openMCPAdd(tab = 'paste') {
      if (!this.mcpInventory) this.loadMCP();
      this.mcpAdd = {
        open: true, mode: 'add', tab, editKey: '',
        pasteText: '', parsing: false, parseNotes: [],
        registryQuery: '', registryHits: [], registryBusy: false, registryLooked: false,
        drafts: [], draftIdx: 0,
        form: this.blankMCPForm(),
        targets: { 'claude:user': true, 'codex:user': false },
        project: '', projectLocal: false, projectShared: false, projectCodex: false,
        plan: null, previewing: false, applying: false, reveal: {},
      };
    },

    openMCPEdit(row, e) {
      this.openMCPAdd('form');
      this.mcpAdd.mode = 'edit';
      this.mcpAdd.editKey = e.locationKey;
      this.mcpAdd.editLabel = e.label;
      this.mcpAdd.form = this.mcpFormFromDraft({ name: row.name, spec: e.spec, sourceClient: e.location.client === 'desktop' ? 'claude' : e.location.client, origin: 'edit' });
    },

    openMCPCopy(row, e) {
      this.openMCPAdd('form');
      this.mcpAdd.form = this.mcpFormFromDraft({ name: row.name, spec: e.spec, sourceClient: e.location.client === 'desktop' ? 'claude' : e.location.client, origin: 'copy:' + e.locationKey });
      this.mcpAdd.targets = { 'claude:user': false, 'codex:user': false };
      this.mcpAdd.copyFrom = e.label;
    },

    openMCPLegacy(d) {
      this.openMCPAdd('form');
      this.loadMCPDrafts([d]);
      this.mcpAdd.form.origin = 'legacy';
    },

    closeMCPAdd() { this.mcpAdd = null; },

    mcpFormFromDraft(d) {
      const spec = d.spec || {};
      const inputs = {};
      for (const i of d.inputs || []) inputs[i.kind + ':' + i.key] = i;
      const rows = (map, kind) => Object.keys(map || {}).sort().map(k => {
        const i = inputs[kind + ':' + k] || {};
        let value = map[k] ?? '';
        // README placeholders ("YOUR_API_KEY", "<token>") must not count as filled.
        const bare = String(value).replace(/^Bearer\s*/, '');
        if (i.key && /^(<.*>|your[-_ ].*|xxx+|\.\.\.|changeme|todo|replace[-_ ]?me)$/i.test(bare)) value = String(value).startsWith('Bearer ') ? 'Bearer ' : '';
        return { key: k, value, secret: !!i.secret, required: !!i.required, desc: i.description || (bare && value !== map[k] ? bare : '') };
      });
      return {
        ...this.blankMCPForm(),
        name: d.name || '', title: d.title || '', description: d.description || '', variant: d.variant || '',
        transport: spec.transport || 'stdio', command: spec.command || '', argsRaw: (spec.args || []).join('\n'),
        cwd: spec.cwd || '', url: spec.url || '', envRows: rows(spec.env, 'env'), headerRows: rows(spec.headers, 'header'),
        bearerTokenEnvVar: spec.bearerTokenEnvVar || '', startupTimeoutSec: spec.startupTimeoutSec || 0, toolTimeoutSec: spec.toolTimeoutSec || 0,
        extra: spec.extra || null, sourceClient: d.sourceClient || '', origin: d.origin || 'manual', notes: d.notes || [], homepage: d.homepage || '',
      };
    },

    loadMCPDrafts(drafts, scope = '') {
      if (!this.mcpAdd || !drafts?.length) return;
      this.mcpAdd.drafts = drafts;
      this.selectMCPDraft(0);
      const hint = scope || drafts[0].scope;
      if (hint === 'user') this.mcpAdd.targets = { 'claude:user': true, 'codex:user': false };
    },

    selectMCPDraft(i) {
      const d = this.mcpAdd.drafts[i];
      if (!d) return;
      this.mcpAdd.draftIdx = i;
      this.mcpAdd.form = this.mcpFormFromDraft(d);
      this.mcpAdd.plan = null;
      if (d.spec?.transport === 'sse' || d.spec?.transport === 'ws') this.mcpAdd.targets['codex:user'] = false;
    },

    async parseMCPPaste() {
      const a = this.mcpAdd;
      if (!a || !a.pasteText.trim()) { this.showError('请先粘贴内容'); return; }
      a.parsing = true;
      try {
        const res = await ParseMCPInput(a.pasteText);
        a.parseNotes = res.notes || [];
        this.loadMCPDrafts(res.drafts || []);
        a.tab = 'form';
      } catch (err) {
        this.showError(`${err}`);
      } finally { a.parsing = false; }
    },

    pickMCPCatalog(item) {
      const d = { name: item.id, title: item.title, description: item.description, spec: item.spec, inputs: item.inputs, origin: 'catalog:' + item.id, notes: item.notes, homepage: item.homepage };
      this.loadMCPDrafts([d]);
      this.mcpAdd.tab = 'form';
      if (item.registryName) this.mcpAdd.form.notes = [...(this.mcpAdd.form.notes || []), `Registry 名称：${item.registryName}（可在 Registry 页签获取最新版本信息）`];
    },

    async searchMCPRegistry() {
      const a = this.mcpAdd;
      const q = (a?.registryQuery || '').trim();
      if (!q) return;
      a.registryBusy = true;
      try {
        if (/^[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+\/[a-zA-Z0-9._-]+$/.test(q)) {
          await this.lookupMCPRegistry(q);
          return;
        }
        a.registryHits = (await SearchMCPRegistry(q)) || [];
        a.registryLooked = true;
      } catch (err) {
        this.showError(`${err}`);
      } finally { a.registryBusy = false; }
    },

    async lookupMCPRegistry(name) {
      const a = this.mcpAdd;
      a.registryBusy = true;
      try {
        const drafts = await LookupMCPRegistry(name);
        this.loadMCPDrafts(drafts || []);
        a.tab = 'form';
      } catch (err) {
        this.showError(`${err}`);
      } finally { a.registryBusy = false; }
    },

    addMCPFormRow(kind) {
      const f = this.mcpAdd?.form;
      if (!f) return;
      f[kind] = [...(f[kind] || []), { key: '', value: '', secret: false, required: false, desc: '' }];
    },

    removeMCPFormRow(kind, idx) {
      const f = this.mcpAdd?.form;
      if (!f) return;
      f[kind] = f[kind].filter((_, i) => i !== idx);
    },

    mcpSpecFromForm() {
      const f = this.mcpAdd.form;
      const toMap = rows => {
        const out = {};
        for (const r of rows || []) { const k = (r.key || '').trim(); if (k) out[k] = r.value ?? ''; }
        return out;
      };
      const spec = { transport: f.transport };
      if (f.transport === 'stdio') {
        spec.command = (f.command || '').trim();
        spec.args = (f.argsRaw || '').split('\n').map(s => s.trim()).filter(Boolean);
        spec.cwd = (f.cwd || '').trim();
        spec.env = toMap(f.envRows);
      } else {
        spec.url = (f.url || '').trim();
        spec.headers = toMap(f.headerRows);
        spec.bearerTokenEnvVar = (f.bearerTokenEnvVar || '').trim();
      }
      spec.startupTimeoutSec = Number(f.startupTimeoutSec || 0);
      spec.toolTimeoutSec = Number(f.toolTimeoutSec || 0);
      if (f.extra && Object.keys(f.extra).length) spec.extra = f.extra;
      return spec;
    },

    mcpTargetsFromForm() {
      const a = this.mcpAdd;
      if (a.mode === 'edit') return [this.mcpLocFromKey(a.editKey)];
      const out = [];
      if (a.targets['claude:user']) out.push({ client: 'claude', scope: 'user' });
      if (a.targets['codex:user']) out.push({ client: 'codex', scope: 'user' });
      if (a.project) {
        if (a.projectLocal) out.push({ client: 'claude', scope: 'local', project: a.project });
        if (a.projectShared) out.push({ client: 'claude', scope: 'project', project: a.project });
        if (a.projectCodex) out.push({ client: 'codex', scope: 'project', project: a.project });
      }
      return out;
    },

    mcpLocFromKey(key) {
      const [client, scope, ...rest] = key.split(':');
      const project = rest.join(':');
      return project ? { client, scope, project } : { client, scope };
    },

    mcpApplyRequest() {
      const f = this.mcpAdd.form;
      return { name: (f.name || '').trim(), spec: this.mcpSpecFromForm(), targets: this.mcpTargetsFromForm(), sourceClient: f.sourceClient || '', origin: f.origin || 'manual' };
    },

    get mcpMissingInputs() {
      const f = this.mcpAdd?.form;
      if (!f) return [];
      const rows = f.transport === 'stdio' ? f.envRows : f.headerRows;
      return (rows || []).filter(r => r.required && !String(r.value || '').replace(/^Bearer\s*/, '').trim()).map(r => r.key);
    },

    get mcpSelectedProject() {
      const p = this.mcpAdd?.project;
      return (this.mcpInventory?.projects || []).find(x => x.path === p) || null;
    },

    async previewMCPApply() {
      const a = this.mcpAdd;
      const req = this.mcpApplyRequest();
      if (!req.targets.length) { this.showError('请至少选择一个写入位置'); return; }
      a.previewing = true;
      try {
        a.plan = await PreviewMCPApply(req);
      } catch (err) {
        this.showError(`预览失败: ${err}`);
      } finally { a.previewing = false; }
    },

    async applyMCPForm() {
      const a = this.mcpAdd;
      if (!a?.plan?.canApply) return;
      const req = this.mcpApplyRequest();
      a.applying = true;
      try {
        await ApplyMCP(req);
        this.showSuccess(`「${req.name}」已写入 ${a.plan.items.map(i => i.label).join('、')}`);
        this.closeMCPAdd();
        await this.loadMCP();
        // Probe right away so the user sees whether it actually starts.
        const row = (this.mcpInventory?.rows || []).find(r => r.name === req.name);
        if (row) {
          this.mcpExpanded = { ...this.mcpExpanded, [row.name]: true };
          const keys = new Set(req.targets.map(t => this.mcpLocKey(t)));
          for (const e of row.entries) if (keys.has(e.locationKey) && e.enabled) this.probeMCPEntry(row, e, true);
        }
      } catch (err) {
        this.showError(`写入失败: ${err}`);
      } finally { if (this.mcpAdd) this.mcpAdd.applying = false; }
    },

    mcpLocKey(loc) {
      return loc.client + ':' + loc.scope + (loc.project ? ':' + loc.project : '');
    },

    // ── Hooks management ──────────────────────────────────────────────
    async loadHooks() {
      this.hooksLoading = true;
      try {
        const [hooks, stats, support, templates] = await Promise.all([
          ListHooks(),
          GetHooksStats(),
          GetHookEventSupport(),
          GetHookTemplates(),
        ]);
        this.hooks = hooks || [];
        this.hooksStats = stats || { total: 0, claude: 0, codex: 0 };
        this.hookEventSupport = support || [];
        this.hookTemplates = templates || [];
      } catch (e) {
        this.showError(`加载 Hooks 失败: ${e}`);
      } finally {
        this.hooksLoading = false;
      }
    },

    hookEventSupports(event, target) {
      const info = (this.hookEventSupport || []).find(e => e.event === event);
      if (!info) return false;
      return target === 'claude' ? !!info.claude : !!info.codex;
    },

    hookEventLabel(event) {
      const info = (this.hookEventSupport || []).find(e => e.event === event);
      return info ? info.label : event;
    },

    hookMatcherHint() {
      const info = (this.hookEventSupport || []).find(e => e.event === (this.hookForm && this.hookForm.event));
      return info ? info.matcherHint : '';
    },

    hookTargetLabel(target) {
      return target === 'claude' ? 'Claude Code' : 'Codex';
    },

    blankHookForm() {
      return { id: '', name: '', description: '', event: 'PostToolUse', matcher: '', command: '', argsRaw: '', timeout: 0, statusMessage: '' };
    },

    hookFormFrom(hook) {
      return {
        id: hook.id || '',
        name: hook.name || '',
        description: hook.description || '',
        event: hook.event || 'PostToolUse',
        matcher: hook.matcher || '',
        command: hook.command || '',
        argsRaw: (hook.args || []).join('\n'),
        timeout: hook.timeout || 0,
        statusMessage: hook.statusMessage || '',
      };
    },

    hookFromForm() {
      const f = this.hookForm || {};
      return {
        id: (f.id || '').trim(),
        name: (f.name || '').trim(),
        description: (f.description || '').trim(),
        event: f.event || 'PostToolUse',
        matcher: (f.matcher || '').trim(),
        command: (f.command || '').trim(),
        args: (f.argsRaw || '').split('\n').map(s => s.trim()).filter(Boolean),
        timeout: Number(f.timeout || 0),
        statusMessage: (f.statusMessage || '').trim(),
      };
    },

    openNewHook() {
      this.hookForm = this.blankHookForm();
      this.hookEditIsNew = true;
      this.hookEditOpen = true;
    },

    openHookEditor(hook) {
      this.hookForm = this.hookFormFrom(hook);
      this.hookEditIsNew = false;
      this.hookEditOpen = true;
    },

    openHookTemplate(template) {
      const copy = JSON.parse(JSON.stringify(template || {}));
      this.hookForm = this.hookFormFrom(copy);
      this.hookEditIsNew = true;
      this.hookEditOpen = true;
      this.hookTemplateGalleryOpen = false;
    },

    closeHookEditor() {
      this.hookEditOpen = false;
      this.hookForm = null;
      this.hookEditIsNew = false;
      this.hookAIInstruction = '';
    },

    openHookTemplateGallery() { this.hookTemplateGalleryOpen = true; },
    closeHookTemplateGallery() { this.hookTemplateGalleryOpen = false; },

    async saveHook() {
      if (!this.hookForm) return;
      this.hookSaving = true;
      try {
        const hook = this.hookFromForm();
        const existing = this.hooks.find(h => h.id === hook.id);
        await SaveHook(hook);
        // Re-read the saved id (backend may generate one for new hooks).
        const saved = (await ListHooks()).find(h =>
          h.id === hook.id || (hook.id === '' && h.name === hook.name));
        const id = saved ? saved.id : hook.id;
        const resynced = [];
        if (existing?.enabledInClaude && this.hookEventSupports(hook.event, 'claude')) {
          await SetHookEnabled(id, 'claude', true);
          resynced.push('Claude Code');
        }
        if (existing?.enabledInCodex && this.hookEventSupports(hook.event, 'codex')) {
          await SetHookEnabled(id, 'codex', true);
          resynced.push('Codex');
        }
        await this.loadHooks();
        this.closeHookEditor();
        const note = resynced.length ? `，已同步到 ${resynced.join(' / ')}` : '';
        this.showSuccess(`Hook「${hook.name || id}」已保存${note}`);
      } catch (e) {
        this.showError(`保存 Hook 失败: ${e}`);
      } finally {
        this.hookSaving = false;
      }
    },

    async deleteHook(id) {
      if (!(await this.askConfirm(`删除 Hook「${id}」？已同步到 Claude / Codex 的条目也会一并移除。`))) return;
      try {
        await DeleteHook(id);
        await this.loadHooks();
        this.showSuccess(`Hook「${id}」已删除`);
      } catch (e) {
        this.showError(`删除 Hook 失败: ${e}`);
      }
    },

    async toggleHookSync(id, target, enabled) {
      const key = id + ':' + target;
      this.hookSyncBusy = { ...this.hookSyncBusy, [key]: true };
      try {
        await SetHookEnabled(id, target, enabled);
        const idx = this.hooks.findIndex(h => h.id === id);
        if (idx >= 0) {
          const h = { ...this.hooks[idx] };
          if (target === 'claude') h.enabledInClaude = enabled;
          else h.enabledInCodex = enabled;
          this.hooks = [...this.hooks.slice(0, idx), h, ...this.hooks.slice(idx + 1)];
        }
        await GetHooksStats().then(st => this.hooksStats = st).catch(() => {});
        const label = this.hookTargetLabel(target);
        this.showSuccess(enabled ? `Hook「${id}」已写入 ${label} 配置` : `Hook「${id}」已从 ${label} 配置移除`);
      } catch (e) {
        await this.loadHooks();
        this.showError(`同步 Hook 失败: ${e}`);
      } finally {
        const b = { ...this.hookSyncBusy };
        delete b[key];
        this.hookSyncBusy = b;
      }
    },

    async suggestHook() {
      const desc = (this.hookSuggestDesc || '').trim();
      if (!desc) { this.showError('请先填写钩子的自然语言描述'); return; }
      this.hookSuggestBusy = true;
      try {
        const draft = await SuggestHookFromDescription(desc);
        if (!this.hookForm) this.hookForm = this.blankHookForm();
        this.hookForm = this.hookFormFrom(draft);
        this.hookEditIsNew = true;
        this.hookEditOpen = true;
        this.showSuccess('AI 已生成钩子草稿，请确认后保存');
      } catch (e) {
        this.showError(`AI 起草失败: ${e}`);
      } finally {
        this.hookSuggestBusy = false;
      }
    },

    async optimizeHookWithAI() {
      const inst = (this.hookAIInstruction || '').trim();
      if (!inst) { this.showError('请先填写优化指令'); return; }
      if (!this.hookForm) return;
      this.hookAIBusy = true;
      try {
        const draft = await SuggestHookEdit(this.hookFromForm(), inst);
        this.hookForm = this.hookFormFrom(draft);
        this.hookAIInstruction = '';
        this.showSuccess('AI 已更新草稿，请确认后保存');
      } catch (e) {
        this.showError(`AI 优化失败: ${e}`);
      } finally {
        this.hookAIBusy = false;
      }
    },

    // ── AI providers ──────────────────────────────────────────────────
    async loadAIProviders() {
      try {
        const r = await GetAIProviders();
        this.aiProviders = r.providers || [];
        this.activeProviderID = r.activeProviderId || '';
      } catch (_) {}
    },

    openProviderEditor(provider) {
      if (provider) {
        this.providerForm = { ...provider, _isNew: false };
      } else {
        this.providerForm = { id: '', name: '', baseUrl: '', apiKey: '', model: '', _isNew: true };
      }
      this.providerTestResult = '';
    },

    closeProviderEditor() {
      this.providerForm = null;
      this.providerTestResult = '';
    },

    fillProviderPreset(preset) {
      if (!this.providerForm) this.openProviderEditor(null);
      const presets = {
        deepseek: { name: 'DeepSeek',  baseUrl: 'https://api.deepseek.com/v1',               model: 'deepseek-chat' },
        glm:      { name: 'GLM',       baseUrl: 'https://open.bigmodel.cn/api/paas/v4',       model: 'glm-4-plus' },
        openai:   { name: 'OpenAI',    baseUrl: 'https://api.openai.com/v1',                  model: 'gpt-4o-mini' },
        claude:   { name: 'Claude',    baseUrl: 'https://api.anthropic.com/v1',               model: 'claude-opus-5-5' },
        kimi:     { name: 'Kimi',      baseUrl: 'https://api.moonshot.cn/v1',                 model: 'moonshot-v1-8k' },
        qwen:     { name: 'Qwen',      baseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1', model: 'qwen-plus' },
      };
      const p = presets[preset];
      if (!p) return;
      this.providerForm.name    = this.providerForm.name    || p.name;
      this.providerForm.baseUrl = p.baseUrl;
      this.providerForm.model   = this.providerForm.model   || p.model;
    },

    async saveProviderForm() {
      if (!this.providerForm) return;
      const f = this.providerForm;
      if (!f.name || !f.baseUrl || !f.apiKey || !f.model) {
        this.showError('请填写名称、URL、Key 和模型名');
        return;
      }
      const provider = {
        id:      f.id || `p-${Date.now()}`,
        name:    f.name.trim(),
        baseUrl: f.baseUrl.trim(),
        apiKey:  f.apiKey.trim(),
        model:   f.model.trim(),
      };
      const existing = this.aiProviders.findIndex(p => p.id === provider.id);
      let updated;
      if (existing >= 0) {
        updated = this.aiProviders.map((p, i) => i === existing ? provider : p);
      } else {
        updated = [...this.aiProviders, provider];
      }
      const activeID = this.activeProviderID || provider.id;
      try {
        await SaveAIProviders(updated, activeID);
        this.aiProviders = updated;
        this.activeProviderID = activeID;
        this.closeProviderEditor();
        this.showSuccess(`提供商「${provider.name}」已保存`);
      } catch (e) { this.showError(`保存失败: ${e}`); }
    },

    async setActiveProvider(id) {
      try {
        await SaveAIProviders(this.aiProviders, id);
        this.activeProviderID = id;
        const p = this.aiProviders.find(p => p.id === id);
        this.showSuccess(`已切换到「${p?.name || id}」`);
      } catch (e) { this.showError(`切换失败: ${e}`); }
    },

    async deleteProvider(id) {
      const p = this.aiProviders.find(p => p.id === id);
      if (!(await this.askConfirm(`删除 AI 提供商「${p?.name || id}」？其 API Key 也会一并删除。`))) return;
      const updated = this.aiProviders.filter(p => p.id !== id);
      const newActive = this.activeProviderID === id
        ? (updated[0]?.id || '')
        : this.activeProviderID;
      try {
        await SaveAIProviders(updated, newActive);
        this.aiProviders = updated;
        this.activeProviderID = newActive;
        if (this.providerForm?.id === id) this.closeProviderEditor();
        this.showSuccess(`已删除「${p?.name || id}」`);
      } catch (e) { this.showError(`删除失败: ${e}`); }
    },

    async testProviderForm() {
      if (!this.providerForm) return;
      const f = this.providerForm;
      this.providerTesting = true;
      this.providerTestResult = '';
      try {
        const result = await TestAIConfig({
          enabled: true,
          baseUrl: f.baseUrl.trim(),
          apiKey:  f.apiKey.trim(),
          model:   f.model.trim(),
        });
        this.providerTestResult = `✓ ${result}`;
      } catch (e) {
        this.providerTestResult = `✗ ${e}`;
      } finally {
        this.providerTesting = false;
      }
    },

    // ── App data paths ────────────────────────────────────────────────
    async loadDataPaths() {
      try {
        this.dataPaths = await GetAppDataPaths();
        this.dataPathsForm = {
          traceLogDir: this.dataPaths.traceLogDir || '',
          skillsDir: this.dataPaths.skillsDir || '',
        };
      } catch (_) {}
    },

    async saveDataPaths() {
      this.dataPathsSaving = true;
      try {
        await SaveAppDataPaths(
          (this.dataPathsForm.traceLogDir || '').trim(),
          (this.dataPathsForm.skillsDir   || '').trim(),
        );
        await this.loadDataPaths();
        this.showSuccess('路径配置已保存');
      } catch (e) { this.showError(`保存路径失败: ${e}`); }
      finally { this.dataPathsSaving = false; }
    },

    async resetDataPaths() {
      this.dataPathsForm = { traceLogDir: '', skillsDir: '' };
      await this.saveDataPaths();
    },

    // ── AI trace log ──────────────────────────────────────────────────
    async loadAITraces() {
      try {
        this.aiTraces = await GetAITraces() || [];
      } catch (_) {}
      await this.loadTraceStorage();
    },

    async loadTraceStorage() {
      try {
        this.aiTraceStorage = await GetAITraceStorageInfo();
      } catch (_) {}
    },

    async clearAITracesLog() {
      if (!(await this.askConfirm('清空全部 AI 调用日志？此操作不可撤销。', { title: '确认清空', confirmText: '清空' }))) return;
      try {
        await ClearAITraces();
        this.aiTraces = [];
        this.aiTraceExpanded = {};
        this.aiTraceStorage = null;
        await this.loadTraceStorage();
        this.showSuccess('调用日志已清空');
      } catch (_) {}
    },

    async saveRetainDays(days) {
      const d = parseInt(days, 10);
      if (!d || d < 1) return;
      this.aiRetainDaysSaving = true;
      try {
        await SaveAIConfig({ ...this.aiForm, traceRetainDays: d });
        this.aiForm.traceRetainDays = d;
        await this.loadTraceStorage();
        this.showSuccess(`日志保留 ${d} 天`);
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.aiRetainDaysSaving = false;
      }
    },

    toggleAITrace(id) {
      this.aiTraceExpanded = { ...this.aiTraceExpanded, [id]: !this.aiTraceExpanded[id] };
    },

    formatBytes(b) {
      if (b < 1024) return b + ' B';
      if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB';
      return (b / 1024 / 1024).toFixed(1) + ' MB';
    },

    // ── Claude Code / Codex usage status ──────────────────────────────
    // ── AI workspace profiles ─────────────────────────────────────────
    async loadWorkspace() {
      this.workspaceLoading = true;
      try {
        const [profiles, snapshots, health, skills, mcps, hooks] = await Promise.all([
          ListAIWorkspaceProfiles(), ListAIWorkspaceSnapshots(), GetAIEnvironmentHealth(),
          this.skills.length ? Promise.resolve(this.skills) : GetSkills(),
          this.mcpInventory ? Promise.resolve(null) : this.loadMCP(),
          this.hooks.length ? Promise.resolve(this.hooks) : ListHooks(),
          // New profiles start from the current agent settings; without these
          // the form would default to empty and applying would reset them.
          this.loadClaudeCodeConfig(), this.loadCodexModelConfig(),
        ]);
        this.workspaceProfiles = profiles || [];
        this.workspaceSnapshots = snapshots || [];
        this.workspaceHealth = health;
        this.skills = skills || [];
        this.hooks = hooks || [];
      } catch (e) { this.showError(`加载工作区失败: ${e}`); }
      finally { this.workspaceLoading = false; }
    },

    workspaceSelections(items, key) {
      return (items || []).map(item => ({ id: item[key], claude: !!item.inClaude || !!item.enabledInClaude, codex: !!item.inCodex || !!item.enabledInCodex }));
    },

    openWorkspaceForm(profile = null) {
      this.workspacePreview = null;
      this.workspaceForm = profile ? JSON.parse(JSON.stringify(profile)) : {
        id: '', name: '', rootPath: '', claudeModel: this.claudeCodeConfig.model || '', claudeEffort: this.claudeCodeConfig.effortLevel || '',
        codexModel: this.codexModelConfig.model || '', codexEffort: this.codexModelConfig.effortLevel || '', codexGoals: !!this.codexModelConfig.goalsEnabled,
        skills: this.workspaceSelections(this.skills, 'name'), mcpServers: this.workspaceSelections(this.mcpServers, 'id'), hooks: this.workspaceSelections(this.hooks, 'id'),
      };
      this.workspaceFormOpen = true;
      this.$nextTick(() => document.querySelector('.workspace-editor')?.scrollIntoView({ behavior: 'smooth', block: 'start' }));
    },

    workspaceSelection(list, id) {
      return (list || []).find(x => x.id === id) || { id, claude: false, codex: false };
    },

    toggleWorkspaceSelection(kind, id, target, enabled) {
      const list = [...(this.workspaceForm[kind] || [])];
      let item = list.find(x => x.id === id);
      if (!item) { item = { id, claude: false, codex: false }; list.push(item); }
      item[target] = enabled;
      this.workspaceForm = { ...this.workspaceForm, [kind]: list.filter(x => x.claude || x.codex) };
    },

    async saveWorkspaceProfile() {
      if (!this.workspaceForm?.name?.trim() || !this.workspaceForm?.rootPath?.trim()) { this.showError('请填写工作区名称和目录'); return; }
      try {
        const saved = await SaveAIWorkspaceProfile(this.workspaceForm);
        this.workspaceForm = saved;
        this.workspaceFormOpen = false;
        await this.loadWorkspace();
        this.showSuccess('工作区画像已保存');
      } catch (e) { this.showError(`保存工作区失败: ${e}`); }
    },

    async previewWorkspaceProfile(profile) {
      this.workspacePreviewing = true;
      this.workspacePreview = null;
      try { this.workspacePreview = await PreviewAIWorkspaceProfileApply(profile.id); }
      catch (e) { this.showError(`预检失败: ${e}`); }
      finally { this.workspacePreviewing = false; }
      this.$nextTick(() => document.querySelector('.workspace-preview')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' }));
    },

    async applyWorkspaceProfile() {
      const profile = this.workspacePreview?.profile;
      if (!profile?.id || !this.workspacePreview?.canApply) return;
      this.workspaceApplying = true;
      try {
        await ApplyAIWorkspaceProfile(profile.id);
        this.workspacePreview = null;
        await this.reloadAfterWorkspaceChange();
        this.showSuccess(`已应用工作区画像「${profile.name}」`);
      } catch (e) { this.showError(`应用失败: ${e}`); }
      finally { this.workspaceApplying = false; }
    },

    async deleteWorkspaceProfile(profile) {
      if (!(await this.askConfirm(`删除工作区画像「${profile.name}」？不会改动当前 Claude/Codex 配置。`))) return;
      try { await DeleteAIWorkspaceProfile(profile.id); await this.loadWorkspace(); this.showSuccess('工作区画像已删除'); }
      catch (e) { this.showError(`删除失败: ${e}`); }
    },

    // Apply / rollback rewrite skills, hooks, MCP and agent settings, so drop
    // the caches loadWorkspace would otherwise reuse and refresh the hero.
    async reloadAfterWorkspaceChange() {
      this.skills = [];
      this.hooks = [];
      this.mcpInventory = null;
      await this.loadWorkspace();
      this.loadAIHeroStats();
    },

    async rollbackWorkspaceSnapshot(snapshot) {
      if (!(await this.askConfirm(`回滚到「${snapshot.profileName}」应用前的状态？当前 Hatch 管理的 AI 配置会被恢复。`, { title: '回滚 AI 工作区', confirmText: '回滚' }))) return;
      try { await RollbackAIWorkspaceSnapshot(snapshot.id); await this.reloadAfterWorkspaceChange(); this.showSuccess('已回滚 AI 工作区快照'); }
      catch (e) { this.showError(`回滚失败: ${e}`); }
    },

    // ── AI tool status ─────────────────────────────────────────────────
    async loadAIUsageStatuses() {
      if (!this.claudeStatus) this.loadClaudeStatus();
      if (!this.codexStatus) this.loadCodexStatus();
      if (!this.skills.length) this.loadSkills();
      this.loadClaudeCodeConfig();
      this.loadClaudeCodePresets();
      this.loadCodexModelConfig();
      this.loadCodexPresets();
    },

    async loadClaudeCodeConfig() {
      try {
        this.claudeCodeConfig = await window['go']['app']['App']['GetClaudeCodeConfig']();
      } catch (e) { /* silent */ }
    },

    async saveClaudeCodeConfig() {
      this.claudeCodeConfigSaving = true;
      try {
        await window['go']['app']['App']['SetClaudeCodeConfig'](
          this.claudeCodeConfig.model || '',
          this.claudeCodeConfig.effortLevel || ''
        );
        await this.loadClaudeCodeConfig();   // refresh effectiveModel (alias resolution)
        this.showSuccess('Claude Code 配置已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.claudeCodeConfigSaving = false;
      }
    },

    async loadClaudeCodePresets() {
      try {
        this.claudeCodePresets = await window['go']['app']['App']['GetClaudeCodePresets']() || [];
      } catch (e) { /* silent */ }
    },

    openNewPreset() {
      this.claudeCodePresetForm = {
        id: '',
        name: '',
        model: this.claudeCodeConfig.model || '',
        effortLevel: this.claudeCodeConfig.effortLevel || '',
      };
      this.claudeCodePresetOpen = true;
    },

    openEditPreset(preset) {
      this.claudeCodePresetForm = { ...preset };
      this.claudeCodePresetOpen = true;
    },

    async saveClaudeCodePreset() {
      if (!this.claudeCodePresetForm.name.trim()) {
        this.showError('预设名称不能为空');
        return;
      }
      try {
        await window['go']['app']['App']['SaveClaudeCodePreset'](this.claudeCodePresetForm);
        this.claudeCodePresetOpen = false;
        await this.loadClaudeCodePresets();
        this.showSuccess('预设已保存');
      } catch (e) {
        this.showError(`保存预设失败: ${e}`);
      }
    },

    async deleteClaudeCodePreset(id) {
      const preset = this.claudeCodePresets.find(p => p.id === id);
      if (!(await this.askConfirm(`删除 Claude Code 预设「${preset?.name || id}」？`))) return;
      try {
        await window['go']['app']['App']['DeleteClaudeCodePreset'](id);
        await this.loadClaudeCodePresets();
        this.showSuccess('预设已删除');
      } catch (e) {
        this.showError(`删除预设失败: ${e}`);
      }
    },

    async applyClaudeCodePreset(preset) {
      this.claudeCodeConfig = {
        ...this.claudeCodeConfig,
        model: preset.model,
        effortLevel: preset.effortLevel,
      };
      await this.saveClaudeCodeConfig();
    },

    isKnownClaudeModel(model) {
      return this.claudeModels.some(item => item.value === model);
    },

    claudeEffortOptions(model) {
      const all = ['low', 'medium', 'high', 'xhigh', 'max'];
      if (!model) return all;   // 系统默认: don't constrain until a model is chosen
      const item = this.claudeModels.find(candidate => candidate.value === model);
      return item ? item.efforts : all;
    },

    normalizeClaudeEffort(config) {
      if (!config?.effortLevel || this.claudeEffortOptions(config.model).includes(config.effortLevel)) return;
      config.effortLevel = '';
    },

    async loadCodexModelConfig() {
      try {
        this.codexModelConfig = await window['go']['app']['App']['GetCodexModelConfig']();
      } catch (e) { /* silent */ }
    },

    async saveCodexModelConfig() {
      this.codexModelConfigSaving = true;
      try {
        await window['go']['app']['App']['SetCodexModelConfig'](
          this.codexModelConfig.model || '',
          this.codexModelConfig.effortLevel || '',
          !!this.codexModelConfig.goalsEnabled
        );
        this.showSuccess('Codex 配置已保存');
        if (this.codexConfigOpen) await this.openCodexConfig();
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.codexModelConfigSaving = false;
      }
    },

    async loadCodexPresets() {
      try {
        this.codexPresets = await window['go']['app']['App']['GetCodexPresets']() || [];
      } catch (e) { /* silent */ }
    },

    openNewCodexPreset() {
      this.codexPresetForm = {
        id: '',
        name: '',
        model: this.codexModelConfig.model || '',
        effortLevel: this.codexModelConfig.effortLevel || '',
      };
      this.codexPresetOpen = true;
    },

    openEditCodexPreset(preset) {
      this.codexPresetForm = { ...preset };
      this.codexPresetOpen = true;
    },

    async saveCodexPreset() {
      if (!this.codexPresetForm.name.trim()) {
        this.showError('预设名称不能为空');
        return;
      }
      try {
        await window['go']['app']['App']['SaveCodexPreset'](this.codexPresetForm);
        this.codexPresetOpen = false;
        await this.loadCodexPresets();
        this.showSuccess('预设已保存');
      } catch (e) {
        this.showError(`保存预设失败: ${e}`);
      }
    },

    async deleteCodexPreset(id) {
      const preset = this.codexPresets.find(p => p.id === id);
      if (!(await this.askConfirm(`删除 Codex 预设「${preset?.name || id}」？`))) return;
      try {
        await window['go']['app']['App']['DeleteCodexPreset'](id);
        await this.loadCodexPresets();
        this.showSuccess('预设已删除');
      } catch (e) {
        this.showError(`删除预设失败: ${e}`);
      }
    },

    async applyCodexPreset(preset) {
      this.codexModelConfig = {
        ...this.codexModelConfig,
        model: preset.model,
        effortLevel: preset.effortLevel,
      };
      await this.saveCodexModelConfig();
    },

    isKnownCodexModel(model) {
      return this.codexModels.some(item => item.value === model);
    },

    codexEffortOptions(model) {
      const all = ['low', 'medium', 'high', 'xhigh', 'max', 'ultra'];
      const item = this.codexModels.find(candidate => candidate.value === model);
      return item ? item.efforts : all;
    },

    normalizeCodexEffort(config) {
      if (!config?.effortLevel || this.codexEffortOptions(config.model).includes(config.effortLevel)) return;
      config.effortLevel = '';
    },

    effortLabel(level) {
      return { low: 'low', medium: 'medium', high: 'high', xhigh: 'xhigh', max: 'max', ultra: 'ultra' }[level] || '系统默认';
    },

    modelLabel(model) {
      return this.claudeModels.find(item => item.value === model)?.label || model || '默认';
    },

    codexModelLabel(model) {
      const label = this.codexModels.find(item => item.value === model)?.label;
      return label ? label.split(' · ')[0] : (model || '默认');
    },

    async loadClaudeStatus() {
      this.claudeStatusLoading = true;
      try {
        this.claudeStatus = await GetClaudeCodeStatus();
      } catch (e) {
        this.showError(`获取 Claude Code 状态失败: ${e}`);
      } finally {
        this.claudeStatusLoading = false;
      }
    },

    // ↻ on the Claude Code card: run a one-line Haiku prompt via the claude CLI so
    // the 5h/7d usage cache refreshes without opening an interactive session.
    async refreshClaudeUsage() {
      if (this.claudeStatusLoading) return;
      this.claudeStatusLoading = true;
      try {
        this.claudeStatus = await window['go']['app']['App']['RefreshClaudeUsage']();
        this.showSuccess('已通过 Haiku 刷新 Claude 用量');
      } catch (e) {
        this.showError(`刷新 Claude 用量失败: ${e}`);
        try { this.claudeStatus = await GetClaudeCodeStatus(); } catch {}
      } finally {
        this.claudeStatusLoading = false;
      }
    },

    async loadCodexStatus() {
      this.codexStatusLoading = true;
      try {
        this.codexStatus = await GetCodexUsageStatus();
      } catch (e) {
        this.showError(`获取 Codex 用量失败: ${e}`);
      } finally {
        this.codexStatusLoading = false;
      }
    },

    usagePill(pct) {
      if (pct >= 80) return 'usage-high';
      if (pct >= 50) return 'usage-mid';
      return 'usage-low';
    },

    // ── Config file editors ────────────────────────────────────────────
    async openClaudeSettings() {
      this.claudeSettingsOpen = true;
      try {
        this.claudeSettingsContent = await GetClaudeSettings();
      } catch (e) {
        this.showError(`读取 Claude settings 失败: ${e}`);
      }
    },

    async saveClaudeSettings() {
      this.claudeSettingsSaving = true;
      try {
        await SaveClaudeSettings(this.claudeSettingsContent);
        this.showSuccess('Claude Code 配置已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.claudeSettingsSaving = false;
      }
    },

    async openCodexConfig() {
      this.codexConfigOpen = true;
      try {
        this.codexConfigContent = await GetCodexConfig();
      } catch (e) {
        this.showError(`读取 Codex 配置失败: ${e}`);
      }
    },

    async saveCodexConfig() {
      this.codexConfigSaving = true;
      try {
        await SaveCodexConfig(this.codexConfigContent);
        this.showSuccess('Codex 配置已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.codexConfigSaving = false;
      }
    },

    // ── Helpers ────────────────────────────────────────────────────────
    showError(msg) {
      this.error = msg;
      clearTimeout(this.errorTimer);
      if (msg) {
        this.errorTimer = setTimeout(() => { this.error = ''; this.errorTimer = null; }, 6000);
      } else {
        this.errorTimer = null;
      }
    },

    showSuccess(msg) {
      this.success = msg;
      clearTimeout(this.successTimer);
      this.successTimer = setTimeout(() => { this.success = ''; this.successTimer = null; }, 3000);
    },

    // ── Shared helpers: clipboard + AI overview + bulk actions ──────────
    async copyToClipboard(text, label) {
      try {
        await navigator.clipboard.writeText(text || '');
        this.showSuccess(`已复制${label ? '：' + label : ''}`);
      } catch (_) {
        this.showError('复制失败');
      }
    },

    async exportHatchBackup() {
      this.backupBusy = true;
      try {
        this.backupResult = await window['go']['app']['App']['ExportBackup']('');
        this.showSuccess('备份已导出');
      } catch (e) { this.showError(`导出失败: ${e}`); }
      finally { this.backupBusy = false; }
    },

    async importHatchBackup() {
      const path = this.backupImportPath.trim();
      if (!path) return;
      if (!(await this.askConfirm(`用「${path}」覆盖当前配置和 Skills？当前 config.json 会先自动备份。`, { title: '从备份恢复', confirmText: '恢复' }))) return;
      this.backupBusy = true;
      try {
        const msg = await window['go']['app']['App']['ImportBackup'](path);
        this.showSuccess(msg || '已恢复');
        this.backupImportPath = '';
        await this.loadConfig();
        this.loadDataInventory();
      } catch (e) { this.showError(`恢复失败: ${e}`); }
      finally { this.backupBusy = false; }
    },

    async loadOverview() {
      try { await this.loadAIUsageStatuses(); } catch (_) {}
      try { if (!this.aiProviders.length) await this.loadAIProviders(); } catch (_) {}
      try { this.skillsStats = await GetSkillsStats(); } catch (_) {}
      try { if (!this.mcpInventory) await this.loadMCP(); } catch (_) {}
      try { this.hooksStats = await GetHooksStats(); } catch (_) {}
      try { if (!this.docLinks.length) await this.loadDocLinks(); } catch (_) {}
      try { if (!this.aiTraces.length) await this.loadAITraces(); } catch (_) {}
    },

    duplicateProvider(p) {
      this.openProviderEditor(null);
      // Cloning is for "same endpoint + key, different model", so the key comes along.
      this.providerForm = { id: '', name: `${p.name || p.model || '提供商'} 副本`, baseUrl: p.baseUrl || '', model: p.model || '', apiKey: p.apiKey || '', _isNew: true };
      this.providerTestResult = '';
    },

    async bulkSkillSync(target, enabled) {
      const list = this.filteredSkills || [];
      if (!list.length) { this.showError('没有可操作的 Skill'); return; }
      let n = 0;
      for (const s of list) {
        const already = target === 'claude' ? s.inClaude : s.inCodex;
        if (already === enabled) continue;
        try { await SetSkillSync(s.name, target, enabled); n++; } catch (_) {}
      }
      await this.loadSkills();
      const label = target === 'claude' ? 'Claude' : 'Codex';
      this.showSuccess(enabled ? `已同步 ${n} 个 Skill 到 ${label}` : `已从 ${label} 取消 ${n} 个 Skill`);
    },

    async bulkSkillUnsyncAll() {
      const list = this.filteredSkills || [];
      if (!list.length) { this.showError('没有可操作的 Skill'); return; }
      let n = 0;
      for (const s of list) {
        if (s.inClaude) { try { await SetSkillSync(s.name, 'claude', false); n++; } catch (_) {} }
        if (s.inCodex)  { try { await SetSkillSync(s.name, 'codex',  false); n++; } catch (_) {} }
      }
      await this.loadSkills();
      this.showSuccess(`已取消 ${n} 项同步`);
    },

    langIcon(lang) {
      return { python: '🐍', shell: '🐚', ruby: '💎', node: '🟩', binary: '⚙️' }[lang] || '📄';
    },

    cmdLabel(cmd) {
      return {
        start:   '▶ start',
        stop:    '■ stop',
        restart: '↺ restart',
        reload:  '↺ reload',
        status:  '? status',
      }[cmd] || cmd;
    },

    // List subtitle: description, else the file name when it differs from the
    // display name, else the kind of file — never the name repeated.
    wbItemSub(s) {
      if (s.desc) return s.desc;
      const file = this.baseName(s.path);
      if (file !== s.name) return file;
      const lang = { shell: 'Shell 脚本', python: 'Python 脚本', ruby: 'Ruby 脚本', node: 'Node 脚本', binary: '可执行文件' }[s.language];
      return lang || '脚本';
    },

    baseName(path) {
      if (!path) return '';
      return path.split('/').filter(Boolean).pop() || path;
    },

    // Shell-like split: whitespace separates, '…' and "…" group, \ escapes.
    splitArgs(raw) {
      const out = [];
      let cur = '', quote = '', has = false;
      const text = String(raw || '');
      for (let i = 0; i < text.length; i++) {
        const c = text[i];
        if (quote) {
          if (c === quote) quote = '';
          else if (c === '\\' && quote === '"' && i + 1 < text.length) cur += text[++i];
          else cur += c;
        } else if (c === '"' || c === "'") { quote = c; has = true; }
        else if (c === '\\' && i + 1 < text.length) { cur += text[++i]; has = true; }
        else if (/\s/.test(c)) { if (has || cur) out.push(cur); cur = ''; has = false; }
        else { cur += c; has = true; }
      }
      if (has || cur) out.push(cur);
      return out;
    },

    // Safe DOM id from a path (avoids btoa encoding issues)
    safeId(path) {
      return path.replace(/[^a-zA-Z0-9]/g, '_');
    },

    normalizePresetList(presets) {
      return (presets || [])
        .map(p => ({
          name: String(p?.name ?? p?.Name ?? '').trim(),
          args: String(p?.args ?? p?.Args ?? '').trim(),
        }))
        .filter(p => p.name || p.args)
        .map(p => ({ ...p, name: p.name || p.args }));
    },

    normalizedPresets(script) {
      return this.normalizePresetList(script?.presets || script?.Presets || []);
    },

    // ── Doc links ──────────────────────────────────────────────────────────
    async loadDocLinks() {
      try {
        this.docLinks = await GetDocLinks() || [];
      } catch (_) {}
    },

    async saveDocLinks() {
      this.docLinksSaving = true;
      try {
        await SaveDocLinks(this.docLinks);
        this.showSuccess('文档链接已保存');
      } catch (e) {
        this.showError(`保存文档链接失败: ${e}`);
      } finally {
        this.docLinksSaving = false;
      }
    },

    async addDocLink() {
      const path = (this.docLinkForm.path || '').trim();
      if (!path) return;
      const label = (this.docLinkForm.label || '').trim() || path.split('/').pop();
      this.docLinks = [...this.docLinks, { path, label }];
      this.docLinkForm = { path: '', label: '' };
      await this.saveDocLinks();
    },

    async removeDocLink(idx) {
      const link = this.docLinks[idx];
      if (!(await this.askConfirm(`移除文档「${link?.label || link?.path || ''}」？只移除链接，不删除文件。`, { title: '确认移除', confirmText: '移除' }))) return;
      this.docLinks = this.docLinks.filter((_, i) => i !== idx);
      await this.saveDocLinks();
    },

    startEditDocLink(idx) {
      const link = this.docLinks[idx];
      this.editingDocLinkIdx = idx;
      this.editingDocLinkForm = { path: link.path, label: link.label };
    },

    cancelEditDocLink() {
      this.editingDocLinkIdx = -1;
      this.editingDocLinkForm = { path: '', label: '' };
    },

    async saveEditDocLink() {
      const path = (this.editingDocLinkForm.path || '').trim();
      if (!path) return;
      const label = (this.editingDocLinkForm.label || '').trim() || path.split('/').pop();
      const newLinks = [...this.docLinks];
      newLinks[this.editingDocLinkIdx] = { path, label };
      this.docLinks = newLinks;
      this.editingDocLinkIdx = -1;
      this.editingDocLinkForm = { path: '', label: '' };
      await this.saveDocLinks();
    },

    async openDocFile(path) {
      try {
        await OpenFile(path);
      } catch (e) {
        this.showError(`打开文件失败: ${e}`);
      }
    },

    docLinkIcon(path) {
      const ext = (path || '').split('.').pop().toLowerCase();
      if (ext === 'pdf') return '📕';
      if (['html', 'htm'].includes(ext)) return '🌐';
      if (['md', 'markdown'].includes(ext)) return '📝';
      if (['txt'].includes(ext)) return '📄';
      return '📎';
    },

    // ── Memo module ────────────────────────────────────────────────────────
    async loadMemoModule() {
      try {
        const [dash, memos, vocab, settings] = await Promise.all([
          GetMemoDashboard(),
          ListMemos(),
          ListVocabEntries(),
          GetMemoSettings(),
        ]);
        this.memoDashboard = dash || this.memoDashboard;
        this.memos = memos || [];
        this.vocabEntries = vocab || [];
        this.memoSettings = settings || this.memoSettings;
        this.loadMemoImagePreviews();
      } catch (e) {
        this.showError(`加载备忘录失败: ${e}`);
      }
    },

    todayDate() {
      const d = new Date();
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      const dd = String(d.getDate()).padStart(2, '0');
      return `${d.getFullYear()}-${mm}-${dd}`;
    },

    unixToLocalInput(unix) {
      if (!unix) return '';
      const d = new Date(unix * 1000);
      const pad = (n) => String(n).padStart(2, '0');
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    },

    localInputToUnix(value) {
      if (!value) return 0;
      const ts = new Date(value).getTime();
      return Number.isFinite(ts) ? Math.floor(ts / 1000) : 0;
    },

    parseTags(raw) {
      return (raw || '').split(/[,，\s]+/).map(s => s.trim()).filter(Boolean);
    },

    tagsRaw(tags) {
      return (tags || []).join(', ');
    },

    // ── Memo: navigation + labels ─────────────────────────────────────
    memoGo(sub) {
      this.memoSubTab = sub;
      this.memoSearch = '';
      if (!['today', 'upcoming', 'done'].includes(sub)) this.memoDraft = null;
      if (sub !== 'words') this.vocabDraft = null;
      this.loadMemoModule();
      if (sub === 'today' || sub === 'upcoming') this.$nextTick(() => this.$refs.memoQuickInput?.focus());
    },

    // Undone first by priority (high → low), then reminder time, then newest.
    memoOrder(a, b) {
      const rank = { high: 0, normal: 1, low: 2 };
      return (rank[a.priority || 'normal'] - rank[b.priority || 'normal'])
        || ((a.reminderAt || Infinity) - (b.reminderAt || Infinity))
        || ((b.createdAt || 0) - (a.createdAt || 0));
    },

    ymd(d) {
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
    },

    memoDateLabel(date, long = false) {
      if (!date) return '无日期';
      const today = new Date(); today.setHours(0, 0, 0, 0);
      const d = new Date(date + 'T00:00:00');
      if (Number.isNaN(d.getTime())) return date;
      const diff = Math.round((d - today) / 86400000);
      const week = '日一二三四五六'[d.getDay()];
      const md = `${d.getMonth() + 1}月${d.getDate()}日`;
      if (diff === 0) return '今天';
      if (diff === 1) return long ? `明天 · ${md}` : '明天';
      if (diff === 2) return long ? `后天 · ${md}` : '后天';
      if (diff === -1) return '昨天';
      if (diff > 2 && diff < 7) return long ? `周${week} · ${md}` : `周${week}`;
      if (diff < -1) return long ? md : `${-diff} 天前`;
      return long ? `${md} · 周${week}` : md;
    },

    // Time only when the reminder falls on the todo's own date (or today).
    memoTimeLabel(unix, date = '') {
      if (!unix) return '';
      const d = new Date(unix * 1000);
      const hm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
      const day = this.ymd(d);
      return day === date || day === this.todayDate() ? hm : `${this.memoDateLabel(day)} ${hm}`;
    },

    memoIsOverdue(item) {
      return item.status !== 'done' && !!item.date && item.date < this.todayDate();
    },

    memoReminderPast(item) {
      return item.status !== 'done' && item.reminderAt && item.reminderAt * 1000 < Date.now();
    },

    // ── Memo: quick add with natural-language parsing ─────────────────
    // "明天 9:30 交周报 #工作 !高" → { title, date, time, tags, priority }.
    // Recognises 今天/明天/后天/大后天, (下)周X/星期X, YYYY-MM-DD, M月D日, M/D,
    // HH:MM, (上午|下午|晚上)H点(半|M分), #tag, !高/!低/!!.
    parseQuickTodo(text, now = new Date()) {
      let rest = ` ${String(text || '')} `;
      const out = { title: '', date: this.ymd(now), time: '', tags: [], priority: 'normal' };
      const take = (re, fn) => { rest = rest.replace(re, (...m) => { fn(...m); return ' '; }); };
      const base = new Date(now); base.setHours(0, 0, 0, 0);
      const setDays = n => { const d = new Date(base); d.setDate(d.getDate() + n); out.date = this.ymd(d); };

      take(/\s#([^\s#!]+)/g, (_, t) => out.tags.push(t));
      take(/\s(?:!{2,}|!(?:高|high|急|1))(?=\s)/gi, () => { out.priority = 'high'; });
      take(/\s!(?:低|low|3)(?=\s)/gi, () => { out.priority = 'low'; });
      // 今晚 / 明早 carry both a day and a part of day ("今晚8点" → today 20:00).
      let impliedPart = '', hadPart = false;
      take(/(今晚|今早|明晚|明早)/, (_, w) => { setDays(w[0] === '明' ? 1 : 0); impliedPart = w[1] === '晚' ? '晚上' : '早上'; });
      take(/(大后天|后天|明天|明日|今天|今日)/, (_, w) => setDays({ 今天: 0, 今日: 0, 明天: 1, 明日: 1, 后天: 2, 大后天: 3 }[w]));
      take(/(下)?(?:周|星期|礼拜)([一二三四五六日天])/, (_, next, w) => {
        const target = '一二三四五六日'.indexOf(w === '天' ? '日' : w); // Monday = 0
        const todayIdx = (base.getDay() + 6) % 7;
        if (next) { setDays(7 - todayIdx + target); return; }  // that day of next week
        const diff = target - todayIdx;
        setDays(diff > 0 ? diff : diff + 7);                    // next occurrence
      });
      take(/(\d{4})-(\d{1,2})-(\d{1,2})/, (_, y, m, d) => { out.date = `${y}-${m.padStart(2, '0')}-${d.padStart(2, '0')}`; });
      take(/(\d{1,2})月(\d{1,2})[日号]?/, (_, m, d) => {
        let y = base.getFullYear();
        const cand = new Date(y, +m - 1, +d);
        if (cand < base) y += 1;
        out.date = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
      });
      take(/\s(\d{1,2})\/(\d{1,2})(?=\s)/, (_, m, d) => {
        let y = base.getFullYear();
        if (new Date(y, +m - 1, +d) < base) y += 1;
        out.date = `${y}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
      });
      take(/(凌晨|上午|早上|中午|下午|傍晚|晚上)?\s*(\d{1,2})(?::|：)(\d{2})/, (_, part, h, m) => { hadPart = !!(part || impliedPart); out.time = this.quickTime(part || impliedPart, +h, +m); });
      if (!out.time) take(/(凌晨|上午|早上|中午|下午|傍晚|晚上)?\s*(\d{1,2})\s*点\s*(半|(\d{1,2})\s*分?)?/, (_, part, h, half, m) => {
        hadPart = !!(part || impliedPart);
        out.time = this.quickTime(part || impliedPart, +h, half === '半' ? 30 : (+m || 0));
      });
      // A bare "8点" typed for today after 8:00 means the evening, not a time already past.
      if (out.time && !hadPart && out.date === this.ymd(now)) {
        const [h, m] = out.time.split(':').map(Number);
        if (h < 12 && h * 60 + m < now.getHours() * 60 + now.getMinutes()) out.time = this.quickTime('晚上', h, m);
      }
      out.title = rest.replace(/\s+/g, ' ').trim();
      return out;
    },

    quickTime(part, h, m) {
      if (['下午', '傍晚', '晚上'].includes(part) && h < 12) h += 12;
      if (part === '中午' && h < 6) h += 12;
      if (h > 23 || m > 59) return '';
      return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
    },

    async addQuickMemo() {
      const p = this.memoQuickParsed;
      if (!p.title) { this.showError('写一点待办内容'); return; }
      const reminderAt = p.time ? this.localInputToUnix(`${p.date}T${p.time}`) : 0;
      try {
        const saved = await SaveMemo({
          id: '', title: p.title, body: '', date: p.date, status: 'todo', priority: p.priority,
          tags: p.tags, reminderAt, reminderMode: 'both', attachments: [],
        });
        this.memos = [saved, ...this.memos];
        this.memoQuick = '';
        const where = p.date === this.todayDate() ? '今天' : this.memoDateLabel(p.date, true);
        this.showSuccess(`已添加到${where}${p.time ? ' · ' + p.time + ' 提醒' : ''}`);
      } catch (e) {
        this.showError(`添加失败: ${e}`);
      }
    },

    // ── Memo: detail panel (auto-save) ────────────────────────────────
    selectMemo(item) {
      this.memoDraft = {
        ...item,
        tagsRaw: this.tagsRaw(item.tags),
        reminderLocal: this.unixToLocalInput(item.reminderAt),
        reminderMode: item.reminderMode || 'both',
        priority: item.priority || 'normal',
        attachments: [...(item.attachments || [])],
      };
      this.memoAttachmentForm = { path: '' };
      if ((item.attachments || []).some(a => a.kind === 'image')) this.loadMemoImagePreviews();
    },

    closeMemoDetail() {
      this.memoDraft = null;
    },

    replaceMemo(saved) {
      const i = this.memos.findIndex(m => m.id === saved.id);
      if (i >= 0) this.memos.splice(i, 1, saved); else this.memos = [saved, ...this.memos];
      this.memos = [...this.memos];
    },

    async saveMemoDraft() {
      const d = this.memoDraft;
      if (!d) return;
      if (!d.title.trim()) { this.showError('标题不能为空'); return; }
      this.memoDraftSaving = true;
      try {
        const { tagsRaw, reminderLocal, ...rest } = d;
        const reminderAt = this.localInputToUnix(reminderLocal);
        const saved = await SaveMemo({
          ...rest,
          title: d.title.trim(),
          date: d.date || this.todayDate(),
          tags: this.parseTags(tagsRaw),
          reminderAt,
          // a new reminder time should fire again
          reminderSentAt: reminderAt !== (rest.reminderAt || 0) ? 0 : rest.reminderSentAt,
        });
        this.replaceMemo(saved);
        if (this.memoDraft?.id === saved.id) Object.assign(this.memoDraft, { reminderAt: saved.reminderAt, reminderSentAt: saved.reminderSentAt, updatedAt: saved.updatedAt });
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.memoDraftSaving = false;
      }
    },

    async toggleMemoDone(item) {
      try {
        const saved = await MarkMemoDone(item.id, item.status !== 'done');
        this.replaceMemo(saved);
        if (this.memoDraft?.id === saved.id) this.memoDraft.status = saved.status;
      } catch (e) {
        this.showError(`更新状态失败: ${e}`);
      }
    },

    async deleteMemoItem(item) {
      if (!(await this.askConfirm(`删除「${item.title}」及其托管附件？此操作不可撤销。`))) return;
      try {
        await DeleteMemo(item.id);
        this.memos = this.memos.filter(m => m.id !== item.id);
        if (this.memoDraft?.id === item.id) this.memoDraft = null;
        this.showSuccess('已删除');
      } catch (e) {
        this.showError(`删除失败: ${e}`);
      }
    },

    async addMemoAttachment() {
      const path = (this.memoAttachmentForm.path || '').trim();
      if (!path || !this.memoDraft) return;
      const kind = /^https?:\/\//.test(path) ? 'link' : /\.(png|jpe?g|gif|webp|heic|bmp|svg)$/i.test(path) ? 'image' : 'file';
      try {
        const att = await ImportMemoAttachment(path, '', kind);
        this.memoDraft.attachments = [...this.memoDraft.attachments, att];
        this.memoAttachmentForm = { path: '' };
        await this.saveMemoDraft();
        if (att.kind === 'image') this.loadMemoImagePreviews();
      } catch (e) {
        this.showError(`添加附件失败: ${e}`);
      }
    },

    async removeMemoAttachment(att) {
      const d = this.memoDraft;
      if (!d) return;
      if (!(await this.askConfirm(`移除附件「${att.label || att.url || '附件'}」？托管目录中的副本也会被删除。`, { title: '移除附件', confirmText: '移除' }))) return;
      try {
        const saved = await DeleteMemoAttachment(d.id, att.id);
        this.replaceMemo(saved);
        d.attachments = saved.attachments || [];
      } catch (e) {
        this.showError(`移除附件失败: ${e}`);
      }
    },

    async openMemoAttachment(att) {
      try {
        await OpenMemoAttachment(att);
      } catch (e) {
        this.showError(`打开附件失败: ${e}`);
      }
    },

    async loadMemoImagePreviews() {
      const images = [];
      for (const m of this.memos || []) {
        for (const a of (m.attachments || [])) if (a.kind === 'image' && !this.memoPreviewCache[a.id]) images.push(a);
      }
      for (const a of images.slice(0, 20)) {
        try {
          const p = await GetMemoAttachmentPreview(a);
          if (p?.dataUrl) this.memoPreviewCache = { ...this.memoPreviewCache, [a.id]: p.dataUrl };
        } catch (_) {}
      }
    },

    attachmentIcon(att) {
      if (att.kind === 'link') return '🔗';
      if (att.kind === 'image') return '🖼';
      return '📄';
    },

    // ── Memo: word book ───────────────────────────────────────────────
    vocabToDraft(v) {
      return {
        ...v,
        word: v.word || '', chinese: v.chinese || '', phonetic: v.phonetic || '', partOfSpeech: v.partOfSpeech || '',
        example: v.example || '', audioUrl: v.audioUrl || '',
        collocations: (v.collocations || []).map(c => ({ pattern: c.pattern || '', meaning: c.meaning || '', example: c.example || '' })),
        tagsRaw: this.tagsRaw(v.tags),
      };
    },

    async selectVocab(v) {
      if (this.vocabDraftDirty && this.vocabDraft?.id !== v.id &&
          !(await this.askConfirm('当前单词有未保存的修改，切换后会丢失。', { title: '放弃修改？', confirmText: '放弃修改' }))) return;
      this.vocabDraft = this.vocabToDraft(v);
      this.vocabDraftSnapshot = JSON.stringify(this.vocabDraft);
    },

    closeVocabDetail() {
      this.vocabDraft = null;
    },

    // Quick capture: existing word → open it; new word → draft + dictionary + AI.
    async startVocabCapture() {
      const word = this.vocabQuick.trim();
      if (!word) return;
      const existing = this.vocabEntries.find(v => (v.word || '').toLowerCase() === word.toLowerCase());
      if (existing) {
        this.vocabQuick = '';
        this.vocabSearch = '';
        this.vocabFilter = 'all';
        await this.selectVocab(existing);
        this.showSuccess(`「${existing.word}」已在单词本中`);
        return;
      }
      this.vocabDraft = this.vocabToDraft({ id: '', word });
      this.vocabDraftSnapshot = '';
      this.vocabQuick = '';
      await this.lookupVocab(true);
      if (this.aiReady && !this.vocabDraft.chinese) await this.suggestVocab(true);
    },

    async lookupVocab(quiet = false) {
      const d = this.vocabDraft;
      const word = (d?.word || '').trim();
      if (!word) return;
      this.vocabLookupBusy = true;
      try {
        const r = await LookupVocabWord(word);
        if (this.vocabDraft !== d) return;
        d.word = r.word || word;
        d.phonetic = d.phonetic || r.phonetic || '';
        d.partOfSpeech = d.partOfSpeech || r.partOfSpeech || '';
        d.audioUrl = d.audioUrl || r.audioUrl || '';
        d.definitions = r.definitions || [];
        if (!d.collocations.length && r.collocations?.length) d.collocations = r.collocations.map(c => ({ pattern: c.pattern || '', meaning: c.meaning || '', example: c.example || '' }));
      } catch (e) {
        if (!quiet) this.showError(`词典查询失败: ${e}`);
      } finally {
        this.vocabLookupBusy = false;
      }
    },

    async suggestVocab(quiet = false) {
      const d = this.vocabDraft;
      if (!d?.word.trim() || !this.aiReady) return;
      this.vocabAIBusy = true;
      try {
        const r = await SuggestVocabEntry(this.vocabDraftPayload(d));
        if (this.vocabDraft !== d) return;
        d.chinese = r.chinese || d.chinese;
        d.phonetic = d.phonetic || r.phonetic || '';
        d.partOfSpeech = d.partOfSpeech || r.partOfSpeech || '';
        d.example = r.example || d.example;
        if ((r.collocations || []).length) d.collocations = r.collocations.map(c => ({ pattern: c.pattern || '', meaning: c.meaning || '', example: c.example || '' }));
        if (!quiet) this.showSuccess('AI 已补全，确认后保存');
      } catch (e) {
        this.showError(`AI 补全失败: ${e}`);
      } finally {
        this.vocabAIBusy = false;
      }
    },

    vocabDraftPayload(d) {
      const { tagsRaw, definitions, ...rest } = d;
      return {
        ...rest,
        word: d.word.trim(),
        tags: this.parseTags(tagsRaw),
        collocations: d.collocations.filter(c => c.pattern.trim() || c.meaning.trim()),
      };
    },

    async saveVocabDraft() {
      const d = this.vocabDraft;
      if (!d?.word.trim()) return;
      this.vocabSaving = true;
      try {
        const saved = await SaveVocabEntry(this.vocabDraftPayload(d));
        const i = this.vocabEntries.findIndex(v => v.id === saved.id);
        if (i >= 0) this.vocabEntries.splice(i, 1, saved); else this.vocabEntries = [saved, ...this.vocabEntries];
        this.vocabEntries = [...this.vocabEntries];
        this.vocabDraft = this.vocabToDraft(saved);
        this.vocabDraftSnapshot = JSON.stringify(this.vocabDraft);
        this.showSuccess(i >= 0 ? '已保存' : '已加入单词本，可在「复习」中练习');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.vocabSaving = false;
      }
    },

    async deleteVocab(v) {
      if (!(await this.askConfirm(`从单词本删除「${v.word}」及其复习记录？`, { title: '删除单词', confirmText: '删除' }))) return;
      try {
        await DeleteVocabEntry(v.id);
        this.vocabEntries = this.vocabEntries.filter(x => x.id !== v.id);
        if (this.vocabDraft?.id === v.id) this.vocabDraft = null;
        this.showSuccess('已删除');
      } catch (e) {
        this.showError(`删除失败: ${e}`);
      }
    },

    vocabLevelTitle(v) {
      if (!v.lastReviewedAt) return '还没复习过';
      const next = v.nextReviewAt ? new Date(v.nextReviewAt * 1000) : null;
      const due = !next || next <= new Date();
      return `熟练度 ${v.reviewLevel || 0} · ` + (due ? '现在可复习' : `下次复习 ${this.memoDateLabel(this.ymd(next), true)}`);
    },

    async speakVocabWord(entry) {
      const word = (entry?.word || '').trim();
      if (!word) { this.showError('英文单词不能为空'); return; }
      try {
        await SpeakVocabWord(word);
      } catch (e) {
        this.showError(`系统发音失败: ${e}`);
      }
    },

    playVocabAudio(entry) {
      const url = entry.audioUrl || entry.audioURL;
      if (!url) {
        this.speakVocabWord(entry);
        return;
      }
      new Audio(url).play().catch(() => this.speakVocabWord(entry));
    },

    // ── Memo: flashcard review (mirrors memo.ApplyReview intervals) ───
    startReview(n) {
      const queue = [...this.reviewDueEntries]
        .sort((a, b) => (a.nextReviewAt || 0) - (b.nextReviewAt || 0) || (a.createdAt || 0) - (b.createdAt || 0))
        .slice(0, n)
        .map(v => v.id);
      this.review = { queue, index: 0, revealed: false, tally: { hard: 0, known: 0, mastered: 0 } };
      if (queue.length) this.$nextTick(() => this.playVocabAudio(this.reviewCard));
    },

    reviewIntervalLabel(v, result) {
      const level = v.reviewLevel || 0;
      const days = result === 'known' ? 3 + (level + 1) * 2 : 14 + (level + 2) * 3;
      return `${days} 天后`;
    },

    async gradeReview(result) {
      const card = this.reviewCard;
      if (!card || !this.review.revealed) return;
      try {
        const saved = await ReviewVocabEntry(card.id, result);
        const i = this.vocabEntries.findIndex(v => v.id === saved.id);
        if (i >= 0) this.vocabEntries.splice(i, 1, saved);
        this.review.tally[result]++;
        this.review.index++;
        this.review.revealed = false;
        if (this.reviewCard) this.$nextTick(() => this.playVocabAudio(this.reviewCard));
      } catch (e) {
        this.showError(`保存复习结果失败: ${e}`);
      }
    },

    reviewKey(e) {
      if (this.tab !== 'memo' || this.memoSubTab !== 'review' || !this.reviewCard || this.palette.open) return;
      if (/^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement?.tagName)) return;
      if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault();
        if (!this.review.revealed) this.review.revealed = true;
      } else if (this.review.revealed && ['1', '2', '3'].includes(e.key)) {
        e.preventDefault();
        this.gradeReview({ 1: 'hard', 2: 'known', 3: 'mastered' }[e.key]);
      }
    },

    async saveMemoSettingsForm() {
      this.memoSettingsSaving = true;
      try {
        await SaveMemoSettings(this.memoSettings.filesDir, this.memoSettings.retainDays | 0, !!this.memoSettings.autoCleanup);
        await this.loadMemoModule();
        this.showSuccess('备忘录存储设置已保存');
      } catch (e) {
        this.showError(`保存存储设置失败: ${e}`);
      } finally {
        this.memoSettingsSaving = false;
      }
    },

    async cleanupMemoFilesNow() {
      try {
        const n = await CleanupMemoFiles();
        await this.loadMemoModule();
        this.showSuccess(`已清理 ${n || 0} 个附件文件`);
      } catch (e) {
        this.showError(`清理附件失败: ${e}`);
      }
    },

    // ── Global search / replace ────────────────────────────────────────────
    setupSearchShortcuts() {
      window.addEventListener('keydown', (e) => {
        if (e.metaKey && e.key === 'f') {
          e.preventDefault();
          this.openSearchBar(false);
        } else if (e.metaKey && e.key === 'r') {
          // Only intercept if a textarea/input is focused (avoid interfering with other Cmd+R uses)
          const el = document.activeElement;
          if (el && (el.tagName === 'TEXTAREA' || el.tagName === 'INPUT')) {
            e.preventDefault();
            this.openSearchBar(true);
          }
        }
      }, true);
    },

    openSearchBar(replaceMode) {
      const el = document.activeElement;
      if (el && (el.tagName === 'TEXTAREA' || el.tagName === 'INPUT')) {
        this.searchActiveEl = el;
      }
      this.searchReplaceMode = replaceMode;
      this.searchOpen = true;
      this.$nextTick(() => {
        document.getElementById('search-bar-input')?.focus();
      });
    },

    closeSearchBar() {
      this.searchOpen = false;
      this.searchQuery = '';
      this.searchReplaceText = '';
      this.searchMatches = [];
      this.searchCurrentIdx = -1;
      if (this.searchActiveEl) {
        this.searchActiveEl.focus();
        this.searchActiveEl = null;
      }
    },

    updateSearchMatches() {
      if (!this.searchActiveEl || !this.searchQuery) {
        this.searchMatches = [];
        this.searchCurrentIdx = -1;
        return;
      }
      const text = this.searchActiveEl.value;
      const query = this.searchQuery;
      const matches = [];
      let idx = 0;
      while (idx < text.length) {
        const pos = text.indexOf(query, idx);
        if (pos === -1) break;
        matches.push({ start: pos, end: pos + query.length });
        idx = pos + 1;
      }
      this.searchMatches = matches;
      if (matches.length > 0) {
        this.searchCurrentIdx = 0;
        this._highlightMatch(0);
      } else {
        this.searchCurrentIdx = -1;
      }
    },

    _highlightMatch(idx) {
      if (!this.searchActiveEl || idx < 0 || idx >= this.searchMatches.length) return;
      const m = this.searchMatches[idx];
      this.searchActiveEl.focus();
      this.searchActiveEl.setSelectionRange(m.start, m.end);
      const ratio = m.start / Math.max(1, this.searchActiveEl.value.length);
      this.searchActiveEl.scrollTop = this.searchActiveEl.scrollHeight * ratio;
    },

    searchNext() {
      if (!this.searchMatches.length) return;
      this.searchCurrentIdx = (this.searchCurrentIdx + 1) % this.searchMatches.length;
      this._highlightMatch(this.searchCurrentIdx);
    },

    searchPrev() {
      if (!this.searchMatches.length) return;
      this.searchCurrentIdx = (this.searchCurrentIdx - 1 + this.searchMatches.length) % this.searchMatches.length;
      this._highlightMatch(this.searchCurrentIdx);
    },

    searchReplaceOne() {
      if (!this.searchActiveEl || this.searchCurrentIdx < 0 || !this.searchMatches.length) return;
      const m = this.searchMatches[this.searchCurrentIdx];
      const val = this.searchActiveEl.value;
      const newVal = val.slice(0, m.start) + this.searchReplaceText + val.slice(m.end);
      this.searchActiveEl.value = newVal;
      this.searchActiveEl.dispatchEvent(new Event('input', { bubbles: true }));
      this.updateSearchMatches();
    },

    searchReplaceAll() {
      if (!this.searchActiveEl || !this.searchQuery) return;
      const count = this.searchMatches.length;
      const newVal = this.searchActiveEl.value.split(this.searchQuery).join(this.searchReplaceText);
      this.searchActiveEl.value = newVal;
      this.searchActiveEl.dispatchEvent(new Event('input', { bubbles: true }));
      this.updateSearchMatches();
      this.showSuccess(`已替换 ${count} 处`);
    },

    get filteredSkills() {
      let list = this.skills;
      if (this.skillFilter === 'claude') list = list.filter(s => s.inClaude);
      else if (this.skillFilter === 'codex')  list = list.filter(s => s.inCodex);
      const q = (this.skillSearch || '').trim().toLowerCase();
      if (q) list = list.filter(s => s.name.toLowerCase().includes(q));
      return list;
    },

    get filteredHooks() {
      let list = this.hooks || [];
      if (this.hooksFilter === 'claude') list = list.filter(h => h.enabledInClaude);
      else if (this.hooksFilter === 'codex') list = list.filter(h => h.enabledInCodex);
      const q = (this.hooksSearch || '').trim().toLowerCase();
      if (q) {
        list = list.filter(h =>
          (h.name || '').toLowerCase().includes(q) ||
          (h.id || '').toLowerCase().includes(q) ||
          (h.event || '').toLowerCase().includes(q) ||
          (h.command || '').toLowerCase().includes(q));
      }
      return list;
    },

    // Rows Claude-global / Codex can switch (overview counts + workspace profiles).
    get mcpServers() {
      return (this.mcpInventory?.rows || []).map(r => {
        const at = key => r.entries.find(e => e.locationKey === key);
        const c = at('claude:user'), x = at('codex:user');
        return { id: r.name, name: r.name, enabledInClaude: !!c?.enabled, enabledInCodex: !!x?.enabled, global: !!(c || x) };
      }).filter(s => s.global);
    },

    get mcpEnabledCount() {
      return (this.mcpInventory?.rows || []).filter(r => r.entries.some(e => e.enabled && !e.readOnly)).length;
    },

    get mcpFailedCount() {
      return (this.mcpInventory?.rows || []).filter(r => r.entries.some(e => ['failed', 'timeout'].includes(e.probe?.status))).length;
    },

    get mcpRowCount() {
      return (this.mcpInventory?.rows || []).filter(r => r.entries.some(e => e.location.client !== 'desktop')).length;
    },

    get recentTraceErrors() {
      return (this.aiTraces || []).filter(t => t.error).slice(0, 3);
    },

    get filteredMCPRows() {
      let rows = this.mcpInventory?.rows || [];
      const f = this.mcpFilter;
      if (f === 'claude') rows = rows.filter(r => r.entries.some(e => e.location.client === 'claude'));
      else if (f === 'codex') rows = rows.filter(r => r.entries.some(e => e.location.client === 'codex'));
      else if (f === 'drift') rows = rows.filter(r => r.drift);
      else if (f === 'off') rows = rows.filter(r => r.entries.some(e => !e.readOnly && !e.enabled));
      else if (f === 'readonly') rows = rows.filter(r => r.entries.every(e => e.readOnly));
      const q = (this.mcpSearch || '').trim().toLowerCase();
      if (q) rows = rows.filter(r => r.name.toLowerCase().includes(q) || r.entries.some(e => this.mcpSummary(e.spec).toLowerCase().includes(q)));
      return rows;
    },

    hostNetActive(n) { return n.receive == null || (n.receive + n.send) > 0; },
    get hostVisibleRates() {
      return this.hostShowIdleNet ? this.hostRates : this.hostRates.filter(n => this.hostNetActive(n));
    },
    get hostIdleNetCount() {
      return this.hostRates.filter(n => !this.hostNetActive(n)).length;
    },

    async openHost() {
      this.hostPrevious = null;
      this.hostRates = [];
      this.refreshHost();
      // Benchmarks are heavy (64 MiB disk write, networkQuality download):
      // run them on the first visit only and keep the result until re-tested.
      if (!this.hostDiskResult) this.hostDiskTest();
      if (!this.hostNetworkResult) this.hostNetworkTest();
      try { this.hostHardware = await window['go']['app']['App']['GetHostHardware'](); }
      catch (error) { this.hostError = '硬件信息获取失败：' + error; }
    },
    async refreshHost() {
      if (this.hostLoading) return;
      this.hostLoading = true;
      try {
        const snapshot = await window['go']['app']['App']['GetHostSnapshot']();
        const previous = this.hostPrevious;
        const seconds = previous ? (snapshot.timestamp - previous.timestamp) / 1000 : 0;
        this.hostRates = (snapshot.networks || []).map(network => {
          const old = (previous?.networks || []).find(item => item.name === network.name);
          return { name: network.name, receive: old && seconds > 0 && network.received >= old.received ? (network.received - old.received) / seconds : null, send: old && seconds > 0 && network.sent >= old.sent ? (network.sent - old.sent) / seconds : null };
        }).sort((left, right) => ((right.receive || 0) + (right.send || 0)) - ((left.receive || 0) + (left.send || 0)));
        this.hostPrevious = snapshot;
        this.hostSnapshot = snapshot;
        this.hostError = (snapshot.errors || []).join('；');
      } catch (error) { this.hostError = '采集失败：' + error; }
      finally {
        this.hostLoading = false;
        if (this.hostTimer) clearTimeout(this.hostTimer);
        if (this.tab === 'host') this.hostTimer = setTimeout(() => this.refreshHost(), 3000);
      }
    },
    hostBytes(value) {
      if (value == null || !Number.isFinite(Number(value))) return '—';
      const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
      let amount = Number(value), unit = 0;
      while (amount >= 1024 && unit < units.length - 1) { amount /= 1024; unit++; }
      return amount.toFixed(unit ? 1 : 0) + ' ' + units[unit];
    },
    async hostDiskTest() {
      if (this.hostDiskBusy) return;
      this.hostDiskBusy = true;
      this.hostDiskResult = null;
      try { this.hostDiskResult = await window['go']['app']['App']['RunHostDiskBenchmark'](); }
      catch (error) { this.showError('磁盘测速失败：' + error); }
      finally { this.hostDiskBusy = false; }
    },
    async hostAnalyzeStorage() {
      if (this.hostStorageBusy) return;
      this.hostStorageBusy = true;
      try { this.hostStorage = await window['go']['app']['App']['AnalyzeHostStorage'](); }
      catch (error) { this.showError('磁盘分析失败：' + error); }
      finally { this.hostStorageBusy = false; }
    },
    async hostLoadPorts() {
      if (this.hostPortsBusy) return;
      this.hostPortsBusy = true;
      try { this.hostPorts = await window['go']['app']['App']['GetHostPorts'](); }
      catch (error) { this.showError('端口读取失败：' + error); }
      finally { this.hostPortsBusy = false; }
    },
    async hostNetworkTest() {
      if (this.hostNetworkBusy) return;
      this.hostNetworkBusy = true;
      this.hostNetworkResult = null;
      try { this.hostNetworkResult = await window['go']['app']['App']['RunHostNetworkBenchmark'](); }
      catch (error) { this.showError('网络测速失败：' + error); }
      finally { this.hostNetworkBusy = false; }
    },
    // ── Header ────────────────────────────────────────────────────────────
    headerRefresh() {
      this.loadModules();
      if (this.tab === 'scripts') {
        this.scanAll();
        this.loadRecentRuns();
      } else if (this.tab === 'vpn') {
        this.refreshVPNStatus();
        if (this.vpnSubTab === 'subs') this.loadVPNSubscriptions();
        else if (this.vpnSubTab === 'rules') this.loadVPNRules();
      } else if (this.tab === 'memo') {
        this.loadMemoModule();
      } else if (this.tab === 'host') {
        this.refreshHost();
      } else if (this.tab === 'ai') {
        this.loadAITraces();
        this.loadAIHeroStats();
        this.aiGo(this.aiSubTab);
      }
    },

    vpnPillTitle() {
      if (this.vpnHealthReport && this.vpnHealthReport.summary) {
        return this.vpnHealthReport.summary + ' — 点击打开 VPN 体检';
      }
      const s = this.vpnStatus;
      if (!s) return 'VPN 状态 — 点击打开配置中心';
      const sp = s.systemProxyEnabled ? '开' : '关';
      const tun = s.tunEnabled ? '开' : '关';
      const ip = (this.vpnCurrentIP && !this.vpnCurrentIP.error)
        ? (this.vpnCurrentIP.country || this.vpnCurrentIP.ip || '—')
        : '—';
      return `系统代理: ${sp} · TUN: ${tun} · 出口: ${ip} — 点击打开配置中心`;
    },

    // ── VPN ───────────────────────────────────────────────────────────────

    async refreshVPNStatus() {
      this.vpnStatusRefreshing = true;
      try {
        const [status, report] = await Promise.all([
          GetVPNStatus(),
          GetVPNHealthReport(),
        ]);
        this.vpnStatus = status;
        this.vpnHealthReport = report;
      } catch (e) {
        this.vpnMsg = '刷新失败: ' + e;
      } finally {
        this.vpnStatusRefreshing = false;
      }
      this.refreshVPNCurrentIP();
    },

    async loadVPNOverview() {
      try {
        this.vpnHealthLoading = true;
        const [profilesRes, dirRes, backupsRes, statusRes, settingsRes, reportRes] = await Promise.allSettled([
          GetVPNProfiles(),
          GetVPNAppDir(),
          ListVPNBackups(),
          GetVPNStatus(),
          GetVPNSettings(),
          GetVPNHealthReport(),
        ]);
        if (profilesRes.status === 'rejected' && reportRes.status === 'rejected') {
          throw profilesRes.reason;
        }
        const profiles = profilesRes.status === 'fulfilled' ? profilesRes.value : [];
        const settings = settingsRes.status === 'fulfilled' ? settingsRes.value : null;
        this.vpnProfiles = profiles || [];
        this.vpnCurrentProfile = this.vpnProfiles.find(p => p.isCurrent) || null;
        this.vpnAppDir = dirRes.status === 'fulfilled' ? (dirRes.value || '') : '';
        this.vpnBackups = backupsRes.status === 'fulfilled' ? (backupsRes.value || []) : [];
        this.vpnStatus = statusRes.status === 'fulfilled' ? (statusRes.value || null) : null;
        this.vpnSettings = settings || null;
        this.vpnHealthReport = reportRes.status === 'fulfilled' ? (reportRes.value || null) : null;
        // Fetch real exit IP in background (network call, takes a moment)
        this.refreshVPNCurrentIP();
        if (settings) {
          this.vpnSettingsForm = { appDir: settings.appDir, backupDir: settings.backupDir };
          this.vpnBackupDir = settings.backupDir;
        }
      } catch (e) {
        this.vpnMsg = '加载失败: ' + e;
      } finally {
        this.vpnHealthLoading = false;
      }
    },

    async saveVPNSettings() {
      this.vpnSettingsSaving = true;
      this.vpnPathValidation = null;
      try {
        const validation = await ValidateVPNSettings(this.vpnSettingsForm.appDir, this.vpnSettingsForm.backupDir);
        this.vpnPathValidation = validation;
        if (validation && validation.errors && validation.errors.length) {
          this.vpnMsg = '目录设置未保存: ' + validation.errors.join('；');
          return;
        }
        await SaveVPNSettings(this.vpnSettingsForm.appDir, this.vpnSettingsForm.backupDir);
        await this.loadVPNOverview();
        this.vpnMsg = '目录设置已保存';
        setTimeout(() => { this.vpnMsg = ''; }, 3000);
      } catch (e) {
        this.vpnMsg = '保存失败: ' + e;
      } finally {
        this.vpnSettingsSaving = false;
      }
    },

    async openVPNDir() {
      try { await OpenVPNDir(); } catch (e) { this.vpnMsg = '打开失败: ' + e; }
    },

    async openVPNBackupDir() {
      try { await OpenVPNBackupDir(); } catch (e) { this.vpnMsg = '打开失败: ' + e; }
    },

    vpnFormatBytes(bytes) {
      if (!bytes) return '—';
      const gb = bytes / 1024 / 1024 / 1024;
      return gb >= 1 ? gb.toFixed(2) + ' GB' : (bytes / 1024 / 1024).toFixed(0) + ' MB';
    },

    async refreshVPNCurrentIP() {
      this.vpnIPLoading = true;
      this.vpnCurrentIP = null;
      try {
        this.vpnCurrentIP = await GetVPNCurrentIP();
      } catch (e) {
        this.vpnCurrentIP = { error: String(e) };
      } finally {
        this.vpnIPLoading = false;
      }
    },

    async loadVPNConfig() {
      this.vpnConfigContent = '';
      this.vpnConfigError = '';
      try {
        const f = this.vpnConfigFile;
        if (f.startsWith('profile:')) {
          this.vpnConfigContent = await GetVPNProfileContent(f.slice(8)) || '';
        } else {
          const r = await GetVPNConfigFile(f);
          this.vpnConfigContent = r.content || '';
        }
      } catch (e) {
        this.vpnConfigError = String(e);
      }
    },

    async saveVPNConfig() {
      this.vpnConfigSaving = true;
      this.vpnConfigError = '';
      try {
        const f = this.vpnConfigFile;
        if (f.startsWith('profile:')) await SaveVPNProfileContent(f.slice(8), this.vpnConfigContent);
        else await SaveVPNConfigFile(f, this.vpnConfigContent);
        this.vpnMsg = '已保存 ' + f.replace(/^profile:/, '');
        setTimeout(() => { this.vpnMsg = ''; }, 3000);
      } catch (e) {
        this.vpnConfigError = String(e);
      } finally {
        this.vpnConfigSaving = false;
      }
    },

    async exportVPNBackup() {
      this.vpnBackupBusy = true;
      this.vpnMsg = '';
      try {
        const r = await ExportVPNBackup();
        this.vpnMsg = `备份成功: ${r.path} (${r.numFiles} 个文件)`;
        this.vpnBackups = (await ListVPNBackups()) || [];   // Go nil slice → null
        this.vpnHealthReport = await GetVPNHealthReport();
      } catch (e) {
        this.vpnMsg = '备份失败: ' + e;
      } finally {
        this.vpnBackupBusy = false;
      }
    },

    async previewAndImportVPNBackup(zipPath) {
      const target = (zipPath || '').trim();
      if (!target) {
        this.vpnMsg = '请输入备份 zip 路径';
        return;
      }
      this.vpnBackupPreview = null;
      try {
        const preview = await PreviewVPNBackup(target);
        this.vpnBackupPreview = preview;
        const keyFiles = preview.keyFiles && preview.keyFiles.length ? preview.keyFiles.join(', ') : '未发现关键配置文件';
        const warnings = preview.warnings && preview.warnings.length ? '\n\n提醒：\n' + preview.warnings.join('\n') : '';
        if (!preview.valid) {
          this.vpnMsg = '备份预览失败: 这不像 Clash Verge Rev 配置备份';
          return;
        }
        if (!(await this.askConfirm(`确认恢复这个备份？\n${target}\n\n包含 ${preview.numFiles} 个文件\n关键文件：${keyFiles}\n\n恢复前 Hatch 会自动创建当前配置备份。建议先关闭 Clash Verge Rev。${warnings}`, { title: '恢复 VPN 备份', confirmText: '恢复' }))) return;
      } catch (e) {
        this.vpnMsg = '备份预览失败: ' + e;
        return;
      }
      await this.importVPNBackup(target);
    },

    async importVPNBackup(zipPath) {
      this.vpnRestoreBusy = true;
      this.vpnMsg = '';
      try {
        const msg = await ImportVPNBackup(zipPath);
        this.vpnMsg = '恢复成功: ' + msg;
        this.vpnBackups = (await ListVPNBackups()) || [];   // Go nil slice → null
        await this.loadVPNOverview();
      } catch (e) {
        this.vpnMsg = '恢复失败: ' + e;
      } finally {
        this.vpnRestoreBusy = false;
      }
    },

    vpnBackupBasename(path) {
      return path.split('/').pop();
    },

    async deleteVPNBackup(zipPath) {
      if (!(await this.askConfirm(`确认删除备份？\n${this.vpnBackupBasename(zipPath)}\n\n此操作不可撤销。`))) return;
      try {
        await DeleteVPNBackup(zipPath);
        this.vpnBackups = (await ListVPNBackups()) || [];   // Go nil slice → null
        this.vpnMsg = '已删除备份';
        setTimeout(() => { this.vpnMsg = ''; }, 3000);
      } catch (e) {
        this.vpnMsg = '删除失败: ' + e;
      }
    },

    vpnProfileTypeBadge(type) {
      const map = { remote: '订阅', merge: '合并', script: '脚本', rules: '规则', proxies: '节点', groups: '分组' };
      return map[type] || type;
    },

    vpnFormatTs(ts) {
      if (!ts) return '—';
      return new Date(ts * 1000).toLocaleString('zh-CN', { dateStyle: 'short', timeStyle: 'short' });
    },

    vpnHealthClass(status) {
      return {
        ok: 'vpn-health-ok',
        warn: 'vpn-health-warn',
        error: 'vpn-health-error',
      }[status] || 'vpn-health-warn';
    },

    vpnHealthLabel(status) {
      return {
        ok: '正常',
        warn: '需检查',
        error: '需处理',
      }[status] || '未知';
    },

    vpnHealthPillText() {
      const overall = this.vpnHealthReport && this.vpnHealthReport.overall;
      if (overall) return 'VPN ' + this.vpnHealthLabel(overall);
      if (this.vpnStatus && (this.vpnStatus.systemProxyEnabled || this.vpnStatus.tunEnabled)) return 'VPN 正常';
      return 'VPN 体检';
    },

    vpnHealthGo(action) {
      if (!action || !action.tab) return;
      this.vpnGo(action.tab === 'advanced' ? 'config' : action.tab);
    },

    // ── VPN hero dashboard ──────────────────────────────────────────────────
    vpnConnState() {
      const s = this.vpnStatus;
      if (!s) return { text: '读取中…', on: false, cls: '' };
      if (!s.coreOnline) return { text: '内核未运行', on: false, cls: 'is-off' };
      if (s.tunEnabled || s.systemProxyEnabled) return { text: '已连接', on: true, cls: 'is-on' };
      return { text: '未接管流量', on: false, cls: 'is-warn' };
    },

    vpnModeLabel(mode) {
      return { rule: '规则分流', global: '全局代理', direct: '全部直连' }[mode] || mode || '';
    },

    vpnTunBypassText() {
      const n = ((this.vpnStatus && this.vpnStatus.tunExcludedCidrs) || []).length;
      return n ? `${n} 个网段绕过 VPN` : '没有绕过的网段';
    },

    vpnFlag(code) {
      if (!/^[A-Za-z]{2}$/.test(code || '')) return '🌐';
      return String.fromCodePoint(...code.toUpperCase().split('').map(c => 0x1F1E6 + c.charCodeAt(0) - 65));
    },

    vpnIPPlace(ip) {
      let country = ip.country || '';
      try {
        if (/^[A-Za-z]{2}$/.test(country)) country = new Intl.DisplayNames(['zh-CN'], { type: 'region' }).of(country.toUpperCase());
      } catch (_) { /* keep the code */ }
      return [country, ip.city].filter(Boolean).join(' · ') || '未知位置';
    },

    vpnIPOrg(org) {
      return (org || '').replace(/^AS\d+\s+/, '');
    },

    vpnDelayText(ms) {
      if (ms > 0) return ms + ' ms';
      if (ms < 0) return '超时';
      return '未测速';
    },

    vpnDelayClass(ms) {
      if (!ms) return '';
      if (ms < 0) return 'is-bad';
      return ms < 200 ? 'is-good' : ms < 500 ? 'is-mid' : 'is-bad';
    },

    async testVPNNodeDelay() {
      const s = this.vpnStatus;
      if (!s || !s.currentGroup || this.vpnDelayTesting) return;
      this.vpnDelayTesting = true;
      try {
        const ms = await window['go']['app']['App']['TestVPNNodeDelay'](s.currentGroup);
        s.nodeDelay = ms;
        if (s.proxyGroups && s.proxyGroups[0]) s.proxyGroups[0].delay = ms;
      } catch (e) {
        this.showError('测速失败：' + e);
      } finally {
        this.vpnDelayTesting = false;
      }
    },

    // ── VPN navigation ──────────────────────────────────────────────────────
    vpnGo(tab) {
      this.vpnSubTab = tab;
      if (tab === 'overview') this.loadVPNOverview();
      else if (tab === 'subs') this.loadVPNSubscriptions();
      else if (tab === 'rules') this.loadVPNRules();
      else if (tab === 'backup') this.loadVPNOverview();
      else if (tab === 'config') { if (!this.vpnSubs) this.loadVPNSubscriptions(); this.loadVPNConfig(); }
    },

    // ── 订阅 ────────────────────────────────────────────────────────────────
    async loadVPNSubscriptions() {
      this.vpnSubsLoading = true;
      this.vpnSubsError = '';
      try {
        this.vpnSubs = await window['go']['app']['App']['GetVPNSubscriptions']();
      } catch (e) {
        this.vpnSubsError = String(e);
      } finally {
        this.vpnSubsLoading = false;
      }
    },

    async copyVPNSubURL(s) {
      try {
        const url = await window['go']['app']['App']['GetVPNSubscriptionURL'](s.uid);
        await this.copyToClipboard(url, '订阅链接');
      } catch (e) {
        this.vpnMsg = '复制失败: ' + e;
      }
    },

    async openVPNApp() {
      try { await window['go']['app']['App']['OpenVPNApp'](); } catch (e) { this.vpnMsg = '打开 Clash Verge 失败: ' + e; }
    },

    async openVPNRawFile(file) {
      if (!this.vpnSubs) await this.loadVPNSubscriptions();
      this.vpnConfigFile = 'profile:' + file;
      this.vpnSubTab = 'config';
      this.loadVPNConfig();
    },

    vpnConfigRemoteFile() {
      const f = this.vpnConfigFile;
      if (!f.startsWith('profile:') || !this.vpnSubs) return null;
      return (this.vpnSubs.files || []).find(x => x.file === f.slice(8) && x.remote) || null;
    },

    vpnExpireText(expire) {
      if (!expire) return '长期有效';
      const days = Math.floor((expire * 1000 - Date.now()) / 86400000);
      const date = new Date(expire * 1000).toLocaleDateString('zh-CN');
      if (days < 0) return `已于 ${date} 到期`;
      return `${date} 到期 · 剩 ${days} 天`;
    },

    vpnExpireClass(expire) {
      if (!expire) return '';
      const days = (expire * 1000 - Date.now()) / 86400000;
      return days < 0 ? 'vpn-text-error' : days < 15 ? 'vpn-text-warn' : '';
    },

    vpnRelTime(ts) {
      if (!ts) return '—';
      const diff = Date.now() / 1000 - ts;
      if (diff < 60) return '刚刚';
      if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前';
      if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前';
      if (diff < 86400 * 30) return Math.floor(diff / 86400) + ' 天前';
      return this.vpnFormatTs(ts);
    },

    vpnIntervalText(min) {
      if (!min) return '';
      if (min === 1440) return '天';
      if (min % 1440 === 0) return ` ${min / 1440} 天`;
      if (min % 60 === 0) return ` ${min / 60} 小时`;
      return ` ${min} 分钟`;
    },

    vpnGroupTypeLabel(type) {
      return { select: '手动选择', 'url-test': '自动测速', fallback: '故障转移', 'load-balance': '负载均衡', relay: '链式代理' }[type] || type || '分组';
    },

    vpnEnhanceLabel(e) {
      return { rules: '分流规则', merge: 'Hosts / DNS 覆写', script: '脚本', proxies: '额外节点', groups: '额外策略组' }[e.kind] || e.label;
    },

    // ── 自定义规则 ──────────────────────────────────────────────────────────
    vpnRulesDirty() {
      return !!this.vpnRulesDraft && JSON.stringify(this.vpnRulesDraft) !== this.vpnRulesSnapshot;
    },

    async loadVPNRules(force = false) {
      if (this.vpnRulesDirty()) {
        if (!force) return;
        if (!(await this.askConfirm('放弃尚未保存的规则修改，重新读取？', { title: '重新读取', confirmText: '放弃修改' }))) return;
      }
      await this.fetchVPNRules();
    },

    async fetchVPNRules() {
      this.vpnRulesLoading = true;
      this.vpnRulesError = '';
      try {
        const state = await window['go']['app']['App']['GetVPNRules']();
        this.vpnRules = state;
        this.vpnRulesDraft = this.vpnDraftShape(state.draft);
        this.vpnRulesSnapshot = JSON.stringify(this.vpnRulesDraft);
        this.vpnRulesIssues = [];
        this.vpnAppendOpen = this.vpnRulesDraft.append.length > 0;
        this.vpnBulk.target = 'DIRECT';
      } catch (e) {
        this.vpnRulesError = String(e);
      } finally {
        this.vpnRulesLoading = false;
      }
    },

    async discardVPNRules() {
      if (!(await this.askConfirm('放弃所有未保存的修改？', { title: '放弃修改', confirmText: '放弃' }))) return;
      await this.fetchVPNRules();
    },

    async saveVPNRules() {
      this.vpnRulesSaving = true;
      try {
        const res = await window['go']['app']['App']['SaveVPNRules'](this.vpnRulesDraft);
        this.vpnRulesIssues = res.issues || [];
        if (this.vpnRulesIssues.length) {
          const first = this.vpnRulesIssues[0].section;
          this.vpnRulesSection = { prepend: 'rules', append: 'rules', hosts: 'hosts', fakeip: 'fakeip', dns: 'dns', tun: 'tun' }[first] || 'rules';
          if (first === 'append') this.vpnAppendOpen = true;
          this.showError(res.message || '有内容需要修改');
          return;
        }
        await this.fetchVPNRules();
        const warn = (res.warnings || []).length ? '（' + res.warnings.join('；') + '）' : '';
        this.showSuccess((res.message || '已保存') + warn);
      } catch (e) {
        this.showError('保存失败：' + e);
      } finally {
        this.vpnRulesSaving = false;
      }
    },

    vpnIssue(section, index) {
      const hit = this.vpnRulesIssues.find(x => x.section === section && x.index === index);
      return hit ? hit.message : '';
    },

    vpnRuleSections() {
      const d = this.vpnRulesDraft || {};
      const issues = (secs) => this.vpnRulesIssues.some(x => secs.includes(x.section));
      return [
        { id: 'rules', title: '分流规则', desc: '网站走代理、直连还是拦截', count: (d.prepend || []).length + (d.append || []).length, issues: issues(['prepend', 'append']) },
        { id: 'hosts', title: 'Hosts', desc: '域名固定解析到指定 IP', count: (d.hosts || []).length, issues: issues(['hosts']) },
        { id: 'fakeip', title: 'Fake-IP 过滤', desc: '内网 / 通话 / 游戏用真实 IP', count: (d.fakeIpFilter || []).length, issues: issues(['fakeip']) },
        { id: 'dns', title: 'DNS 分流', desc: '指定域名用指定 DNS', count: (d.dnsPolicy || []).length, issues: issues(['dns']) },
        { id: 'tun', title: 'TUN 绕过', desc: '这些网段完全不走 VPN', count: (d.tunExclude || []).length, issues: issues(['tun']) },
      ];
    },

    vpnRuleTypeInfo(type) {
      return ((this.vpnRules && this.vpnRules.ruleTypes) || []).find(t => t.type === type) || null;
    },

    vpnRuleTypeOptions(current) {
      const types = (this.vpnRules && this.vpnRules.ruleTypes) || [];
      const basic = types.filter(t => !t.advanced || t.type === current);
      return { basic, advanced: types.filter(t => t.advanced && t.type !== current) };
    },

    vpnRuleToString(r) {
      if (r.raw && r.raw.trim()) return r.raw.trim();
      const info = this.vpnRuleTypeInfo(r.type);
      let s = `${r.type},${(r.value || '').trim()},${(r.target || '').trim()}`;
      if (info && info.noResolve && r.noResolve) s += ',no-resolve';
      return s;
    },

    vpnSetRuleType(r, type) {
      if (type === '__raw') {
        r.raw = this.vpnRuleToString(r);
        return;
      }
      const info = this.vpnRuleTypeInfo(type);
      r.type = type;
      if (info && info.noResolve && /^IP-|GEOIP/.test(type)) r.noResolve = true;
    },

    // Raw → form editing when the rule has the simple TYPE,value,target[,no-resolve] shape.
    vpnRuleToSimple(r) {
      const parts = (r.raw || '').split(',').map(x => x.trim());
      const type = (parts[0] || '').toUpperCase();
      const info = this.vpnRuleTypeInfo(type);
      const nr = parts.length === 4 && parts[3].toLowerCase() === 'no-resolve' && info && info.noResolve;
      if (!info || !(parts.length === 3 || nr)) {
        this.showError('这条规则包含高级语法，只能原样编辑');
        return;
      }
      Object.assign(r, { type, value: parts[1], target: parts[2], noResolve: !!nr, raw: '' });
    },

    vpnTargetOptions(current, proxyOnly = false) {
      const groups = (this.vpnRules && this.vpnRules.groups) || [];
      const opts = proxyOnly ? [] : [
        { value: 'DIRECT', label: '直连（不走代理）' },
        { value: 'REJECT', label: '拦截（禁止访问）' },
      ];
      groups.forEach(g => opts.push({ value: g, label: '代理 · ' + g }));
      if (current && !opts.some(o => o.value === current)) opts.push({ value: current, label: current });
      return opts;
    },

    vpnTargetClass(target) {
      if (!target) return '';
      if (target === 'DIRECT') return 'vpn-target-direct';
      if (/^REJECT/.test(target)) return 'vpn-target-reject';
      return 'vpn-target-proxy';
    },

    vpnDefaultTarget() {
      const groups = (this.vpnRules && this.vpnRules.groups) || [];
      return groups[0] || 'DIRECT';
    },

    vpnAddRule(list) {
      list.push({ type: 'DOMAIN-SUFFIX', value: '', target: this.vpnDefaultTarget(), noResolve: false, raw: '', note: '' });
      this.$nextTick(() => {
        const rows = document.querySelectorAll('.vpn-rule-list-' + (list === this.vpnRulesDraft.append ? 'append' : 'prepend') + ' .vpn-rule-value');
        if (rows.length) rows[rows.length - 1].focus();
      });
    },

    vpnAddRow(section) {
      const d = this.vpnRulesDraft;
      if (section === 'hosts') d.hosts.push({ domain: '', value: '', note: '' });
      else if (section === 'fakeip') d.fakeIpFilter.push({ value: '', note: '' });
      else if (section === 'dns') d.dnsPolicy.push({ domain: '', server: '', note: '' });
      else if (section === 'tun') d.tunExclude.push({ value: '' });
      this.$nextTick(() => {
        const inputs = document.querySelectorAll('.vpn-kv-list-' + section + ' .vpn-kv-first');
        if (inputs.length) inputs[inputs.length - 1].focus();
      });
    },

    vpnMoveRow(list, i, delta) {
      const j = i + delta;
      if (j < 0 || j >= list.length) return;
      const [item] = list.splice(i, 1);
      list.splice(j, 0, item);
    },

    async vpnBulkAdd() {
      if (!this.vpnBulk.text.trim()) return;
      this.vpnBulk.busy = true;
      try {
        const rules = await window['go']['app']['App']['ParseVPNRuleLines'](this.vpnBulk.text, this.vpnBulk.target);
        const n = this.vpnMergeDraft({ prepend: rules || [] });
        this.vpnBulk = { open: false, text: '', target: this.vpnBulk.target, busy: false };
        this.showSuccess(`已添加 ${n} 条规则，确认无误后点「保存并生效」`);
      } catch (e) {
        this.showError('解析失败：' + e);
      } finally {
        this.vpnBulk.busy = false;
      }
    },

    // Appends template / bulk output to the draft, skipping exact duplicates.
    vpnMergeDraft(add) {
      const d = this.vpnRulesDraft;
      let n = 0;
      const ruleKeys = new Set([...d.prepend, ...d.append].map(r => this.vpnRuleToString(r)));
      (add.prepend || []).forEach(r => {
        const k = this.vpnRuleToString(r);
        if (!ruleKeys.has(k)) { ruleKeys.add(k); d.prepend.push(r); n++; }
      });
      (add.hosts || []).forEach(h => {
        const cur = d.hosts.find(x => x.domain === h.domain);
        if (cur) { if (cur.value !== h.value) { cur.value = h.value; n++; } } else { d.hosts.push(h); n++; }
      });
      (add.fakeIpFilter || []).forEach(f => {
        if (!d.fakeIpFilter.some(x => x.value === f.value)) { d.fakeIpFilter.push(f); n++; }
      });
      (add.dnsPolicy || []).forEach(p => {
        const cur = d.dnsPolicy.find(x => x.domain === p.domain);
        if (cur) { if (cur.server !== p.server) { cur.server = p.server; n++; } } else { d.dnsPolicy.push(p); n++; }
      });
      return n;
    },

    vpnRestoreFakeIP(mode) {
      const sub = (this.vpnRules && this.vpnRules.subscriptionFakeIp) || [];
      if (mode === 'reset') {
        this.vpnRulesDraft.fakeIpFilter = sub.map(v => ({ value: v, note: '' }));
        return;
      }
      const have = new Set(this.vpnRulesDraft.fakeIpFilter.map(f => f.value));
      sub.filter(v => !have.has(v)).forEach(v => this.vpnRulesDraft.fakeIpFilter.push({ value: v, note: '订阅默认' }));
    },

    vpnMissingFakeIP() {
      const sub = (this.vpnRules && this.vpnRules.subscriptionFakeIp) || [];
      const have = new Set(((this.vpnRulesDraft && this.vpnRulesDraft.fakeIpFilter) || []).map(f => f.value));
      return sub.filter(v => !have.has(v));
    },

    vpnDraftShape(d) {
      d = d || {};
      return {
        prepend: d.prepend || [], append: d.append || [], hosts: d.hosts || [],
        fakeIpFilter: d.fakeIpFilter || [], dnsPolicy: d.dnsPolicy || [], tunExclude: d.tunExclude || [],
      };
    },

    vpnSectionLabel(section) {
      return { rules: '分流', hosts: 'Hosts', fakeip: 'Fake-IP', dns: 'DNS', tun: 'TUN 绕过' }[section] || section;
    },

    vpnCIDRHint(v) {
      const m = /\/(\d+)$/.exec((v || '').trim());
      if (!m) return '';
      const bits = Number(m[1]);
      if ((v || '').includes(':')) return bits >= 128 ? '单个 IP' : '';
      return bits >= 32 ? '单个 IP' : `约 ${Math.pow(2, 32 - bits).toLocaleString()} 个地址`;
    },

    vpnAddCommonTun() {
      const have = new Set(this.vpnRulesDraft.tunExclude.map(c => c.value));
      ['10.0.0.0/8', '172.16.0.0/12', '192.168.0.0/16', '127.0.0.0/8'].forEach(v => {
        if (!have.has(v)) this.vpnRulesDraft.tunExclude.push({ value: v });
      });
    },

    // ── AI：一句话改规则 ──
    vpnAIExamples() {
      return ['让 notion.so 和 figma.com 走代理', '公司 corp.example.com 直连，用 10.0.0.53 解析', '屏蔽 doubleclick.net', '让 192.168.50.0/24 不经过 VPN'];
    },

    async runVPNRuleAI() {
      const st = this.vpnAI;
      if (!st.text.trim() || st.busy) return;
      st.busy = true;
      st.error = '';
      st.preview = null;
      try {
        st.preview = await window['go']['app']['App']['SuggestVPNRuleChange'](st.text, this.vpnRulesDraft);
      } catch (e) {
        st.error = 'AI 没能给出有效修改：' + e;
      } finally {
        st.busy = false;
      }
    },

    async applyVPNRuleAI(saveNow) {
      const p = this.vpnAI.preview;
      if (!p) return;
      this.vpnRulesDraft = this.vpnDraftShape(p.draft);
      const first = (p.changes[0] || {}).section || 'rules';
      this.vpnRulesSection = first;
      this.vpnAI = { text: '', busy: false, error: '', preview: null };
      if (saveNow) await this.saveVPNRules();
      else this.showSuccess('已放进下方列表，确认无误后点「保存并生效」');
    },

    vpnTestTargetLabel(expect, t) {
      if (expect === 'BYPASS') return '绕过 TUN';
      if (expect === 'DIRECT') return '直连';
      if (/^REJECT/.test(expect)) return '拦截';
      return expect || t.actual;
    },

    // ── AI 修改增强文件（订阅页） ──
    async openVPNEnhanceAI(row, sub) {
      this.vpnEnhAI = { row, sub, info: null, text: '', busy: false, error: '', preview: null, applying: false };
      try {
        const info = await window['go']['app']['App']['GetVPNEnhanceInfo'](row.file);
        if (this.vpnEnhAI && this.vpnEnhAI.row === row) this.vpnEnhAI.info = info;
      } catch (e) {
        if (this.vpnEnhAI) this.vpnEnhAI.error = String(e);
      }
    },

    closeVPNEnhanceAI() {
      if (this.vpnEnhAI && (this.vpnEnhAI.busy || this.vpnEnhAI.applying)) return;
      this.vpnEnhAI = null;
    },

    vpnEnhAIScope() {
      const st = this.vpnEnhAI;
      if (!st) return '';
      const scope = st.info ? st.info.scope : (st.sub ? (st.sub.isCurrent ? 'current' : 'other') : 'global');
      if (scope === 'global') return '所有订阅（全局）';
      if (scope === 'current') return '当前正在使用的订阅，应用后立即生效';
      return '未在使用的订阅「' + (st.sub ? st.sub.name : '') + '」，切换到它后生效';
    },

    async runVPNEnhanceAI() {
      const st = this.vpnEnhAI;
      if (!st || !st.text.trim() || st.busy) return;
      st.busy = true;
      st.error = '';
      st.preview = null;
      try {
        st.preview = await window['go']['app']['App']['SuggestVPNEnhanceEdit'](st.row.file, st.text);
      } catch (e) {
        st.error = 'AI 没能给出有效修改：' + e;
      } finally {
        st.busy = false;
      }
    },

    async applyVPNEnhanceAI() {
      const st = this.vpnEnhAI;
      if (!st || !st.preview) return;
      st.applying = true;
      st.error = '';
      try {
        const res = await window['go']['app']['App']['ApplyVPNEnhanceEdit'](st.row.file, st.preview.content);
        this.vpnEnhAI = null;
        this.showSuccess(res.message || '已保存');
        await this.loadVPNSubscriptions();
        if (this.vpnRules && !this.vpnRulesDirty()) this.fetchVPNRules();
        this.refreshVPNStatus();
      } catch (e) {
        st.error = '应用失败：' + e;
      } finally {
        if (this.vpnEnhAI) this.vpnEnhAI.applying = false;
      }
    },

    // ── Rule tester ──
    async runVPNRuleTest() {
      const t = this.vpnRuleTest;
      if (!t.input.trim()) return;
      t.busy = true;
      t.error = '';
      try {
        t.result = await window['go']['app']['App']['TestVPNRule'](t.input, this.vpnRulesDraft);
      } catch (e) {
        t.result = null;
        t.error = String(e);
      } finally {
        t.busy = false;
      }
    },

    vpnTestVerdict(r) {
      if (!r.matched) return '没有规则能确定结果';
      if (r.target === 'DIRECT') return '✅ 直连，不走代理';
      if (/^REJECT/.test(r.target)) return '⛔ 被拦截，无法访问';
      return '🌐 走代理：' + r.target;
    },

    vpnTestSource(r) {
      if (!r.matched) return '';
      if (r.source === 'custom') return `命中你的第 ${r.position} 条自定义规则`;
      if (r.source === 'append') return `命中第 ${r.position} 条追加规则`;
      return `命中订阅自带的第 ${r.position} 条规则`;
    },

    vpnTestNotes(r) {
      const notes = [];
      if (r.hostsValue) notes.push(`Hosts：${r.host} 固定解析到 ${r.hostsValue}`);
      if (r.fakeIpBypass) notes.push(`Fake-IP 过滤命中「${r.fakeIpBypass}」，会解析出真实 IP`);
      if (r.dnsServer) notes.push(`DNS 分流：由 ${r.dnsServer} 解析`);
      if (r.tunBypass) notes.push(`该 IP 在 TUN 排除网段 ${r.tunBypass} 内：TUN 模式下完全不经过 Clash，以上规则对它不生效`);
      if (r.uncertainCount) notes.push(`在它之前有 ${r.uncertainCount} 条规则要在真实连接时才能判断（如 ${r.uncertain[0]}），实际可能先命中它们`);
      return notes;
    },
  };
};

Alpine.start();
