# 模型榜单数据来源（2026-10-28 重构后）

只接入**当前在更新、可匿名下载、明确 CC BY 4.0** 的来源。每个榜单保留来源自己的名次与分数，不跨来源合成。

| 领域 | 榜单 | 来源文件 | 口径 | 本机实网（2026-10-28） |
|---|---|---|---|---|
| 编程 | Text Arena · 编程 | Arena `text_style_control` / `coding` | 真人盲测，风格控制（arena.ai 默认） | 413 个模型，2026-10-22 |
| 编程 | Agent Arena | Arena `agent` / `overall` | 编码智能体会话 IPS 分（可为负） | 50，2026-10-23 |
| 前端 | Code Arena · WebDev | Arena `webdev` / `overall` | 真人对比生成网页 | 151，2026-10-27 |
| 前端 | Code Arena · 截图转网页 | Arena `webdev` / `image_to_webdev` | 设计稿还原 | 64，2026-10-19 |
| 图像 | Text-to-Image / Image Edit | Arena `text_to_image`、`image_edit` / `overall` | 真人盲测 | 82 / 61 |
| 图像 | Vision Arena | Arena `vision_style_control` / `overall` | 风格控制 | 167，2026-10-19 |
| 综合 | Text Arena · 总榜 / 中文 | Arena `text_style_control` / `overall`、`chinese` | 风格控制 | 418 / 394，2026-10-22 |
| 综合 | Epoch Capabilities Index | Epoch `benchmark_data.zip` → `eci_scores.csv` | 官方综合指数 + 区间 + 发布日期 | 270 |
| 综合 | GPQA Diamond | 同上 → `gpqa_diamond.csv` | Epoch 自测 mean_score ×100 | 326 |

## 本次修正的问题

- **风格控制**：Arena 自 2025-05-16 起默认风格控制榜。旧实现把原始分（`text`/`vision`、`?scoring=raw`）作为主榜，和官网对不上；现在统一用 `*_style_control`。
- **过时来源移除**：EvalPlus（榜首仍为 2024 年 O1 Mini）、GenEval（5 个 2023 年模型）、Artificial Analysis 空占位卡。Epoch SWE-bench Verified / FrontierMath 自 2026-06 起不再覆盖新模型，未接入。
- **Agent Arena 解析**：该文件没有 `rating` 列；parquet-go 对缺失列解码为 0 而不是 nil，因此按文件 schema 选择 `rating*` 或 `score*` 列。
- **名称**：Arena 的 `Claude Opus 5.5 (High)` 拆成模型 + 档位；Epoch 的 `gpt-6-astra_max` 通过 `processed_data_for_eci.csv` 映射为 `GPT-6 Astra` + `Max`，映射不到的保留原始版本名。厂商名统一（`anthropic`/`Anthropic`、`Z.ai (Zhipu AI)`/`zai` 等）。

## 机制

- 每个来源文件只下载一次，供该文件上的所有榜单使用（旧实现 text parquet、Epoch zip 各下载两次）。7 个文件并行，整体 90s 超时，单响应 16 MiB 上限。
- 缓存只保留每榜前 50 名，12 小时内不自动重拉；失败保留旧数据并在 60s 内不重试；同时发起的刷新合并为一次。
- 坏行跳过而不是让整个榜单失败；Arena 每个分类只取最新发布日期的行。

复验：`cd src && HATCH_LEADERBOARD_LIVE=1 go test ./ai/leaderboard/ -run Live -v`。
