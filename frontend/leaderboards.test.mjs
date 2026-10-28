import test from 'node:test';
import assert from 'node:assert/strict';
import { leaderboardState, chartScale, recommend, modelKey } from './src/leaderboards.js';

const entry = (rank, score, extra = {}) => ({ rank, model: `Model ${rank}`, org: 'Org', brand: '', score, ...extra });

test('rating axes are fitted to the data; accuracy keeps a true zero', () => {
  const elo = chartScale('elo', [entry(1, 1551.6, { lower: 1539, upper: 1564 }), entry(2, 1500)]);
  assert.ok(elo.min < 1500 && elo.min > 1300, `elo min ${elo.min}`);
  assert.ok(elo.max >= 1564, `elo max ${elo.max}`);
  assert.equal(elo.zeroBased, false);
  assert.equal((elo.max - elo.min) % elo.step, 0);
  assert.deepEqual(chartScale('percent', [entry(1, 96.4)]), { min: 0, max: 100, step: 25, zeroBased: true });
  const ips = chartScale('ips', [entry(1, 0.2077), entry(2, -0.02)]);
  assert.ok(ips.min <= -0.02 && ips.max >= 0.2077, 'ips handles negative scores');
});

test('domain and board selection stay independent; query and expansion reset', () => {
  const s = leaderboardState();
  s.lbBoards = [
    { id: 'a', domain: 'coding', entries: [] },
    { id: 'b', domain: 'coding', entries: [] },
    { id: 'c', domain: 'image', entries: [] },
  ];
  assert.equal(s.lbBoard().id, 'a');
  s.lbSelectBoard('b');
  s.lbQuery = 'x'; s.lbExpanded = true;
  s.lbSelectDomain('image');
  assert.equal(s.lbBoard().id, 'c');
  assert.equal(s.lbQuery, '');
  assert.equal(s.lbExpanded, false);
  s.lbSelectDomain('coding');
  assert.equal(s.lbBoard().id, 'b', 'remembers the board picked per domain');
});

test('rows show 20 by default, 50 when expanded, and filter by model, variant or org', () => {
  const s = leaderboardState();
  const entries = Array.from({ length: 50 }, (_, i) => entry(i + 1, 1500 - i, { org: i === 30 ? 'Anthropic' : 'Org', variant: i === 40 ? 'Max' : '' }));
  s.lbBoards = [{ id: 'a', domain: 'coding', format: 'elo', unit: 'Arena 分', entries }];
  assert.equal(s.lbRows().length, 20);
  assert.ok(s.lbCanExpand());
  s.lbExpanded = true;
  assert.equal(s.lbRows().length, 50);
  s.lbQuery = 'anthropic';
  assert.deepEqual(s.lbRows().map(e => e.rank), [31]);
  s.lbQuery = 'max';
  assert.deepEqual(s.lbRows().map(e => e.rank), [41]);
  assert.equal(s.lbCanExpand(), false);
});

test('chart takes the top 10 with whiskers only where the source gives an interval', () => {
  const s = leaderboardState();
  const entries = Array.from({ length: 12 }, (_, i) => entry(i + 1, 1500 - i * 5, i === 0 ? { lower: 1490, upper: 1510, brand: 'anthropic' } : {}));
  s.lbBoards = [{ id: 'a', domain: 'coding', format: 'elo', unit: 'Arena 分', entries }];
  const chart = s.lbChart();
  assert.equal(chart.cols.length, 10);
  assert.equal(chart.cols[0].ci, true);
  assert.equal(chart.cols[1].ci, false);
  assert.ok(chart.cols[0].h > chart.cols[9].h);
  assert.ok(chart.cols[0].hi > chart.cols[0].h && chart.cols[0].lo < chart.cols[0].h);
  assert.match(chart.cols[0].brand.icon, /anthropic\.svg$/);
  assert.match(chart.note, /纵轴从 \d+ 起/);
  assert.equal(chart.cols[0].b, 0, 'rating bars start at the axis floor');
  assert.ok(chart.ticks.length >= 3);
});

test('scores straddling zero grow from a zero baseline', () => {
  const s = leaderboardState();
  s.lbBoards = [{ id: 'a', domain: 'coding', format: 'ips', unit: 'IPS', entries: [entry(1, 0.2), entry(2, -0.05)] }];
  const [up, down] = s.lbChart().cols;
  assert.ok(up.b > 0 && down.b < up.b, 'negative bar hangs below the zero line');
  assert.equal(Math.round(up.b), Math.round(down.b + down.h));
  assert.match(s.lbChart().note, /以 0 为基线/);
});

