<div align="center">

<img src="build/appicon.png" alt="Hatch" width="120" />

# Hatch

**macOS 脚本管理仪表盘 — Go + Wails + Alpine.js**

![Platform](https://img.shields.io/badge/platform-macOS%20arm64-000000?logo=apple&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2.12-DF0000?logo=wails&logoColor=white)
![Alpine.js](https://img.shields.io/badge/Alpine.js-v3-8BC0D0?logo=alpinedotjs&logoColor=white)
![Pico CSS](https://img.shields.io/badge/Pico%20CSS-v2-13795B)
![Vite](https://img.shields.io/badge/Vite-v3-646CFF?logo=vite&logoColor=white)

</div>

Hatch 是一个 macOS 脚本管理仪表盘，基于 **Wails v2** 构建：Go 后端 + WKWebView 外壳 + Alpine.js 前端。
面向 Apple Silicon macOS，在 App Store 之外分发，**不需要 Apple 开发者账号**。

## 截图

<!-- 将应用截图放到 docs/screenshots/ 后替换下方路径 -->
<div align="center">

<img src="docs/screenshots/dashboard.png" alt="Hatch 主界面" width="800" />

</div>

> 暂未提供截图。运行 `make dev` 后对窗口截图，保存为 `docs/screenshots/dashboard.png` 即可在此显示。

---

## 技术栈

### 桌面框架
- **Wails v2** (`v2.12.0`) — 将 Go 二进制嵌入 macOS 原生 **WKWebView** 外壳，Go 后端方法绑定到前端，并支持双向事件通信。

### 后端 (Go)
- **Go 1.23+**
- 直接依赖：
  - `github.com/wailsapp/wails/v2` — 桌面运行时
  - `gopkg.in/yaml.v3` — YAML 解析
- 其余为 Wails 间接依赖（macOS 通知、websocket、uuid 等）。
- 后端按职责拆分为多个 `app_*.go`（脚本 CRUD、执行器、cron 调度、诊断、备份、zshrc、AI、VPN、MCP、hooks、Claude/Codex 配置等）；核心逻辑位于 `internal/`（config / scanner / executor / scheduler / autostart / store / zshrc）。
- AI 功能走 **OpenAI 兼容协议** 的 provider，支持流式响应。

### 前端
- **Alpine.js v3** (`3.15.12`) — 轻量响应式状态管理
- **Pico CSS v2** (`2.1.1`) — 基础样式，由自定义 CSS 变量覆盖（暖色 Anthropic 风浅色主题，不使用 Tailwind）
- **Vite v3** — 前端打包构建
- 入口约定：`main.js` 先 import Pico，再 import `./style.css`

### 数据存储
无数据库，全部为本地 JSON / 文件：

| 用途 | 路径 |
|---|---|
| 配置 | `~/Library/Application Support/Hatch/config.json` |
| 脚本缓存 | `~/Library/Application Support/Hatch/scripts.json` |
| AI traces | `~/Library/Application Support/Hatch/traces/<date>.jsonl` |
| 运行历史 | `~/Library/Application Support/Hatch/runs/<script>/<id>.{log,json}` |
| Master skills | `~/Library/Application Support/Hatch/skills/` |
| PID 文件 | `/tmp/hatch/<script-name>.pid` |
| zshrc 托管块 | `~/.zshrc`（仅 Hatch block）|
| App 自启动 | `~/Library/LaunchAgents/com.hatch.app.plist` |

---

## 开发

```bash
make dev          # wails dev，热更新开发模式（自动打开窗口）
make test         # go test ./internal/... -v
make test-short   # 只显示 PASS/FAIL
```

也可直接使用 `$(go env GOPATH)/bin/wails dev`。开发服务器同时运行在 http://localhost:34115，可在浏览器 devtools 中调用 Go 方法。

---

## 打包部署

目标平台固定为 **Apple Silicon macOS (`darwin/arm64`)**，通过 `Makefile` 驱动。

| 命令 | 作用 |
|---|---|
| `make build` | `wails build -platform darwin/arm64`（增量编译 → `build/bin/Hatch.app`）|
| `make build-clean` | 加 `-clean` 完整重新编译 |
| `make sign` | 对 `Hatch.app` 做 **ad-hoc 自签名**（`codesign --sign -`，无需开发者证书）|
| `make dmg` | 用 `create-dmg` 打成 `Hatch-<version>-arm64.dmg`（依赖 `sign`）|
| `make release` | 完整发布链：`build-clean → sign → dmg` |
| `make install` | 复制到 `/Applications` 并 `xattr -cr` 清除隔离属性 |
| `make open` | `xattr -cr` 后直接运行，绕过 Gatekeeper |

### 发布要点
- **签名策略**：使用 ad-hoc 自签名而非证书签名，因此用户首次运行需清除 quarantine 属性（`make install` / `make open` 已内置 `xattr -cr`）以绕过 Gatekeeper。
- **分发物**：单个 `.dmg`，用户拖入 `/Applications` 即可。
- **依赖工具**：`create-dmg`（`/opt/homebrew/bin`）、`wails` CLI（`$(go env GOPATH)/bin`）。

### 发布版本号
版本号在多处独立维护（`Makefile` 的 `VERSION`、`wails.json` 的 `productVersion`），发布时通过参数覆盖并保持同步：

```bash
make release VERSION=2.2.2
```

同时记得更新 `wails.json` 中的 `productVersion`，避免 DMG 文件名与实际版本不一致。

---

## 仓库结构

```text
hatch/
├── app.go / app_*.go       # Wails 绑定的后端方法
├── internal/               # config / scanner / executor / scheduler / autostart / store / zshrc
├── frontend/               # Alpine 模板、CSS、资源、生成的 wailsjs
├── build/                  # 构建产物与图标
├── wails.json              # Wails 项目配置
└── Makefile                # 构建 / 测试 / 发布命令
```

> `frontend/wailsjs/**` 由 Wails 在 `wails dev` / `wails build` 时自动生成，请勿手动编辑。
