// AI › 模型榜单. Boards come from GetModelLeaderboards (cache, instant) and
// RefreshModelLeaderboards (network, stale feeds only unless forced). Each board
// keeps its source ranking; scores are never combined or compared across boards.

// Muted vendor tones so the chart stays inside Hatch's warm palette.
const BRANDS = {
  anthropic: { color: '#c47a5a', icon: 'anthropic' },
  openai:    { color: '#4a4540', icon: 'openai' },
  google:    { color: '#4f8f64', icon: 'google-color' },
  meta:      { color: '#3f7fb8', icon: 'meta-color' },
  xai:       { color: '#5b5650', icon: 'grok' },
  deepseek:  { color: '#5671c4', icon: 'deepseek-color' },
  alibaba:   { color: '#7466b8', icon: 'qwen-color' },
  moonshot:  { color: '#3b78b5', icon: 'kimi' },
  zai:       { color: '#55525e', icon: 'zai' },
  minimax:   { color: '#b65a7c', icon: 'minimax-color' },
  mistral:   { color: '#cf8040', icon: 'mistral-color' },
  xiaomi:    { color: '#d6844a', icon: 'xiaomimimo' },
};
const OTHER_TONES = ['#a08c78', '#8d9a7c', '#7d919c', '#9c8296', '#b09868'];

const DECIMALS = { elo: 0, index: 1, ips: 3, percent: 1 };
const DAY = 86400000;
const CHART_SIZE = 10;
const LIST_SIZE = 20;
const REC_DEPTH = 20; // places that earn recommendation points: #1 → 20 … #20 → 1
const REC_COUNT = 4;

// "Claude Opus 5.5" (Arena) and "claude-opus-5-5" (Epoch slug) are the same model.
export function modelKey(name) {
  return String(name).toLowerCase().replace(/[^a-z0-9\u4e00-\u9fa5]/g, '');
}
const looksDisplayed = name => /[A-Z\s]/.test(name);

// Recommendation = sum of place points over every board in the domain, best
// variant per model. Places, not scores: Elo, IPS and accuracy are not comparable.
export function recommend(boards) {
  const models = new Map();
  for (const board of boards) {
    const seen = new Set();
    for (const e of board.entries) {
      if (e.rank > REC_DEPTH) break;
      const key = modelKey(e.model);
      if (!key || seen.has(key)) continue;
      seen.add(key);
      const m = models.get(key) || { key, name: e.model, org: e.org, brand: e.brand, points: 0, hits: [], best: null };
      if (!looksDisplayed(m.name) && looksDisplayed(e.model)) m.name = e.model;
      if (!m.org && e.org) m.org = e.org;
      m.points += REC_DEPTH + 1 - e.rank;
      const hit = { board: board.id, title: board.title, tag: board.tag, rank: e.rank, variant: e.variant || '' };
      m.hits.push(hit);
      if (!m.best || hit.rank < m.best.rank) m.best = hit;
      models.set(key, m);
    }
  }
  return [...models.values()]
    .sort((a, b) => b.points - a.points || b.hits.length - a.hits.length || a.best.rank - b.best.rank)
    .slice(0, REC_COUNT);
}

function niceStep(raw) {
  const pow = 10 ** Math.floor(Math.log10(raw));
  const unit = raw / pow;
  return (unit <= 1 ? 1 : unit <= 2 ? 2 : unit <= 2.5 ? 2.5 : unit <= 5 ? 5 : 10) * pow;
}

function hash(text) {
  let h = 0;
  for (const ch of String(text)) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
  return h;
}

// Chart scale: accuracy keeps its true 0–100 axis; ratings (Elo/IPS/ECI) are
// interval scales where zero means nothing, so the axis is fitted to the data
// and the note says where it starts.
export function chartScale(format, entries) {
  if (format === 'percent') return { min: 0, max: 100, step: 25, zeroBased: true };
  const values = entries.flatMap(e => [e.score, e.lower || e.score, e.upper || e.score]);
  const lo = Math.min(...values), hi = Math.max(...values);
  const span = hi - lo || Math.abs(hi) || 1;
  const step = niceStep((span * 1.5) / 4);
  const min = Math.floor((lo - span * 0.4) / step) * step;
  const max = Math.ceil((hi + span * 0.05) / step) * step;
  return { min, max, step, zeroBased: false };
}

