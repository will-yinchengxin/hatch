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
} from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

// New methods added to Go but not yet in wailsjs/App.js (regenerated on next build).
const SaveAIConfig        = (ai)      => window['go']['main']['App']['SaveAIConfig'](ai);
const TestAIConfig        = (ai)      => window['go']['main']['App']['TestAIConfig'](ai);
const SuggestScriptMeta   = (meta)    => window['go']['main']['App']['SuggestScriptMeta'](meta);
const GetAutoLaunchEnabled = () => window['go']['main']['App']['GetAutoLaunchEnabled']();
const SetAutoLaunchEnabled = (enabled) => window['go']['main']['App']['SetAutoLaunchEnabled'](enabled);
const GetClaudeCodeStatus   = ()                            => window['go']['main']['App']['GetClaudeCodeStatus']();
const GetCodexUsageStatus   = ()                            => window['go']['main']['App']['GetCodexUsageStatus']();
const ReadScriptFile        = (path)                        => window['go']['main']['App']['ReadScriptFile'](path);
const WriteScriptFile       = (path, content)               => window['go']['main']['App']['WriteScriptFile'](path, content);
const CreateScriptFile      = (dirPath, filename, content)  => window['go']['main']['App']['CreateScriptFile'](dirPath, filename, content);
const SuggestScriptContent  = (path, cur, inst)             => window['go']['main']['App']['SuggestScriptContent'](path, cur, inst);
const GetAITraces           = ()                            => window['go']['main']['App']['GetAITraces']();
const ClearAITraces         = ()                            => window['go']['main']['App']['ClearAITraces']();
const GetAITraceStorageInfo = ()                            => window['go']['main']['App']['GetAITraceStorageInfo']();
const GetSkills             = ()                            => window['go']['main']['App']['GetSkills']();
const GetSkillsStats        = ()                            => window['go']['main']['App']['GetSkillsStats']();
const GetSkillContent       = (name)                        => window['go']['main']['App']['GetSkillContent'](name);
const SaveSkillContent      = (name, content)               => window['go']['main']['App']['SaveSkillContent'](name, content);
const DeleteSkill           = (name)                        => window['go']['main']['App']['DeleteSkill'](name);
const DeleteSkillEverywhere = (name)                        => window['go']['main']['App']['DeleteSkillEverywhere'](name);
const SetSkillSync          = (name, target, enabled)       => window['go']['main']['App']['SetSkillSync'](name, target, enabled);
const ImportSkillsFrom      = (source)                      => window['go']['main']['App']['ImportSkillsFrom'](source);
const ResolveSkillImportConflict = (name, source, useSource) => window['go']['main']['App']['ResolveSkillImportConflict'](name, source, useSource);
const CreateSkillTemplate   = (name)                        => window['go']['main']['App']['CreateSkillTemplate'](name);
const GetAppDataPaths       = ()                            => window['go']['main']['App']['GetAppDataPaths']();
const SaveAppDataPaths      = (traceDir, skillsDir)         => window['go']['main']['App']['SaveAppDataPaths'](traceDir, skillsDir);
const GetAIProviders        = ()                            => window['go']['main']['App']['GetAIProviders']();
const SaveAIProviders       = (providers, activeID)         => window['go']['main']['App']['SaveAIProviders'](providers, activeID);
const GetDocLinks           = ()                            => window['go']['main']['App']['GetDocLinks']();
const SaveDocLinks          = (links)                       => window['go']['main']['App']['SaveDocLinks'](links);
const OpenFile              = (path)                        => window['go']['main']['App']['OpenFile'](path);
const GetVPNAppDir          = ()                            => window['go']['main']['App']['GetVPNAppDir']();
const GetVPNProfiles        = ()                            => window['go']['main']['App']['GetVPNProfiles']();
const GetVPNConfigFile      = (name)                        => window['go']['main']['App']['GetVPNConfigFile'](name);
const SaveVPNConfigFile     = (name, content)               => window['go']['main']['App']['SaveVPNConfigFile'](name, content);
const GetVPNProfileContent  = (uid)                         => window['go']['main']['App']['GetVPNProfileContent'](uid);
const SaveVPNProfileContent = (uid, content)                => window['go']['main']['App']['SaveVPNProfileContent'](uid, content);
const ExportVPNBackup       = ()                            => window['go']['main']['App']['ExportVPNBackup']();
const ImportVPNBackup       = (zipPath)                     => window['go']['main']['App']['ImportVPNBackup'](zipPath);
const ListVPNBackups        = ()                            => window['go']['main']['App']['ListVPNBackups']();
const OpenVPNDir            = ()                            => window['go']['main']['App']['OpenVPNDir']();
const OpenVPNBackupDir      = ()                            => window['go']['main']['App']['OpenVPNBackupDir']();
const GetVPNSettings        = ()                            => window['go']['main']['App']['GetVPNSettings']();
const SaveVPNSettings       = (appDir, backupDir)           => window['go']['main']['App']['SaveVPNSettings'](appDir, backupDir);
const GetVPNStatus          = ()                            => window['go']['main']['App']['GetVPNStatus']();
const GetVPNCurrentIP       = ()                            => window['go']['main']['App']['GetVPNCurrentIP']();
const GetClaudeSettings     = ()                            => window['go']['main']['App']['GetClaudeSettings']();
const SaveClaudeSettings    = (content)                     => window['go']['main']['App']['SaveClaudeSettings'](content);
const GetCodexConfig        = ()                            => window['go']['main']['App']['GetCodexConfig']();
const SaveCodexConfig       = (content)                     => window['go']['main']['App']['SaveCodexConfig'](content);

