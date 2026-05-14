# ScriptNest → Hatch：综合可执行方案

## 背景需求

```markdown
我现在想要再 mac M1 下(当然要适配 M1 后的版本) 完成这样一个管控的功能, 不要需要apple商店认证, 本地快捷安装下, 使用的技术栈尽量简单, 但是是可视化的图像界面

1.因为我个人喜欢使用zshrc, 那么开启软件请为我检测是否安装了 zshrc, 没有则安装下, 其次这里要允许创建自定的export 路径, 比如 我输出 Users/yinchengxin/tools, 自动为我追加 export SQLToDSL=/Users/yinchengxin/tools 类似这
样的语句到 zshrc 中并生效, 可以标记目录是否为脚本目录

2.完成 1 后, 我这里输人的目录可能包含各种脚本, 大多是 python shell 脚本等, 支持动态添加, 给我描述, 脚本内容, 如果目录中以存在, 自动分析能力, 列表形式展示, 支持分页搜索, 比如
 ✘  🐍 base  WillYin  ~  ls tools
WS_Create                 git_command               obs_backup.sh             start_mkcard.sh
cdn.sh                    github-recovery-codes.txt obs_restore.sh            time_now
codex_limit.sh            go.sh                     product_doc_install.sh    tmpwa.py
cug_logfiler.sh           goland-init.sh            pycharm-init.sh           webp2mp4
excalidraw.log            icon_generator.py         redis_sys.sh              x_following_person.sh
excalidraw_start.sh       imac-accent.sh            reset_navicat.sh
gcurl                     kops                      skill-init.sh
generate_media_pic.sh     kubectl                   start_goodle.sh

分析脚本能力的能后, 脚本的形态可能有差异, 比如有些执行一次就完成, 直接输出结果, 有些是可用控制起停 start/stop/reload 等, 把这些能力分析出来, 也要作为控制按钮, 因为我在界面进行统一控制, 
针对这个, 你给我推荐下技术栈, 执行方案, 其次为工程想一个好点的名字
你看方案是否合适, 因为我是后端工程师, 熟悉 python golang shell 等, 尽量满足我需求的前期下为我指定合理的方案

为我输出一份综合的可执行方案
```

## 一、最终技术选型

**Wails v2 (Go + WebView)**,理由前一份说清楚了,补充几个你会关心的实操点:

- macOS 上 Wails 用系统 WKWebView,**不打包 Chromium**,最终 .app 大约 8-15MB
- Apple Silicon 原生支持,`wails build -platform darwin/arm64` 一行出包
- **不需要 Apple 开发者账号**,本地 `codesign --sign -` 自签名即可,绕过 Gatekeeper 用 `xattr -cr Hatch.app` 一次性解决
- Go 后端代码就是你写 CLI 的姿势,前端就是 HTML/CSS/JS,两边通过自动生成的 binding 互调

**前端方案再精简一层**:不用 Vue/React,直接 **Alpine.js (15KB) + Pico.css (无类名 CSS)**。你后端工程师不需要为这个项目学构建工具链,Alpine 直接 `<script>` 引入,声明式绑定够用了。

**工程名:Hatch** ✅。短、好打、`hatch` 这个词本身就是"舱口/孵化"的意思,跟"统一入口管理脚本"的隐喻吻合。

---

## 二、对前方案的关键修正

### 修正 1:zshrc 检测的真实含义

zsh 本身是 macOS 13+ 的默认 shell,**一定存在**。要检测的不是 zsh,而是:

1. 用户当前 shell 是不是 zsh (`echo $SHELL`)
2. `~/.zshrc` 文件是否存在(不存在就创建空文件)
3. 是否需要切换默认 shell (`chsh -s /bin/zsh`)

不要无脑"安装 zshrc",这个表述不准确。

### 修正 2:export 写入要幂等且可回滚

直接 `echo "export X=Y" >> ~/.zshrc` 是新手写法,几个月后 zshrc 会变成垃圾堆。正确做法:

```bash
# >>> Hatch managed block >>>
export SQLToDSL="/Users/yinchengxin/tools"
export MyScripts="/Users/yinchengxin/scripts"
# <<< Hatch managed block <
```

Hatch 只管理这个块内的内容,**整块覆盖式写入**,块外的用户自己的配置一行不动。这是 conda/rbenv/nvm 都在用的成熟模式。

### 修正 3:执行脚本时的环境隔离

前一份提到"App 不依赖用户 shell 环境",这点对但要讲透:

Hatch 执行脚本时,**显式启动一个 login shell** 来 source zshrc,这样脚本里能用到用户配的 PATH、别名、函数。具体:

```go
cmd := exec.Command("/bin/zsh", "-l", "-c", scriptPath + " " + args)
```

`-l` (login) 会 source `/etc/zprofile` → `~/.zprofile` → `~/.zshrc`(交互式情况下),保证脚本环境跟用户在终端里手动跑一致。这点对你的 `kubectl`、`kops` 这种依赖 PATH 的脚本特别重要。

### 修正 4:脚本能力识别 — 三层降级

