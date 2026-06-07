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
tab = 'scripts' | 'ai' | 'vpn'
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'skills' | 'mcp' | 'hooks' | 'tools' | 'prompts' | 'data'
vpnSubTab = 'overview' | 'config' | 'profiles' | 'backup' | 'enhance'
```

- `scripts + !selectedDir`: 目录卡片 + workspace settings accordion。
- `scripts + selectedDir`: 脚本列表。
- `ai`: AI 侧边栏布局，含连接、skills、mcp、hooks、tools、prompts、data。
- `vpn`: VPN 侧边栏布局，含概览、配置、profiles、备份、自定义规则。
- `editTarget != null`: 右侧编辑抽屉。

## 关键状态

- 脚本运行：`statuses`、`runningSet`、`scriptLogs`、`scriptExitCodes`、`logTarget`。
- 编辑抽屉：`editTarget`、`editForm`、`editSaving`，所有 per-script 行为都从这里配置。
- AI：`aiForm`、`aiProviders`、`activeProviderID`、`quickAIOpen`、`traceLog`、`promptTemplates`、AI tool model config state。
- Skills：`skills`、`skillsStats`、`skillSearch`、`skillFilter`、`filteredSkills`、`skillAIInstruction`、`skillImportResult`、`skillConflictBusy`。
- MCP：`mcpSettings`、`mcpServers`、`mcpTemplates`、`mcpForm`、`mcpSyncBusy`、`mcpTestBusy`、`mcpImportBusy`、`mcpManualOpen`、`filteredMCPServers`。MCP 页面分两段：获取/安装 与 配置同步；配置值默认明文展示；批量启用/测试迭代 `filteredMCPServers`。
- Hooks：`hooks`、`hooksStats`、`hooksSearch`、`hooksFilter`、`filteredHooks`、`hookEventSupport`、`hookForm`/`hookEditOpen`、`hookSyncBusy`、`hookTemplates`、`hookSuggestDesc`（头部起草）、`hookAIInstruction`/`hookAIBusy`（编辑器内 AI 优化）。`hookEventSupports(event,target)` 控制同步开关置灰；同步开关复用 `.skill-sync-toggle`，编辑用表单（`.hook-editor-panel`）而非 markdown 编辑器；`SuggestHookFromDescription` 起草新钩子，`SuggestHookEdit` 改进现有钩子。
- Data：`dataPaths`、`dataInventory`、`retention`。
- VPN：`vpnStatus`、`vpnCurrentIP`、`vpnProfiles`、`vpnEnhanceData`、`vpnBackups`、`vpnSettingsForm`。
- Workspace：`wsShellOpen`、`wsDirsOpen`、`wsEnvOpen`、`wsZshrcOpen`。

## CSS 规则

- Pico semantic spacing reset:

```css
header, nav, main, footer, section, article, aside { padding: 0; margin: 0; }
```

- 覆盖 `--pico-primary-*` 到 Hatch clay 色。
- 覆盖 `button[aria-busy="true"]`，否则 Pico 会变蓝。
- 重置非按钮 `[role="button"]` 卡片样式。
- AI/VPN 使用 `.ai-layout`、`.ai-sidenav`、`.ai-content` 共享布局；AI Tools 快捷配置共用 `agent-*` 布局类，Goals 开关放在共享动作行内。
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
