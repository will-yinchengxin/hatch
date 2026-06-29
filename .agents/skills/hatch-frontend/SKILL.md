---
name: hatch-frontend
version: 1.1.0
description: "Hatch 前端参考：修改 frontend/index.html、frontend/src/main.js 或 frontend/src/style.css 前必读。覆盖 Alpine 状态机、Pico 覆盖、UI token、AI/VPN/Skills/Data 页面和常见坑。"
---

# Hatch 前端工作流

使用条件：任何修改 `frontend/index.html`、`frontend/src/main.js`、`frontend/src/style.css`、前端交互、页面布局、CSS class、Alpine 状态字段、Wails 前端调用方式的任务。

## 先读

- `.codex/reference/frontend.md`
- `.codex/reference/gotchas.md`
- 如果涉及 Go 方法调用，再读 `.codex/reference/bound-methods.md`
- 如果涉及数据字段，再读 `.codex/reference/data-model.md`

## 约束

- Hatch 前端使用 Alpine.js v3 + Pico CSS v2 + Vite，不使用 Tailwind。
- `main.js` 必须先 import Pico，再 import `./style.css`。
- `index.html` 不能添加 `<link rel="stylesheet" href="/src/style.css" />`。
- `frontend/wailsjs/**` 禁止手动编辑。
- 新 Go 方法在 Wails 重新生成前用 `window['go']['main']['App']['MethodName'](...)` 调用。
- `setupEvents()` 注册事件前必须先 `EventsOff`，避免 HMR 重复监听。
- 新功能不要写 inline `style="..."`；添加具名 CSS class。
- 维持 warm off-white + charcoal + clay accent 方向，不要改回冷蓝或霓虹暗色主题。

## 当前页面模型

```text
tab = 'scripts' | 'ai' | 'vpn' | 'memo'
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'tools' | 'workspace' | 'skills' | 'mcp' | 'hooks' | 'docs' | 'prompts' | 'data'
vpnSubTab = 'overview' | 'config' | 'profiles' | 'backup' | 'enhance'
memoSubTab = 'today' | 'all' | 'words' | 'review' | 'storage'
```

- `scripts + !selectedDir`: 目录卡片 + workspace settings accordion。
- `scripts + selectedDir`: 脚本列表。
- `ai`: AI 侧边栏顺序：概览 → 连接 → **助手**（tools，🤖 Claude Code/Codex 状态与配置）→ Skills → MCP → Hooks → 文档 → **提示词**（prompts，system prompt 覆盖）→ 数据。
- AI hero 顶部统计卡包含 AI 状态、模型、Skills、MCP、Hooks；保持同一套 `hero-stat-*` 卡片样式和响应式 grid。
- `vpn`: VPN 侧边栏布局，含体检、备份、自定义规则，以及高级配置/Profile 原文编辑；TUN 排除网段在 hero 卡片内紧凑常驻展示，不使用隐藏式双击交互。
- `memo`: 备忘录侧边栏布局，含今日、全部、单词、复盘、存储；提醒只在 Hatch 运行时触发，附件导入到托管目录后展示/清理；单词发音优先音频 URL，缺失或播放失败时调用 `SpeakVocabWord` 走 macOS 系统发音。
- `editTarget != null`: 右侧编辑抽屉。
- `?aiSubTab=mcp` 等 query 可直接打开 AI 子页，`?aiSubTab=mcp&mcpManual=1` 可打开 MCP 手册弹窗，便于本地截图验证。

## 关键状态

- 脚本运行：`statuses`、`runningSet`、`scriptLogs`、`scriptExitCodes`、`logTarget`。
- 编辑抽屉：`editTarget`、`editForm`、`editSaving`，所有 per-script 行为都从这里配置。
- AI：`aiForm`、`aiProviders`、`activeProviderID`、`quickAIOpen`、`traceLog`、`promptTemplates`、AI tool model config state。
- AI 工作区：`workspaceProfiles`、`workspaceSnapshots`、`workspaceHealth`、`workspacePreview`、`workspaceForm`。应用只能走后端预检与确认，不在前端直接同步资源。
- Skills：`skills`、`skillsStats`、`skillSearch`、`skillFilter`、`filteredSkills`、`skillAIInstruction`、`skillImportResult`、`skillConflictBusy`。
- MCP：`mcpSettings`、`mcpServers`、`mcpTemplates`、`mcpForm`、`mcpSyncBusy`、`mcpTestBusy`、`mcpImportBusy`、`mcpManualOpen`、`mcpRecommendQuery`、`mcpRecommendation`、`mcpSelectedRecommendationID`、`filteredMCPServers`。MCP 页面包含 AI 推荐选型、获取/安装、配置同步；推荐候选只填入编辑器，不自动保存；配置值默认明文展示；批量启用/测试迭代 `filteredMCPServers`。
- Hooks：`hooks`、`hooksStats`、`hooksSearch`、`hooksFilter`、`filteredHooks`、`hookEventSupport`、`hookForm`/`hookEditOpen`、`hookSyncBusy`、`hookTemplates`、`hookSuggestDesc`（头部起草）、`hookAIInstruction`/`hookAIBusy`（编辑器内 AI 优化）。`hookEventSupports(event,target)` 控制同步开关置灰；同步开关复用 `.skill-sync-toggle`，编辑用表单（`.hook-editor-panel`）而非 markdown 编辑器；`SuggestHookFromDescription` 起草新钩子，`SuggestHookEdit` 改进现有钩子。
- Data：`dataPaths`、`dataInventory`、`retention`。
- VPN：`vpnStatus`、`vpnCurrentIP`、`vpnHealthReport`、`vpnPathValidation`、`vpnProfiles`、`vpnEnhanceData`、`vpnBackups`、`vpnBackupPreview`、`vpnSettingsForm`。
- Memo：`memoDashboard`、`memos`、`memoForm`、`memoSettings`、`memoPreviewCache`、`vocabEntries`、`vocabForm`、`reviewReveal`。
- Workspace：`wsShellOpen`、`wsDirsOpen`、`wsEnvOpen`、`wsZshrcOpen`。

## CSS 规则

- Pico semantic spacing reset:

```css
header, nav, main, footer, section, article, aside { padding: 0; margin: 0; }
```

- 覆盖 `--pico-primary-*` 到 Hatch clay 色。
- 覆盖 `button[aria-busy="true"]`，否则 Pico 会变蓝。
- 重置非按钮 `[role="button"]` 卡片样式。
- AI/VPN/Memo 使用 `.ai-layout`、`.ai-sidenav`、`.ai-content` 共享布局；AI Docs 是 MCP/Hooks 平级侧栏页，概览应提供文档数量摘要和跳转；AI Tools 快捷配置共用 `agent-*` 布局类，Claude Code/Codex 推荐组合区块需对齐复用 `.cc-rec-*` chip 结构，Goals 开关放在共享动作行内。
- 新 v1.1+ 功能类放在 `style.css` 末尾对应分区，保持命名可读。

## Alpine 注意事项

嵌套 `x-for` 的 preset click 不要通过方法“先 set 再读”，用完全 inline handler：

```html
@click.stop="$event.preventDefault(); scriptArgs[s.path] = p.args; runScript(s)"
```

`<template x-if>` 内必须包元素，不能直接放裸文本。

## 完成前检查

- 搜索是否误加了 `/src/style.css` link。
- 搜索新增 `style="`，确认没有为新功能添加 inline style。
- 如果改了 bound 方法调用，确认动态调用或生成绑定路径正确。
- 如果改了事件监听，确认不会 HMR 重复注册。
- 需要视觉验证时启动 dev server 并用浏览器检查桌面和窄屏布局。
