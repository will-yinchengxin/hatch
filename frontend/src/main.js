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
const ParseZshrc          = ()        => window['go']['main']['App']['ParseZshrc']();
const SaveZshrcAnnotation = (key, note) => window['go']['main']['App']['SaveZshrcAnnotation'](key, note);
const SaveAIConfig        = (ai)      => window['go']['main']['App']['SaveAIConfig'](ai);
const TestAIConfig        = (ai)      => window['go']['main']['App']['TestAIConfig'](ai);
const SuggestScriptMeta   = (meta)    => window['go']['main']['App']['SuggestScriptMeta'](meta);
const SuggestZshrcAnnotation = (entry) => window['go']['main']['App']['SuggestZshrcAnnotation'](entry);
const SuggestZshrcOrganization = (mode) => window['go']['main']['App']['SuggestZshrcOrganization'](mode);
const ApplyZshrcOrganization = (content) => window['go']['main']['App']['ApplyZshrcOrganization'](content);
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
    aiBusyScript: null,
    aiBusyEntry: null,
    zshrcInfo: null,
    newDir: { path: '', exportName: '', isScriptDir: true },
    savingExports: false,

    quickExportDir: null,
    quickExportName: '',

    // Edit drawer
    editTarget: null,
    editForm: {},
    editSaving: false,

    // zshrc analysis
    zshrcEntries: [],
    zshrcFilter: 'all',
    zshrcLoading: false,
    zshrcOrganizing: false,
    zshrcApplying: false,
    zshrcOrganizationMode: '按功能模块优先分组，同时把 export/PATH、alias、函数、source 初始化分别归类。',
    zshrcOrganization: null,
    zshrcApplyResult: null,
    annotationTarget: null,
    annotationText: '',
    expandedFns: {},

    zshrcFilterOptions: [
      { key: 'all',      label: '全部' },
      { key: 'function', label: '函数' },
      { key: 'alias',    label: '别名' },
      { key: 'export',   label: '导出' },
      { key: 'path',     label: 'PATH' },
      { key: 'source',   label: 'Source' },
      { key: 'other',    label: '其他' },
    ],

    autoLaunchEnabled: false,
    autoLaunchSaving: false,

    // Claude Code status
    claudeStatus: null,
    claudeStatusLoading: false,

    // Codex usage status
    codexStatus: null,
    codexStatusLoading: false,

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
      return !!(this.aiForm.enabled && this.aiForm.baseUrl && this.aiForm.apiKey && this.aiForm.model);
    },

    get filteredZshrcEntries() {
      if (this.zshrcFilter === 'all') return this.zshrcEntries;
      // 'section' is shown in 'all' only
      return this.zshrcEntries.filter(e => e.type === this.zshrcFilter);
    },

    // ── Lifecycle ──────────────────────────────────────────────────────
    async init() {
      await this.loadConfig();
      await this.scanAll();
      this.setupEvents();
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
        await AddDirectory(this.newDir.path, this.newDir.exportName || '', this.newDir.isScriptDir);
        await this.loadConfig();
        await this.scanAll();
        this.newDir = { path: '', exportName: '', isScriptDir: true };
      } catch (e) { this.showError(`添加目录失败: ${e}`);   }
    },

    async removeDirectory(path) {
      try {
        await RemoveDirectory(path);
        await this.loadConfig();
        await this.scanAll();
      } catch (e) { this.showError(`移除目录失败: ${e}`);   }
    },

    quickExport(dir) {
      this.quickExportDir = dir;
      const seg = dir.path.split('/').filter(Boolean).pop() || '';
      this.quickExportName = seg.replace(/[^a-zA-Z0-9_]/g, '_').toUpperCase();
    },

    async confirmQuickExport() {
      if (!this.quickExportDir || !this.quickExportName) return;
      try {
        await UpdateDirectoryExport(this.quickExportDir.path, this.quickExportName);
        await this.loadConfig();
        this.quickExportDir = null;
        this.quickExportName = '';
      } catch (e) { this.showError(`设置导出失败: ${e}`);   }
    },

    async clearExport(dir) {
      try {
        await UpdateDirectoryExport(dir.path, '');
        await this.loadConfig();
      } catch (e) { this.showError(`取消导出失败: ${e}`);   }
    },

    async saveExports() {
      this.savingExports = true;
      try {
        await SaveExports(this.cfg.exports || []);
        await this.loadConfig();
      } catch (e) { this.showError(`保存失败: ${e}`);   }
      finally { this.savingExports = false; }
    },

    addExportRow() {
      if (!this.cfg.exports) this.cfg.exports = [];
      this.cfg.exports = [...this.cfg.exports, { name: '', value: '' }];
    },

    removeExportRow(i) {
      this.cfg.exports = this.cfg.exports.filter((_, idx) => idx !== i);
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

    quickExportPreview() {
      return this.exportsPreview({
        name: this.quickExportName,
        value: this.quickExportDir?.path || '',
      });
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
        });
        await this.loadConfig();
        this.aiTestResult = '已保存';
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
      };
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
        };
        await SaveScriptMeta(meta);
        if (this.selectedDir) await this.scanDir(this.selectedDir);
        else await this.scanAll();
        this.cancelEdit();
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

    // ── zshrc analysis ────────────────────────────────────────────────
    async loadZshrcAnalysis() {
      this.zshrcLoading = true;
      try {
        this.zshrcEntries = await ParseZshrc() || [];
      } catch (e) { this.showError(`解析 zshrc 失败: ${e}`);   }
      finally { this.zshrcLoading = false; }
    },

    async organizeZshrcWithAI() {
      if (!this.aiReady) {
        this.selectedDir = null;
        this.tab = 'ai';
        this.showError('请先完成并启用 AI 配置');  
        return;
      }
      this.zshrcOrganizing = true;
      this.zshrcApplyResult = null;
      try {
        this.zshrcOrganization = await SuggestZshrcOrganization(this.zshrcOrganizationMode || '');
      } catch (e) {
        this.showError(`AI 整理 zshrc 失败: ${e}`);  
      } finally {
        this.zshrcOrganizing = false;
      }
    },

    async applyZshrcOrganization() {
      const content = this.zshrcOrganization?.content || '';
      if (!content.trim()) return;
      this.zshrcApplying = true;
      try {
        this.zshrcApplyResult = await ApplyZshrcOrganization(content);
        await this.loadZshrcAnalysis();
        await this.loadConfig();
      } catch (e) {
        this.showError(`写入 zshrc 失败: ${e}`);  
      } finally {
        this.zshrcApplying = false;
      }
    },

    clearZshrcOrganization() {
      this.zshrcOrganization = null;
      this.zshrcApplyResult = null;
    },

    zshrcCount(filter) {
      if (filter === 'all') return this.zshrcEntries.length;
      return this.zshrcEntries.filter(e => e.type === filter).length;
    },

    startAnnotate(entry) {
      this.annotationTarget = entry.key;
      this.annotationText = entry.note || '';
    },

    cancelAnnotate() {
      this.annotationTarget = null;
      this.annotationText = '';
    },

    async saveAnnotation(key) {
      try {
        await SaveZshrcAnnotation(key, this.annotationText);
        const e = this.zshrcEntries.find(x => x.key === key);
        if (e) e.note = this.annotationText;
        this.cancelAnnotate();
      } catch (e) { this.showError(`保存备注失败: ${e}`);   }
    },

    async deleteAnnotation(key) {
      try {
        await SaveZshrcAnnotation(key, '');
        const e = this.zshrcEntries.find(x => x.key === key);
        if (e) e.note = '';
        this.cancelAnnotate();
      } catch (e) { this.showError(`删除备注失败: ${e}`);   }
    },

    async fillAnnotationWithAI(entry) {
      if (!this.aiReady) {
        this.selectedDir = null;
        this.tab = 'ai';
        this.showError('请先完成并启用 AI 配置');  
        return;
      }
      this.aiBusyEntry = entry.key;
      try {
        const note = await SuggestZshrcAnnotation(entry);
        this.annotationTarget = entry.key;
        this.annotationText = note;
      } catch (e) {
        this.showError(`AI 生成备注失败: ${e}`);  
      } finally {
        this.aiBusyEntry = null;
      }
    },

    toggleFn(key) {
      this.expandedFns = { ...this.expandedFns, [key]: !this.expandedFns[key] };
    },

    // ── AI trace log ──────────────────────────────────────────────────
    async loadAITraces() {
      try {
        this.aiTraces = await GetAITraces() || [];
      } catch (_) {}
    },

    async clearAITracesLog() {
      try {
        await ClearAITraces();
        this.aiTraces = [];
        this.aiTraceExpanded = {};
      } catch (_) {}
    },

    toggleAITrace(id) {
      this.aiTraceExpanded = { ...this.aiTraceExpanded, [id]: !this.aiTraceExpanded[id] };
    },

    aiTraceStatus(t) {
      return t.error ? 'error' : 'ok';
    },

    zshrcTypeIcon(type) {
      return {
        function: 'ƒ',
        alias:    '≡',
        export:   '$',
        path:     '⊕',
        source:   '↯',
        section:  '#',
        other:    '…',
      }[type] || '?';
    },

    // ── Claude Code / Codex usage status ──────────────────────────────
    async loadAIUsageStatuses() {
      if (!this.claudeStatus) this.loadClaudeStatus();
      if (!this.codexStatus) this.loadCodexStatus();
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
        await CreateScriptFile(
          this.selectedDir.path,
          this.newScriptForm.filename.trim(),
          this.newScriptForm.content
        );
        await this.scanDir(this.selectedDir);
        this.closeNewScript();
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
        this.errorTimer = setTimeout(() => {
          this.error = '';
          this.errorTimer = null;
        }, 6000);
      } else {
        this.errorTimer = null;
      }
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
  };
};

Alpine.start();
