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
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

window.hatchApp = function () {
  return {
    // ── State ──────────────────────────────────────────────────────────
    tab: 'scripts',

    // Directory navigation
    selectedDir: null,       // { path, exportName, isScriptDir }
    allScripts: [],          // result of ScanAll (all dirs combined)
    currentScripts: [],      // scripts in selectedDir

    query: '',
    page: 1,
    pageSize: 15,
    loading: false,
    error: '',

    statuses: {},
    runningSet: {},
    scriptArgs: {},
    scriptLogs: {},
    logTarget: null,

    cfg: { directories: [], exports: [] },
    zshrcInfo: null,
    newDir: { path: '', exportName: '', isScriptDir: true },
    savingExports: false,

    quickExportDir: null,
    quickExportName: '',

    // ── Computed ───────────────────────────────────────────────────────
    get scriptDirs() {
      return (this.cfg.directories || []).filter(d => d.isScriptDir);
    },

    // Count of scripts per dir, built from allScripts
    get dirScriptCount() {
      const counts = {};
      for (const s of this.allScripts) {
        const dir = this.cfg.directories?.find(d => s.path.startsWith(d.path + '/') || s.path.startsWith(d.path + '\\'));
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
      } catch (e) { this.error = `加载配置失败: ${e}`; }
      try { this.zshrcInfo = await CheckZshrc(); } catch (_) {}
    },

    // Scan all dirs to populate directory card counters (background, no loading spinner)
    async scanAll() {
      this.loading = true;
      try {
        this.allScripts = await ScanAll() || [];
        await this.refreshStatuses();
      } catch (e) { this.error = `扫描失败: ${e}`; }
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
        // Also refresh allScripts count
        const idx = this.allScripts.findIndex(s => s.path.startsWith(dir.path));
        if (idx === -1) this.scanAll(); // first time
      } catch (e) { this.error = `扫描目录失败: ${e}`; }
      finally { this.loading = false; }
    },

    // ── Events ─────────────────────────────────────────────────────────
    setupEvents() {
      EventsOn('executor:log', (ev) => {
        const { id, line, stream } = ev;
        if (!this.scriptLogs[id]) this.scriptLogs[id] = [];
        this.scriptLogs[id].push({ line, stream });
        if (this.scriptLogs[id].length > 500) this.scriptLogs[id] = this.scriptLogs[id].slice(-500);
        this.$nextTick(() => {
          const el = document.getElementById('log-' + btoa(id));
          if (el) el.scrollTop = el.scrollHeight;
        });
      });
      EventsOn('executor:done', (ev) => {
        delete this.runningSet[ev.id];
        this.runningSet = { ...this.runningSet };
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
      this.logTarget = id;
      const raw = (this.scriptArgs[id] || '').trim();
      const args = raw ? raw.split(/\s+/) : [];
      try {
        await RunScript(id, script.path, args);
      } catch (e) {
        this.error = `运行失败: ${e}`;
        delete this.runningSet[id];
        this.runningSet = { ...this.runningSet };
      }
    },

    async serviceAction(script, action) {
      try {
        await ServiceAction(script.path, action, []);
        await this.refreshStatuses();
      } catch (e) { this.error = `${action} 失败: ${e}`; }
    },

    logLines(path) { return this.scriptLogs[path] || []; },

    clearLog(path) {
      this.scriptLogs[path] = [];
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
      } catch (e) { this.error = `添加目录失败: ${e}`; }
    },

    async removeDirectory(path) {
      try {
        await RemoveDirectory(path);
        await this.loadConfig();
        await this.scanAll();
      } catch (e) { this.error = `移除目录失败: ${e}`; }
    },

    // Quick export flow
    quickExport(dir) {
      this.quickExportDir = dir;
      // Suggest a name from the last segment of the path
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
      } catch (e) { this.error = `设置导出失败: ${e}`; }
    },

    async clearExport(dir) {
      try {
        await UpdateDirectoryExport(dir.path, '');
        await this.loadConfig();
      } catch (e) { this.error = `取消导出失败: ${e}`; }
    },

    async saveExports() {
      this.savingExports = true;
      try {
        await SaveExports(this.cfg.exports || []);
        await this.loadConfig();
      } catch (e) { this.error = `保存失败: ${e}`; }
      finally { this.savingExports = false; }
    },

    addExportRow() {
      if (!this.cfg.exports) this.cfg.exports = [];
      this.cfg.exports = [...this.cfg.exports, { name: '', value: '' }];
    },

    removeExportRow(i) {
      this.cfg.exports = this.cfg.exports.filter((_, idx) => idx !== i);
    },

    // ── Helpers ────────────────────────────────────────────────────────
    langIcon(lang) {
      return { python: '🐍', shell: '🐚', ruby: '💎', node: '🟩', binary: '⚙️' }[lang] || '📄';
    },

    cmdLabel(cmd) {
      return { start: '▶ start', stop: '■ stop', restart: '↺ restart', reload: '↺ reload', status: '? status' }[cmd] || cmd;
    },

    baseName(path) {
      if (!path) return '';
      return path.split('/').filter(Boolean).pop() || path;
    },

    dirTypeSummary(dirPath) {
      const scripts = this.allScripts.filter(s => s.path.startsWith(dirPath + '/') || s.path.startsWith(dirPath + '\\'));
      const svcCount = scripts.filter(s => s.type === 'service').length;
      const oneshotCount = scripts.filter(s => s.type === 'oneshot').length;
      const parts = [];
      if (svcCount > 0) parts.push(`${svcCount} service`);
      if (oneshotCount > 0) parts.push(`${oneshotCount} oneshot`);
      return parts.join(' · ');
    },
  };
};

Alpine.start();
