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
const SetSkillSync          = (name, target, enabled)       => window['go']['main']['App']['SetSkillSync'](name, target, enabled);
const ImportSkillsFrom      = (source)                      => window['go']['main']['App']['ImportSkillsFrom'](source);
const CreateSkillTemplate   = (name)                        => window['go']['main']['App']['CreateSkillTemplate'](name);
const GetAppDataPaths       = ()                            => window['go']['main']['App']['GetAppDataPaths']();
const SaveAppDataPaths      = (traceDir, skillsDir)         => window['go']['main']['App']['SaveAppDataPaths'](traceDir, skillsDir);
const GetAIProviders        = ()                            => window['go']['main']['App']['GetAIProviders']();
const SaveAIProviders       = (providers, activeID)         => window['go']['main']['App']['SaveAIProviders'](providers, activeID);
const GetDocLinks           = ()                            => window['go']['main']['App']['GetDocLinks']();
const SaveDocLinks          = (links)                       => window['go']['main']['App']['SaveDocLinks'](links);
const OpenFile              = (path)                        => window['go']['main']['App']['OpenFile'](path);
const GetClaudeSettings     = ()                            => window['go']['main']['App']['GetClaudeSettings']();
const SaveClaudeSettings    = (content)                     => window['go']['main']['App']['SaveClaudeSettings'](content);
const GetCodexConfig        = ()                            => window['go']['main']['App']['GetCodexConfig']();
const SaveCodexConfig       = (content)                     => window['go']['main']['App']['SaveCodexConfig'](content);

window.hatchApp = function () {
  return {
    // ── State ──────────────────────────────────────────────────────────
    tab: 'scripts',

    selectedDir: null,
    allScripts: [],
    currentScripts: [],

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
    skillEditTarget: null,   // skill name being edited, or null
    skillEditContent: '',
    skillEditSaving: false,
    skillEditIsNew: false,
    skillNewName: '',
    skillSyncBusy: {},       // { "name:target": true }
    skillImporting: null,    // null | "claude" | "codex"
    skillImportResult: null,

    // AI sub-navigation
    aiSubTab: 'overview',    // 'overview' | 'connect' | 'skills' | 'tools'

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
      };
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

    async saveSkillContent() {
      if (!this.skillEditTarget) return;
      this.skillEditSaving = true;
      try {
        const name = this.skillEditTarget;
        await SaveSkillContent(name, this.skillEditContent);
        await this.loadSkills();
        this.closeSkillEditor();
        this.showSuccess(`Skill「${name}」已保存`);
      } catch (e) { this.showError(`保存 Skill 失败: ${e}`); }
      finally { this.skillEditSaving = false; }
    },

    async deleteSkill(name) {
      try {
        await DeleteSkill(name);
        if (this.skillEditTarget === name) this.closeSkillEditor();
        await this.loadSkills();
        this.showSuccess(`Skill「${name}」已删除`);
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
        if (n > 0) this.showSuccess(`已从 ${source === 'claude' ? 'Claude' : 'Codex'} 导入 ${n} 个 Skill`);
        else this.showSuccess('无新 Skill 可导入（已全部存在）');
      } catch (e) { this.showError(`导入失败: ${e}`); }
      finally { this.skillImporting = null; }
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
  };
};

Alpine.start();
