---
### [2026-05-21 16:05] Codex 模型与努力程度配置 + 快捷预设

**Files modified**: `internal/config/config.go`, `app_files.go`, `frontend/index.html`, `frontend/src/main.js`, `frontend/src/style.css`

#### 设计决策
- Codex 的当前配置读写 `~/.codex/config.toml` 顶层 `model` 与 `model_reasoning_effort`，前端仍展示为用户理解的 `model` / `effortLevel`。
- 写入采用行级 merge：只移除并重写这两个顶层 key，其他 TOML section、插件、项目 trust 配置和 UI 设置保持原样。
- 快捷预设存 Hatch 自己的 `config.json`，新增 `codexPresets`；首次读取时写入 4 个默认组合，并用 `codexPresetsSeeded` 避免用户删除后再次自动恢复。
- Codex UI 放在 AI → 工具 → Codex 卡片的状态区下方、`~/.codex/config.toml` 编辑器上方，保持与 Claude Code 卡片相同的信息层级，但使用独立状态名、方法名和 CSS class。

#### 偏差
- 没有照搬 Claude Code 的 `max` 档位，因为当前 Codex 本机模型缓存和配置实际使用的是 `low / medium / high / xhigh`。

#### 权衡
- 未引入 TOML 依赖，避免为了两个顶层字符串字段增加解析依赖；代价是这里只支持管理顶层同名字符串 key，符合当前配置需求。

#### 开放问题
- Codex 模型列表目前按本机已知模型固定在前端；后续可以从 `~/.codex/models_cache.json` 动态读取。
---

---
### [2026-05-21] Claude Code 模型与努力程度配置 + 快捷预设

**Files modified**: `internal/config/config.go`, `app_files.go`, `frontend/index.html`, `frontend/src/main.js`, `frontend/src/style.css`

#### 设计决策
- `model` 和 `effortLevel` 直接读写 `~/.claude/settings.json`，用 merge 方式（不覆盖其他字段），`SetClaudeCodeConfig` 用 `map[string]interface{}` 先读后写。
- 空字符串 = 删除该 key（即恢复默认），与 Claude Code 原生行为一致。
- 快捷预设存在 Hatch 自己的 `config.json`（新增 `claudeCodePresets` 字段），而非写入 `~/.claude/`，避免污染 Claude Code 的 settings 文件。
- 预设 ID 用 `strconv.FormatInt(time.Now().UnixNano(), 36)` 生成短字符串，够唯一、够短。
- UI 放在 Claude Code 卡片内部（状态区下方、cfg-editor 上方），内联折叠表单而非独立 modal，与现有 cfg-editor-block 风格一致。
- 努力程度 badge 用三色区分（绿/低、黄/中、橙/高），与脚本卡片的 accent 色系相符。

#### 权衡
- 努力程度没有放到 settings.json 的原生字段外再造一层"映射层"，直接读写 `effortLevel`（Claude Code 已支持 low/medium/high），简单直接。
- 没有为新方法导入 wailsjs binding（用 `window['go']['main']['App']` 调用），因为 wails dev 已自动重新生成了绑定文件，两种调法等价。

---
### [2026-05-20 ~] VPN 模块 — Clash Verge Rev 配置管理

**Files added**: `app_vpn.go`  
**Files modified**: `frontend/index.html`, `frontend/src/main.js`, `frontend/src/style.css`, `go.mod` (+ `gopkg.in/yaml.v3`)

#### 设计决策
- 新增顶层 `tab='vpn'` 导航按钮，与 `scripts` / `ai` 并列，而非放在 AI 子 tab 下 — 因为 VPN 配置管理与 AI 功能无关，独立入口更清晰。
- 备份目录存放在 `~/Library/Application Support/Hatch/vpn-backups/` 而非桌面，避免污染桌面、且跟 Hatch 数据集中管理。
- 引入 `gopkg.in/yaml.v3` 解析 `profiles.yaml`，而非手动 text-parse — profiles.yaml 结构复杂（嵌套 option 字段），手动解析出错风险高。
- 写操作均复用现有 `backupIfExists(path)` 模式（`app_files.go:158`），统一创建 `.bak` 备份，用户不会意外丢失原始文件。
- ExportVPNBackup 排除 `cache.db`、`*.dat`（geoip/geosite/country mmdb，几十 MB）和 `logs/` 目录，只保留配置相关 yaml/js 文件。
- Profile content 读取时尝试 `uid.yaml` → `uid.js` → `uid`（原始 file 字段）三种路径，兼容 Clash Verge 不同类型的 profile 文件。

#### 权衡
- 考虑过将 VPN 作为 AI tab 的第七个子 tab — 但 VPN 是完全独立的功能域，单独 tab 避免 AI tab 过于臃肿，也更便于后续扩展（比如支持多个代理工具）。
- 考虑过不引入 yaml.v3（项目原来无 YAML 依赖）— 但 profiles.yaml 解析是核心功能，手动解析代码量大且易出 bug，yaml.v3 是标准选择且 license 兼容。

#### 开放问题
- 目前没有「切换当前激活 profile」功能（需要写 profiles.yaml 的 `current` 字段），可后续按需加。
- 外部 zip 恢复目前需用户手动输入路径；如果 Wails 支持 FileDialog，可升级为文件选择器。
- 没有 Clash Verge Rev 进程检测（恢复时建议用户先关 app），可后续加 ps 检测 + 提示。
---
