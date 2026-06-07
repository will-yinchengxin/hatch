package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"hatch/internal/config"
)

// ─── Hooks management ─────────────────────────────────────────────────────────
//
// A Hatch hook is a single command-type hook authored once and synced to:
//   - Claude Code: merged into ~/.claude/settings.json under the "hooks" key
//   - Codex:       merged into ~/.codex/hooks.json under the "hooks" key
//
// Both tools share the same hooks JSON schema:
//
//	{"hooks": {"<Event>": [{"matcher": "...", "hooks": [{"type":"command", ...}]}]}}
//
// Hatch-owned handlers carry a "hatchId" marker so sync can update/remove its own
// entries without touching hand-written hooks. Only type:"command" is used, which
// both Claude Code and Codex execute identically.

const (
	hookTargetClaude = "claude"
	hookTargetCodex  = "codex"
)

// HooksStats are the quick counts shown in the AI sidenav.
type HooksStats struct {
	Total  int `json:"total"`
	Claude int `json:"claude"`
	Codex  int `json:"codex"`
}

// HookEventInfo describes one supported hook event and which targets accept it.
type HookEventInfo struct {
	Event       string `json:"event"`
	Label       string `json:"label"`
	Claude      bool   `json:"claude"`
	Codex       bool   `json:"codex"`
	MatcherHint string `json:"matcherHint"`
}

// hookEvents is the canonical event catalog offered in the UI. Codex supports a
// smaller set; events it cannot run are gated off for the Codex sync toggle.
var hookEvents = []HookEventInfo{
	{Event: "SessionStart", Label: "会话开始", Claude: true, Codex: true, MatcherHint: "startup | resume | clear | compact"},
	{Event: "UserPromptSubmit", Label: "提交提示词", Claude: true, Codex: true, MatcherHint: "（无 matcher）"},
	{Event: "PreToolUse", Label: "工具调用前", Claude: true, Codex: true, MatcherHint: "Bash | Edit|Write | mcp__.*"},
	{Event: "PostToolUse", Label: "工具调用后", Claude: true, Codex: true, MatcherHint: "Bash | Edit|Write | mcp__.*"},
	{Event: "PreCompact", Label: "压缩前", Claude: true, Codex: true, MatcherHint: "manual | auto"},
	{Event: "PostCompact", Label: "压缩后", Claude: true, Codex: true, MatcherHint: "manual | auto"},
	{Event: "SubagentStop", Label: "子代理结束", Claude: true, Codex: true, MatcherHint: "子代理类型"},
	{Event: "Stop", Label: "回合结束", Claude: true, Codex: true, MatcherHint: "（无 matcher）"},
	{Event: "Notification", Label: "通知（仅 Claude）", Claude: true, Codex: false, MatcherHint: "permission_prompt | auth_success"},
	{Event: "SessionEnd", Label: "会话结束（仅 Claude）", Claude: true, Codex: false, MatcherHint: "clear | logout"},
}

// GetHookEventSupport returns the event catalog for the UI dropdown and toggle gating.
func (a *App) GetHookEventSupport() []HookEventInfo {
	out := make([]HookEventInfo, len(hookEvents))
	copy(out, hookEvents)
	return out
}

func hookEventInfo(event string) (HookEventInfo, bool) {
	for _, e := range hookEvents {
		if e.Event == event {
			return e, true
		}
	}
	return HookEventInfo{}, false
}

func validHookEvent(event string) bool {
	_, ok := hookEventInfo(event)
	return ok
}

func hookEventSupportedBy(event, target string) bool {
	info, ok := hookEventInfo(event)
	if !ok {
		return false
	}
	switch target {
	case hookTargetClaude:
		return info.Claude
	case hookTargetCodex:
		return info.Codex
	}
	return false
}

func validHookID(id string) bool {
	return mcpIDPattern.MatchString(strings.TrimSpace(id))
}

// ─── Registry CRUD ────────────────────────────────────────────────────────────

// defaultHelloHook is seeded once so users have a ready-to-use hook that proves
// the pipeline works: typing "hello" injects a confirmation into the context.
func defaultHelloHook() config.HookDef {
	return config.HookDef{
		ID:          "hello-test",
		Name:        "Hello 触发测试",
		Description: "用户输入包含 hello 时注入提示，用于验证 hook 是否生效",
		Event:       "UserPromptSubmit",
		Command:     `P=$(jq -r '.prompt // empty'); echo "$P" | grep -qi hello && echo '{"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"⚠️ 检测到 hello 关键字！hook 已触发！"}}' || true`,
		UpdatedAt:   time.Now().Unix(),
	}
}

