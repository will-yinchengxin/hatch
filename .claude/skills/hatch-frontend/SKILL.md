---
name: hatch-frontend
version: 1.0.0
description: "Hatch 前端参考：目录结构、设计 token、页面视图组成、组件 class 规范、Alpine.js 状态机、常见坑。修改 Hatch 前端（style.css / index.html / main.js）前必读。"
---

# Hatch 前端参考

Hatch 是一个 macOS 脚本管理仪表盘，前端基于 **Alpine.js v3 + Pico CSS v2 + Vite**，不使用 Tailwind。

---

## 目录结构

```
frontend/
├── index.html                  # 唯一 HTML 文件，所有视图在此通过 x-show 切换
├── vite.config.js              # Vite 配置（server.open: true）
├── src/
│   ├── main.js                 # Alpine 组件 hatchApp()，所有状态 + 方法，Alpine.start() 在末尾
│   ├── style.css               # 全部自定义样式（先 Pico 后 style.css 覆盖）
│   └── assets/
│       ├── fonts/              # 字体文件
│       └── images/
│           └── logo.png        # 应用 logo（28×28，圆角 7px）
└── wailsjs/                    # Wails 自动生成，禁止手动编辑
    ├── go/main/App.js          # Go 方法的 JS 绑定
    ├── go/main/App.d.ts
    ├── go/models.ts
    └── runtime/runtime.js      # EventsOn / EventsOff / EventsEmit 等
```

### CSS 加载顺序（关键）

`main.js` 开头：
```js
import '@picocss/pico/css/pico.min.css';  // 先 Pico
import './style.css';                       // 后 custom，覆盖 Pico
```

`index.html` 中**不能**有 `<link rel="stylesheet" href="/src/style.css" />`，否则 Pico 会后来居上覆盖自定义样式。

---

## 设计 Token（style.css :root）

### 颜色

| Token | 值 | 用途 |
|---|---|---|
| `--primary` | `#c15f3c` | 主色调，clay 橙红 |
| `--primary-hover` | `#a94f32` | 主色 hover |
| `--primary-dim` | `rgba(193,95,60,0.1)` | 主色淡底 |
| `--primary-glow` | `rgba(193,95,60,0.22)` | 主色光晕 |
| `--surface-0` | `#f6f0e8` | 最深暖色底（body bg） |
| `--surface` | `#fbf7f0` | 基础底色 |
| `--surface-2` | `#fffdf8` | 卡片 / 面板 |
| `--surface-3` | `#f1e8dc` | 输入框 / hover 低阶 |
| `--surface-4` | `#e7dacb` | 更高阶 hover |
| `--border` | `#ded0c0` | 普通边框 |
| `--border-light` | `#c9b5a1` | hover 时边框 |
| `--text` | `#2b211b` | 主文本（炭灰） |
| `--text-muted` | `#6f6258` | 次要文本 |
| `--text-dim` | `#9d8e80` | 更淡文本 / placeholder |
| `--green` | `#4f7b58` | 成功 / service running |
| `--red` | `#b94c3f` | 错误 / stop |
| `--yellow` | `#a7772e` | service 类型 / restart |
| `--cyan` | `#4f6f7b` | source 类型 / cmd-tag |
| `--purple` | `#7b5c7d` | function 类型 |

### 类型 accent 颜色

| Token | 颜色 |
|---|---|
| `--accent-oneshot` | primary（clay） |
| `--accent-service` | green |
| `--accent-function` | purple |
| `--accent-alias` | yellow |
| `--accent-path` | red |
| `--accent-export` | green |
| `--accent-source` | cyan |
| `--accent-other` | text-dim |

### 圆角 / 阴影 / 尺寸

```css
--radius:    10px;
--radius-sm: 6px;
--radius-xs: 4px;

--control-height:    34px;
--control-height-sm: 30px;

--font-mono: 'SF Mono', 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;

--shadow-sm:   0 1px 2px rgba(43,33,27,0.08);
--shadow-md:   0 18px 48px rgba(43,33,27,0.13);
--shadow-card: 0 10px 28px rgba(43,33,27,0.08);
--shadow-glow: 0 0 0 2px var(--primary-glow);
```

