import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';

// Exercise the production Alpine methods without importing CSS or Wails runtime.
const source = fs.readFileSync(new URL('./src/main.js', import.meta.url), 'utf8');
const methods = source.slice(source.indexOf('    get filteredDocLinks()'), source.indexOf('    async openDocFile(path)'));
const docs = [
  { label: 'Claude 指南', path: '/docs/claude.html' },
  { label: 'Go 入门', path: '/docs/backend.pdf' },
  { label: 'Codex 指南', path: '/docs/codex.html' },
  { label: 'Python 入门', path: '/docs/python.md' },
];
let hitRow = null;
function fixture(save = async () => {}) {
  const s = vm.runInNewContext(`({${methods}})`, { SaveDocLinks: save, GetDocLinks: async () => docs, document: { elementFromPoint: () => hitRow } });
  Object.assign(s, { docLinks: [...docs], docLinkSearch: '', docLinksSaving: false,
    docLinkPointer: null, docLinkDragIdx: -1, docLinkDropIdx: -1, docLinkDropAfter: false,
    editingDocLinkIdx: -1, editingDocLinkForm: { path: '', label: '' },
    docLinkForm: { path: '', label: '' }, showSuccess() {}, showError(e) { this.error = e; }, askConfirm: async () => true });
  return s;
}
const labels = s => Array.from(s.docLinks, l => l.label);

test('search matches Chinese label and case-insensitive path, trims whitespace; retains original indices', () => {
  const s = fixture();
  s.docLinkSearch = ' 指南 ';
  assert.deepEqual(Array.from(s.filteredDocLinks, r => r.idx), [0, 2]);
  s.docLinkSearch = 'BACKEND.PDF';
  assert.equal(s.filteredDocLinks[0].idx, 1);
  s.docLinkSearch = 'absent'; assert.equal(s.filteredDocLinks.length, 0);
  s.docLinkSearch = ' '; assert.equal(s.filteredDocLinks.length, 4);
});

test('filtered edit changes the intended document and keeps draft visible until saved', async () => {
  const s = fixture(); s.docLinkSearch = 'Codex';
  s.startEditDocLink(s.filteredDocLinks[0].idx);
  s.docLinkSearch = 'absent'; assert.equal(s.filteredDocLinks[0].idx, 2);
  s.editingDocLinkForm = { path: '/new.md', label: '改名' };
  await s.saveEditDocLink();
  assert.equal(s.docLinks[2].label, '改名'); assert.equal(s.docLinks[0].label, docs[0].label);
  assert.equal(s.editingDocLinkIdx, -1);
});

test('filtered delete removes original index, cancellation preserves list', async () => {
  const s = fixture(); s.docLinkSearch = 'Codex';
  s.askConfirm = async () => false; await s.removeDocLink(2); assert.equal(s.docLinks.length, 4);
  s.askConfirm = async () => true; await s.removeDocLink(s.filteredDocLinks[0].idx);
  assert.deepEqual(labels(s), ['Claude 指南', 'Go 入门', 'Python 入门']);
});

test('drop before and after persist the whole list, including hidden items; duplicates remain distinct', async () => {
  let saved;
  const s = fixture(async links => { saved = links; }); s.docLinkSearch = '指南';
  await s.moveDocLink(2, 0);
  assert.deepEqual(labels(s), ['Codex 指南', 'Claude 指南', 'Go 入门', 'Python 入门']);
  assert.equal(saved.length, 4);
  await s.moveDocLink(0, 3, true);
  assert.deepEqual(labels(s), ['Claude 指南', 'Go 入门', 'Python 入门', 'Codex 指南']);
  s.docLinks = [docs[0], { ...docs[0] }, docs[1]];
  await s.moveDocLink(1, 2, true); assert.equal(s.docLinks[2].path, docs[0].path);
});

test('save failure preserves order and edit/add drafts', async () => {
  const s = fixture(async () => { throw new Error('disk full'); });
  await s.moveDocLink(0, 3, true); assert.deepEqual(labels(s), docs.map(l => l.label));
  assert.match(s.error, /disk full/); assert.equal(s.docLinksSaving, false);
  s.startEditDocLink(1); s.editingDocLinkForm.label = 'draft'; await s.saveEditDocLink();
  assert.equal(s.editingDocLinkIdx, 1); assert.equal(s.editingDocLinkForm.label, 'draft');
  s.cancelEditDocLink(); s.docLinkForm = { path: '/new.md', label: 'draft' }; await s.addDocLink();
  assert.equal(s.docLinkForm.label, 'draft'); assert.equal(s.docLinks.length, 4);
});

test('in-flight saves block conflicting mutations; no-op moves never write', async () => {
  let finish; let count = 0;
  const s = fixture(() => { count++; return new Promise(resolve => { finish = resolve; }); });
  await s.moveDocLink(0, 0); await s.moveDocLink(-1, 1); await s.moveDocLink(0, undefined);
  assert.equal(count, 0);
  const pending = s.moveDocLink(0, 3, true); assert.equal(s.docLinksSaving, true);
  await s.moveDocLink(1, 0); s.startEditDocLink(1); await s.removeDocLink(1); await s.loadDocLinks();
  assert.equal(count, 1); assert.equal(s.editingDocLinkIdx, -1);
  finish(); await pending; assert.equal(s.docLinksSaving, false);
});

test('pointer drag cancels safely, ignores clicks/outside drops, and saves valid drops', async () => {
  let count = 0; const s = fixture(async () => { count++; });
  const handle = { setPointerCapture() {}, hasPointerCapture: () => true, releasePointerCapture() {} };
  const start = { button: 0, pointerId: 1, clientX: 10, clientY: 10, currentTarget: handle };
  const move = { ...start, clientY: 45 };
  hitRow = { closest() { return this; }, dataset: { docIdx: '2' }, getBoundingClientRect: () => ({ top: 10, height: 40 }) };
  await s.dropDocLink(start); assert.equal(count, 0);
  s.startDocLinkDrag(start, 0); await s.dropDocLink(start); assert.equal(count, 0);
  s.startDocLinkDrag(start, 0); s.overDocLink(move); assert.equal(s.docLinkDropAfter, true);
  s.endDocLinkDrag(); assert.equal(count, 0);
  s.startDocLinkDrag(start, 0); s.overDocLink(move); await s.dropDocLink(move);
  assert.equal(count, 1); assert.equal(s.docLinkDropIdx, -1); assert.equal(s.docLinkPointer, null);
  assert.deepEqual(labels(s), ['Go 入门', 'Codex 指南', 'Claude 指南', 'Python 入门']);
  hitRow = null; s.startDocLinkDrag(start, 0); s.overDocLink(move); await s.dropDocLink(move);
  assert.equal(count, 1);
});