// ListHooks returns the hook registry, newest first. On first call it seeds the
// default "hello" test hook and enables it for Claude Code (unless a hand-written
// hook would conflict), so the capability is demonstrable out of the box.
func (a *App) ListHooks() ([]config.HookDef, error) {
	a.mu.Lock()
	if !a.cfg.HooksSeeded {
		def := defaultHelloHook()
		if conflict, err := claudeHookEntryExists(def); err == nil && !conflict {
			if werr := writeClaudeHookConfig(def, true); werr == nil {
				def.EnabledInClaude = true
			}
		}
		a.cfg.Hooks = append(a.cfg.Hooks, def)
		a.cfg.HooksSeeded = true
		_ = a.cfg.Save()
	}
	out := cloneHooks(a.cfg.Hooks)
	a.mu.Unlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == out[j].UpdatedAt {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out, nil
}

// SaveHook creates or updates a hook (matched by ID), preserving sync flags.
// It does not write any target file; the frontend re-syncs enabled targets after save.
func (a *App) SaveHook(hook config.HookDef) error {
	hook, err := normalizeHook(hook)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, h := range a.cfg.Hooks {
		if h.ID == hook.ID {
			hook.EnabledInClaude = h.EnabledInClaude
			hook.EnabledInCodex = h.EnabledInCodex
			a.cfg.Hooks[i] = hook
			return a.cfg.Save()
		}
	}
	a.cfg.Hooks = append(a.cfg.Hooks, hook)
	return a.cfg.Save()
}

// DeleteHook removes a hook from the registry and from any target it was synced to.
func (a *App) DeleteHook(id string) error {
	if !validHookID(id) {
		return fmt.Errorf("invalid hook id")
	}
	a.mu.Lock()
	var removed config.HookDef
	found := false
	keep := a.cfg.Hooks[:0]
	for _, h := range a.cfg.Hooks {
		if h.ID == id {
			removed = h
			found = true
			continue
		}
		keep = append(keep, h)
	}
	if !found {
		a.mu.Unlock()
		return nil
	}
	a.cfg.Hooks = keep
	if err := a.cfg.Save(); err != nil {
		a.mu.Unlock()
		return err
	}
	a.mu.Unlock()

	if removed.EnabledInClaude {
		if err := writeClaudeHookConfig(removed, false); err != nil {
			return err
		}
	}
	if removed.EnabledInCodex {
		if err := a.syncCodexHooks(); err != nil {
			return err
		}
	}
	return nil
}

// SetHookEnabled syncs (or removes) one hook to/from the given target.
func (a *App) SetHookEnabled(id, target string, enabled bool) error {
	if !validHookID(id) {
		return fmt.Errorf("invalid hook id")
	}
	if target != hookTargetClaude && target != hookTargetCodex {
		return fmt.Errorf("unknown hook target: %s", target)
	}

	a.mu.RLock()
	var hook config.HookDef
	found := false
	for _, h := range a.cfg.Hooks {
		if h.ID == id {
			hook = h
			found = true
			break
		}
	}
	a.mu.RUnlock()
	if !found {
		return fmt.Errorf("hook not found: %s", id)
	}

	if enabled && !hookEventSupportedBy(hook.Event, target) {
		label := "Codex"
		if target == hookTargetClaude {
			label = "Claude Code"
		}
		return fmt.Errorf("事件「%s」不被 %s 支持", hook.Event, label)
	}

	switch target {
	case hookTargetClaude:
		// First-time enable must not clobber a hand-written hook. Re-syncing an
		// already-enabled hook (edit/update) overwrites its own marked handler.
		if enabled && !hook.EnabledInClaude {
			exists, err := claudeHookEntryExists(hook)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("Claude Code 配置中已存在匹配的手写钩子，已跳过以避免覆盖；如需 Hatch 接管请先在 ~/.claude/settings.json 中删除该条目")
			}
		}
		if err := writeClaudeHookConfig(hook, enabled); err != nil {
			return err
		}
	case hookTargetCodex:
		if enabled && !hook.EnabledInCodex {
			exists, err := codexHookEntryExists(hook)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("Codex 配置中已存在匹配的手写钩子，已跳过以避免覆盖；如需 Hatch 接管请先在 ~/.codex/hooks.json 中删除该条目")
			}
		}
		// Build the desired full Codex list with the flag applied, then write the
		// file before committing the flag so a failed write leaves state untouched.
		a.mu.RLock()
		list := cloneHooks(a.cfg.Hooks)
		a.mu.RUnlock()
		for i := range list {
			if list[i].ID == id {
				list[i].EnabledInCodex = enabled
			}
		}
		if err := writeCodexHooksFile(list); err != nil {
			return err
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.cfg.Hooks {
		if a.cfg.Hooks[i].ID != id {
			continue
		}
		if target == hookTargetClaude {
			a.cfg.Hooks[i].EnabledInClaude = enabled
		} else {
			a.cfg.Hooks[i].EnabledInCodex = enabled
		}
		a.cfg.Hooks[i].UpdatedAt = time.Now().Unix()
		break
	}
	return a.cfg.Save()
}

// GetHooksStats returns the sidenav counts.
func (a *App) GetHooksStats() HooksStats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	st := HooksStats{Total: len(a.cfg.Hooks)}
	for _, h := range a.cfg.Hooks {
		if h.EnabledInClaude {
			st.Claude++
		}
		if h.EnabledInCodex {
			st.Codex++
		}
	}
	return st
}

