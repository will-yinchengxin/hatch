package hooks

import "hatch/core/config"

// Templates returns the built-in starter hooks gallery. All templates are
// command-type and work identically under Claude Code and Codex (events not
// supported by Codex are gated off in the UI at sync time).
func Templates() []config.HookDef {
	return []config.HookDef{
		// ── 通知 ──────────────────────────────────────────────
		{
			ID:            "notify-on-stop",
			Name:          "完成时桌面通知",
			Description:   "每次回合结束时弹出 macOS 通知",
			Event:         "Stop",
			Command:       `osascript -e 'display notification "任务完成" with title "Claude Code"'`,
			StatusMessage: "发送通知…",
		},
		{
			ID:            "notify-subagent",
			Name:          "子代理完成通知",
			Description:   "子代理结束时弹出 macOS 通知",
			Event:         "SubagentStop",
			Command:       `osascript -e 'display notification "子代理已完成" with title "Claude Code"'`,
			StatusMessage: "发送通知…",
		},
		{
			ID:            "notify-webhook",
			Name:          "完成时 Webhook 通知",
			Description:   "回合结束时 POST 到 Webhook（先把 URL 改成你的地址）",
			Event:         "Stop",
			Command:       `curl -s -X POST -H 'Content-Type: application/json' -d '{"text":"Claude 任务完成"}' "https://example.com/webhook" >/dev/null || true`,
			Timeout:       15,
			StatusMessage: "通知 Webhook…",
		},
		// ── 格式化 / Lint ─────────────────────────────────────
		{
			ID:            "format-prettier",
			Name:          "Prettier 格式化（编辑后）",
			Description:   "编辑或写入文件后用 Prettier 格式化该文件",
			Event:         "PostToolUse",
			Matcher:       "Edit|Write",
			Command:       `jq -r '.tool_input.file_path // empty' | xargs -r npx prettier --write`,
			Timeout:       60,
			StatusMessage: "Prettier 格式化中…",
		},
		{
			ID:            "format-gofmt",
			Name:          "gofmt 格式化 Go 文件",
			Description:   "编辑 .go 文件后自动 gofmt -w",
			Event:         "PostToolUse",
			Matcher:       "Edit|Write",
			Command:       `jq -r '.tool_input.file_path // empty' | grep '\.go$' | xargs -r gofmt -w`,
			Timeout:       30,
			StatusMessage: "gofmt 中…",
		},
		{
			ID:            "format-black",
			Name:          "Black 格式化 Python 文件",
			Description:   "编辑 .py 文件后自动 black 格式化",
			Event:         "PostToolUse",
			Matcher:       "Edit|Write",
			Command:       `jq -r '.tool_input.file_path // empty' | grep '\.py$' | xargs -r black`,
			Timeout:       30,
			StatusMessage: "Black 格式化中…",
		},
		{
			ID:            "lint-eslint",
			Name:          "ESLint 自动修复",
			Description:   "编辑 JS/TS 文件后运行 eslint --fix",
			Event:         "PostToolUse",
			Matcher:       "Edit|Write",
			Command:       `jq -r '.tool_input.file_path // empty' | grep -E '\.(js|jsx|ts|tsx)$' | xargs -r npx eslint --fix`,
			Timeout:       60,
			StatusMessage: "ESLint 中…",
		},
		{
			ID:            "typecheck-tsc",
			Name:          "TypeScript 类型检查",
			Description:   "编辑 TS 文件后运行 tsc --noEmit（结果回灌上下文）",
			Event:         "PostToolUse",
			Matcher:       "Edit|Write",
			Command:       `npx tsc --noEmit 2>&1 | head -40`,
			Timeout:       120,
			StatusMessage: "类型检查中…",
		},
		// ── 安全 / 拦截 ───────────────────────────────────────
		{
			ID:          "block-rm",
			Name:        "拦截危险删除",
			Description: "拦截包含 rm -rf 的 Bash 命令",
			Event:       "PreToolUse",
			Matcher:     "Bash",
			Command:     `jq -r '.tool_input.command' | grep -q 'rm -rf' && echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Destructive command blocked"}}' || true`,
		},
		{
			ID:          "block-force-push",
			Name:        "拦截 git 强制推送",
			Description: "拦截 git push --force / -f",
			Event:       "PreToolUse",
			Matcher:     "Bash",
			Command:     `jq -r '.tool_input.command' | grep -qE 'git +push.*(--force|-f)' && echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"force push blocked"}}' || true`,
		},
		{
			ID:          "protect-secrets",
			Name:        "保护敏感文件",
			Description: "拦截对 .env / 密钥文件的编辑",
			Event:       "PreToolUse",
			Matcher:     "Edit|Write",
			Command:     `jq -r '.tool_input.file_path // empty' | grep -qE '(\.env|\.pem|id_rsa|\.key)$' && echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"protected file"}}' || true`,
		},
		{
			ID:            "audit-bash",
			Name:          "Bash 命令审计日志",
			Description:   "把每条 Bash 命令追加写入审计日志",
			Event:         "PreToolUse",
			Matcher:       "Bash",
			Command:       `C=$(jq -r '.tool_input.command'); echo "$(date '+%F %T') $C" >> ~/hatch-bash-audit.log`,
			StatusMessage: "记录命令…",
		},
		// ── 会话 / 上下文 ─────────────────────────────────────
		{
			ID:          "session-context",
			Name:        "会话开始注入 git 分支",
			Description: "会话启动时把当前 git 分支注入上下文",
			Event:       "SessionStart",
			Matcher:     "startup",
			Command:     `echo "{\"hookSpecificOutput\":{\"hookEventName\":\"SessionStart\",\"additionalContext\":\"branch: $(git branch --show-current 2>/dev/null)\"}}"`,
		},
		{
			ID:          "prompt-timestamp",
			Name:        "提示词附加时间戳",
			Description: "用户提交提示词时注入当前时间作为上下文",
			Event:       "UserPromptSubmit",
			Command:     `echo "{\"hookSpecificOutput\":{\"hookEventName\":\"UserPromptSubmit\",\"additionalContext\":\"now: $(date '+%F %T')\"}}"`,
		},
		{
			ID:            "backup-transcript",
			Name:          "压缩前备份会话记录",
			Description:   "上下文压缩前把 transcript 复制到本地目录",
			Event:         "PreCompact",
			Command:       `T=$(jq -r '.transcript_path // empty'); [ -n "$T" ] && mkdir -p ~/hatch-transcripts && cp "$T" ~/hatch-transcripts/ 2>/dev/null || true`,
			StatusMessage: "备份记录…",
		},
	}
}
