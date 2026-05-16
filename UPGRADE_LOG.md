# Hatch 整改日志（v1.1 → v1.3）

> 用于记录本轮所有的整改任务、过程、遇到的问题、解决方式、最终效果，方便后续核实和回滚。
> 每完成一个版本（v1.1 / v1.2 / v1.3）会执行一次 git commit。

---

## 整体路线图

| 版本 | 主题 | 任务数 |
|---|---|---|
| v1.1 | 收紧地基（重构 + 立刻能用的执行体验） | 6 + 1 commit |
| v1.2 | 真实开发场景（自动化 + AI 体验） | 5 + 1 commit |
| v1.3 | 工作流粘合（启动器 / 备份 / 编排） | 4 + 1 commit |

---

## v1.1 收紧地基

### 任务列表

| # | 任务 | 状态 |
|---|---|---|
| 1.1-1 | 拆分 app.go 为多文件（按职责切 receiver） | ✅ |
| 1.1-2 | 拆分前端 main.js 为多 mixin | ⏸ 延后 |
| 1.1-3 | 设置抽屉化（齿轮按钮 → 抽屉） | ⏸ 延后 |
| 1.1-4 | 执行历史归档 + 重跑 | ✅ |
| 1.1-5 | 脚本级 Env 变量 | ✅ |
| 1.1-6 | macOS 原生通知 | ✅ |

### 1.1-1 拆分 app.go

**问题**
- `app.go` 已经膨胀到 1560 行，60+ 个 `(a *App)` 方法堆在一起：脚本 CRUD、AI、trace、Claude/Codex、Skills、zshrc、文件读写……
- 每加一个能力就会更乱，后续不可维护。

**整改方式**
- 按职责拆为：
  - `app.go`：核心结构 + startup + 公共工具
  - `app_scripts.go`：脚本 CRUD / 扫描 / Override
  - `app_executor.go`：RunScript / ServiceAction / 状态
  - `app_zshrc.go`：zshrc 解析 / 注释 / 整理
  - `app_ai.go`：AI 调用 / Provider / Suggest 系列
  - `app_observability.go`：trace 日志 + Claude/Codex status
  - `app_skills.go`：保留现状（已经独立）
  - `app_files.go`：脚本文件读写 + OpenFile + DocLinks + Settings
- Wails 仅看方法集，拆分零行为变化。

**执行结果**
- `app.go` 从 1560 行精简到 ~110 行（仅保留 App 结构 / startup / 共享 helper / migrate / restore）。
- 新增 6 个职责清晰的文件：
  - `app_scripts.go` — 目录管理 + 扫描 + Override
  - `app_executor.go` — RunScript / ServiceAction + 新增 ListRuns/GetRunLog/DeleteRun + scriptEnv/notifyHook 辅助
  - `app_zshrc.go` — zshrc 解析 / 注释 / AI 整理 / syncZshrc
  - `app_ai.go` — Provider / Suggest / callAI / Trace 调用
  - `app_observability.go` — Claude/Codex status + AI Trace 文件持久化 + macOS 通知发送 (`sendMacNotification`)
  - `app_files.go` — 脚本文件 I/O + AppDataPaths + AutoLaunch + ClaudeSettings/CodexConfig + DocLinks + OpenFile
- 行为零变化：Wails 只看 `(a *App)` 方法集合。
- `go test ./...` 全部通过（zshrc 7/7、scanner、executor 含 history、app）。

**遇到的问题 & 解决**
- 单元测试 `runOneshotWithTestEmitter` 因 `runOneshotInternal` 多了 `env` 参数报错，加 `nil` 即修复（`internal/executor/executor_test.go:248`）。

### 1.1-2 拆分前端 main.js

**问题**
- `main.js` 1435 行 / `index.html` 1341 行，所有 tab 的状态和方法塞在一个 Alpine 组件里。
- 想改一处样式或加一个字段，都要在巨型对象里翻。

**整改方式**
- 按 tab 拆 mixin：
  - `mixins/scripts.js`
  - `mixins/ai.js`
  - `mixins/skills.js`
  - `mixins/settings.js`
- `main.js` 仅做：导入 wails 绑定、合并 mixin、注册 Alpine.data。
- CSS 用 `@import` 分块：`tokens.css` / `layout.css` / `scripts.css` / `ai.css` / `forms.css`。

**执行结果 / 问题**
- ⏸ **延后到 v2.0 重构期**。理由：纯前端组织重构、用户不可见、风险高（一处出错全 App 不可用）；当前 1435 行 main.js 单文件仍可维护；优先把 v1.2/v1.3 的可见功能交付出来后再统一做。
- 已在新功能（运行历史、env 编辑）中保持「相关方法相邻分组」的写法，给后续拆分留好边界。

### 1.1-3 设置抽屉化