完全同意前一份的思路,但**约定注释块的格式我定细一点**,让你直接能用:

```bash
#!/bin/bash
# @hatch.name: Redis 系统管理
# @hatch.desc: 启停本地 Redis 服务
# @hatch.type: service        # oneshot | service
# @hatch.commands: start, stop, restart, status
# @hatch.pidfile: /tmp/hatch/redis_sys.pid
```

- `type: oneshot` → 界面显示一个"运行"按钮,可带参数输入框
- `type: service` → 界面显示 start/stop/restart 按钮 + 状态指示灯(绿/灰)
- 没有注释块 → 尝试 `script --help` (2秒超时) 解析,失败则降级为 oneshot

### 修正 5:服务进程的状态恢复

PID 文件方案对,补充一个细节:**App 启动时扫一遍所有 service 类脚本的 pidfile**,用 `kill -0 <pid>` 判断进程是否还活着,死了就清理 stale pidfile。这样 App 重启后状态自动恢复。

---

## 三、工程结构(最终版)

```
hatch/
├── main.go
├── app.go                    # 暴露给前端的方法集合
├── wails.json
├── go.mod
├── internal/
│   ├── config/
│   │   └── config.go         # ~/Library/Application Support/Hatch/config.json
│   ├── zshrc/
│   │   ├── manager.go        # 管理块的读写
│   │   └── manager_test.go
│   ├── scanner/
│   │   ├── scanner.go        # 目录扫描
│   │   ├── parser.go         # 注释块解析 + --help 探测
│   │   └── metadata.go       # ScriptMeta 结构
│   ├── executor/
│   │   ├── oneshot.go        # 一次性执行 + 输出流
│   │   ├── service.go        # 服务管理 + PID 文件
│   │   └── stream.go         # 实时输出推送(EventsEmit)
│   └── store/
│       └── store.go          # 脚本列表持久化
├── frontend/
│   ├── index.html
│   ├── src/
│   │   ├── app.js            # Alpine.js 应用逻辑
│   │   └── style.css         # 自定义样式(在 Pico 之上)
│   └── package.json          # 仅用于 Wails 的 dev server,不需要真的 npm install 啥
└── build/
    └── darwin/
        └── Info.plist        # macOS 应用元信息
```

---

## 四、核心数据结构

```go
// internal/scanner/metadata.go
type ScriptMeta struct {
    Path        string            `json:"path"`
    Name        string            `json:"name"`         // 显示名,默认文件名
    Desc        string            `json:"desc"`         // 描述
    Type        ScriptType        `json:"type"`         // oneshot | service
    Language    string            `json:"language"`     // shell | python | binary
    Commands    []string          `json:"commands"`     // service 类的子命令
    PidFile     string            `json:"pidFile"`      // service 类的 PID 文件
    Args        []ArgDef          `json:"args"`         // 参数定义(可选)
    Tags        []string          `json:"tags"`         // 用户自定义标签
    LastRunAt   int64             `json:"lastRunAt"`
    UpdatedAt   int64             `json:"updatedAt"`
}

type ScriptType string
const (
    TypeOneshot ScriptType = "oneshot"
    TypeService ScriptType = "service"
)

// internal/config/config.go
type Config struct {
    Directories []Directory      `json:"directories"`
    Exports     []ExportVar      `json:"exports"`
    OverrideMeta map[string]ScriptMeta `json:"overrideMeta"` // 用户手动改过的元信息
}

type Directory struct {
    Path       string `json:"path"`
    IsScriptDir bool  `json:"isScriptDir"`
    ExportName string `json:"exportName"`  // 比如 SQLToDSL
}

type ExportVar struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}
```

---

## 五、Day-by-Day 执行计划

### Day 1:环境 + 骨架(2-3小时)

```bash
# 安装 Wails
brew install go node      # node 用于前端 dev server
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 验证环境
wails doctor

# 创建项目(用 vanilla 模板,不带 Vue/React)
wails init -n hatch -t vanilla
cd hatch
wails dev    # 跑起来确认 hello world
```

### Day 2:zshrc 管理器 + 配置存储

先把 Go 部分写成可独立运行的 package,**带单元测试**。zshrc 管理器是最容易写错又最影响用户文件的地方,必须测好。

### Day 3:扫描器 + 注释块解析

扫描 + 注释块解析 + `--help` 探测。这部分纯字符串处理,Go 写起来很舒服。

### Day 4:执行器 + 实时输出流

oneshot 执行 + service 启停 + 输出通过 Wails EventsEmit 推送到前端。这步联调前端。

### Day 5:前端界面

Alpine.js + Pico.css,列表 + 搜索 + 分页 + 详情面板 + 实时日志面板。

### Day 6:打包 + 自签名 + DMG

```bash
wails build -platform darwin/arm64 -clean
codesign --force --deep --sign - build/bin/Hatch.app
# 用 create-dmg 做 DMG
brew install create-dmg
create-dmg Hatch.dmg build/bin/Hatch.app
```

首次打开:右键 → 打开,或 `xattr -cr /Applications/Hatch.app`。
