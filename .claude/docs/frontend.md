# Hatch — Frontend Reference

## Alpine State Machine (`main.js` → `hatchApp()`)

```
tab = 'scripts' | 'ai'   (settings removed — accordion on scripts home)
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'skills' | 'tools'

scripts + !selectedDir  →  Directory cards + workspace-settings accordion
scripts + selectedDir   →  Script list
ai                      →  AI left-sidebar layout
editTarget != null      →  Edit drawer overlay
```

### Key State Fields

| Field | Purpose |
|---|---|
| `allScripts` | `ScanAll()` result, used for dir counters |
| `currentScripts` | `GetScriptsByDir()`, shown in script list |
| `statuses` | `{path: "running"|"stopped"}`, polled every 3s |
| `runningSet` | In-progress oneshot paths |
| `scriptLogs` | `{path: [{line,stream}]}`, capped at 500 lines |
| `scriptExitCodes` | `{path: exitCode}` |
| `logTarget` | Which script's log panel is open |
| `editTarget / editForm / editSaving` | Edit drawer state |
| `aiForm` | `{enabled, baseUrl, apiKey, model}` |
| `aiSubTab` | Active AI sub-tab |
| `claudeStatus / claudeStatusLoading` | Claude Code status card |
| `codexStatus / codexStatusLoading` | Codex status card |
| `skills` | `[]` from `ListSkills()` |
| `skillsStats` | `{total, claude, codex}` |
| `traceLog / traceLogDate` | Trace log viewer state |
| `dataPaths / dataPathsForm` | App data paths config |
| `zshrcEntries / zshrcFilter` | Parsed zshrc + filter |
| `autoLaunchEnabled` | Login autostart toggle |
| `wsShellOpen / wsDirsOpen / wsEnvOpen / wsZshrcOpen` | Workspace accordion states |
| `success / successTimer` | Success toast (3s auto-dismiss) |
| `error` | Error toast (6s auto-dismiss) |

### Toast System

```js
showSuccess(msg)  // green toast, 3s
showError(msg)    // red toast, 6s (already existed)
```

Call `showSuccess()` after every write operation so users see confirmation.

---

## UI Design Tokens (`style.css :root`)

```
--surface-0: #f6f0e8   warm page bg (body)
--surface:   #fbf7f0   base bg
--surface-2: #fffdf8   cards / panels
--surface-3: #f1e8dc   inputs / low hover
--surface-4: #e7dacb   elevated hover

--primary:   #c15f3c   clay accent
--text:      #2b211b   charcoal text
--text-muted: #8a7060
--border:    #e0d5c8
--green:     #4f7b58
--red:       #b94c3f
--yellow:    #a7772e
--cyan:      #4f6f7b
--purple:    #7b5c7d
```

Warm off-white, charcoal text, muted clay accent. **No neon/cold-blue dark theme.**

---

## Layout Structure

- Sticky header (52px, blur) — logo left · tab nav center · search+refresh right
- Content: `max-width: 1080px`, centered, `padding: 24px 22px 44px`
- `page-hero` band: eyebrow + h1 + subtitle + stat row
- Responsive at 820px — cards/forms go single-column, drawer full-width

### AI Module Layout

```html
<main class="ai-main">
  <section class="page-hero ai-hero"> ... hero stats ... </section>
  <div class="ai-layout">
    <nav class="ai-sidenav"> 4 sub-tab buttons </nav>
    <div class="ai-content"> sub-tab content panels </div>
  </div>
</main>
```

### Hero Stat Cards (AI hero)

`.hero-stat-card` — unified card style (border, border-radius: 14px, min-height: 72px).
`.hero-stat-card--skills` — purple border (`#9b6bb5`), with total + breakdown layout.
Regular cards show label (small caps, muted) + large value.

---

## Script Card Anatomy

- Left accent border: clay for oneshot, ochre for service
- Header: language icon + name + badges + `✦ 智能填写` + `✏ 编辑` + log toggle
- Tags (clay chips) + service command chips
- Preset row: `🔖 Name` ochre pill buttons
- Footer (oneshot): args input + `▶ 运行`
- Footer (service): status dot + start/stop/restart
- Log panel: dark terminal, auto-opens on run

---

## Edit Drawer Flow

1. `openEdit(s)` populates `editForm` from script meta
2. User edits name/desc/tags/type/commands/pidFile/needsArgs/presets
3. `saveEdit()` → `SaveScriptMeta()` → `scanDir()` → close
4. `resetEdit()` → `DeleteScriptMeta()` → auto-detection resumes
5. `✦ 智能填写` → `SuggestScriptMeta()` → fills form → user saves manually

---

## Settings (now workspace-settings accordion on scripts home)

`.workspace-settings` > `.ws-section` accordion with `.ws-section-header` toggle.

Sections (wsShellOpen / wsDirsOpen / wsEnvOpen / wsZshrcOpen):
1. ⚙ 环境与路径配置 — shell status, autostart, data paths (trace log dir, skills dir)
2. 📁 目录管理 — add/remove dirs with export
3. 🔑 环境变量 — managed block preview + edit
4. 📄 Shell 配置解析 — parsed zshrc with annotations and AI notes

---

## Zshrc Parser Entry Types

| Type | Key |
|---|---|
| `function` | function name |
| `alias` | alias name |
| `export` | variable name |
| `path` | `"PATH"` |
| `source` | `"source:N"` |
| `section` | `"comment:N"` |
| `other` | `"other:N"` |

Positional keys (`source:N`, `other:N`, `comment:N`) change when the file is edited.
