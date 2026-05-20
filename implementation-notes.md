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