// GetHookTemplates returns the built-in starter hooks gallery. All templates are
// command-type and work identically under Claude Code and Codex (events not
// supported by Codex are gated off in the UI at sync time).
func (a *App) GetHookTemplates() []config.HookDef {
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

func hookAISystemPrompt() string {
	events := make([]string, 0, len(hookEvents))
	for _, e := range hookEvents {
		events = append(events, e.Event)
	}
	return `你是 Hatch 的 Hook 起草助手。Hook 是给 Claude Code / Codex 使用的命令型钩子。
请生成或改进一个 command 类型钩子，只输出一个 JSON 对象，不要解释，不要代码块。
字段要求：
- name: 简短中文名称
- description: 一句话说明
- event: 必须是以下之一: ` + strings.Join(events, ", ") + `
- matcher: 触发过滤（工具名/来源），不需要则空字符串
- command: 要执行的 shell 命令（命令通过 stdin 收到事件 JSON，可用 jq 读取 .tool_input / .prompt 等字段）
- timeout: 秒数，整数，不确定填 0
只输出形如 {"name":"","description":"","event":"","matcher":"","command":"","timeout":0} 的 JSON。`
}

// parseHookDraft strips fences/prose and unmarshals an AI response into a HookDef.
func parseHookDraft(raw string) (config.HookDef, error) {
	raw = strings.TrimSpace(raw)
	for _, fence := range []string{"```json", "```"} {
		if strings.HasPrefix(raw, fence) {
			raw = strings.TrimPrefix(raw, fence)
			raw = strings.TrimLeft(raw, " \t\r\n")
			break
		}
	}
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "```")
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, "{"); i > 0 {
		raw = raw[i:]
	}
	var draft config.HookDef
	if err := json.Unmarshal([]byte(raw), &draft); err != nil {
		return config.HookDef{}, fmt.Errorf("AI 返回内容无法解析为钩子：%w", err)
	}
	if !validHookEvent(draft.Event) {
		draft.Event = "PostToolUse"
	}
	return draft, nil
}

// SuggestHookFromDescription asks the configured AI to draft a command hook from a
// natural-language description. The draft is returned for the UI to confirm; it is
// never persisted automatically.
func (a *App) SuggestHookFromDescription(description string) (config.HookDef, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return config.HookDef{}, err
	}
	userBytes, _ := json.Marshal(map[string]any{"description": description})
	raw, err := a.callAIOp("SuggestHookFromDescription", ai, hookAISystemPrompt(), string(userBytes))
	if err != nil {
		return config.HookDef{}, err
	}
	return parseHookDraft(raw)
}