---

## 页面视图结构

所有视图共用同一个 `<body x-data="hatchApp()" x-init="init()">` Alpine 组件，通过 `x-show` 切换。

```
header.app-header（sticky 52px，毛玻璃）
  ├── .header-brand
  │   ├── .btn-back（仅在 selectedDir 时显示）
  │   ├── .logo（img.logo-img + .logo-text）
  │   └── .breadcrumb（仅在 selectedDir 时，x-if）
  ├── nav.tab-nav（居中，3 个 tab）
  └── .header-right（搜索框 + 刷新 + 自启开关）

.error-toast（fixed 右下角，错误提示）

main（max-width: 1080px, centered, padding: 24px 22px 44px）
  ├── 脚本首页（tab=scripts && !selectedDir）
  │   ├── .page-hero（eyebrow + h1 + subtitle + hero-stats）
  │   └── .dir-card-list（2列网格）
  │       └── .dir-card × N
  │
  ├── 脚本目录内（tab=scripts && selectedDir）
  │   ├── .directory-toolbar（同 page-hero 样式）
  │   └── .script-list（flex 列）
  │       └── article.script-card × N
  │           ├── .card-header（.card-title + .card-actions）
  │           ├── .card-desc
  │           ├── .cmd-tags.meta-tags（tag chips）
  │           ├── .cmd-tags（service sub-command chips）
  │           ├── .preset-row（preset 快捷按钮）
  │           ├── .card-footer（oneshot: args input + 运行按钮）
  │           ├── .card-footer.service-footer（service: 状态 + 控制按钮）
  │           └── .log-panel（内联日志，dark terminal）
  │
  ├── 脚本首页工作区设置（.workspace-settings，accordion）
  │   ├── .ws-section（wsShellOpen）— ⚙ 环境与路径配置
  │   ├── .ws-section（wsDirsOpen）— 📁 目录管理
  │   ├── .ws-section（wsEnvOpen）— 🔑 环境变量
  │   └── .ws-section（wsZshrcOpen）— 📄 Shell 配置解析
  │
  └── AI 页（tab=ai && !selectedDir）
      ├── .page-hero.ai-hero（+ .hero-stat-card × 3）
      └── .ai-layout
          ├── nav.ai-sidenav（4 个 sub-tab 按钮）
          └── .ai-content（aiSubTab 切换）
              ├── 概览（overview）— 能力介绍
              ├── 连接（connect）— 模型配置 + trace log 查看器
              ├── Skills — 列表 / 同步 / 创建
              └── 工具（tools）— .ai-usage-row（claude card + codex card）

.overlay（fixed 全屏遮罩，右侧 edit drawer 或居中 modal）
  ├── .edit-drawer（460px，右侧滑入）
  │   ├── .drawer-header
  │   ├── .drawer-path
  │   ├── .drawer-body（滚动区）
  │   └── .drawer-footer
  └── .modal（居中，快速导出）
```

---

## 核心 CSS 类速查

### 布局类

| 类 | 用途 |
|---|---|
| `.page-hero` / `.directory-toolbar` | 页面顶部信息条，渐变卡片 |
| `.settings-section` | 设置页各区块，统一白卡样式 |
| `.section-header-row` | 区块标题行（标题 + 右侧操作） |
| `.overlay` | 全屏遮罩（drawer / modal 父容器） |
| `.export-actions` | 操作按钮行（flex + gap） |

### 按钮类

