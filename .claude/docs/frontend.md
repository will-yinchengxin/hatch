# Hatch — Frontend Reference

## Alpine State Machine (`main.js` → `hatchApp()`)

```
tab = 'scripts' | 'ai'   (default: 'ai' on launch; settings removed — accordion on scripts home)
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'skills' | 'tools' | 'prompts' | 'data'

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
| `skills` | `[]` from `GetSkills()` — sorted by modTime desc, SKILL.md-validated only |
| `skillsStats` | `{total, claude, codex}` |
| `skillSearch` | Search query string for filtering skills by name |
| `skillFilter` | `'all' \| 'claude' \| 'codex'` — active category filter tab |
| `filteredSkills` | Getter: `skills` filtered by `skillFilter` then `skillSearch` |
| `traceLog / traceLogDate` | Trace log viewer state |
| `dataPaths / dataPathsForm` | App data paths config |
| `zshrcEntries / zshrcFilter` | Parsed zshrc + filter |
| `autoLaunchEnabled` | Login autostart toggle |
| `wsShellOpen / wsDirsOpen / wsEnvOpen / wsZshrcOpen` | Workspace accordion states |
| `success / successTimer` | Success toast (3s auto-dismiss) |
| `error` | Error toast (6s auto-dismiss) |
| `historyTarget / historyRecords / historyOpenId / historyLogText / historyLoading` | v1.1 run-history modal |
| `editForm.envPairs / .notifyOnDone / .schedule / .requiresRaw / .healthUrl / .preScriptsRaw` | v1.1/v1.2/v1.3 edit-drawer fields |
| `promptTemplates / promptDefaults / promptDefaultsOpen / promptOps / promptSaving` | v1.2 Prompt sub-tab |
| `dataInventory / retention / retentionSaving` | v1.3 Data sub-tab |
| `aiProviders / activeProviderID / quickAIOpen / quickAIActiveID / quickAITraceID` | Header quick-AI popover |
| `skillAIInstruction / skillAIBusy` | v1.3 Skill ✦ AI optimize bar |
| `vpnSubTab` | `'overview' \| 'config' \| 'profiles' \| 'backup' \| 'enhance'` |
| `vpnStatus` | `VPNStatus` from `GetVPNStatus()` — loaded on `init()` + tab enter + manual refresh |
| `vpnCurrentIP / vpnIPLoading` | Exit IP card — refreshed separately via `refreshVPNCurrentIP()` |
| `vpnStatusRefreshing` | Hero refresh button busy state |
| `vpnTUNExpanded` | Toggle TUN card's excluded-CIDR expansion (double-click) |
| `vpnProfiles / vpnCurrentProfile` | All profiles + the current active one |
| `vpnEnhanceData / vpnEnhanceLoading` | Enhance tab: prepend rules, hosts, DNS fake-IP |
| `vpnProfileEditing / vpnProfileContent` | Inline profile YAML editor |
| `vpnBackups / vpnBackupDir / vpnBackupBusy / vpnRestoreBusy` | Backup tab state |
| `vpnMsg` | Inline message bar (success/error, auto-dismiss) |
| `vpnSettingsForm / vpnSettingsSaving` | App-dir + backup-dir settings |
| `vpnConfigFile / vpnConfigContent / vpnConfigSaving / vpnConfigError` | Config editor tab |

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

- Sticky header (52px, blur) — logo left · tab nav center · search + refresh + **`✦ 当前 provider ▼` quick-AI button** + **VPN status pill** + autostart toggle right
- Content: `max-width: 1080px`, centered, `padding: 24px 22px 44px`
- `page-hero` band: eyebrow + h1 + subtitle + stat row
- Responsive at 820px — cards/forms go single-column, drawer full-width

### AI Module Layout

```html
<main class="ai-main">
  <section class="page-hero ai-hero"> ... hero stats ... </section>
  <div class="ai-layout">
    <nav class="ai-sidenav"> 6 sub-tab buttons (overview/connect/skills/tools/prompts/data) </nav>
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

1. `openEdit(s)` populates `editForm` from script meta — including `envPairs`, `notifyOnDone`, `schedule`, `requiresRaw`, `healthUrl`, `preScriptsRaw`
2. User edits any field; the drawer is the **only** place to configure all per-script behavior
3. `saveEdit()` packages everything (including `env` map and `preScripts` array) → `SaveScriptMeta()` → `scanDir()` → re-applies scheduler
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

---

## v1.1 – v1.3 UI Surfaces

### Run-history modal (`historyTarget`)
- Wide modal (900px) reusing `.se-modal` + `.history-modal-wide`
- Each record: `.history-record` card with status mark, timestamp, args, expandable log `<pre>`
- Actions: `▶ 重跑` (reuses `runScript` with archived args), `📋 日志` (toggle full log), `✕` delete
- Backend: `ListRuns(path) / GetRunLog(path,id) / DeleteRun(path,id)` — dynamic bindings, not in wailsjs until rebuild

### Prompt sub-tab (`aiSubTab='prompts'`)
- `.prompt-card` per op (5 fixed ops in `promptOps`)
- Header: title + op key + `● 已覆盖`/`○ 使用默认` badge + `查看默认` + `↓ 以默认为模板`
- `.prompt-default-block` — read-only dark pre showing built-in default
- Backend: `GetDefaultPrompts()` returns canonical defaults; UI must keep `promptOps` list in sync with backend keys (`SuggestScriptMeta / SuggestScriptContent / SuggestZshrcAnnotation / SuggestZshrcOrganization / TestAIConfig / SuggestSkillContent`)

