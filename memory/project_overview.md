---
name: project-overview
description: Hatch app features, data model, and key implementation details
metadata:
  type: project
---

Hatch is a macOS desktop app (Wails v2 + Go backend + Alpine.js frontend) for managing and running shell scripts.

**Why:** Personal productivity tool for script management on Apple Silicon, no App Store distribution.

**Key features added in May 2026:**
- `CommandPreset` struct in `ScriptMeta` for named fixed-args run buttons
- `ZshrcAnnotations` map in `Config` for annotating ~/.zshrc entries
- `internal/zshrc/parser.go` — parses ~/.zshrc into typed entries (function/alias/export/path/source/section/other)
- `app.go` — `ParseZshrc()` and `SaveZshrcAnnotation()` backend methods
- UI: script cards have colored left accent border by type, preset quick-run buttons, auto-open log panel with exit code display
- Settings: "Shell 配置解析" section parses and annotates ~/.zshrc

**How to apply:** When adding features, new Go methods need manual window['go']['main']['App']['Method']() wrappers in main.js since wailsjs/App.js is auto-generated on build.