| 类 | 外观 |
|---|---|
| `button`（默认） | surface-2 背景，border，13px |
| `button.secondary` | text-muted 颜色 |
| `button.active` | primary 填充 |
| `.run-btn` | primary 填充，运行脚本 |
| `.btn-start` | 绿色透明底 |
| `.btn-stop` | 红色透明底 |
| `.btn-restart` | text-muted，hover 黄色 |
| `.btn-back` | 圆角 pill，返回 |
| `.btn-refresh` | 圆角 pill，刷新 |
| `.btn-edit` | 圆角 pill，小号，编辑/AI |
| `.btn-sm` | 圆角 pill，更小号 |
| `.icon-btn` | 方形小图标按钮 |
| `.icon-btn.danger` | hover 红色 |
| `.preset-btn` | 圆角 pill，primary 透明底，快捷运行 |

### Badge 类

| 类 | 颜色 |
|---|---|
| `.badge` | 通用 pill，surface-3 底 |
| `.badge-ok` | 绿色 |
| `.badge-warn` | 黄色 |
| `.badge-export` | primary，等宽字体 |
| `.lang-python/shell/ruby/node/binary` | 各语言对应色 |
| `.type-oneshot` | primary（clay） |
| `.type-service` | yellow |

### 脚本卡片 accent（左边框）

```css
.script-card[data-type="oneshot"] → border-left: 3px solid var(--primary)
.script-card[data-type="service"] → border-left: 3px solid var(--yellow)
```

### 目录卡片

```css
.dir-card-list        → grid, 2列
.dir-card             → flex row，surface-2，hover lift + primary glow
.dir-enter-btn        → 右侧 pill，pointer-events: none（由父级 click 处理）
```

### 日志面板

```css
.log-panel            → 暗色终端容器，background: #2b211b
.log-toolbar          → 顶部状态栏，background: #3b2e27
.log-pre              → 输出区，max-height: 340px，font-mono 12px
.log-stdout           → color: #f3e7dc（暖白）
.log-stderr           → color: #ffb4a9（淡红）
```

### 状态点

```css
.dot-running  → 绿色 + pulse 动画
.dot-stopped  → surface-4，灰色边框
```

### zshrc 条目边框颜色

```css
.zshrc-function → purple
.zshrc-alias    → yellow
.zshrc-export   → green
.zshrc-path     → red
.zshrc-source   → cyan
```

---

## Alpine.js 状态机（main.js hatchApp()）

### 视图路由

```
tab = 'scripts' | 'ai'   （settings 已合并入 scripts 首页 accordion）
selectedDir = null | { path, exportName, isScriptDir }
aiSubTab = 'overview' | 'connect' | 'skills' | 'tools'

scripts + !selectedDir  → 目录卡片 + workspace-settings accordion
scripts + selectedDir   → 脚本列表
ai                      → AI 侧边栏布局
editTarget != null      → 右侧编辑抽屉叠加
```

### 关键状态字段

| 字段 | 类型 | 用途 |
|---|---|---|
| `allScripts` | `ScriptMeta[]` | ScanAll() 结果，用于目录卡片计数 |
| `currentScripts` | `ScriptMeta[]` | 当前目录脚本列表 |
| `statuses` | `{path: string}` | `"running"\|"stopped"`，每 3s 轮询 |
| `runningSet` | `{path: true}` | oneshot 运行中标志 |
| `scriptLogs` | `{path: [{line,stream}]}` | 每脚本最多 500 行 |
| `logTarget` | `string\|null` | 当前展开日志的脚本 path |
| `editTarget` / `editForm` | — | 编辑抽屉 |
| `aiForm` | `{enabled,baseUrl,apiKey,model}` | AI 连接配置 |
| `aiSubTab` | `string` | AI 子页面（overview/connect/skills/tools/prompts/data） |
| `claudeStatus` / `codexStatus` | `object\|null` | AI 工具状态卡片数据 |
| `skills` | `array` | ListSkills() 结果 |
| `skillsStats` | `{total,claude,codex}` | hero stat 用 |
| `traceLog / traceLogDate` | — | trace log 查看器 |
| `dataPaths / dataPathsForm` | — | trace log dir / skills dir 配置 |
| `zshrcEntries / zshrcFilter` | — | 解析后 zshrc + 类型过滤 |
| `autoLaunchEnabled` | `bool` | 开机自启 |
| `wsShellOpen / wsDirsOpen / wsEnvOpen / wsZshrcOpen` | `bool` | workspace accordion 展开状态 |
| `success / successTimer` | — | 成功 toast，3s 自动消失 |
| `error` | `string` | 错误 toast，6s 自动消失 |
| **v1.1 – v1.3** ||
| `historyTarget / historyRecords / historyOpenId / historyLogText / historyLoading` | — | 运行历史 modal |
| `editForm.envPairs / .notifyOnDone / .schedule / .requiresRaw / .healthUrl / .preScriptsRaw` | — | 编辑抽屉新字段 |
| `promptTemplates / promptDefaults / promptDefaultsOpen / promptOps / promptSaving` | — | Prompt 模板编辑器 |
| `dataInventory / retention / retentionSaving` | — | 数据目录面板 |
| `aiProviders / activeProviderID / quickAIOpen / quickAIActiveID / quickAITraceID` | — | 顶栏快捷弹层 |
| `skillAIInstruction / skillAIBusy` | — | Skill ✦ AI 优化栏 |

