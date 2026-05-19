# Hatch — Known Gotchas & Pico CSS Notes

## Known Patterns & Gotchas

- **New Go methods not in wailsjs**: Call as `window['go']['main']['App']['Method']()` until next `wails build` regenerates bindings.
- **`mergeOverride` non-zero only** — `Commands: []` in override does NOT clear auto-detected commands; use `DeleteScriptMeta` to fully reset.
- **`Args` nil check** — `ov.Args != nil` (not `len > 0`), so `[]ArgDef{}` correctly hides the args input.
- **`Presets`** — fully replaced on every `SaveScriptMeta` call.
- **`buildBlock` adds PATH line** — every exported dir var is auto-appended to `export PATH=`.
- **`ZshrcAnnotations` keys** — named entries use identifier; positional entries use `"type:lineNum"` (changes on file edit).
- **Service detection** requires both `start` AND `stop` in first 200 lines.
- **Login shell + explicit source** — scripts run as `/bin/zsh -l -c 'source ~/.zshrc; exec script args'`. Required because Wails GUI apps don't get interactive shells.
- **`Setpgid: true`** on services — `Stop()` sends SIGTERM to the script's own process group.
- **Corrupt store JSON** silently resets to empty — normal on first run.
- **`buildBlock` sorts keys** — alphabetical order for stable diffs.
- **`safeId(path)`** replaces `btoa(path)` — avoids encoding errors on non-Latin-1 paths.
- **AI key is plaintext** in `config.json` — never log or screenshot it.
- **AI suggestions are not auto-saved** — always populate UI for user to confirm.
- **OpenAI-compatible URL** — `BaseURL` accepts `.../v1` base or full `/chat/completions`; backend appends suffix only when missing.
- **CLI detection in GUI** — `exec.LookPath` only searches process PATH. Use `findCLI()` helper which falls back to `zsh -l -c "command -v <name>"` for tools in `~/.local/bin` etc.
- **`<template x-if>` with bare text** — Alpine.js `<template>` must wrap an element, not bare text. Use `x-text` on a `<p>` or `<span>` instead.
- **Alpine.js reactive batching in nested `x-for`** — setting a reactive key then immediately reading it inside a called method can silently fail. Use fully inline `@click` handler instead:
  ```html
  @click.stop="$event.preventDefault(); scriptArgs[s.path] = p.args; runScript(s)"
  ```

---

## Pico CSS Integration Notes

### CSS Load Order (critical)

`index.html` must NOT have `<link rel="stylesheet" href="/src/style.css" />`. Both load via JS in `main.js`:

```js
import '@picocss/pico/css/pico.min.css';  // first
import './style.css';                       // second — wins the cascade
```

If also linked in HTML, the `<link>` fires before JS, Pico overrides our rules.

### Primary Color Override

```css
--pico-primary:                  #c15f3c;
--pico-primary-background:       #c15f3c;
--pico-primary-border:           #c15f3c;
--pico-primary-hover:            #a94f32;
--pico-primary-hover-background: #a94f32;
--pico-primary-focus:            rgba(193,95,60,0.35);
--pico-primary-inverse:          #fffaf4;
```

### Semantic Element Reset

```css
header, nav, main, footer, section, article, aside { padding: 0; margin: 0; }
```
Do NOT use `display: revert` — breaks display types.

### `aria-busy` Button Color

```css
button[aria-busy="true"] { background: var(--primary); border-color: var(--primary); color: #fffaf4; }
button.secondary[aria-busy="true"], button.btn-refresh[aria-busy="true"] {
  background: var(--surface-3); border-color: var(--border); color: var(--text-muted);
}
```

### `role="button"` Reset

```css
[role="button"]:not(button):not(a):not(input):not(select):not(textarea) {
  background: var(--surface-2); color: var(--text);
  border: 1px solid var(--border); border-radius: var(--radius); box-shadow: var(--shadow-card);
}
```

---

## v1.1 – v1.3 Specific

### Wails Binding Regeneration
Methods added in `app_v12.go`, `app_v13.go`, `app_datadir.go`, `skills.go` (v1.3+) are **not** in `frontend/wailsjs/go/main/App.js` until you run `wails dev` or `wails build`. Frontend calls them via `window.go.main.App.MethodName(...)`. Don't try to import them at the top of `main.js` until bindings regenerate.

### Default-prompt drift
`GetDefaultPrompts()` in `app_datadir.go` is a **manual copy** of the system prompts hard-coded in the Suggest* methods inside `app_ai.go` (and `SuggestSkillContent` in `app_datadir.go`). When you change a default prompt, update **both places** or the UI will misrepresent what's running.

### Scheduler reload after SaveScriptMeta
`SaveScriptMeta` triggers `go a.ApplyAllSchedules()` so a saved `Schedule` field takes effect immediately. The 500ms startup delay before the first call is intentional — gives `ScanAll` time to populate `store` so the schedule loop sees current meta.

### Retention ticker tied to App ctx
`ApplyRetention` runs on a 1-hour `time.NewTicker` inside a goroutine started in `startup`. The goroutine watches `ctx.Done()` so closing the window stops it cleanly. **Do not** spawn the ticker from anywhere else — duplicate timers will double-prune.

### Run-history scope by basename
Run-history dirs use `filepath.Base(scriptPath)` as the grouping key. Two scripts named `deploy.sh` in different directories will share a single history bucket. Acceptable for v1.x; revisit if/when scripts collide for real.

### Pre-scripts don't emit
Pre-scripts in `RunWithPreScripts` execute with `nil` emitter — they only show as `[hatch] pre: <path>` lines in the **main** script's log stream, not as standalone runs. That's by design so the user sees one continuous log for the whole chain. Pre-scripts still archive to history individually.

### Skill save auto-resyncs to Claude / Codex
`saveSkillContent()` checks `skill.inClaude` / `skill.inCodex` **before** calling `SaveSkillContent`, then calls `SetSkillSync(name, target, true)` for each enabled target after the master save. This keeps the synced directories up to date without manual re-toggle. New skills (not yet synced anywhere) skip this step silently.

### Quick-AI popover provider sync
`switchQuickAIProvider()` calls `SaveAIProviders(providers, newActiveID)` and **also** copies the new provider's `baseUrl/apiKey/model` into `aiForm` so the `🔌 连接` page reflects the switch. If you change provider config independently, update both stores together.

### Inline-style rule for new features
v1.1+ rule: no inline `style="..."` on new feature markup. Extract to a class in the bottom section of `style.css`. Pre-existing inline styles remain (cleanup batched for a future pass).
