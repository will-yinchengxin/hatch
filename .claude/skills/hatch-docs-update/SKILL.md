---
name: hatch-docs-update
version: 1.0.0
description: "更新 Hatch 项目文档。当需要同步 CLAUDE.md 和 .claude/docs/ 以反映新功能、重构或 API 变更时使用。"
---

# Hatch 文档更新

## 文档结构

```
CLAUDE.md                          # 主参考文档 (目标 <200 行，保持精简)
.claude/docs/frontend.md           # 前端架构、UI 设计、Alpine 状态字段
.claude/docs/internals.md          # 内部包：scanner、executor、zshrc、AI 调用
.claude/docs/gotchas.md            # 已知坑 + Pico CSS 集成注意事项
.claude/skills/hatch-frontend/SKILL.md  # 前端组件模式 + CSS class 规范
```

## 更新流程

### 1. 分析变更

```bash
git diff main..HEAD -- app.go internal/ frontend/src/main.js frontend/index.html
git log main..HEAD --oneline
```

重点关注：
- `app.go` 新增/删除/修改的 bound 方法 → 更新 CLAUDE.md 的 Methods 表
- `internal/config/config.go` 数据结构变更 → 更新 CLAUDE.md 的 Core Data Model
- `frontend/src/main.js` 新状态字段 → 更新 `.claude/docs/frontend.md` 的 State Fields 表
- UI 布局变更 → 更新 `.claude/docs/frontend.md`

### 2. 更新原则

**CLAUDE.md** (精简主文档):
- 保持 < 200 行
- Methods 表：每行一个方法，只写用途，不写实现
- 数据模型：只写 Go struct，不写行为说明（行为在 internals.md）
- 新模块添加到 Tab Navigation 或相应 section

**`.claude/docs/` 详细文档**:
- `frontend.md` — 状态字段、UI 设计 token、布局、组件
- `internals.md` — 执行模型、文件操作、AI 调用细节
- `gotchas.md` — 非显而易见的坑（只写代码读不出来的东西）

**`hatch-frontend` skill**:
- 新 CSS 类添加到 Components 或 Layout 章节
- 新 Alpine 状态字段加入 State Machine 章节

### 3. 检查标准

更新完成后验证：
- [ ] CLAUDE.md 方法表与 `app.go` 所有 `func (a *App)` 方法一致
- [ ] 新的 `not yet in wailsjs` 方法已列出
- [ ] Config struct 字段与 `internal/config/config.go` 一致
- [ ] 行数检查：`wc -l CLAUDE.md`（目标 < 200）

```bash
grep "^func (a \*App)" app.go | sed 's/func (a \*App) //' | sed 's/(.*$//'
wc -l CLAUDE.md
```

## 常见变更场景

| 变更类型 | 需更新的文件 |
|---------|------------|
| 新 Go bound 方法 | CLAUDE.md Methods 表 |
| Config struct 新字段 | CLAUDE.md Core Data Model |
| 新 Alpine 状态字段 | `.claude/docs/frontend.md` State Fields |
| 新 CSS 类/组件 | `.claude/docs/frontend.md` + `hatch-frontend` skill |
| 新 UI 交互模式 | `.claude/docs/frontend.md` 对应 section |
| 非显而易见的 bug/坑 | `.claude/docs/gotchas.md` |
| 新内部包行为 | `.claude/docs/internals.md` |
