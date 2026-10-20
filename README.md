<div align="center">

<img src="build/appicon.png" alt="Hatch" width="120" />

# Hatch

**macOS 脚本管理与开发环境仪表盘 — Go + Wails + Alpine.js**

![Platform](https://img.shields.io/badge/platform-macOS%20arm64-000000?logo=apple&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2.12-DF0000?logo=wails&logoColor=white)
![Alpine.js](https://img.shields.io/badge/Alpine.js-v3-8BC0D0?logo=alpinedotjs&logoColor=white)
![Pico CSS](https://img.shields.io/badge/Pico%20CSS-v2-13795B)
![Vite](https://img.shields.io/badge/Vite-v3-646CFF?logo=vite&logoColor=white)

</div>

Hatch 是一个面向 Apple Silicon macOS 的本地桌面工具，基于 **Wails v2**（Go 后端 + WKWebView 外壳 + Alpine.js 前端）。
它把脚本管理、AI 编码助手配置、VPN（Clash Verge Rev）维护和日常备忘集中在一个窗口里。
在 App Store 之外分发，**不需要 Apple 开发者账号**，所有数据都保存在本机。

---

## 功能模块

顶部四个模块均可在配置文件中单独开关（见 [模块开关](#模块开关)）。

| 模块 | 主要能力 |
|---|---|
| 🗂 **工作台** | 按目录管理脚本：自动扫描与 `@hatch.*` 注解识别、一次性脚本 / 常驻服务（start/stop/restart）、实时日志与运行历史、定时任务、前置脚本链、单脚本环境变量、完成通知；`~/.zshrc` export 托管块、zshrc 解析 / 备注 / AI 整理；PATH 与依赖体检；内置脚本编辑器与 AI 生成 |
| ✦ **AI** | OpenAI 兼容 provider 管理与调用日志；Claude Code / Codex 模型与推理档位配置、预设、用量状态；Skills 主仓库与双端同步；MCP 注册、安装（本地 / git / npm）、同步、连接测试与 AI 选型推荐；命令型 Hooks 双端同步；项目级工作区画像（预览 → 快照 → 应用 / 回滚）；提示词覆盖；本地数据目录与保留策略 |
| 🔒 **VPN** | Clash Verge Rev 系统代理 / TUN 状态、当前节点与出口 IP、一键体检；配置文件与 Profile 原文编辑（保存前 YAML 校验并自动备份）；配置备份 / 预览 / 恢复；自定义规则汇总 |
| ✓ **备忘录** | 待办与提醒（横幅 / 声音）、图片与文件附件托管；单词本：词典查询、发音、AI 补全、间隔复盘 |

所有 AI 能力只生成草稿供确认，不会自动写入配置、脚本或 zshrc。

---

## 快速开始

### 环境要求

- Apple Silicon Mac（`darwin/arm64`）
- Go 1.23+
- Node.js + npm（前端构建）
- Wails CLI v2：`go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
- 可选：`create-dmg`（`brew install create-dmg`，仅打 DMG 时需要）

### 编译并安装

```bash
make build      # 编译打包 → build/bin/Hatch.app
make install    # 复制到 /Applications 并清除隔离属性
```

> **覆盖安装注意**：`make install` 使用 `cp -r`，若 `/Applications/Hatch.app` 已存在，新包会被复制到旧包内部而不是覆盖。重新安装前请先退出并删除旧版本：
>
> ```bash
> osascript -e 'quit app "Hatch"'
> rm -rf /Applications/Hatch.app
> make build && make install
> ```

只想本地运行、不安装：`make open`。

---

## 模块开关

配置文件 `~/Library/Application Support/Hatch/config.json` 中的 `modules` 控制顶部模块是否显示：

```json
{
  "modules": {
    "workbench": true,
    "ai": true,
    "vpn": false,
    "memo": true
  }
}
```

| 键 | 模块 |
|---|---|
| `workbench` | 🗂 工作台 |
| `ai` | ✦ AI（同时隐藏右上角「✦ AI」快捷面板） |
| `vpn` | 🔒 VPN（同时隐藏右上角 VPN 状态按钮） |
| `memo` | ✓ 备忘录 |

- 未写 `modules` 或缺少某个键时，该模块默认**显示**；新版 Hatch 第一次保存设置时会自动写入四个 `true`。
- 修改后**无需重启**：点击顶部「↻ 刷新」或切回 Hatch 窗口即生效；当前模块被隐藏时自动切到其他可见模块。
- 四个全部关闭时会保留工作台，窗口不会为空。
- Hatch 自身保存配置时不会覆盖你手动改的 `modules`。
- 隐藏只影响界面：定时任务、备忘提醒、数据清理等后台功能照常运行。

> `config.json` 中以明文保存 AI API Key 和 MCP 环境变量，请勿外传。

---

## 开发

```bash
make dev          # cd src && wails dev，热更新开发模式（自动打开窗口）
make test         # cd src && go test ./... -v
make test-short   # 只显示 PASS / FAIL
```

- 开发服务器同时运行在 http://localhost:34115，可在浏览器 devtools 中通过 `window.go.app.App.<Method>()` 调用 Go 方法。
- 新克隆的仓库需要先构建一次前端（`make build` 或 `cd frontend && npm run build`），生成 `src/web/dist` 后 `go test` / `go vet` 才能编译通过（该目录会被嵌入二进制）。
- `frontend/wailsjs/**` 由 Wails 在 `wails dev` / `wails build` 时自动生成，请勿手动编辑。

### 架构约定

- **前后端分离**：`frontend/` 是独立的 Vite 工程；`src/` 是 Go module（`module hatch`）与 Wails 工程目录。所有 Go / Wails 命令都在 `src/` 下执行（Makefile 已内置 `cd src`）。
- **后端分层**：`src/app` 是唯一绑定到前端的包（JS 中为 `window.go.app.App`），按界面模块拆分文件，只负责加锁、保存配置、发送事件；具体逻辑放在对应的模块包（`core` / `workbench` / `ai` / `vpn` / `memo`）。模块包不依赖 Wails 和 `app`，`core` 不依赖任何业务模块。
- **前端约定**：Alpine.js v3 + Pico CSS v2，不使用 Tailwind；`main.js` 先 import Pico 再 import `./style.css`；新功能不写内联 `style`，使用 `style.css` 中的具名 class。

---

## 打包发布

目标平台固定为 **Apple Silicon macOS（`darwin/arm64`）**，由根目录 `Makefile` 驱动：

| 命令 | 作用 |
|---|---|
| `make build` | `cd src && wails build -platform darwin/arm64`，增量编译 → `build/bin/Hatch.app` |
| `make build-clean` | 加 `-clean` 完整重新编译 |
| `make sign` | 对 `Hatch.app` 做 **ad-hoc 自签名**（`codesign --sign -`，无需开发者证书） |
| `make dmg` | 用 `create-dmg` 打成 `build/bin/Hatch-<version>-arm64.dmg`（依赖 `sign`） |
| `make release` | 完整发布链：`build-clean → sign → dmg` |
| `make install` | 复制到 `/Applications` 并 `xattr -cr` 清除隔离属性 |
| `make open` | `xattr -cr` 后直接运行，绕过 Gatekeeper |

也可以使用 `./scripts/release.sh <version>` 一次完成构建、签名和 DMG 打包。

### 发布要点

- **签名策略**：ad-hoc 自签名，用户首次运行需清除 quarantine 属性（`make install` / `make open` 已内置 `xattr -cr`）。
- **分发物**：单个 `.dmg`，拖入 `/Applications` 即可。
- **版本号**：`Makefile` 的 `VERSION` 与 `src/wails.json` 的 `productVersion` 需要保持一致：

  ```bash
  make release VERSION=2.2.2   # 同时更新 src/wails.json 中的 productVersion
  ```

---

## 仓库结构

```text
hatch/
├── frontend/                  # 前端工程（Vite + Alpine.js + Pico CSS）
│   ├── index.html             # 页面模板
│   ├── src/                   # main.js（状态与交互）、style.css、assets
│   ├── wailsjs/               # Wails 自动生成的 Go 调用绑定
│   └── vite.config.js         # 构建产物输出到 ../src/web/dist
├── src/                       # 后端工程（Go module + Wails 项目）
│   ├── main.go                # 入口：嵌入 web/dist 并调用 app.Run
│   ├── go.mod / go.sum
│   ├── wails.json             # 前端目录 ../frontend，打包目录 ../build
│   ├── app/                   # Wails 绑定层：system_* / workbench_* / ai_* / vpn / memo*
│   ├── core/                  # 基础设施：config、fsutil、macos、strutil、backup、datadir、autostart
│   ├── workbench/             # 工作台：scanner、executor、scheduler、store、zshrc、diagnostics
│   ├── ai/                    # AI：llm、prompts、agent、skills、hooks、mcp、workspace
│   ├── vpn/                   # VPN：Clash Verge Rev 管理
│   └── memo/                  # 备忘录：待办、附件、单词、复盘
├── build/                     # 打包资源（图标、Info.plist）；产物在 build/bin（不入库）
├── scripts/release.sh         # 构建 → 签名 → DMG
└── Makefile                   # 构建 / 测试 / 发布命令入口
```

---

## 技术栈

| 层 | 技术 |
|---|---|
| 桌面框架 | Wails v2.12：Go 二进制 + macOS 原生 WKWebView，Go 方法绑定到前端并支持双向事件 |
| 后端 | Go 1.23+；直接依赖 `wailsapp/wails/v2`、`gopkg.in/yaml.v3`、`pelletier/go-toml` |
| 前端 | Alpine.js v3、Pico CSS v2（暖色浅色主题，自定义 CSS 变量覆盖）、Vite v3 |
| AI | OpenAI 兼容协议 provider（`/chat/completions`） |
| 存储 | 无数据库，全部为本地 JSON / 文件 |

---

## 本地数据

| 用途 | 路径 |
|---|---|
| 主配置（含模块开关） | `~/Library/Application Support/Hatch/config.json` |
| 脚本缓存 | `~/Library/Application Support/Hatch/scripts.json` |
| 备忘录 / 单词 | `~/Library/Application Support/Hatch/memos.json` |
| 备忘录附件 | `~/Library/Application Support/Hatch/memo-files/`（可配置） |
| AI 调用日志 | `~/Library/Application Support/Hatch/traces/<date>.jsonl`（可配置） |
| 运行历史 | `~/Library/Application Support/Hatch/runs/<script>/<id>.{log,json}` |
| Skills 主仓库 | `~/Library/Application Support/Hatch/skills/`（可配置） |
| AI 工作区快照 | `~/Library/Application Support/Hatch/ai-workspace-snapshots/`（保留最近 20 个） |
| VPN 备份 | `~/Library/Application Support/Hatch/vpn-backups/`（可配置） |
| PID 文件 | `/tmp/hatch/<script-name>.pid` |
| zshrc 托管块 / 备份 | `~/.zshrc`（仅 Hatch block）/ `~/.zshrc.hatch-backup-<时间戳>` |
| 开机自启 | `~/Library/LaunchAgents/com.hatch.app.plist` |
| 配置备份导出 | `~/Desktop/hatch-backup-<时间戳>.zip`（默认） |

AI 调用日志、运行历史和备忘录附件按保留策略在启动时和每小时自动清理，可在 AI → 数据 页面调整。