**问题**
- Settings 当前是 scripts 首页的 accordion，杂项太多（环境路径、目录、Exports、zshrc、AI、AppDataPaths、DocLinks……）。
- 用户找一项设置要在脚本首页滚来滚去。

**整改方式**
- 顶部右上加一个 ⚙ 按钮 → 唤出全屏右侧抽屉。
- 抽屉内 tab 化：环境/目录/Exports/zshrc/AI/路径/外观。
- scripts 首页只保留：目录卡片、AI 入口、最近运行等脚本核心信息。

**执行结果 / 问题**
- ⏸ **延后到 v1.3**。理由：抽屉化是布局重构、需要重新组织 4 块 settings markup + 改导航逻辑，工作量约等于一个独立小版本；v1.3 增加导入导出/PATH 体检/Menubar 时一并做更经济。
- 当前 scripts 首页 accordion 仍可用。

### 1.1-4 执行历史归档 + 重跑

**问题**
- 当前 `executor:log` 只在前端内存里，刷新或关闭面板就丢。
- 没办法回看"昨天那次跑挂了输出是什么"。

**整改方式**
- 路径：`~/Library/Application Support/Hatch/runs/<script-basename>/<timestamp>.json`，结构 `{script, args, startedAt, endedAt, exitCode, lines:[]}`。
- 后端 `RunOneshot` 写一个 sink；提供 `ListRuns(path)` / `GetRun(id)` / `DeleteRun(id)`。
- 前端：脚本详情下新增"历史"区块，每条带 ▶ 重跑（参数自动回填）。
- 保留策略：默认保留每脚本最近 50 条，超过 LRU 淘汰。

**执行结果**
- 后端：新增 `internal/executor/history.go`，`runRecorder` 把每行 stdout/stderr 写入 `~/Library/Application Support/Hatch/runs/<basename>/<id>.log`，结束时落盘 `<id>.json` 元数据（参数、退出码、耗时、stdout 末 50 行预览）。
- `RunOneshotWithEnv` 同时驱动事件流和归档，互不影响。
- 启动时 `executor.PruneAllHistory(50)` 异步清理超量记录。
- 暴露 `(a *App) ListRuns / GetRunLog / DeleteRun`。
- 前端：新增 "📜 历史" 按钮 + 全屏 modal（`historyTarget` / `historyRecords` 状态），支持查看完整日志、▶ 重跑（参数自动回填）、✕ 删除单条。
- 因为 `wails build` 才会重新生成 `wailsjs/go/main/App.js`，前端通过 `window.go.main.App.ListRuns` 等动态绑定调用，无需手工维护。

**问题 & 解决**
- 无。一次性通过。

### 1.1-5 脚本级 Env 变量

**问题**
- 全局 exports 已有，但脚本级/目录级 env 没有。
- 用例：`kops` 想固定一个 `KUBECONFIG`，目前只能改 zshrc 或参数塞。

**整改方式**
- `ScriptMeta.Env map[string]string`（config.go）。
- `executor.RunOneshotWithEnv` + `ServiceManager.StartWithEnv/RestartWithEnv` 注入 env，`mergeEnv` 函数把 `os.Environ()` 和自定义 env 合并（override 优先）。
- `app_executor.go` 的 `scriptEnv(path)` 从 Override 或 store 中查找。
- UI：编辑抽屉新增「环境变量」区，K=V 行可增删。

**执行结果**
- 后端测试无回归。
- 执行时日志会打印 `[hatch] env: KEY=val`，便于排查变量是否生效。

### 1.1-6 macOS 原生通知

**问题**
- 长跑脚本完成后用户不知道，要回到 App 看日志。

**整改方式**
- `sendMacNotification(title, body)`（`app_observability.go`）—— 直接调用 `osascript -e 'display notification "..." with title "..."'`，错误忽略（用户可能关闭通知权限）。
- `ScriptMeta.NotifyOnDone bool`。
- `app_executor.go` 的 `notifyHook(path)` 返回 `func(exitCode, durationMs)` 回调；`RunOneshotWithEnv` 在 doneFn 内调用，弹通知带 ✓/✗ + 耗时。
- UI：编辑抽屉新增「脚本运行完成后弹出 macOS 通知」复选框。

**执行结果**
- 首次触发 macOS 会请求通知权限。
- 仅 oneshot 触发；service 是 long-running，不需要这种通知。

---

## v1.2 真实开发场景

### 任务列表

| # | 任务 | 状态 |
|---|---|---|
| 1.2-1 | 内置 cron 定时执行 | ✅ |
| 1.2-2 | 依赖检查 + 健康检查 | ✅ |
| 1.2-3 | AI 流式输出 + Prompt 模板可编 | ◐ 后端就绪 |
| 1.2-4 | 脚本上下文 AI 对话 | ⏸ 延后 |
| 1.2-5 | Trace 调试台升级 | ⏸ 延后 |