### 事件监听（setupEvents）

**重要**：`setupEvents()` 在调用前会先 `EventsOff` 清除旧监听，防止 HMR 重复注册：

```js
setupEvents() {
  EventsOff('executor:log', 'executor:done', 'executor:status');
  EventsOn('executor:log',    (ev) => { /* 追加日志，滚到底 */ });
  EventsOn('executor:done',   (ev) => { /* 更新 runningSet + exitCode */ });
  EventsOn('executor:status', (ev) => { /* 更新 statuses */ });
  // v1.2: AI streaming + trace fanout
  EventsOn('ai:trace',  (t)  => { /* prepend to aiTraces */ });
  EventsOn('ai:stream', (ev) => { /* {id, chunk, done, error} */ });
}
```

### 新 Go 方法调用方式

wailsjs/App.js 只在 `wails build` 后才重新生成。新增方法在 build 前用：
```js
window['go']['main']['App']['MethodName'](arg)
```

---

## Pico CSS 覆盖要点

1. **semantic element 内边距清零**：`header, nav, main, footer, section, article, aside { padding:0; margin:0; }`
2. **article 卡片样式清零**：`article { background: transparent; box-shadow: none; border-radius: 0; }`
3. **aria-busy button 颜色**：`button[aria-busy="true"] { background: var(--primary); }` （否则 Pico 用蓝色）
4. **role="button" 背景**：dir-card 用了 `role="button"`，需要显式覆盖 Pico 的蓝色 primary 背景
5. **Pico primary 变量**：在 `:root` 中覆盖所有 `--pico-primary-*` 变量为 clay 色

### AI 模块 CSS 类

| 类 | 用途 |
|---|---|
| `.ai-layout` | 侧边栏 + 内容区，flex row |
| `.ai-sidenav` | 左侧 sub-tab 导航，flex column |
| `.ai-sidenav button.active` | 激活 sub-tab，primary 色 |
| `.ai-content` | 右侧内容区，flex: 1 |
| `.ai-usage-row` | 2列 grid，1fr 1fr，equal height |
| `.ai-usage-card` | flex column，撑满高度 |
| `.hero-stat-card` | hero 统计卡片，14px 圆角，72px 最小高度 |
| `.hero-stat-card--skills` | purple 边框 + 总数+分项布局 |
| `.hero-stat-value` | 26px 粗体值 |
| `.hero-stat-value--sm` | 15px 小值（长文本） |
| `.hero-stat-total` | 36px purple 数字 |
| `.success-toast` | 成功提示，绿色，右下角 fixed |
| `.workspace-settings` | 脚本首页 accordion 容器 |
| `.ws-section` | accordion 单区块 |
| `.ws-section-header` | accordion 标题行（cursor pointer） |
| `.ws-chevron` | 折叠箭头 |

### v1.1 – v1.3 新增 CSS 类

> 所有新功能 CSS 在 `style.css` 末尾 `v1.1 – v1.3 additions` 分区。**禁止给新功能写 inline `style=`** —— 一律抽到具名 class。