### Data sub-tab (`aiSubTab='data'`)
- `.retention-form` 4-col grid: trace days / runs days / per-name cap / save
- `.data-dir-card` per directory from `GetAppDataInventory()`
- `📂` calls `OpenFile(path)` (macOS `open`); `🗑` calls `CleanupDataDir(key)` only when `d.cleanable`
- `formatBytes(n)` / `formatTime(unix)` helpers in `main.js`

### Skills sub-tab filter bar
- `.skill-filter-bar` — flex row: segment-control tabs + search input
- `.skill-filter-tabs` / `.skill-filter-tab` — 全部 / Claude / Codex segment control; active tab uses `var(--primary)` text
- `.skill-search-input` — name search; uses `!important` overrides to suppress Pico's large form-element padding
- `filteredSkills` getter applies category filter first, then name substring match
- Backend `GetSkills()` skips dirs without `SKILL.md` and returns skills sorted by `modTime` desc (newest first)

### Skill editor `✦ AI` bar
- `.se-ai-bar.skill-ai-bar` mirrors the script-editor AI bar
- `optimizeSkillWithAI()` → `SuggestSkillContent(name, current, instruction)` — overwrites `skillEditContent`, user reviews then saves

### Header quick-AI popover (`quickAIOpen`)
- `.quick-ai-wrap` + absolute-positioned `.quick-ai-popover` (420px)
- Outside-click closes via `@click.outside`
- Model switcher binds to `aiProviders`; `switchQuickAIProvider()` calls `SaveAIProviders(providers, newActiveID)` and syncs `aiForm`
- Trace list shows newest 8; click to expand inline `.quick-ai-trace-pre` (user prompt + raw response)

---

### VPN Module Layout (`tab='vpn'`)

Same left-sidebar layout as AI (`ai-layout` + `ai-sidenav` + `ai-content`).

```html
<main class="ai-main">
  <section class="page-hero ai-hero vpn-hero"> <!-- position: relative for refresh btn -->
    <button class="vpn-hero-refresh-btn">↻</button>   <!-- absolute top-right -->
    <div class="vpn-hero-left"> eyebrow + h1 + subtitle </div>
    <div class="vpn-hero-right">
      <div class="hero-stats vpn-hero-stats">
        <!-- 系统代理 card: vpn-status-card + vpn-stat-on/off -->
        <!-- TUN card: @dblclick toggles vpnTUNExpanded, shows tunInterface + CIDR chips -->
        <!-- 出口 IP card: vpn-ip-card -->
        <!-- 代理节点 card: vpn-node-card, iterates vpnStatus.proxyGroups -->
      </div>
      <div class="vpn-hero-traffic"> upload · download · total · expire </div>
    </div>
  </section>
  <div class="ai-layout">
    <nav class="ai-sidenav"> 概览 / 配置编辑 / Profiles / 备份恢复 / ✦ 自定义规则 </nav>
    <div class="ai-content"> ... </div>
  </div>
</main>
```

**Header VPN pill** (`.vpn-hdr-pill`): shows two 5px dots (system proxy + TUN state) + truncated node name. Loaded via `GetVPNStatus()` during `init()`. Clicking jumps to VPN tab.

**TUN card double-click**: `vpnTUNExpanded` toggles `.vpn-tun-cidrs` section showing CIDR chips from `vpnStatus.tunExcludedCidrs`.

**Proxy groups card** (`.vpn-node-card`): iterates `vpnStatus.proxyGroups` (live from Clash API), sorted real-node groups first, shows up to 3 rows with "+N 个" overflow.

**Enhance tab** (`vpnSubTab='enhance'`): calls `GetVPNEnhanceData()` which reads `merge` + `rules` type profiles. Shows: prepend rules table (type/value/target color-coded), hosts override table, DNS fake-IP filter chips.

### VPN CSS Classes

| Prefix / Class | Purpose |
|---|---|
| `.vpn-hdr-pill` / `.vpn-hdr-pill-on` | Header status pill (green border when proxy on) |
| `.vpn-hdr-dots` / `.vpn-hdr-dot` | Two 5px indicator dots in pill |
| `.vpn-hero-refresh-btn` | Absolute-positioned ↻ in hero top-right |
| `.vpn-status-card` / `.vpn-stat-on` / `.vpn-stat-off` | Status cards (系统代理 + TUN) |
| `.vpn-status-dot` / `.vpn-dot-on` / `.vpn-dot-off` | Animated green / gray dot in status cards |
| `.vpn-tun-iface` | Monospace interface+IP line below TUN status |
| `.vpn-tun-cidrs` / `.vpn-cidr-tag` | Expanded CIDR chip list |
| `.vpn-node-card` / `.vpn-proxy-groups` / `.vpn-proxy-group-row` | Proxy groups list card |
| `.vpn-pg-label` / `.vpn-pg-node` | Group name tag + selected node text |
| `.vpn-enhance-table` / `.vpn-enhance-row` / `.vpn-enhance-thead` | Enhance tab rule tables |
| `.vpn-rule-type-tag` / `.vpn-rule-value` / `.vpn-rule-target` | Rule cell styling |
| `.vpn-target-direct` / `.vpn-target-proxy` | DIRECT=green, proxy=clay |
| `.vpn-enhance-chips` / `.vpn-enhance-chip` | DNS fake-IP filter chip list |

---

## CSS Class Conventions (v1.1+)

All new feature CSS is appended to `style.css` under the `v1.1 – v1.3 additions` section. **Do not add inline `style=` for any new feature** — extract to a named class. Existing inline styles are pre-existing and left alone for now.

Class prefixes by feature:
- `history-*` — run history modal
- `prompt-*` — Prompt sub-tab cards
- `data-dir-*` / `retention-*` — Data sub-tab
- `quick-ai-*` — header popover
- `skill-filter-*` — Skills filter bar (segment tabs + search)
- `field-hint`, `ok-mark`, `err-mark` — shared small atoms

