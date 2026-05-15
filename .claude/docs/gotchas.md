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
