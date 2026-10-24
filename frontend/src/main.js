import '@picocss/pico/css/pico.min.css';
import './style.css';
import Alpine from 'alpinejs';

import {
  GetConfig,
  ScanAll,
  GetScriptsByDir,
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
const GetMCPSettings        = ()                            => window['go']['app']['App']['GetMCPSettings']();
const SaveMCPSettings       = (mcpDir)                      => window['go']['app']['App']['SaveMCPSettings'](mcpDir);
const ListMCPServers        = ()                            => window['go']['app']['App']['ListMCPServers']();
const SaveMCPServer         = (server)                      => window['go']['app']['App']['SaveMCPServer'](server);
const DeleteMCPServer       = (id)                          => window['go']['app']['App']['DeleteMCPServer'](id);
const ImportMCPFromLocalDir = (path)                        => window['go']['app']['App']['ImportMCPFromLocalDir'](path);
const InstallMCPFromGit     = (url, name)                   => window['go']['app']['App']['InstallMCPFromGit'](url, name);
const SaveNpmMCPServer      = (pkg, name)                   => window['go']['app']['App']['SaveNpmMCPServer'](pkg, name);
const SetMCPEnabled         = (id, target, enabled)         => window['go']['app']['App']['SetMCPEnabled'](id, target, enabled);
const TestMCPConnection     = (id, target)                  => window['go']['app']['App']['TestMCPConnection'](id, target);
const GetMCPTemplates       = ()                            => window['go']['app']['App']['GetMCPTemplates']();
const UninstallMCPFromDisk  = (id)                          => window['go']['app']['App']['UninstallMCPFromDisk'](id);
const UpdateMCPFromGit      = (id)                          => window['go']['app']['App']['UpdateMCPFromGit'](id);
const SuggestMCPRecommendations = (query)                   => window['go']['app']['App']['SuggestMCPRecommendations'](query);
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
    // ── State ──────────────────────────────────────────────────────────
    tab: 'ai',
    // Top-level module visibility, from config.json "modules" (GetModuleSettings).
    modules: { workbench: true, ai: true, vpn: true, memo: true },

    selectedDir: null,
    allScripts: [],
    currentScripts: [],

    dragSrcPath: null,
    dragOverPath: null,
    sortMode: false,
    _dragScrollY: 0,
    _dragScrollRAF: null,
    _dragScrollHandler: null,

    query: '',
    page: 1,
    pageSize: 15,
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
    logTarget: null,

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

    exportsDirty: false,

    // Edit drawer
    editTarget: null,
    editForm: {},
    editSaving: false,

    // v1.3+: data dir inventory & retention
    dataInventory: [],
    retention: { traceRetainDays: 3, runsRetainDays: 14, runsKeepPerName: 50 },
    retentionSaving: false,

    // v1.2: prompt templates
    promptTemplates: { SuggestScriptMeta: '', SuggestZshrcAnnotation: '', SuggestScriptContent: '', SuggestZshrcOrganization: '', TestAIConfig: '', SuggestHookFromDescription: '', SuggestHookEdit: '', SuggestMCPRecommendations: '', SuggestVPNRuleChange: '', SuggestVPNEnhanceEdit: '' },
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
      { key: 'SuggestMCPRecommendations',   label: 'MCP 选型',         group: 'MCP' },
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
    // Codex model catalog (checked 2026-10 against the Codex models page). kind: current | legacy.
    codexModels: [
      { value: 'gpt-6.1-sol',   label: 'GPT-6.1 Sol · 复杂编码 / 代理任务', kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh', 'max', 'ultra'] },
      { value: 'gpt-6-astra',   label: 'GPT-6 Astra · 最难的端到端任务',    kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh'] },
      { value: 'gpt-6-sol',     label: 'GPT-6 Sol',                         kind: 'current', efforts: ['low', 'medium', 'high', 'xhigh', 'max', 'ultra'] },
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

    // Script content editor
    scriptEditorTarget: null,   // ScriptMeta of the script being edited
    scriptEditorContent: '',
    scriptEditorSaving: false,
    scriptEditorAIBusy: false,
    scriptEditorInstruction: '',

    // New script creation
    newScriptOpen: false,
    newScriptForm: { filename: '', content: '#!/bin/zsh\n\n' },
    newScriptSaving: false,
    newScriptAIBusy: false,

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

    // MCP management
    mcpSettings: {},
    mcpSettingsForm: { mcpDir: '' },
    mcpSettingsSaving: false,
    mcpServers: [],
    mcpTemplates: [],
    mcpLoading: false,
    mcpSearch: '',
    mcpFilter: 'all',        // 'all' | 'claude' | 'codex' | 'failed'
    mcpEditTarget: null,
    mcpForm: null,
    mcpSyncBusy: {},         // { "id:target": true }
    mcpTestBusy: {},         // { "id:target": true }
    mcpLocalPath: '',
    mcpGitURL: '',
    mcpGitName: '',
    mcpNpmPackage: '',
    mcpNpmName: '',
    mcpImportBusy: '',
    mcpManualOpen: false,
    mcpAddModalOpen: false,
    mcpAddTab: 'git',            // 'git' | 'npm' | 'local'
    mcpDeleteModal: { open: false, id: '', name: '', canCleanDisk: false, localPath: '' },
    confirmDialog: { open: false, title: '', message: '', confirmText: '', danger: true },
    confirmResolve: null,
    mcpTemplateGalleryOpen: false,
    mcpUpdateBusy: {},           // { id: true }
    mcpRecommendQuery: 'es 的 MCP，适配 Codex 和 Claude Code',
    mcpRecommendBusy: false,
    mcpRecommendation: null,
    mcpSelectedRecommendationID: '',

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

    // Memo module
    memoSubTab: 'today',      // 'today' | 'all' | 'words' | 'review' | 'storage'
    memoDashboard: { today: 0, overdue: 0, reminders: 0, reviewDue: 0, totalMemos: 0, totalVocab: 0 },
    memos: [],
    memoSearch: '',
    memoFilter: 'open',       // open | done | all
    memoFormOpen: false,
    memoEditingID: '',
    memoForm: { title: '', body: '', date: '', priority: 'normal', tagsRaw: '', reminderLocal: '', reminderMode: 'both', attachments: [] },
    memoAttachmentForm: { path: '', label: '', kind: 'file' },
    memoSaving: false,
    memoSettings: { filesDir: '', retainDays: 90, autoCleanup: true },
    memoSettingsSaving: false,
    memoPreviewCache: {},
    vocabEntries: [],
    vocabSearch: '',
    vocabSort: 'alpha',       // alpha | created
    vocabDateFilter: '',
    vocabPage: 1,
    vocabPageSize: 10,
    vocabDeleteConfirmID: '',
    vocabFormOpen: false,
    vocabEditingID: '',
    vocabForm: { word: '', chinese: '', phonetic: '', partOfSpeech: '', collocationsRaw: '', example: '', audioUrl: '', tagsRaw: '' },
    vocabLookupBusy: false,
    vocabAIBusy: false,
    vocabSaving: false,
    reviewReveal: {},

    // Workspace-settings accordion state
    wsShellOpen: false,
    wsDirsOpen: true,
    wsEnvOpen: true,

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

    get filteredScripts() {
      const src = this.currentScripts;
      if (!this.query.trim()) return src;
      const q = this.query.toLowerCase();
      return src.filter(s =>
        (s.name || '').toLowerCase().includes(q) ||
        (s.desc || '').toLowerCase().includes(q) ||
        (s.language || '').toLowerCase().includes(q) ||
        (s.tags || []).some(t => t.toLowerCase().includes(q)) ||
        (s.commands || []).some(c => c.toLowerCase().includes(q))
      );
    },

    get todayMemos() {
      const today = this.todayDate();
      return this.memos
        .filter(m => m.date === today && m.status !== 'done')
        .sort((a, b) => (a.reminderAt || 0) - (b.reminderAt || 0));
    },

    get filteredMemos() {
      const q = (this.memoSearch || '').trim().toLowerCase();
      return this.memos.filter(m => {
        if (this.memoFilter === 'open' && m.status === 'done') return false;
        if (this.memoFilter === 'done' && m.status !== 'done') return false;
        if (!q) return true;
        return [m.title, m.body, ...(m.tags || [])].some(v => (v || '').toLowerCase().includes(q));
      });
    },

    get filteredVocabEntries() {
      const q = (this.vocabSearch || '').trim().toLowerCase();
      const items = (this.vocabEntries || []).filter(v => {
        if (this.vocabDateFilter && this.vocabEntryDate(v) !== this.vocabDateFilter) return false;
        if (!q) return true;
        return [v.word, v.chinese, v.phonetic, v.partOfSpeech, v.example, ...(v.tags || [])]
          .some(x => (x || '').toLowerCase().includes(q));
      });
      return items.sort((a, b) => {
        if (this.vocabSort === 'created') return (b.createdAt || 0) - (a.createdAt || 0);
        const aw = (a.word || '').toLowerCase();
        const bw = (b.word || '').toLowerCase();
        if (aw !== bw) return aw.localeCompare(bw);
        return (b.createdAt || 0) - (a.createdAt || 0);
      });
    },

    get groupedVocabEntries() {
      const byDay = new Map();
      for (const entry of this.pagedVocabEntries) {
        const key = this.vocabEntryDate(entry) || 'unknown';
        const label = key === 'unknown' ? '未记录日期' : this.vocabDayLabel(key);
        if (!byDay.has(key)) {
          byDay.set(key, { key, label, items: [] });
        }
        byDay.get(key).items.push(entry);
      }
      return [...byDay.values()].sort((a, b) => {
        if (a.key === 'unknown') return 1;
        if (b.key === 'unknown') return -1;
        return b.key.localeCompare(a.key);
      });
    },

    get vocabTotalPages() {
      return Math.max(1, Math.ceil(this.filteredVocabEntries.length / this.vocabPageSize));
    },

    get pagedVocabEntries() {
      const page = Math.min(Math.max(this.vocabPage, 1), this.vocabTotalPages);
      const start = (page - 1) * this.vocabPageSize;
      return this.filteredVocabEntries.slice(start, start + this.vocabPageSize);
    },

    get vocabPageStart() {
      if (!this.filteredVocabEntries.length) return 0;
      return (Math.min(this.vocabPage, this.vocabTotalPages) - 1) * this.vocabPageSize + 1;
    },

    get vocabPageEnd() {
      return Math.min(this.filteredVocabEntries.length, Math.min(this.vocabPage, this.vocabTotalPages) * this.vocabPageSize);
    },

    get reviewDueEntries() {
      const now = Math.floor(Date.now() / 1000);
      return this.vocabEntries.filter(v => !v.nextReviewAt || v.nextReviewAt <= now);
    },

    get totalPages() {
      return Math.max(1, Math.ceil(this.filteredScripts.length / this.pageSize));
    },

    get pagedScripts() {
      if (this.sortMode) return this.filteredScripts;
      const start = (this.page - 1) * this.pageSize;
      return this.filteredScripts.slice(start, start + this.pageSize);
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
      await this.scanAll();
      this.loadDataPaths();
      this.loadDocLinks();
      this.loadAIProviders();
      if (this.tab === 'ai' && this.aiSubTab === 'mcp') this.loadMCP();
      if (this.tab === 'ai' && this.aiSubTab === 'skills') this.loadSkills();
      if (this.tab === 'ai' && this.aiSubTab === 'workspace') this.loadWorkspace();
      if (this.tab === 'ai' && this.aiSubTab === 'overview') this.loadOverview();
      if (this.tab === 'memo') this.loadMemoModule();
      if (this.tab === 'vpn') { this.loadVPNOverview(); if (this.vpnSubTab !== 'overview') this.vpnGo(this.vpnSubTab); }
      this.setupEvents();
      this.setupModuleReload();
      this.setupSearchShortcuts();
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

    openTab(tab) {
      if (!this.moduleOn(tab)) return;
      this.tab = tab;
      this.selectedDir = null;
      if (tab === 'ai') { this.loadAIUsageStatuses(); this.loadAITraces(); }
      else if (tab === 'vpn') this.loadVPNOverview();
      else if (tab === 'memo') this.loadMemoModule();
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
          };
        }
      } catch (_) {}
      if (this.moduleOn(this.tab)) return;
      const next = ['ai', 'scripts', 'vpn', 'memo'].find(t => this.moduleOn(t)) || 'scripts';
      if (init) {
        this.tab = next;
        this.selectedDir = null;
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
      const allowedAI = ['overview', 'connect', 'tools', 'workspace', 'skills', 'mcp', 'hooks', 'docs', 'prompts', 'data'];
      const allowedMemo = ['today', 'all', 'words', 'review', 'storage'];
      if (allowedAI.includes(ai)) {
        this.tab = 'ai';
        this.selectedDir = null;
        this.aiSubTab = ai;
      }
      if (['overview', 'subs', 'rules', 'backup', 'config'].includes(vpn)) {
        this.tab = 'vpn';
        this.selectedDir = null;
        this.vpnSubTab = vpn;
      }
      if (allowedMemo.includes(memo)) {
        this.tab = 'memo';
        this.selectedDir = null;
        this.memoSubTab = memo;
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
        await this.refreshStatuses();
      } catch (e) { this.showError(`扫描失败: ${e}`);   }
      finally { this.loading = false; }
    },

    // ── Directory navigation ───────────────────────────────────────────
    async enterDir(dir) {
      this.selectedDir = dir;
      this.query = '';
      this.page = 1;
      await this.scanDir(dir);
    },

    exitDir() {
      this.selectedDir = null;
      this.currentScripts = [];
      this.query = '';
    },

    async scanDir(dir) {
      this.loading = true;
      try {
        this.currentScripts = await GetScriptsByDir(dir.path) || [];
        this.page = 1;
        await this.refreshStatuses();
        const idx = this.allScripts.findIndex(s => s.path.startsWith(dir.path));
        if (idx === -1) this.scanAll();
      } catch (e) { this.showError(`扫描目录失败: ${e}`);   }
      finally { this.loading = false; }
    },

    // ── Drag-to-reorder ────────────────────────────────────────────────
    enterSortMode() {
      this.query = '';
      this.page = 1;
      this.sortMode = true;
      this._dragScrollHandler = (e) => { this._dragScrollY = e.clientY; };
      document.addEventListener('dragover', this._dragScrollHandler);
    },

    exitSortMode() {
      this.sortMode = false;
      this.dragSrcPath = null;
      this.dragOverPath = null;
      this._stopDragScroll();
      if (this._dragScrollHandler) {
        document.removeEventListener('dragover', this._dragScrollHandler);
        this._dragScrollHandler = null;
      }
    },

    scriptDragStart(e, path) {
      this.dragSrcPath = path;
      e.dataTransfer.effectAllowed = 'move';
      this._startDragScroll();
    },

    scriptDragOver(e, path) {
      e.preventDefault();
      e.dataTransfer.dropEffect = 'move';
      if (path !== this.dragSrcPath) this.dragOverPath = path;
    },

    scriptDragLeave(e) {
      if (!e.currentTarget.contains(e.relatedTarget)) {
        this.dragOverPath = null;
      }
    },

    async scriptDrop(e, targetPath) {
      e.preventDefault();
      this._stopDragScroll();
      const src = this.dragSrcPath;
      this.dragSrcPath = null;
      this.dragOverPath = null;
      if (!src || src === targetPath) return;

      const scripts = [...this.currentScripts];
      const si = scripts.findIndex(s => s.path === src);
      const ti = scripts.findIndex(s => s.path === targetPath);
      if (si === -1 || ti === -1) return;
      const [item] = scripts.splice(si, 1);
      scripts.splice(ti, 0, item);
      this.currentScripts = scripts;

      try {
        await window['go']['app']['App']['SaveScriptOrder'](
          this.selectedDir.path,
          scripts.map(s => s.path)
        );
      } catch (err) { this.showError(`保存排序失败: ${err}`); }
    },

    scriptDragEnd() {
      this.dragSrcPath = null;
      this.dragOverPath = null;
      this._stopDragScroll();
    },

    _startDragScroll() {
      const ZONE = 100;
      const MAX_SPEED = 20;
      const loop = () => {
        if (!this.dragSrcPath) return;
        const y = this._dragScrollY;
        const h = window.innerHeight;
        let speed = 0;
        if (y < ZONE) {
          speed = -MAX_SPEED * (1 - y / ZONE);
        } else if (y > h - ZONE) {
          speed = MAX_SPEED * (1 - (h - y) / ZONE);
        }
        if (speed !== 0) window.scrollBy(0, speed);
        this._dragScrollRAF = requestAnimationFrame(loop);
      };
      this._dragScrollRAF = requestAnimationFrame(loop);
    },

    _stopDragScroll() {
      if (this._dragScrollRAF) {
        cancelAnimationFrame(this._dragScrollRAF);
        this._dragScrollRAF = null;
      }
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
        this.$nextTick(() => {
          const el = document.getElementById('log-' + this.safeId(id));
          if (el) el.scrollTop = el.scrollHeight;
        });
      });

      EventsOn('executor:done', (ev) => {
        delete this.runningSet[ev.id];
        this.runningSet = { ...this.runningSet };
        this.scriptExitCodes = { ...this.scriptExitCodes, [ev.id]: ev.exitCode };
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

    // ── Script operations ──────────────────────────────────────────────
    async runScript(script) {
      const id = script.path;
      this.runningSet = { ...this.runningSet, [id]: true };
      this.scriptLogs[id] = [];
      delete this.scriptExitCodes[id];
      this.scriptExitCodes = { ...this.scriptExitCodes };
      this.logTarget = id;
      const args = this.splitArgs(this.scriptArgs[id] || '');
      try {
        await RunScript(id, script.path, args);
      } catch (e) {
        this.showError(`运行失败: ${e}`);  
        delete this.runningSet[id];
        this.runningSet = { ...this.runningSet };
      }
    },

    async runPreset(script, preset) {
      this.scriptArgs[script.path] = preset.args || '';
      await this.runScript(script);
    },

    async serviceAction(script, action, args = []) {
      const id = script.path;
      if (!this.scriptLogs[id]) this.scriptLogs[id] = [];
      this.scriptLogs[id].push({ line: `[hatch] ${action}…`, stream: 'stdout' });
      this.logTarget = id;
      try {
        await ServiceAction(id, action, args);
        this.scriptLogs[id].push({ line: `[hatch] ${action} OK`, stream: 'stdout' });
        await this.refreshStatuses();
      } catch (e) {
        this.scriptLogs[id].push({ line: `[hatch] ${action} error: ${e}`, stream: 'stderr' });
        this.showError(`${action} 失败: ${e}`);  
      }
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
      if (this.logTarget === path) this.logTarget = null;
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

    async saveExports() {
      this.savingExports = true;
      try {
        await SaveExports(this.cfg.exports || []);
        await this.loadConfig();
        this.exportsDirty = false;
        this.showSuccess('环境变量已写入 ~/.zshrc');
      } catch (e) { this.showError(`保存失败: ${e}`);   }
      finally { this.savingExports = false; }
    },

    addExportRow() {
      if (!this.cfg.exports) this.cfg.exports = [];
      this.cfg.exports = [...this.cfg.exports, { name: '', value: '' }];
      this.exportsDirty = true;
      this.$nextTick(() => {
        const inputs = document.querySelectorAll('.export-row .var-name');
        const last = inputs[inputs.length - 1];
        if (last) last.focus();
      });
    },

    removeExportRow(i) {
      this.cfg.exports = this.cfg.exports.filter((_, idx) => idx !== i);
      this.exportsDirty = true;
    },

    exportsPreview(extra = null) {
      const exports = [...(this.cfg.exports || [])];
      if (extra?.name) {
        const idx = exports.findIndex(e => e.name === extra.name);
        if (idx >= 0) exports[idx] = extra;
        else exports.push(extra);
      }
      const rows = exports
        .filter(e => (e.name || '').trim())
        .map(e => ({ name: String(e.name).trim(), value: String(e.value || '') }))
        .sort((a, b) => a.name.localeCompare(b.name));
      const lines = ['# >>> Hatch managed block >>>'];
      for (const e of rows) {
        lines.push(`export ${e.name}=${this.shellValue(e.value)}`);
      }
      if (rows.length > 0) {
        lines.push(`export PATH=${rows.map(e => '$' + e.name).join(':')}:$PATH`);
      }
      lines.push('# <<< Hatch managed block <<<');
      return lines.join('\n');
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
    openEdit(script) {
      this.editTarget = script;
      this.editForm = {
        name: script.name || '',
        desc: script.desc || '',
        type: script.type || 'oneshot',
        commandsRaw: (script.commands || []).join(', '),
        pidFile: script.pidFile || '',
        tagsRaw: (script.tags || []).join(', '),
        needsArgs: script.args === null || (Array.isArray(script.args) && script.args.length > 0),
        presets: this.normalizedPresets(script),
        envPairs: Object.entries(script.env || {}).map(([k, v]) => ({ key: k, value: v })),
        notifyOnDone: !!script.notifyOnDone,
        schedule: script.schedule || '',
        requiresRaw: (script.requires || []).join(', '),
        healthUrl: script.healthUrl || '',
        preScriptsRaw: (script.preScripts || []).join('\n'),
      };
    },

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
      const path = this.historyTarget.path;
      this.scriptArgs[path] = (r.args || []).join(' ');
      this.closeHistory();
      // Trigger the normal run path so logs and UI update.
      const script = (this.scripts || []).find(s => s.path === path) || { path };
      this.runScript(script);
    },

    addEnvPair() {
      if (!this.editForm.envPairs) this.editForm.envPairs = [];
      this.editForm.envPairs = [...this.editForm.envPairs, { key: '', value: '' }];
    },
    removeEnvPair(i) {
      this.editForm.envPairs = this.editForm.envPairs.filter((_, idx) => idx !== i);
    },

    cancelEdit() {
      this.editTarget = null;
      this.editForm = {};
    },

    async saveEdit() {
      if (!this.editTarget) return;
      this.editSaving = true;
      try {
        const meta = {
          ...this.editTarget,
          name: this.editForm.name,
          desc: this.editForm.desc,
          type: this.editForm.type,
          commands: this.editForm.type === 'service'
            ? this.editForm.commandsRaw.split(',').map(s => s.trim()).filter(Boolean)
            : [],
          pidFile: this.editForm.type === 'service' ? this.editForm.pidFile : '',
          tags: this.editForm.tagsRaw.split(',').map(s => s.trim()).filter(Boolean),
          args: this.editForm.type === 'oneshot'
            ? (this.editForm.needsArgs ? null : [])
            : null,
          presets: this.normalizedEditPresets(),
          env: (this.editForm.envPairs || []).reduce((acc, p) => {
            const k = (p.key || '').trim();
            if (k) acc[k] = p.value || '';
            return acc;
          }, {}),
          notifyOnDone: !!this.editForm.notifyOnDone,
          schedule: (this.editForm.schedule || '').trim(),
          requires: (this.editForm.requiresRaw || '').split(',').map(s => s.trim()).filter(Boolean),
          healthUrl: (this.editForm.healthUrl || '').trim(),
          preScripts: (this.editForm.preScriptsRaw || '').split('\n').map(s => s.trim()).filter(Boolean),
        };
        await SaveScriptMeta(meta);
        if (this.selectedDir) await this.scanDir(this.selectedDir);
        else await this.scanAll();
        this.cancelEdit();
        this.showSuccess('脚本元信息已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);  
      } finally {
        this.editSaving = false;
      }
    },

    async resetEdit() {
      if (!this.editTarget) return;
      if (!(await this.askConfirm(`重置「${this.baseName(this.editTarget.path)}」的手动配置，恢复自动检测？此操作不可撤销。`, { title: '重置脚本配置', confirmText: '重置' }))) return;
      this.editSaving = true;
      try {
        await DeleteScriptMeta(this.editTarget.path);
        if (this.selectedDir) await this.scanDir(this.selectedDir);
        else await this.scanAll();
        this.cancelEdit();
        this.showSuccess('已重置为自动检测');
      } catch (e) {
        this.showError(`重置失败: ${e}`);  
      } finally {
        this.editSaving = false;
      }
    },

    addPreset() {
      if (!this.editForm.presets) this.editForm.presets = [];
      this.editForm.presets = [...this.editForm.presets, { name: '', args: '' }];
      this.$nextTick(() => {
        const drawer = document.querySelector('.edit-drawer .drawer-body');
        const latest = document.querySelector('.preset-edit-row:last-child input');
        if (latest) {
          latest.scrollIntoView({ block: 'center', behavior: 'smooth' });
          latest.focus();
        } else if (drawer) {
          drawer.scrollTop = drawer.scrollHeight;
        }
      });
    },

    removePreset(i) {
      this.editForm.presets = this.editForm.presets.filter((_, idx) => idx !== i);
    },

    async fillScriptWithAI(script) {
      if (!this.aiReady) {
        if (this.moduleOn('ai')) {
          this.selectedDir = null;
          this.tab = 'ai';
        }
        this.showError('请先完成并启用 AI 配置');  
        return;
      }
      this.aiBusyScript = script.path;
      try {
        const suggestion = await SuggestScriptMeta(script);
        this.openEdit(script);
        this.editForm.name = suggestion.name || this.editForm.name;
        this.editForm.desc = suggestion.desc || this.editForm.desc;
        this.editForm.type = suggestion.type || this.editForm.type;
        this.editForm.commandsRaw = (suggestion.commands || []).join(', ');
        this.editForm.pidFile = suggestion.pidFile || this.editForm.pidFile;
        this.editForm.tagsRaw = (suggestion.tags || []).join(', ');
        this.editForm.needsArgs = !!suggestion.needsArgs;
        this.editForm.presets = this.normalizePresetList(suggestion.presets || []);
      } catch (e) {
        this.showError(`AI 分析脚本失败: ${e}`);  
      } finally {
        this.aiBusyScript = null;
      }
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

    // ── MCP management ────────────────────────────────────────────────
    async loadMCP() {
      this.mcpLoading = true;
      try {
        const [settings, servers, templates] = await Promise.all([
          GetMCPSettings(),
          ListMCPServers(),
          GetMCPTemplates(),
        ]);
        this.mcpSettings = settings || {};
        this.mcpSettingsForm = { mcpDir: this.mcpSettings.mcpDir || '' };
        this.mcpServers = servers || [];
        this.mcpTemplates = templates || [];
      } catch (e) {
        this.showError(`加载 MCP 失败: ${e}`);
      } finally {
        this.mcpLoading = false;
      }
    },

    async suggestMCPRecommendations() {
      const query = (this.mcpRecommendQuery || '').trim();
      if (!query) { this.showError('请先描述想要的 MCP，例如 Elasticsearch / ES'); return; }
      this.mcpRecommendBusy = true;
      try {
        const result = await SuggestMCPRecommendations(query);
        this.mcpRecommendation = result || null;
        const first = (result?.candidates || [])[0];
        this.mcpSelectedRecommendationID = first?.id || '';
        if (first) this.showSuccess(`已生成 MCP 推荐：${first.name || first.id}`);
      } catch (e) {
        this.showError(`生成 MCP 推荐失败: ${e}`);
      } finally {
        this.mcpRecommendBusy = false;
      }
    },

    selectedMCPRecommendation() {
      const list = this.mcpRecommendation?.candidates || [];
      return list.find(c => c.id === this.mcpSelectedRecommendationID) || list[0] || null;
    },

    applyMCPRecommendation(candidate) {
      const c = candidate || this.selectedMCPRecommendation();
      if (!c?.server) { this.showError('没有可应用的 MCP 推荐'); return; }
      const copy = JSON.parse(JSON.stringify(c.server));
      copy.enabledInClaude = false;
      copy.enabledInCodex = false;
      copy.lastTestResults = {};
      this.openMCPServerEditor(copy);
      if (c.quickEnv && this.mcpForm?.transport === 'stdio') {
        const existing = {};
        for (const row of this.mcpForm.envRows || []) {
          const key = (row.key || '').trim();
          if (key) existing[key] = row.value || '';
        }
        this.mcpForm.envRows = Object.entries({ ...c.quickEnv, ...existing })
          .map(([key, value]) => ({ key, value }));
      }
      this.showSuccess(`已套用「${c.name || c.id}」，请补齐参数后保存`);
    },

    async saveMCPSettings() {
      this.mcpSettingsSaving = true;
      try {
        await SaveMCPSettings((this.mcpSettingsForm.mcpDir || '').trim());
        await this.loadMCP();
        this.showSuccess('MCP 管理目录已保存');
      } catch (e) {
        this.showError(`保存 MCP 目录失败: ${e}`);
      } finally {
        this.mcpSettingsSaving = false;
      }
    },

    openMCPServerEditor(server) {
      this.mcpEditTarget = server?.id || 'new';
      this.mcpForm = this.mcpFormFromServer(server || {
        id: '',
        name: '',
        sourceType: 'local',
        transport: 'stdio',
        env: {},
        httpHeaders: {},
      });
    },

    openMCPTemplate(template) {
      const copy = JSON.parse(JSON.stringify(template || {}));
      copy.enabledInClaude = false;
      copy.enabledInCodex = false;
      copy.lastTestResults = {};
      this.openMCPServerEditor(copy);
    },

    closeMCPEditor() {
      this.mcpEditTarget = null;
      this.mcpForm = null;
    },

    openMCPManual() {
      this.mcpManualOpen = true;
    },

    closeMCPManual() {
      this.mcpManualOpen = false;
    },

    mcpFormFromServer(server) {
      const env = server.env || {};
      const headers = server.httpHeaders || {};
      return {
        ...server,
        argsRaw: (server.args || []).join('\n'),
        tagsRaw: (server.tags || []).join(', '),
        envRows: Object.keys(env).sort().map(k => ({ key: k, value: env[k] })),
        headerRows: Object.keys(headers).sort().map(k => ({ key: k, value: headers[k] })),
      };
    },

    mcpServerFromForm() {
      const f = this.mcpForm || {};
      const env = {};
      for (const row of f.envRows || []) {
        const key = (row.key || '').trim();
        if (key) env[key] = row.value || '';
      }
      const httpHeaders = {};
      for (const row of f.headerRows || []) {
        const key = (row.key || '').trim();
        if (key) httpHeaders[key] = row.value || '';
      }
      return {
        id: (f.id || '').trim(),
        name: (f.name || '').trim(),
        description: (f.description || '').trim(),
        sourceType: f.sourceType || 'local',
        sourceUrl: (f.sourceUrl || '').trim(),
        localPath: (f.localPath || '').trim(),
        packageName: (f.packageName || '').trim(),
        transport: f.transport || 'stdio',
        command: (f.command || '').trim(),
        args: (f.argsRaw || '').split('\n').map(s => s.trim()).filter(Boolean),
        cwd: (f.cwd || '').trim(),
        url: (f.url || '').trim(),
        env,
        httpHeaders,
        bearerTokenEnvVar: (f.bearerTokenEnvVar || '').trim(),
        startupTimeoutSec: Number(f.startupTimeoutSec || 0),
        toolTimeoutSec: Number(f.toolTimeoutSec || 0),
        tags: (f.tagsRaw || '').split(',').map(s => s.trim()).filter(Boolean),
      };
    },

    addMCPEnvRow() {
      if (!this.mcpForm) return;
      this.mcpForm.envRows = [...(this.mcpForm.envRows || []), { key: '', value: '' }];
    },

    removeMCPEnvRow(idx) {
      if (!this.mcpForm) return;
      this.mcpForm.envRows = (this.mcpForm.envRows || []).filter((_, i) => i !== idx);
    },

    addMCPHeaderRow() {
      if (!this.mcpForm) return;
      this.mcpForm.headerRows = [...(this.mcpForm.headerRows || []), { key: '', value: '' }];
    },

    removeMCPHeaderRow(idx) {
      if (!this.mcpForm) return;
      this.mcpForm.headerRows = (this.mcpForm.headerRows || []).filter((_, i) => i !== idx);
    },

    async saveMCPForm() {
      if (!this.mcpForm) return;
      try {
        const server = this.mcpServerFromForm();
        const existing = this.mcpServers.find(s => s.id === server.id);
        await SaveMCPServer(server);
        const resynced = [];
        if (existing?.enabledInClaude) {
          await SetMCPEnabled(server.id, 'claude', true);
          resynced.push('Claude Code');
        }
        if (existing?.enabledInCodex) {
          await SetMCPEnabled(server.id, 'codex', true);
          resynced.push('Codex');
        }
        await this.loadMCP();
        this.closeMCPEditor();
        const syncNote = resynced.length ? `，已同步到 ${resynced.join(' / ')}` : '';
        this.showSuccess(`MCP「${server.name || server.id}」已保存${syncNote}`);
      } catch (e) {
        this.showError(`保存 MCP 失败: ${e}`);
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

    openMCPDeleteModal(s) {
      const canCleanDisk = (s.sourceType === 'git' || s.sourceType === 'local') && !!s.localPath;
      this.mcpDeleteModal = { open: true, id: s.id, name: s.name || s.id, canCleanDisk, localPath: s.localPath || '' };
    },

    closeMCPDeleteModal() {
      this.mcpDeleteModal = { open: false, id: '', name: '', canCleanDisk: false, localPath: '' };
    },

    async confirmDeleteMCP(cleanupDisk) {
      const { id } = this.mcpDeleteModal;
      this.closeMCPDeleteModal();
      try {
        if (cleanupDisk) {
          await UninstallMCPFromDisk(id);
          this.showSuccess(`MCP「${id}」已从 Hatch 删除并清理本地文件`);
        } else {
          await DeleteMCPServer(id);
          this.showSuccess(`MCP「${id}」已从 Hatch 删除`);
        }
        if (this.mcpEditTarget === id) this.closeMCPEditor();
        await this.loadMCP();
      } catch (e) {
        this.showError(`删除 MCP 失败: ${e}`);
      }
    },

    openMCPAddModal(tab) {
      this.mcpAddTab = tab || 'git';
      this.mcpAddModalOpen = true;
    },

    closeMCPAddModal() {
      this.mcpAddModalOpen = false;
    },

    async updateMCPFromGit(id) {
      this.mcpUpdateBusy = { ...this.mcpUpdateBusy, [id]: true };
      try {
        await UpdateMCPFromGit(id);
        await this.loadMCP();
        this.showSuccess(`MCP「${id}」已从 Git 更新`);
      } catch (e) {
        this.showError(`更新 MCP 失败: ${e}`);
      } finally {
        const b = { ...this.mcpUpdateBusy };
        delete b[id];
        this.mcpUpdateBusy = b;
      }
    },

    openMCPTemplateGallery() {
      this.mcpTemplateGalleryOpen = true;
    },

    closeMCPTemplateGallery() {
      this.mcpTemplateGalleryOpen = false;
    },

    async deleteMCP(id) {
      try {
        await DeleteMCPServer(id);
        if (this.mcpEditTarget === id) this.closeMCPEditor();
        await this.loadMCP();
        this.showSuccess(`MCP「${id}」已从 Hatch 删除`);
      } catch (e) {
        this.showError(`删除 MCP 失败: ${e}`);
      }
    },

    async importMCPLocal() {
      const path = (this.mcpLocalPath || '').trim();
      if (!path) { this.showError('请先填写本地 MCP 目录'); return; }
      this.mcpImportBusy = 'local';
      try {
        const server = await ImportMCPFromLocalDir(path);
        this.mcpLocalPath = '';
        this.closeMCPAddModal();
        await this.loadMCP();
        this.showSuccess(`已导入 MCP「${server.name || server.id}」`);
      } catch (e) {
        this.showError(`导入本地 MCP 失败: ${e}`);
      } finally {
        this.mcpImportBusy = '';
      }
    },

    async installMCPGit() {
      const url = (this.mcpGitURL || '').trim();
      if (!url) { this.showError('请先填写 Git URL'); return; }
      this.mcpImportBusy = 'git';
      try {
        const server = await InstallMCPFromGit(url, (this.mcpGitName || '').trim());
        this.mcpGitURL = '';
        this.mcpGitName = '';
        this.closeMCPAddModal();
        await this.loadMCP();
        this.showSuccess(`已安装 MCP「${server.name || server.id}」`);
      } catch (e) {
        this.showError(`Git 安装 MCP 失败: ${e}`);
      } finally {
        this.mcpImportBusy = '';
      }
    },

    async addNpmMCP() {
      const pkg = (this.mcpNpmPackage || '').trim();
      if (!pkg) { this.showError('请先填写 npm 包名'); return; }
      this.mcpImportBusy = 'npm';
      try {
        const server = await SaveNpmMCPServer(pkg, (this.mcpNpmName || '').trim());
        this.mcpNpmPackage = '';
        this.mcpNpmName = '';
        this.closeMCPAddModal();
        await this.loadMCP();
        this.showSuccess(`已添加 npm MCP「${server.name || server.id}」`);
      } catch (e) {
        this.showError(`添加 npm MCP 失败: ${e}`);
      } finally {
        this.mcpImportBusy = '';
      }
    },

    async toggleMCPEnabled(id, target, enabled) {
      const key = id + ':' + target;
      this.mcpSyncBusy = { ...this.mcpSyncBusy, [key]: true };
      try {
        await SetMCPEnabled(id, target, enabled);
        const idx = this.mcpServers.findIndex(s => s.id === id);
        if (idx >= 0) {
          const s = { ...this.mcpServers[idx] };
          if (target === 'claude') s.enabledInClaude = enabled;
          else s.enabledInCodex = enabled;
          this.mcpServers = [...this.mcpServers.slice(0, idx), s, ...this.mcpServers.slice(idx + 1)];
        }
        this.showSuccess(enabled ? `MCP「${id}」已写入 ${this.mcpTargetLabel(target)} 全局配置` : `MCP「${id}」已从 ${this.mcpTargetLabel(target)} 全局配置移除`);
      } catch (e) {
        await this.loadMCP();
        this.showError(`同步 MCP 失败: ${e}`);
      } finally {
        const b = { ...this.mcpSyncBusy };
        delete b[key];
        this.mcpSyncBusy = b;
      }
    },

    async testMCP(id, target) {
      const key = id + ':' + target;
      this.mcpTestBusy = { ...this.mcpTestBusy, [key]: true };
      try {
        const result = await TestMCPConnection(id, target);
        const idx = this.mcpServers.findIndex(s => s.id === id);
        if (idx >= 0) {
          const s = { ...this.mcpServers[idx] };
          s.lastTestResults = { ...(s.lastTestResults || {}), [target]: result };
          this.mcpServers = [...this.mcpServers.slice(0, idx), s, ...this.mcpServers.slice(idx + 1)];
        }
        if (result.status === 'success') this.showSuccess(result.summary || 'MCP 连接成功');
        else this.showError(result.summary || 'MCP 连接测试失败');
      } catch (e) {
        this.showError(`MCP 测试失败: ${e}`);
      } finally {
        const b = { ...this.mcpTestBusy };
        delete b[key];
        this.mcpTestBusy = b;
      }
    },

    mcpTargetLabel(target) {
      return target === 'claude' ? 'Claude Code' : 'Codex';
    },

    mcpStatusClass(result) {
      if (!result) return 'mcp-test-none';
      if (result.status === 'success') return 'mcp-test-ok';
      if (result.status === 'failed') return 'mcp-test-fail';
      return 'mcp-test-unknown';
    },

    mcpConfigLines(server) {
      if (!server) return [];
      const lines = [];
      if (server.transport === 'http') {
        if (server.url) lines.push({ label: 'URL', value: server.url });
        if (server.bearerTokenEnvVar) lines.push({ label: 'Bearer Env', value: server.bearerTokenEnvVar });
        for (const [key, value] of Object.entries(server.httpHeaders || {}).sort()) {
          lines.push({ label: `Header ${key}`, value });
        }
      } else {
        if (server.command) lines.push({ label: 'Command', value: server.command });
        if ((server.args || []).length) lines.push({ label: 'Args', value: (server.args || []).join(' ') });
        if (server.cwd) lines.push({ label: 'CWD', value: server.cwd });
        for (const [key, value] of Object.entries(server.env || {}).sort()) {
          lines.push({ label: `Env ${key}`, value });
        }
      }
      if (server.startupTimeoutSec) lines.push({ label: '启动超时', value: `${server.startupTimeoutSec}s` });
      if (server.toolTimeoutSec) lines.push({ label: '工具超时', value: `${server.toolTimeoutSec}s` });
      return lines;
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
          this.mcpServers.length ? Promise.resolve(this.mcpServers) : ListMCPServers(),
          this.hooks.length ? Promise.resolve(this.hooks) : ListHooks(),
        ]);
        this.workspaceProfiles = profiles || [];
        this.workspaceSnapshots = snapshots || [];
        this.workspaceHealth = health;
        this.skills = skills || [];
        this.mcpServers = mcps || [];
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
    },

    async applyWorkspaceProfile() {
      const profile = this.workspacePreview?.profile;
      if (!profile?.id || !this.workspacePreview?.canApply) return;
      this.workspaceApplying = true;
      try {
        await ApplyAIWorkspaceProfile(profile.id);
        this.workspacePreview = null;
        await this.loadWorkspace();
        this.showSuccess(`已应用工作区画像「${profile.name}」`);
      } catch (e) { this.showError(`应用失败: ${e}`); }
      finally { this.workspaceApplying = false; }
    },

    async deleteWorkspaceProfile(profile) {
      if (!(await this.askConfirm(`删除工作区画像「${profile.name}」？不会改动当前 Claude/Codex 配置。`))) return;
      try { await DeleteAIWorkspaceProfile(profile.id); await this.loadWorkspace(); this.showSuccess('工作区画像已删除'); }
      catch (e) { this.showError(`删除失败: ${e}`); }
    },

    async rollbackWorkspaceSnapshot(snapshot) {
      if (!(await this.askConfirm(`回滚到「${snapshot.profileName}」应用前的状态？当前 Hatch 管理的 AI 配置会被恢复。`, { title: '回滚 AI 工作区', confirmText: '回滚' }))) return;
      try { await RollbackAIWorkspaceSnapshot(snapshot.id); await this.loadWorkspace(); this.showSuccess('已回滚 AI 工作区快照'); }
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

    // ── Script content editor ──────────────────────────────────────────
    async openScriptEditor(script) {
      this.scriptEditorTarget = script;
      this.scriptEditorInstruction = '';
      try {
        this.scriptEditorContent = await ReadScriptFile(script.path);
      } catch (e) {
        this.showError(`读取文件失败: ${e}`);
        this.scriptEditorTarget = null;
      }
    },

    closeScriptEditor() {
      this.scriptEditorTarget = null;
      this.scriptEditorContent = '';
      this.scriptEditorInstruction = '';
    },

    async saveScriptContent() {
      if (!this.scriptEditorTarget) return;
      this.scriptEditorSaving = true;
      try {
        await WriteScriptFile(this.scriptEditorTarget.path, this.scriptEditorContent);
        if (this.selectedDir) await this.scanDir(this.selectedDir);
        this.closeScriptEditor();
        this.showSuccess('脚本已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.scriptEditorSaving = false;
      }
    },

    async scriptAIGenerate() {
      if (!this.aiReady) {
        if (this.moduleOn('ai')) this.tab = 'ai';
        this.showError('请先完成并启用 AI 配置');
        return;
      }
      if (!this.scriptEditorInstruction.trim()) {
        this.showError('请输入 AI 指令（如：加错误处理、重写为 python3）');
        return;
      }
      this.scriptEditorAIBusy = true;
      try {
        const result = await SuggestScriptContent(
          this.scriptEditorTarget?.path || '',
          this.scriptEditorContent,
          this.scriptEditorInstruction
        );
        this.scriptEditorContent = result;
      } catch (e) {
        this.showError(`AI 生成失败: ${e}`);
      } finally {
        this.scriptEditorAIBusy = false;
      }
    },

    // ── New script creation ────────────────────────────────────────────
    openNewScript() {
      this.newScriptForm = { filename: '', content: '#!/bin/zsh\n\n' };
      this.newScriptOpen = true;
    },

    closeNewScript() {
      this.newScriptOpen = false;
      this.newScriptForm = { filename: '', content: '#!/bin/zsh\n\n' };
    },

    async saveNewScript() {
      if (!this.selectedDir || !this.newScriptForm.filename.trim()) return;
      this.newScriptSaving = true;
      try {
        const fname = this.newScriptForm.filename.trim();
        await CreateScriptFile(
          this.selectedDir.path,
          fname,
          this.newScriptForm.content
        );
        await this.scanDir(this.selectedDir);
        this.closeNewScript();
        this.showSuccess(`脚本「${fname}」已创建`);
      } catch (e) {
        this.showError(`创建失败: ${e}`);
      } finally {
        this.newScriptSaving = false;
      }
    },

    async newScriptAIGenerate() {
      if (!this.aiReady) {
        if (this.moduleOn('ai')) this.tab = 'ai';
        this.showError('请先完成并启用 AI 配置');
        return;
      }
      if (!this.scriptEditorInstruction.trim()) {
        this.showError('请输入 AI 指令');
        return;
      }
      this.newScriptAIBusy = true;
      try {
        const result = await SuggestScriptContent(
          this.newScriptForm.filename || 'new-script',
          this.newScriptForm.content,
          this.scriptEditorInstruction
        );
        this.newScriptForm.content = result;
      } catch (e) {
        this.showError(`AI 生成失败: ${e}`);
      } finally {
        this.newScriptAIBusy = false;
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

    // Render a copy-paste config snippet for one MCP, mirroring the backend
    // renderClaudeMCPServer / renderCodexMCPServer output.
    renderMCPSnippet(server, target) {
      if (!server) return '';
      const id = server.id || 'server';
      if (target === 'claude') {
        const body = {};
        if (server.transport === 'http') {
          body.type = 'http';
          body.url = server.url || '';
          if (server.httpHeaders && Object.keys(server.httpHeaders).length) body.headers = server.httpHeaders;
        } else {
          body.type = 'stdio';
          body.command = server.command || '';
          body.args = server.args || [];
          body.env = server.env || {};
          if (server.cwd) body.cwd = server.cwd;
        }
        return JSON.stringify({ [id]: body }, null, 2);
      }
      const q = (s) => JSON.stringify(s ?? '');
      const lines = [`[mcp_servers.${id}]`];
      if (server.transport === 'http') {
        lines.push(`url = ${q(server.url || '')}`);
        if (server.bearerTokenEnvVar) lines.push(`bearer_token_env_var = ${q(server.bearerTokenEnvVar)}`);
        const hk = Object.keys(server.httpHeaders || {}).sort();
        if (hk.length) lines.push(`http_headers = { ${hk.map(k => `${q(k)} = ${q(server.httpHeaders[k])}`).join(', ')} }`);
      } else {
        lines.push(`command = ${q(server.command || '')}`);
        if ((server.args || []).length) lines.push(`args = [${server.args.map(q).join(', ')}]`);
        if (server.cwd) lines.push(`cwd = ${q(server.cwd)}`);
      }
      if (server.startupTimeoutSec) lines.push(`startup_timeout_sec = ${server.startupTimeoutSec}`);
      if (server.toolTimeoutSec) lines.push(`tool_timeout_sec = ${server.toolTimeoutSec}`);
      lines.push('enabled = true');
      const ek = server.transport !== 'http' ? Object.keys(server.env || {}).sort() : [];
      if (ek.length) {
        lines.push('', `[mcp_servers.${id}.env]`);
        ek.forEach(k => lines.push(`${k} = ${q(server.env[k])}`));
      }
      return lines.join('\n');
    },

    async loadOverview() {
      try { await this.loadAIUsageStatuses(); } catch (_) {}
      try { if (!this.aiProviders.length) await this.loadAIProviders(); } catch (_) {}
      try { this.skillsStats = await GetSkillsStats(); } catch (_) {}
      try { if (!this.mcpServers.length) this.mcpServers = (await ListMCPServers()) || []; } catch (_) {}
      try { this.hooksStats = await GetHooksStats(); } catch (_) {}
      try { if (!this.docLinks.length) await this.loadDocLinks(); } catch (_) {}
      try { if (!this.aiTraces.length) await this.loadAITraces(); } catch (_) {}
    },

    duplicateProvider(p) {
      this.openProviderEditor(null);
      this.providerForm = { id: '', name: `${p.name || p.model || '提供商'} 副本`, baseUrl: p.baseUrl || '', model: p.model || '', apiKey: '', _isNew: true };
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

    async bulkMCPEnable(target, enabled) {
      const list = this.filteredMCPServers || [];
      if (!list.length) { this.showError('没有可操作的 MCP'); return; }
      let ok = 0, fail = 0;
      for (const s of list) {
        const already = target === 'claude' ? s.enabledInClaude : s.enabledInCodex;
        if (already === enabled) continue;
        try { await SetMCPEnabled(s.id, target, enabled); ok++; } catch (_) { fail++; }
      }
      await this.loadMCP();
      const label = this.mcpTargetLabel(target);
      let msg = enabled ? `已启用 ${ok} 个到 ${label}` : `已从 ${label} 移除 ${ok} 个`;
      if (fail) msg += `，${fail} 个跳过（冲突或失败）`;
      this.showSuccess(msg);
    },

    async bulkTestMCP() {
      const list = this.filteredMCPServers || [];
      let tested = 0, failed = 0;
      for (const s of list) {
        for (const target of ['claude', 'codex']) {
          const enabled = target === 'claude' ? s.enabledInClaude : s.enabledInCodex;
          if (!enabled) continue;
          try {
            const r = await TestMCPConnection(s.id, target);
            tested++;
            if (r && r.status === 'failed') failed++;
          } catch (_) { tested++; failed++; }
        }
      }
      await this.loadMCP();
      if (!tested) { this.showError('没有已启用的 MCP 可测试'); return; }
      this.showSuccess(`已测试 ${tested} 项${failed ? `，${failed} 项失败` : '，全部通过'}`);
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

    baseName(path) {
      if (!path) return '';
      return path.split('/').filter(Boolean).pop() || path;
    },

    splitArgs(raw) {
      const text = String(raw || '').trim();
      return text ? text.split(/\s+/) : [];
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

    normalizedEditPresets() {
      return this.normalizePresetList(this.editForm?.presets || []);
    },

    dirTypeSummary(dirPath) {
      const scripts = this.allScripts.filter(s =>
        s.path.startsWith(dirPath + '/') || s.path.startsWith(dirPath + '\\')
      );
      const svcCount     = scripts.filter(s => s.type === 'service').length;
      const oneshotCount = scripts.filter(s => s.type === 'oneshot').length;
      const parts = [];
      if (svcCount > 0)     parts.push(`${svcCount} service`);
      if (oneshotCount > 0) parts.push(`${oneshotCount} oneshot`);
      return parts.join(' · ');
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
        this.clampVocabPage();
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

    vocabEntryDate(entry) {
      const unix = entry?.createdAt || entry?.updatedAt || 0;
      if (!unix) return '';
      const d = new Date(unix * 1000);
      if (Number.isNaN(d.getTime())) return '';
      const mm = String(d.getMonth() + 1).padStart(2, '0');
      const dd = String(d.getDate()).padStart(2, '0');
      return `${d.getFullYear()}-${mm}-${dd}`;
    },

    vocabDayLabel(day) {
      if (!day) return '全部日期';
      const [year, month, date] = day.split('-');
      if (!year || !month || !date) return day;
      return `${Number(month)} 月 ${Number(date)} 日`;
    },

    resetVocabPage() {
      this.vocabPage = 1;
      this.vocabDeleteConfirmID = '';
    },

    clampVocabPage() {
      this.vocabPage = Math.min(Math.max(this.vocabPage || 1, 1), this.vocabTotalPages);
    },

    setVocabSearch(value) {
      this.vocabSearch = value;
      this.resetVocabPage();
    },

    setVocabSort(value) {
      this.vocabSort = value || 'alpha';
      this.resetVocabPage();
    },

    setVocabDateFilter(value) {
      this.vocabDateFilter = value || '';
      this.resetVocabPage();
    },

    gotoVocabPage(page) {
      this.vocabPage = Math.min(Math.max(page, 1), this.vocabTotalPages);
      this.vocabDeleteConfirmID = '';
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

    openNewMemo() {
      this.memoEditingID = '';
      this.memoFormOpen = true;
      this.memoForm = {
        title: '',
        body: '',
        date: this.todayDate(),
        priority: 'normal',
        tagsRaw: '',
        reminderLocal: '',
        reminderMode: 'both',
        attachments: [],
      };
      this.memoAttachmentForm = { path: '', label: '', kind: 'file' };
    },

    editMemo(item) {
      this.memoEditingID = item.id;
      this.memoFormOpen = true;
      this.memoForm = {
        title: item.title || '',
        body: item.body || '',
        date: item.date || this.todayDate(),
        priority: item.priority || 'normal',
        tagsRaw: this.tagsRaw(item.tags),
        reminderLocal: this.unixToLocalInput(item.reminderAt),
        reminderMode: item.reminderMode || 'both',
        attachments: [...(item.attachments || [])],
      };
      this.memoAttachmentForm = { path: '', label: '', kind: 'file' };
    },

    memoPayloadFromForm() {
      const existing = this.memos.find(m => m.id === this.memoEditingID) || {};
      return {
        ...existing,
        id: this.memoEditingID || existing.id || '',
        title: this.memoForm.title,
        body: this.memoForm.body,
        date: this.memoForm.date || this.todayDate(),
        priority: this.memoForm.priority || 'normal',
        tags: this.parseTags(this.memoForm.tagsRaw),
        reminderAt: this.localInputToUnix(this.memoForm.reminderLocal),
        reminderMode: this.memoForm.reminderMode || 'both',
        attachments: this.memoForm.attachments || [],
        status: existing.status || 'todo',
      };
    },

    async saveMemoForm() {
      if (!this.memoForm.title.trim()) { this.showError('请填写待办标题'); return; }
      this.memoSaving = true;
      try {
        const saved = await SaveMemo(this.memoPayloadFromForm());
        const idx = this.memos.findIndex(m => m.id === saved.id);
        if (idx >= 0) this.memos.splice(idx, 1, saved);
        else this.memos = [saved, ...this.memos];
        this.memoFormOpen = false;
        this.memoEditingID = '';
        await this.loadMemoModule();
        this.showSuccess('备忘录已保存');
      } catch (e) {
        this.showError(`保存备忘录失败: ${e}`);
      } finally {
        this.memoSaving = false;
      }
    },

    async toggleMemoDone(item) {
      try {
        await MarkMemoDone(item.id, item.status !== 'done');
        await this.loadMemoModule();
      } catch (e) {
        this.showError(`更新状态失败: ${e}`);
      }
    },

    async deleteMemoItem(item) {
      if (!(await this.askConfirm(`删除「${item.title}」及其托管附件？此操作不可撤销。`))) return;
      try {
        await DeleteMemo(item.id);
        await this.loadMemoModule();
        this.showSuccess('备忘录已删除');
      } catch (e) {
        this.showError(`删除备忘录失败: ${e}`);
      }
    },

    async addMemoAttachment() {
      const path = (this.memoAttachmentForm.path || '').trim();
      if (!path) return;
      try {
        const att = await ImportMemoAttachment(path, this.memoAttachmentForm.label || '', this.memoAttachmentForm.kind || 'file');
        this.memoForm.attachments = [...(this.memoForm.attachments || []), att];
        this.memoAttachmentForm = { path: '', label: '', kind: 'file' };
        if (att.kind === 'image') this.loadMemoImagePreviews();
      } catch (e) {
        this.showError(`添加附件失败: ${e}`);
      }
    },

    async removeMemoAttachment(att) {
      if (this.memoEditingID) {
        if (!(await this.askConfirm(`删除附件「${att.label || att.url || '附件'}」？托管目录中的副本也会被删除。`))) return;
        try {
          const saved = await DeleteMemoAttachment(this.memoEditingID, att.id);
          this.memoForm.attachments = saved.attachments || [];
          await this.loadMemoModule();
          return;
        } catch (e) {
          this.showError(`删除附件失败: ${e}`);
          return;
        }
      }
      this.memoForm.attachments = (this.memoForm.attachments || []).filter(a => a.id !== att.id);
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

    reminderLabel(item) {
      if (!item.reminderAt) return '';
      const when = this.formatTime(item.reminderAt);
      if (item.status !== 'done' && item.reminderAt < Math.floor(Date.now() / 1000)) return `已过期 · ${when}`;
      return when;
    },

    openNewVocab() {
      this.vocabEditingID = '';
      this.vocabDeleteConfirmID = '';
      this.vocabFormOpen = true;
      this.vocabForm = { word: '', chinese: '', phonetic: '', partOfSpeech: '', collocationsRaw: '', example: '', audioUrl: '', tagsRaw: '' };
    },

    editVocab(entry) {
      this.vocabEditingID = entry.id;
      this.vocabDeleteConfirmID = '';
      this.vocabFormOpen = true;
      this.vocabForm = {
        word: entry.word || '',
        chinese: entry.chinese || '',
        phonetic: entry.phonetic || '',
        partOfSpeech: entry.partOfSpeech || '',
        collocationsRaw: this.collocationsToRaw(entry.collocations || []),
        example: entry.example || '',
        audioUrl: entry.audioUrl || '',
        tagsRaw: this.tagsRaw(entry.tags),
      };
    },

    collocationsToRaw(list) {
      return (list || []).map(c => [c.pattern, c.meaning, c.example].filter(Boolean).join(' | ')).join('\n');
    },

    parseCollocations(raw) {
      return (raw || '').split('\n').map(line => {
        const parts = line.split('|').map(s => s.trim());
        return { pattern: parts[0] || '', meaning: parts[1] || '', example: parts[2] || '' };
      }).filter(c => c.pattern || c.meaning || c.example);
    },

    vocabPayloadFromForm() {
      const existing = this.vocabEntries.find(v => v.id === this.vocabEditingID) || {};
      return {
        ...existing,
        id: this.vocabEditingID || existing.id || '',
        word: this.vocabForm.word,
        chinese: this.vocabForm.chinese,
        phonetic: this.vocabForm.phonetic,
        partOfSpeech: this.vocabForm.partOfSpeech,
        collocations: this.parseCollocations(this.vocabForm.collocationsRaw),
        example: this.vocabForm.example,
        audioUrl: this.vocabForm.audioUrl,
        tags: this.parseTags(this.vocabForm.tagsRaw),
      };
    },

    async lookupVocab() {
      const word = (this.vocabForm.word || '').trim();
      if (!word) { this.showError('请先填写英文单词'); return; }
      this.vocabLookupBusy = true;
      try {
        const r = await LookupVocabWord(word);
        this.vocabForm.word = r.word || word;
        this.vocabForm.phonetic = r.phonetic || this.vocabForm.phonetic;
        this.vocabForm.partOfSpeech = r.partOfSpeech || this.vocabForm.partOfSpeech;
        this.vocabForm.audioUrl = r.audioUrl || this.vocabForm.audioUrl;
        if (!this.vocabForm.collocationsRaw && r.collocations?.length) this.vocabForm.collocationsRaw = this.collocationsToRaw(r.collocations);
        this.showSuccess('词典信息已填入草稿');
      } catch (e) {
        this.showError(`词典查询失败: ${e}`);
      } finally {
        this.vocabLookupBusy = false;
      }
    },

    async suggestVocab() {
      if (!this.vocabForm.word.trim()) { this.showError('请先填写英文单词'); return; }
      this.vocabAIBusy = true;
      try {
        const r = await SuggestVocabEntry(this.vocabPayloadFromForm());
        this.vocabForm = {
          ...this.vocabForm,
          word: r.word || this.vocabForm.word,
          chinese: r.chinese || this.vocabForm.chinese,
          phonetic: r.phonetic || this.vocabForm.phonetic,
          partOfSpeech: r.partOfSpeech || this.vocabForm.partOfSpeech,
          collocationsRaw: this.collocationsToRaw(r.collocations || this.parseCollocations(this.vocabForm.collocationsRaw)),
          example: r.example || this.vocabForm.example,
          audioUrl: r.audioUrl || this.vocabForm.audioUrl,
        };
        this.showSuccess('AI 已补全单词草稿');
      } catch (e) {
        this.showError(`AI 补全失败: ${e}`);
      } finally {
        this.vocabAIBusy = false;
      }
    },

    async saveVocabForm() {
      if (!this.vocabForm.word.trim()) { this.showError('请填写英文单词'); return; }
      this.vocabSaving = true;
      try {
        await SaveVocabEntry(this.vocabPayloadFromForm());
        this.vocabFormOpen = false;
        this.vocabEditingID = '';
        await this.loadMemoModule();
        this.showSuccess('单词已保存');
      } catch (e) {
        this.showError(`保存单词失败: ${e}`);
      } finally {
        this.vocabSaving = false;
      }
    },

    requestDeleteVocab(entry) {
      this.vocabDeleteConfirmID = entry?.id || '';
    },

    cancelDeleteVocab() {
      this.vocabDeleteConfirmID = '';
    },

    async confirmDeleteVocab(entry) {
      if (!entry?.id || this.vocabDeleteConfirmID !== entry.id) return;
      try {
        await DeleteVocabEntry(entry.id);
        this.vocabDeleteConfirmID = '';
        await this.loadMemoModule();
        this.showSuccess('单词已删除');
      } catch (e) {
        this.showError(`删除单词失败: ${e}`);
      }
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

    async reviewVocab(entry, result) {
      try {
        await ReviewVocabEntry(entry.id, result);
        this.reviewReveal = { ...this.reviewReveal, [entry.id]: false };
        await this.loadMemoModule();
      } catch (e) {
        this.showError(`保存复盘结果失败: ${e}`);
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

    get mcpEnabledCount() {
      return (this.mcpServers || []).filter(s => s.enabledInClaude || s.enabledInCodex).length;
    },

    get mcpFailedCount() {
      return (this.mcpServers || []).filter(s => Object.values(s.lastTestResults || {}).some(r => r.status === 'failed')).length;
    },

    get recentTraceErrors() {
      return (this.aiTraces || []).filter(t => t.error).slice(0, 3);
    },

    get filteredMCPServers() {
      let list = this.mcpServers || [];
      if (this.mcpFilter === 'claude') list = list.filter(s => s.enabledInClaude);
      else if (this.mcpFilter === 'codex') list = list.filter(s => s.enabledInCodex);
      else if (this.mcpFilter === 'failed') {
        list = list.filter(s => Object.values(s.lastTestResults || {}).some(r => r.status === 'failed'));
      }
      const q = (this.mcpSearch || '').trim().toLowerCase();
      if (q) {
        list = list.filter(s =>
          (s.id || '').toLowerCase().includes(q) ||
          (s.name || '').toLowerCase().includes(q) ||
          (s.description || '').toLowerCase().includes(q)
        );
      }
      return list;
    },

    // ── Header ────────────────────────────────────────────────────────────
    headerRefresh() {
      this.loadModules();
      if (this.tab === 'scripts') {
        this.selectedDir ? this.scanDir(this.selectedDir) : this.scanAll();
      } else if (this.tab === 'vpn') {
        this.refreshVPNStatus();
        if (this.vpnSubTab === 'subs') this.loadVPNSubscriptions();
        else if (this.vpnSubTab === 'rules') this.loadVPNRules();
      } else if (this.tab === 'memo') {
        this.loadMemoModule();
      } else if (this.tab === 'ai') {
        this.loadAIUsageStatuses();
        this.loadAITraces();
        if (this.aiSubTab === 'overview') this.loadOverview();
        else if (this.aiSubTab === 'mcp') this.loadMCP();
        else if (this.aiSubTab === 'hooks') this.loadHooks();
        else if (this.aiSubTab === 'skills') this.loadSkills();
        else if (this.aiSubTab === 'workspace') this.loadWorkspace();
        else if (this.aiSubTab === 'docs') this.loadDocLinks();
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
        this.vpnBackups = await ListVPNBackups();
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
        this.vpnBackups = await ListVPNBackups();
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
        this.vpnBackups = await ListVPNBackups();
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
