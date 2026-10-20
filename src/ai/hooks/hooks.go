// Package hooks manages Hatch's command-hook registry entries and their sync
// into Claude Code and Codex.
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
package hooks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hatch/core/config"
	"hatch/core/strutil"
)

const (
	TargetClaude = "claude"
	TargetCodex  = "codex"
)

// Stats are the quick counts shown in the AI sidenav.
type Stats struct {
	Total  int `json:"total"`
	Claude int `json:"claude"`
	Codex  int `json:"codex"`
}

// EventInfo describes one supported hook event and which targets accept it.
type EventInfo struct {
	Event       string `json:"event"`
	Label       string `json:"label"`
	Claude      bool   `json:"claude"`
	Codex       bool   `json:"codex"`
	MatcherHint string `json:"matcherHint"`
}

// events is the canonical event catalog offered in the UI. Codex supports a
// smaller set; events it cannot run are gated off for the Codex sync toggle.
var events = []EventInfo{
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

// Events returns a copy of the event catalog.
func Events() []EventInfo {
	out := make([]EventInfo, len(events))
	copy(out, events)
	return out
}

// EventNames returns the event identifiers in catalog order.
func EventNames() []string {
	names := make([]string, 0, len(events))
	for _, e := range events {
		names = append(names, e.Event)
	}
	return names
}

func eventInfo(event string) (EventInfo, bool) {
	for _, e := range events {
		if e.Event == event {
			return e, true
		}
	}
	return EventInfo{}, false
}

// ValidEvent reports whether event is in the catalog.
func ValidEvent(event string) bool {
	_, ok := eventInfo(event)
	return ok
}

// SupportedBy reports whether target ("claude" | "codex") can run event.
func SupportedBy(event, target string) bool {
	info, ok := eventInfo(event)
	if !ok {
		return false
	}
	switch target {
	case TargetClaude:
		return info.Claude
	case TargetCodex:
		return info.Codex
	}
	return false
}

// ValidID reports whether id is a safe hook id.
func ValidID(id string) bool {
	return strutil.ValidID(id)
}

// DefaultHello is seeded once so users have a ready-to-use hook that proves the
// pipeline works: typing "hello" injects a confirmation into the context.
func DefaultHello() config.HookDef {
	return config.HookDef{
		ID:          "hello-test",
		Name:        "Hello 触发测试",
		Description: "用户输入包含 hello 时注入提示，用于验证 hook 是否生效",
		Event:       "UserPromptSubmit",
		Command:     `P=$(jq -r '.prompt // empty'); echo "$P" | grep -qi hello && echo '{"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"⚠️ 检测到 hello 关键字！hook 已触发！"}}' || true`,
		UpdatedAt:   time.Now().Unix(),
	}
}

// Normalize trims fields, derives an id and validates event/command.
func Normalize(hook config.HookDef) (config.HookDef, error) {
	hook.ID = strutil.SanitizeID(strutil.OrDefault(strings.TrimSpace(hook.ID), strings.TrimSpace(hook.Name)))
	if hook.ID == "" {
		hook.ID = strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if !ValidID(hook.ID) {
		return config.HookDef{}, fmt.Errorf("invalid hook id")
	}
	hook.Name = strutil.OrDefault(strings.TrimSpace(hook.Name), hook.ID)
	hook.Event = strings.TrimSpace(hook.Event)
	if !ValidEvent(hook.Event) {
		return config.HookDef{}, fmt.Errorf("unsupported hook event: %s", hook.Event)
	}
	hook.Matcher = strings.TrimSpace(hook.Matcher)
	hook.Command = strings.TrimSpace(hook.Command)
	if hook.Command == "" {
		return config.HookDef{}, fmt.Errorf("hook command is required")
	}
	hook.Args = strutil.CleanSlice(hook.Args)
	hook.StatusMessage = strings.TrimSpace(hook.StatusMessage)
	if hook.Timeout < 0 {
		hook.Timeout = 0
	}
	hook.UpdatedAt = time.Now().Unix()
	return hook, nil
}

// Clone deep-copies a hook list.
func Clone(in []config.HookDef) []config.HookDef {
	out := make([]config.HookDef, len(in))
	for i, h := range in {
		out[i] = h
		out[i].Args = append([]string(nil), h.Args...)
	}
	return out
}

// Find returns the hook with the given id.
func Find(in []config.HookDef, id string) (config.HookDef, bool) {
	for _, h := range in {
		if h.ID == id {
			return h, true
		}
	}
	return config.HookDef{}, false
}

// CountStats returns sidenav counts for a hook list.
func CountStats(list []config.HookDef) Stats {
	st := Stats{Total: len(list)}
	for _, h := range list {
		if h.EnabledInClaude {
			st.Claude++
		}
		if h.EnabledInCodex {
			st.Codex++
		}
	}
	return st
}

// ParseDraft strips fences/prose and unmarshals an AI response into a HookDef.
// Unknown events fall back to PostToolUse.
func ParseDraft(raw string) (config.HookDef, error) {
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
	if !ValidEvent(draft.Event) {
		draft.Event = "PostToolUse"
	}
	return draft, nil
}