| 前缀 / 类 | 用途 |
|---|---|
| `.history-record / .history-record-row / .history-record-actions / .history-record-args / .history-record-log` | 运行历史 modal 单条记录 |
| `.history-modal-body / .history-modal-wide / .history-empty` | 历史 modal 容器 |
| `.ok-mark / .err-mark` | ✓/✗ 颜色原子（绿 / 红） |
| `.field-hint` | 输入框下小灰提示行 |
| `.prompt-card / .prompt-card-header / .prompt-card-title / .prompt-card-op / .prompt-card-actions` | Prompt 模板卡 |
| `.prompt-status-on / .prompt-status-off / .prompt-default-block / .prompt-actions-row` | Prompt 状态徽章 + 默认值预览 |
| `.retention-form / .retention-note` | 数据 tab 顶部保留策略 4 列网格 |
| `.data-dir-list / .data-dir-card / .data-dir-row / .data-dir-info / .data-dir-head` | 数据目录卡片 |
| `.data-dir-key / .data-dir-missing / .data-dir-path / .data-dir-desc / .data-dir-stats / .data-dir-actions / .data-empty` | 卡片字段 |
| `.quick-ai-wrap / .quick-ai-popover / .quick-ai-header` | 顶栏弹层容器（绝对定位，420px） |
| `.quick-ai-section / .quick-ai-section-label / .quick-ai-section-title` | 弹层内部分块 |
| `.quick-ai-traces-header / .quick-ai-trace + .is-open / .quick-ai-trace-row / .quick-ai-trace-err / .quick-ai-trace-pre / .quick-ai-empty / .quick-ai-actions / .quick-ai-manage-row / .quick-ai-toggle-arrow` | trace 列表细节 |
| `.skill-ai-bar` | Skill 编辑器 ✦ AI bar 间距（继承 `.se-ai-bar`） |

---

## 响应式断点

```css
@media (max-width: 820px) {
  .dir-card-list        → 单列
  .ai-form-grid         → 单列
  .ai-capability-list   → 单列
  .edit-drawer          → max-width: 100vw（全宽）
}
```

---

## 常见坑

- **CSS 加载顺序**：`index.html` 里不加 style.css 的 `<link>`，只通过 `main.js` 的 import 加载
- **EventsOn 重复注册**：setupEvents() 调用前必须 EventsOff，否则 HMR 后事件回调加倍
- **Alpine 嵌套 x-for + 方法调用**：在嵌套 x-for 里调用设置响应式数据后立即读取的方法会静默失败，改用内联 `@click.stop` 处理
- **args nil vs []**：Go 的 `nil` 序列化为 JSON `null`（显示 args 输入框），`[]ArgDef{}` 序列化为 `[]`（隐藏输入框）
- **safeId(path)**：日志面板 DOM id 用 `safeId()` 替代 `btoa()`，避免中文路径编码错误
- **wailsjs 禁止手动编辑**：`frontend/wailsjs/` 完全由 `wails build` 生成

---

## 快速上手

### 添加新页面视图

1. 在 `index.html` 增加 `<main x-show="tab==='xxx'">`
2. 在 tab-nav 加 `<button class="tab" :class="tab==='xxx'?'active':''" @click="tab='xxx'">` 
3. 用 `.page-hero` 作顶部信息条，`.settings-section` 作内容卡片

### 添加新 script-card 操作按钮

放在 `.card-actions` 内，使用 `.btn-edit` 类（圆角 pill，小号）：
```html
<button class="btn-edit" @click.stop="myAction(s)">按钮名</button>
```

### 添加新 settings 区块

```html
<section class="settings-section">
  <div class="section-header-row">
    <h3>区块标题</h3>
    <button class="btn-sm">操作</button>
  </div>
  <!-- 内容 -->
</section>
```

### 新增 badge

直接用 `<span class="badge badge-ok">文字</span>`，或按需用 `badge-warn` / `badge-export` / `type-oneshot` 等。