// SuggestHookEdit asks the configured AI to refine an existing hook given an
// instruction. The current hook plus the instruction are sent; the improved draft
// is returned for UI confirmation and is never persisted automatically.
func (a *App) SuggestHookEdit(current config.HookDef, instruction string) (config.HookDef, error) {
	a.mu.RLock()
	ai := a.activeAIConfig()
	a.mu.RUnlock()
	if err := validateAIConfig(ai); err != nil {
		return config.HookDef{}, err
	}
	userBytes, _ := json.Marshal(map[string]any{
		"current":     current,
		"instruction": instruction,
	})
	raw, err := a.callAIOp("SuggestHookEdit", ai, hookAISystemPrompt()+"\n用户会给出现有钩子 current 与改进指令 instruction，请在保留合理字段的前提下按指令改进并输出完整钩子 JSON。", string(userBytes))
	if err != nil {
		return config.HookDef{}, err
	}
	draft, err := parseHookDraft(raw)
	if err != nil {
		return config.HookDef{}, err
	}
	// Preserve the id so an edit updates in place rather than creating a new hook.
	if current.ID != "" {
		draft.ID = current.ID
	}
	return draft, nil
}

// ─── Normalization ────────────────────────────────────────────────────────────

func normalizeHook(hook config.HookDef) (config.HookDef, error) {
	hook.ID = sanitizeMCPID(orDefault(strings.TrimSpace(hook.ID), strings.TrimSpace(hook.Name)))
	if hook.ID == "" {
		hook.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if !validHookID(hook.ID) {
		return config.HookDef{}, fmt.Errorf("invalid hook id")
	}
	hook.Name = orDefault(strings.TrimSpace(hook.Name), hook.ID)
	hook.Event = strings.TrimSpace(hook.Event)
	if !validHookEvent(hook.Event) {
		return config.HookDef{}, fmt.Errorf("unsupported hook event: %s", hook.Event)
	}
	hook.Matcher = strings.TrimSpace(hook.Matcher)
	hook.Command = strings.TrimSpace(hook.Command)
	if hook.Command == "" {
		return config.HookDef{}, fmt.Errorf("hook command is required")
	}
	hook.Args = cleanStringSlice(hook.Args)
	hook.StatusMessage = strings.TrimSpace(hook.StatusMessage)
	if hook.Timeout < 0 {
		hook.Timeout = 0
	}
	hook.UpdatedAt = time.Now().Unix()
	return hook, nil
}

func cloneHooks(in []config.HookDef) []config.HookDef {
	out := make([]config.HookDef, len(in))
	for i, h := range in {
		out[i] = h
		out[i].Args = append([]string(nil), h.Args...)
	}
	return out
}

// ─── Target writers (shared JSON hooks-object helpers) ────────────────────────

func writeClaudeHookConfig(hook config.HookDef, enabled bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".claude", "settings.json")
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("invalid ~/.claude/settings.json: %w", err)
		}
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}
	hooksObj, _ := raw["hooks"].(map[string]interface{})
	if hooksObj == nil {
		hooksObj = map[string]interface{}{}
	}
	removeHatchHandler(hooksObj, hook.ID)
	if enabled {
		insertHatchHandler(hooksObj, hook)
	}
	if len(hooksObj) == 0 {
		delete(raw, "hooks")
	} else {
		raw["hooks"] = hooksObj
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func (a *App) syncCodexHooks() error {
	a.mu.RLock()
	list := cloneHooks(a.cfg.Hooks)
	a.mu.RUnlock()
	return writeCodexHooksFile(list)
}

// writeCodexHooksFile regenerates the Hatch-managed portion of ~/.codex/hooks.json
// from the given registry, preserving any hand-written (non-hatchId) handlers.
func writeCodexHooksFile(hooks []config.HookDef) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".codex", "hooks.json")
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil && len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("invalid ~/.codex/hooks.json: %w", err)
		}
	}
	if raw == nil {
		raw = map[string]interface{}{}
	}
	hooksObj, _ := raw["hooks"].(map[string]interface{})
	if hooksObj == nil {
		hooksObj = map[string]interface{}{}
	}
	removeHatchHandler(hooksObj, "") // strip all Hatch-managed handlers
	for _, h := range hooks {
		if h.EnabledInCodex && hookEventSupportedBy(h.Event, hookTargetCodex) {
			insertHatchHandler(hooksObj, h)
		}
	}
	if len(hooksObj) == 0 {
		delete(raw, "hooks")
	} else {
		raw["hooks"] = hooksObj
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := backupIfExists(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// renderHookHandler builds the JSON handler object Hatch writes into both targets.
func renderHookHandler(hook config.HookDef) map[string]interface{} {
	h := map[string]interface{}{
		"type":    "command",
		"command": hook.Command,
		"hatchId": hook.ID,
	}
	if len(hook.Args) > 0 {
		args := make([]interface{}, len(hook.Args))
		for i, a := range hook.Args {
			args[i] = a
		}
		h["args"] = args
	}
	if hook.Timeout > 0 {
		h["timeout"] = hook.Timeout
	}
	if hook.StatusMessage != "" {
		h["statusMessage"] = hook.StatusMessage
	}
	return h
}

// removeHatchHandler drops Hatch-owned handlers from a hooks object. When id is
// empty, all hatchId-marked handlers are removed; otherwise only the matching id.
// Hand-written handlers (no hatchId) are never touched. Empty groups/events prune.
func removeHatchHandler(hooksObj map[string]interface{}, id string) {
	for event, v := range hooksObj {
		groups, ok := v.([]interface{})
		if !ok {
			continue
		}
		var keptGroups []interface{}
		for _, g := range groups {
			gm, ok := g.(map[string]interface{})
			if !ok {
				keptGroups = append(keptGroups, g)
				continue
			}
			handlers, _ := gm["hooks"].([]interface{})
			var keptHandlers []interface{}
			for _, h := range handlers {
				if hm, ok := h.(map[string]interface{}); ok {
					if hid, has := hm["hatchId"].(string); has && hid != "" && (id == "" || hid == id) {
						continue
					}
				}
				keptHandlers = append(keptHandlers, h)
			}
			if len(keptHandlers) == 0 {
				continue
			}
			gm["hooks"] = keptHandlers
			keptGroups = append(keptGroups, gm)
		}
		if len(keptGroups) == 0 {
			delete(hooksObj, event)
		} else {
			hooksObj[event] = keptGroups
		}
	}
}

// insertHatchHandler appends the hook's rendered handler into the right event,
// reusing an existing matcher group when one matches, otherwise creating one.
func insertHatchHandler(hooksObj map[string]interface{}, hook config.HookDef) {
	handler := renderHookHandler(hook)
	groups, _ := hooksObj[hook.Event].([]interface{})
	for _, g := range groups {
		gm, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		if matcherEquals(gm["matcher"], hook.Matcher) {
			handlers, _ := gm["hooks"].([]interface{})
			gm["hooks"] = append(handlers, handler)
			hooksObj[hook.Event] = groups
			return
		}
	}
	newGroup := map[string]interface{}{"hooks": []interface{}{handler}}
	if hook.Matcher != "" {
		newGroup["matcher"] = hook.Matcher
	}
	hooksObj[hook.Event] = append(groups, newGroup)
}

func matcherEquals(v interface{}, matcher string) bool {
	s, _ := v.(string)
	return s == matcher
}

// ─── First-enable guards ──────────────────────────────────────────────────────

func claudeHookEntryExists(hook config.HookDef) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	return hookEntryExistsInFile(filepath.Join(home, ".claude", "settings.json"), hook)
}

func codexHookEntryExists(hook config.HookDef) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	return hookEntryExistsInFile(filepath.Join(home, ".codex", "hooks.json"), hook)
}

// hookEntryExistsInFile reports whether the file already has a hand-written
// (non-hatchId) handler under the same event+matcher with the same command.
func hookEntryExistsInFile(path string, hook config.HookDef) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return false, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("invalid %s: %w", path, err)
	}
	hooksObj, _ := raw["hooks"].(map[string]interface{})
	groups, _ := hooksObj[hook.Event].([]interface{})
	for _, g := range groups {
		gm, ok := g.(map[string]interface{})
		if !ok || !matcherEquals(gm["matcher"], hook.Matcher) {
			continue
		}
		handlers, _ := gm["hooks"].([]interface{})
		for _, h := range handlers {
			hm, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if hid, has := hm["hatchId"].(string); has && hid != "" {
				continue // Hatch-owned, not a hand-written conflict
			}
			if cmd, _ := hm["command"].(string); cmd == hook.Command {
				return true, nil
			}
		}
	}
	return false, nil
}
