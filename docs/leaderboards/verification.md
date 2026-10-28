# 本地验收与安装记录

日期：2026-10-28（Asia/Shanghai）。设计输入保留在 `design.md`；本记录描述实际交付边界。

- `cd src && go test ./...`：所有包通过。
- `cd src && go test -race ./ai/leaderboard/...`：通过，覆盖完整文件解析、类别过滤、原始并列名次、整数/小数 schema、24h 缓存、失败保留、60s 退避、重试、同榜并发合并、缓存重启、取消、响应上限和未知日期。
- `node --test frontend/leaderboards.test.mjs`：3 个测试通过，覆盖本地日历天数、未来/不完整/未知日期、领域来源筛选、切页中的异步返回、缓存跳过、逐榜错误隔离。
- `HATCH_LEADERBOARD_LIVE=1 go test ./ai/leaderboard/... -run TestLiveAnonymousSources -v -parallel 2`：9/9 真实官方数据文件成功，无 Key、账号或 LLM，数量/榜单日期见 `sources.md`。
- `npm run build`、`make build`：通过，Wails 自动生成绑定；保留现有 `time.Time` 生成器告警，不影响构建。生成器本身输出的 models.ts 空行包含 tab，未手工改动。
- 原生 WKWebView 验证：AI 侧栏入口、默认编程前十、展开 413 条与收起、图片三个子类、财务未接入、来源筛选、重启缓存恢复、手动刷新中显示旧缓存、默认 1024px 并排与约 900px 单列布局。发现并修正 Pico 深色默认色影响下拉框/标题/模型名/评分的对比度。
- 已安装至 `/Applications/Hatch.app`，`codesign --verify --deep --strict` 通过，安装二进制与构建产物逐字节一致。
- 最终 SHA-256：`a5169b806c0407c4c7f8e103ed980e0cd9588fa35cf2a37df8f190a331b16536`。
- 运行路径核对：`/Applications/Hatch.app/Contents/MacOS/Hatch`。旧版在 `/private/tmp/hatch-before-leaderboards.jk0dAy/Hatch.app` 保留临时备份。

边界：当前实际同步来源为 Arena 的九个不同评测；Artificial Analysis 需要 Key，HELM Finance 未验证稳定接口，均明确显示暂未接入。模型精确版本发布日期不在来源字段中，显示未知。没有生成示例排名或猜测日期。没有提交/推送代码、创建子代理、hooks 或后台任务。

审阅：按设计逐项检查缓存锁与原子写入、失败成功时间、原始名次/分数、日期口径、页面异步切换、生成绑定和安装产物；未发现需要阻断安装的问题。

## 后续：移除财务并增加榜单图表

- 用户确认使用横向条形图展示模型排名与分数；当前领域为编程、图片、前端、综合。
- 默认 SVG 条形图，独立零基准线性刻度，保留原始排名/分数，可切换详情列表；前三名使用较深 clay 颜色。
- 前端 4 个测试、全量 Go、前端构建、原生 Wails 构建均通过；实际检查默认图表、详情列表切换、1024px 并排及约 900px 单列布局。
- 已更新 `/Applications/Hatch.app`，签名及构建/安装二进制一致性通过，安装版实际显示四个领域和横向条形图。
- 当前二进制 SHA-256：`fe59998da2d59c9df8b9ea97ed5e60bffcf4c1ef7fe472b1d962f6901f7b7ff8`。上节旧摘要属于首次交付。
- 本次旧版备份：`/private/tmp/hatch-before-charts.2HNT8t/Hatch.app`。

## 后续：固定前 20 名