window.hatchApp = function () {
  return {
    // ── State ──────────────────────────────────────────────────────────
    tab: 'ai',

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
    promptTemplates: { SuggestScriptMeta: '', SuggestZshrcAnnotation: '', SuggestScriptContent: '', SuggestZshrcOrganization: '', TestAIConfig: '' },
    promptDefaults: {},
    promptDefaultsOpen: {},
    promptSaving: false,
    promptOps: [
      { key: 'SuggestScriptMeta',         label: '脚本元信息分析' },
      { key: 'SuggestScriptContent',      label: '脚本内容生成' },
      { key: 'SuggestZshrcAnnotation',    label: 'zshrc 条目备注' },
      { key: 'SuggestZshrcOrganization',  label: 'zshrc 整理重排' },
      { key: 'TestAIConfig',              label: '配置连通性测试' },
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

    // Claude Code model + effort config
    claudeCodeConfig: { model: '', effortLevel: '' },
    claudeCodeConfigSaving: false,
    claudeCodePresets: [],
    claudeCodePresetForm: { id: '', name: '', model: '', effortLevel: '' },
    claudeCodePresetOpen: false,  // inline add/edit form visible

    // Codex model + reasoning effort + goal-mode config
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

    // AI sub-navigation
    aiSubTab: 'overview',    // 'overview' | 'connect' | 'skills' | 'tools'

    // VPN module
    vpnSubTab: 'overview',   // 'overview' | 'config' | 'profiles' | 'backup' | 'enhance'
    vpnProfiles: [],
    vpnCurrentProfile: null,
    vpnAppDir: '',
    vpnStatus: null,
    vpnCurrentIP: null,
    vpnIPLoading: false,
    vpnSettings: null,
    vpnSettingsForm: { appDir: '', backupDir: '' },
    vpnSettingsSaving: false,
    vpnConfigFile: 'verge.yaml',
    vpnConfigContent: '',
    vpnConfigSaving: false,
    vpnConfigError: '',
    vpnProfileEditing: null,
    vpnProfileContent: '',
    vpnProfileSaving: false,
    vpnProfileError: '',
    vpnBackups: [],
    vpnBackupDir: '',
    vpnBackupBusy: false,
    vpnRestoreBusy: false,
    vpnMsg: '',
    vpnTUNExpanded: false,
    vpnStatusRefreshing: false,
    vpnEnhanceData: null,
    vpnEnhanceLoading: false,

    // Workspace-settings accordion state
    wsShellOpen: false,
    wsDirsOpen: true,
    wsEnvOpen: true,

    // App data paths config
    dataPaths: {},           // AppDataPaths from Go
    dataPathsForm: { traceLogDir: '', skillsDir: '' },
    dataPathsSaving: false,

    // Doc links (AI > Tools quick-open documents)
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
      await this.loadConfig();
      await this.scanAll();
      this.loadDataPaths();
      this.loadDocLinks();
      this.loadAIProviders();
      this.setupEvents();
      this.setupSearchShortcuts();
      setInterval(() => this.refreshStatuses(), 3000);
      try { this.vpnStatus = await GetVPNStatus(); } catch (_) {}
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
        await window['go']['main']['App']['SaveScriptOrder'](
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
      EventsOff('executor:log', 'executor:done', 'executor:status', 'ai:trace');
      EventsOn('ai:trace', (trace) => {
        this.aiTraces = [trace, ...this.aiTraces].slice(0, 50);
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
        const getInv = window?.go?.main?.App?.GetAppDataInventory;
        const getRet = window?.go?.main?.App?.GetRetentionPolicy;
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
        const fn = window?.go?.main?.App?.SaveRetentionPolicy;
        if (typeof fn !== 'function') throw new Error('binding 未生成');
        await fn(this.retention.traceRetainDays|0, this.retention.runsRetainDays|0, this.retention.runsKeepPerName|0);
        this.showSuccess('保留策略已保存并立即应用。');
        await this.loadDataInventory();
      } catch (e) { this.showError(`保存失败: ${e}`); }
      finally { this.retentionSaving = false; }
    },
    async cleanupDataDir(d) {
      if (!d.cleanable) return;
      if (!confirm(`确认清空「${d.label}」(${d.fileCount} 个文件, ${this.formatBytes(d.totalBytes)})？此操作不可撤销。`)) return;
      try {
        const fn = window?.go?.main?.App?.CleanupDataDir;
        if (typeof fn !== 'function') throw new Error('binding 未生成');
        const msg = await fn(d.key);
        this.showSuccess(msg || '已清理');
        await this.loadDataInventory();
      } catch (e) { this.showError(`清理失败: ${e}`); }
    },
    async openDataDir(path) {
      try {
        const fn = window?.go?.main?.App?.OpenFile;
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
        const getOverrides = window?.go?.main?.App?.GetPromptTemplates;
        const getDefaults = window?.go?.main?.App?.GetDefaultPrompts;
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
        const fn = window?.go?.main?.App?.SavePromptTemplates;
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
      if (!confirm('确认清空所有自定义 Prompt 模板，恢复内置默认？')) return;
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
        const fn = window?.go?.main?.App?.ListRuns;
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
        const fn = window?.go?.main?.App?.GetRunLog;
        if (typeof fn === 'function') {
          this.historyLogText = await fn(this.historyTarget.path, r.id);
        }
      } catch (e) {
        this.historyLogText = `读取失败: ${e}`;
      }
    },
    async deleteHistory(r) {
      try {
        const fn = window?.go?.main?.App?.DeleteRun;
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
        this.selectedDir = null;
        this.tab = 'ai';
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
        const fn = window?.go?.main?.App?.SuggestSkillContent;
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
        claude:   { name: 'Claude',    baseUrl: 'https://api.anthropic.com/v1',               model: 'claude-sonnet-4-6' },
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
        this.claudeCodeConfig = await window['go']['main']['App']['GetClaudeCodeConfig']();
      } catch (e) { /* silent */ }
    },

    async saveClaudeCodeConfig() {
      this.claudeCodeConfigSaving = true;
      try {
        await window['go']['main']['App']['SetClaudeCodeConfig'](
          this.claudeCodeConfig.model || '',
          this.claudeCodeConfig.effortLevel || ''
        );
        this.showSuccess('Claude Code 配置已保存');
      } catch (e) {
        this.showError(`保存失败: ${e}`);
      } finally {
        this.claudeCodeConfigSaving = false;
      }
    },

    async loadClaudeCodePresets() {
      try {
        this.claudeCodePresets = await window['go']['main']['App']['GetClaudeCodePresets']() || [];
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
        await window['go']['main']['App']['SaveClaudeCodePreset'](this.claudeCodePresetForm);
        this.claudeCodePresetOpen = false;
        await this.loadClaudeCodePresets();
        this.showSuccess('预设已保存');
      } catch (e) {
        this.showError(`保存预设失败: ${e}`);
      }
    },

    async deleteClaudeCodePreset(id) {
      try {
        await window['go']['main']['App']['DeleteClaudeCodePreset'](id);
        await this.loadClaudeCodePresets();
        this.showSuccess('预设已删除');
      } catch (e) {
        this.showError(`删除预设失败: ${e}`);
      }
    },

    async applyClaudeCodePreset(preset) {
      this.claudeCodeConfig = { model: preset.model, effortLevel: preset.effortLevel };
      await this.saveClaudeCodeConfig();
    },

    async loadCodexModelConfig() {
      try {
        this.codexModelConfig = await window['go']['main']['App']['GetCodexModelConfig']();
      } catch (e) { /* silent */ }
    },

    async saveCodexModelConfig() {
      this.codexModelConfigSaving = true;
      try {
        await window['go']['main']['App']['SetCodexModelConfig'](
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
        this.codexPresets = await window['go']['main']['App']['GetCodexPresets']() || [];
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
        await window['go']['main']['App']['SaveCodexPreset'](this.codexPresetForm);
        this.codexPresetOpen = false;
        await this.loadCodexPresets();
        this.showSuccess('预设已保存');
      } catch (e) {
        this.showError(`保存预设失败: ${e}`);
      }
    },

    async deleteCodexPreset(id) {
      try {
        await window['go']['main']['App']['DeleteCodexPreset'](id);
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

    effortLabel(level) {
      return { low: 'low', medium: 'medium', high: 'high', xhigh: 'xhigh', max: 'max' }[level] || '系统默认';
    },

    // xhigh: Opus + Sonnet 支持（官方文档：extended thinking 要求）
    effortSupportsXhigh(model) {
      if (!model) return true; // 不知道模型时不限制
      return model.includes('opus') || model.includes('sonnet');
    },

    // max: 仅 Opus 支持
    effortSupportsMax(model) {
      if (!model) return true; // 不知道模型时不限制
      return model.includes('opus');
    },

    modelLabel(model) {
      const map = {
        'claude-opus-4-7': 'Opus 4.7',
        'claude-opus-4-6': 'Opus 4.6',
        'claude-sonnet-4-6': 'Sonnet 4.6',
        'claude-haiku-4-5': 'Haiku 4.5',
        'claude-opus-4-5': 'Opus 4.5',
      };
      return map[model] || model || '默认';
    },

    codexModelLabel(model) {
      const map = {
        'gpt-5.5': 'GPT-5.5',
        'gpt-5.4': 'GPT-5.4',
        'gpt-5.4-mini': 'GPT-5.4 mini',
        'gpt-5.3-codex': 'GPT-5.3 Codex',
        'gpt-5.2': 'GPT-5.2',
      };
      return map[model] || model || '默认';
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
        this.tab = 'ai';
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
        this.tab = 'ai';
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

    // ── VPN ───────────────────────────────────────────────────────────────

    async refreshVPNStatus() {
      this.vpnStatusRefreshing = true;
      try {
        this.vpnStatus = await GetVPNStatus();
      } catch (e) {
        this.vpnMsg = '刷新失败: ' + e;
      } finally {
        this.vpnStatusRefreshing = false;
      }
      this.refreshVPNCurrentIP();
    },

    async loadVPNOverview() {
      try {
        const [profiles, dir, backups, status, settings] = await Promise.all([
          GetVPNProfiles(),
          GetVPNAppDir(),
          ListVPNBackups(),
          GetVPNStatus(),
          GetVPNSettings(),
        ]);
        this.vpnProfiles = profiles || [];
        this.vpnCurrentProfile = this.vpnProfiles.find(p => p.isCurrent) || null;
        this.vpnAppDir = dir || '';
        this.vpnBackups = backups || [];
        this.vpnStatus = status || null;
        this.vpnSettings = settings || null;
        // Fetch real exit IP in background (network call, takes a moment)
        this.refreshVPNCurrentIP();
        if (settings) {
          this.vpnSettingsForm = { appDir: settings.appDir, backupDir: settings.backupDir };
          this.vpnBackupDir = settings.backupDir;
        }
      } catch (e) {
        this.vpnMsg = '加载失败: ' + e;
      }
    },

    async saveVPNSettings() {
      this.vpnSettingsSaving = true;
      try {
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
        const r = await GetVPNConfigFile(this.vpnConfigFile);
        this.vpnConfigContent = r.content || '';
      } catch (e) {
        this.vpnConfigError = String(e);
      }
    },

    async saveVPNConfig() {
      this.vpnConfigSaving = true;
      this.vpnConfigError = '';
      try {
        await SaveVPNConfigFile(this.vpnConfigFile, this.vpnConfigContent);
        this.vpnMsg = '已保存 ' + this.vpnConfigFile;
        setTimeout(() => { this.vpnMsg = ''; }, 3000);
      } catch (e) {
        this.vpnConfigError = String(e);
      } finally {
        this.vpnConfigSaving = false;
      }
    },

    async openVPNProfile(profile) {
      this.vpnProfileEditing = profile;
      this.vpnProfileContent = '';
      this.vpnProfileError = '';
      try {
        this.vpnProfileContent = await GetVPNProfileContent(profile.file);
      } catch (e) {
        this.vpnProfileError = String(e);
      }
    },

    async saveVPNProfile() {
      if (!this.vpnProfileEditing) return;
      this.vpnProfileSaving = true;
      this.vpnProfileError = '';
      try {
        await SaveVPNProfileContent(this.vpnProfileEditing.file, this.vpnProfileContent);
        this.vpnMsg = '已保存 ' + this.vpnProfileEditing.file;
        setTimeout(() => { this.vpnMsg = ''; }, 3000);
      } catch (e) {
        this.vpnProfileError = String(e);
      } finally {
        this.vpnProfileSaving = false;
      }
    },

    closeVPNProfile() {
      this.vpnProfileEditing = null;
      this.vpnProfileContent = '';
      this.vpnProfileError = '';
    },

    async exportVPNBackup() {
      this.vpnBackupBusy = true;
      this.vpnMsg = '';
      try {
        const r = await ExportVPNBackup();
        this.vpnMsg = `备份成功: ${r.path} (${r.numFiles} 个文件)`;
        this.vpnBackups = await ListVPNBackups();
      } catch (e) {
        this.vpnMsg = '备份失败: ' + e;
      } finally {
        this.vpnBackupBusy = false;
      }
    },

    async importVPNBackup(zipPath) {
      if (!confirm(`确认从以下备份恢复？\n${zipPath}\n\n建议先关闭 Clash Verge Rev。`)) return;
      this.vpnRestoreBusy = true;
      this.vpnMsg = '';
      try {
        const msg = await ImportVPNBackup(zipPath);
        this.vpnMsg = '恢复成功: ' + msg;
      } catch (e) {
        this.vpnMsg = '恢复失败: ' + e;
      } finally {
        this.vpnRestoreBusy = false;
      }
    },

    vpnBackupBasename(path) {
      return path.split('/').pop();
    },

    vpnProfileTypeBadge(type) {
      const map = { remote: '订阅', merge: '合并', script: '脚本', rules: '规则', proxies: '节点', groups: '分组' };
      return map[type] || type;
    },

    vpnFormatTs(ts) {
      if (!ts) return '—';
      return new Date(ts * 1000).toLocaleString('zh-CN', { dateStyle: 'short', timeStyle: 'short' });
    },

    async loadVPNEnhanceData() {
      this.vpnEnhanceLoading = true;
      try {
        this.vpnEnhanceData = await window['go']['main']['App']['GetVPNEnhanceData']();
      } catch (e) {
        this.vpnMsg = '加载增强配置失败: ' + e;
      } finally {
        this.vpnEnhanceLoading = false;
      }
    },

    vpnEnhanceRuleType(rule) {
      if (!rule) return '';
      const prefix = rule.split(',')[0].trim().toUpperCase();
      const map = {
        'DOMAIN': '域名', 'DOMAIN-SUFFIX': '域名后缀', 'DOMAIN-KEYWORD': '域名关键词',
        'IP-CIDR': 'IP段', 'IP-CIDR6': 'IPv6段', 'GEOIP': '地理IP',
        'PROCESS-NAME': '进程', 'MATCH': '默认'
      };
      return map[prefix] || prefix;
    },

    vpnEnhanceRuleTarget(rule) {
      if (!rule) return '';
      const parts = rule.split(',');
      return parts[parts.length - 1]?.replace(',no-resolve', '').trim() || '';
    },

    vpnEnhanceRuleValue(rule) {
      if (!rule) return '';
      const parts = rule.split(',');
      return parts.length >= 2 ? parts[1]?.trim() : '';
    },
  };
};

Alpine.start();