test('numbers, intervals and meta follow the board format', () => {
  const s = leaderboardState();
  const elo = { format: 'elo', entries: [] }, pct = { format: 'percent', entries: [] }, ips = { format: 'ips', entries: [] };
  assert.equal(s.lbNumber(elo, 1551.57), '1552');
  assert.equal(s.lbNumber(pct, 96.4), '96.4%');
  assert.equal(s.lbNumber(ips, 0.2077), '0.208');
  assert.equal(s.lbInterval(elo, { score: 1548, lower: 1535.2, upper: 1560.8 }), '±13');
  assert.equal(s.lbInterval({ format: 'index' }, { score: 168.83, lower: 165.25, upper: 173.95 }), '+5.1 / −3.6');
  assert.equal(s.lbInterval(elo, { score: 1500 }), '');
  assert.equal(s.lbMeta(elo, { votes: 12345 }), '12,345 票');
  assert.equal(s.lbMeta(ips, { votes: 50 }), '50 会话');
  assert.equal(s.lbMeta(pct, { released: '2026-09-22' }), '2026-09-22 发布');
});

test('new badge uses the local calendar and never marks future or old releases', () => {
  const s = leaderboardState();
  const iso = d => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  const now = new Date();
  assert.equal(s.lbIsNew({ released: iso(now) }), true);
  assert.equal(s.lbIsNew({ released: iso(new Date(now.getTime() - 40 * 86400000)) }), false);
  assert.equal(s.lbIsNew({ released: '2099-01-01' }), false);
  assert.equal(s.lbIsNew({}), false);
});

test('unknown vendors get a stable tinted initial', () => {
  const s = leaderboardState();
  const a = s.lbBrand({ brand: 'kling', org: 'Kling', model: 'Kling 3' });
  assert.equal(a.icon, '');
  assert.equal(a.initial, 'K');
  assert.equal(a.color, s.lbBrand({ brand: 'kling', org: 'Kling', model: 'other' }).color);
});

test('load joins a stale cache with one non-forced refresh', async () => {
  const calls = [];
  globalThis.window = { go: { app: { App: {
    GetModelLeaderboards: async () => ({ boards: [{ id: 'a', domain: 'coding', entries: [] }], stale: true, syncedAt: 0 }),
    RefreshModelLeaderboards: async force => { calls.push(force); return { boards: [{ id: 'a', domain: 'coding', entries: [entry(1, 1)] }], stale: false, syncedAt: 100 }; },
  } } } };
  const s = leaderboardState();
  await s.loadLeaderboards();
  await new Promise(r => setTimeout(r, 0));
  assert.deepEqual(calls, [false]);
  assert.equal(s.lbBoards[0].entries.length, 1);
  assert.equal(s.lbSyncedAt, 100);
  assert.equal(s.lbRefreshing, false);
  delete globalThis.window;
});

test('recommendations sum place points across boards, matching names across sources', () => {
  assert.equal(modelKey('Claude Opus 5.5'), modelKey('claude-opus-5-5'));
  const arena = { id: 'arena', title: 'Text Arena', tag: '人类偏好', entries: [
    { rank: 1, model: 'Gemini 4 Argon', variant: 'High', org: 'Google', brand: 'google' },
    { rank: 2, model: 'Claude Opus 5.5', variant: 'High', org: 'Anthropic', brand: 'anthropic' },
    { rank: 3, model: 'Claude Opus 5.5', variant: 'Max', org: 'Anthropic', brand: 'anthropic' },
    { rank: 4, model: 'GPT-6 Astra', variant: 'Max', org: 'OpenAI', brand: 'openai' },
    { rank: 21, model: 'Too Low', org: 'X', brand: '' },
  ] };
  const epoch = { id: 'epoch', title: 'ECI', tag: '综合能力指数', entries: [
    { rank: 1, model: 'claude-opus-5-5', variant: 'Max', org: 'Anthropic', brand: 'anthropic' },
    { rank: 2, model: 'GPT-6 Astra', org: 'OpenAI', brand: 'openai' },
    { rank: 3, model: 'Solo Model', org: 'Y', brand: '' },
  ] };
  const recs = recommend([arena, epoch]);
  assert.deepEqual(recs.map(r => r.name), ['Claude Opus 5.5', 'GPT-6 Astra', 'Gemini 4 Argon', 'Solo Model']);
  assert.equal(recs[0].points, 19 + 20, 'best variant only: #2 on Arena, #1 on Epoch');
  assert.deepEqual(recs[0].best, { board: 'epoch', title: 'ECI', tag: '综合能力指数', rank: 1, variant: 'Max' });
  assert.equal(recs[0].hits.length, 2);
  assert.ok(!recs.some(r => r.name === 'Too Low'));
});

test('clicking a recommendation opens its best board filtered to the model', () => {
  const s = leaderboardState();
  s.lbBoards = [
    { id: 'a', domain: 'general', title: 'A', tag: 'a', entries: [{ rank: 5, model: 'M', org: '' }] },
    { id: 'b', domain: 'general', title: 'B', tag: 'b', entries: [{ rank: 1, model: 'M', org: '' }] },
  ];
  s.lbSelectDomain('general');
  const [rec] = s.lbRecommend();
  s.lbFocusRec(rec);
  assert.equal(s.lbBoard().id, 'b');
  assert.equal(s.lbQuery, 'M');
  assert.equal(s.lbDomainName(), '综合');
});