export function leaderboardState() {
  return {
    lbBoards: [], lbLoading: false, lbRefreshing: false, lbStale: false, lbSyncedAt: 0, lbError: '',
    lbDomain: 'coding', lbPick: {}, lbQuery: '', lbExpanded: false,
    lbDomains: [
      { id: 'coding', name: '编程' }, { id: 'image', name: '图像' },
      { id: 'frontend', name: '前端' }, { id: 'general', name: '综合' },
    ],

    lbApply(snap) {
      if (!snap) return;
      this.lbBoards = snap.boards || [];
      this.lbStale = !!snap.stale;
      this.lbSyncedAt = snap.syncedAt || 0;
    },
    async loadLeaderboards() {
      if (this.lbLoading) return;
      this.lbLoading = true; this.lbError = '';
      let snap = null;
      try { snap = await window.go.app.App.GetModelLeaderboards(); this.lbApply(snap); }
      catch (e) { this.lbError = String(e); }
      finally { this.lbLoading = false; }
      // A refresh already running (page reopened mid-flight) is joined, not duplicated.
      if (snap && (snap.stale || snap.refreshing)) this.refreshLeaderboards(false);
    },
    async refreshLeaderboards(force) {
      if (this.lbRefreshing) return;
      this.lbRefreshing = true; this.lbError = '';
      try { this.lbApply(await window.go.app.App.RefreshModelLeaderboards(!!force)); }
      catch (e) { this.lbError = String(e); }
      finally { this.lbRefreshing = false; }
    },
    lbOpen(url) { if (url) window.runtime.BrowserOpenURL(url); },

    lbDomainBoards(domain) { return this.lbBoards.filter(b => b.domain === domain); },
    lbBoard() {
      const boards = this.lbDomainBoards(this.lbDomain);
      return boards.find(b => b.id === this.lbPick[this.lbDomain]) || boards[0] || null;
    },
    lbSelectDomain(id) { this.lbDomain = id; this.lbQuery = ''; this.lbExpanded = false; },
    lbSelectBoard(id) { this.lbPick = { ...this.lbPick, [this.lbDomain]: id }; this.lbQuery = ''; this.lbExpanded = false; },

    lbDomainName() { return (this.lbDomains.find(d => d.id === this.lbDomain) || {}).name || ''; },
    lbRecommend() { return recommend(this.lbDomainBoards(this.lbDomain).filter(b => b.entries.length)); },
    lbRecNote() {
      const n = this.lbDomainBoards(this.lbDomain).filter(b => b.entries.length).length;
      return `按本领域 ${n} 个榜单的名次汇总：各榜第 1 名记 ${REC_DEPTH} 分，依次递减到第 ${REC_DEPTH} 名记 1 分，同一模型只计最佳配置；不合并各榜原始分数，仅供参考。`;
    },
    // Jump to the board where the model places best and filter the list to it.
    lbFocusRec(rec) { this.lbSelectBoard(rec.best.board); this.lbQuery = rec.name; },

    lbRows() {
      const board = this.lbBoard();
      if (!board) return [];
      const q = this.lbQuery.trim().toLowerCase();
      if (q) return board.entries.filter(e => `${e.model} ${e.variant || ''} ${e.org}`.toLowerCase().includes(q));
      return this.lbExpanded ? board.entries : board.entries.slice(0, LIST_SIZE);
    },
    lbCanExpand() {
      const board = this.lbBoard();
      return !!board && !this.lbQuery.trim() && board.entries.length > LIST_SIZE;
    },

    lbChart() {
      const board = this.lbBoard();
      const entries = board ? board.entries.slice(0, CHART_SIZE) : [];
      if (!entries.length) return { cols: [], ticks: [], note: '' };
      const s = chartScale(board.format, entries);
      const pct = v => Math.max(0, Math.min(100, ((v - s.min) / (s.max - s.min)) * 100));
      const ticks = [];
      for (let v = s.min; v <= s.max + s.step / 1e6; v += s.step) ticks.push({ y: pct(v), label: this.lbNumber(board, v, true) });
      // Scores that straddle zero (Agent Arena IPS) grow from 0; otherwise from the axis floor.
      const base = pct(s.min < 0 && s.max > 0 ? 0 : s.min);
      const cols = entries.map(e => ({
        key: `${e.rank}-${e.model}-${e.variant || ''}`, entry: e, brand: this.lbBrand(e),
        b: Math.min(base, pct(e.score)), h: Math.abs(pct(e.score) - base), top: Math.max(base, pct(e.score)),
        score: this.lbNumber(board, e.score),
        ci: !!(e.lower && e.upper), lo: pct(e.lower || e.score), hi: pct(e.upper || e.score),
        title: `#${e.rank} ${e.model}${e.variant ? ` (${e.variant})` : ''} · ${this.lbNumber(board, e.score)} ${board.unit}${e.lower && e.upper ? ` · 区间 ${this.lbNumber(board, e.lower)}–${this.lbNumber(board, e.upper)}` : ''}`,
      }));
      const axis = s.zeroBased ? '纵轴 0–100%'
        : base > 0 ? '柱以 0 为基线'
        : `纵轴从 ${this.lbNumber(board, s.min, true)} 起，柱高仅表示相对差距`;
      const note = `${axis}${cols.some(c => c.ci) ? ' · 细线为置信区间' : ''}`;
      return { cols, ticks, note };
    },

    lbNumber(board, value, axis = false) {
      const digits = DECIMALS[board.format] ?? 2;
      const text = new Intl.NumberFormat('en-US', {
        minimumFractionDigits: axis ? 0 : digits, maximumFractionDigits: digits, useGrouping: false,
      }).format(value);
      return board.format === 'percent' ? `${text}%` : text;
    },
    lbInterval(board, e) {
      if (!e.lower || !e.upper) return '';
      const up = e.upper - e.score, down = e.score - e.lower;
      const fmt = v => this.lbNumber(board, v);
      return Math.abs(up - down) < 10 ** -(DECIMALS[board.format] ?? 2) ? `±${fmt(up)}` : `+${fmt(up)} / −${fmt(down)}`;
    },
    lbMeta(board, e) {
      if (e.votes) return `${e.votes.toLocaleString('en-US')} ${board.format === 'ips' ? '会话' : '票'}`;
      if (e.released) return `${e.released} 发布`;
      return '';
    },
    lbMetaLabel(board) {
      if (!board) return '';
      if (board.entries.some(e => e.votes)) return board.format === 'ips' ? '会话数' : '投票数';
      return board.entries.some(e => e.released) ? '发布日期' : '';
    },
    lbIsNew(e) {
      if (!e.released) return false;
      const [y, m, d] = e.released.split('-').map(Number);
      const now = new Date();
      const days = (Date.UTC(now.getFullYear(), now.getMonth(), now.getDate()) - Date.UTC(y, m - 1, d)) / DAY;
      return days >= 0 && days <= 30;
    },
    lbBrand(e) {
      const known = BRANDS[e.brand];
      const label = e.org || e.model;
      if (known) return { color: known.color, icon: `./leaderboard-brands/${known.icon}.svg`, initial: '' };
      return { color: OTHER_TONES[hash(e.brand || label) % OTHER_TONES.length], icon: '', initial: (label.match(/[A-Za-z0-9一-龥]/) || ['?'])[0].toUpperCase() };
    },

    lbSyncLabel() {
      if (this.lbRefreshing) return '正在同步…';
      if (!this.lbSyncedAt) return '尚未同步';
      const minutes = Math.floor((Date.now() / 1000 - this.lbSyncedAt) / 60);
      const ago = minutes < 1 ? '刚刚' : minutes < 60 ? `${minutes} 分钟前` : minutes < 1440 ? `${Math.floor(minutes / 60)} 小时前` : `${Math.floor(minutes / 1440)} 天前`;
      return `${ago}同步`;
    },
    lbBoardState(board) {
      if (!board) return '';
      if (board.error && board.entries.length) return `最近一次同步失败，显示缓存：${board.error}`;
      return board.error;
    },
  };
}