- 图表和详情列表统一调用 `modelBoardRows(board)`，固定取前 20 条；删除展开全部按钮和 expanded 状态，来源完整缓存继续保留。
- 前端 4 个测试通过（25 条输入只展示 20 条且不修改源数组），前端构建及 Wails 原生构建通过。
- 安装版实际检查：编程两个榜单共 40 个条形图，均到第 20 名且没有第 21 名；详情列表末尾也均到第 20 名，没有展开按钮。
- 已更新 `/Applications/Hatch.app`，签名和二进制一致性校验通过。当前 SHA-256：`23f06a6575e657ebb35c9a0ebc6b1b1a408d31778af572ec64c0310e351af927`。
- 本次旧版备份：`/private/tmp/hatch-before-top20.eR0in9/Hatch.app`。

## 后续：多来源与参考图竖向榜单

- 新增 EvalPlus HumanEval+ / MBPP+、Epoch AI ECI / GPQA Diamond、GenEval 历史图片对齐，共 5 榜；原有 Arena 9 榜继续正常。总计 14/14 匿名实网同步通过。
- 新来源没有源 rank 时按单项成绩降序生成同分竞赛排名，显示「按来源分数排序」；保留原始分数、Chat/Completion 设置及精确 model version，不合成 Hatch 总分。历史覆盖在卡片中明示。
- 参考用户图片实现竖向厂商配色柱：柱内两位小数概览、厂商本地 SVG、完整斜向多行名称、点击高亮与原始数据详情；单榜最多 20 条，横向滚动，键盘 Tab 可访问第 20 名并以空格选中。
- 原生 WKWebView 已验证柱体与纵轴对齐、标识加载、模型标签无截断、选中第 2/20 名详情、横向自动滚动、来源筛选 Epoch AI 仅显示两榜（40 柱）、不同评分单位及版本日期出处。默认窗口与约 900px 窄窗口均检查。
- Go 全量、leaderboard race 与新增适配器解析边界测试通过。前端 5 个测试通过：缓存、异步领域切换、前二十、原始分数、正负/零刻度、竖向几何、厂商回退、完整标签、选择与日期。Wails 生产构建通过。
- SVG 来自 LobeHub Lobe Icons，MIT 许可与来源信息随应用打包在 `leaderboard-brands/`；不需要图标 CDN 或额外运行时。
- 已更新 `/Applications/Hatch.app`，`codesign --verify --deep --strict` 与构建/安装二进制一致性检查通过。安装后重新打开模型榜单，确认四个领域、每榜前 20 名、竖向彩色柱、柱内分数、厂商图标、倾斜名称及选中详情正常显示。
- 最终二进制 SHA-256：`ef724c6ee4339dc80e011d86fad76e4952018243acbc83c573108dc3fb6c5aac`。本次旧版备份：`/private/tmp/hatch-before-columns.4GFmyI/Hatch.app`。

## 后续：数据源与界面重构（2026-10-28 晚）

- 数据：Arena 改用风格控制子集（与 arena.ai 默认一致），新增 Agent Arena、截图转网页、中文榜；移除 EvalPlus、GenEval 与 Artificial Analysis 占位。详见 `sources.md`。
- 机制：按来源文件下载与缓存（7 个文件 → 11 个榜单），每榜存前 50；旧缓存格式自动丢弃重建。修复 Agent Arena 缺列被读成 0 分的问题。
- 界面：领域分段控件 → 榜单卡片选择 → 单榜视图（前 10 柱状图，按榜单拟合纵轴并标明起点，IPS 以 0 为基线，置信区间须线；下方完整排名，可筛选、展开至 50）。窄窗口自动切 6 列并隐藏次要列。
- 验证：`go vet ./...`、`go test ./...`、`go test -race ./ai/leaderboard/`、实网 11/11 榜单、`node --test frontend/leaderboards.test.mjs`（9 项）、`npm run build`、`make build` 均通过；`wails dev` 浏览器中检查四个领域、Agent/GPQA/图像榜、筛选、展开、窄宽度。
- 已安装 `/Applications/Hatch.app`，`codesign --verify --deep --strict` 通过，与构建产物逐字节一致，安装版打开模型榜单显示新界面。旧版备份：`/private/tmp/hatch-before-lb-redesign.nIyxcj/Hatch.app`。
