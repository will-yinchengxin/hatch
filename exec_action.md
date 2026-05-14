# Hatch 执行日志

## Day 1 — 环境 + 骨架 ✅

**日期**: 2026-05-14

### 完成内容

| 步骤 | 状态 | 说明 |
|------|------|------|
| Go 环境 | ✅ | go1.26.1 darwin/arm64 |
| Node 环境 | ✅ | v24.3.0 / npm 11.4.2 |
| Wails 安装 | ✅ | v2.12.0，`wails doctor` 全绿 |
| `wails init -t vanilla` | ✅ | 项目骨架生成，合并入 hatch/ |
| `internal/` 目录结构 | ✅ | config / zshrc / scanner / executor / store 目录已创建 |
| `wails build -platform darwin/arm64` | ✅ | 编译 + 自签名成功，17s，产物在 `build/bin/hatch.app` |
| git init + 初始提交 | ✅ | commit: `0137ee2` |

### 工程结构（当前）

```
hatch/
├── main.go              # Wails 入口
├── app.go               # App 结构体 + Greet stub
├── wails.json           # 项目配置
├── go.mod / go.sum
├── internal/
│   ├── config/          # (待实现 Day 2)
│   ├── zshrc/           # (待实现 Day 2)
│   ├── scanner/         # (待实现 Day 3)
│   ├── executor/        # (待实现 Day 4)
│   └── store/           # (待实现 Day 2)
├── frontend/
│   ├── index.html
│   └── src/
│       ├── main.js
│       └── style.css
├── build/
│   ├── bin/hatch.app    # 编译产物（.gitignore 已排除）
│   └── darwin/Info.plist
└── execution_plan.md
```

### 验证结果

```
wails doctor → SUCCESS: Your system is ready for Wails development!
wails build  → Built 'hatch.app' in 17.346s (自签名完成)
```

---

## Day 2 — zshrc 管理器 + 配置存储 ✅

**日期**: 2026-05-14

| 步骤 | 状态 | 说明 |
|------|------|------|
| `internal/config/config.go` | ✅ | Config / Directory / ExportVar / ScriptMeta 结构体，JSON Load/Save，存储路径 `~/Library/Application Support/Hatch/config.json` |
| `internal/zshrc/manager.go` | ✅ | Hatch managed block 读写；幂等覆盖（整块替换）；key 排序保证每次输出一致；块外内容零修改 |
| `internal/zshrc/manager_test.go` | ✅ | 7 个测试全部通过：创建块、幂等、更新、读取、空块返回 nil、删除、保留块后内容 |
| `internal/store/store.go` | ✅ | 脚本列表持久化，存储路径 `~/Library/Application Support/Hatch/scripts.json`；Replace / UpdateMeta / All，并发安全（RWMutex） |

```
go test ./internal/zshrc/... -v → PASS (7/7)
go build ./internal/...        → OK
```

---

## Day 3 — 扫描器 + 注释块解析（待执行）

- [ ] `internal/scanner/metadata.go` — ScriptMeta / ScriptType 数据结构
- [ ] `internal/scanner/scanner.go` — 目录扫描（可执行文件过滤）
- [ ] `internal/scanner/parser.go` — `@hatch.*` 注释解析 + `--help` 探测降级

---

## Day 4 — 执行器 + 实时输出流（待执行）

- [ ] `internal/executor/oneshot.go` — `/bin/zsh -l -c` 执行 + 输出捕获
- [ ] `internal/executor/service.go` — start/stop/restart + PID 文件管理
- [ ] `internal/executor/stream.go` — Wails EventsEmit 推送实时日志

---

## Day 5 — 前端界面（待执行）

- [ ] 替换 Pico.css + Alpine.js（直接 `<script>` 引入，无构建工具）
- [ ] 脚本列表 + 分页 + 搜索
- [ ] oneshot 运行按钮 + 参数输入框
- [ ] service 状态指示灯 + start/stop/restart 控制

---

## Day 6 — 打包 + DMG（待执行）

- [ ] `wails build -platform darwin/arm64 -clean`
- [ ] `codesign --force --deep --sign - build/bin/Hatch.app`
- [ ] `create-dmg Hatch.dmg build/bin/Hatch.app`
