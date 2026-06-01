# Hatch — Frontend Reference

Use this with `.codex/skills/hatch-frontend/SKILL.md` before editing `frontend/index.html`, `frontend/src/main.js`, or `frontend/src/style.css`.

## Alpine State Machine

```text
tab = 'scripts' | 'ai' | 'vpn'
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'skills' | 'mcp' | 'tools' | 'prompts' | 'data'
vpnSubTab = 'overview' | 'config' | 'profiles' | 'backup' | 'enhance'

scripts + !selectedDir  -> directory cards + workspace-settings accordion
scripts + selectedDir   -> script list
ai                      -> AI sidebar layout
vpn                     -> VPN sidebar layout
editTarget != null      -> edit drawer overlay
```

`?aiSubTab=<overview|connect|skills|mcp|tools|prompts|data>` deep-links directly into an AI subview for local verification. `?aiSubTab=mcp&mcpManual=1` opens the MCP operation manual modal.

## Key State

| Field | Purpose |
|---|---|
| `allScripts / currentScripts` | Scan results and current directory scripts |
| `statuses / runningSet / scriptLogs / scriptExitCodes / logTarget` | Runtime script state |
| `editTarget / editForm / editSaving` | Edit drawer state |
| `aiForm / aiProviders / activeProviderID / quickAIOpen / quickAITraceID` | Provider config and header quick-AI popover |
| `aiSubTab / traceLog / traceLogDate / promptTemplates / promptDefaults` | AI page subviews |
| AI tool model config state | Model/effort shortcuts; the CLI config state also tracks `goalsEnabled` |
| `skills / skillsStats / skillSearch / skillFilter / filteredSkills` | Skill list and filter UI |
| `skillImportResult / skillConflictBusy` | Skill import results and per-conflict action state |
| `skillAIInstruction / skillAIBusy` | Skill AI optimize bar |
| `mcpSettings / mcpServers / mcpTemplates / filteredMCPServers` | MCP registry, templates, filters, and config directory |
| `mcpForm / mcpSyncBusy / mcpTestBusy / mcpImportBusy / mcpManualOpen` | MCP editor, target sync, connection checks, separated source-install state, and operation-manual modal |
| `dataPaths / dataPathsForm / dataInventory / retention` | Data paths, inventory, retention |
| `zshrcEntries / zshrcFilter` | Parsed zshrc |
| `autoLaunchEnabled` | Login autostart |
| `wsShellOpen / wsDirsOpen / wsEnvOpen / wsZshrcOpen` | Workspace accordion |
| `historyTarget / historyRecords / historyOpenId / historyLogText` | Run-history modal |
| `vpnStatus / vpnCurrentIP / vpnProfiles / vpnCurrentProfile` | VPN status and profile state |
| `vpnEnhanceData / vpnProfileEditing / vpnBackups / vpnSettingsForm` | VPN enhance, edit, backup, settings state |
| `success / successTimer / error` | Toasts |

Call `showSuccess()` after write operations. `showError()` is for failures.

## UI Direction

Warm off-white surfaces, charcoal text, muted clay accent. Do not introduce a cold blue or neon dark theme unless explicitly requested.

Core tokens:

```css
--surface-0: #f6f0e8;
--surface:   #fbf7f0;
--surface-2: #fffdf8;
--surface-3: #f1e8dc;
--surface-4: #e7dacb;
--primary:   #c15f3c;
--text:      #2b211b;
--green:     #4f7b58;
--red:       #b94c3f;
--yellow:    #a7772e;
--cyan:      #4f6f7b;
--purple:    #7b5c7d;
```

## Layout

- Sticky 52px header: logo, tab nav, search/refresh, quick-AI provider button, VPN status pill, autostart toggle.
- Content max-width 1080px, centered, `padding: 24px 22px 44px`.
- `page-hero` bands use eyebrow, h1, subtitle, and stat cards.
- Responsive breakpoint: 820px. Cards/forms become single column; edit drawer becomes full width.
- AI and VPN use the same left-sidebar layout: `.ai-layout`, `.ai-sidenav`, `.ai-content`.

## Feature Surfaces

- Script cards: accent border by type, language/name badges, AI fill, edit, log toggle, tags, service chips, preset buttons, footer controls, dark log panel.
- Edit drawer: all per-script behavior lives here, including env pairs, notify, schedule, requires, health URL, and pre-scripts.
- Workspace settings: scripts home accordion with shell/data paths, directories, exports, zshrc parser.
- AI: overview, connect/providers/traces, skills, MCP, tools/status, prompts, data/retention. Skills import can surface same-name content conflicts for Hatch-vs-source selection. MCP starts with a double-click operation manual, separates local/Git/npm acquisition from Claude Code/Codex sync, shows MCP config values in plaintext for review, writes global target config, and shows CLI test results. Model shortcut panels share the `agent-*` config layout classes; the Goals toggle sits in the shared action row.
- VPN: overview/status hero, config editor, profiles editor, backup/restore, enhance rules/hosts/DNS view.

## CSS And Alpine Pitfalls

- `main.js` imports Pico first and `style.css` second. Do not link `/src/style.css` in `index.html`.
- Reset Pico semantic spacing with `header, nav, main, footer, section, article, aside { padding: 0; margin: 0; }`.
- Override Pico `--pico-primary-*` variables to clay colors.
- Override `button[aria-busy="true"]` so busy buttons do not turn Pico blue.
- Reset non-button `[role="button"]` cards to Hatch card styles.
- `setupEvents()` must call `EventsOff` before `EventsOn` to avoid HMR duplicates.
- Nested Alpine `x-for` preset clicks should use inline handlers:

```html
@click.stop="$event.preventDefault(); scriptArgs[s.path] = p.args; runScript(s)"
```

- Newly added Go methods can be called through `window['go']['main']['App']` until Wails bindings regenerate.
- New feature CSS belongs in named classes, not inline `style=`.
