# Hatch — Frontend Reference

Use this with `.claude/skills/hatch-frontend/SKILL.md` before editing `frontend/index.html`, `frontend/src/main.js`, or `frontend/src/style.css`.

## Alpine State Machine

```text
tab = 'scripts' | 'ai' | 'vpn'
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'skills' | 'mcp' | 'hooks' | 'tools' | 'prompts' | 'data'
vpnSubTab = 'overview' | 'config' | 'profiles' | 'backup' | 'enhance'

scripts + !selectedDir  -> directory cards + workspace-settings accordion
scripts + selectedDir   -> script list
ai                      -> AI sidebar layout
vpn                     -> VPN sidebar layout
editTarget != null      -> edit drawer overlay
```

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
| `mcpRecommendQuery / mcpRecommendBusy / mcpRecommendation / mcpSelectedRecommendationID` | MCP AI recommendation panel and selected draft candidate |
| `hooks / hooksStats / hooksSearch / hooksFilter / filteredHooks` | Hook list and filter UI |
| `hookEventSupport / hookForm / hookEditOpen / hookEditIsNew / hookSaving` | Hook event catalog + edit form state |
| `hookSyncBusy / hookTemplates / hookTemplateGalleryOpen / hookSuggestDesc / hookSuggestBusy` | Hook sync busy keys, template gallery, header AI draft bar |
| `hookAIInstruction / hookAIBusy` | In-editor AI optimize bar for refining an existing hook |
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

- Sticky 52px header: logo, tab nav (segmented control — `.tab-nav` is a single rounded container; `.tab.active` is a charcoal fill with a soft shadow; all three tabs carry an icon), tab-aware refresh button (`headerRefresh()` dispatches per `tab`), quick-AI provider button (tooltip shows active provider name + baseUrl), VPN status pill (`vpnPillTitle()` tooltip surfaces system-proxy/TUN/exit-IP), autostart toggle.
- Content max-width 1080px, centered, `padding: 24px 22px 44px`.
- `page-hero` bands use eyebrow, h1, subtitle, and stat cards.
- AI hero stat cards show AI status, active/default model, Skills, MCP, and Hooks; MCP/Hooks reuse existing overview state and responsive grid rules.
- Responsive breakpoint: 820px. Cards/forms become single column; edit drawer becomes full width.
- AI and VPN use the same left-sidebar layout: `.ai-layout`, `.ai-sidenav`, `.ai-content`.

## Feature Surfaces

- Script cards: accent border by type, language/name badges, AI fill, edit, log toggle, tags, service chips, preset buttons, footer controls, dark log panel.
- Edit drawer: all per-script behavior lives here, including env pairs, notify, schedule, requires, health URL, and pre-scripts.
- Workspace settings: scripts home accordion with shell/data paths, directories, exports, zshrc parser.
- AI: overview, connect/providers/traces, skills, MCP, hooks, tools/status, prompts, data/retention. Skills import can surface same-name content conflicts for Hatch-vs-source selection. MCP starts with a double-click operation manual, includes an AI recommendation panel for choosing a server/install method, separates local/Git/npm acquisition from Claude Code/Codex sync, shows MCP config values in plaintext for review, writes global target config, and shows CLI test results. Model shortcut panels share the `agent-*` config layout classes; the Goals toggle sits in the shared action row.
  - Overview is a live dashboard (`.ai-dashboard-grid` / `.ai-dash-card`): active provider, Claude/Codex install+usage, skills counts, MCP enabled/failed counts, hooks counts (total + Claude/Codex enabled), and recent trace errors, each clickable to its subtab. `loadOverview()` aggregates already-bound status methods (including `GetHooksStats`) and is called on the overview sidenav button and from `init()`.
  - Bulk actions (`.bulk-action-bar`) on skills and MCP: `bulkSkillSync` / `bulkSkillUnsyncAll` / `bulkMCPEnable` / `bulkTestMCP` iterate `filteredSkills` / `filteredMCPServers` serially.
  - `copyToClipboard(text, label)` is the shared clipboard helper (provider URL/model, MCP `renderMCPSnippet` Claude/Codex blocks, default prompts, trace blocks). `renderMCPSnippet(server, target)` mirrors backend `renderClaudeMCPServer`/`renderCodexMCPServer`.
  - `duplicateProvider(p)` opens the provider editor prefilled (baseUrl/model, blank apiKey, "副本" name).
  - Hooks (`aiSubTab==='hooks'`, `.ai-hooks-section`): registry of `command` hooks. Each `.hook-item` shows event badge + matcher + command preview and two `.skill-sync-toggle` switches (Claude/Codex); a switch gets `.sync-disabled` when `hookEventSupports(h.event, target)` is false. Editing uses a form (`.hook-editor-panel` / `.hook-form-grid`), not a markdown editor. The editor has an inline AI bar (`.se-ai-bar`, `hookAIInstruction` / `hookAIBusy`): `optimizeHookWithAI()` sends the current form via `SuggestHookEdit` to refine the existing hook in place. `📦 模板库` opens `.hook-modal` from `hookTemplates`; the header `✦ AI 起草` calls `suggestHook()` (`SuggestHookFromDescription`) to prefill a new hook. `toggleHookSync` updates the row in place and refreshes `hooksStats`; `saveHook` re-syncs already-enabled targets after `SaveHook`.
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