### 1.2-1 内置 cron 定时执行

**问题**
- 日常想跑备份、健康检查、定时清理脚本，每个都注册 LaunchAgent 太重。

**整改方式**
- 新增 `internal/scheduler` 包，零外部依赖。支持三种简单 spec：
  - `@every 30s` / `@every 5m` / `@every 1h`
  - `@daily HH:MM` —— 每天某时
  - `@hourly :MM` —— 每小时第 MM 分钟
  - 故意不引入 robfig/cron，避免 `go mod tidy` 触发网络访问；后续如需 cron 表达式可平滑替换。
- `ScriptMeta.Schedule` 字段 + 注释解析 `@hatch.schedule`。
- `app_v12.go` 的 `ApplyAllSchedules` 在启动 + SaveScriptMeta 后调用，全量重建。
- UI：编辑抽屉新增 "定时执行 schedule" 输入框，含提示语。

**执行结果**
- `internal/scheduler/scheduler_test.go` 5 个测试通过（解析、间隔、当日/隔日触发、错误用例、实际触发）。
- 计划任务通过 `executor.RunOneshotWithEnv` 复用现有归档 + 通知链路，自动出现在历史里。

### 1.2-2 依赖检查 + 健康检查

**问题**
- 依赖 `kubectl/jq` 之类命令的脚本，启动后才报错；service 类只看 PID 不知道实际是否健康。

**整改方式**
- `ScriptMeta.Requires []string` + `@hatch.requires: kubectl, jq` 注释解析。
- `ScriptMeta.HealthURL` + `@hatch.health: http://localhost:8080/healthz` 注释解析。
- `(a *App) CheckScriptHealth(path)` —— 同步检查 PATH 上的命令存在性 + HTTP 2s 超时 GET，返回 `HealthCheckResult{MissingDeps, HealthOK, HealthStatus, HealthLatency, HealthError}`。
- `(a *App) CheckCommandAvailable(name)` —— 单命令探测。
- UI：编辑抽屉新增 requires / healthUrl 字段。前端调用 `window.go.main.App.CheckScriptHealth(path)` 渲染状态徽章（留给 v1.3 进一步整合到列表里）。

**执行结果 / 后续**
- 后端 + 编辑表单可写入。脚本列表上的可视化"健康灯"留到 v1.3 settings 重构时一并加。

### 1.2-3 AI 流式输出 + Prompt 模板可编

**问题**
- 现有 Suggest 系列是阻塞调用，体验差；prompt 写死在 Go 里，用户无法按自己脚本风格调教。

**整改方式（部分）**
- `Config.PromptTemplates map[string]string` 持久化；`GetPromptTemplates/SavePromptTemplates`。
- `(a *App) effectivePrompt(op, default)` 返回用户覆盖 prompt 或默认 prompt。
- `(a *App) SuggestScriptContentStream(id, ...)` —— 走异步路径，通过 `ai:stream` 事件 emit `{id, chunk, done, error}`。当前实现仍调用阻塞 endpoint 后一次性 emit，为前端保留稳定接口；真正的 SSE 改造放在后续小版本。

**执行结果 / 后续**
- 后端接口和事件协议固定（`ai:stream`），前端可以先按事件流写 UI。
- 真正 SSE（分行解码 OpenAI `data: ` chunk）会作为 v1.3.1 增量优化。
- Prompt 模板的 UI 编辑器留到 v1.3 settings 抽屉一起做。

### 1.2-4 脚本上下文 AI 对话

**执行结果 / 问题**
- ⏸ **延后到 v1.3 之后**。理由：完整的多轮 chat 需要后端状态机 + 前端聊天 UI + history token 限制策略，规模 ≈ 一个独立小版本。当前 `SuggestScriptContent` 已经把脚本路径和现有内容塞给 AI，单轮版"针对脚本提建议"已经能用。

### 1.2-5 Trace 调试台升级

**执行结果 / 问题**
- ⏸ **延后到 v1.3 settings 重构时一并做**。理由：filter/聚合 UI 要在 settings 抽屉里有合适落位；当前 trace 列表仍可用。

---

## v1.3 工作流粘合

| # | 任务 | 状态 |
|---|---|---|
| 1.3-1 | 快速启动器 / Menubar 极简模式 | ⏳ |
| 1.3-2 | PATH 体检 + alias 管理 | ⏳ |
| 1.3-3 | 导入导出备份 | ⏳ |
| 1.3-4 | 脚本编排 / 前置依赖 | ⏳ |

> 详细问题/方案/结果会在进入 v1.3 时补完。

---

## 提交记录

| 版本 | Commit | 时间 |
|---|---|---|
| v1.1 | `fc755fb` | 2026-05-16 |
| v1.2 | — | — |
| v1.3 | — | — |
