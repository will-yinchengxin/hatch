# 模型榜单实施计划

> 状态：已被取代（2026-10-28）。EvalPlus / GenEval 等来源已移除，现行数据来源与实现见 `docs/leaderboards/sources.md`。此文件仅作历史记录。

**Goal:** 在 AI 侧栏提供无需账号、Key、LLM 的领域模型榜单，验证后安装本机。
**Architecture:** Go `ai/leaderboard` 管理官方数据适配、独立缓存和并发请求；app 薄绑定；Alpine 领域导航与响应式卡片。
**Tech Stack:** Go 标准库、Alpine、Pico、Wails。
**Spec:** `docs/leaderboards/design.md`（用户通过实施请求授权执行）。

## 全局约束
- 编程、图片（生成/编辑/理解）、前端、财务、综合；不接入工具/Agent 系统榜。
- 真实来源原始名次/分数；不合成总分；缺失模型发布日期显示未知。
- 缓存先展示，24 小时过期；按榜独立错误、成功/尝试时间，超时/大小上限、失败退避和合并并发刷新。
- 不修改用户配置、不新增后台任务；不手工修改 wailsjs。

## Review Focus
- 文件截断、混合 category、同分并列：完整读取官方文件并只保留目标分类，保留原 rank。
- 网络/格式失败：旧数据和成功时间不变，短时禁止重复请求。
- 同时刷新不同榜单及同榜单：缓存写入串行、同榜只请求一次。
- 模型日期与榜单发布日期混淆：独立展示，未知不计算天数。
- 前端切页中的异步返回：按 ID 合并，不把旧领域覆盖当前领域。

### Task 1: 来源与后端
- [x] 记录官方归属、CC BY 4.0、匿名路径及覆盖表。
- [x] 先写失败测试：完整文件、分类/并列、坏响应、旧缓存、24h、失败退避、并发合并、重启读取。
- [x] 实现 `New(path) *Service`, `List() []Board`, `Refresh(ctx,id,force) (Board,error)` 和 app 的 `GetModelLeaderboards` / `RefreshModelLeaderboard`。
- [x] `cd src && go test -race ./ai/leaderboard/...`。

### Task 2: 页面
- [x] AI 侧栏/aiGo/query route 加入 leaderboards；默认 coding/gen、所有来源、前十展开。
- [x] 缓存先显示，逐榜刷新独立状态；来源链接/许可/榜单日期/日期未知清晰展示。
- [x] 前端构建、浏览器桌面与窄屏实际渲染验证。

### Task 3: 验收及安装
- [x] 同步项目 reference 和前端 skill。
- [x] 匿名实网逐榜验证；`cd src && go test ./...`；`make build`；签名校验。
- [x] 审阅差异，安装到 /Applications/Hatch.app，启动并验证运行中的新二进制。

Ruling: 按用户明确要求直接实施；留在当前干净工作区，不创建额外 worktree、子代理或提交。

Ruling: datasets-server JSON 接口实网不稳定；最终使用官方 latest Parquet，新增 parquet-go v0.25.1（纯 Go、兼容项目 Go 版本）以移除分页和中转服务依赖。
Completion: 九榜匿名实网、全量 Go、race、前端状态与构建、原生默认/最小窗口、安装签名/摘要/运行路径已验证。

## 用户后续调整
- [x] 移除财务分类及来源定义。
- [x] 默认横向条形图，可切换详情列表；各榜独立零基准线性刻度，保留原始分数/并列排名，不合成成绩。
- [x] 图表计算测试覆盖正值、负值、零值、空数据和前十/展开。
- [x] 原生 UI 验证后更新本机安装。
