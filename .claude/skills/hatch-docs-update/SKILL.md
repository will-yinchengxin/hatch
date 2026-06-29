---
name: hatch-docs-update
version: 1.1.0
description: "更新 Hatch 项目文档。当新功能、重构、配置字段、Wails API、前端状态或已知坑变化时，同步 CLAUDE.md、.claude/reference 和 Hatch skills。"
---

# Hatch 文档更新工作流

使用条件：需要让项目文档反映新功能、API 变更、数据结构变更、前端状态变化、内部行为变化、构建/测试规则变化或已知坑。

## 文档结构

```text
CLAUDE.md                                  # Claude Code 主参考，精简规则 + 路由
.claude/reference/data-model.md             # Config / ScriptMeta / VPN 类型
.claude/reference/bound-methods.md          # Wails methods + events
.claude/docs/frontend.md                    # Alpine state, layout, CSS conventions
.claude/docs/internals.md                   # scanner/executor/scheduler/zshrc/AI/data/VPN
.claude/docs/gotchas.md                     # 非显而易见的坑
.claude/skills/hatch-frontend/SKILL.md      # 前端修改前必读 workflow
.claude/skills/hatch-docs-update/SKILL.md   # 本 workflow
```

## 分析变更

优先查看：

```bash
git diff -- app*.go skills.go internal/ frontend/src/main.js frontend/index.html frontend/src/style.css
rg "func \\(a \\*App\\)" -n *.go
```

重点：

- 新增/删除/修改 Wails-bound methods -> 更新 `.claude/reference/bound-methods.md` 和 `CLAUDE.md`。
- `internal/config/config.go` 数据结构变更 -> 更新 `.claude/reference/data-model.md` 和 `CLAUDE.md`。
- `frontend/src/main.js` 新状态字段 -> 更新 `.claude/docs/frontend.md` 和 `.claude/skills/hatch-frontend/SKILL.md`。
- 新 UI 布局、CSS class、交互模式 -> 更新 `.claude/docs/frontend.md` 和 `.claude/skills/hatch-frontend/SKILL.md`。
- scanner/executor/scheduler/zshrc/AI/data/VPN 行为变化 -> 更新 `.claude/docs/internals.md`。
- 非显而易见的问题 -> 更新 `.claude/docs/gotchas.md`。

## 更新原则

- `CLAUDE.md` 只放高频规则、source of truth、命令、边界和文档路由。
- 细节进入 `.claude/docs/**` 和 `.claude/reference/**`，不要把入口文档重新变成完整手册。
- Skill 只保留可执行工作流和任务前检查，不承载所有参考内容。
- 不创建 hooks、agents、subagents 或后台自动化，除非用户明确同意。

## 检查标准

- `rg "func \\(a \\*App\\)" -n *.go` 的方法表与 bound-methods 文档没有明显遗漏。
- Config / ScriptMeta 字段与 data-model 文档没有明显漂移。
- 前端状态字段、sub-tab、主要页面结构与 frontend 文档没有明显漂移。
- `CLAUDE.md` 不包含过长实现细节，能路由到对应 reference 或 skill。
- 读回改动文件，确认 markdown frontmatter 和表格完整。
