import '@picocss/pico/css/pico.min.css';
import './style.css';
import Alpine from 'alpinejs';

import {
  GetConfig,
  ScanAll,
  AddDirectory,
  RemoveDirectory,
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
    scripts: [],
    query: '',
    page: 1,
    pageSize: 15,
    loading: false,
    error: '',

    statuses: {},       // { [path]: 'running' | 'stopped' }
    runningSet: {},     // { [path]: true } for oneshot in-flight
    scriptArgs: {},     // { [path]: argsString }
    scriptLogs: {},     // { [path]: [{line, stream}] }
    logTarget: null,    // path whose log panel is open

    cfg: { directories: [], exports: [] },
    zshrcInfo: null,
    newDir: { path: '', exportName: '', isScriptDir: true },
    savingExports: false,

    // ── Computed ───────────────────────────────────────────────────────
    get filteredScripts() {
      if (!this.query.trim()) return this.scripts;
      const q = this.query.toLowerCase();
      return this.scripts.filter(s =>
        (s.name || '').toLowerCase().includes(q) ||
        (s.desc || '').toLowerCase().includes(q) ||
        (s.language || '').toLowerCase().includes(q) ||
        (s.tags || []).some(t => t.toLowerCase().includes(q))
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
      await this.scan();
      this.setupEvents();
      // Poll service statuses every 3 s
      setInterval(() => this.refreshStatuses(), 3000);
    },

    async loadConfig() {
      try {
        this.cfg = await GetConfig() || { directories: [], exports: [] };
      } catch (e) {
        this.error = `加载配置失败: ${e}`;
      }
      try {
        this.zshrcInfo = await CheckZshrc();
      } catch (_) {}
    },

    async scan() {
      this.loading = true;
      try {
        this.scripts = await ScanAll() || [];
        this.page = 1;
        await this.refreshStatuses();
      } catch (e) {
        this.error = `扫描失败: ${e}`;
      } finally {
        this.loading = false;
      }
    },

    async refreshStatuses() {
      try {
        const s = await GetAllServiceStatuses();
        this.statuses = s || {};
      } catch (_) {}
    },

    // ── Events ─────────────────────────────────────────────────────────
    setupEvents() {
      EventsOn('executor:log', (ev) => {
        const { id, line, stream } = ev;
        if (!this.scriptLogs[id]) this.scriptLogs[id] = [];
        this.scriptLogs[id].push({ line, stream });
        if (this.scriptLogs[id].length > 500) {
          this.scriptLogs[id] = this.scriptLogs[id].slice(-500);
        }
        // Auto-scroll log panel
        this.$nextTick(() => {
          const el = document.getElementById('log-' + btoa(id));
          if (el) el.scrollTop = el.scrollHeight;
        });
      });
      EventsOn('executor:done', (ev) => {
        const { id } = ev;
        delete this.runningSet[id];
        this.runningSet = { ...this.runningSet };
      });
      EventsOn('executor:status', (ev) => {
        const { id, status } = ev;
        this.statuses = { ...this.statuses, [id]: status };
        if (status !== 'running') {
          delete this.runningSet[id];
          this.runningSet = { ...this.runningSet };
        }
      });
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
      } catch (e) {
        this.error = `${action} 失败: ${e}`;
      }
    },

    logLines(path) {
      return this.scriptLogs[path] || [];
    },

    clearLog(path) {
      this.scriptLogs[path] = [];
      if (this.logTarget === path) this.logTarget = null;
    },

    // ── Settings operations ────────────────────────────────────────────
    async addDirectory() {
      if (!this.newDir.path) return;
      try {
        await AddDirectory(this.newDir.path, this.newDir.exportName || '', this.newDir.isScriptDir);
        await this.loadConfig();
        if (this.newDir.isScriptDir) await this.scan();
        this.newDir = { path: '', exportName: '', isScriptDir: true };
      } catch (e) {
        this.error = `添加目录失败: ${e}`;
      }
    },

    async removeDirectory(path) {
      try {
        await RemoveDirectory(path);
        await this.loadConfig();
        await this.scan();
      } catch (e) {
        this.error = `移除目录失败: ${e}`;
      }
    },

    async saveExports() {
      this.savingExports = true;
      try {
        await SaveExports(this.cfg.exports || []);
        await this.loadConfig();
      } catch (e) {
        this.error = `保存失败: ${e}`;
      } finally {
        this.savingExports = false;
      }
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
  };
};

Alpine.start();
